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
	"unicode/utf8"
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
	je.buf.Write(appendJSONString(je.buf.AvailableBuffer(), key))
	je.buf.WriteByte(':')

	if isString {
		je.buf.Write(appendJSONString(je.buf.AvailableBuffer(), value))
	} else {
		je.buf.Write(value)
	}
	je.buf.WriteByte(',')
}

// appendJSONString quotes a string directly into reusable buffer storage.
func appendJSONString[S string | []byte](dst []byte, value S) []byte {
	const hex = "0123456789abcdef"
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(value); {
		c := value[i]
		if c >= utf8.RuneSelf {
			end := i + min(len(value)-i, utf8.UTFMax)
			r, size := utf8.DecodeRuneInString(string(value[i:end]))
			if r == utf8.RuneError && size == 1 {
				dst = append(dst, value[start:i]...)
				dst = append(dst, `\ufffd`...)
				i++
				start = i
				continue
			}
			i += size
			continue
		}
		if c >= 0x20 && c != '"' && c != '\\' {
			i++
			continue
		}
		dst = append(dst, value[start:i]...)
		switch c {
		case '"', '\\':
			dst = append(dst, '\\', c)
		case '\b':
			dst = append(dst, `\b`...)
		case '\f':
			dst = append(dst, `\f`...)
		case '\n':
			dst = append(dst, `\n`...)
		case '\r':
			dst = append(dst, `\r`...)
		case '\t':
			dst = append(dst, `\t`...)
		default:
			dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
		}
		i++
		start = i
	}
	dst = append(dst, value[start:]...)
	return append(dst, '"')
}
