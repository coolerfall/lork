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

type classicLogger struct {
	name        string
	multiWriter *MultiWriter
}

// NewClassicLogger create a classic ILogger. This logger is a builtin implementation.
func NewClassicLogger(name string, writer *MultiWriter) ILogger {
	return &classicLogger{
		name:        name,
		multiWriter: writer,
	}
}

func (l *classicLogger) Name() string {
	return l.name
}

func (l *classicLogger) SetLevel(Level) {
}

func (l *classicLogger) Trace() Record {
	return l.makeRecord(TraceLevel)
}

func (l *classicLogger) Debug() Record {
	return l.makeRecord(DebugLevel)
}

func (l *classicLogger) Info() Record {
	return l.makeRecord(InfoLevel)
}

func (l *classicLogger) Warn() Record {
	return l.makeRecord(WarnLevel)
}

func (l *classicLogger) Error() Record {
	return l.makeRecord(ErrorLevel)
}

func (l *classicLogger) Fatal() Record {
	return l.makeRecord(FatalLevel)
}

func (l *classicLogger) Panic() Record {
	return l.makeRecord(PanicLevel)
}

func (l *classicLogger) Level(lvl Level) Record {
	return l.makeRecord(lvl)
}

func (l *classicLogger) Event(e *LogEvent) {
	err := l.multiWriter.WriteEvent(e)
	if err != nil {
		l.Error().Err(err).Msg("write raw event error")
	}
}

func (l *classicLogger) makeRecord(lvl Level) Record {
	return newClassicRecord(lvl, l.multiWriter)
}
