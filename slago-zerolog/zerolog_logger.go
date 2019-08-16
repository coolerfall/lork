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

package salzerolog

import (
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/anbillon/slago/slago-api"
	"gitlab.com/anbillon/slago/slago-api/helpers"
)

var (
	slagoLvlToZeroLvl = map[slago.Level]zerolog.Level{
		slago.TraceLevel: zerolog.NoLevel,
		slago.DebugLevel: zerolog.DebugLevel,
		slago.InfoLevel:  zerolog.InfoLevel,
		slago.WarnLevel:  zerolog.WarnLevel,
		slago.ErrorLevel: zerolog.ErrorLevel,
		slago.FatalLevel: zerolog.FatalLevel,
		slago.PanicLevel: zerolog.PanicLevel,
	}
)

type zeroLogger struct {
	mutex   sync.Mutex
	logger  zerolog.Logger
	writers []io.Writer
}

func init() {
	slago.Bind(newZeroLogger())
}

func newZeroLogger() *zeroLogger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	zerolog.LevelFieldName = helpers.LevelFieldKey
	zerolog.TimestampFieldName = helpers.TimestampFieldKey
	zerolog.MessageFieldName = helpers.MessageFieldKey
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		return strings.ToUpper(l.String())
	}
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()
	log.Logger = logger

	return &zeroLogger{
		logger:  logger,
		writers: make([]io.Writer, 0),
	}
}

func (l *zeroLogger) Name() string {
	return "zerolog"
}

func (l *zeroLogger) AddWriter(w io.Writer) {
	l.mutex.Lock()
	l.writers = append(l.writers, w)
	mw := zerolog.MultiLevelWriter(l.writers...)
	l.logger = l.logger.Output(mw)
	log.Logger = l.logger
	l.mutex.Unlock()
}

func (l *zeroLogger) SetLevel(lvl slago.Level) {
	zerolog.SetGlobalLevel(slagoLvlToZeroLvl[lvl])
}

func (l *zeroLogger) Level(lvl slago.Level) slago.Record {
	return newZeroRecord(l.logger.WithLevel(slagoLvlToZeroLvl[lvl]))
}

func (l *zeroLogger) Trace() slago.Record {
	return newZeroRecord(l.logger.WithLevel(zerolog.NoLevel))
}

func (l *zeroLogger) Debug() slago.Record {
	return newZeroRecord(l.logger.Debug())
}

func (l *zeroLogger) Info() slago.Record {
	return newZeroRecord(l.logger.Info())
}

func (l *zeroLogger) Warn() slago.Record {
	return newZeroRecord(l.logger.Warn())
}

func (l *zeroLogger) Error() slago.Record {
	return newZeroRecord(l.logger.Error())
}

func (l *zeroLogger) Fatal() slago.Record {
	return newZeroRecord(l.logger.Fatal())
}

func (l *zeroLogger) Panic() slago.Record {
	return newZeroRecord(l.logger.Panic())
}

func (l *zeroLogger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

func (l *zeroLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}
