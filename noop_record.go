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

func (r *noopRecord) Str(_, _ string) Record {
	return r
}

func (r *noopRecord) Strs(_ string, _ []string) Record {
	return r
}

func (r *noopRecord) Bytes(_ string, _ []byte) Record {
	return r
}

func (r *noopRecord) Hex(_ string, _ []byte) Record {
	return r
}

func (r *noopRecord) Err(_ error) Record {
	return r
}

func (r *noopRecord) Errs(_ string, _ []error) Record {
	return r
}

func (r *noopRecord) Bool(_ string, _ bool) Record {
	return r
}

func (r *noopRecord) Bools(_ string, _ []bool) Record {
	return r
}

func (r *noopRecord) Int(_ string, _ int) Record {
	return r
}

func (r *noopRecord) Ints(_ string, _ []int) Record {
	return r
}

func (r *noopRecord) Int8(_ string, _ int8) Record {
	return r
}

func (r *noopRecord) Ints8(_ string, _ []int8) Record {
	return r
}

func (r *noopRecord) Int16(_ string, _ int16) Record {
	return r
}

func (r *noopRecord) Ints16(_ string, _ []int16) Record {
	return r
}

func (r *noopRecord) Int32(_ string, _ int32) Record {
	return r
}

func (r *noopRecord) Ints32(_ string, _ []int32) Record {
	return r
}

func (r *noopRecord) Int64(_ string, _ int64) Record {
	return r
}

func (r *noopRecord) Ints64(_ string, _ []int64) Record {
	return r
}

func (r *noopRecord) Uint(_ string, _ uint) Record {
	return r
}

func (r *noopRecord) Uints(_ string, _ []uint) Record {
	return r
}

func (r *noopRecord) Uint8(_ string, _ uint8) Record {
	return r
}

func (r *noopRecord) Uints8(_ string, _ []uint8) Record {
	return r
}

func (r *noopRecord) Uint16(_ string, _ uint16) Record {
	return r
}

func (r *noopRecord) Uints16(_ string, _ []uint16) Record {
	return r
}

func (r *noopRecord) Uint32(_ string, _ uint32) Record {
	return r
}

func (r *noopRecord) Uints32(_ string, _ []uint32) Record {
	return r
}

func (r *noopRecord) Uint64(_ string, _ uint64) Record {
	return r
}

func (r *noopRecord) Uints64(_ string, _ []uint64) Record {
	return r
}

func (r *noopRecord) Float32(_ string, _ float32) Record {
	return r
}

func (r *noopRecord) Floats32(_ string, _ []float32) Record {
	return r
}

func (r *noopRecord) Float64(_ string, _ float64) Record {
	return r
}

func (r *noopRecord) Floats64(_ string, _ []float64) Record {
	return r
}

func (r *noopRecord) Time(_ string, _ time.Time) Record {
	return r
}

func (r *noopRecord) Times(_ string, _ []time.Time) Record {
	return r
}

func (r *noopRecord) Dur(_ string, _ time.Duration) Record {
	return r
}

func (r *noopRecord) Durs(_ string, _ []time.Duration) Record {
	return r
}

func (r *noopRecord) Interface(_ string, _ interface{}) Record {
	return r
}

func (r *noopRecord) Msg(_ ...string) {
	recordPool.Put(r)
}

func (r *noopRecord) Msgf(_ string, _ ...interface{}) {
	recordPool.Put(r)
}
