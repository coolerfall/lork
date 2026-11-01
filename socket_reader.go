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
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type SocketReader struct {
	locker   sync.Mutex
	run      *socketReaderRun
	upgrader *websocket.Upgrader
	path     string
	port     int
}

// Each run owns its listener and connections, including hijacked connections
// whose WebSocket handshake has not finished yet.
type socketReaderRun struct {
	server      *http.Server
	listener    net.Listener
	connections map[net.Conn]struct{}
	stopping    bool
	done        chan struct{}
}

type SocketReaderOption struct {
	Path string
	Port int
}

// NewSocketReader creates a new instance of socket reader.
func NewSocketReader(options ...func(*SocketReaderOption)) *SocketReader {
	opts := &SocketReaderOption{
		Path: "/ws/log",
		Port: 6060,
	}

	for _, f := range options {
		f(opts)
	}

	return &SocketReader{
		upgrader: &websocket.Upgrader{},
		path:     opts.Path,
		port:     opts.Port,
	}
}

func (sr *SocketReader) Start() {
	sr.locker.Lock()
	if sr.run != nil {
		sr.locker.Unlock()
		return
	}

	run := &socketReaderRun{
		connections: make(map[net.Conn]struct{}),
		done:        make(chan struct{}),
	}
	mux, err := sr.newMux(run)
	if err != nil {
		sr.locker.Unlock()
		Reportf("socket reader route error: %v", err)
		return
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", sr.port))
	if err != nil {
		sr.locker.Unlock()
		Reportf("socket reader listen error: %v", err)
		return
	}
	run.listener = listener
	run.server = &http.Server{
		Handler: mux,
		ConnState: func(conn net.Conn, state http.ConnState) {
			sr.locker.Lock()
			if state == http.StateNew {
				if run.stopping {
					sr.locker.Unlock()
					_ = conn.Close()
					return
				}
				run.connections[conn] = struct{}{}
			} else if state == http.StateClosed {
				delete(run.connections, conn)
			}
			sr.locker.Unlock()
		},
	}
	sr.run = run
	sr.locker.Unlock()
	defer sr.finishRun(run)

	LoggerC().Info().Msgf("socket reader is listening on %v with path %v", sr.port, sr.path)
	if err := run.server.Serve(listener); err != nil &&
		!errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
		Reportf("socket reader serve error: %v", err)
	}
}

func (sr *SocketReader) newMux(run *socketReaderRun) (mux *http.ServeMux, err error) {
	// ServeMux reports invalid route patterns by panicking. Treat configuration
	// failures like listen failures, before acquiring any network resources.
	defer func() {
		if cause := recover(); cause != nil {
			err = fmt.Errorf("invalid path %q: %v", sr.path, cause)
		}
	}()
	mux = http.NewServeMux()
	mux.HandleFunc(sr.path, func(w http.ResponseWriter, r *http.Request) {
		sr.readLog(run, w, r)
	})
	return mux, nil
}

func (sr *SocketReader) Stop() {
	sr.locker.Lock()
	run := sr.run
	if run == nil {
		sr.locker.Unlock()
		return
	}
	run.stopping = true
	sr.locker.Unlock()
	// Close the listener explicitly as Stop can precede the call to Serve.
	_ = run.listener.Close()
	_ = run.server.Close()
	<-run.done
}

func (sr *SocketReader) finishRun(run *socketReaderRun) {
	sr.locker.Lock()
	run.stopping = true
	connections := run.connections
	run.connections = nil
	sr.locker.Unlock()

	_ = run.listener.Close()
	_ = run.server.Close()
	for conn := range connections {
		_ = conn.Close()
	}

	sr.locker.Lock()
	if sr.run == run {
		sr.run = nil
	}
	close(run.done)
	sr.locker.Unlock()
}

func (sr *SocketReader) running(run *socketReaderRun) bool {
	sr.locker.Lock()
	defer sr.locker.Unlock()
	return sr.run == run && !run.stopping
}

func (sr *SocketReader) readLog(run *socketReaderRun, w http.ResponseWriter, r *http.Request) {
	if !sr.running(run) {
		http.Error(w, "socket reader stopped", http.StatusServiceUnavailable)
		return
	}
	conn, err := sr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		LoggerC().Error().Err(err).Msg("read log upgrade error")
		return
	}
	defer func() {
		_ = conn.Close()
		sr.locker.Lock()
		delete(run.connections, conn.UnderlyingConn())
		sr.locker.Unlock()
	}()

	LoggerC().Info().Msgf("socket reader got a client connected: %v", r.Host)

	for {
		if !sr.running(run) {
			return
		}

		if msgType, data, err := conn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || !sr.running(run) {
				LoggerC().Info().Msg("socket client has closed")
				return
			}

			LoggerC().Error().Err(err).Msg("read log err")
			return
		} else {
			if !sr.running(run) {
				return
			}
			if msgType != websocket.BinaryMessage {
				LoggerC().Debug().Msg("not binary message, skipping")
				continue
			}

			LoggerC().Event(MakeEvent(data))
		}
	}
}
