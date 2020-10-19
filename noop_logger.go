// Copyright (c) 2019-2020 Vincent Cheung (coolingfall@gmail.com).
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

type noopLogger struct {
}

// newNoopLogger creates a new instance of no-operation logger.
// This logger will write all data to /dev/null.
func newNoopLogger() SlaLogger {
	return &noopLogger{}
}

func (l *noopLogger) Name() string {
	return "noop"
}

func (l *noopLogger) AddWriter(w ...Writer) {
}

func (l *noopLogger) ResetWriter() {
}

func (l *noopLogger) SetLevel(lvl Level) {
}

func (l *noopLogger) Level(lvl Level) Record {
	return newNoopRecord()
}

func (l *noopLogger) Trace() Record {
	return newNoopRecord()
}

func (l *noopLogger) Debug() Record {
	return newNoopRecord()
}

func (l *noopLogger) Info() Record {
	return newNoopRecord()
}

func (l *noopLogger) Warn() Record {
	return newNoopRecord()
}

func (l *noopLogger) Error() Record {
	return newNoopRecord()
}

func (l *noopLogger) Fatal() Record {
	return newNoopRecord()
}

func (l *noopLogger) Panic() Record {
	return newNoopRecord()
}

func (l *noopLogger) WriteRaw(p []byte) {
}
