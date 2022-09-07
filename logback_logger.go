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

type logbackLogger struct {
	multiWriter *MultiWriter
	lvl         Level
}

func NewLogbackLogger() SlaLogger {
	return &logbackLogger{
		multiWriter: NewMultiWriter(),
	}
}

func (l *logbackLogger) Name() string {
	return "github.com/coolerfall/slago"
}

func (l *logbackLogger) AddWriter(w ...Writer) {
	l.multiWriter.AddWriter(w...)
}

func (l *logbackLogger) ResetWriter() {
	l.multiWriter.Reset()
}

func (l *logbackLogger) SetLevel(lvl Level) {
	l.lvl = lvl
}

func (l *logbackLogger) Trace() Record {
	return newLogbackRecord(TraceLevel, l.multiWriter)
}

func (l *logbackLogger) Debug() Record {
	return newLogbackRecord(DebugLevel, l.multiWriter)
}

func (l *logbackLogger) Info() Record {
	return newLogbackRecord(InfoLevel, l.multiWriter)
}

func (l *logbackLogger) Warn() Record {
	return newLogbackRecord(WarnLevel, l.multiWriter)
}

func (l *logbackLogger) Error() Record {
	return newLogbackRecord(ErrorLevel, l.multiWriter)
}

func (l *logbackLogger) Fatal() Record {
	return newLogbackRecord(FatalLevel, l.multiWriter)
}

func (l *logbackLogger) Panic() Record {
	return newLogbackRecord(PanicLevel, l.multiWriter)
}

func (l *logbackLogger) WriteRaw(p []byte) {
	_, err := l.multiWriter.Write(p)
	if err != nil {
		l.Error().Err(err).Msg("write raw error")
	}
}
