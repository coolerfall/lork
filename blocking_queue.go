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

const DefaultQueueSize = 512

type BlockingQueue struct {
	locker   *sync.Mutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
	items    []interface{}

	count     int
	takeIndex int
	putIndex  int
}

// NewBlockingQueue creates a new blocking queue.
func NewBlockingQueue(capacity int) *BlockingQueue {
	lock := new(sync.Mutex)
	if capacity == 0 {
		capacity = DefaultQueueSize
	}

	return &BlockingQueue{
		locker:   lock,
		notEmpty: sync.NewCond(lock),
		notFull:  sync.NewCond(lock),
		items:    make([]interface{}, capacity),
	}
}

// RemainCapacity gets remain capacity in queue.
func (q *BlockingQueue) RemainCapacity() int {
	q.locker.Lock()
	defer q.locker.Unlock()

	return len(q.items) - q.count
}

// Len gets the count in current queue.
func (q *BlockingQueue) Len() int {
	q.locker.Lock()
	defer q.locker.Unlock()

	return q.count
}

// Put puts an item into queue.
func (q *BlockingQueue) Put(item interface{}) {
	q.locker.Lock()
	defer q.locker.Unlock()

	for q.count == len(q.items) {
		q.notFull.Wait()
	}

	q.items[q.putIndex] = item
	q.putIndex++
	if q.putIndex == len(q.items) {
		q.putIndex = 0
	}
	q.count++

	q.notEmpty.Signal()
}

// Take takes an item from queue.
func (q *BlockingQueue) Take() interface{} {
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

	q.notFull.Signal()

	return next
}

// Clear clears the data in queue and reset all index.
func (q *BlockingQueue) Clear() {
	q.locker.Lock()
	defer q.locker.Unlock()

	q.items = q.items[:0]
	q.count = 0
	q.putIndex = 0
	q.takeIndex = 0
}
