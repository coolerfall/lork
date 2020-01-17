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
	"errors"
	"strings"
)

// Filter represents a logging filter for slago.
type Filter interface {
	// Do filters the logging. True means filterd, otherwise pass through.
	Do(e *LogEvent) bool
}

// levelFilter represents a level filter.
type levelFilter struct {
	level Level
}

// NewLevelFilter creates a new instance of levelFilter.
func NewLevelFilter(lvl Level) Filter {
	return &levelFilter{
		level: lvl,
	}
}

// Do will execute the filter.
func (f *levelFilter) Do(e *LogEvent) bool {
	return f.level > e.LevelInt()
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

func (f *keywordFilter) Do(e *LogEvent) bool {
	var filtered bool
	e.Fields(func(k, v []byte, _ bool) {
		if f.compare(k, v) {
			filtered = true
			return
		}
	})

	return filtered
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
