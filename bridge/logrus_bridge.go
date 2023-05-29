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

package bridge

import (
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
}

// NewLogrusBridge creates a new lork bridge for logrus.
func NewLogrusBridge() lork.Bridge {
	bridge := &logrusBridge{}
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: lork.TimestampFormat,
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
	lork.BridgeWrite(b, p)

	return len(p), nil
}
