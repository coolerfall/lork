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

package slago

import (
	"io"
	"sync"
)

type EventWriter interface {
	Write(event *LogEvent) (err error)
}

// Writer is the interface that wraps the io.Writer, add
// Encoder and Filter func for slago to encode and filter logs.
type Writer interface {
	io.Writer

	// Encoder returns encoder used in current writer.
	Encoder() Encoder

	// Filter returns filter used in current writer.
	Filter() Filter
}

// MultiWriter represents multiple writer which implements slago.Writer.
// This writer is used as output which will implement SlaLogger.
type MultiWriter struct {
	locker       sync.Mutex
	writers      []Writer
	asyncWriters []Writer
}

// NewMultiWriter creates a new multiple writer.
func NewMultiWriter() *MultiWriter {
	return &MultiWriter{
		writers:      make([]Writer, 0),
		asyncWriters: make([]Writer, 0),
	}
}

// AddWriter adds a slago writer into multi writer.
func (mw *MultiWriter) AddWriter(writers ...Writer) {
	for _, w := range writers {
		if lc, ok := w.(Lifecycle); ok {
			lc.Start()
		}

		if _, ok := w.(*asyncWriter); ok {
			mw.asyncWriters = append(mw.asyncWriters, w)
		} else {
			mw.writers = append(mw.writers, w)
		}
	}
}

// Reset will remove all writers.
func (mw *MultiWriter) Reset() {
	mw.locker.Lock()
	defer mw.locker.Unlock()

	for _, w := range mw.writers {
		if lc, ok := w.(Lifecycle); ok {
			lc.Stop()
		}
	}
	mw.writers = make([]Writer, 0)
	mw.asyncWriters = make([]Writer, 0)
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.locker.Lock()
	defer mw.locker.Unlock()

	if n, err = mw.writeAsync(p); err != nil {
		return
	}

	if n, err = mw.writeNormal(p); err != nil {
		return
	}

	return len(p), nil
}

func (mw *MultiWriter) writeAsync(p []byte) (n int, err error) {
	for _, w := range mw.asyncWriters {
		if n, err = w.Write(p); err != nil {
			return
		}
	}

	return len(p), nil
}

func (mw *MultiWriter) writeNormal(p []byte) (n int, err error) {
	if len(mw.writers) == 0 {
		return 0, nil
	}

	event := MakeEvent(p)
	defer event.Recycle()
	for _, w := range mw.writers {
		if w.Filter() != nil && w.Filter().Do(event) {
			return
		}

		encoded := p
		if w.Encoder() != nil {
			encoded, err = w.Encoder().Encode(event)
			if err != nil {
				return 0, err
			}
		}
		n, err = w.Write(encoded)
	}

	return len(p), nil
}
