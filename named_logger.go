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

// namedLogger represents a logger with name which can be used as category.
type namedLogger struct {
	name  string
	level Level

	realLogger ILogger
	parent     ILogger
	children   []*namedLogger

	locker      sync.Mutex
	multiWriter *MultiWriter
}

// newNamedLogger creates a new instance of named logger.
func newNamedLogger(name string, parent ILogger, writer *MultiWriter) *namedLogger {
	nl := &namedLogger{
		name:        name,
		parent:      parent,
		level:       TraceLevel,
		multiWriter: writer,
	}
	nl.realLogger = nl.findRealLogger()

	return nl
}

func (nl *namedLogger) Name() string {
	return nl.name
}

func (nl *namedLogger) SetLevel(lvl Level) {
	nl.locker.Lock()
	defer nl.locker.Unlock()

	if nl.level == lvl {
		// nothing to do
		return
	}

	nl.level = lvl

	for _, child := range nl.children {
		child.SetLevel(lvl)
	}
}

func (nl *namedLogger) Trace() Record {
	return nl.makeRecord(TraceLevel, nl.realLogger.Trace)
}

func (nl *namedLogger) Debug() Record {
	return nl.makeRecord(DebugLevel, nl.realLogger.Debug)
}

func (nl *namedLogger) Info() Record {
	return nl.makeRecord(InfoLevel, nl.realLogger.Info)
}

func (nl *namedLogger) Warn() Record {
	return nl.makeRecord(WarnLevel, nl.realLogger.Warn)
}

func (nl *namedLogger) Error() Record {
	return nl.makeRecord(ErrorLevel, nl.realLogger.Error)
}

func (nl *namedLogger) Fatal() Record {
	return nl.makeRecord(FatalLevel, nl.realLogger.Fatal)
}

func (nl *namedLogger) Panic() Record {
	return nl.makeRecord(PanicLevel, nl.realLogger.Panic)
}

func (nl *namedLogger) Level(lvl Level) Record {
	switch lvl {
	case DebugLevel:
		return nl.makeRecord(lvl, nl.realLogger.Debug)
	case InfoLevel:
		return nl.makeRecord(lvl, nl.realLogger.Info)
	case WarnLevel:
		return nl.makeRecord(lvl, nl.realLogger.Warn)
	case ErrorLevel:
		return nl.makeRecord(lvl, nl.realLogger.Error)
	case FatalLevel:
		return nl.makeRecord(lvl, nl.realLogger.Fatal)
	case PanicLevel:
		return nl.makeRecord(lvl, nl.realLogger.Panic)
	case TraceLevel:
		fallthrough
	default:
		return nl.makeRecord(lvl, nl.realLogger.Trace)
	}
}

func (nl *namedLogger) Event(e *LogEvent) {
	nl.parent.Event(e)
}

func (nl *namedLogger) AddWriter(writers ...Writer) {
	nl.multiWriter.AddWriter(writers...)
}

func (nl *namedLogger) GetWriter(name string) Writer {
	return nl.multiWriter.GetWriter(name)
}

func (nl *namedLogger) Attached(writer Writer) bool {
	return nl.multiWriter.Attached(writer)
}

func (nl *namedLogger) ResetWriter() {
	nl.multiWriter.ResetWriter()
}

func (nl *namedLogger) CreateChild(name string) *namedLogger {
	child := newNamedLogger(name, nl, nl.multiWriter)
	nl.children = append(nl.children, child)
	child.level = nl.level

	return child
}

func (nl *namedLogger) FindChild(name string) *namedLogger {
	for _, child := range nl.children {
		if child.Name() == name {
			return child
		}
	}

	return nil
}

func (nl *namedLogger) findRealLogger() ILogger {
	p, ok := nl.parent.(*namedLogger)
	if ok {
		return p.findRealLogger()
	}

	return nl.parent
}

func (nl *namedLogger) isRootLogger() bool {
	return nl.name == RootLoggerName
}

func (nl *namedLogger) makeRecord(lvl Level, newRecord func() Record) Record {
	var record Record

	if nl.level > lvl {
		record = newNoopRecord()
	} else {
		record = newRecord()
	}

	// append logger name
	return record.Str(LoggerNameFieldKey, nl.name)
}
