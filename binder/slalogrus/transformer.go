// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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

package slalogrus

import (
	"bytes"
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.com/anbillon/slago"
)

var (
	logrusLvlToSlagoLvl = map[logrus.Level]slago.Level{
		logrus.TraceLevel: slago.TraceLevel,
		logrus.DebugLevel: slago.DebugLevel,
		logrus.InfoLevel:  slago.InfoLevel,
		logrus.WarnLevel:  slago.WarnLevel,
		logrus.ErrorLevel: slago.ErrorLevel,
		logrus.FatalLevel: slago.FatalLevel,
		logrus.PanicLevel: slago.PanicLevel,
	}
)

type transformer struct {
	buf         *bytes.Buffer
	locker      sync.Mutex
	multiWriter *slago.MultiWriter
}

func newTransformer(w *slago.MultiWriter) *transformer {
	return &transformer{
		buf:         new(bytes.Buffer),
		multiWriter: w,
	}
}

func (t *transformer) Write(p []byte) (n int, err error) {
	t.locker.Lock()
	defer t.locker.Unlock()

	_ = slago.ReplaceJson(p, t.buf, slago.LevelFieldKey,
		func(k, v []byte) ([]byte, []byte, error) {
			lvl, err := logrus.ParseLevel(string(v))
			if err != nil {
				return k, v, err
			} else {
				return k, []byte(logrusLvlToSlagoLvl[lvl].String()), nil
			}
		})
	p = t.buf.Bytes()
	t.buf.Reset()

	return t.multiWriter.Write(p)
}
