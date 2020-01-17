// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

const (
	LevelFieldKey     = "level"
	TimestampFieldKey = "time"
	MessageFieldKey   = "message"
	LoggerFieldKey    = "logger_name"

	TimestampFormat = time.RFC3339Nano

	Slash = "/"
)

// BrigeWrite writes data from bridge to slago logger.
func BrigeWrite(bridge Bridge, p []byte) error {
	lvl, _ := jsonparser.GetString(p, LevelFieldKey)
	msg, _ := jsonparser.GetString(p, MessageFieldKey)

	record := makeRecord(bridge.ParseLevel(lvl))
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

func makeRecord(lvl Level) Record {
	switch lvl {
	case DebugLevel:
		return Logger().Debug()
	case InfoLevel:
		return Logger().Info()
	case WarnLevel:
		return Logger().Warn()
	case ErrorLevel:
		return Logger().Error()
	case FatalLevel:
		return Logger().Fatal()
	case PanicLevel:
		return Logger().Panic()
	case TraceLevel:
		fallthrough
	default:
		return Logger().Trace()
	}
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

// indexOfSlash gets the position of slash, starting at fromIndex.
func indexOfSlash(name string, fromIndex int) int {
	if len(name) < fromIndex || fromIndex < 0 {
		return -1
	}

	var sub = name
	if fromIndex > 0 {
		sub = name[fromIndex:]
	}

	i := strings.Index(sub, Slash)
	if i < 0 {
		return i
	}

	return fromIndex + i
}

// rename creates directory if not existed, and rename file to a new name.
func rename(oldPath, newFilename string) (err error) {
	dir := filepath.Dir(oldPath)
	err = os.MkdirAll(dir, os.FileMode(0666))
	if err != nil {
		return
	}

	newPath := filepath.Join(dir, newFilename)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return
	}

	return
}

// ReplaceJson replaces key/value with given search key.
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
