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

package zero

import (
	"github.com/coolerfall/lork"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	lorkLvlToZeroLvl = map[lork.Level]zerolog.Level{
		lork.TraceLevel: zerolog.TraceLevel,
		lork.DebugLevel: zerolog.DebugLevel,
		lork.InfoLevel:  zerolog.InfoLevel,
		lork.WarnLevel:  zerolog.WarnLevel,
		lork.ErrorLevel: zerolog.ErrorLevel,
		lork.FatalLevel: zerolog.FatalLevel,
		lork.PanicLevel: zerolog.PanicLevel,
	}
)

// zeroLogger is an implementation of ILogger.
type zeroLogger struct {
	name        string
	logger      zerolog.Logger
	multiWriter *lork.MultiWriter
}

// NewZeroLogger creates a new instance of zeroLogger used to be bound to lork.
func NewZeroLogger(name string, writer *lork.MultiWriter) lork.ILogger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	zerolog.TimeFieldFormat = lork.TimestampFormat
	zerolog.LevelFieldName = lork.LevelFieldKey
	zerolog.TimestampFieldName = lork.TimestampFieldKey
	zerolog.MessageFieldName = lork.MessageFieldKey
	zerolog.LevelFieldMarshalFunc = capitalLevel

	logger := zerolog.New(writer).With().Timestamp().Logger()
	log.Logger = logger

	return &zeroLogger{
		name:        name,
		logger:      logger,
		multiWriter: writer,
	}
}

func (l *zeroLogger) Name() string {
	return l.name
}

func (l *zeroLogger) AddWriter(w ...lork.Writer) {
	l.multiWriter.AddWriter(w...)
}

func (l *zeroLogger) ResetWriter() {
	l.multiWriter.ResetWriter()
}

func (l *zeroLogger) SetLevel(lvl lork.Level) {
	zerolog.SetGlobalLevel(lorkLvlToZeroLvl[lvl])
}

func (l *zeroLogger) Trace() lork.Record {
	return newZeroRecord(l.logger.Trace())
}

func (l *zeroLogger) Debug() lork.Record {
	return newZeroRecord(l.logger.Debug())
}

func (l *zeroLogger) Info() lork.Record {
	return newZeroRecord(l.logger.Info())
}

func (l *zeroLogger) Warn() lork.Record {
	return newZeroRecord(l.logger.Warn())
}

func (l *zeroLogger) Error() lork.Record {
	return newZeroRecord(l.logger.Error())
}

func (l *zeroLogger) Fatal() lork.Record {
	return newZeroRecord(l.logger.Fatal())
}

func (l *zeroLogger) Panic() lork.Record {
	return newZeroRecord(l.logger.Panic())
}

func (l *zeroLogger) Level(lvl lork.Level) lork.Record {
	return newZeroRecord(l.logger.WithLevel(lorkLvlToZeroLvl[lvl]))
}

func (l *zeroLogger) Event(e *lork.LogEvent) {
	if err := l.multiWriter.WriteEvent(e); err != nil {
		l.Error().Err(err).Msg("write raw event error")
	}
}

func capitalLevel(l zerolog.Level) string {
	switch l {
	case zerolog.DebugLevel:
		return "DEBUG"
	case zerolog.InfoLevel:
		return "INFO"
	case zerolog.WarnLevel:
		return "WARN"
	case zerolog.ErrorLevel:
		return "ERROR"
	case zerolog.FatalLevel:
		return "FATAL"
	case zerolog.PanicLevel:
		return "PANIC"
	case zerolog.NoLevel:
		fallthrough
	case zerolog.TraceLevel:
		fallthrough
	default:
		return "TRACE"
	}
}
