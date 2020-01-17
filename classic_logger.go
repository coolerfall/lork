// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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

func (dl *classicLogger) Name() string {
	return dl.name
}

func (dl *classicLogger) AddWriter(w ...Writer) {
	dl.parent.AddWriter(w...)
}

func (dl *classicLogger) ResetWriter() {
	dl.parent.ResetWriter()
}

func (dl *classicLogger) SetLevel(lvl Level) {
	dl.lvl = lvl
}

func (dl *classicLogger) Trace() Record {
	return dl.makeRecord(TraceLevel, dl.root.Trace)
}

func (dl *classicLogger) Debug() Record {
	return dl.makeRecord(DebugLevel, dl.root.Debug)
}

func (dl *classicLogger) Info() Record {
	return dl.makeRecord(InfoLevel, dl.root.Info)
}

func (dl *classicLogger) Warn() Record {
	return dl.makeRecord(WarnLevel, dl.root.Warn)
}

func (dl *classicLogger) Error() Record {
	return dl.makeRecord(ErrorLevel, dl.root.Error)
}

func (dl *classicLogger) Fatal() Record {
	return dl.makeRecord(FatalLevel, dl.root.Fatal)
}

func (dl *classicLogger) Panic() Record {
	return dl.makeRecord(PanicLevel, dl.root.Panic)
}

func (dl *classicLogger) Print(args ...interface{}) {
	dl.parent.Print(args...)
}

func (dl *classicLogger) Printf(format string, args ...interface{}) {
	dl.parent.Printf(format, args...)
}

func (dl *classicLogger) WriteRaw(p []byte) {
	dl.parent.WriteRaw(p)
}

func (dl *classicLogger) makeRecord(lvl Level, newRecord func() Record) Record {
	var record Record

	if dl.checkLevel(lvl) {
		record = newRecord()
	} else {
		record = newNoopRecord()
	}

	// append logger name
	return record.Str(LoggerFieldKey, dl.name)
}

func (dl *classicLogger) checkLevel(lvl Level) bool {
	if dl.lvl > lvl {
		return false
	}

	for l := dl; l != nil; {
		p, ok := l.parent.(*classicLogger)
		if !ok {
			break
		}

		if dl.lvl > l.lvl {
			return false
		}

		l = p
	}

	return true
}
