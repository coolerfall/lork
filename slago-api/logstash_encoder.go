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

var (
	oldKey  = []byte(`"` + TimestampFieldKey + `"`)
	newKey  = []byte(`"@timestamp"`)
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

	var gotKey bool
	var start, end int
	for i, b := range p {
		if b == '"' {
			if start == 0 {
				start = i
			} else {
				end = i
			}
		} else {
			if start != 0 && end != 0 {
				s := p[start : end+1]
				if bytes.Compare(s, oldKey) == 0 {
					e.buf.Write(newKey)
					gotKey = true
				} else {
					e.buf.Write(s)
					if gotKey {
						e.buf.WriteByte(',')
						e.buf.Write(version)
						gotKey = false
					}
				}
				e.buf.WriteByte(b)

				start = 0
				end = 0
			} else if start == 0 && end == 0 {
				e.buf.WriteByte(b)
			}
		}
	}

	data = e.buf.Bytes()
	e.buf.Reset()

	return data, nil
}
