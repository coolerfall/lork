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
	"bytes"
	"sync"
)

const jsonTimeFormat = "2006-01-02T15:04:05.000Z07:00"

// jsonEncoder encodes logging event into json format.
type jsonEncoder struct {
	buf    *bytes.Buffer
	tsBuf  *bytes.Buffer
	locker sync.Mutex
}

// NewJsonEncoder creates a new instance of encoder to encode data to json.
func NewJsonEncoder() *jsonEncoder {
	return &jsonEncoder{
		buf:   new(bytes.Buffer),
		tsBuf: new(bytes.Buffer),
	}
}

func (e *jsonEncoder) Encode(p []byte) ([]byte, error) {
	e.locker.Lock()
	defer e.locker.Unlock()

	// wrtite key and value as json string
	err := ReplaceJson(p, e.buf, TimestampFieldKey,
		func(k, v []byte) (nk []byte, nv []byte, err error) {
			bufData := e.tsBuf.Bytes()
			bufData, err = convertFormat(bufData, v, TimestampFormat, jsonTimeFormat)
			if err != nil {
				return nil, nil, err
			}
			e.tsBuf.Reset()
			e.tsBuf.Write(bufData)
			timestamp := e.tsBuf.Bytes()
			e.tsBuf.Reset()

			return k, timestamp, nil
		})

	p = e.buf.Bytes()
	e.buf.Reset()

	return p, err
}
