// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
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
	"log"
	"strings"

	"gitlab.com/anbillon/slago"
)

type logBridge struct {
}

// NewLogBridge creates a new slago bridge for standard log.
func NewLogBridge() *logBridge {
	bridge := &logBridge{}
	log.SetOutput(bridge)
	// clear all flags, just output message
	log.SetFlags(0)

	return bridge
}

func (b *logBridge) Name() string {
	return "log"
}

func (b *logBridge) ParseLevel(lvl string) slago.Level {
	return slago.TraceLevel
}

func (b *logBridge) Write(p []byte) (n int, err error) {
	slago.Logger().Print(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
