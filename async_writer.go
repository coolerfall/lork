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

const defaultWriterQueueSize = 512

type asyncWriter struct {
	name      string
	ref       Writer
	locker    sync.Mutex
	queue     *blockingQueue
	isRunning bool
}

// AsyncWriterOption represents available options for async writer.
type AsyncWriterOption struct {
	Name      string
	RefWriter Writer
	QueueSize int
}

// NewAsyncWriter creates a new instance of asynchronous writer.
func NewAsyncWriter(options ...func(*AsyncWriterOption)) Writer {
	opts := &AsyncWriterOption{
		QueueSize: defaultWriterQueueSize,
	}

	for _, f := range options {
		f(opts)
	}

	if opts.RefWriter == nil {
		ReportfExit("async writer need a referenced writer")
	}

	return &asyncWriter{
		name:  opts.Name,
		ref:   opts.RefWriter,
		queue: NewBlockingQueue(opts.QueueSize),
	}
}

func (w *asyncWriter) Start() {
	if w.isRunning {
		return
	}
	if lc, ok := w.ref.(Lifecycle); ok {
		lc.Start()
	}
	w.isRunning = true
	go w.startWorker()
}

func (w *asyncWriter) Stop() {
	w.locker.Lock()
	defer w.locker.Unlock()
	if lc, ok := w.ref.(Lifecycle); ok {
		lc.Stop()
	}
	w.isRunning = false
}

func (w *asyncWriter) DoWrite(event *LogEvent) error {
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

func (w *asyncWriter) Name() string {
	return w.name
}

func (w *asyncWriter) startWorker() {
	for {
		if !w.isRunning {
			break
		}

		p := (w.queue.Take()).(*LogEvent)
		w.write(p)
	}
}

func (w *asyncWriter) write(event *LogEvent) {
	defer event.Recycle()

	if err := w.ref.DoWrite(event); err != nil {
		Reportf("async writer write error: %v", err)
	}

	return
}
