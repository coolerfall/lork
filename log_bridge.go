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

import (
	"log"
)

type logBridge struct {
	opts *LogBridgeOption
}

type LogBridgeOption struct {
	Name  string
	Level Level
}

// NewLogBridge creates a new lork bridge for standard log.
func NewLogBridge(options ...func(*LogBridgeOption)) *logBridge {
	opts := &LogBridgeOption{
		Name:  "log",
		Level: TraceLevel,
	}
	for _, f := range options {
		f(opts)
	}

	bridge := &logBridge{
		opts: opts,
	}

	log.SetOutput(bridge)
	// clear all flags, just output message
	log.SetFlags(0)

	return bridge
}

func (b *logBridge) Name() string {
	return "log"
}

func (b *logBridge) ParseLevel(string) Level {
	return TraceLevel
}

func (b *logBridge) Write(p []byte) (n int, err error) {
	Logger(b.opts.Name).Level(b.opts.Level).Msg(string(p))

	return len(p), nil
}
