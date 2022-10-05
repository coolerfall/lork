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
	"github.com/coolerfall/lork"
	"github.com/sirupsen/logrus"
)

var (
	lorkLvlToLogrusLvl = map[lork.Level]logrus.Level{
		lork.TraceLevel: logrus.TraceLevel,
		lork.DebugLevel: logrus.DebugLevel,
		lork.InfoLevel:  logrus.InfoLevel,
		lork.WarnLevel:  logrus.WarnLevel,
		lork.ErrorLevel: logrus.ErrorLevel,
		lork.FatalLevel: logrus.FatalLevel,
		lork.PanicLevel: logrus.PanicLevel,
	}
)

// logrusLogger is an implementation of ILogger.
type logrusLogger struct {
	multiWriter *lork.MultiWriter
}

// NewLogrusLogger creates a new instance of logrusLogger used to be bound to lork
func NewLogrusLogger() lork.ILogger {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: lork.TimestampFormat,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: lork.LevelFieldKey,
			logrus.FieldKeyTime:  lork.TimestampFieldKey,
			logrus.FieldKeyMsg:   lork.MessageFieldKey,
		},
	})
	logrus.SetLevel(logrus.TraceLevel)

	writer := lork.NewMultiWriter()
	transformer := newTransformer(writer)
	logrus.SetOutput(transformer)

	return &logrusLogger{
		multiWriter: writer,
	}
}

func (l *logrusLogger) Name() string {
	return "github.com/sirupsen/logrus"
}

func (l *logrusLogger) AddWriter(w ...lork.Writer) {
	l.multiWriter.AddWriter(w...)
}

func (l *logrusLogger) ResetWriter() {
	l.multiWriter.Reset()
}

func (l *logrusLogger) SetLevel(lvl lork.Level) {
	logrus.SetLevel(lorkLvlToLogrusLvl[lvl])
}

func (l *logrusLogger) Trace() lork.Record {
	return newLogrusRecord(logrus.TraceLevel)
}

func (l *logrusLogger) Debug() lork.Record {
	return newLogrusRecord(logrus.DebugLevel)
}

func (l *logrusLogger) Info() lork.Record {
	return newLogrusRecord(logrus.InfoLevel)
}

func (l *logrusLogger) Warn() lork.Record {
	return newLogrusRecord(logrus.WarnLevel)
}

func (l *logrusLogger) Error() lork.Record {
	return newLogrusRecord(logrus.ErrorLevel)
}

func (l *logrusLogger) Fatal() lork.Record {
	return newLogrusRecord(logrus.FatalLevel)
}

func (l *logrusLogger) Panic() lork.Record {
	return newLogrusRecord(logrus.PanicLevel)
}

func (l *logrusLogger) Level(lvl lork.Level) lork.Record {
	return newLogrusRecord(lorkLvlToLogrusLvl[lvl])
}

func (l *logrusLogger) WriteEvent(e *lork.LogEvent) {
	if err := l.multiWriter.WriteEvent(e); err != nil {
		l.Error().Err(err).Msg("write raw event error")
	}
}
