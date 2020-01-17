// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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
	"time"

	"gitlab.com/anbillon/slago"
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

// NewZapBrige creates a new slago bridge for zap.
func NewZapBrige() *zapBridge {
	bridge := &zapBridge{}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = slago.LevelFieldKey
	encoderConfig.TimeKey = slago.TimestampFieldKey
	encoderConfig.MessageKey = slago.MessageFieldKey
	encoderConfig.EncodeTime = rf3339Encoder
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(bridge), atomicLevel)
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
	err := slago.BrigeWrite(b, p)
	if err != nil {
		slago.Reportf("zap bridge write error", err)
	}

	return len(p), err
}

func rf3339Encoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(slago.TimestampFormat))
}
