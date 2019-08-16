// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package zeroslago

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/anbillon/slago/slago-api"
	"gitlab.com/anbillon/slago/slago-api/helpers"
)

var (
	zapLvlToSlagoLvl = map[zerolog.Level]slago.Level{
		zerolog.NoLevel:    slago.TraceLevel,
		zerolog.DebugLevel: slago.DebugLevel,
		zerolog.InfoLevel:  slago.InfoLevel,
		zerolog.WarnLevel:  slago.WarnLevel,
		zerolog.ErrorLevel: slago.ErrorLevel,
		zerolog.FatalLevel: slago.FatalLevel,
	}
)

type zerologBridge struct {
}

func init() {
	slago.Install(newZerologBridge())
}

func newZerologBridge() slago.Bridge {
	bridge := &zerologBridge{}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	zerolog.LevelFieldName = helpers.LevelFieldKey
	zerolog.TimestampFieldName = helpers.TimestampFieldKey
	zerolog.MessageFieldName = helpers.MessageFieldKey
	logger := zerolog.New(bridge).With().Timestamp().Logger()
	log.Logger = logger

	return bridge
}

func (b *zerologBridge) Name() string {
	return "zerolog"
}

func (b *zerologBridge) ParseLevel(lvl string) slago.Level {
	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.NoLevel
		slago.Reportf("parse zerolog level error: %s", err)
	}

	return zapLvlToSlagoLvl[level]
}

func (b *zerologBridge) Write(p []byte) (int, error) {
	err := helpers.BrigeWrite(b, p)
	if err != nil {
		slago.Reportf("zerolog bridge write error", err)
	}

	return len(p), err
}
