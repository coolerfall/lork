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

package slago

import (
	"os"
)

type consoleWriter struct {
	encoder Encoder
	filter  Filter
}

// ConsoleWriterOption represents available options for console writer.
type ConsoleWriterOption struct {
	Encoder Encoder
	Filter  Filter
}

// NewConsoleWriter creates a new instance of console writer.
func NewConsoleWriter(options ...func(*ConsoleWriterOption)) Writer {
	opt := &ConsoleWriterOption{
		Encoder: NewPatternEncoder(),
	}

	for _, f := range options {
		f(opt)
	}

	return &consoleWriter{
		encoder: opt.Encoder,
		filter:  opt.Filter,
	}
}

func (w *consoleWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (w *consoleWriter) Encoder() Encoder {
	return w.encoder
}

func (w *consoleWriter) Filter() Filter {
	return w.filter
}
