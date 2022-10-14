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

package lork

import (
	"fmt"
	"sync"
	"time"
)

var (
	classicRecordPool = &sync.Pool{
		New: func() interface{} {
			return &classicRecord{}
		},
	}
)

type classicRecord struct {
	event  *LogEvent
	writer *MultiWriter
}

func newClassicRecord(lvl Level, multiWriter *MultiWriter) Record {
	r := classicRecordPool.Get().(*classicRecord)
	r.event = NewLogEvent()
	r.writer = multiWriter
	r.event.appendLevel(lvl)

	return r
}

func (r *classicRecord) Str(key, val string) Record {
	r.event.appendString(key, val)
	return r
}

func (r *classicRecord) Strs(key string, val []string) Record {
	r.event.appendStrings(key, val)
	return r
}

func (r *classicRecord) Bytes(key string, val []byte) Record {
	r.event.appendBytes(key, val)
	return r
}

func (r *classicRecord) Err(err error) Record {
	r.event.appendString(ErrorFieldKey, err.Error())
	return r
}

func (r *classicRecord) Errs(key string, errs []error) Record {
	r.event.appendErrors(key, errs)
	return r
}

func (r *classicRecord) Bool(key string, val bool) Record {
	r.event.appendBool(key, val)
	return r
}

func (r *classicRecord) Bools(key string, val []bool) Record {
	r.event.appendBools(key, val)
	return r
}

func (r *classicRecord) Int(key string, val int) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *classicRecord) Ints(key string, val []int) Record {
	r.event.appendInts(key, val)
	return r
}

func (r *classicRecord) Int8(key string, val int8) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *classicRecord) Ints8(key string, val []int8) Record {
	r.event.appendInts8(key, val)
	return r
}

func (r *classicRecord) Int16(key string, val int16) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *classicRecord) Ints16(key string, val []int16) Record {
	r.event.appendInts16(key, val)
	return r
}

func (r *classicRecord) Int32(key string, val int32) Record {
	r.event.appendInt(key, int64(val))
	return r
}

func (r *classicRecord) Ints32(key string, val []int32) Record {
	r.event.appendInts32(key, val)
	return r
}

func (r *classicRecord) Int64(key string, val int64) Record {
	r.event.appendInt(key, val)
	return r
}

func (r *classicRecord) Ints64(key string, val []int64) Record {
	r.event.appendInts64(key, val)
	return r
}

func (r *classicRecord) Uint(key string, val uint) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *classicRecord) Uints(key string, val []uint) Record {
	r.event.appendUints(key, val)
	return r
}

func (r *classicRecord) Uint8(key string, val uint8) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *classicRecord) Uints8(key string, val []uint8) Record {
	r.event.appendUints8(key, val)
	return r
}

func (r *classicRecord) Uint16(key string, val uint16) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *classicRecord) Uints16(key string, val []uint16) Record {
	r.event.appendUints16(key, val)
	return r
}

func (r *classicRecord) Uint32(key string, val uint32) Record {
	r.event.appendUint(key, uint64(val))
	return r
}

func (r *classicRecord) Uints32(key string, val []uint32) Record {
	r.event.appendUints32(key, val)
	return r
}

func (r *classicRecord) Uint64(key string, val uint64) Record {
	r.event.appendUint(key, val)
	return r
}

func (r *classicRecord) Uints64(key string, val []uint64) Record {
	r.event.appendUints64(key, val)
	return r
}

func (r *classicRecord) Float32(key string, val float32) Record {
	r.event.appendFloat32(key, val)
	return r
}

func (r *classicRecord) Floats32(key string, val []float32) Record {
	r.event.appendFloats32(key, val)
	return r
}

func (r *classicRecord) Float64(key string, val float64) Record {
	r.event.appendFloat64(key, val)
	return r
}

func (r *classicRecord) Floats64(key string, val []float64) Record {
	r.event.appendFloats64(key, val)
	return r
}

func (r *classicRecord) Time(key string, val time.Time) Record {
	r.event.appendTime(key, val)
	return r
}

func (r *classicRecord) Times(key string, val []time.Time) Record {
	r.event.appendTimes(key, val)
	return r
}

func (r *classicRecord) Dur(key string, val time.Duration) Record {
	r.event.appendDuration(key, val)
	return r
}

func (r *classicRecord) Durs(key string, val []time.Duration) Record {
	r.event.appendDurations(key, val)
	return r
}

func (r *classicRecord) Any(key string, val interface{}) Record {
	r.event.appendAny(key, val)
	return r
}

func (r *classicRecord) Msge() {
	if err := r.writer.WriteEvent(r.event); err != nil {
		Reportf("fail to write event: %v", err)
	}

	classicRecordPool.Put(r)
}

func (r *classicRecord) Msg(msg string) {
	r.event.appendMessage(msg)
	r.Msge()
}

func (r *classicRecord) Msgf(format string, v ...interface{}) {
	r.event.appendMessage(fmt.Sprintf(format, v...))
	r.Msge()
}
