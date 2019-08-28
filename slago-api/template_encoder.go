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
	"sync"
	"text/template"
	"time"
)

const (
	colorLevel = 0

	timestampFormat = "2006-01-02 15:04:05.000"
	DefaultLayout   = `{{ clr "cyan" .timestamp }} {{ clr "lvl" .level }} ` +
		`{{ .message}}`
)

var (
	colorMap = map[string]int{
		"lvl":       colorLevel,
		"black":     colorBlack,
		"red":       colorRed,
		"green":     colorGreen,
		"yellow":    colorYellow,
		"blue":      colorBlue,
		"magenta":   colorMagenta,
		"cyan":      colorCyan,
		"white":     colorWhite,
		"blackbr":   colorBrightBlack,
		"redbr":     colorBrightRed,
		"greenbr":   colorBrightGreen,
		"yellowbr":  colorBrightYellow,
		"bluebr":    colorBrightBlue,
		"magentabr": colorBrightMagenta,
		"cyanbr":    colorBrightCyan,
		"whitebr":   colorBrightWhite,
	}

	levelColorMap = map[string]int{
		"TRACE": colorBrightWhite,
		"DEBUG": colorBlue,
		"INFO":  colorGreen,
		"WARN":  colorYellow,
		"ERROR": colorRed,
		"FATAL": colorRed,
		"PANIC": colorRed,
	}
)

// TemplateEncoder encodes logging event with template.
type TemplateEncoder struct {
	mutex sync.Mutex
	buf   *bytes.Buffer
	tpl   *template.Template
}

func NewTemplateEncoder(layout string) *TemplateEncoder {
	if len(layout) == 0 {
		layout = DefaultLayout
	}

	tplFuncMap := make(template.FuncMap)
	tplFuncMap["clr"] = func(c string, s string) string {
		color, ok := colorMap[c]
		if !ok {
			color = colorWhite
		}
		if color == colorLevel {
			color = levelColorMap[s]
		}
		return colorize(color, s)
	}
	tpl, err := template.New("").Funcs(tplFuncMap).Parse(layout)
	if err != nil {
		Reportf("process log template error: %s", err)
		os.Exit(0)
	}

	return &TemplateEncoder{
		buf: &bytes.Buffer{},
		tpl: tpl,
	}
}

func (e *TemplateEncoder) Encode(p []byte) (data []byte, err error) {
	var event map[string]interface{}
	err = json.Unmarshal(p, &event)
	if err != nil {
		return nil, err
	}

	lvl := getAndRemove(LevelFieldKey, event)
	ts := getAndRemove(TimestampFieldKey, event)
	msg := getAndRemove(MessageFieldKey, event)
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, err
	}
	level := ParseLevel(lvl)

	e.mutex.Lock()
	if err = e.tpl.Execute(e.buf, map[string]interface{}{
		"timestamp": t.Format(timestampFormat),
		"level":     level.String(),
		"message":   msg,
	}); err == nil {
		if len(msg) != 0 {
			e.buf.WriteString(" ")
		}
		for k, v := range event {
			e.buf.WriteString(fmt.Sprintf("%s=%v ", k, v))
		}
		e.buf.WriteString("\n")

		data = e.buf.Bytes()
		e.buf.Reset()
	}
	e.mutex.Unlock()

	return data, err
}
