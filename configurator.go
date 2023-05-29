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

const (
	StatusNeutral ExecutionStatus = iota
	StatusNext
	StatusNoNext
)

type ExecutionStatus int8

// Configurator represents a configurator for logger.
type Configurator interface {
	// Configure will configure the logger with writers and return ExecutionStatus.
	Configure(ctx *LoggerContext) ExecutionStatus
}

type ManualConfigurator struct {
	isConfigured bool
	writers      []Writer
	context      *LoggerContext
}

var manual = &ManualConfigurator{}

// Manual gets ManualConfigurator to use.
func Manual() *ManualConfigurator {
	return manual
}

func (c *ManualConfigurator) Configure(ctx *LoggerContext) ExecutionStatus {
	if len(c.writers) == 0 {
		return StatusNext
	}

	ctx.RealLogger(RootLoggerName).AddWriter(c.writers...)
	c.context = ctx

	return StatusNoNext
}

func (c *ManualConfigurator) AddWriter(writers ...Writer) {
	c.writers = append(c.writers, writers...)
}

func (c *ManualConfigurator) GetWriter(name string) Writer {
	for _, w := range c.writers {
		if w.Name() == name {
			return w
		}
	}

	return nil
}

func (c *ManualConfigurator) Attached(writer Writer) bool {
	for _, w := range c.writers {
		if w == writer {
			return true
		}
	}

	return false
}

func (c *ManualConfigurator) ResetWriter() {
	c.writers = c.writers[:0]
	if c.context != nil {
		c.context.RealLogger(RootLoggerName).ResetWriter()
	}
}

type basicConfigurator struct {
}

func newBasicConfigurator() *basicConfigurator {
	return &basicConfigurator{}
}

func (c *basicConfigurator) Configure(ctx *LoggerContext) ExecutionStatus {
	cw := NewConsoleWriter(func(o *ConsoleWriterOption) {
		o.Name = "CONSOLE"
	})
	ctx.RealLogger(RootLoggerName).AddWriter(cw)

	return StatusNeutral
}
