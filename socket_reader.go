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
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type SocketReader struct {
	locker    sync.Mutex
	isRunning bool
	upgrader  *websocket.Upgrader
	path      string
	port      int
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
	sr.isRunning = true
	sr.locker.Unlock()

	LoggerC().Info().Msgf("socket reader is listening on %v with path %v", sr.port, sr.path)
	http.HandleFunc(sr.path, sr.readLog)
	fmt.Print(http.ListenAndServe(fmt.Sprintf(":%v", sr.port), nil))
}

func (sr *SocketReader) Stop() {
	sr.locker.Lock()
	sr.isRunning = false
	sr.locker.Unlock()
}

func (sr *SocketReader) readLog(w http.ResponseWriter, r *http.Request) {
	conn, err := sr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		LoggerC().Error().Err(err).Msg("read log upgrade error")
	}
	defer func() {
		_ = conn.Close()
	}()

	LoggerC().Info().Msgf("socket reader got a client connected: %v", r.Host)

	for {
		if !sr.isRunning {
			LoggerC().Info().Msg("socket log reader stopped, closing...")
			break
		}

		if msgType, data, err := conn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				LoggerC().Info().Msg("socket client has closed")
				break
			}

			LoggerC().Error().Err(err).Msg("read log err")
			continue
		} else {
			if msgType != websocket.BinaryMessage {
				LoggerC().Debug().Msg("not binary message, skipping")
				continue
			}

			LoggerC().Event(MakeEvent(data))
		}
	}
}
