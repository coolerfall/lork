// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package zapslago

import (
	"gitlab.com/anbillon/slago/slago-api"
	"gitlab.com/anbillon/slago/slago-api/helpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapLvlToSlagoLvl = map[zapcore.Level]slago.Level{
		zapcore.DebugLevel: slago.DebugLevel,
		zapcore.InfoLevel:  slago.InfoLevel,
		zapcore.WarnLevel:  slago.WarnLevel,
		zapcore.ErrorLevel: slago.ErrorLevel,
		zapcore.FatalLevel: slago.FatalLevel,
	}
)

type zapBridge struct {
}

func init() {
	slago.Install(newZapBrige())
}

func newZapBrige() *zapBridge {
	bridge := &zapBridge{}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = helpers.LevelFieldKey
	encoderConfig.TimeKey = helpers.TimestampFieldKey
	encoderConfig.MessageKey = helpers.MessageFieldKey
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(bridge),
		zap.DebugLevel)
	zap.ReplaceGlobals(zap.New(core))

	return bridge
}

func (b *zapBridge) Name() string {
	return "zap"
}

func (b *zapBridge) ParseLevel(lvl string) slago.Level {
	var level = zapcore.DebugLevel
	if err := (&level).UnmarshalText([]byte(lvl)); err != nil {
		slago.Reportf("parse zap level error: %s", err)
	}

	return zapLvlToSlagoLvl[level]
}

func (b *zapBridge) Write(p []byte) (int, error) {
	err := helpers.BrigeWrite(b, p)
	if err != nil {
		slago.Reportf("logrus bridge write error", err)
	}

	return len(p), err
}
