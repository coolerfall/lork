// Copyright (c) 2019-2020 Vincent Cheung (coolingfall@gmail.com).
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

package bridge

import (
	"bytes"
	"sync"
	"time"

	"github.com/coolerfall/slago"
	"github.com/sirupsen/logrus"
)

var (
	logrusLvlToSlagoLvl = map[logrus.Level]slago.Level{
		logrus.TraceLevel: slago.TraceLevel,
		logrus.DebugLevel: slago.DebugLevel,
		logrus.InfoLevel:  slago.InfoLevel,
		logrus.WarnLevel:  slago.WarnLevel,
		logrus.ErrorLevel: slago.ErrorLevel,
		logrus.FatalLevel: slago.FatalLevel,
	}
)

type logrusBridge struct {
	buf    *bytes.Buffer
	locker sync.Mutex
}

// NewLogrusBridge creates a new slago bridge for logrus.
func NewLogrusBridge() slago.Bridge {
	bridge := &logrusBridge{
		buf: new(bytes.Buffer),
	}
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: slago.LevelFieldKey,
			logrus.FieldKeyTime:  slago.TimestampFieldKey,
			logrus.FieldKeyMsg:   slago.MessageFieldKey,
		},
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(bridge)

	return bridge
}

func (b *logrusBridge) Name() string {
	return "github.com/sirupsen/logrus"
}

func (b *logrusBridge) ParseLevel(lvl string) slago.Level {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		slago.Reportf("parse logrus level error: %s", err)
		level = logrus.TraceLevel
	}

	return logrusLvlToSlagoLvl[level]
}

func (b *logrusBridge) Write(p []byte) (int, error) {
	b.locker.Lock()
	defer b.locker.Unlock()

	_ = slago.ReplaceJson(p, b.buf, slago.LevelFieldKey,
		func(k, v []byte) (nk []byte, nv []byte, err error) {
			lvl, err := logrus.ParseLevel(string(v))
			if err != nil {
				return k, v, err
			} else {
				return k, []byte(logrusLvlToSlagoLvl[lvl].String()), nil
			}
		})
	p = b.buf.Bytes()
	b.buf.Reset()

	err := slago.BrigeWrite(b, p)
	if err != nil {
		slago.Reportf("logrus bridge write error", err)
	}

	return len(p), err
}
