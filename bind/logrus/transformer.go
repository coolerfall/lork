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
	"bytes"
	"sync"

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
		logrus.PanicLevel: lork.PanicLevel,
	}
)

type transformer struct {
	buf         *bytes.Buffer
	locker      sync.Mutex
	multiWriter *lork.MultiWriter
}

func newTransformer(w *lork.MultiWriter) *transformer {
	return &transformer{
		buf:         new(bytes.Buffer),
		multiWriter: w,
	}
}

func (t *transformer) Write(p []byte) (n int, err error) {
	t.locker.Lock()
	defer t.locker.Unlock()

	_ = lork.ReplaceJson(p, t.buf, lork.LevelFieldKey,
		func(k, v []byte) ([]byte, []byte, error) {
			lvl, err := logrus.ParseLevel(string(v))
			if err != nil {
				return k, v, err
			} else {
				return k, []byte(logrusLvlToLorkLvl[lvl].String()), nil
			}
		})
	p = t.buf.Bytes()
	t.buf.Reset()

	return t.multiWriter.Write(p)
}
