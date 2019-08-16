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
	"time"
)

type Record interface {
	Str(key, val string) Record

	Strs(key string, val []string) Record

	Bytes(key string, val []byte) Record

	Hex(key string, val []byte) Record

	Err(err error) Record

	Errs(key string, errs []error) Record

	Bool(key string, val bool) Record

	Bools(key string, val []bool) Record

	Int(key string, val int) Record

	Ints(key string, val []int) Record

	Int8(key string, val int8) Record

	Ints8(key string, val []int8) Record

	Int16(key string, val int16) Record

	Ints16(key string, val []int16) Record

	Int32(key string, val int32) Record

	Ints32(key string, val []int32) Record

	Int64(key string, val int64) Record

	Ints64(key string, val []int64) Record

	Uint(key string, val uint) Record

	Uints(key string, val []uint) Record

	Uint8(key string, val uint8) Record

	Uints8(key string, val []uint8) Record

	Uint16(key string, val uint16) Record

	Uints16(key string, val []uint16) Record

	Uint32(key string, val uint32) Record

	Uints32(key string, val []uint32) Record

	Uint64(key string, val uint64) Record

	Uints64(key string, val []uint64) Record

	Float32(key string, val float32) Record

	Floats32(key string, val []float32) Record

	Float64(key string, val float64) Record

	Floats64(key string, val []float64) Record

	Time(key string, val time.Time) Record

	Times(key string, val []time.Time) Record

	Dur(key string, val time.Duration) Record

	Durs(key string, val []time.Duration) Record

	Interface(key string, val interface{}) Record

	Msg(msg string)

	Msgf(format string, msg string)
}
