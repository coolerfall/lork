// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package main

import (
	"testing"

	"gitlab.com/anbillon/slago/slago-api"
)

func BenchmarkSlagoZerolog(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slago.Logger().Info().Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}
