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

package lork

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
	Name              string
	RemoteUrl         string
	QueueSize         int
	ReconnectionDelay time.Duration
	Filter            Filter
}

type socketWriter struct {
	opts    *SocketWriterOption
	encoder Encoder

	locker    sync.Mutex
	conn      *websocket.Conn
	queue     *BlockingQueue
	isStarted bool

	remoteUrl *url.URL
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
	}

	return NewBytesWriter(sw)
}

func (w *socketWriter) Start() {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.opts.QueueSize <= 0 {
		w.opts.QueueSize = defaultSocketQueueSize
	}
	if w.opts.ReconnectionDelay <= 0 {
		w.opts.ReconnectionDelay = defaultReconnectionDelay
	}

	remoteUrl, err := url.Parse(w.opts.RemoteUrl)
	if err != nil {
		ReportfExit("socket writer needs a available remote url: %v", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(remoteUrl.String(), nil)
	if err != nil {
		ReportfExit("connect socket server error, check your remote url: %v", err)
	}

	w.remoteUrl = remoteUrl
	w.conn = conn
	w.queue = NewBlockingQueue(w.opts.QueueSize)

	if w.isStarted {
		return
	}
	w.isStarted = true
	go w.startWorker()
}

func (w *socketWriter) Stop() {
	w.locker.Lock()
	defer w.locker.Unlock()

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

func (w *socketWriter) Name() string {
	return w.opts.Name
}

func (w *socketWriter) Encoder() Encoder {
	return w.encoder
}

func (w *socketWriter) Filter() Filter {
	return w.opts.Filter
}

func (w *socketWriter) startWorker() {
	for {
		if !w.isStarted {
			break
		}

		event := (w.queue.Take()).([]byte)
		err := w.conn.WriteMessage(websocket.BinaryMessage, event)
		if err == nil {
			continue
		}

		// close first
		_ = w.conn.Close()
		Reportf("socket writer write error: %v", err)

		// delay before reconnect
		time.Sleep(w.opts.ReconnectionDelay)
		conn, _, err := websocket.DefaultDialer.Dial(w.remoteUrl.String(), nil)
		if err != nil {
			Reportf("socket writer reconnect error: %v", err)
		} else {
			w.conn = conn
		}
	}
}
