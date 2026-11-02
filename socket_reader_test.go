package lork

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func stopSocketReaderTest(t *testing.T, reader *SocketReader) {
	t.Helper()
	done := make(chan struct{})
	go func() { reader.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("SocketReader.Stop did not complete")
	}
}

func startSocketReaderTest(t *testing.T, reader *SocketReader) (*socketReaderRun, <-chan struct{}) {
	t.Helper()
	done := make(chan struct{})
	go func() { reader.Start(); close(done) }()
	t.Cleanup(func() { stopSocketReaderTest(t, reader) })
	deadline := time.Now().Add(2 * time.Second)
	for {
		reader.locker.Lock()
		run := reader.run
		reader.locker.Unlock()
		if run != nil {
			return run, done
		}
		select {
		case <-done:
			t.Fatal("SocketReader.Start returned without starting")
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("SocketReader did not start")
		}
		time.Sleep(time.Millisecond)
	}
}

func socketReaderAddress(run *socketReaderRun) string {
	return fmt.Sprintf("127.0.0.1:%d", run.listener.Addr().(*net.TCPAddr).Port)
}

func newSocketReaderTest() *SocketReader {
	return NewSocketReader(func(o *SocketReaderOption) { o.Port = 0 })
}

func TestSocketReaderStopAndRestart(t *testing.T) {
	reader := newSocketReaderTest()
	reader.Stop()
	for range 2 {
		run, done := startSocketReaderTest(t, reader)
		reader.Start()
		if reader.run != run {
			t.Fatal("repeated Start replaced the running server")
		}
		address := socketReaderAddress(run)
		conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+reader.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		stopSocketReaderTest(t, reader)
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Start remained blocked after Stop")
		}
		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		if _, _, err := conn.ReadMessage(); err == nil {
			t.Fatal("idle WebSocket stayed open after Stop")
		}
		if conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond); err == nil {
			_ = conn.Close()
			t.Fatal("listener accepted a connection after Stop")
		}
		reader.Stop()
		// The next run must also be able to reuse the same listening port.
		reader.port = run.listener.Addr().(*net.TCPAddr).Port
	}
}

func TestSocketReaderPrivateRoutesAndFailedUpgrade(t *testing.T) {
	previous := http.DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()
	http.HandleFunc("/ws/log", func(http.ResponseWriter, *http.Request) {
		t.Error("reader used the process-wide route")
	})
	t.Cleanup(func() { http.DefaultServeMux = previous })
	for range 2 {
		reader := newSocketReaderTest()
		run, _ := startSocketReaderTest(t, reader)
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "http://reader.test/ws/log", nil)
		reader.readLog(run, response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("failed upgrade returned %d, want 400", response.Code)
		}
		conn, _, err := websocket.DefaultDialer.Dial("ws://"+socketReaderAddress(run)+reader.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.Close()
	}
}

func TestSocketReaderStartFailureCanRetry(t *testing.T) {
	reservation, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reservation.Close() })
	reader := NewSocketReader(func(o *SocketReaderOption) {
		o.Port = reservation.Addr().(*net.TCPAddr).Port
	})
	reader.Start()
	if reader.run != nil {
		t.Fatal("failed listen left an active run")
	}
	reader.Stop()
	_ = reservation.Close()
	for _, path := range []string{"", "/{"} {
		reader.path = path
		reader.Start()
		if reader.run != nil {
			t.Fatal("invalid route left an active run")
		}
		reader.Stop()
	}
	reader.path = "/ws/log"
	startSocketReaderTest(t, reader)
	stopSocketReaderTest(t, reader)
}

func TestSocketReaderTerminalReadError(t *testing.T) {
	reader := newSocketReaderTest()
	run, _ := startSocketReaderTest(t, reader)
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+socketReaderAddress(run)+reader.path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	err = conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "closed"), time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		reader.locker.Lock()
		count := len(run.connections)
		reader.locker.Unlock()
		if count == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("terminal read error did not release the connection")
		}
		time.Sleep(time.Millisecond)
	}
}

type socketReaderHijacker struct {
	*httptest.ResponseRecorder
	conn     net.Conn
	hijacked chan struct{}
}

func (w *socketReaderHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	close(w.hijacked)
	return w.conn, bufio.NewReadWriter(bufio.NewReader(w.conn), bufio.NewWriter(w.conn)), nil
}

func TestSocketReaderStopPendingUpgrade(t *testing.T) {
	reader := newSocketReaderTest()
	run, _ := startSocketReaderTest(t, reader)
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	run.server.ConnState(server, http.StateNew)
	response := &socketReaderHijacker{httptest.NewRecorder(), server, make(chan struct{})}
	request := httptest.NewRequest(http.MethodGet, "http://reader.test/ws/log", nil)
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Version", "13")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	done := make(chan struct{})
	go func() { reader.readLog(run, response, request); close(done) }()
	select {
	case <-response.hijacked:
	case <-time.After(time.Second):
		t.Fatal("upgrade did not hijack the connection")
	}
	stopSocketReaderTest(t, reader)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop did not interrupt the pending WebSocket handshake")
	}
}

func TestSocketReaderConcurrentLifecycle(t *testing.T) {
	reader := newSocketReaderTest()
	t.Cleanup(func() { stopSocketReaderTest(t, reader) })
	finish := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			select {
			case <-finish:
				return
			case <-time.After(time.Millisecond):
				reader.Stop()
			}
		}
	}()
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 10 {
				reader.Start()
			}
		}()
	}
	done := make(chan struct{})
	go func() { workers.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("concurrent Start and Stop did not complete")
	}
	close(finish)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("concurrent Stop remained blocked")
	}
	stopSocketReaderTest(t, reader)
	reader.locker.Lock()
	defer reader.locker.Unlock()
	if reader.run != nil {
		t.Fatal("concurrent lifecycle left an active run")
	}
}
