// Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
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
	"errors"
	"strings"
)

// FilterReply defines the result of filter.
type FilterReply int8

const (
	// Accept represents the logging will be accepted by filter.
	Accept FilterReply = iota + 1
	// Deny represents the logging will be denied by filter.
	Deny
)

// Filter represents a logging filter for lork.
type Filter interface {
	// Do filters the logging. The logs will be accepted or denied according to FilterReply.
	Do(e *LogEvent) FilterReply
}

// thresholdFilter represents a filter with threshold.
type thresholdFilter struct {
	level Level
}

// NewThresholdFilter creates a new instance of thresholdFilter.
func NewThresholdFilter(lvl Level) Filter {
	return &thresholdFilter{
		level: lvl,
	}
}

func (f *thresholdFilter) Do(e *LogEvent) FilterReply {
	if e.LevelInt() >= f.level {
		return Accept
	}

	return Deny
}

// keywordFilter represents a filter by key word rule.
type keywordFilter struct {
	keywords []string
}

// NewKeywordFilter creates a new instance of keywordFilter.
func NewKeywordFilter(keywords ...string) Filter {
	return &keywordFilter{
		keywords: keywords,
	}
}

var errFound = errors.New("found")

func (f *keywordFilter) Do(e *LogEvent) FilterReply {
	err := e.Fields(func(k, v []byte, _ bool) error {
		if f.compare(k, v) {
			return errFound
		}

		return nil
	})

	if err == errFound {
		return Accept
	}

	return Deny
}

func (f *keywordFilter) compare(key []byte, value []byte) bool {
	for _, keyword := range f.keywords {
		if strings.Contains(keyword, "=") &&
			keyword == string(key)+"="+string(value) {
			return true
		}

		if keyword == string(key) || keyword == string(value) {
			return true
		}
	}

	return false
}
