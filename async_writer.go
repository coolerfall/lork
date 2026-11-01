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
	"sync"
)

type AsyncWriter struct {
	name        string
	locker      sync.Mutex
	queue       chan *LogEvent
	stop        chan struct{}
	done        chan struct{}
	producers   sync.WaitGroup
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
	if opts.QueueSize == 0 {
		opts.QueueSize = DefaultQueueSize
	}

	return &AsyncWriter{
		name:        opts.Name,
		queue:       make(chan *LogEvent, opts.QueueSize),
		stop:        make(chan struct{}),
		multiWriter: NewMultiWriter(),
	}
}

// Start begins consuming events, restarting retained downstream writers after Stop.
func (w *AsyncWriter) Start() {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.isRunning {
		return
	}
	select {
	case <-w.stop:
		w.queue = make(chan *LogEvent, cap(w.queue))
		w.stop = make(chan struct{})
		for _, writer := range w.multiWriter.writers {
			if lc, ok := writer.(Lifecycle); ok {
				lc.Start()
			}
		}
	default:
	}
	w.done = make(chan struct{})
	w.isRunning = true
	go w.startWorker()
}

// Stop drains pending events and stops downstream writers, retaining their bindings.
func (w *AsyncWriter) Stop() {
	w.locker.Lock()
	defer w.locker.Unlock()

	select {
	case <-w.stop:
		return
	default:
		close(w.stop)
	}
	// Stop accepting events before waiting for in-flight enqueues.
	w.producers.Wait()
	close(w.queue)
	if w.isRunning {
		<-w.done
		w.isRunning = false
	} else {
		// Events can also be queued before Start.
		for event := range w.queue {
			w.writeEvent(event)
		}
	}
	for _, writer := range w.multiWriter.writers {
		if lc, ok := writer.(Lifecycle); ok {
			lc.Stop()
		}
	}
}

func (w *AsyncWriter) DoWrite(event *LogEvent) error {
	w.locker.Lock()
	select {
	case <-w.stop:
		w.locker.Unlock()
		return nil
	default:
	}
	queue := w.queue
	w.producers.Add(1)
	w.locker.Unlock()
	defer w.producers.Done()

	if event.LevelInt() <= InfoLevel && cap(queue)-len(queue) < cap(queue)/5 {
		return nil
	}

	// copy a log event for further usage
	copied := event.Copy()
	select {
	case queue <- copied:
	default:
		// A full queue discards every level without waiting for space.
		copied.Recycle()
	}

	return nil
}

func (w *AsyncWriter) Name() string {
	return w.name
}

func (w *AsyncWriter) AddWriter(writers ...Writer) {
	w.locker.Lock()
	defer w.locker.Unlock()

	select {
	case <-w.stop:
		// Newly attached writers stay stopped until the next Start.
		w.multiWriter.writers = append(w.multiWriter.writers, writers...)
	default:
		w.multiWriter.AddWriter(writers...)
	}
}

func (w *AsyncWriter) GetWriter(name string) Writer {
	return w.multiWriter.GetWriter(name)
}

func (w *AsyncWriter) Attached(writer Writer) bool {
	return w.multiWriter.Attached(writer)
}

func (w *AsyncWriter) ResetWriter() {
	w.locker.Lock()
	defer w.locker.Unlock()

	select {
	case <-w.stop:
		// Existing writers are stopped; newly attached writers have not started.
		w.multiWriter.writers = w.multiWriter.writers[:0]
	default:
		w.multiWriter.ResetWriter()
	}
}

func (w *AsyncWriter) startWorker() {
	defer close(w.done)
	for event := range w.queue {
		w.writeEvent(event)
	}
}

func (w *AsyncWriter) writeEvent(event *LogEvent) {
	if err := w.multiWriter.WriteEvent(event); err != nil {
		Reportf("async writer write error: %v", err)
	}
}
