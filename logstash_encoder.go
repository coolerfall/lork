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

	_ = ReplaceJson(p, e.buf, TimestampFieldKey,
		func(k, v []byte) (nk, kv []byte, err error) {
			return newKey, v, nil
		})

	data = e.buf.Bytes()
	e.buf.Truncate(len(data) - 2)
	e.buf.WriteByte(',')
	e.buf.Write(version)
	e.buf.WriteString("}\n")

	data = e.buf.Bytes()
	e.buf.Reset()

	return data, nil
}
