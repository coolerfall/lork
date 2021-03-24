// Copyright (c) 2019-2020 Vincent Cheung (coolingfall@gmail.com).
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

package slazero

import (
	"sync"
	"time"

	"github.com/coolerfall/slago"
	"github.com/rs/zerolog"
)

var (
	recordPool = &sync.Pool{
		New: func() interface{} {
			return &zeroRecord{}
		},
	}
)

type zeroRecord struct {
	event *zerolog.Event
}

func newZeroRecord(e *zerolog.Event) *zeroRecord {
	r := recordPool.Get().(*zeroRecord)
	r.event = e
	return r
}

func (r *zeroRecord) Str(key, val string) slago.Record {
	r.event.Str(key, val)
	return r
}

func (r *zeroRecord) Strs(key string, val []string) slago.Record {
	r.event.Strs(key, val)
	return r
}

func (r *zeroRecord) Bytes(key string, val []byte) slago.Record {
	r.event.Bytes(key, val)
	return r
}

func (r *zeroRecord) Hex(key string, val []byte) slago.Record {
	r.event.Hex(key, val)
	return r
}

func (r *zeroRecord) Err(err error) slago.Record {
	r.event.Err(err)
	return r
}

func (r *zeroRecord) Errs(key string, errs []error) slago.Record {
	r.event.Errs(key, errs)
	return r
}

func (r *zeroRecord) Bool(key string, val bool) slago.Record {
	r.event.Bool(key, val)
	return r
}

func (r *zeroRecord) Bools(key string, val []bool) slago.Record {
	r.event.Bools(key, val)
	return r
}

func (r *zeroRecord) Int(key string, val int) slago.Record {
	r.event.Int(key, val)
	return r
}

func (r *zeroRecord) Ints(key string, val []int) slago.Record {
	r.event.Ints(key, val)
	return r
}

func (r *zeroRecord) Int8(key string, val int8) slago.Record {
	r.event.Int8(key, val)
	return r
}

func (r *zeroRecord) Ints8(key string, val []int8) slago.Record {
	r.event.Ints8(key, val)
	return r
}

func (r *zeroRecord) Int16(key string, val int16) slago.Record {
	r.event.Int16(key, val)
	return r
}

func (r *zeroRecord) Ints16(key string, val []int16) slago.Record {
	r.event.Ints16(key, val)
	return r
}

func (r *zeroRecord) Int32(key string, val int32) slago.Record {
	r.event.Int32(key, val)
	return r
}

func (r *zeroRecord) Ints32(key string, val []int32) slago.Record {
	r.event.Ints32(key, val)
	return r
}

func (r *zeroRecord) Int64(key string, val int64) slago.Record {
	r.event.Int64(key, val)
	return r
}

func (r *zeroRecord) Ints64(key string, val []int64) slago.Record {
	r.event.Ints64(key, val)
	return r
}

func (r *zeroRecord) Uint(key string, val uint) slago.Record {
	r.event.Uint(key, val)
	return r
}

func (r *zeroRecord) Uints(key string, val []uint) slago.Record {
	r.event.Uints(key, val)
	return r
}

func (r *zeroRecord) Uint8(key string, val uint8) slago.Record {
	r.event.Uint8(key, val)
	return r
}

func (r *zeroRecord) Uints8(key string, val []uint8) slago.Record {
	r.event.Uints8(key, val)
	return r
}

func (r *zeroRecord) Uint16(key string, val uint16) slago.Record {
	r.event.Uint16(key, val)
	return r
}

func (r *zeroRecord) Uints16(key string, val []uint16) slago.Record {
	r.event.Uints16(key, val)
	return r
}

func (r *zeroRecord) Uint32(key string, val uint32) slago.Record {
	r.event.Uint32(key, val)
	return r
}

func (r *zeroRecord) Uints32(key string, val []uint32) slago.Record {
	r.event.Uints32(key, val)
	return r
}

func (r *zeroRecord) Uint64(key string, val uint64) slago.Record {
	r.event.Uint64(key, val)
	return r
}

func (r *zeroRecord) Uints64(key string, val []uint64) slago.Record {
	r.event.Uints64(key, val)
	return r
}

func (r *zeroRecord) Float32(key string, val float32) slago.Record {
	r.event.Float32(key, val)
	return r
}

func (r *zeroRecord) Floats32(key string, val []float32) slago.Record {
	r.event.Floats32(key, val)
	return r
}

func (r *zeroRecord) Float64(key string, val float64) slago.Record {
	r.event.Float64(key, val)
	return r
}

func (r *zeroRecord) Floats64(key string, val []float64) slago.Record {
	r.event.Floats64(key, val)
	return r
}

func (r *zeroRecord) Time(key string, val time.Time) slago.Record {
	r.event.Time(key, val)
	return r
}

func (r *zeroRecord) Times(key string, val []time.Time) slago.Record {
	r.event.Times(key, val)
	return r
}

func (r *zeroRecord) Dur(key string, val time.Duration) slago.Record {
	r.event.Dur(key, val)
	return r
}

func (r *zeroRecord) Durs(key string, val []time.Duration) slago.Record {
	r.event.Durs(key, val)
	return r
}

func (r *zeroRecord) Interface(key string, val interface{}) slago.Record {
	r.event.Interface(key, val)
	return r
}

func (r *zeroRecord) Msg(originMsg ...string) {
	var msg string
	if len(originMsg) != 0 {
		msg = originMsg[0]
	}
	r.event.Msg(msg)
	recordPool.Put(r)
}

func (r *zeroRecord) Msgf(format string, v ...interface{}) {
	r.event.Msgf(format, v...)
	recordPool.Put(r)
}
