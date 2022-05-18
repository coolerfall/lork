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
	"time"

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
	appender    *bytes.Buffer
	tmp         *bytes.Buffer
}

var (
	eventPool = &sync.Pool{
		New: func() interface{} {
			tmp := new(bytes.Buffer)
			tmp.Grow(128)
			return &LogEvent{
				level:       new(bytes.Buffer),
				rfc3339Nano: new(bytes.Buffer),
				logger:      new(bytes.Buffer),
				caller:      new(bytes.Buffer),
				message:     new(bytes.Buffer),
				fields:      new(bytes.Buffer),
				fieldsIndex: new(bytes.Buffer),
				appender:    new(bytes.Buffer),
				tmp:         tmp,
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
			event.appendLevelBytes(v)
		case LoggerFieldKey:
			event.appendLogger(v)
		case MessageFieldKey:
			event.appendMessageBytes(v)

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

func (e *LogEvent) appendLevel(lvl Level) {
	e.tmp.WriteString(lvl.String())
	data := e.tmp.Bytes()
	e.tmp.Reset()
	e.appendLevelBytes(data)
}

func (e *LogEvent) appendLevelBytes(v []byte) {
	e.level.Write(v)
}

func (e *LogEvent) appendLogger(v []byte) {
	e.logger.Write(v)
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

func (e *LogEvent) appendTimestamp() {
	data := e.tmp.Bytes()
	data, err := appendFormat(data, time.Now(), TimestampFormat)
	if err != nil {
		return
	}
	e.makeTimestamp(data)
}

func (e *LogEvent) appendMessageBytes(msg []byte) {
	e.message.Grow(len(msg))
	temp := e.message.Bytes()
	m, _ := jsonparser.Unescape(msg, temp)
	e.message.Write(bytes.TrimRight(m, "\n"))
}

func (e *LogEvent) appendMessage(msg string) {
	e.message.WriteString(msg)
	v := e.message.Bytes()
	e.message.Reset()
	e.appendMessageBytes(v)
}

func (e *LogEvent) appendKeyValue(key string, value []byte, isString bool) {
	e.fields.WriteString(key)
	e.fieldsIndex.WriteString(strconv.Itoa(len(key)))
	e.fieldsIndex.WriteByte(',')
	e.fields.Write(value)
	e.fieldsIndex.WriteString(strconv.Itoa(len(value)))
	var bitSet = byte(0)
	if isString {
		bitSet = byte(1)
	}
	e.fieldsIndex.WriteByte(bitSet)
	e.fieldsIndex.WriteByte('|')
}

func (e *LogEvent) appendString(key, value string) {
	if key == LoggerFieldKey {
		e.logger.WriteString(value)
		return
	}
	e.appender.WriteString(value)
	data := e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, true)
}

func (e *LogEvent) appendStrings(key string, value []string) {
	e.appender.WriteString("[")
	for _, v := range value {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		e.appender.WriteString(`"`)
		e.appender.WriteString(v)
		e.appender.WriteString(`"`)
	}
	e.appender.WriteString("]")
	data := e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendBytes(key string, value []byte) {
	e.appender.WriteString("[")
	for _, v := range value {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		e.appender.WriteString(strconv.Itoa(int(v)))
	}
	e.appender.WriteString("]")
	data := e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendBool(key string, value bool) {
	data := e.tmp.Bytes()
	data = strconv.AppendBool(data, value)
	e.tmp.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendBools(key string, value []bool) {
	data := e.tmp.Bytes()
	e.appender.WriteString("[")
	for _, v := range value {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		data = strconv.AppendBool(data[:0], v)
		e.appender.Write(data)
	}
	e.appender.WriteString("]")
	data = e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendInt(key string, value int64) {
	data := e.tmp.Bytes()
	data = strconv.AppendInt(data, value, 10)
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendInts(key string, value []int) {
	data := e.tmp.Bytes()
	e.appender.WriteString("[")
	for _, v := range value {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		data = strconv.AppendInt(data[:0], int64(v), 10)
		e.appender.Write(data)
	}
	e.appender.WriteString("]")
	data = e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendInts8(key string, value []int8) {
	data := e.tmp.Bytes()
	e.appender.WriteString("[")
	for _, v := range value {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		data = strconv.AppendInt(data[:0], int64(v), 10)
		e.appender.Write(data)
	}
	e.appender.WriteString("]")
	data = e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendUint(key string, value uint64) {
	data := e.tmp.Bytes()
	data = strconv.AppendUint(data, value, 10)
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendField(key string, val interface{}) {
	switch value := val.(type) {
	case string:
		e.appendString(key, value)
	case []string:
		e.appendStrings(key, value)
	case bool:
		e.appendBool(key, value)
	case []bool:
		e.appendBools(key, value)
	case int:
		e.appendInt(key, int64(value))
	case []int:
		e.appendInts(key, value)
	case int8:
		e.appendInt(key, int64(value))
	case []int8:
		e.appendInts8(key, value)
	case int32:
		e.appendInt(key, int64(value))
	case int64:
		e.appendInt(key, value)
	case uint:
		e.appendUint(key, uint64(value))
	case uint8:
		e.appendUint(key, uint64(value))
	case uint16:
		e.appendUint(key, uint64(value))
	case uint32:
		e.appendUint(key, uint64(value))
	case uint64:
		e.appendUint(key, value)
	default:
		return
	}
}

func (e *LogEvent) Recycle() {
	e.rfc3339Nano.Reset()
	e.level.Reset()
	e.logger.Reset()
	e.caller.Reset()
	e.message.Reset()
	e.fields.Reset()
	e.fieldsIndex.Reset()
	e.appender.Reset()
	e.tmp.Reset()
	eventPool.Put(e)
}
