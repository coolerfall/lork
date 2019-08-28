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
	"bytes"
	"fmt"

	"github.com/json-iterator/go"
)

const (
	LevelFieldKey     = "level"
	TimestampFieldKey = "time"
	MessageFieldKey   = "message"

	TimestampFormat = "2006-01-02T15:04:05.000Z07:00"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func getAndRemove(key string, event map[string]interface{}) string {
	field, ok := event[key]
	if !ok {
		return ""
	}
	delete(event, key)

	return field.(string)
}

// findValue find value for specified key with given json bytes.
func findValue(p []byte, key string, valBuf *bytes.Buffer) {
	var start, end int
	var gotKey bool
	var gotColon bool
	keyBytes := []byte(key)

	for i, b := range p {
		if i == 0 {
			continue
		}
		prev := p[i-1]
		if b == '"' && prev != '\\' {
			if start == 0 {
				start = i
			} else {
				end = i
			}
		} else {
			if !gotKey && start != 0 && end != 0 {
				s := p[start+1 : end]
				if bytes.Compare(s, keyBytes) == 0 {
					gotKey = true
				}
				start = 0
				end = 0
			}

			if !gotColon && gotKey && b == ':' {
				gotColon = true
				continue
			}

			if gotKey && gotColon {
				if (start != 0 && end != 0 || start == 0) && (b == ',' || b == '}') {
					break
				}

				valBuf.WriteByte(b)
			}
		}
	}
}

// BrigeWrite writes data from bridge to slago logger.
func BrigeWrite(bridge Bridge, p []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(p, &event); err != nil {
		return err
	}

	lvl := getAndRemove(LevelFieldKey, event)
	msg := getAndRemove(MessageFieldKey, event)
	delete(event, TimestampFieldKey)

	record := Logger().Level(bridge.ParseLevel(lvl))
	for k, v := range event {
		record.Interface(k, v)
	}
	record.Msg(msg)

	return nil
}

// Report reports message in stdout.
func Report(msg string) {
	fmt.Printf("slago: %v\n", msg)
}

// Reportf reports message with arguments in stdout.
func Reportf(format string, args ...interface{}) {
	format = "slago: " + format
	fmt.Println(colorize(colorRed, fmt.Sprintf(format, args...)))
}

// colorize adds ANSI color for given string.
func colorize(color int, s string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", color, s)
}
