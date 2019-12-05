// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"

	"gitlab.com/anbillon/slago/slago-api"
)

func init() {
	fw := slago.NewFileWriter(func(o *slago.FileWriterOption) {
		o.Encoder = slago.NewPatternEncoder("#date{2006-01-02} #level #message #fields")
		o.Filter = slago.NewLevelFilter(slago.InfoLevel)
		o.Filename = "slago-test.log"
		o.RollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
			func(o *slago.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "slago-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
			})
	})

	//cw := slago.NewConsoleWriter(slago.NewPatternEncoder(""), nil)
	slago.Logger().AddWriter(fw)
}

func BenchmarkSlagoZerolog(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slago.Logger().Info().Int("int", 88).Msg(
				"The quick brown fox jumps over the lazy dog")
		}
	})
}
