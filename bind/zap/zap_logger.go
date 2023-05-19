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

package zap

import (
	"time"

	"github.com/coolerfall/lork"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	lorkLvlToZapLvl = map[lork.Level]zapcore.Level{
		lork.TraceLevel: zapcore.DebugLevel,
		lork.DebugLevel: zapcore.DebugLevel,
		lork.InfoLevel:  zapcore.InfoLevel,
		lork.WarnLevel:  zapcore.WarnLevel,
		lork.ErrorLevel: zapcore.ErrorLevel,
		lork.FatalLevel: zapcore.FatalLevel,
		lork.PanicLevel: zapcore.PanicLevel,
	}
)

// zapLogger is an implementation of ILogger.
type zapLogger struct {
	name        string
	atomicLevel zap.AtomicLevel
	multiWriter *lork.MultiWriter
}

// NewZapLogger creates a new instance of zapLogger used to be bound to lork.
func NewZapLogger(name string, writer *lork.MultiWriter) lork.ILogger {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = lork.LevelFieldKey
	encoderConfig.MessageKey = lork.MessageFieldKey
	encoderConfig.TimeKey = lork.TimestampFieldKey
	encoderConfig.EncodeTime = rf3339Encoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		atomicLevel,
	))

	zap.ReplaceGlobals(logger)

	return &zapLogger{
		name:        name,
		atomicLevel: atomicLevel,
		multiWriter: writer,
	}
}

func (l *zapLogger) Name() string {
	return l.name
}

func (l *zapLogger) AddWriter(w ...lork.Writer) {
	l.multiWriter.AddWriter(w...)
}

func (l *zapLogger) ResetWriter() {
	l.multiWriter.ResetWriter()
}

func (l *zapLogger) SetLevel(lvl lork.Level) {
	l.atomicLevel.SetLevel(lorkLvlToZapLvl[lvl])
}

func (l *zapLogger) Trace() lork.Record {
	return l.Debug()
}

func (l *zapLogger) Debug() lork.Record {
	return newZapRecord(zapcore.DebugLevel)
}

func (l *zapLogger) Info() lork.Record {
	return newZapRecord(zapcore.InfoLevel)
}

func (l *zapLogger) Warn() lork.Record {
	return newZapRecord(zapcore.WarnLevel)
}

func (l *zapLogger) Error() lork.Record {
	return newZapRecord(zapcore.ErrorLevel)
}

func (l *zapLogger) Fatal() lork.Record {
	return newZapRecord(zapcore.FatalLevel)
}

func (l *zapLogger) Panic() lork.Record {
	return newZapRecord(zapcore.PanicLevel)
}

func (l *zapLogger) Level(lvl lork.Level) lork.Record {
	return newZapRecord(lorkLvlToZapLvl[lvl])
}

func (l *zapLogger) Event(e *lork.LogEvent) {
	if err := l.multiWriter.WriteEvent(e); err != nil {
		l.Error().Err(err).Msg("write raw event error")
	}
}

func rf3339Encoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(lork.TimestampFormat))
}
