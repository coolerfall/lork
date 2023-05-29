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

package lork

import (
	"os"
	"sync"
)

type consoleWriter struct {
	opts   *ConsoleWriterOption
	locker sync.Locker
}

// ConsoleWriterOption represents available options for console writer.
type ConsoleWriterOption struct {
	Name    string
	Encoder Encoder
	Filter  Filter
}

// NewConsoleWriter creates a new instance of console writer.
func NewConsoleWriter(options ...func(*ConsoleWriterOption)) Writer {
	opts := &ConsoleWriterOption{
		Encoder: NewPatternEncoder(),
	}

	for _, f := range options {
		f(opts)
	}

	cw := &consoleWriter{
		opts:   opts,
		locker: new(sync.Mutex),
	}

	return NewBytesWriter(cw)
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	return os.Stdout.Write(p)
}

func (w *consoleWriter) Name() string {
	return w.opts.Name
}

func (w *consoleWriter) Encoder() Encoder {
	return w.opts.Encoder
}

func (w *consoleWriter) Filter() Filter {
	return w.opts.Filter
}
