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

package main

import (
	"flag"
	"log"
	"time"

	"github.com/coolerfall/lork"
	logrusb "github.com/coolerfall/lork/binder/logrus"
	zapb "github.com/coolerfall/lork/binder/zap"
	"github.com/coolerfall/lork/binder/zero"
	"github.com/coolerfall/lork/bridge"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func main() {
	var binderName string
	flag.StringVar(&binderName, "b", "builtin", "")
	flag.Parse()

	switch binderName {
	case "logrus":
		lork.Bind(logrusb.NewLogrusLogger())
		lork.Install(bridge.NewZerologBridge())
		lork.Install(bridge.NewZapBrige())

	case "zerolog":
		lork.Bind(zero.NewZeroLogger())
		lork.Install(bridge.NewLogrusBridge())
		lork.Install(bridge.NewZapBrige())

	case "zap":
		lork.Bind(zapb.NewZapLogger())
		lork.Install(bridge.NewLogrusBridge())
		lork.Install(bridge.NewZerologBridge())

	default:
		lork.Bind(lork.NewClassicLogger())
		lork.Install(bridge.NewLogrusBridge())
		lork.Install(bridge.NewZapBrige())
		lork.Install(bridge.NewZerologBridge())
	}

	lork.Install(lork.NewLogBridge())

	lork.Logger().AddWriter(lork.NewConsoleWriter(func(o *lork.ConsoleWriterOption) {
		o.Encoder = lork.NewPatternEncoder(func(opt *lork.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{36}]){magenta} : #message #fields"
		})
	}))
	fw := lork.NewFileWriter(func(o *lork.FileWriterOption) {
		o.Encoder = lork.NewJsonEncoder()
		o.Filter = lork.NewThresholdFilter(lork.DebugLevel)
		o.Filename = "/tmp/lork/lork-test.log"
		o.RollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
			func(o *lork.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "/tmp/lork/lork-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
				o.MaxHistory = 10
			})
	})
	aw := lork.NewAsyncWriter(func(o *lork.AsyncWriterOption) {
		o.Ref = fw
	})
	lork.Logger().AddWriter(aw)

	lork.Logger().Info().Msgf("bind with: %s", lork.Logger().Name())
	lork.Logger().Trace().Msg("lork\nThis is a message with new line \n\n")
	lork.Logger("github.com/coolerfall/lork/foo").Info().Int("int", 88).
		Any("lork", "val").Msge()
	logrus.WithField("logrus", "yes").Errorln("this is from logrus")
	zap.L().With().Warn("this is zap")
	log.Printf("this is builtin logger\n\n")

	logger := lork.Logger("github.com/lork")
	logger.Debug().Msg("lork sub logger")
	logger.SetLevel(lork.InfoLevel)
	logger.Trace().Msg("this will not print")
	logger.Info().Any("any", map[string]interface{}{
		"name": "dog",
		"age":  2,
	}).Msg("this is interface")
	lork.LoggerC().Info().Bytes("sss", []byte("ABCK")).Msg("test for auto logger name")

	time.Sleep(time.Second * 5)
}
