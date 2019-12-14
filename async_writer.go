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
	"bytes"
	"sync"
)

const defaultQueueSize = 256

type asyncWriter struct {
	ref       Writer
	locker    sync.Mutex
	queue     *blockingQueue
	buf       *bytes.Buffer
	isStarted bool
}

// NewAsyncWriter creates a new instance of asynchronous writer.
func NewAsyncWriter(ref Writer) *asyncWriter {
	return &asyncWriter{
		ref:   ref,
		queue: NewBlockingQueue(defaultQueueSize),
		buf:   new(bytes.Buffer),
	}
}

func (w *asyncWriter) Start() {
	if w.isStarted {
		return
	}
	go w.startWorker()
	w.isStarted = true
}

func (w *asyncWriter) Write(p []byte) (n int, err error) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.queue.RemainCapacity() <= 16 {
		// discard
		return 0, nil
	}

	w.buf.Write(p)
	w.queue.Put(w.buf.String())
	w.buf.Reset()

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
		p := []byte(w.queue.Take())

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

type blockingQueue struct {
	locker   *sync.Mutex
	notEmpty *sync.Cond
	items    []string

	count     int
	takeIndex int
	putIndex  int
}

func NewBlockingQueue(capacity int) *blockingQueue {
	lock := new(sync.Mutex)
	return &blockingQueue{
		locker:   lock,
		notEmpty: sync.NewCond(lock),
		items:    make([]string, capacity),
	}
}

func (q *blockingQueue) RemainCapacity() int {
	q.locker.Lock()
	defer q.locker.Unlock()

	return len(q.items) - q.count
}

func (q *blockingQueue) Put(item string) {
	q.locker.Lock()
	defer q.locker.Unlock()

	q.items[q.putIndex] = item
	q.putIndex++
	if q.putIndex == len(q.items) {
		q.putIndex = 0
	}
	q.count++

	q.notEmpty.Signal()
}

func (q *blockingQueue) Take() string {
	q.locker.Lock()
	defer q.locker.Unlock()

	for q.count == 0 {
		q.notEmpty.Wait()
	}

	next := q.items[q.takeIndex]
	q.takeIndex++
	if q.takeIndex == len(q.items) {
		q.takeIndex = 0
	}
	q.count--

	return next
}
