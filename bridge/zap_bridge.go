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
	"time"

	"github.com/coolerfall/lork"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapLvlToLorkLvl = map[zapcore.Level]lork.Level{
		zapcore.DebugLevel: lork.DebugLevel,
		zapcore.InfoLevel:  lork.InfoLevel,
		zapcore.WarnLevel:  lork.WarnLevel,
		zapcore.ErrorLevel: lork.ErrorLevel,
		zapcore.FatalLevel: lork.FatalLevel,
	}
)

type zapBridge struct {
}

// NewZapBridge creates a new lork bridge for zap.
func NewZapBridge() *zapBridge {
	bridge := &zapBridge{}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = lork.LevelFieldKey
	encoderConfig.TimeKey = lork.TimestampFieldKey
	encoderConfig.MessageKey = lork.MessageFieldKey
	encoderConfig.EncodeTime = rf3339Encoder
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(bridge), atomicLevel)
	zap.ReplaceGlobals(zap.New(core))

	return bridge
}

func (b *zapBridge) Name() string {
	return "go.uber.org/zap"
}

func (b *zapBridge) ParseLevel(lvl string) lork.Level {
	var level = zapcore.DebugLevel
	if err := (&level).UnmarshalText([]byte(lvl)); err != nil {
		lork.Reportf("parse zap level error: %s", err)
	}

	return zapLvlToLorkLvl[level]
}

func (b *zapBridge) Write(p []byte) (int, error) {
	err := lork.BridgeWrite(b, p)
	if err != nil {
		lork.Reportf("zap bridge write error", err)
	}

	return len(p), err
}

func rf3339Encoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(lork.TimestampFormat))
}
