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
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

type socketWriter struct {
	conn    *websocket.Conn
	mutex   sync.Mutex
	encoder Encoder
}

// TODO: reconnection
// NewSocketWriter create a logging writter via socket.
func NewSocketWriter(u *url.URL) *socketWriter {
	if u == nil {
		Logger().Error().Msg("connect socket server error")
		return nil
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Logger().Error().Err(err).Msg("connect socket server error")
	}

	return &socketWriter{
		conn:    conn,
		encoder: NewJsonEncoder(),
	}
}

func (w *socketWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	err = w.conn.WriteMessage(websocket.BinaryMessage, p)
	defer w.mutex.Unlock()
	return len(p), err
}

func (w *socketWriter) Close() error {
	return w.conn.Close()
}

func (w *socketWriter) Encoder() Encoder {
	return w.encoder
}

func (w *socketWriter) Filter() *Filter {
	return nil
}
