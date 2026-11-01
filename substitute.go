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
	"sync/atomic"
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
	lock       sync.Mutex
	delegate   atomic.Pointer[substituteDelegate]
}

type substituteDelegate struct {
	factory ILoggerFactory
}

func newSubstituteFactory() *substituteFactory {
	return &substituteFactory{
		eventQueue: NewBlockingQueue(DefaultQueueSize),
	}
}

func (f *substituteFactory) Logger(name string) ILogger {
	if delegate := f.delegate.Load(); delegate != nil {
		return delegate.factory.Logger(name)
	}
	f.lock.Lock()
	if delegate := f.delegate.Load(); delegate != nil {
		f.lock.Unlock()
		return delegate.factory.Logger(name)
	}
	defer f.lock.Unlock()

	if logger, ok := f.loggers.Load(name); ok {
		return logger.(ILogger)
	}

	logger := newSubstituteLogger(name, f)
	f.loggers.Store(name, logger)

	return logger
}

func (f *substituteFactory) activate(delegate ILoggerFactory) {
	f.lock.Lock()
	defer f.lock.Unlock()
	// Writers use this same lock: after publication, no event can enter the queue.
	f.delegate.Store(&substituteDelegate{factory: delegate})
}

func (f *substituteFactory) reset() {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.delegate.Store(nil)
}

func (f *substituteFactory) clear() {
	f.eventQueue.Clear()
}

type substituteLogger struct {
	name             string
	factory          *substituteFactory
	eventCacheLogger ILogger
}

func newSubstituteLogger(name string, factory *substituteFactory) ILogger {
	return &substituteLogger{
		name:             name,
		factory:          factory,
		eventCacheLogger: newEventCacheLogger(name, newSubstituteWriter(factory)),
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

func (l *substituteLogger) delegate() ILogger {
	if delegate := l.factory.delegate.Load(); delegate != nil {
		return delegate.factory.Logger(l.name)
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
	factory *substituteFactory
}

func newSubstituteWriter(factory *substituteFactory) EventRecorder {
	return &substituteWriter{
		factory: factory,
	}
}

func (w *substituteWriter) WriteEvent(event *LogEvent) error {
	f := w.factory
	f.lock.Lock()
	delegate := f.delegate.Load()
	if delegate == nil {
		if f.eventQueue.RemainCapacity() != 0 {
			f.eventQueue.Put(event)
			f.lock.Unlock()
			return nil
		}
		// Prepare may log synchronously; a full startup queue must not block it.
		f.lock.Unlock()
		event.Recycle()
		return nil
	}
	f.lock.Unlock()

	// Records created during Prepare may be completed after replay has finished.
	delegate.factory.Logger(string(event.LoggerName())).Event(event)

	return nil
}
