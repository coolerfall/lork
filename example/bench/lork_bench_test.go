// Copyright (c) 2019-2023 Vincent Cheung (coolingfall@gmail.com).
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
	"sync"
	"testing"
	"time"

	"github.com/coolerfall/lork"
)

func init() {

}

var (
	longStr = "this is super long long long long long long long text from lork to hello world"
	strs    = []string{"hello world", "hello go"}
	ints    = []int{5, 1, 2}
	bools   = []bool{true, false, true}
	bytes   = []byte{0x36, 0x37, 0x88}
	t       = time.Now()
	times   = []time.Time{time.Now(), time.Now()}
	d       = time.Second * 13
	ds      = []time.Duration{time.Second * 14, time.Minute * 2}

	rollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
		func(o *lork.SizeAndTimeBasedRPOption) {
			o.FilenamePattern = "/tmp/lork/lork-archive.#date{2006-01-02}.#index.log"
			o.MaxFileSize = "50MB"
		})
)

func init() {
	lork.Load(lork.NewClassicProvider())
}

var onceJson sync.Once

func BenchmarkJsonFileWriter(b *testing.B) {
	onceJson.Do(func() {
		fw := lork.NewFileWriter(func(o *lork.FileWriterOption) {
			o.Encoder = lork.NewJsonEncoder()
			o.Filename = "/tmp/lork/lork-test.log"
			o.RollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
				func(o *lork.SizeAndTimeBasedRPOption) {
					o.FilenamePattern = "/tmp/lork/lork-archive.#date{2006-01-02}.#index.log"
					o.MaxFileSize = "10MB"
				})
		})

		lork.Manual().AddWriter(fw)
	})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lork.Logger("github.com/coolerfall/lork/bench").
				Info().
				Bytes("bytes", bytes).
				Int("int", 88888).Ints("ints", ints).
				Bool("bool", true).Bools("bools", bools).
				Float32("float32", 9999.1).Uint("uint", 999).
				Time("time", t).Times("times", times).
				Dur("dur", d).Durs("durs", ds).
				Str("str", longStr).Strs("strs", strs).
				Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}

var oncePattern sync.Once

func BenchmarkPatternFileWriter(b *testing.B) {
	oncePattern.Do(func() {
		lork.Reset()
		fw := lork.NewFileWriter(func(o *lork.FileWriterOption) {
			o.Encoder = lork.NewPatternEncoder(func(opt *lork.PatternEncoderOption) {
				opt.Pattern = "#date{2006-01-02} #level #message #fields"
			})
			o.Filename = "/tmp/lork/lork-test.log"
			o.RollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
				func(o *lork.SizeAndTimeBasedRPOption) {
					o.FilenamePattern = "/tmp/lork/lork-archive.#date{2006-01-02}.#index.log"
					o.MaxFileSize = "10MB"
				})
		})

		lork.Manual().AddWriter(fw)
	})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lork.Logger("github.com/coolerfall/lork/bench").
				Info().
				Bytes("bytes", bytes).
				Int("int", 88888).Ints("ints", ints).
				Bool("bool", true).Bools("bools", bools).
				Float32("float32", 9999.1).Uint("uint", 999).
				Time("time", t).Times("times", times).
				Dur("dur", d).Durs("durs", ds).
				Str("str", longStr).Strs("strs", strs).
				Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}

var onceAsync sync.Once

func BenchmarkAsyncFileWriter(b *testing.B) {
	onceAsync.Do(func() {
		lork.Reset()
		fw := lork.NewFileWriter(func(o *lork.FileWriterOption) {
			o.Encoder = lork.NewJsonEncoder()
			o.Filename = "/tmp/lork/lork-test.log"
			o.RollingPolicy = rollingPolicy
		})

		aw := lork.NewAsyncWriter(func(o *lork.AsyncWriterOption) {
			o.Name = "ASYNC"
		})
		aw.AddWriter(fw)
		lork.Manual().AddWriter(aw)
	})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lork.Logger("github.com/coolerfall/lork/bench").
				Info().
				Bytes("bytes", bytes).
				Int("int", 88888).Ints("ints", ints).
				Bool("bool", true).Bools("bools", bools).
				Float32("float32", 9999.1).Uint("uint", 999).
				Time("time", t).Times("times", times).
				Dur("dur", d).Durs("durs", ds).
				Str("str", longStr).Strs("strs", strs).
				Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}

var onceNone sync.Once

func BenchmarkNoWriter(b *testing.B) {
	onceNone.Do(func() {
		lork.Reset()
		lork.Manual().AddWriter(&DiscardWriter{})
	})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lork.Logger("github.com/coolerfall/lork/bench").
				Info().
				Bytes("bytes", bytes).
				Int("int", 88888).Ints("ints", ints).
				Bool("bool", true).Bools("bools", bools).
				Float32("float32", 9999.1).Uint("uint", 999).
				Time("time", t).Times("times", times).
				Dur("dur", d).Durs("durs", ds).
				Str("str", longStr).Strs("strs", strs).
				Msg("The quick brown fox jumps over the lazy dog")
		}
	})
}

type DiscardWriter struct {
}

func (w *DiscardWriter) DoWrite(*lork.LogEvent) error {
	return nil
}

func (w *DiscardWriter) Name() string {
	return ""
}
