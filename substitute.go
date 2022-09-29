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

type substituteProvider struct {
	factory *substituteFactory
}

func newSubstituteProvider() *substituteProvider {
	return &substituteProvider{
		factory: newSubstituteFactory(),
	}
}

func (p *substituteProvider) Name() string {
	return "github.com/coolerfall/lork"
}

func (p *substituteProvider) Prepare() {
}

func (p *substituteProvider) LoggerFactory() ILoggerFactory {
	return p.factory
}

func (p *substituteProvider) SubstLoggerFactory() *substituteFactory {
	return p.factory
}

type substituteFactory struct {
	loggers    sync.Map
	eventQueue *BlockingQueue
}

func newSubstituteFactory() *substituteFactory {
	return &substituteFactory{
		eventQueue: NewBlockingQueue(DefaultQueueSize),
	}
}

func (f *substituteFactory) Logger(name string) ILogger {
	if logger, ok := f.loggers.Load(name); ok {
		return logger.(ILogger)
	}

	logger := newSubstituteLogger(name, f.eventQueue)
	f.loggers.Store(name, logger)

	return logger
}

func (f *substituteFactory) Loggers() []*substituteLogger {
	var loggers []*substituteLogger
	f.loggers.Range(func(_, logger interface{}) bool {
		loggers = append(loggers, logger.(*substituteLogger))
		return true
	})

	return loggers
}

func (f *substituteFactory) clear() {
	f.eventQueue.Clear()
}

type substituteLogger struct {
	name             string
	delegateLogger   ILogger
	eventQueue       *BlockingQueue
	eventWriter      EventRecorder
	eventCacheLogger ILogger
}

func newSubstituteLogger(name string, eventQueue *BlockingQueue) ILogger {
	return &substituteLogger{
		eventQueue:  eventQueue,
		name:        name,
		eventWriter: newSubstituteWriter(eventQueue),
	}
}

func (l *substituteLogger) Name() string {
	return l.name
}

func (l *substituteLogger) SetLevel(Level) {
}

func (l *substituteLogger) Trace() Record {
	return l.delegate().Trace()
}

func (l *substituteLogger) Debug() Record {
	return l.delegate().Debug()
}

func (l *substituteLogger) Info() Record {
	return l.delegate().Info()
}

func (l *substituteLogger) Warn() Record {
	return l.delegate().Warn()
}

func (l *substituteLogger) Error() Record {
	return l.delegate().Error()
}

func (l *substituteLogger) Fatal() Record {
	return l.delegate().Fatal()
}

func (l *substituteLogger) Panic() Record {
	return l.delegate().Panic()
}

func (l *substituteLogger) Level(lvl Level) Record {
	return l.delegate().Level(lvl)
}

func (l *substituteLogger) Event(e *LogEvent) {
	l.delegate().Event(e)
}

func (l *substituteLogger) SetDelegate(logger ILogger) {
	l.delegateLogger = logger
}

func (l *substituteLogger) delegate() ILogger {
	if l.delegateLogger != nil {
		return l.delegateLogger
	}

	if l.eventCacheLogger == nil {
		l.eventCacheLogger = newEventCacheLogger(l.name, l.eventWriter)
	}

	return l.eventCacheLogger
}

type eventCacheLogger struct {
	name     string
	recorder EventRecorder
}

func newEventCacheLogger(name string, r EventRecorder) ILogger {
	return &eventCacheLogger{
		name:     name,
		recorder: r,
	}
}

func (l *eventCacheLogger) Name() string {
	return l.name
}

func (l *eventCacheLogger) SetLevel(Level) {
}

func (l *eventCacheLogger) Trace() Record {
	return l.newRecord(TraceLevel)
}

func (l *eventCacheLogger) Debug() Record {
	return l.newRecord(DebugLevel)
}

func (l *eventCacheLogger) Info() Record {
	return l.newRecord(InfoLevel)
}

func (l *eventCacheLogger) Warn() Record {
	return l.newRecord(WarnLevel)
}

func (l *eventCacheLogger) Error() Record {
	return l.newRecord(ErrorLevel)
}

func (l *eventCacheLogger) Fatal() Record {
	return l.newRecord(FatalLevel)
}

func (l *eventCacheLogger) Panic() Record {
	return l.newRecord(PanicLevel)
}

func (l *eventCacheLogger) Level(lvl Level) Record {
	return l.newRecord(lvl)
}

func (l *eventCacheLogger) Event(e *LogEvent) {
	if err := l.recorder.WriteEvent(e); err != nil {
		Reportf("fail to write event: %v", err)
	}
}

func (l *eventCacheLogger) newRecord(lvl Level) Record {
	return newClassicRecord(lvl, l.recorder).Str(LoggerNameFieldKey, l.name)
}

type substituteWriter struct {
	eventQueue *BlockingQueue
}

func newSubstituteWriter(eventQueue *BlockingQueue) EventRecorder {
	return &substituteWriter{
		eventQueue: eventQueue,
	}
}

func (w *substituteWriter) WriteEvent(event *LogEvent) error {
	w.eventQueue.Put(event)

	return nil
}
