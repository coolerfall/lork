// Copyright (c) 2019-2021 Vincent Cheung (coolingfall@gmail.com).
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

type blockingQueue struct {
	locker   *sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	items    []*bytes.Buffer

	count     int
	takeIndex int
	putIndex  int
}

// NewBlockingQueue creates a new blocking queue.
func NewBlockingQueue(capacity int) *blockingQueue {
	lock := new(sync.Mutex)

	items := make([]*bytes.Buffer, capacity)
	for i := 0; i < capacity; i++ {
		items[i] = new(bytes.Buffer)
	}

	return &blockingQueue{
		locker:   lock,
		notEmpty: sync.NewCond(lock),
		notFull:  sync.NewCond(lock),
		items:    items,
	}
}

// RemainCapacity gets remain capacity in queue.
func (q *blockingQueue) RemainCapacity() int {
	q.locker.Lock()
	defer q.locker.Unlock()

	return len(q.items) - q.count
}

// Put puts an item into queue.
func (q *blockingQueue) Put(item []byte) {
	q.locker.Lock()
	defer q.locker.Unlock()

	for q.count == len(q.items) {
		q.notFull.Wait()
	}

	q.items[q.putIndex].Write(item)
	q.putIndex++
	if q.putIndex == len(q.items) {
		q.putIndex = 0
	}
	q.count++

	q.notEmpty.Signal()
}

// Take takes an item from queue.
func (q *blockingQueue) Take() []byte {
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

	data := next.Bytes()
	next.Reset()

	q.notFull.Signal()

	return data
}
