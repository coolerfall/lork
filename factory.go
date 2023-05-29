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

const (
	stateUninitialized = iota
	stateInitializing
	stateSuccess
	stateFail
)

var (
	factory = &loggerFactory{
		substProvider: newSubstituteProvider(),
	}
)

type loggerFactory struct {
	initialState  int
	lock          sync.Mutex
	boundProvider Provider
	substProvider *substituteProvider
	providers     []Provider
	bridges       []Bridge
}

// Load loads an implementation of lork provider.
func Load(provider Provider) {
	factory.Load(provider)
}

// Install installs a logging framework bridge into lork. All the log of the bridge
// will be delegated to lork if the logging framework bridge was installed.
func Install(bridge Bridge) {
	factory.Install(bridge)
}

// Reset will reset all providers and stop writers.
func Reset() {
	Manual().ResetWriter()
	factory.Reset()
}

func getLoggerFactory() ILoggerFactory {
	return factory.provider().LoggerFactory()
}

func (f *loggerFactory) ILoggerFactory() ILoggerFactory {
	return f.provider().LoggerFactory()
}

func (f *loggerFactory) Logger(name string) ILogger {
	return f.ILoggerFactory().Logger(name)
}

func (f *loggerFactory) Reset() {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.initialState = stateUninitialized
}

func (f *loggerFactory) provider() Provider {
	if f.initialState == stateUninitialized {
		f.lock.Lock()
		if f.initialState == stateUninitialized {
			f.initialState = stateInitializing
		}
		f.bind()
		f.lock.Unlock()
	}

	switch f.initialState {
	case stateSuccess:
		return f.boundProvider
	case stateInitializing:
		return f.substProvider
	case stateFail:
		fallthrough
	default:
		panic("fail to initialize provider")
	}
}

func (f *loggerFactory) bind() {
	length := len(f.providers)
	var provider Provider
	if length == 0 {
		provider = NewClassicProvider()
	} else if length > 0 {
		if length > 1 {
			Report("multiple lork binder found")
		}
		provider = f.providers[0]
	}

	for _, b := range f.bridges {
		if provider.Name() == b.Name() {
			ReportfExit("cycle checked, %s -> lork -> %s", b.Name(), provider.Name())
		}
	}

	f.boundProvider = provider

	provider.Prepare()
	// the provider has been initialized successfully
	f.initialState = stateSuccess

	LoggerC().Debug().Msg("lork provider has been initialized")

	// replay events
	f.fixSubstLoggers()
	f.replayEvents()
}

func (f *loggerFactory) fixSubstLoggers() {
	loggers := f.substProvider.SubstLoggerFactory().Loggers()
	for _, logger := range loggers {
		logger.SetDelegate(f.Logger(logger.Name()))
	}
}

func (f *loggerFactory) replayEvents() {
	queue := f.substProvider.SubstLoggerFactory().eventQueue

	for {
		if queue.Len() == 0 {
			break
		}

		item := queue.Take()
		if item == nil {
			continue
		}

		event := item.(*LogEvent)
		f.Logger(string(event.LoggerName())).Event(event)
	}
	f.substProvider.SubstLoggerFactory().clear()
}

func (f *loggerFactory) Load(provider Provider) {
	f.providers = append(f.providers, provider)
}

func (f *loggerFactory) Install(bridge Bridge) {
	f.bridges = append(f.bridges, bridge)
}
