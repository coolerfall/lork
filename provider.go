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

package lork

// BaseProvider is a base Provider with rich features.
type BaseProvider struct {
	context       *LoggerContext
	configurators []Configurator
}

// NewBaseProvider creates a new BaseProvider.
func NewBaseProvider(ctx *LoggerContext) *BaseProvider {
	return &BaseProvider{
		context: ctx,
	}
}

func (p *BaseProvider) Name() string {
	panic("add name for your provider")
}

func (p *BaseProvider) Prepare() {
	p.configurators = append(p.configurators, manual)

	for _, c := range p.configurators {
		if c.Configure(p.context) == StatusNoNext {
			return
		}
	}

	// if no configurator was successful to execute, rollback to basic configurator
	newBasicConfigurator().Configure(p.context)
}

func (p *BaseProvider) LoggerFactory() ILoggerFactory {
	return p.context
}
