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

// Record represents a log record to hold the data to log.
type Record interface {
	// Str adds string value to this record.
	Str(key, val string) Record

	// Strs adds string array value to this record.
	Strs(key string, val []string) Record

	// Bytes adds byte array value to this record.
	Bytes(key string, val []byte) Record

	// Hex adds hex byte array value to this record.
	Hex(key string, val []byte) Record

	// Err adds err to this record.
	Err(err error) Record

	// Errs adds err array to this record.
	Errs(key string, errs []error) Record

	// Bool adds bool value to this record.
	Bool(key string, val bool) Record

	// Bools adds bool array value to this record.
	Bools(key string, val []bool) Record

	// Int adds int value to this record.
	Int(key string, val int) Record

	// Ints adds int array value to this record.
	Ints(key string, val []int) Record

	// Int8 adds int8 value to this record.
	Int8(key string, val int8) Record

	// ints8 adds int8 array value to this record.
	Ints8(key string, val []int8) Record

	// Int16 adds int16 value to this record.
	Int16(key string, val int16) Record

	// Ints16 adds int16 array value to this record.
	Ints16(key string, val []int16) Record

	// Int32 adds int32 value to this record.
	Int32(key string, val int32) Record

	// Int32 adds int32 array value to this record.
	Ints32(key string, val []int32) Record

	// Int64 adds int64 value to this record.
	Int64(key string, val int64) Record

	// Int64 adds int64 array value to this record.
	Ints64(key string, val []int64) Record

	// Uint adds uint value to this record.
	Uint(key string, val uint) Record

	// Uints adds uint array value to this record.
	Uints(key string, val []uint) Record

	// Uint8 adds uint8 value to this record.
	Uint8(key string, val uint8) Record

	// Uints8 adds uint8 array value to this record.
	Uints8(key string, val []uint8) Record

	// Uint16 adds uint16 value to this record.
	Uint16(key string, val uint16) Record

	// Uints16 adds uint16 array value to this record.
	Uints16(key string, val []uint16) Record

	// Uint32 adds uint32 value to this record.
	Uint32(key string, val uint32) Record

	// Uint32 adds uint32 array value to this record.
	Uints32(key string, val []uint32) Record

	// Uint64 adds uint64 value to this record.
	Uint64(key string, val uint64) Record

	// Uints64 adds uint64 array value to this record.
	Uints64(key string, val []uint64) Record

	// Float32 adds float32 value to this record.
	Float32(key string, val float32) Record

	// Floats32 adds float32 array value to this record.
	Floats32(key string, val []float32) Record

	// Float64 adds float64 value to this record.
	Float64(key string, val float64) Record

	// Floats64 adds float64 array value to this record.
	Floats64(key string, val []float64) Record

	// Time adds time value to this record.
	Time(key string, val time.Time) Record

	// Times adds time array value to this record.
	Times(key string, val []time.Time) Record

	// Dur adds duration value to this record.
	Dur(key string, val time.Duration) Record

	// Time adds duration array value to this record.
	Durs(key string, val []time.Duration) Record

	// Interface adds interface value to this record.
	Interface(key string, val interface{}) Record

	// Msg adds a message to this record and output log.
	Msg(msg string)

	// Msgf adds a message with format to this record and output log.
	Msgf(format string, msg string)
}
