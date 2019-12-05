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
	"strconv"
	"sync"
	"time"

	"github.com/buger/jsonparser"
)

const (
	DefaultLayout = "#color(#date{2006-01-02 15:04:05}){cyan} " +
		"#color(#level) #message #fields"
)

var (
	colorMap = map[string]int{
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
		"TRACE": colorWhite,
		"DEBUG": colorBlue,
		"INFO":  colorGreen,
		"WARN":  colorYellow,
		"ERROR": colorRed,
		"FATAL": colorRed,
		"PANIC": colorRed,
	}
)

// PatternEncoder encodes logging event with pattern.
type PatternEncoder struct {
	mutex     sync.Mutex
	buf       *bytes.Buffer
	converter Converter
}

// NewPatternEncoder creates a new instance of pattern encoder.
func NewPatternEncoder(layouts ...string) *PatternEncoder {
	var layout string
	if len(layouts) == 0 || len(layouts[0]) == 0 {
		layout = DefaultLayout
	} else {
		layout = layouts[0]
	}

	patternParser := NewPatternParser(layout)
	node, err := patternParser.Parse()
	if err != nil {
		ReportfExit("parse pattern error, %v", err)
	}

	converters := map[string]NewConverter{
		"color":   newColorConverter,
		"level":   newLevelConverter,
		"date":    newLogDateConverter,
		"message": newMessageConverter,
		"fields":  newFieldsConverter,
	}
	converter, err := NewPatternCompiler(node, converters).Compile()
	if err != nil {
		ReportfExit("compile pattern error, %v", err)
	}

	return &PatternEncoder{
		buf:       &bytes.Buffer{},
		converter: converter,
	}
}

func (pe *PatternEncoder) Encode(p []byte) (data []byte, err error) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	for c := pe.converter; c != nil; c = c.Next() {
		var result string
		// implicit conversation will allocate memory
		switch c.(type) {
		case *colorConverter:
			result = c.(*colorConverter).Convert(p)

		case *levelConverter:
			result = c.(*levelConverter).Convert(p)

		case *logDateConverter:
			result = c.(*logDateConverter).Convert(p)

		case *messageConverter:
			result = c.(*messageConverter).Convert(p)

		case *fieldsConverter:
			result = c.(*fieldsConverter).Convert(p)
		}

		pe.buf.WriteString(result)
	}
	pe.buf.WriteByte('\n')
	data = pe.buf.Bytes()
	pe.buf.Reset()

	return data, err
}

type colorConverter struct {
	next  Converter
	child Converter
	opts  []string
	buf   *bytes.Buffer
}

func newColorConverter() Converter {
	return &colorConverter{
		buf: new(bytes.Buffer),
	}
}

func (cc *colorConverter) AttatchNext(next Converter) {
	cc.next = next
}

func (cc *colorConverter) Next() Converter {
	return cc.next
}

func (cc *colorConverter) AttachChild(child Converter) {
	cc.child = child
}

func (cc *colorConverter) AttachOptions(opts []string) {
	cc.opts = opts
}

func (cc *colorConverter) Convert(event interface{}) string {
	data, ok := event.([]byte)
	if !ok {
		return "-"
	}
	level, _, _, _ := jsonparser.Get(data, LevelFieldKey)

	if len(cc.opts) != 0 {
		color, ok := colorMap[cc.opts[0]]
		if !ok {
			color = colorWhite
		}
		cc.writeColor(color)
	}

	for c := cc.child; c != nil; c = c.Next() {
		result := c.Convert(event)
		if _, ok := c.(*levelConverter); ok {
			color, ok := levelColorMap[string(level)]
			if !ok {
				color = colorWhite
			}

			cc.writeColor(color)
		}
		cc.buf.WriteString(result)
		cc.writeColorEnd()
	}

	cc.writeColorEnd()

	result := cc.buf.String()
	cc.buf.Reset()

	return result
}

func (cc *colorConverter) writeColor(color int) {
	cc.buf.WriteString("\x1b[")
	cc.buf.WriteString(strconv.Itoa(color))
	cc.buf.WriteByte('m')
}

func (cc *colorConverter) writeColorEnd() {
	cc.buf.WriteString("\x1b[0m")
}

type levelConverter struct {
	next Converter
}

func newLevelConverter() Converter {
	return &levelConverter{}
}

func (lc *levelConverter) AttatchNext(next Converter) {
	lc.next = next
}

func (lc *levelConverter) Next() Converter {
	return lc.next
}

func (lc *levelConverter) AttachChild(child Converter) {
}

func (lc *levelConverter) AttachOptions(opts []string) {
}

func (lc *levelConverter) Convert(event interface{}) string {
	data, ok := event.([]byte)
	if !ok {
		return ""
	}

	level, _, _, _ := jsonparser.Get(data, LevelFieldKey)

	return string(level)
}

type logDateConverter struct {
	next  Converter
	child Converter
	opts  []string
}

func newLogDateConverter() Converter {
	return &logDateConverter{
		opts: []string{"2006-01-02"},
	}
}

func (c *logDateConverter) AttatchNext(next Converter) {
	c.next = next
}

func (c *logDateConverter) Next() Converter {
	return c.next
}

func (c *logDateConverter) AttachChild(child Converter) {
	c.child = child
}

func (c *logDateConverter) AttachOptions(opts []string) {
	if len(opts) != 0 && len(opts[0]) != 0 {
		c.opts = opts
	}
}

func (c *logDateConverter) Convert(event interface{}) string {
	data, ok := event.([]byte)
	if !ok {
		return ""
	}

	tsValue, _, _, _ := jsonparser.Get(data, TimestampFieldKey)
	ts, err := time.Parse(TimestampFormat, string(tsValue))
	if err != nil {
		ts = time.Now()
	}
	// TODO: AppendFormat
	return ts.Format(c.opts[0])
}

type messageConverter struct {
	next Converter
}

func newMessageConverter() Converter {
	return &messageConverter{}
}

func (mc *messageConverter) AttatchNext(next Converter) {
	mc.next = next
}

func (mc *messageConverter) Next() Converter {
	return mc.next
}

func (mc *messageConverter) AttachChild(child Converter) {
}

func (mc *messageConverter) AttachOptions(opts []string) {
}

func (mc *messageConverter) Convert(event interface{}) string {
	data, ok := event.([]byte)
	if !ok {
		return "-"
	}

	message, _, _, _ := jsonparser.Get(data, MessageFieldKey)
	if len(message) == 0 {
		return "-"
	}

	return string(message)
}

type fieldsConverter struct {
	next Converter
	buf  *bytes.Buffer
}

func newFieldsConverter() Converter {
	return &fieldsConverter{
		buf: new(bytes.Buffer),
	}
}

func (fc *fieldsConverter) AttatchNext(next Converter) {
	fc.next = next
}

func (fc *fieldsConverter) Next() Converter {
	return fc.next
}

func (fc *fieldsConverter) AttachChild(child Converter) {
}

func (fc *fieldsConverter) AttachOptions(opts []string) {
}

func (fc *fieldsConverter) Convert(event interface{}) string {
	data, ok := event.([]byte)
	if !ok {
		return ""
	}

	_ = jsonparser.ObjectEach(data, func(key []byte, value []byte,
		dataType jsonparser.ValueType, offset int) error {
		jsonKey := string(key)
		switch jsonKey {
		case TimestampFieldKey:
		case LevelFieldKey:
		case MessageFieldKey:
			// do nothing for these keys

		default:
			fc.buf.Write(key)
			fc.buf.WriteString("=")
			fc.buf.Write(value)
			fc.buf.WriteByte(' ')
		}

		return nil
	})

	result := fc.buf.String()
	fc.buf.Reset()

	return result
}
