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
	"sync"
	"time"
)

var (
	recordPool = &sync.Pool{
		New: func() interface{} {
			return &noopRecord{}
		},
	}
)

type noopRecord struct {
}

func newNoopRecord() *noopRecord {
	return recordPool.Get().(*noopRecord)
}

func (r *noopRecord) Str(key, val string) Record {
	return r
}

func (r *noopRecord) Strs(key string, val []string) Record {
	return r
}

func (r *noopRecord) Bytes(key string, val []byte) Record {
	return r
}

func (r *noopRecord) Hex(key string, val []byte) Record {
	return r
}

func (r *noopRecord) Err(err error) Record {
	return r
}

func (r *noopRecord) Errs(key string, errs []error) Record {
	return r
}

func (r *noopRecord) Bool(key string, val bool) Record {
	return r
}

func (r *noopRecord) Bools(key string, val []bool) Record {
	return r
}

func (r *noopRecord) Int(key string, val int) Record {
	return r
}

func (r *noopRecord) Ints(key string, val []int) Record {
	return r
}

func (r *noopRecord) Int8(key string, val int8) Record {
	return r
}

func (r *noopRecord) Ints8(key string, val []int8) Record {
	return r
}

func (r *noopRecord) Int16(key string, val int16) Record {
	return r
}

func (r *noopRecord) Ints16(key string, val []int16) Record {
	return r
}

func (r *noopRecord) Int32(key string, val int32) Record {
	return r
}

func (r *noopRecord) Ints32(key string, val []int32) Record {
	return r
}

func (r *noopRecord) Int64(key string, val int64) Record {
	return r
}

func (r *noopRecord) Ints64(key string, val []int64) Record {
	return r
}

func (r *noopRecord) Uint(key string, val uint) Record {
	return r
}

func (r *noopRecord) Uints(key string, val []uint) Record {
	return r
}

func (r *noopRecord) Uint8(key string, val uint8) Record {
	return r
}

func (r *noopRecord) Uints8(key string, val []uint8) Record {
	return r
}

func (r *noopRecord) Uint16(key string, val uint16) Record {
	return r
}

func (r *noopRecord) Uints16(key string, val []uint16) Record {
	return r
}

func (r *noopRecord) Uint32(key string, val uint32) Record {
	return r
}

func (r *noopRecord) Uints32(key string, val []uint32) Record {
	return r
}

func (r *noopRecord) Uint64(key string, val uint64) Record {
	return r
}

func (r *noopRecord) Uints64(key string, val []uint64) Record {
	return r
}

func (r *noopRecord) Float32(key string, val float32) Record {
	return r
}

func (r *noopRecord) Floats32(key string, val []float32) Record {
	return r
}

func (r *noopRecord) Float64(key string, val float64) Record {
	return r
}

func (r *noopRecord) Floats64(key string, val []float64) Record {
	return r
}

func (r *noopRecord) Time(key string, val time.Time) Record {
	return r
}

func (r *noopRecord) Times(key string, val []time.Time) Record {
	return r
}

func (r *noopRecord) Dur(key string, val time.Duration) Record {
	return r
}

func (r *noopRecord) Durs(key string, val []time.Duration) Record {
	return r
}

func (r *noopRecord) Interface(key string, val interface{}) Record {
	return r
}

func (r *noopRecord) Msg(msg string) {
	r.Msgf(msg)
}

func (r *noopRecord) Msgf(format string, v ...interface{}) {
	recordPool.Put(r)
}
