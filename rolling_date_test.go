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
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
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
	})
	ginkgo.It("end of next nPeriod", func() {
		rd := newRollingDate("2006-01-02")
		now := time.Now()
		nextAdd := now.Add(time.Hour * 48)
		next := rd.endOfNextNPeriod(now, 2)
		Expect(next.Day()).To(Equal(nextAdd.Day()))
	})
})
