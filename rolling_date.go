// Copyright (c) 2019-2020 Vincent Cheung (coolingfall@gmail.com).
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
	"time"
)

const (
	topOfSecond periodicType = iota + 1
	topOfMinute
	topOfHour
	topOfDay
	topOfMonth
)

var (
	periods = []periodicType{
		topOfSecond, topOfMinute, topOfHour, topOfDay, topOfMonth,
	}
)

type periodicType int8

type rollingDate struct {
	datePattern string
	_type       periodicType
}

func newRollingDate(datePattern string) *rollingDate {
	rd := &rollingDate{
		datePattern: datePattern,
	}
	rd._type = rd.calcPeriodType()

	return rd
}

func (rd *rollingDate) calcPeriodType() periodicType {
	now := time.Now()
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
			now.Year(), now.Month()+time.Month(periods), 0, 0, 0, 0, 0, now.Location())
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

func (rd *rollingDate) next() time.Time {
	return rd.endOfNextNPeriod(time.Now(), 1)
}

func (rd *rollingDate) periodCrossed(start int64, end int64) int {
	diff := end - start
	switch rd._type {
	case topOfMinute:
		return int(diff / secondsInOneMinite)

	case topOfHour:
		return int(diff / secondsInOneHour)

	case topOfDay:
		return int(diff / secondsInOneDay)

	case topOfMonth:
		startTime := time.Unix(start, 0)
		endTime := time.Unix(end, 0)
		yearDiff := endTime.Year() - startTime.Year()
		monthDiff := endTime.Month() - startTime.Month()
		return yearDiff*12 + int(monthDiff)

	default:
		return 0
	}
}
