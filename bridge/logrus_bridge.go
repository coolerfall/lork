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

package bridge

import (
	"bytes"
	"sync"
	"time"

	"github.com/coolerfall/lork"
	"github.com/sirupsen/logrus"
)

var (
	logrusLvlToLorkLvl = map[logrus.Level]lork.Level{
		logrus.TraceLevel: lork.TraceLevel,
		logrus.DebugLevel: lork.DebugLevel,
		logrus.InfoLevel:  lork.InfoLevel,
		logrus.WarnLevel:  lork.WarnLevel,
		logrus.ErrorLevel: lork.ErrorLevel,
		logrus.FatalLevel: lork.FatalLevel,
	}
)

type logrusBridge struct {
	buf    *bytes.Buffer
	locker sync.Mutex
}

// NewLogrusBridge creates a new lork bridge for logrus.
func NewLogrusBridge() lork.Bridge {
	bridge := &logrusBridge{
		buf: new(bytes.Buffer),
	}
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: lork.LevelFieldKey,
			logrus.FieldKeyTime:  lork.TimestampFieldKey,
			logrus.FieldKeyMsg:   lork.MessageFieldKey,
		},
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(bridge)

	return bridge
}

func (b *logrusBridge) Name() string {
	return "github.com/sirupsen/logrus"
}

func (b *logrusBridge) ParseLevel(lvl string) lork.Level {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		lork.Reportf("parse logrus level error: %s", err)
		level = logrus.TraceLevel
	}

	return logrusLvlToLorkLvl[level]
}

func (b *logrusBridge) Write(p []byte) (int, error) {
	b.locker.Lock()
	defer b.locker.Unlock()

	_ = lork.ReplaceJson(p, b.buf, lork.LevelFieldKey,
		func(k, v []byte) (nk []byte, nv []byte, err error) {
			lvl, err := logrus.ParseLevel(string(v))
			if err != nil {
				return k, v, err
			} else {
				return k, []byte(logrusLvlToLorkLvl[lvl].String()), nil
			}
		})
	p = b.buf.Bytes()
	b.buf.Reset()

	lork.Logger().WriteEvent(lork.MakeEvent(p))

	return len(p), nil
}
