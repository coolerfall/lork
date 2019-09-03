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
	"time"
)

type Converter interface {
	AttatchNext(next Converter)
	Next() Converter
	AttachChild(child Converter)
	AttachOptions(opts []string)
	Convert(event interface{}) string
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

func (c *literalConverter) Convert(event interface{}) string {
	return c.value
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

func (c *dateConverter) Convert(event interface{}) string {
	t, ok := event.(time.Time)
	if !ok {
		return ""
	}

	return t.Format(c.opts[0])
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

func (c *indexConverter) Convert(e interface{}) string {
	if _, ok := e.(int); !ok {
		return ""
	}
	return fmt.Sprintf("%v", e)
}
