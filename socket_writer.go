// Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
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

package slago

import (
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
	RemoteUrl         *url.URL
	QueueSize         int
	ReconnectionDelay time.Duration
	Filter            Filter
}

type socketWriter struct {
	encoder Encoder
	filter  Filter

	locker    sync.Mutex
	conn      *websocket.Conn
	queue     *blockingQueue
	isStarted bool

	remoteUrl   *url.URL
	reconnDelay time.Duration
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

	if opts.RemoteUrl == nil {
		ReportfExit("socket writer need a remote url")
	}

	if opts.QueueSize <= 0 {
		opts.QueueSize = defaultSocketQueueSize
	}
	if opts.ReconnectionDelay <= 0 {
		opts.ReconnectionDelay = defaultReconnectionDelay
	}

	conn, _, err := websocket.DefaultDialer.Dial(opts.RemoteUrl.String(), nil)
	if err != nil {
		ReportfExit("connect socket server error: %v", err)
	}

	return &socketWriter{
		encoder:     NewJsonEncoder(),
		filter:      opts.Filter,
		conn:        conn,
		queue:       NewBlockingQueue(opts.QueueSize),
		reconnDelay: opts.ReconnectionDelay,
		remoteUrl:   opts.RemoteUrl,
	}
}

func (w *socketWriter) Start() {
	if w.isStarted {
		return
	}
	w.isStarted = true
	go w.startWorker()
}

func (w *socketWriter) Stop() {
	err := w.conn.Close()
	if err != nil {
		Reportf("stop socket writer error: %v", err)
	}
}

func (w *socketWriter) Write(p []byte) (int, error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.queue.RemainCapacity() <= 2 {
		// discard
		return 0, nil
	}

	w.queue.Put(p)

	return len(p), nil
}

func (w *socketWriter) Encoder() Encoder {
	return w.encoder
}

func (w *socketWriter) Filter() Filter {
	return w.filter
}

func (w *socketWriter) startWorker() {
	for {
		if !w.isStarted {
			break
		}

		p := w.queue.Take()

		err := w.conn.WriteMessage(websocket.BinaryMessage, p)
		if err == nil {
			continue
		}

		// close first
		_ = w.conn.Close()
		Reportf("socket writer write error: %v", err)

		// delay before reconnect
		time.Sleep(w.reconnDelay)
		conn, _, err := websocket.DefaultDialer.Dial(w.remoteUrl.String(), nil)
		if err != nil {
			Reportf("socket writer reconnect error: %v", err)
		} else {
			w.conn = conn
		}
	}
}
