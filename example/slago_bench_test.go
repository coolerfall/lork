// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package main

import (
	"testing"

	"gitlab.com/anbillon/slago/slago-api"
)

func init() {
	fw := slago.NewFileWriter(&slago.FileWriterOption{
		Encoder:  slago.NewLogstashEncoder(),
		Filter:   slago.NewFilter(slago.InfoLevel),
		Filename: "slago-test.log",
		RollingPolicy: slago.NewSizeAndTimeBasedRollingPolicy(
			`slago-arch.{{date "2006-01-02"}}.{{.index}}.log`, "5MB"),
	})

	//cw := slago.NewConsoleWriter(slago.NewTemplateEncoder(""), nil)
	slago.Logger().AddWriter(fw)
}

func BenchmarkSlagoZerolog(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slago.Logger().Info().Int("int", 88).Msg(
				"The quick brown fox jumps over the lazy dog")
			//slago.Logger().Debug().Int("int", 88).Msg(
			//	"The quick brown fox jumps over the lazy dog")
			//log.Info().Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}
