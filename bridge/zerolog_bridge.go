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

package bridge

import (
	"github.com/coolerfall/lork"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	zeroLvlToLorkLvl = map[zerolog.Level]lork.Level{
		zerolog.NoLevel:    lork.TraceLevel,
		zerolog.TraceLevel: lork.TraceLevel,
		zerolog.DebugLevel: lork.DebugLevel,
		zerolog.InfoLevel:  lork.InfoLevel,
		zerolog.WarnLevel:  lork.WarnLevel,
		zerolog.ErrorLevel: lork.ErrorLevel,
		zerolog.FatalLevel: lork.FatalLevel,
	}
)

type zerologBridge struct {
}

// NewZerologBridge creates a new lork bridge for zerolog.
func NewZerologBridge() lork.Bridge {
	bridge := &zerologBridge{}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = lork.TimestampFormat
	zerolog.LevelFieldName = lork.LevelFieldKey
	zerolog.TimestampFieldName = lork.TimestampFieldKey
	zerolog.MessageFieldName = lork.MessageFieldKey
	logger := zerolog.New(bridge).With().Timestamp().Logger()
	log.Logger = logger

	return bridge
}

func (b *zerologBridge) Name() string {
	return "github.com/rs/zerolog"
}

func (b *zerologBridge) ParseLevel(lvl string) lork.Level {
	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.TraceLevel
		lork.Reportf("parse zerolog level error: %s", err)
	}

	return zeroLvlToLorkLvl[level]
}

func (b *zerologBridge) Write(p []byte) (int, error) {
	lork.Logger().WriteEvent(lork.MakeEvent(p))

	return len(p), nil
}
