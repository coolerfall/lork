package lork

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func socketWriterForTest(options ...func(*SocketWriterOption)) (Writer, *socketWriter) {
	writer := NewSocketWriter(options...)
	return writer, writer.(*syncWriter).ref.(*bytesWriter).ref.(*socketWriter)
}

func TestSocketWriterOwnsQueuedData(t *testing.T) {
	writer, socket := socketWriterForTest()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	socket.run = &socketRun{ctx: ctx, queue: make(chan *bytes.Buffer, 2)}
	for _, message := range []string{"first", "other"} {
		event := NewLogEvent()
		event.appendMessage(message)
		if err := writer.DoWrite(event); err != nil {
			t.Fatal(err)
		}
		event.Recycle()
	}
	for _, want := range []string{"first", "other"} {
		buf := <-socket.run.queue
		var event map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		if got := event[MessageFieldKey]; got != want {
			t.Fatalf("message = %v, want %s", got, want)
		}
		socket.buffers.Put(buf)
	}
}

func TestSocketWriterQueueDoesNotBlock(t *testing.T) {
	_, socket := socketWriterForTest()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	socket.run = &socketRun{ctx: ctx, queue: make(chan *bytes.Buffer, 1)}
	data := []byte("first")
	if n, err := socket.Write(data); err != nil || n != len(data) {
		t.Fatalf("Write = %d, %v", n, err)
	}
	copy(data, "other")
	if n, err := socket.Write(data); err != nil || n != len(data) {
		t.Fatalf("full Write = %d, %v", n, err)
	}
	buf := <-socket.run.queue
	if got := buf.String(); got != "first" {
		t.Fatalf("queued bytes = %q, want first", got)
	}
	socket.buffers.Put(buf)
	cancel()
	_, _ = socket.Write(data)
	if len(socket.run.queue) != 0 {
		t.Fatal("Write accepted data after cancellation")
	}
}

func BenchmarkSocketWriterEnqueue(b *testing.B) {
	writer, socket := socketWriterForTest()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	socket.run = &socketRun{ctx: ctx, queue: make(chan *bytes.Buffer, 1)}
	mw := NewMultiWriter()
	mw.AddWriter(writer) // The synthetic run prevents starting a network connection.
	logger := NewClassicLogger("socket", mw)
	write := func() {
		logger.Info().Str("field", `say "hello"`).Msg("hello\nworld")
		socket.buffers.Put(<-socket.run.queue)
	}
	write()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		write()
	}
}

type socketTestTransport struct {
	locker         sync.Mutex
	peers          []net.Conn
	closed         chan struct{}
	handshakes     chan int
	writes         chan struct{}
	observeWrites  atomic.Bool
	stallWrites    bool
	stallHandshake int
}

type socketTestConn struct {
	net.Conn
	transport *socketTestTransport
}

func (c *socketTestConn) Write(p []byte) (int, error) {
	if c.transport.observeWrites.Load() {
		select {
		case c.transport.writes <- struct{}{}:
		default:
		}
	}
	return c.Conn.Write(p)
}

func (s *socketTestTransport) dial(context.Context, string, string) (net.Conn, error) {
	s.locker.Lock()
	defer s.locker.Unlock()
	select {
	case <-s.closed:
		return nil, net.ErrClosed
	default:
	}
	client, server := net.Pipe()
	s.peers = append(s.peers, server)
	index := len(s.peers)
	go func() {
		defer server.Close()
		request, err := http.ReadRequest(bufio.NewReader(server))
		if err != nil {
			return
		}
		s.handshakes <- index
		if index == s.stallHandshake {
			_, _ = io.Copy(io.Discard, server)
			return
		}
		sum := sha1.Sum([]byte(request.Header.Get("Sec-WebSocket-Key") + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		_, err = fmt.Fprintf(server, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", base64.StdEncoding.EncodeToString(sum[:]))
		if err != nil {
			return
		}
		if s.stallWrites {
			<-s.closed
		} else {
			_, _ = io.Copy(io.Discard, server)
		}
	}()
	return &socketTestConn{Conn: client, transport: s}, nil
}

func (s *socketTestTransport) close() {
	s.locker.Lock()
	defer s.locker.Unlock()
	close(s.closed)
	for _, conn := range s.peers {
		_ = conn.Close()
	}
}

func newSocketLifecycleTest(t *testing.T, stallWrites bool, stallHandshake int) (*socketWriter, *socketTestTransport) {
	t.Helper()
	transport := &socketTestTransport{
		closed: make(chan struct{}), handshakes: make(chan int, 8), writes: make(chan struct{}, 1),
		stallWrites: stallWrites, stallHandshake: stallHandshake,
	}
	previous := websocket.DefaultDialer
	websocket.DefaultDialer = &websocket.Dialer{NetDialContext: transport.dial, HandshakeTimeout: 5 * time.Second}
	_, socket := socketWriterForTest(func(o *SocketWriterOption) {
		o.RemoteUrl = "ws://socket.test/"
		o.ReconnectionDelay = time.Millisecond
	})
	t.Cleanup(func() {
		transport.close()
		socket.Stop()
		websocket.DefaultDialer = previous
	})
	return socket, transport
}

func waitSocketHandshake(t *testing.T, transport *socketTestTransport, want int) {
	t.Helper()
	select {
	case got := <-transport.handshakes:
		if got != want {
			t.Fatalf("handshake = %d, want %d", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("handshake did not start")
	}
}

func stopSocketTest(t *testing.T, socket *socketWriter) {
	t.Helper()
	done := make(chan struct{})
	go func() { socket.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop did not terminate the worker")
	}
}

func TestSocketWriterStopAndRestart(t *testing.T) {
	socket, transport := newSocketLifecycleTest(t, false, 0)
	socket.Stop() // Stopping before Start is safe.
	socket.Start()
	waitSocketHandshake(t, transport, 1)
	firstRun := socket.run
	socket.Start()
	if socket.run != firstRun {
		t.Fatal("repeated Start replaced the active session")
	}
	stopSocketTest(t, socket)
	select {
	case <-firstRun.done:
	default:
		t.Fatal("Stop returned before the worker exited")
	}
	_, _ = socket.Write([]byte("after stop"))
	if len(firstRun.queue) != 0 {
		t.Fatal("Write queued data after Stop")
	}
	socket.Stop()
	socket.Start()
	waitSocketHandshake(t, transport, 2)
	if socket.run == firstRun {
		t.Fatal("restart reused the stopped session")
	}
	stopSocketTest(t, socket)
}

func TestSocketWriterStopPendingWrite(t *testing.T) {
	socket, transport := newSocketLifecycleTest(t, true, 0)
	socket.Start()
	transport.observeWrites.Store(true)
	_, _ = socket.Write([]byte("blocked write"))
	select {
	case <-transport.writes:
	case <-time.After(time.Second):
		t.Fatal("worker did not begin writing")
	}
	run := socket.run
	for i := 0; i < 3; i++ {
		_, _ = socket.Write([]byte("pending"))
	}
	stopSocketTest(t, socket)
	if len(run.queue) != 0 {
		t.Fatal("Stop left queued buffers unreleased")
	}
}

func TestSocketWriterStopInitialHandshake(t *testing.T) {
	socket, transport := newSocketLifecycleTest(t, false, 1)
	started := make(chan struct{})
	go func() { socket.Start(); close(started) }()
	waitSocketHandshake(t, transport, 1)
	stopSocketTest(t, socket)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("Start remained blocked after cancellation")
	}
}

func TestSocketWriterStopProxyHandshake(t *testing.T) {
	socket, transport := newSocketLifecycleTest(t, false, 1)
	websocket.DefaultDialer.Proxy = func(*http.Request) (*url.URL, error) {
		return url.Parse("http://proxy.test:8080")
	}
	started := make(chan struct{})
	go func() { socket.Start(); close(started) }()
	waitSocketHandshake(t, transport, 1)
	stopSocketTest(t, socket)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("Start remained blocked in the proxy handshake")
	}
}

func TestSocketWriterStopReconnectHandshake(t *testing.T) {
	socket, transport := newSocketLifecycleTest(t, false, 2)
	socket.Start()
	waitSocketHandshake(t, transport, 1)
	transport.locker.Lock()
	_ = transport.peers[0].Close()
	transport.locker.Unlock()
	_, _ = socket.Write([]byte("connection failed"))
	waitSocketHandshake(t, transport, 2)
	stopSocketTest(t, socket)
}

func TestSocketWriterConcurrentLifecycle(t *testing.T) {
	socket, transport := newSocketLifecycleTest(t, false, 0)
	// Keep draining handshake observations throughout repeated restarts.
	drainDone := make(chan struct{})
	defer close(drainDone)
	go func() {
		for {
			select {
			case <-transport.handshakes:
			case <-drainDone:
				return
			}
		}
	}()
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 20; j++ {
				socket.Start()
				_, _ = socket.Write([]byte("concurrent lifecycle"))
				socket.Stop()
			}
		}()
	}
	done := make(chan struct{})
	go func() { workers.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent lifecycle did not complete")
	}
	stopSocketTest(t, socket)
	socket.locker.Lock()
	defer socket.locker.Unlock()
	if socket.run != nil {
		t.Fatal("Stop left an active session")
	}
}
