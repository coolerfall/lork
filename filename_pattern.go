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
	"errors"
	"time"
)

var errPlace = errors.New("index pattern should be placed after date pattern")

type filenamePattern struct {
	converter Converter
}

func newFilenamePattern(pattern string) (*filenamePattern, error) {
	patternParser := newPatternParser(pattern)
	node, err := patternParser.Parse()
	if err != nil {
		return nil, err
	}

	converters := map[string]NewConverter{
		"index": newIndexConverter,
		"date":  newDateConverter,
	}
	converter, err := newPatternCompiler(node, converters).Compile()
	if err != nil {
		return nil, err
	}

	var gotIndex bool
	for c := converter; c != nil; c = c.Next() {
		switch c.(type) {
		case *dateConverter:
			if gotIndex {
				return nil, errPlace
			}
			break

		case *indexConverter:
			gotIndex = true
		}
	}

	return &filenamePattern{
		converter: converter,
	}, nil
}

func (fp *filenamePattern) headConverter() Converter {
	return fp.converter
}

func (fp *filenamePattern) convert(current time.Time, index int) string {
	buf := new(bytes.Buffer)

	for c := fp.converter; c != nil; c = c.Next() {
		switch c.(type) {
		case *literalConverter:
			c.Convert(nil, buf)

		case *dateConverter:
			ts := current.Format(time.RFC3339)
			c.Convert([]byte(ts), buf)

		case *indexConverter:
			c.Convert(index, buf)
		}
	}

	return buf.String()
}

func (fp *filenamePattern) toFilenameRegex() string {
	return fp.toFilenameRegexForFixed(time.Now())
}

func (fp *filenamePattern) toFilenameRegexForFixed(t time.Time) string {
	var buf = new(bytes.Buffer)
	for c := fp.converter; c != nil; c = c.Next() {
		switch c.(type) {
		case *literalConverter:
			c.Convert(nil, buf)

		case *dateConverter:
			ts := t.Format(time.RFC3339)
			c.Convert([]byte(ts), buf)

		case *indexConverter:
			buf.WriteString("(\\d{1,3})")
		}
	}

	reg := buf.String()
	buf.Reset()

	return reg
}

func (fp *filenamePattern) hasIndexConverter() bool {
	for c := fp.converter; c != nil; c = c.Next() {
		if _, ok := c.(*indexConverter); ok {
			return true
		}
	}

	return false
}

func (fp *filenamePattern) datePattern() string {
	for c := fp.converter; c != nil; c = c.Next() {
		if dc, ok := c.(*dateConverter); ok {
			return dc.DatePattern()
		}
	}

	return ""
}
