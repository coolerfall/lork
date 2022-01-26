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

package slago

// classicLogger represents a classic logger with name which can be used as category.
type classicLogger struct {
	name   string
	root   SlaLogger
	parent SlaLogger
	lvl    Level
}

// newClassicLogger creates a new instance of classic logger.
func newClassicLogger(name string, root SlaLogger, parent SlaLogger) SlaLogger {
	return &classicLogger{
		name:   name,
		root:   root,
		parent: parent,
		lvl:    TraceLevel,
	}
}

func (cl *classicLogger) Name() string {
	return cl.name
}

func (cl *classicLogger) AddWriter(w ...Writer) {
	cl.parent.AddWriter(w...)
}

func (cl *classicLogger) ResetWriter() {
	cl.parent.ResetWriter()
}

func (cl *classicLogger) SetLevel(lvl Level) {
	cl.lvl = lvl
}

func (cl *classicLogger) Trace() Record {
	return cl.makeRecord(TraceLevel, cl.root.Trace)
}

func (cl *classicLogger) Debug() Record {
	return cl.makeRecord(DebugLevel, cl.root.Debug)
}

func (cl *classicLogger) Info() Record {
	return cl.makeRecord(InfoLevel, cl.root.Info)
}

func (cl *classicLogger) Warn() Record {
	return cl.makeRecord(WarnLevel, cl.root.Warn)
}

func (cl *classicLogger) Error() Record {
	return cl.makeRecord(ErrorLevel, cl.root.Error)
}

func (cl *classicLogger) Fatal() Record {
	return cl.makeRecord(FatalLevel, cl.root.Fatal)
}

func (cl *classicLogger) Panic() Record {
	return cl.makeRecord(PanicLevel, cl.root.Panic)
}

func (cl *classicLogger) WriteRaw(p []byte) {
	cl.parent.WriteRaw(p)
}

func (cl *classicLogger) makeRecord(lvl Level, newRecord func() Record) Record {
	var record Record

	if cl.checkLevel(lvl) {
		record = newRecord()
	} else {
		record = newNoopRecord()
	}

	// append logger name
	return record.Str(LoggerFieldKey, cl.name)
}

func (cl *classicLogger) checkLevel(lvl Level) bool {
	if cl.lvl > lvl {
		return false
	}

	for l := cl; l != nil; {
		p, ok := l.parent.(*classicLogger)
		if !ok {
			break
		}

		if cl.lvl > l.lvl {
			return false
		}

		l = p
	}

	return true
}
