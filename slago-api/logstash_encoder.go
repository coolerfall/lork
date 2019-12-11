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

	"github.com/buger/jsonparser"
)

var (
	newKey  = []byte(`@timestamp`)
	version = []byte(`"@version":"1"`)
)

// logstashEncoder encodes logging event into logstash json format.
type logstashEncoder struct {
	mutex sync.Mutex
	buf   *bytes.Buffer
}

// NewLogstashEncoder creates a new instance of logstash encoder.
func NewLogstashEncoder() *logstashEncoder {
	return &logstashEncoder{
		buf: &bytes.Buffer{},
	}
}

func (e *logstashEncoder) Encode(p []byte) (data []byte, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.buf.WriteByte('{')
	_ = jsonparser.ObjectEach(p, func(key []byte, value []byte,
		dataType jsonparser.ValueType, _ int) error {
		e.buf.WriteByte('"')

		if string(key) == TimestampFieldKey {
			e.buf.Write(newKey)
		} else {
			e.buf.Write(key)
		}
		e.buf.WriteByte('"')
		e.buf.WriteByte(':')

		switch dataType {
		case jsonparser.String:
			e.buf.WriteByte('"')
			e.buf.Write(value)
			e.buf.WriteByte('"')

		default:
			e.buf.Write(value)
		}
		e.buf.WriteByte(',')

		return nil
	})
	e.buf.Write(version)
	e.buf.WriteByte('}')
	e.buf.WriteByte('\n')

	data = e.buf.Bytes()
	e.buf.Reset()

	return data, nil
}
