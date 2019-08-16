// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package slago

type Formatter interface {
	LevelKey() string
	TimestampKey() string
	MessageKey() string
}

type Encoder interface {
	Encode() string
}

type LogstashFormatter struct {
}

type PlainTextFormatter struct {
}

type JsonFormatter struct {
}
