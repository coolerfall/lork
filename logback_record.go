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
	"fmt"
	"sync"
	"time"
)

var (
	logbackRecordPool = &sync.Pool{
		New: func() interface{} {
			return &logbackRecord{}
		},
	}
)

type logbackRecord struct {
	event       *LogEvent
	multiWriter *MultiWriter
}

func newLogbackRecord(lvl Level, multiWriter *MultiWriter) Record {
	r := logbackRecordPool.Get().(*logbackRecord)
	r.event = NewLogEvent()
	r.multiWriter = multiWriter
	r.event.appendLevel(lvl)

	return r
}

func (r *logbackRecord) Str(key, val string) Record {
	r.event.appendString(key, val)
	return r
}

func (r *logbackRecord) Strs(key string, val []string) Record {
	r.event.appendStrings(key, val)
	return r
}

func (r *logbackRecord) Bytes(key string, val []byte) Record {
	r.event.appendBytes(key, val)
	return r
}

func (r *logbackRecord) Hex(key string, val []byte) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Err(err error) Record {
	r.event.appendString(ErrorFieldKey, err.Error())
	return r
}

func (r *logbackRecord) Errs(key string, errs []error) Record {
	r.event.appendField(key, errs)
	return r
}

func (r *logbackRecord) Bool(key string, val bool) Record {
	r.event.appendBool(key, val)
	return r
}

func (r *logbackRecord) Bools(key string, val []bool) Record {
	r.event.appendBools(key, val)
	return r
}

func (r *logbackRecord) Int(key string, val int) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *logbackRecord) Ints(key string, val []int) Record {
	r.event.appendInts(key, val)
	return r
}

func (r *logbackRecord) Int8(key string, val int8) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *logbackRecord) Ints8(key string, val []int8) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Int16(key string, val int16) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *logbackRecord) Ints16(key string, val []int16) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Int32(key string, val int32) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *logbackRecord) Ints32(key string, val []int32) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Int64(key string, val int64) Record {
	r.event.appendInt(key, val)
	return r
}

func (r *logbackRecord) Ints64(key string, val []int64) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Uint(key string, val uint) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *logbackRecord) Uints(key string, val []uint) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Uint8(key string, val uint8) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *logbackRecord) Uints8(key string, val []uint8) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Uint16(key string, val uint16) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *logbackRecord) Uints16(key string, val []uint16) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Uint32(key string, val uint32) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *logbackRecord) Uints32(key string, val []uint32) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Uint64(key string, val uint64) Record {
	r.event.appendUint(key, val)
	return r
}

func (r *logbackRecord) Uints64(key string, val []uint64) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Float32(key string, val float32) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Floats32(key string, val []float32) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Float64(key string, val float64) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Floats64(key string, val []float64) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Time(key string, val time.Time) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Times(key string, val []time.Time) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Dur(key string, val time.Duration) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Durs(key string, val []time.Duration) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Interface(key string, val interface{}) Record {
	r.event.appendField(key, val)
	return r
}

func (r *logbackRecord) Msge() {
	r.event.appendTimestamp()
	if _, err := r.multiWriter.WriteEvent(r.event); err != nil {
		Reportf("fail to write event: %v", err)
	}

	logbackRecordPool.Put(r)
}

func (r *logbackRecord) Msg(msg string) {
	r.event.appendTimestamp()
	r.event.appendMessage(msg)
	if _, err := r.multiWriter.WriteEvent(r.event); err != nil {
		Reportf("fail to write event: %v", err)
	}
	logbackRecordPool.Put(r)
}

func (r *logbackRecord) Msgf(format string, v ...interface{}) {
	r.event.appendTimestamp()
	r.event.appendMessage(fmt.Sprintf(format, v...))
	if _, err := r.multiWriter.WriteEvent(r.event); err != nil {
		Reportf("fail to write event: %v", err)
	}

	logbackRecordPool.Put(r)
}
