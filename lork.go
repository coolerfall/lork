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
	"time"
)

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

const (
	LevelFieldKey     = "level"
	TimestampFieldKey = "time"
	MessageFieldKey   = "message"
	LoggerFieldKey    = "logger_name"
	ErrorFieldKey     = "error"

	TimestampFormat   = time.RFC3339Nano
	TimeFormatRFC3339 = "2006-01-02T15:04:05.000Z07:00"

	RootLoggerName = "ROOT"
)

var (
	levelMap = map[string]Level{
		"TRACE": TraceLevel,
		"DEBUG": DebugLevel,
		"INFO":  InfoLevel,
		"WARN":  WarnLevel,
		"ERROR": ErrorLevel,
		"FATAL": FatalLevel,
		"PANIC": PanicLevel,
	}
)

type Level int8

// ILogger represents lork logging interface defination.
type ILogger interface {
	// Name returns the name of current lork logger implementation.
	Name() string

	// AddWriter add one or more writer to this logger.
	AddWriter(w ...Writer)

	// ResetWriter will remove all writers added before.
	ResetWriter()

	// SetLevel sets global level for logger.
	SetLevel(lvl Level)

	// Trace logs with trace level.
	Trace() Record

	// Debug logs with debug level.
	Debug() Record

	// Info logs with info level.
	Info() Record

	// Warn logs with warn level.
	Warn() Record

	// Error logs with error level.
	Error() Record

	// Fatal logs with fatal level.
	Fatal() Record

	// Panic logs with panic level.
	Panic() Record

	// Level logs with specified level.
	Level(lvl Level) Record

	// WriteEvent writes raw logging event.
	WriteEvent(e *LogEvent)
}

// Bridge represents bridge between other logging framework and lork logger.
type Bridge interface {
	// Name returns the name of this bridge.
	Name() string

	// ParseLevel parses the given level string into lork level.
	ParseLevel(lvl string) Level
}

var (
	loggers = make([]ILogger, 0)
	bridges = make([]Bridge, 0)

	onceLogger   sync.Once
	loggerLocker sync.Mutex
	loggerCache  map[string]ILogger
)

// Logger get a global lork logger to use. The name will only get the first one.
func Logger(name ...string) ILogger {
	onceLogger.Do(func() {
		loggerLen := len(loggers)
		if loggerLen > 1 {
			Report("multiple lork logger implementation found")
		} else if loggerLen == 0 {
			Bind(NewClassicLogger())
		}
		logger := loggers[0]

		for _, b := range bridges {
			if logger.Name() == b.Name() {
				ReportfExit("cycle logger checked, %s -> lork -> %s",
					b.Name(), logger.Name())
			}
		}

		loggerCache = make(map[string]ILogger)
		loggerCache[RootLoggerName] = logger
	})

	return findLogger(name...)
}

// LoggerC get a global lork logger with a caller package name.
// Note: this will call runtime.Caller function.
func LoggerC() ILogger {
	pkgName := PackageName(1)
	return Logger(pkgName)
}

func findLogger(name ...string) ILogger {
	loggerLocker.Lock()
	defer loggerLocker.Unlock()

	rootLogger := loggerCache[RootLoggerName]
	var realName string
	if len(name) > 0 {
		realName = name[0]
	} else {
		return rootLogger
	}

	child, ok := loggerCache[realName]
	if ok {
		return child
	}

	var i = 0
	var logger = rootLogger
	var childName string
	for {
		index := indexOfSlash(realName, i)
		if index == -1 {
			childName = realName
		} else {
			childName = realName[:index]
		}
		i = index + 1
		child, ok = loggerCache[childName]
		if !ok {
			child = newNamedLogger(childName, rootLogger, logger)
			loggerCache[childName] = child
		}
		logger = child

		if index == -1 {
			return child
		}
	}
}

// Bind binds an implementation of lork logger as output logger.
func Bind(logger ILogger) {
	loggers = append(loggers, logger)
}

// Install installs a logging framework bridge into lork. All the log of the bridge
// will be delegated to lork if the logging framework bridge was installed.
func Install(bridge Bridge) {
	bridges = append(bridges, bridge)
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	case TraceLevel:
		fallthrough
	default:
		return "TRACE"
	}
}

// ParseLevel converts a level string into lork level value.
func ParseLevel(lvl string) Level {
	level, ok := levelMap[strings.ToUpper(lvl)]
	if !ok {
		return TraceLevel
	}

	return level
}
