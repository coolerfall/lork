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
	"time"

	"github.com/coolerfall/slago"
)

func init() {
	//slago.Bind(slazero.NewZeroLogger())
	slago.Bind(slago.NewLogbackLogger())

	fw := slago.NewFileWriter(func(o *slago.FileWriterOption) {
		//o.Encoder = slago.NewPatternEncoder(func(opt *slago.PatternEncoderOption) {
		//	opt.Pattern = "#date{2006-01-02} #level #message #fields"
		//})
		o.Encoder = slago.NewJsonEncoder()
		o.Filter = slago.NewLevelFilter(slago.InfoLevel)
		o.Filename = "/tmp/slago/slago-test.log"
		o.RollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
			func(o *slago.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "/tmp/slago/slago-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
			})
	})

	aw := slago.NewAsyncWriter(func(o *slago.AsyncWriterOption) {
		o.Ref = fw
	})
	slago.Logger().AddWriter(aw)
}

var (
	longStr = "this is super long long long long long long long text from slago to hello wrold"
	strs    = []string{"hello world", "hello go"}
	ints    = []int{5, 1, 2}
	bools   = []bool{true, false, true}
	bytes   = []byte{0x36, 0x37, 0x88}
	t       = time.Now()
	times   = []time.Time{time.Now(), time.Now()}
	d       = time.Second * 13
	ds      = []time.Duration{time.Second * 14, time.Minute * 2}
)

func BenchmarkSlagoBuiltin(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slago.Logger("github.com/coolerfall/slago/bench").
				Info().
				Int("int", 88888).Ints("ints", ints).
				Bool("bool", true).Bools("bools", bools).
				Float32("float32", 9999.1).Uint("uint", 999).
				Time("timef", t).Times("times", times).
				Dur("dur", d).Durs("durs", ds).
				Str("str", longStr).Strs("strs", strs).
				Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}
