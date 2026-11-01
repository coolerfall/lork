// Copyright (c) 2019-2023 Vincent Cheung (coolingfall@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lork

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultSocketQueueSize   = 128
	defaultReconnectionDelay = 5_000
)

type SocketWriterOption struct {
	Name              string
	RemoteUrl         string
	QueueSize         int
	ReconnectionDelay time.Duration
	Filter            Filter
}

type socketWriter struct {
	opts    *SocketWriterOption
	encoder Encoder

	locker  sync.Mutex
	buffers sync.Pool
	run     *socketRun
}

// socketRun owns the connection and pending data for one Start/Stop cycle.
type socketRun struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	ready  chan error
	queue  chan *bytes.Buffer
	conn   *websocket.Conn // Replaced and closed under socketWriter.locker.
	url    string
}

// NewSocketWriter create a logging writer via socket.
func NewSocketWriter(options ...func(*SocketWriterOption)) Writer {
	opts := &SocketWriterOption{
		QueueSize:         defaultSocketQueueSize,
		ReconnectionDelay: defaultReconnectionDelay,
	}

	for _, f := range options {
		f(opts)
	}

	sw := &socketWriter{
		opts:    opts,
		encoder: NewJsonEncoder(),
		buffers: sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
	}

	return NewBytesWriter(sw)
}

func (w *socketWriter) Start() {
	w.locker.Lock()

	if w.run != nil {
		w.locker.Unlock()
		return
	}

	if w.opts.QueueSize <= 0 {
		w.opts.QueueSize = defaultSocketQueueSize
	}
	if w.opts.ReconnectionDelay <= 0 {
		w.opts.ReconnectionDelay = defaultReconnectionDelay
	}

	remoteUrl, err := url.Parse(w.opts.RemoteUrl)
	if err != nil {
		w.locker.Unlock()
		ReportfExit("socket writer needs a available remote url: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	run := &socketRun{
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
		ready:  make(chan error, 1),
		queue:  make(chan *bytes.Buffer, w.opts.QueueSize),
		url:    remoteUrl.String(),
	}
	w.run = run
	go w.startWorker(run)
	w.locker.Unlock()

	if err := <-run.ready; err != nil && err != context.Canceled {
		ReportfExit("connect socket server error, check your remote url: %v", err)
	}
}

func (w *socketWriter) Stop() {
	w.locker.Lock()
	run := w.run
	if run == nil {
		w.locker.Unlock()
		return
	}
	if run.ctx.Err() == nil {
		run.cancel()
		if run.conn != nil {
			_ = run.conn.Close()
		}
	}
	w.locker.Unlock()

	<-run.done
	w.locker.Lock()
	if w.run == run {
		w.run = nil
	}
	w.locker.Unlock()
}

func (w *socketWriter) Write(p []byte) (int, error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	run := w.run
	if run == nil || run.ctx.Err() != nil || len(run.queue) == cap(run.queue) {
		return len(p), nil
	}

	// The encoder reuses p after Write returns; queued data needs its own storage.
	buf := w.buffers.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write(p)
	run.queue <- buf

	return len(p), nil
}

func (w *socketWriter) Name() string {
	return w.opts.Name
}

func (w *socketWriter) Encoder() Encoder {
	return w.encoder
}

func (w *socketWriter) Filter() Filter {
	return w.opts.Filter
}

func (w *socketWriter) startWorker(run *socketRun) {
	var conn *websocket.Conn
	defer close(run.done)
	defer func() {
		w.locker.Lock()
		run.cancel()
		w.locker.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
		for {
			select {
			case buf := <-run.queue:
				w.buffers.Put(buf)
			default:
				return
			}
		}
	}()

	var err error
	conn, err = dialSocket(run.ctx, run.url)
	if err != nil {
		if run.ctx.Err() != nil {
			err = run.ctx.Err()
		}
		run.ready <- err
		return
	}
	w.locker.Lock()
	if run.ctx.Err() != nil {
		w.locker.Unlock()
		run.ready <- run.ctx.Err()
		return
	}
	run.conn = conn
	w.locker.Unlock()
	run.ready <- nil

	for {
		var buf *bytes.Buffer
		select {
		case <-run.ctx.Done():
			return
		case buf = <-run.queue:
		}

		if run.ctx.Err() != nil {
			w.buffers.Put(buf)
			return
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
		w.buffers.Put(buf)
		if run.ctx.Err() != nil {
			return
		}
		if err == nil {
			continue
		}

		// close first
		_ = conn.Close()
		Reportf("socket writer write error: %v", err)

		// Both the reconnect delay and the handshake must stop on cancellation.
		timer := time.NewTimer(w.opts.ReconnectionDelay)
		select {
		case <-run.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		reconnected, err := dialSocket(run.ctx, run.url)
		if err != nil {
			if run.ctx.Err() != nil {
				return
			}
			Reportf("socket writer reconnect error: %v", err)
		} else {
			w.locker.Lock()
			if run.ctx.Err() != nil {
				w.locker.Unlock()
				_ = reconnected.Close()
				return
			}
			run.conn = reconnected
			conn = reconnected
			w.locker.Unlock()
		}
	}
}

func dialSocket(ctx context.Context, remoteURL string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 45 * time.Second}
	if websocket.DefaultDialer != nil {
		dialer = *websocket.DefaultDialer
	}
	dial := dialer.NetDialContext
	if dial == nil {
		if dialer.NetDial != nil {
			dial = func(_ context.Context, network, addr string) (net.Conn, error) {
				return dialer.NetDial(network, addr)
			}
		} else {
			dial = (&net.Dialer{}).DialContext
		}
	}
	// Register before proxy CONNECT as well as the WebSocket upgrade. DialContext
	// alone cannot cancel either handshake once it is blocked reading a response.
	var stop func() bool
	watch := func(dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
		return func(dialCtx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dial(dialCtx, network, addr)
			if err == nil {
				stop = context.AfterFunc(ctx, func() { _ = conn.Close() })
			}
			return conn, err
		}
	}
	dialer.NetDialContext = watch(dial)
	if dialer.NetDialTLSContext != nil {
		dialer.NetDialTLSContext = watch(dialer.NetDialTLSContext)
	}
	conn, _, err := dialer.DialContext(ctx, remoteURL, nil)
	if stop != nil {
		stop()
	}
	return conn, err
}
