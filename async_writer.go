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
	ref       Writer
	locker    sync.Mutex
	queue     *blockingQueue
	isRunning bool
}

// AsyncWriterOption represents available options for async writer.
type AsyncWriterOption struct {
	RefWriter Writer
	QueueSize int
}

// NewAsyncWriter creates a new instance of asynchronous writer.
func NewAsyncWriter(options ...func(*AsyncWriterOption)) Writer {
	opt := &AsyncWriterOption{
		QueueSize: defaultWriterQueueSize,
	}

	for _, f := range options {
		f(opt)
	}

	if opt.RefWriter == nil {
		ReportfExit("async writer need a referenced writer")
	}

	return &asyncWriter{
		ref:   opt.RefWriter,
		queue: NewBlockingQueue(opt.QueueSize),
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

func (w *asyncWriter) WriteEvent(event *LogEvent) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.queue.RemainCapacity() <= 16 {
		// discard
		event.Recycle()
		return 0, nil
	}

	w.queue.Put(event)

	return 0, nil
}

func (w *asyncWriter) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.queue.RemainCapacity() <= 16 {
		// discard
		return 0, nil
	}

	w.queue.Put(MakeEvent(p))

	return len(p), nil
}

func (w *asyncWriter) Encoder() Encoder {
	return nil
}

func (w *asyncWriter) Filter() Filter {
	return nil
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

	var err error
	if w.ref.Filter() != nil && w.ref.Filter().Do(event) == Deny {
		return
	}

	var encoded []byte
	if w.ref.Encoder() == nil {
		Reportf("no encoder found for reference writer")
		return
	}

	encoded, err = w.ref.Encoder().Encode(event)
	if err != nil {
		Reportf("async writer encode error: %v", err)
		return
	}

	_, err = w.ref.Write(encoded)
	if err != nil {
		Reportf("async writer write error: %v", err)
	}
}
