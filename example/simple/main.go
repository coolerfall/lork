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

	"github.com/coolerfall/slago"
	logrusb "github.com/coolerfall/slago/binder/logrus"
	zapb "github.com/coolerfall/slago/binder/zap"
	"github.com/coolerfall/slago/binder/zero"
	"github.com/coolerfall/slago/bridge"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func main() {
	var binderName string
	flag.StringVar(&binderName, "b", "builtin", "")
	flag.Parse()

	switch binderName {
	case "logrus":
		slago.Bind(logrusb.NewLogrusLogger())
		slago.Install(bridge.NewZerologBridge())
		slago.Install(bridge.NewZapBrige())

	case "zerolog":
		slago.Bind(zero.NewZeroLogger())
		slago.Install(bridge.NewLogrusBridge())
		slago.Install(bridge.NewZapBrige())

	case "zap":
		slago.Bind(zapb.NewZapLogger())
		slago.Install(bridge.NewLogrusBridge())
		slago.Install(bridge.NewZerologBridge())

	default:
		slago.Bind(slago.NewClassicLogger())
		slago.Install(bridge.NewLogrusBridge())
		slago.Install(bridge.NewZapBrige())
		slago.Install(bridge.NewZerologBridge())
	}

	slago.Install(slago.NewLogBridge())

	slago.Logger().AddWriter(slago.NewConsoleWriter(func(o *slago.ConsoleWriterOption) {
		o.Encoder = slago.NewPatternEncoder(func(opt *slago.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{16}]){magenta} : #message #fields"
		})
	}))
	fw := slago.NewFileWriter(func(o *slago.FileWriterOption) {
		o.Encoder = slago.NewJsonEncoder()
		o.Filter = slago.NewLevelFilter(slago.DebugLevel)
		o.Filename = "/tmp/slago/slago-test.log"
		o.RollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
			func(o *slago.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "/tmp/slago/slago-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
				o.MaxHistory = 10
			})
	})
	aw := slago.NewAsyncWriter(func(o *slago.AsyncWriterOption) {
		o.Ref = fw
	})
	slago.Logger().AddWriter(aw)

	slago.Logger().Info().Msgf("bind with: %s", slago.Logger().Name())
	slago.Logger().Trace().Msg("slago\nThis is a message \n\n")
	slago.Logger("github.com/coolerfall/slago/foo").Info().Int("int", 88).
		Any("slago", "val").Msge()
	logrus.WithField("logrus", "yes").Errorln("this is from logrus")
	zap.L().With().Warn("this is zap")
	log.Printf("this is builtin logger\n\n")

	logger := slago.Logger("github.com/slago")
	logger.Debug().Msg("slago sub logger")
	logger.SetLevel(slago.InfoLevel)
	logger.Trace().Msg("this will not print")
	logger.Info().Any("any", map[string]interface{}{
		"name": "dog",
		"age":  2,
	}).Msg("this is interface")
	slago.LoggerC().Info().Bytes("sss", []byte("ABCK")).Msg("test for auto logger name")

	time.Sleep(time.Second * 5)
}
