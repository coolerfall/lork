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

type Filter interface {
	Do() bool
}

// LevelFilter represents a level filter.
type LevelFilter struct {
	level Level
}

// NewLevelFilter creates a new instance of filter.
func NewLevelFilter(lvl Level) *LevelFilter {
	return &LevelFilter{
		level: lvl,
	}
}

// Do will do the filter.
func (f *LevelFilter) Do(level Level) bool {
	return f.level > level
}
