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

package logrus

import (
	"sync"
	"time"

	"github.com/coolerfall/lork"
	"github.com/sirupsen/logrus"
)

var (
	logrusRecordPool = &sync.Pool{
		New: func() interface{} {
			return &logrusRecord{}
		},
	}
)

type logrusRecord struct {
	entry *logrus.Entry
	level logrus.Level
}

func newLogrusRecord(lvl logrus.Level) *logrusRecord {
	r := logrusRecordPool.Get().(*logrusRecord)
	r.entry = logrus.NewEntry(logrus.StandardLogger())
	r.level = lvl

	return r
}

func (r *logrusRecord) Str(key, val string) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Strs(key string, val []string) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Bytes(key string, val []byte) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Err(err error) lork.Record {
	r.entry = r.entry.WithError(err)
	return r
}

func (r *logrusRecord) Errs(key string, errs []error) lork.Record {
	r.entry = r.entry.WithField(key, errs)
	return r
}

func (r *logrusRecord) Bool(key string, val bool) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Bools(key string, val []bool) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Int(key string, val int) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Ints(key string, val []int) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Int8(key string, val int8) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Ints8(key string, val []int8) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Int16(key string, val int16) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Ints16(key string, val []int16) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Int32(key string, val int32) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Ints32(key string, val []int32) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Int64(key string, val int64) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Ints64(key string, val []int64) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uint(key string, val uint) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uints(key string, val []uint) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uint8(key string, val uint8) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uints8(key string, val []uint8) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uint16(key string, val uint16) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uints16(key string, val []uint16) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uint32(key string, val uint32) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uints32(key string, val []uint32) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uint64(key string, val uint64) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Uints64(key string, val []uint64) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Float32(key string, val float32) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Floats32(key string, val []float32) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Float64(key string, val float64) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Floats64(key string, val []float64) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Time(key string, val time.Time) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Times(key string, val []time.Time) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Dur(key string, val time.Duration) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Durs(key string, val []time.Duration) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Any(key string, val interface{}) lork.Record {
	r.entry = r.entry.WithField(key, val)
	return r
}

func (r *logrusRecord) Msge() {
	r.Msg("")
}

func (r *logrusRecord) Msg(msg string) {
	r.entry.Log(r.level, msg)
	logrusRecordPool.Put(r)
}

func (r *logrusRecord) Msgf(format string, v ...interface{}) {
	r.entry.Logf(r.level, format, v...)
	logrusRecordPool.Put(r)
}
