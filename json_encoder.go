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
	"bytes"
	"sync"
)

// jsonEncoder encodes logging event into json format.
type jsonEncoder struct {
	locker sync.Mutex
	buf    *bytes.Buffer
	tsBuf  *bytes.Buffer
}

// NewJsonEncoder creates a new instance of encoder to encode data to json.
func NewJsonEncoder() Encoder {
	return &jsonEncoder{
		buf:   new(bytes.Buffer),
		tsBuf: new(bytes.Buffer),
	}
}

func (je *jsonEncoder) Encode(e *LogEvent) ([]byte, error) {
	je.locker.Lock()
	defer je.locker.Unlock()

	var err error
	bufData := je.tsBuf.Bytes()
	bufData, err = appendFormatUnix(bufData, e.Timestamp(), TimeFormatRFC3339)
	if err != nil {
		return nil, err
	}
	je.tsBuf.Reset()
	je.tsBuf.Write(bufData)
	timestamp := je.tsBuf.Bytes()
	je.tsBuf.Reset()

	// write key and value as json string
	je.buf.WriteString("{")
	je.writeKeyAndValue(TimestampFieldKey, timestamp, true)
	je.writeKeyAndValue(LevelFieldKey, e.Level(), true)
	je.writeKeyAndValue(LoggerNameFieldKey, e.LoggerName(), true)
	je.writeKeyAndValue(MessageFieldKey, e.Message(), true)

	_ = e.Fields(func(k, v []byte, isString bool) error {
		je.writeKeyAndValue(string(k), v, isString)
		return nil
	})

	je.buf.Truncate(je.buf.Len() - 1)
	je.buf.WriteString("}\n")

	p := je.buf.Bytes()
	je.buf.Reset()

	return p, err
}

func (je *jsonEncoder) writeKeyAndValue(key string, value []byte, isString bool) {
	je.buf.WriteByte('"')
	je.buf.WriteString(key)
	je.buf.WriteByte('"')
	je.buf.WriteByte(':')

	if isString {
		je.buf.WriteByte('"')
	}
	je.buf.Write(value)
	if isString {
		je.buf.WriteByte('"')
	}
	je.buf.WriteByte(',')
}
