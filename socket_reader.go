// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
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
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type SocketReader struct {
	locker    sync.Mutex
	isRunning bool
	upgrader  *websocket.Upgrader
}

// NewSocketReader creates a new instance of socket reader.
func NewSocketReader() *SocketReader {
	return &SocketReader{
		upgrader: &websocket.Upgrader{},
	}
}

func (sr *SocketReader) Start() {
	sr.locker.Lock()
	sr.isRunning = true
	sr.locker.Unlock()

	http.HandleFunc("/log/socket", sr.readLog)
	fmt.Print(http.ListenAndServe(":6060", nil))
}

func (sr *SocketReader) Stop() {
	sr.locker.Lock()
	sr.isRunning = false
	sr.locker.Unlock()
}

func (sr *SocketReader) readLog(w http.ResponseWriter, r *http.Request) {
	conn, err := sr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger().Error().Err(err).Msg("read log upgrade error")
	}
	defer func() {
		_ = conn.Close()
	}()

	Logger().Info().Msgf("socket reader got a client connected: %v", r.Host)

	for {
		if !sr.isRunning {
			Logger().Info().Msg("socket log reader stopped, closing...")
			break
		}

		if msgType, data, err := conn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				Logger().Info().Msg("socket client has closed")
				break
			}

			Logger().Error().Err(err).Msg("read log err")
			continue
		} else {
			if msgType != websocket.BinaryMessage {
				Logger().Debug().Msg("not binary message for log, skip")
				continue
			}

			Logger().WriteRaw(data)
		}
	}
}
