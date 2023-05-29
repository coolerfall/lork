// Copyright (c) 2019-2023 Vincent Cheung (coolingfall@gmail.com).
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
