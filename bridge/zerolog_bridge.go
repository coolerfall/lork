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
	"github.com/coolerfall/slago"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	zeroLvlToSlagoLvl = map[zerolog.Level]slago.Level{
		zerolog.NoLevel:    slago.TraceLevel,
		zerolog.TraceLevel: slago.TraceLevel,
		zerolog.DebugLevel: slago.DebugLevel,
		zerolog.InfoLevel:  slago.InfoLevel,
		zerolog.WarnLevel:  slago.WarnLevel,
		zerolog.ErrorLevel: slago.ErrorLevel,
		zerolog.FatalLevel: slago.FatalLevel,
	}
)

type zerologBridge struct {
}

// NewZerologBridge creates a new slago bridge for zerolog.
func NewZerologBridge() slago.Bridge {
	bridge := &zerologBridge{}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = slago.TimestampFormat
	zerolog.LevelFieldName = slago.LevelFieldKey
	zerolog.TimestampFieldName = slago.TimestampFieldKey
	zerolog.MessageFieldName = slago.MessageFieldKey
	logger := zerolog.New(bridge).With().Timestamp().Logger()
	log.Logger = logger

	return bridge
}

func (b *zerologBridge) Name() string {
	return "github.com/rs/zerolog"
}

func (b *zerologBridge) ParseLevel(lvl string) slago.Level {
	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.TraceLevel
		slago.Reportf("parse zerolog level error: %s", err)
	}

	return zeroLvlToSlagoLvl[level]
}

func (b *zerologBridge) Write(p []byte) (int, error) {
	err := slago.BridgeWrite(b, p)
	if err != nil {
		slago.Reportf("zerolog bridge write error", err)
	}

	return len(p), err
}
