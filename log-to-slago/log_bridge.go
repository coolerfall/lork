// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package gologslago

import (
	"log"
	"strings"

	"gitlab.com/anbillon/slago/slago-api"
)

type logBridge struct {
}

func init() {
	slago.Install(newGologBridge())
}

func newGologBridge() *logBridge {
	bridge := &logBridge{}
	log.SetOutput(bridge)
	log.SetFlags(0)

	return bridge
}

func (b *logBridge) Name() string {
	return "log"
}

func (b *logBridge) ParseLevel(lvl string) slago.Level {
	return slago.DebugLevel
}

func (b *logBridge) Write(p []byte) (n int, err error) {
	slago.Logger().Print(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
