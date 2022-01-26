// Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
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

package bench

import (
	"testing"

	"github.com/coolerfall/slago"
	"github.com/coolerfall/slago/binder/slazero"
)

func init() {
	slago.Bind(slazero.NewZeroLogger())

	fw := slago.NewFileWriter(func(o *slago.FileWriterOption) {
		//o.Encoder = slago.NewPatternEncoder(func(opt *slago.PatternEncoderOption) {
		//	opt.Layout = "#date{2006-01-02} #level #message #fields"
		//})
		o.Encoder = slago.NewJsonEncoder()
		o.Filter = slago.NewLevelFilter(slago.InfoLevel)
		o.Filename = "slago-archive.2020-10-16.0.log"
		o.RollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
			func(o *slago.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "/tmp/log/slago/slago-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
			})
	})

	aw := slago.NewAsyncWriter(func(o *slago.AsyncWriterOption) {
		o.Ref = fw
	})
	slago.Logger().AddWriter(aw)
}

func BenchmarkSlagoZerolog(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slago.Logger().Info().Int("int", 88).Bool("bool", true).
				Float32("float32", 2.1).Uint("uint", 9).Str("str", "wrold").Msg(
				"The quick brown fox jumps over the lazy dog")
		}
	})
}
