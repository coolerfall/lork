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

	"github.com/coolerfall/slago"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	slagoLvlToZapLvl = map[slago.Level]zapcore.Level{
		slago.TraceLevel: zapcore.DebugLevel,
		slago.DebugLevel: zapcore.DebugLevel,
		slago.InfoLevel:  zapcore.InfoLevel,
		slago.WarnLevel:  zapcore.WarnLevel,
		slago.ErrorLevel: zapcore.ErrorLevel,
		slago.FatalLevel: zapcore.FatalLevel,
		slago.PanicLevel: zapcore.PanicLevel,
	}
)

// zapLogger is an implementation of SlaLogger.
type zapLogger struct {
	atomicLevel zap.AtomicLevel
	multiWriter *slago.MultiWriter
}

// NewZapLogger creates a new instance of zapLogger used to be bound to slago.
func NewZapLogger() slago.SlaLogger {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = slago.LevelFieldKey
	encoderConfig.MessageKey = slago.MessageFieldKey
	encoderConfig.TimeKey = slago.TimestampFieldKey
	encoderConfig.EncodeTime = rf3339Encoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	writer := slago.NewMultiWriter()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		atomicLevel,
	))

	zap.ReplaceGlobals(logger)

	return &zapLogger{
		atomicLevel: atomicLevel,
		multiWriter: writer,
	}
}

func (l *zapLogger) Name() string {
	return "go.uber.org/zap"
}

func (l *zapLogger) AddWriter(w ...slago.Writer) {
	l.multiWriter.AddWriter(w...)
}

func (l *zapLogger) ResetWriter() {
	l.multiWriter.Reset()
}

func (l *zapLogger) SetLevel(lvl slago.Level) {
	l.atomicLevel.SetLevel(slagoLvlToZapLvl[lvl])
}

func (l *zapLogger) Trace() slago.Record {
	return l.Debug()
}

func (l *zapLogger) Debug() slago.Record {
	return newZapRecord(zapcore.DebugLevel)
}

func (l *zapLogger) Info() slago.Record {
	return newZapRecord(zapcore.InfoLevel)
}

func (l *zapLogger) Warn() slago.Record {
	return newZapRecord(zapcore.WarnLevel)
}

func (l *zapLogger) Error() slago.Record {
	return newZapRecord(zapcore.ErrorLevel)
}

func (l *zapLogger) Fatal() slago.Record {
	return newZapRecord(zapcore.FatalLevel)
}

func (l *zapLogger) Panic() slago.Record {
	return newZapRecord(zapcore.PanicLevel)
}

func (l *zapLogger) Level(lvl slago.Level) slago.Record {
	return newZapRecord(slagoLvlToZapLvl[lvl])
}

func (l *zapLogger) WriteRaw(p []byte) {
	_, err := l.multiWriter.Write(p)
	if err != nil {
		l.Error().Err(err).Msg("write raw error")
	}
}

func rf3339Encoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(slago.TimestampFormat))
}
