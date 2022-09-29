// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package main

import (
	"github.com/coolerfall/lork"
)

func main() {
	lork.Install(lork.NewLogBridge())
	lork.Manual().AddWriter(lork.NewConsoleWriter(func(o *lork.ConsoleWriterOption) {
		o.Encoder = lork.NewPatternEncoder(func(opt *lork.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{32}]){magenta} : #message #fields"
		})
	}))
	reader := lork.NewSocketReader(func(o *lork.SocketReaderOption) {
		o.Path = "/ws/log"
		o.Port = 6060
	})
	reader.Start()
}
