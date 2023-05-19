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
	"strings"
	"sync"
)

type LoggerContext struct {
	loggerLocker sync.Mutex

	rootLogger  *namedLogger
	loggerCache map[string]*namedLogger
}

type NewLogger func(name string, writer *MultiWriter) ILogger

// NewLoggerContext creates a new instance of LoggerContext.
func NewLoggerContext(newLogger NewLogger) *LoggerContext {
	writer := NewMultiWriter()
	realLogger := newLogger(RootLoggerName, writer)
	rootLogger := newNamedLogger(RootLoggerName, realLogger, writer)
	ctx := &LoggerContext{
		rootLogger:  rootLogger,
		loggerCache: make(map[string]*namedLogger),
	}

	return ctx
}

// RealLogger gets a namedLogger with given name.
func (c *LoggerContext) RealLogger(name string) *namedLogger {
	c.loggerLocker.Lock()
	defer c.loggerLocker.Unlock()

	if strings.EqualFold(RootLoggerName, name) {
		return c.rootLogger
	}

	child, ok := c.loggerCache[name]
	if ok {
		return child
	}

	var i = 0
	var logger = c.rootLogger
	var childName string
	for {
		index := indexOfSlash(name, i)
		if index == -1 {
			childName = name
		} else {
			childName = name[:index]
		}
		i = index + 1
		child, ok = c.loggerCache[childName]
		if !ok {
			child = logger.CreateChild(childName)
			c.loggerCache[childName] = child
		}
		logger = child

		if index == -1 {
			return child
		}
	}
}

// Logger is implementation for ILoggerFactory.
func (c *LoggerContext) Logger(name string) ILogger {
	return c.RealLogger(name)
}
