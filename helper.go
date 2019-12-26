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
	"os"
	"path/filepath"

	"github.com/buger/jsonparser"
)

const (
	LevelFieldKey     = "level"
	TimestampFieldKey = "time"
	MessageFieldKey   = "message"

	TimestampFormat = "2006-01-02T15:04:05.999999999Z07:00"
)

// BrigeWrite writes data from bridge to slago logger.
func BrigeWrite(bridge Bridge, p []byte) error {
	lvl, _ := jsonparser.GetString(p, LevelFieldKey)
	msg, _ := jsonparser.GetString(p, MessageFieldKey)

	record := Logger().Level(bridge.ParseLevel(lvl))
	_ = jsonparser.ObjectEach(p, func(key []byte, value []byte,
		dataType jsonparser.ValueType, _ int) error {
		realKey := string(key)
		switch realKey {
		case LevelFieldKey:
		case TimestampFieldKey:
			// do nothing

		case MessageFieldKey:
			record.Msg(msg)

		default:
			record.Bytes(realKey, value)
		}

		return nil
	})

	return nil
}

// Report reports message in stdou
func Report(msg string) {
	Reportf(msg)
}

// Reportf reports message with arguments in stdou
func Reportf(format string, args ...interface{}) {
	format = "slago: " + format
	fmt.Println(colorize(colorRed, fmt.Sprintf(format, args...)))
}

// ReportfExit reportes message with arguments in stdout and exit process.
func ReportfExit(format string, args ...interface{}) {
	Reportf(format, args...)
	os.Exit(0)
}

// colorize adds ANSI color for given string.
func colorize(color int, s string) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", color, s)
}

// rename creates directory if not existed, and rename file to a new name.
func rename(oldPath, newPath string) (err error) {
	dir := filepath.Dir(newPath)
	err = os.MkdirAll(dir, os.FileMode(0666))
	if err != nil {
		return
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return
	}

	return
}

// ReplaceJson replace key/value with given search key.
func ReplaceJson(p []byte, buf *bytes.Buffer, searchKey string,
	transform func(k, v []byte) (nk, kv []byte, e error)) error {
	buf.WriteByte('{')
	var start = false
	var err error
	_ = jsonparser.ObjectEach(p, func(key []byte, value []byte,
		dataType jsonparser.ValueType, _ int) error {
		if start {
			buf.WriteByte(',')
		} else {
			start = true
		}

		if string(key) == searchKey {
			key, value, err = transform(key, value)
			if err != nil {
				return err
			}
		}

		buf.WriteByte('"')
		buf.Write(key)
		buf.WriteByte('"')
		buf.WriteByte(':')

		switch dataType {
		case jsonparser.String:
			buf.WriteByte('"')
			buf.Write(value)
			buf.WriteByte('"')

		default:
			buf.Write(value)
		}

		return nil
	})
	buf.WriteByte('}')
	buf.WriteByte('\n')

	return nil
}
