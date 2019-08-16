// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package logrusslago

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/anbillon/slago/slago-api"
	"gitlab.com/anbillon/slago/slago-api/helpers"
)

var (
	logrusLvlToSlagoLvl = map[logrus.Level]slago.Level{
		logrus.TraceLevel: slago.TraceLevel,
		logrus.DebugLevel: slago.DebugLevel,
		logrus.InfoLevel:  slago.InfoLevel,
		logrus.WarnLevel:  slago.WarnLevel,
		logrus.ErrorLevel: slago.ErrorLevel,
		logrus.FatalLevel: slago.FatalLevel,
	}
)

type logrusBridge struct {
}

func init() {
	slago.Install(newLogrusBridge())
}

func newLogrusBridge() slago.Bridge {
	bridge := &logrusBridge{}
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: helpers.LevelFieldKey,
			logrus.FieldKeyTime:  helpers.TimestampFieldKey,
			logrus.FieldKeyMsg:   helpers.MessageFieldKey,
		},
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(bridge)

	return bridge
}

func (b *logrusBridge) Name() string {
	return "logrus"
}

func (b *logrusBridge) ParseLevel(lvl string) slago.Level {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		slago.Reportf("parse logrus level error: %s", err)
		level = logrus.TraceLevel
	}

	return logrusLvlToSlagoLvl[level]
}

func (b *logrusBridge) Write(p []byte) (int, error) {
	err := helpers.BrigeWrite(b, p)
	if err != nil {
		slago.Reportf("logrus bridge write error", err)
	}

	return len(p), err
}
