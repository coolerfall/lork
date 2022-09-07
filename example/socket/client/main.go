// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package main

import (
	"net/url"
	"time"

	"github.com/coolerfall/slago"
)

func main() {
	slago.Bind(slago.NewLogbackLogger())

	sw := slago.NewSocketWriter(func(o *slago.SocketWriterOption) {
		o.RemoteUrl, _ = url.Parse("ws://localhost:6060/ws/log")
	})
	slago.Logger().AddWriter(slago.NewConsoleWriter(func(o *slago.ConsoleWriterOption) {
		o.Encoder = slago.NewPatternEncoder(func(opt *slago.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{16}]){magenta} : #message #fields"
		})
	}))
	slago.Logger().AddWriter(sw)

	for now := range time.Tick(5 * time.Second) {
		slago.Logger("slago/example/socket/client").Info().Time("tick",
			now).Msg("This is log from client")
	}
}
