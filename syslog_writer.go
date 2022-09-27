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

//go:build !windows && !plan9
// +build !windows,!plan9

package lork

import (
	"bytes"
	"log/syslog"
	"strconv"
)

type syslogWriter struct {
	sw   *syslog.Writer
	opts *SyslogWriterOption

	isStarted bool
}

type SyslogWriterOption struct {
	Tag string
	// See net.Dial
	Address, Network string
	Filter           Filter

	encoder Encoder
}

// NewSyslogWriter create a logging writer via syslog.
func NewSyslogWriter(options ...func(option *SyslogWriterOption)) Writer {
	opts := &SyslogWriterOption{
		encoder: NewPatternEncoder(func(o *PatternEncoderOption) {
			o.Pattern = "#level #message #fields"
			o.Converters = map[string]NewConverter{
				"level": newLevelIntConverter,
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

func (w *syslogWriter) Start() {
	if w.isStarted {
		return
	}

	sw, err := syslog.Dial(w.opts.Address, w.opts.Network, syslog.LOG_DEBUG, w.opts.Tag)
	if err != nil {
		ReportfExit("failed to dial syslog: %v", err)
	}

	w.sw = sw
	w.isStarted = true
}

func (w *syslogWriter) Stop() {
	_ = w.sw.Close()
}

func (w *syslogWriter) Write(p []byte) (n int, err error) {
	lvl, err := atoi(p[:1])
	if err != nil {
		return 0, err
	}
	msg := string(p[1:])

	switch Level(lvl) {
	case InfoLevel:
		err = w.sw.Info(msg)
	case WarnLevel:
		err = w.sw.Warning(msg)
	case ErrorLevel:
		err = w.sw.Err(msg)
	case FatalLevel:
		err = w.sw.Crit(msg)
	case PanicLevel:
		err = w.sw.Emerg(msg)
	case DebugLevel:
	case TraceLevel:
		fallthrough
	default:
		err = w.sw.Debug(msg)
	}

	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *syslogWriter) Encoder() Encoder {
	return w.opts.encoder
}

func (w *syslogWriter) Filter() Filter {
	return w.opts.Filter
}

type levelIntConverter struct {
	next Converter
}

func newLevelIntConverter() Converter {
	return &levelIntConverter{}
}

func (c *levelIntConverter) AttachNext(next Converter) {
	c.next = next
}

func (c *levelIntConverter) Next() Converter {
	return c.next
}

func (c *levelIntConverter) AttachChild(Converter) {
}

func (c *levelIntConverter) AttachOptions([]string) {
}

func (c *levelIntConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	event := origin.(*LogEvent)
	buf.WriteString(strconv.Itoa(int(event.LevelInt())))
}
