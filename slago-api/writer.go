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
	"io"
	"sync"
)

// Writer is the interface that wraps the io.Writer, add adds
// Encoder and Filter func for slago to ecnode and filter logs.
type Writer interface {
	io.Writer

	// Encoder returns encoder used in current writer.
	Encoder() Encoder

	// Filter returns filter used in current writer.
	Filter() *Filter
}

// SyncMultiWriter represents synchronous multi writer which implements io.Writer.
// This writer is used as output which will implement SlaLogger.
type SyncMultiWriter struct {
	writers []Writer
	mutex   sync.Mutex
	buf     *bytes.Buffer
}

// NewSyncMultiWriter creates a new synchronous multi writer.
func NewSyncMultiWriter() *SyncMultiWriter {
	return &SyncMultiWriter{
		writers: make([]Writer, 0),
		buf:     new(bytes.Buffer),
	}
}

// AddWriter adds a slago writer into sync multi writer.
func (smw *SyncMultiWriter) AddWriter(w ...Writer) {
	smw.writers = append(smw.writers, w...)
}

func (smw *SyncMultiWriter) Write(p []byte) (n int, err error) {
	smw.mutex.Lock()
	findValue(p, LevelFieldKey, smw.buf)
	level := ParseLevel(smw.buf.String())
	smw.buf.Reset()
	smw.mutex.Unlock()

	for _, w := range smw.writers {
		if w.Filter() != nil && w.Filter().Do(level) {
			return
		}

		encoded := p
		if w.Encoder() != nil {
			encoded, err = w.Encoder().Encode(p)
			if err != nil {
				return
			}
		}
		n, err = w.Write(encoded)
	}

	return len(p), nil
}
