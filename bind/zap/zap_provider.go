// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package zap

import (
	"github.com/coolerfall/lork"
)

type zapProvider struct {
	*lork.BaseProvider
}

func NewZapProvider() lork.Provider {
	ctx := lork.NewLoggerContext(NewZapLogger)
	return &zapProvider{
		BaseProvider: lork.NewBaseProvider(ctx),
	}
}

func (p *zapProvider) Name() string {
	return "go.uber.org/zap"
}
