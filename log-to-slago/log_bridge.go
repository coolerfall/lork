// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package gologslago

import (
	"log"
	"strings"

	"gitlab.com/anbillon/slago/slago-api"
)

type logBridge struct {
}

// NewLogBridge creates a new slago bridge for standard log.
func NewLogBridge() *logBridge {
	bridge := &logBridge{}
	log.SetOutput(bridge)
	// clear all flags, just output message
	log.SetFlags(0)

	return bridge
}

func (b *logBridge) Name() string {
	return "log"
}

func (b *logBridge) ParseLevel(lvl string) slago.Level {
	return slago.TraceLevel
}

func (b *logBridge) Write(p []byte) (n int, err error) {
	slago.Logger().Print(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
