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

package slago

import (
	"fmt"
	"os"

	"github.com/buger/jsonparser"
)

const (
	LevelFieldKey     = "level"
	TimestampFieldKey = "time"
	MessageFieldKey   = "message"

	TimestampFormat = "2006-01-02T15:04:05.999999999Z07:00"
	//TimestampFormat = "2006-01-02T15:04:05.000Z07:00"
)

// BrigeWrite writes data from bridge to slago logger.
func BrigeWrite(bridge Bridge, p []byte) error {
	lvl, _ := jsonparser.GetString(p, LevelFieldKey)
	msg, _ := jsonparser.GetString(p, MessageFieldKey)

	record := Logger().Level(bridge.ParseLevel(lvl))
	_ = jsonparser.ObjectEach(p, func(key []byte, value []byte,
		dataType jsonparser.ValueType, offset int) error {
		realKey := string(key)
		switch realKey {
		case LevelFieldKey:
		case TimestampFieldKey:
		case MessageFieldKey:
			record.Msg(msg)

		default:
			record.Bytes(realKey, value)
		}

		return nil
	})

	return nil
}

// Report reports message in stdout.
func Report(msg string) {
	fmt.Println(colorize(colorRed, fmt.Sprintf("slago: %v", msg)))
}

// Reportf reports message with arguments in stdout.
func Reportf(format string, args ...interface{}) {
	format = "slago: " + format
	fmt.Println(colorize(colorRed, fmt.Sprintf(format, args...)))
}

// ReportfExit reportes message with arguments in stdout and exit process.
func ReportfExit(format string, args ...interface{}) {
	Reportf(format, args)
	os.Exit(0)
}

// colorize adds ANSI color for given string.
func colorize(color int, s string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", color, s)
}
