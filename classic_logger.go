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

type classicLogger struct {
	multiWriter *MultiWriter
}

func NewClassicLogger() ILogger {
	return &classicLogger{
		multiWriter: NewMultiWriter(),
	}
}

func (l *classicLogger) Name() string {
	return "github.com/coolerfall/lork"
}

func (l *classicLogger) AddWriter(w ...Writer) {
	l.multiWriter.AddWriter(w...)
}

func (l *classicLogger) ResetWriter() {
	l.multiWriter.Reset()
}

func (l *classicLogger) SetLevel(_ Level) {
}

func (l *classicLogger) Trace() Record {
	return newClassicRecord(TraceLevel, l.multiWriter)
}

func (l *classicLogger) Debug() Record {
	return newClassicRecord(DebugLevel, l.multiWriter)
}

func (l *classicLogger) Info() Record {
	return newClassicRecord(InfoLevel, l.multiWriter)
}

func (l *classicLogger) Warn() Record {
	return newClassicRecord(WarnLevel, l.multiWriter)
}

func (l *classicLogger) Error() Record {
	return newClassicRecord(ErrorLevel, l.multiWriter)
}

func (l *classicLogger) Fatal() Record {
	return newClassicRecord(FatalLevel, l.multiWriter)
}

func (l *classicLogger) Panic() Record {
	return newClassicRecord(PanicLevel, l.multiWriter)
}

func (l *classicLogger) Level(lvl Level) Record {
	return newClassicRecord(lvl, l.multiWriter)
}

func (l *classicLogger) WriteEvent(e *LogEvent) {
	_, err := l.multiWriter.WriteEvent(e)
	if err != nil {
		l.Error().Err(err).Msg("write raw event error")
	}
}
