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
	"time"
)

const (
	topOfSecond periodicType = iota + 1
	topOfMinute
	topOfHour
	topOfDay
	topOfMonth
)

const (
	secondsInOneMinute = 60
	secondsInOneHour   = 60 * secondsInOneMinute
	secondsInOneDay    = 24 * secondsInOneHour
)

var (
	periods = []periodicType{
		topOfSecond, topOfMinute, topOfHour, topOfDay, topOfMonth,
	}
)

type periodicType int8

type rollingDate struct {
	datePattern      string
	_type            periodicType
	cleanupSupported bool
}

func newRollingDate(datePattern string) *rollingDate {
	rd := &rollingDate{
		datePattern: datePattern,
	}
	rd._type = rd.calcPeriodType()
	// Every field needed to identify a period must survive formatting and
	// parsing. Use non-default fields and an afternoon hour to reject missing
	// dates and ambiguous 12-hour clocks. The year rejects two-digit layouts.
	reference := time.Date(2100, time.November, 23, 17, 43, 51, 0, time.UTC)
	parsed, err := time.Parse(datePattern, reference.Format(datePattern))
	rd.cleanupSupported = err == nil && rd.periodIndex(parsed) == rd.periodIndex(reference)

	return rd
}

func (rd *rollingDate) calcPeriodType() periodicType {
	// Use a fixed baseline so a smaller period cannot cross a larger boundary.
	now := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, t := range periods {
		tl := now.Format(rd.datePattern)
		next := rd.endOfThisPeriod(t, now)
		tr := next.Format(rd.datePattern)
		if tl != tr {
			return t
		}
	}

	return topOfSecond
}

func (rd *rollingDate) _endOfNextNPeriod(
	periodicType periodicType, now time.Time, periods int) time.Time {
	switch periodicType {
	case topOfMinute:
		return time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(),
			now.Minute()+periods, 0, 0, now.Location())
	case topOfHour:
		return time.Date(
			now.Year(), now.Month(), now.Day(),
			now.Hour()+periods, 0, 0, 0, now.Location())
	case topOfDay:
		return time.Date(
			now.Year(), now.Month(), now.Day()+periods, 0, 0, 0, 0, now.Location())
	case topOfMonth:
		return time.Date(
			now.Year(), now.Month()+time.Month(periods), 1, 0, 0, 0, 0, now.Location())
	case topOfSecond:
		fallthrough
	default:
		return time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(),
			now.Minute(), now.Second()+periods, 0, now.Location())
	}
}

func (rd *rollingDate) endOfNextNPeriod(now time.Time, periods int) time.Time {
	return rd._endOfNextNPeriod(rd._type, now, periods)
}

func (rd *rollingDate) endOfThisPeriod(pt periodicType, now time.Time) time.Time {
	return rd._endOfNextNPeriod(pt, now, 1)
}

func (rd *rollingDate) next(t time.Time) time.Time {
	return rd.endOfNextNPeriod(t, 1)
}

func (rd *rollingDate) periodCrossed(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	// Count calendar boundaries in the archive's location, rather than complete
	// elapsed durations. A local day can contain 23 or 25 hours.
	return max(0, int(rd.periodIndex(end)-rd.periodIndex(start.In(end.Location()))))
}

func (rd *rollingDate) periodIndex(t time.Time) int64 {
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	switch rd._type {
	case topOfSecond:
		return time.Date(year, month, day, hour, minute, second, 0, time.UTC).Unix()
	case topOfMinute:
		return time.Date(year, month, day, hour, minute, 0, 0, time.UTC).Unix() / secondsInOneMinute
	case topOfHour:
		return time.Date(year, month, day, hour, 0, 0, 0, time.UTC).Unix() / secondsInOneHour
	case topOfDay:
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / secondsInOneDay
	case topOfMonth:
		return int64(year)*12 + int64(month-1)
	default:
		return 0
	}
}

func (rd *rollingDate) periodForCleanup(now time.Time, offset int) (time.Time, bool) {
	// Enumerate calendar labels in UTC so nonexistent local times cannot
	// normalize into another, possibly retained, period during a DST change.
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	calendar := time.Date(year, month, day, hour, minute, second, 0, time.UTC)
	calendar = rd.endOfNextNPeriod(calendar, offset)
	year, month, day = calendar.Date()
	hour, minute, second = calendar.Clock()
	local := time.Date(year, month, day, hour, minute, second, 0, now.Location())
	index := rd.periodIndex(calendar)
	if rd.periodIndex(local) == index {
		return local, true
	}
	// Some zones skip midnight or part of an hour. Use the transition itself
	// if the calendar period exists; skip periods that never occurred.
	start, end := local.ZoneBounds()
	for _, candidate := range [...]time.Time{start, end} {
		if !candidate.IsZero() && rd.periodIndex(candidate) == index {
			return candidate, true
		}
	}
	return time.Time{}, false
}
