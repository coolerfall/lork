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

var _ = ginkgo.Describe("time format", func() {
	ginkgo.It("to unix nano", func() {
		utcTime := "2022-10-09T09:37:43Z"
		nano, err := toUTCUnixNano([]byte(utcTime), time.RFC3339)
		Expect(err).To(BeNil())
		t, _ := time.Parse(time.RFC3339, utcTime)
		Expect(nano).To(Equal(t.UnixNano()))

		zTime := "2022-10-09T09:37:43+08:00"
		nano, err = toUTCUnixNano([]byte(zTime), time.RFC3339)
		Expect(err).To(BeNil())
		t, _ = time.Parse(time.RFC3339, zTime)
		Expect(nano).To(Equal(t.UnixNano()))
	})
	ginkgo.It("convert format", func() {
		var b []byte
		now := time.Now()
		oldTime := now.Format(time.RFC3339Nano)
		newTime, err := convertFormat(b, []byte(oldTime), time.RFC3339Nano, TimeFormatRFC3339)
		Expect(err).To(BeNil())
		Expect(newTime).To(Equal([]byte(now.Format(TimeFormatRFC3339))))
	})
})
