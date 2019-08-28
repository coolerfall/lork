// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package slago

import (
	"strings"
	"sync"
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

var (
	levelMap = map[string]Level{
		"TRACE":  TraceLevel,
		"DEBUG":  DebugLevel,
		"INFO":   InfoLevel,
		"WARN":   WarnLevel,
		"ERROR":  ErrorLevel,
		"FALTAL": FatalLevel,
		"PANIC":  PanicLevel,
	}
)

type Level int8

// SlaLogger represents a logging abstraction.
type SlaLogger interface {
	// Name returns the name of current slago logger implementation.
	Name() string

	// AddWriter add one or more writer to this logger.
	AddWriter(w ...Writer)

	// SetLevel sets global level for root logger.
	SetLevel(lvl Level)

	// Level logs with a level.
	Level(lvl Level) Record

	// Trace logs with trace level.
	Trace() Record

	// Trace logs with debug level.
	Debug() Record

	// Trace logs with info level.
	Info() Record

	// Trace logs with warn level.
	Warn() Record

	// Trace logs with error level.
	Error() Record

	// Trace logs with faltal level.
	Fatal() Record

	// Trace logs with panic level.
	Panic() Record

	// Print prints the given args.
	Print(args ...interface{})

	// Printf prints with given format and args.
	Printf(format string, args ...interface{})

	// WriteRaw writes raw logging event.
	WriteRaw(p []byte)
}

// Bridge represents bridge between other logging framework and slago logger.
type Bridge interface {
	// Name returns the name of this bridge.
	Name() string

	// ParseLevel parses the given level string into slago level.
	ParseLevel(lvl string) Level
}

var (
	loggers = make([]SlaLogger, 0)
	bridges = make([]Bridge, 0)

	onceLogger sync.Once
	logger     SlaLogger
)

// Logger get a global slago logger to use.
func Logger() SlaLogger {
	onceLogger.Do(func() {
		loggerLen := len(loggers)
		if loggerLen > 1 {
			Report("multiple slago logger implementation found")
		} else if loggerLen == 0 {
			Bind(newNoopLogger())
			Report("no slago logger found, default to " +
				"no-operation (NOOP) logger implementation")
		}
		logger = loggers[0]

		for _, b := range bridges {
			if logger.Name() == b.Name() {
				Reportf("cycle logger checked, %s -> slago -> %s",
					b.Name(), logger.Name())
			}
		}
	})

	return logger
}

// Bind binds an implementation of slago logger as output logger.
func Bind(logger SlaLogger) {
	loggers = append(loggers, logger)
}

// Install installs a logging framework bridge into slago. All the log of the bridge
// will be delegated to slagto if the logging framework bridge was installed.
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
		return "FALTAL"
	case PanicLevel:
		return "PANIC"
	case TraceLevel:
		fallthrough
	default:
		return "TRACE"
	}
}

// ParseLevel converts a level string into slago level value.
func ParseLevel(lvl string) Level {
	level, ok := levelMap[strings.ToUpper(lvl)]
	if !ok {
		return TraceLevel
	}

	return level
}
