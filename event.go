// Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
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

// NewLogEvent gets a LogEvent from pool.
func NewLogEvent() *LogEvent {
	return eventPool.Get().(*LogEvent)
}

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
func (e *LogEvent) Fields(callback func(k, v []byte, isString bool) error) error {
	var startIndex = 0
	var kvIndex = 0
	var ik, iv int
	kvArr := e.fields.Bytes()
	indexArr := e.fieldsIndex.Bytes()
	for i := 0; i < len(indexArr); i++ {
		if indexArr[i] == ',' {
			ik, _ = atoi(indexArr[startIndex:i])
			startIndex = i + 1
		} else if indexArr[i] == '|' {
			iv, _ = atoi(indexArr[startIndex : i-1])
			startIndex = i + 1
			bitSet := indexArr[i-1]
			var isString bool
			if bitSet == byte(1) {
				isString = true
			}
			keyEndIndex := kvIndex + ik
			valueEndIndex := kvIndex + ik + iv
			err := callback(kvArr[kvIndex:keyEndIndex],
				kvArr[keyEndIndex:valueEndIndex], isString)
			if err != nil {
				return err
			}
			kvIndex = valueEndIndex
		}
	}

	return nil
}

func MakeEvent(p []byte) *LogEvent {
	event := eventPool.Get().(*LogEvent)
	_ = jsonparser.ObjectEach(p, func(k []byte, v []byte,
		dataType jsonparser.ValueType, _ int) error {
		switch string(k) {
		case TimestampFieldKey:
			event.makeTimestamp(v)
		case LevelFieldKey:
			event.makeLevel(v)
		case LoggerFieldKey:
			event.makeLogger(v)
		case MessageFieldKey:
			event.makeMessage(v)

		default:
			event.makeFileds(k, v, dataType == jsonparser.String)
		}

		return nil
	})

	return event
}

func (e *LogEvent) makeTimestamp(v []byte) {
	e.rfc3339Nano.Write(v)
}

func (e *LogEvent) makeLevel(v []byte) {
	e.level.Write(v)
}

func (e *LogEvent) makeLogger(v []byte) {
	e.logger.Write(v)
}

func (e *LogEvent) makeMessage(v []byte) {
	e.message.Grow(len(v))
	temp := e.message.Bytes()
	m, _ := jsonparser.Unescape(v, temp)
	e.message.Write(bytes.TrimRight(m, "\n"))
}

func (e *LogEvent) makeFileds(k, v []byte, isString bool) {
	e.fields.Write(k)
	e.fields.Write(v)
	e.fieldsIndex.WriteString(strconv.Itoa(len(k)))
	e.fieldsIndex.WriteByte(',')
	e.fieldsIndex.WriteString(strconv.Itoa(len(v)))
	var bitSet = byte(0)
	if isString {
		bitSet = byte(1)
	}
	e.fieldsIndex.WriteByte(bitSet)
	e.fieldsIndex.WriteByte('|')
}

func (e *LogEvent) Recycle() {
	e.rfc3339Nano.Reset()
	e.level.Reset()
	e.logger.Reset()
	e.caller.Reset()
	e.message.Reset()
	e.fields.Reset()
	e.fieldsIndex.Reset()
	eventPool.Put(e)
}
