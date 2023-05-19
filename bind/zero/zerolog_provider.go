// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package zero

import (
	"github.com/coolerfall/lork"
)

type zeroProvider struct {
	*lork.BaseProvider
}

func NewZeroProvider() lork.Provider {
	ctx := lork.NewLoggerContext(NewZeroLogger)
	return &zeroProvider{
		BaseProvider: lork.NewBaseProvider(ctx),
	}
}

func (p *zeroProvider) Name() string {
	return "github.com/rs/zerolog"
}
