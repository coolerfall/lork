// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package main

import (
	"net/url"
	"time"

	"github.com/coolerfall/lork"
)

func main() {
	lork.Install(lork.NewLogBridge())
	lork.Bind(lork.NewClassicLogger())

	sw := lork.NewSocketWriter(func(o *lork.SocketWriterOption) {
		o.RemoteUrl, _ = url.Parse("ws://localhost:6060/ws/log")
	})
	lork.Logger().AddWriter(lork.NewConsoleWriter(func(o *lork.ConsoleWriterOption) {
		o.Encoder = lork.NewPatternEncoder(func(opt *lork.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{16}]){magenta} : #message #fields"
		})
	}))
	lork.Logger().AddWriter(sw)

	for now := range time.Tick(5 * time.Second) {
		lork.Logger("lork/example/socket/client").Info().Time("tick", now).
			Msg("This is log from client")
	}
}
