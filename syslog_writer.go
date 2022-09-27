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

package lork

import (
	"bytes"
	"log/syslog"
	"os"
	"strconv"
)

var (
	lorkLvlToSyslogPriority = map[Level]syslog.Priority{
		TraceLevel: syslog.LOG_DEBUG,
		DebugLevel: syslog.LOG_DEBUG,
		InfoLevel:  syslog.LOG_INFO,
		WarnLevel:  syslog.LOG_WARNING,
		ErrorLevel: syslog.LOG_ERR,
		FatalLevel: syslog.LOG_EMERG,
		PanicLevel: syslog.LOG_CRIT,
	}
)

type syslogWriter struct {
	w    *syslog.Writer
	opts *SyslogWriterOption
}

type SyslogWriterOption struct {
	Filter  Filter
	encoder Encoder
}

func NewSyslogWriter(options ...func(option *SyslogWriterOption)) Writer {
	opts := &SyslogWriterOption{
		encoder: NewPatternEncoder(func(opt *PatternEncoderOption) {
			opt.Pattern = "#syslog: #message #fields"
			opt.Converters = map[string]NewConverter{
				"syslog": newSyslogConverter,
			}
		}),
	}

	for _, f := range options {
		f(opts)
	}

	return &syslogWriter{
		opts: opts,
	}
}

func (w *syslogWriter) Write(p []byte) (n int, err error) {
	panic("implement me")
}

func (w *syslogWriter) Encoder() Encoder {
	return w.opts.encoder
}

func (w *syslogWriter) Filter() Filter {
	return w.opts.Filter
}

type syslogConverter struct {
	next Converter
}

func newSyslogConverter() Converter {
	return &syslogConverter{}
}

func (c *syslogConverter) AttachNext(next Converter) {
	c.next = next
}

func (c *syslogConverter) Next() Converter {
	return c.next
}

func (c *syslogConverter) AttachChild(_ Converter) {
}

func (c *syslogConverter) AttachOptions(_ []string) {
}

func (c *syslogConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	event := origin.(*LogEvent)
	buf.WriteByte('<')
	priority := lorkLvlToSyslogPriority[event.LevelInt()]
	buf.WriteString(strconv.Itoa(int(priority)))
	buf.WriteByte('>')
	buf.WriteString(strconv.Itoa(os.Getpid()))
}
