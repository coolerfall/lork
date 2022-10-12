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
	"io"
	"sync"
)

// Writer is interface represents the raw writer of lork.
// The lork will write LogEvent without any transform.
type Writer interface {
	Nameable

	// DoWrite this is where a writer accomplishes its work.
	// Lork will send LogEvent to this writer.
	DoWrite(event *LogEvent) error
}

// EventWriter represents writer which will write raw LogEvent with filter.
// Create Writer with NewEventWriter if writer implemented EventWriter.
type EventWriter interface {
	Nameable

	// Write Writes given event.
	Write(event *LogEvent) error

	// Filter returns filter used in current writer.
	Filter() Filter
}

// BytesWriter represents a writer which will write bytes with Encoder and Filter.
// Create Writer with NewBytesWriter if writer implemented BytesWriter.
type BytesWriter interface {
	io.Writer

	Nameable

	// Encoder returns encoder used in current writer.
	Encoder() Encoder

	// Filter returns filter used in current writer.
	Filter() Filter
}

// MultiWriter represents multiple writer which implements EventWriter.
// This writer is used as output which will implement ILogger.
type MultiWriter struct {
	locker  sync.Mutex
	writers []Writer
}

// NewMultiWriter creates a new multiple writer.
func NewMultiWriter() *MultiWriter {
	return &MultiWriter{
		writers: make([]Writer, 0),
	}
}

// AddWriter adds a lork writer into multi writer.
func (mw *MultiWriter) AddWriter(writers ...Writer) {
	for _, w := range writers {
		if lc, ok := w.(Lifecycle); ok {
			lc.Start()
		}

		mw.writers = append(mw.writers, w)
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
	mw.writers = mw.writers[:0]
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.locker.Lock()
	defer mw.locker.Unlock()

	if len(mw.writers) == 0 {
		return 0, nil
	}

	err = mw.WriteEvent(MakeEvent(p))

	return len(p), err
}

func (mw *MultiWriter) WriteEvent(event *LogEvent) (err error) {
	defer event.Recycle()

	for _, w := range mw.writers {
		if err = w.DoWrite(event); err != nil {
			Reportf("write event with writer [%v] error: %v", w.Name(), err)
		}
	}

	return nil
}

type eventWriter struct {
	ref EventWriter
}

// NewEventWriter creates a Writer with given EventWriter.
func NewEventWriter(w EventWriter) Writer {
	return &eventWriter{
		ref: w,
	}
}

func (w *eventWriter) Start() {
	if lw, ok := w.ref.(Lifecycle); ok {
		lw.Start()
	}
}

func (w *eventWriter) Stop() {
	if lw, ok := w.ref.(Lifecycle); ok {
		lw.Stop()
	}
}

func (w *eventWriter) Name() string {
	return w.ref.Name()
}

func (w *eventWriter) DoWrite(event *LogEvent) error {
	if w.ref.Filter() != nil && w.ref.Filter().Do(event) == Deny {
		return nil
	}

	return w.ref.Write(event)
}

type bytesWriter struct {
	ref BytesWriter
}

// NewBytesWriter creates a Writer with given BytesWriter.
func NewBytesWriter(w BytesWriter) Writer {
	return &bytesWriter{
		ref: w,
	}
}

func (w *bytesWriter) Start() {
	if w.ref.Encoder() == nil {
		ReportfExit("no encoder found in writer: %v", w.ref.Name())
	}

	if lw, ok := w.ref.(Lifecycle); ok {
		lw.Start()
	}
}

func (w *bytesWriter) Stop() {
	if lw, ok := w.ref.(Lifecycle); ok {
		lw.Stop()
	}
}

func (w *bytesWriter) Name() string {
	return w.ref.Name()
}

func (w *bytesWriter) DoWrite(event *LogEvent) error {
	if w.ref.Filter() != nil && w.ref.Filter().Do(event) == Deny {
		return nil
	}

	encoded, err := w.ref.Encoder().Encode(event)
	if err != nil {
		return err
	}
	_, err = w.ref.Write(encoded)

	return err
}
