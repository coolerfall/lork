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

var _ = ginkgo.Describe("time based rolling policy", func() {
	ginkgo.It("should trigger", func() {
		tbrp := NewTimeBasedRollingPolicy(func(o *TimeBasedRPOption) {
			o.FilenamePattern = "lork-archive.#date{2006-01-02}.log"
		})
		_ = NewFileWriter(func(o *FileWriterOption) {
			o.RollingPolicy = tbrp
		})
		_ = tbrp.Prepare()
		result := tbrp.ShouldTrigger(0)
		Expect(result).To(BeFalse())
		tbrp.(*timeBasedRollingPolicy).nextCheck = time.Now().Truncate(time.Hour * 36)
		result = tbrp.ShouldTrigger(0)
		Expect(result).To(BeTrue())
	})
})

var _ = ginkgo.Describe("size and time based rolling policy", func() {
	ginkgo.It("should trigger", func() {
		stbrp := NewSizeAndTimeBasedRollingPolicy(func(o *SizeAndTimeBasedRPOption) {
			o.MaxFileSize = "10Kb"
		})
		_ = NewFileWriter(func(o *FileWriterOption) {
			o.RollingPolicy = stbrp
		})
		_ = stbrp.Prepare()
		result := stbrp.ShouldTrigger(10)
		Expect(result).To(BeFalse())
		result = stbrp.ShouldTrigger(10241)
		Expect(result).To(BeTrue())
	})
})
