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

//go:build !windows && !plan9
// +build !windows,!plan9

package lork

import (
	"log/syslog"
	"sync"
)

type syslogWriter struct {
	sw   *syslog.Writer
	opts *SyslogWriterOption

	encoder   Encoder
	locker    sync.Mutex
	isStarted bool
}

type SyslogWriterOption struct {
	Name string
	Tag  string
	// See net.Dial
	Address, Network string
	Filter           Filter
}

// NewSyslogWriter create a logging writer via syslog.
func NewSyslogWriter(options ...func(option *SyslogWriterOption)) Writer {
	opts := &SyslogWriterOption{}

	for _, f := range options {
		f(opts)
	}

	sw := &syslogWriter{
		opts: opts,
		encoder: NewPatternEncoder(func(o *PatternEncoderOption) {
			o.Pattern = "#message #fields"
		}),
	}

	return NewEventWriter(sw)
}

func (w *syslogWriter) Start() {
	w.locker.Lock()
	defer w.locker.Unlock()

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
	w.locker.Lock()
	defer w.locker.Unlock()

	_ = w.sw.Close()
	w.isStarted = false
}

func (w *syslogWriter) Write(event *LogEvent) (err error) {
	data, err := w.encoder.Encode(event)
	if err != nil {
		return err
	}
	msg := string(data)

	switch event.LevelInt() {
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

	return err
}

func (w *syslogWriter) Name() string {
	return w.opts.Name
}

func (w *syslogWriter) Filter() Filter {
	return w.opts.Filter
}

func (w *syslogWriter) Synchronized() bool {
	return false
}
