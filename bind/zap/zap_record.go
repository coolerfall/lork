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

package zap

import (
	"sync"
	"time"

	"github.com/coolerfall/lork"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapRecordPool = &sync.Pool{
		New: func() interface{} {
			return &zapRecord{}
		},
	}
)

type zapRecord struct {
	logger *zap.Logger
	level  zapcore.Level
}

func newZapRecord(lvl zapcore.Level) *zapRecord {
	r := zapRecordPool.Get().(*zapRecord)
	r.logger = zap.L()
	r.level = lvl

	return r
}

func (r *zapRecord) Str(key, val string) lork.Record {
	r.logger = r.logger.With(zap.String(key, val))
	return r
}

func (r *zapRecord) Strs(key string, val []string) lork.Record {
	r.logger = r.logger.With(zap.Strings(key, val))
	return r
}

func (r *zapRecord) Bytes(key string, val []byte) lork.Record {
	r.logger = r.logger.With(zap.ByteString(key, val))
	return r
}

func (r *zapRecord) Err(err error) lork.Record {
	r.logger = r.logger.With(zap.Error(err))
	return r
}

func (r *zapRecord) Errs(key string, errs []error) lork.Record {
	r.logger = r.logger.With(zap.Errors(key, errs))
	return r
}

func (r *zapRecord) Bool(key string, b bool) lork.Record {
	r.logger = r.logger.With(zap.Bool(key, b))
	return r
}

func (r *zapRecord) Bools(key string, b []bool) lork.Record {
	r.logger = r.logger.With(zap.Bools(key, b))
	return r
}

func (r *zapRecord) Int(key string, val int) lork.Record {
	r.logger = r.logger.With(zap.Int(key, val))
	return r
}

func (r *zapRecord) Ints(key string, val []int) lork.Record {
	r.logger = r.logger.With(zap.Ints(key, val))
	return r
}

func (r *zapRecord) Int8(key string, val int8) lork.Record {
	r.logger = r.logger.With(zap.Int8(key, val))
	return r
}

func (r *zapRecord) Ints8(key string, val []int8) lork.Record {
	r.logger = r.logger.With(zap.Int8s(key, val))
	return r
}

func (r *zapRecord) Int16(key string, val int16) lork.Record {
	r.logger = r.logger.With(zap.Int16(key, val))
	return r
}

func (r *zapRecord) Ints16(key string, val []int16) lork.Record {
	r.logger = r.logger.With(zap.Int16s(key, val))
	return r
}

func (r *zapRecord) Int32(key string, val int32) lork.Record {
	r.logger = r.logger.With(zap.Int32(key, val))
	return r
}

func (r *zapRecord) Ints32(key string, val []int32) lork.Record {
	r.logger = r.logger.With(zap.Int32s(key, val))
	return r
}

func (r *zapRecord) Int64(key string, val int64) lork.Record {
	r.logger = r.logger.With(zap.Int64(key, val))
	return r
}

func (r *zapRecord) Ints64(key string, val []int64) lork.Record {
	r.logger = r.logger.With(zap.Int64s(key, val))
	return r
}

func (r *zapRecord) Uint(key string, val uint) lork.Record {
	r.logger.With(zap.Uint(key, val))
	return r
}

func (r *zapRecord) Uints(key string, val []uint) lork.Record {
	r.logger = r.logger.With(zap.Uints(key, val))
	return r
}

func (r *zapRecord) Uint8(key string, val uint8) lork.Record {
	r.logger = r.logger.With(zap.Uint8(key, val))
	return r
}

func (r *zapRecord) Uints8(key string, val []uint8) lork.Record {
	r.logger = r.logger.With(zap.Uint8s(key, val))
	return r
}

func (r *zapRecord) Uint16(key string, val uint16) lork.Record {
	r.logger = r.logger.With(zap.Uint16(key, val))
	return r
}

func (r *zapRecord) Uints16(key string, val []uint16) lork.Record {
	r.logger = r.logger.With(zap.Uint16s(key, val))
	return r
}

func (r *zapRecord) Uint32(key string, val uint32) lork.Record {
	r.logger = r.logger.With(zap.Uint32(key, val))
	return r
}

func (r *zapRecord) Uints32(key string, val []uint32) lork.Record {
	r.logger = r.logger.With(zap.Uint32s(key, val))
	return r
}

func (r *zapRecord) Uint64(key string, val uint64) lork.Record {
	r.logger = r.logger.With(zap.Uint64(key, val))
	return r
}

func (r *zapRecord) Uints64(key string, val []uint64) lork.Record {
	r.logger = r.logger.With(zap.Uint64s(key, val))
	return r
}

func (r *zapRecord) Float32(key string, val float32) lork.Record {
	r.logger = r.logger.With(zap.Float32(key, val))
	return r
}

func (r *zapRecord) Floats32(key string, val []float32) lork.Record {
	r.logger = r.logger.With(zap.Float32s(key, val))
	return r
}

func (r *zapRecord) Float64(key string, val float64) lork.Record {
	r.logger = r.logger.With(zap.Float64(key, val))
	return r
}

func (r *zapRecord) Floats64(key string, val []float64) lork.Record {
	r.logger = r.logger.With(zap.Float64s(key, val))
	return r
}

func (r *zapRecord) Time(key string, val time.Time) lork.Record {
	r.logger = r.logger.With(zap.Time(key, val))
	return r
}

func (r *zapRecord) Times(key string, val []time.Time) lork.Record {
	r.logger = r.logger.With(zap.Times(key, val))
	return r
}

func (r *zapRecord) Dur(key string, val time.Duration) lork.Record {
	r.logger = r.logger.With(zap.Duration(key, val))
	return r
}

func (r *zapRecord) Durs(key string, val []time.Duration) lork.Record {
	r.logger = r.logger.With(zap.Durations(key, val))
	return r
}

func (r *zapRecord) Any(key string, val interface{}) lork.Record {
	r.logger = r.logger.With(zap.Any(key, val))

	return r
}

func (r *zapRecord) Msge() {
	r.Msg("")
}

func (r *zapRecord) Msg(msg string) {
	switch r.level {
	case zapcore.DebugLevel:
		r.logger.Debug(msg)
	case zapcore.InfoLevel:
		r.logger.Info(msg)
	case zapcore.WarnLevel:
		r.logger.Warn(msg)
	case zapcore.ErrorLevel:
		r.logger.Error(msg)
	case zapcore.FatalLevel:
		r.logger.Fatal(msg)
	case zapcore.PanicLevel:
		r.logger.Panic(msg)
	}

	zapRecordPool.Put(r)
}

func (r *zapRecord) Msgf(format string, v ...interface{}) {
	sl := r.logger.Sugar()

	switch r.level {
	case zapcore.DebugLevel:
		sl.Debugf(format, v...)
	case zapcore.InfoLevel:
		sl.Infof(format, v...)
	case zapcore.WarnLevel:
		sl.Warnf(format, v...)
	case zapcore.ErrorLevel:
		sl.Errorf(format, v...)
	case zapcore.FatalLevel:
		sl.Fatalf(format, v...)
	case zapcore.PanicLevel:
		sl.Panicf(format, v...)
	}

	zapRecordPool.Put(r)
}
