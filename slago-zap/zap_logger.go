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

package salzap

import (
	"io"
	"io/ioutil"

	"gitlab.com/anbillon/slago/slago-api"
	"gitlab.com/anbillon/slago/slago-api/helpers"
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

type zapLogger struct {
	atomicLevel zap.AtomicLevel
	writers     []io.Writer
}

func init() {
	slago.Bind(newZapLogger())
}

func newZapLogger() *zapLogger {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = helpers.LevelFieldKey
	encoderConfig.MessageKey = helpers.MessageFieldKey
	encoderConfig.TimeKey = helpers.TimestampFieldKey
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(ioutil.Discard),
		atomicLevel,
	))

	zap.ReplaceGlobals(logger)

	return &zapLogger{
		atomicLevel: atomicLevel,
		writers:     make([]io.Writer, 0),
	}
}

func (l *zapLogger) Name() string {
	return "zap"
}

func (l *zapLogger) AddWriter(w io.Writer) {
	l.writers = append(l.writers, w)
}

func (l *zapLogger) SetLevel(lvl slago.Level) {
	l.atomicLevel.SetLevel(slagoLvlToZapLvl[lvl])
}

func (l *zapLogger) Level(lvl slago.Level) slago.Record {
	zapLevel := slagoLvlToZapLvl[lvl]

	switch zapLevel {
	case zapcore.InfoLevel:
		return l.Info()
	case zapcore.WarnLevel:
		return l.Warn()
	case zapcore.ErrorLevel:
		return l.Error()
	case zapcore.FatalLevel:
		return l.Fatal()
	case zapcore.PanicLevel:
		return l.Panic()
	case zapcore.DebugLevel:
		fallthrough
	default:
		return l.Debug()
	}
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

func (l *zapLogger) Print(args ...interface{}) {
	zap.S().Debug(args...)
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	zap.S().Debugf(format, args...)
}
