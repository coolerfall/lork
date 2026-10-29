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

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("rolling date", func() {
	ginkgo.It("calc period type", func() {
		rd := newRollingDate("2006-01")
		Expect(rd.calcPeriodType()).To(Equal(topOfMonth))
		rd = newRollingDate("2006-01-02")
		Expect(rd.calcPeriodType()).To(Equal(topOfDay))
		rd = newRollingDate("2006-01-02-15")
		Expect(rd.calcPeriodType()).To(Equal(topOfHour))
		rd = newRollingDate("2006-01-02-15-04")
		Expect(rd.calcPeriodType()).To(Equal(topOfMinute))
		rd = newRollingDate("2006-01-02-15-04-05")
		Expect(rd.calcPeriodType()).To(Equal(topOfSecond))
	})
	ginkgo.It("end of next nPeriod", func() {
		rd := newRollingDate("2006-01-02")
		location := time.FixedZone("UTC+8", 8*60*60)
		for _, tc := range []struct {
			now      time.Time
			expected time.Time
		}{
			{
				time.Date(2026, time.October, 29, 23, 59, 59, 0, location),
				time.Date(2026, time.October, 31, 0, 0, 0, 0, location),
			},
			{
				time.Date(2026, time.December, 31, 23, 59, 59, 0, location),
				time.Date(2027, time.January, 2, 0, 0, 0, 0, location),
			},
			{
				time.Date(2024, time.February, 28, 23, 59, 59, 0, location),
				time.Date(2024, time.March, 1, 0, 0, 0, 0, location),
			},
		} {
			next := rd.endOfNextNPeriod(tc.now, 2)
			Expect(next).To(Equal(tc.expected))
		}
	})
})
