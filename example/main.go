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
	"log"

	"github.com/sirupsen/logrus"
	"gitlab.com/anbillon/slago"
	"gitlab.com/anbillon/slago/binder/slazero"
	"gitlab.com/anbillon/slago/bridge"
	//_ "gitlab.com/anbillon/slago/zerolog-to-slago"
	//_ "gitlab.com/anbillon/slago/slalogrus"
	//_ "gitlab.com/anbillon/slago/slazap"
	"go.uber.org/zap"
)

func main() {
	slago.Install(bridge.NewLogBridge())
	slago.Install(bridge.NewLogrusBridge())
	slago.Install(bridge.NewZapBrige())
	slago.Bind(slazero.NewZeroLogger())
	//slago.Bind(slalogrus.NewLogrusLogger())

	slago.Logger().AddWriter(slago.NewConsoleWriter(func(o *slago.ConsoleWriterOption) {
		o.Encoder = slago.NewPatternEncoder(
			"#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #message #fields")
	}))
	fw := slago.NewFileWriter(func(o *slago.FileWriterOption) {
		o.Encoder = slago.NewJsonEncoder()
		//o.Encoder = slago.NewLogstashEncoder()
		o.Filter = slago.NewLevelFilter(slago.InfoLevel)
		o.Filename = "slago-test.log"
		o.RollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
			func(o *slago.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "slago-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
			})
	})
	slago.Logger().AddWriter(fw)

	slago.Logger().Trace().Msg("slago")
	slago.Logger().Info().Int("int", 88).Interface("slago", "val").Msg("")
	logrus.WithField("logrus", "yes").Errorln("this is from logrus")
	zap.L().With().Warn("this is zap")

	log.Printf("this is builtin logger")
}
