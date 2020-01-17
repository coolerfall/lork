// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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
	"strconv"
	"sync"

	"github.com/buger/jsonparser"
)

type LogEvent struct {
	rfc3339Nano *bytes.Buffer
	level       *bytes.Buffer
	logger      *bytes.Buffer
	caller      *bytes.Buffer
	message     *bytes.Buffer
	fields      *bytes.Buffer
	fieldsIndex *bytes.Buffer
}

var (
	eventPool = &sync.Pool{
		New: func() interface{} {
			return &LogEvent{
				level:       new(bytes.Buffer),
				rfc3339Nano: new(bytes.Buffer),
				logger:      new(bytes.Buffer),
				caller:      new(bytes.Buffer),
				message:     new(bytes.Buffer),
				fields:      new(bytes.Buffer),
				fieldsIndex: new(bytes.Buffer),
			}
		},
	}
)

// Time returns rfc3339nano bytes.
func (e *LogEvent) Time() []byte {
	return e.rfc3339Nano.Bytes()
}

// LevelInt returns level int value.
func (e *LogEvent) LevelInt() Level {
	return ParseLevel(e.level.String())
}

// Level returns level string bytes.
func (e *LogEvent) Level() []byte {
	return e.level.Bytes()
}

// Logger return logger name bytes.
func (e *LogEvent) Logger() []byte {
	return e.logger.Bytes()
}

// Message returns message bytes.
func (e *LogEvent) Message() []byte {
	return e.message.Bytes()
}

// Fields gets extra key and value bytes.
func (e *LogEvent) Fields(callback func(k, v []byte, isString bool)) {
	var kvIndex = 0
	var startIndex = 0
	var ik, iv int
	kvArr := e.fields.Bytes()
	indexArr := e.fieldsIndex.Bytes()
	for i := 0; i < len(indexArr); i++ {
		if indexArr[i] == ',' {
			var s = string(indexArr[startIndex:i])
			ik, _ = strconv.Atoi(s)
			startIndex = i + 1
		} else if indexArr[i] == '|' {
			var s = string(indexArr[startIndex : i-1])
			iv, _ = strconv.Atoi(s)
			startIndex = i + 1
			bitSet := indexArr[i-1]
			var isString bool
			if bitSet == byte(1) {
				isString = true
			}
			callback(kvArr[kvIndex:kvIndex+ik], kvArr[kvIndex+ik:kvIndex+ik+iv], isString)
			kvIndex += ik + iv
		}
	}
}

func makeEvent(p []byte) *LogEvent {
	event := eventPool.Get().(*LogEvent)

	_ = jsonparser.ObjectEach(p, func(k []byte, v []byte,
		dataType jsonparser.ValueType, _ int) error {
		switch string(k) {
		case TimestampFieldKey:
			event.rfc3339Nano.Write(v)
		case LevelFieldKey:
			event.level.Write(v)
		case LoggerFieldKey:
			event.logger.Write(v)
		case MessageFieldKey:
			event.message.Write(v)

		default:
			event.fields.Write(k)
			event.fields.Write(v)
			event.fieldsIndex.WriteString(strconv.Itoa(len(k)))
			event.fieldsIndex.WriteByte(',')
			event.fieldsIndex.WriteString(strconv.Itoa(len(v)))
			isString := dataType == jsonparser.String
			var bitSet = byte(0)
			if isString {
				bitSet = byte(1)
			}
			event.fieldsIndex.WriteByte(bitSet)
			event.fieldsIndex.WriteByte('|')
		}

		return nil
	})

	return event
}

func (e *LogEvent) recycle() {
	e.rfc3339Nano.Reset()
	e.level.Reset()
	e.logger.Reset()
	e.caller.Reset()
	e.message.Reset()
	e.fields.Reset()
	e.fieldsIndex.Reset()
	eventPool.Put(e)
}
