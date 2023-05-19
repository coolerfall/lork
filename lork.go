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
	LevelFieldKey      = "level"
	TimestampFieldKey  = "time"
	MessageFieldKey    = "message"
	LoggerNameFieldKey = "logger_name"
	ErrorFieldKey      = "error"

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

// Level type definition for logging level.
type Level int8

// Nameable a simple interface to define an object to return name.
type Nameable interface {
	// Name returns the name of this object. The name uniquely identifies the object.
	Name() string
}

// ILogger represents lork logging interface definition.
type ILogger interface {
	Nameable

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

	// Event writes raw logging event.
	Event(e *LogEvent)
}

// ILoggerFactory represents a factory to get loggers.
type ILoggerFactory interface {
	// Logger gets a ILogger with given name
	Logger(name string) ILogger
}

// Provider provides ILoggerFactory to get logger.
type Provider interface {
	// Name the name of this provider.
	Name() string

	// Prepare prepares the provider.
	Prepare()

	// LoggerFactory gets ILoggerFactory from this provider.
	LoggerFactory() ILoggerFactory
}

// Bridge represents bridge between other logging framework and lork logger.
type Bridge interface {
	Nameable

	// ParseLevel parses the given level string into lork level.
	ParseLevel(lvl string) Level
}

// Logger gets a global lork logger to use. The name will only get the first one.
func Logger(name ...string) ILogger {
	realName := ""
	if len(name) > 0 {
		realName = name[0]
	}
	return getLoggerFactory().Logger(realName)
}

// LoggerC get a global lork logger with a caller package name.
// Note: this will call runtime.Caller function.
func LoggerC() ILogger {
	pkgName := PackageName(1)
	return getLoggerFactory().Logger(pkgName)
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
