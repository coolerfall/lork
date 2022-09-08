// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package main

import (
	"github.com/coolerfall/slago"
)

func main() {
	slago.Install(slago.NewLogBridge())
	slago.Logger().AddWriter(slago.NewConsoleWriter(func(o *slago.ConsoleWriterOption) {
		o.Encoder = slago.NewPatternEncoder(func(opt *slago.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{32}]){magenta} : #message #fields"
		})
	}))
	reader := slago.NewSocketReader(func(o *slago.SocketReaderOption) {
		o.Path = "/ws/log"
		o.Port = 6060
	})
	reader.Start()
}
