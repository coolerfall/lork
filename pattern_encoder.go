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
	"bytes"
	"strconv"
	"sync"
)

const (
	DefaultPattern = "#color(#date{2006-01-02T15:04:05.000}){cyan} #color(#level) " +
		"#color([#logger]){magenta} : #message #fields"
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

	levelColorMap = map[Level]int{
		TraceLevel: colorWhite,
		DebugLevel: colorBlue,
		InfoLevel:  colorGreen,
		WarnLevel:  colorYellow,
		ErrorLevel: colorBrightRed,
		FatalLevel: colorRed,
		PanicLevel: colorRed,
	}
)

// patternEncoder encodes logging event with pattern.
type patternEncoder struct {
	locker    sync.Mutex
	buf       *bytes.Buffer
	converter Converter
}

type PatternEncoderOption struct {
	Pattern    string
	Converters map[string]NewConverter
}

// NewPatternEncoder creates a new instance of pattern encoder.
func NewPatternEncoder(options ...func(*PatternEncoderOption)) Encoder {
	opts := &PatternEncoderOption{}
	for _, f := range options {
		f(opts)
	}

	var pattern = DefaultPattern
	if len(opts.Pattern) != 0 {
		pattern = opts.Pattern
	}

	patternParser := newPatternParser(pattern)
	node, err := patternParser.Parse()
	if err != nil {
		ReportfExit("parse pattern error, %v", err)
	}

	converters := map[string]NewConverter{
		"color":   newColorConverter,
		"level":   newLevelConverter,
		"date":    newLogDateConverter,
		"logger":  newLoggerNameConverter,
		"message": newMessageConverter,
		"fields":  newFieldsConverter,
	}
	for k, c := range opts.Converters {
		converters[k] = c
	}
	converter, err := newPatternCompiler(node, converters).Compile()
	if err != nil {
		ReportfExit("compile pattern error, %v", err)
	}

	return &patternEncoder{
		buf:       new(bytes.Buffer),
		converter: converter,
	}
}

func (pe *patternEncoder) Encode(e *LogEvent) (data []byte, err error) {
	pe.locker.Lock()
	defer pe.locker.Unlock()

	for c := pe.converter; c != nil; c = c.Next() {
		c.Convert(e, pe.buf)
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

func (cc *colorConverter) AttachNext(next Converter) {
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

func (cc *colorConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	e, ok := origin.(*LogEvent)
	if !ok {
		return
	}

	if len(cc.opts) != 0 {
		color, ok := colorMap[cc.opts[0]]
		if !ok {
			color = colorWhite
		}
		cc.writeColor(color)
	}

	level := e.LevelInt()
	for c := cc.child; c != nil; c = c.Next() {
		switch c.(type) {
		case *levelConverter:
			color, ok := levelColorMap[level]
			if !ok {
				color = colorWhite
			}

			cc.writeColor(color)
			c.Convert(origin, cc.buf)
			cc.writeColorEnd()

		default:
			c.Convert(origin, cc.buf)
		}
	}

	cc.writeColorEnd()

	buf.Write(cc.buf.Bytes())
	cc.buf.Reset()
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

func (lc *levelConverter) AttachNext(next Converter) {
	lc.next = next
}

func (lc *levelConverter) Next() Converter {
	return lc.next
}

func (lc *levelConverter) AttachChild(_ Converter) {
}

func (lc *levelConverter) AttachOptions(_ []string) {
}

func (lc *levelConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	e, ok := origin.(*LogEvent)
	if !ok {
		return
	}
	buf.Write(e.Level())
}

type logDateConverter struct {
	next  Converter
	child Converter
	opt   string
}

func newLogDateConverter() Converter {
	return &logDateConverter{
		opt: "2006-01-02",
	}
}

func (c *logDateConverter) AttachNext(next Converter) {
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
		c.opt = opts[0]
	}
}

func (c *logDateConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	e, ok := origin.(*LogEvent)
	if !ok {
		return
	}
	bufData := buf.Bytes()
	bufData, _ = appendFormatUnix(bufData, e.Timestamp(), c.opt)
	buf.Reset()
	buf.Write(bufData)
}

type loggerConverter struct {
	next Converter
	opt  int
}

func newLoggerNameConverter() Converter {
	return &loggerConverter{
		opt: -1,
	}
}

func (lc *loggerConverter) AttachNext(next Converter) {
	lc.next = next
}

func (lc *loggerConverter) Next() Converter {
	return lc.next
}

func (lc *loggerConverter) AttachChild(_ Converter) {
}

func (lc *loggerConverter) AttachOptions(opts []string) {
	if len(opts) == 0 {
		return
	}

	opt, err := strconv.Atoi(opts[0])
	if err != nil {
		return
	}

	lc.opt = opt
}

func (lc *loggerConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	e, ok := origin.(*LogEvent)
	if !ok {
		buf.WriteByte('-')
		return
	}

	loggerName := e.LoggerName()
	if !ok || len(loggerName) == 0 {
		buf.WriteByte('-')
		return
	}

	buf.Write(lc.abbreviator(loggerName))
}

func (lc *loggerConverter) abbreviator(name []byte) []byte {
	length := len(name)
	if lc.opt <= 0 || length <= lc.opt {
		return name
	}

	var abbr []byte
	var gotAbbr bool
	index := bytes.LastIndex(name, []byte("/"))
	if index <= 0 {
		return name
	}

	for i := 0; i < length-1; i++ {
		tmp := name[i]
		if tmp == '/' || tmp == '.' {
			if i == index || (len(abbr)+length-i+1) <= lc.opt {
				abbr = append(abbr, name[i:]...)
				break
			}

			abbr = append(abbr, name[i])
			gotAbbr = false
			continue
		}

		if gotAbbr {
			continue
		}

		abbr = append(abbr, name[i])
		gotAbbr = true
	}

	return abbr
}

type messageConverter struct {
	next Converter
}

func newMessageConverter() Converter {
	return &messageConverter{}
}

func (mc *messageConverter) AttachNext(next Converter) {
	mc.next = next
}

func (mc *messageConverter) Next() Converter {
	return mc.next
}

func (mc *messageConverter) AttachChild(_ Converter) {
}

func (mc *messageConverter) AttachOptions(_ []string) {
}

func (mc *messageConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	e, ok := origin.(*LogEvent)
	if !ok {
		buf.WriteByte('-')
		return
	}

	message := e.Message()
	if len(message) == 0 {
		buf.WriteByte('-')
		return
	}

	buf.Write(message)
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

func (fc *fieldsConverter) AttachNext(next Converter) {
	fc.next = next
}

func (fc *fieldsConverter) Next() Converter {
	return fc.next
}

func (fc *fieldsConverter) AttachChild(_ Converter) {
}

func (fc *fieldsConverter) AttachOptions(_ []string) {
}

func (fc *fieldsConverter) Convert(origin interface{}, buf *bytes.Buffer) {
	e, ok := origin.(*LogEvent)
	if !ok {
		return
	}

	_ = e.Fields(func(k, v []byte, isString bool) error {
		buf.Write(k)
		buf.WriteString("=")
		buf.Write(v)
		buf.WriteByte(' ')
		return nil
	})

	// remove last space
	buf.Truncate(buf.Len() - 1)
}
