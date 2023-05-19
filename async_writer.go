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
	"sync"
)

type AsyncWriter struct {
	name        string
	locker      sync.Mutex
	queue       *BlockingQueue
	isRunning   bool
	multiWriter *MultiWriter
}

// AsyncWriterOption represents available options for async writer.
type AsyncWriterOption struct {
	Name      string
	QueueSize int
}

// NewAsyncWriter creates a new instance of asynchronous writer.
func NewAsyncWriter(options ...func(*AsyncWriterOption)) *AsyncWriter {
	opts := &AsyncWriterOption{
		QueueSize: DefaultQueueSize,
	}

	for _, f := range options {
		f(opts)
	}

	return &AsyncWriter{
		name:        opts.Name,
		queue:       NewBlockingQueue(opts.QueueSize),
		multiWriter: NewMultiWriter(),
	}
}

func (w *AsyncWriter) Start() {
	if w.isRunning {
		return
	}
	w.isRunning = true
	go w.startWorker()
}

func (w *AsyncWriter) Stop() {
	w.locker.Lock()
	defer w.locker.Unlock()
	w.multiWriter.ResetWriter()

	w.isRunning = false
}

func (w *AsyncWriter) DoWrite(event *LogEvent) error {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.queue.RemainCapacity() <= 16 {
		// discard
		return nil
	}

	// copy a log event for further usage
	w.queue.Put(event.Copy())

	return nil
}

func (w *AsyncWriter) Name() string {
	return w.name
}

func (w *AsyncWriter) AddWriter(writers ...Writer) {
	w.multiWriter.AddWriter(writers...)
}

func (w *AsyncWriter) GetWriter(name string) Writer {
	return w.multiWriter.GetWriter(name)
}

func (w *AsyncWriter) Attached(writer Writer) bool {
	return w.multiWriter.Attached(writer)
}

func (w *AsyncWriter) ResetWriter() {
	w.multiWriter.ResetWriter()
}

func (w *AsyncWriter) startWorker() {
	for {
		if !w.isRunning {
			break
		}

		p := (w.queue.Take()).(*LogEvent)
		if err := w.multiWriter.WriteEvent(p); err != nil {
			Reportf("async writer write error: %v", err)
		}
	}
}
