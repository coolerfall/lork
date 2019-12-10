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
	"github.com/buger/jsonparser"
	"sync"
)

const jsonTimeFormat = "2006-01-02T15:04:05.000Z07:00"

// jsonEncoder encodes logging event into json format.
type jsonEncoder struct {
	buf   *bytes.Buffer
	mutex sync.Mutex
}

// NewJsonEncoder creates a new instance of encoder to encode data to json.
func NewJsonEncoder() *jsonEncoder {
	return &jsonEncoder{
		buf: &bytes.Buffer{},
	}
}

func (e *jsonEncoder) Encode(p []byte) ([]byte, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	var err error
	ts, _, _, _ := jsonparser.Get(p, TimestampFieldKey)
	bufData := e.buf.Bytes()
	bufData, err = convertFormat(bufData, string(ts), jsonTimeFormat)
	if err != nil {
		return nil, err
	}
	e.buf.Reset()
	e.buf.Write(bufData)
	timestamp := e.buf.String()
	e.buf.Reset()

	e.buf.WriteByte('{')
	var start = false
	_ = jsonparser.ObjectEach(p, func(key []byte, value []byte,
		dataType jsonparser.ValueType, offset int) error {
		if start {
			e.buf.WriteByte(',')
		} else {
			start = true
		}

		e.buf.WriteByte('"')
		e.buf.Write(key)
		e.buf.WriteByte('"')
		e.buf.WriteByte(':')

		switch dataType {
		case jsonparser.String:
			e.buf.WriteByte('"')
			if string(key) == TimestampFieldKey {
				e.buf.WriteString(timestamp)
			} else {
				e.buf.Write(value)
			}
			e.buf.WriteByte('"')

		default:
			e.buf.Write(value)
		}

		return nil
	})
	e.buf.WriteByte('}')
	e.buf.WriteByte('\n')
	p = e.buf.Bytes()
	e.buf.Reset()

	// the default encoder is json, just return origin data
	return p, nil
}
