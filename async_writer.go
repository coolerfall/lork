// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
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
	"sync"
)

const defaultWriterQueueSize = 256

type asyncWriter struct {
	ref       Writer
	locker    sync.Mutex
	queue     *blockingQueue
	isStarted bool
}

// NewAsyncWriter creates a new instance of asynchronous writer.
func NewAsyncWriter(ref Writer) Writer {
	return &asyncWriter{
		ref:   ref,
		queue: NewBlockingQueue(defaultWriterQueueSize),
	}
}

func (w *asyncWriter) Start() {
	if w.isStarted {
		return
	}
	w.isStarted = true
	go w.startWorker()
}

func (w *asyncWriter) Stop() {
	w.locker.Lock()
	defer w.locker.Unlock()
	w.isStarted = false
}

func (w *asyncWriter) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.queue.RemainCapacity() <= 16 {
		// discard
		return 0, nil
	}

	w.queue.Put(p)

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
		if !w.isStarted {
			break
		}

		p := w.queue.Take()

		var err error
		if w.ref.Filter() != nil && w.ref.Filter().Do(p) {
			continue
		}

		encoded := p
		if w.ref.Encoder() != nil {
			encoded, err = w.ref.Encoder().Encode(p)
			if err != nil {
				Reportf("async writer encode error: %v", err)
				continue
			}
		}
		_, err = w.ref.Write(encoded)
		if err != nil {
			Reportf("async writer write error: %v", err)
		}
	}
}
