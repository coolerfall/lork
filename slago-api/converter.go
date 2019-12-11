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
	"time"
)

// Converter represents a pattern converter which will convert pattern to string.
type Converter interface {
	// AttatchNext attatches next converter to the chain.
	AttatchNext(next Converter)

	// Next gets next from the chain.
	Next() Converter

	// AttachChild attaches child converter to current converter.
	AttachChild(child Converter)

	// AttachOptions attaches options to current converter.
	AttachOptions(opts []string)

	// Convert converts given data into buffer.
	Convert(origin []byte, buf *bytes.Buffer)
}

type NewConverter func() Converter

type literalConverter struct {
	value string
	next  Converter
}

func NewLiteralConverter(value string) *literalConverter {
	return &literalConverter{
		value: value,
	}
}

func (c *literalConverter) AttatchNext(next Converter) {
	c.next = next
}

func (c *literalConverter) Next() Converter {
	return c.next
}

func (c *literalConverter) AttachChild(child Converter) {
}

func (c *literalConverter) AttachOptions(opt []string) {
}

func (c *literalConverter) Convert(origin []byte, buf *bytes.Buffer) {
	buf.WriteString(c.value)
}

type dateConverter struct {
	next Converter
	opts []string
}

func newDateConverter() Converter {
	return &dateConverter{
		opts: []string{"2006-01-02"},
	}
}

func (c *dateConverter) AttatchNext(next Converter) {
	c.next = next
}

func (c *dateConverter) Next() Converter {
	return c.next
}

func (c *dateConverter) AttachChild(child Converter) {
}

func (c *dateConverter) AttachOptions(opts []string) {
	if len(opts) != 0 && len(opts[0]) != 0 {
		c.opts = opts
	}
}

func (c *dateConverter) Convert(origin []byte, buf *bytes.Buffer) {
	var err error
	bufData := buf.Bytes()
	bufData, err = convertFormat(bufData, origin, time.RFC3339, c.opts[0])
	if err != nil {
		return
	}
	buf.Reset()
	buf.Write(bufData)
}

func (c *dateConverter) DatePattern() string {
	return c.opts[0]
}

type indexConverter struct {
	next Converter
}

func newIndexConverter() Converter {
	return &indexConverter{}
}

func (c *indexConverter) AttatchNext(next Converter) {
	c.next = next
}

func (c *indexConverter) Next() Converter {
	return c.next
}

func (c *indexConverter) AttachChild(child Converter) {
}

func (c *indexConverter) AttachOptions(opts []string) {
}

func (c *indexConverter) Convert(origin []byte, buf *bytes.Buffer) {
	index, err := atoi(origin)
	if err != nil {
		return
	}

	bufData := buf.Bytes()
	bufData = strconv.AppendInt(bufData, int64(index), 10)
	buf.Reset()
	buf.Write(bufData)
}
