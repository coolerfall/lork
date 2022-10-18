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

// namedLogger represents a logger with name which can be used as category.
type namedLogger struct {
	name   string
	root   ILogger
	parent ILogger
	lvl    Level
	locker sync.Mutex
}

// newNamedLogger creates a new instance of named logger.
func newNamedLogger(name string, root ILogger, parent ILogger) ILogger {
	return &namedLogger{
		name:   name,
		root:   root,
		parent: parent,
		lvl:    TraceLevel,
	}
}

func (cl *namedLogger) Name() string {
	return cl.name
}

func (cl *namedLogger) AddWriter(w ...Writer) {
	cl.parent.AddWriter(w...)
}

func (cl *namedLogger) ResetWriter() {
	cl.parent.ResetWriter()
}

func (cl *namedLogger) SetLevel(lvl Level) {
	cl.locker.Lock()
	defer cl.locker.Unlock()

	cl.lvl = lvl
}

func (cl *namedLogger) Trace() Record {
	return cl.makeRecord(TraceLevel, cl.root.Trace)
}

func (cl *namedLogger) Debug() Record {
	return cl.makeRecord(DebugLevel, cl.root.Debug)
}

func (cl *namedLogger) Info() Record {
	return cl.makeRecord(InfoLevel, cl.root.Info)
}

func (cl *namedLogger) Warn() Record {
	return cl.makeRecord(WarnLevel, cl.root.Warn)
}

func (cl *namedLogger) Error() Record {
	return cl.makeRecord(ErrorLevel, cl.root.Error)
}

func (cl *namedLogger) Fatal() Record {
	return cl.makeRecord(FatalLevel, cl.root.Fatal)
}

func (cl *namedLogger) Panic() Record {
	return cl.makeRecord(PanicLevel, cl.root.Panic)
}

func (cl *namedLogger) Level(lvl Level) Record {
	switch lvl {
	case DebugLevel:
		return cl.makeRecord(lvl, cl.root.Debug)
	case InfoLevel:
		return cl.makeRecord(lvl, cl.root.Info)
	case WarnLevel:
		return cl.makeRecord(lvl, cl.root.Warn)
	case ErrorLevel:
		return cl.makeRecord(lvl, cl.root.Error)
	case FatalLevel:
		return cl.makeRecord(lvl, cl.root.Fatal)
	case PanicLevel:
		return cl.makeRecord(lvl, cl.root.Panic)
	case TraceLevel:
		fallthrough
	default:
		return cl.makeRecord(lvl, cl.root.Trace)
	}
}

func (cl *namedLogger) Event(e *LogEvent) {
	cl.parent.Event(e)
}

func (cl *namedLogger) makeRecord(lvl Level, newRecord func() Record) Record {
	var record Record

	if cl.checkLevel(lvl) {
		record = newRecord()
	} else {
		record = newNoopRecord()
	}

	// append logger name
	return record.Str(LoggerNameFieldKey, cl.name)
}

func (cl *namedLogger) checkLevel(lvl Level) bool {
	if cl.lvl > lvl {
		return false
	}

	for l := cl; l != nil; {
		p, ok := l.parent.(*namedLogger)
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
