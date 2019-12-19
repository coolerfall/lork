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

package slalogrus

import (
	"bytes"
	"sync"

	"github.com/buger/jsonparser"
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

	t.buf.WriteByte('{')
	var start = false
	_ = jsonparser.ObjectEach(p, func(key []byte, value []byte,
		dataType jsonparser.ValueType, _ int) error {
		if start {
			t.buf.WriteByte(',')
		} else {
			start = true
		}

		t.buf.WriteByte('"')
		t.buf.Write(key)
		t.buf.WriteByte('"')
		t.buf.WriteByte(':')

		switch dataType {
		case jsonparser.String:
			t.buf.WriteByte('"')
			if string(key) == slago.LevelFieldKey {
				lvl, err := logrus.ParseLevel(string(value))
				if err != nil {
					t.buf.Write(value)
				} else {
					t.buf.WriteString(logrusLvlToSlagoLvl[lvl].String())
				}
			} else {
				t.buf.Write(value)
			}
			t.buf.WriteByte('"')

		default:
			t.buf.Write(value)
		}

		return nil
	})
	t.buf.WriteByte('}')
	t.buf.WriteByte('\n')
	p = t.buf.Bytes()
	t.buf.Reset()

	return t.multiWriter.Write(p)
}
