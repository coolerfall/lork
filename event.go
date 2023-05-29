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
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/buger/jsonparser"
)

type LogEvent struct {
	unixNano    int64
	level       *bytes.Buffer
	loggerName  *bytes.Buffer
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
				loggerName:  new(bytes.Buffer),
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
	event := eventPool.Get().(*LogEvent)

	return event
}

// MakeEvent makes LogEvent from json string. The LevelFieldKey, TimestampFieldKey and
// MessageFieldKey field key must keep the same with lork.
func MakeEvent(p []byte) *LogEvent {
	event := eventPool.Get().(*LogEvent)
	_ = jsonparser.ObjectEach(p, func(k []byte, v []byte,
		dataType jsonparser.ValueType, _ int) error {
		switch string(k) {
		case TimestampFieldKey:
			event.appendRFC3999Nano(v)
		case LevelFieldKey:
			event.appendLevelBytes(v)
		case LoggerNameFieldKey:
			event.appendLogger(v)
		case MessageFieldKey:
			event.appendMessageBytes(v)

		default:
			event.makeFields(k, v, dataType == jsonparser.String)
		}

		return nil
	})

	return event
}

func (e *LogEvent) Copy() *LogEvent {
	cp := eventPool.Get().(*LogEvent)
	cp.unixNano = e.unixNano
	cp.level.Write(e.level.Bytes())
	cp.loggerName.Write(e.loggerName.Bytes())
	cp.caller.Write(e.caller.Bytes())
	cp.message.Write(e.message.Bytes())
	cp.fields.Write(e.fields.Bytes())
	cp.fieldsIndex.Write(e.fieldsIndex.Bytes())

	return cp
}

// Timestamp returns unix timestamp in nano second.
func (e *LogEvent) Timestamp() int64 {
	if e.unixNano == 0 {
		e.appendTimestamp()
	}
	return e.unixNano
}

// LevelInt returns level int value.
func (e *LogEvent) LevelInt() Level {
	return ParseLevel(e.level.String())
}

// Level returns level string bytes.
func (e *LogEvent) Level() []byte {
	return e.level.Bytes()
}

// LoggerName return logger name bytes.
func (e *LogEvent) LoggerName() []byte {
	return e.loggerName.Bytes()
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

func (e *LogEvent) appendLevel(lvl Level) {
	e.tmp.WriteString(lvl.String())
	data := e.tmp.Bytes()
	e.tmp.Reset()
	e.appendLevelBytes(data)
}

func (e *LogEvent) appendLevelBytes(v []byte) {
	e.level.Reset()
	e.level.Write(v)
}

func (e *LogEvent) appendLogger(v []byte) {
	e.loggerName.Reset()
	e.loggerName.Write(v)
}

func (e *LogEvent) makeFields(k, v []byte, isString bool) {
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
	e.unixNano = time.Now().UnixNano()
}

func (e *LogEvent) appendRFC3999Nano(rfc3339Nano []byte) {
	nano, err := toUTCUnixNano(rfc3339Nano, TimestampFormat)
	if err != nil {
		return
	}
	e.unixNano = nano
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

func (e *LogEvent) appendArray(key string, len int, f func(data []byte, index int) ([]byte, bool)) {
	data := e.tmp.Bytes()
	e.appender.WriteString("[")
	for i := 0; i < len; i++ {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		d, isString := f(data[:0], i)
		if isString {
			e.appender.WriteByte('"')
		}
		e.appender.Write(d)
		if isString {
			e.appender.WriteByte('"')
		}
	}
	e.appender.WriteString("]")
	data = e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendString(key, value string) {
	if key == LoggerNameFieldKey {
		e.loggerName.WriteString(value)
		return
	}
	e.appender.WriteString(value)
	data := e.appender.Bytes()
	e.appender.Reset()
	e.appendKeyValue(key, data, true)
}

func (e *LogEvent) appendStrings(key string, value []string) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		e.tmp.WriteString(value[index])
		d := e.tmp.Bytes()
		e.tmp.Reset()
		return d, true
	})
}

func (e *LogEvent) appendBytes(key string, value []byte) {
	e.appender.WriteString("[")
	for _, v := range value {
		if e.appender.Len() > 1 {
			e.appender.WriteString(",")
		}
		data := e.tmp.Bytes()
		data = strconv.AppendInt(data, int64(v), 10)
		e.tmp.Reset()
		e.appender.Write(data)
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
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendBool(data, value[index]), false
	})
}

func (e *LogEvent) appendInt(key string, value int64) {
	data := e.tmp.Bytes()
	data = strconv.AppendInt(data, value, 10)
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendInts(key string, value []int) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendInt(data, int64(value[index]), 10), false
	})
}

func (e *LogEvent) appendInts8(key string, value []int8) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendInt(data, int64(value[index]), 10), false
	})
}

func (e *LogEvent) appendInts16(key string, value []int16) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendInt(data, int64(value[index]), 10), false
	})
}

func (e *LogEvent) appendInts32(key string, value []int32) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendInt(data, int64(value[index]), 10), false
	})
}

func (e *LogEvent) appendInts64(key string, value []int64) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendInt(data, value[index], 10), false
	})
}

func (e *LogEvent) appendUint(key string, value uint64) {
	data := e.tmp.Bytes()
	data = strconv.AppendUint(data, value, 10)
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendUints(key string, value []uint) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendUint(data, uint64(value[index]), 10), false
	})
}

func (e *LogEvent) appendUints8(key string, value []uint8) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendUint(data, uint64(value[index]), 10), false
	})
}

func (e *LogEvent) appendUints16(key string, value []uint16) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendUint(data, uint64(value[index]), 10), false
	})
}

func (e *LogEvent) appendUints32(key string, value []uint32) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendUint(data, uint64(value[index]), 10), false
	})
}

func (e *LogEvent) appendUints64(key string, value []uint64) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendUint(data, value[index], 10), false
	})
}

func (e *LogEvent) appendFloat(dst []byte, value float64, bitSize int) []byte {
	switch {
	case math.IsNaN(value):
		return append(dst, `"NaN"`...)
	case math.IsInf(value, 1):
		return append(dst, `"+Inf"`...)
	case math.IsInf(value, -1):
		return append(dst, `"-Inf"`...)
	default:
		return strconv.AppendFloat(dst, value, 'f', -1, bitSize)
	}
}

func (e *LogEvent) appendFloat32(key string, value float32) {
	data := e.tmp.Bytes()
	data = e.appendFloat(data, float64(value), 32)
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendFloats32(key string, value []float32) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return e.appendFloat(data, float64(value[index]), 32), false
	})
}

func (e *LogEvent) appendFloat64(key string, value float64) {
	data := e.tmp.Bytes()
	data = e.appendFloat(data, value, 64)
	e.appendKeyValue(key, data, false)
}

func (e *LogEvent) appendFloats64(key string, value []float64) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return e.appendFloat(data, value[index], 64), false
	})
}

func (e *LogEvent) appendTime(key string, value time.Time) {
	data := e.tmp.Bytes()
	data, err := appendFormat(data, value, TimeFormatRFC3339)
	if err != nil {
		return
	}
	e.appendKeyValue(key, data, true)
}

func (e *LogEvent) appendTimes(key string, value []time.Time) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		d, err := appendFormat(data, value[index], TimeFormatRFC3339)
		if err != nil {
			return data, true
		}
		return d, true
	})
}

func (e *LogEvent) appendDuration(key string, value time.Duration) {
	e.appendInt(key, value.Nanoseconds())
}

func (e *LogEvent) appendDurations(key string, value []time.Duration) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		return strconv.AppendInt(data, value[index].Nanoseconds(), 10), false
	})
}

func (e *LogEvent) appendErrors(key string, value []error) {
	e.appendArray(key, len(value), func(data []byte, index int) ([]byte, bool) {
		e.tmp.WriteString(value[index].Error())
		d := e.tmp.Bytes()
		e.tmp.Reset()
		return d, true
	})
}

func (e *LogEvent) appendAny(key string, val interface{}) {
	switch val.(type) {
	case string:
		e.appendString(key, val.(string))
	case []string:
		e.appendStrings(key, val.([]string))
	case []byte:
		e.appendBytes(key, val.([]byte))
	case []error:
		e.appendErrors(key, val.([]error))
	case bool:
		e.appendBool(key, val.(bool))
	case []bool:
		e.appendBools(key, val.([]bool))
	case int:
		e.appendInt(key, int64(val.(int)))
	case int8:
		e.appendInt(key, int64(val.(int8)))
	case int16:
		e.appendInt(key, int64(val.(int16)))
	case int32:
		e.appendInt(key, int64(val.(int32)))
	case int64:
		e.appendInt(key, val.(int64))
	case []int:
		e.appendInts(key, val.([]int))
	case []int8:
		e.appendInts8(key, val.([]int8))
	case []int16:
		e.appendInts16(key, val.([]int16))
	case []int32:
		e.appendInts32(key, val.([]int32))
	case []int64:
		e.appendInts64(key, val.([]int64))
	case uint:
		e.appendUint(key, uint64(val.(uint)))
	case uint8:
		e.appendUint(key, uint64(val.(uint8)))
	case uint16:
		e.appendUint(key, uint64(val.(uint16)))
	case uint32:
		e.appendUint(key, uint64(val.(uint32)))
	case uint64:
		e.appendUint(key, val.(uint64))
	case []uint:
		e.appendUints(key, val.([]uint))
	case []uint16:
		e.appendUints16(key, val.([]uint16))
	case []uint32:
		e.appendUints32(key, val.([]uint32))
	case []uint64:
		e.appendUints64(key, val.([]uint64))
	case float32:
		e.appendFloat32(key, val.(float32))
	case float64:
		e.appendFloat64(key, val.(float64))
	case time.Time:
		e.appendTime(key, val.(time.Time))
	case []time.Time:
		e.appendTimes(key, val.([]time.Time))
	case time.Duration:
		e.appendDuration(key, val.(time.Duration))
	case []time.Duration:
		e.appendDurations(key, val.([]time.Duration))
	default:
		data, err := json.Marshal(val)
		if err != nil {
			e.appendString(key, fmt.Sprintf("marshaling error: %v", err))
			return
		}
		e.appendKeyValue(key, data, false)
	}
}

func (e *LogEvent) Recycle() {
	e.unixNano = 0
	e.level.Reset()
	e.loggerName.Reset()
	e.caller.Reset()
	e.message.Reset()
	e.fields.Reset()
	e.fieldsIndex.Reset()
	e.appender.Reset()
	e.tmp.Reset()
	eventPool.Put(e)
}
