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

package slagologrus

import (
	"io"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gitlab.com/anbillon/slago/slago-api"
)

var (
	slagoLvlToLogrusLvl = map[slago.Level]logrus.Level{
		slago.TraceLevel: logrus.TraceLevel,
		slago.DebugLevel: logrus.DebugLevel,
		slago.InfoLevel:  logrus.InfoLevel,
		slago.WarnLevel:  logrus.WarnLevel,
		slago.ErrorLevel: logrus.ErrorLevel,
		slago.FatalLevel: logrus.FatalLevel,
		slago.PanicLevel: logrus.PanicLevel,
	}
)

type logrusLogger struct {
	writers []io.Writer
}

func init() {
	slago.Bind(newLogrusLogger())
}

func newLogrusLogger() *logrusLogger {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)

	return &logrusLogger{
		writers: make([]io.Writer, 0),
	}
}

func (l *logrusLogger) Name() string {
	return "logrus"
}

func (l *logrusLogger) AddWriter(w io.Writer) {
	l.writers = append(l.writers, w)
	logrus.SetOutput(io.MultiWriter(l.writers...))
}

func (l *logrusLogger) SetLevel(lvl slago.Level) {
	logrus.SetLevel(slagoLvlToLogrusLvl[lvl])
}

func (l *logrusLogger) Level(lvl slago.Level) slago.Record {
	logrusLevel := slagoLvlToLogrusLvl[lvl]

	switch logrusLevel {
	case logrus.DebugLevel:
		return l.Debug()
	case logrus.InfoLevel:
		return l.Info()
	case logrus.WarnLevel:
		return l.Warn()
	case logrus.ErrorLevel:
		return l.Error()
	case logrus.FatalLevel:
		return l.Fatal()
	case logrus.PanicLevel:
		return l.Panic()
	case logrus.TraceLevel:
		fallthrough
	default:
		return l.Trace()
	}
}

func (l *logrusLogger) Trace() slago.Record {
	return newLogrusRecord(logrus.TraceLevel)
}

func (l *logrusLogger) Debug() slago.Record {
	return newLogrusRecord(logrus.DebugLevel)
}

func (l *logrusLogger) Info() slago.Record {
	return newLogrusRecord(logrus.InfoLevel)
}

func (l *logrusLogger) Warn() slago.Record {
	return newLogrusRecord(logrus.WarnLevel)
}

func (l *logrusLogger) Error() slago.Record {
	return newLogrusRecord(logrus.ErrorLevel)
}

func (l *logrusLogger) Fatal() slago.Record {
	return newLogrusRecord(logrus.FatalLevel)
}

func (l *logrusLogger) Panic() slago.Record {
	return newLogrusRecord(logrus.PanicLevel)
}

func (l *logrusLogger) Print(args ...interface{}) {
	logrus.Print(args...)
}

func (l *logrusLogger) Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}
