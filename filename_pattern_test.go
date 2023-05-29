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
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("filename pattern", func() {
	ginkgo.It("new", func() {
		_, err := newFilenamePattern("archive.#date{2006-01-02}.#index.log")
		Expect(err).To(BeNil())
	})
	ginkgo.It("to filename regex", func() {
		fp, _ := newFilenamePattern("archive.#date{2006-01-02}.#index.log")
		now := time.Now()
		Expect(fp.toFilenameRegexForFixed(now)).
			To(Equal(fmt.Sprintf("archive.%v.(\\d{1,3}).log", now.Format("2006-01-02"))))
	})
	ginkgo.It("has index", func() {
		fp, _ := newFilenamePattern("archive.#date{2006-01-02}.log")
		Expect(fp.hasIndexConverter()).To(BeFalse())
	})
	ginkgo.It("date pattern", func() {
		fp, _ := newFilenamePattern("archive.#date{2006-01-02}.#index.log")
		Expect(fp.datePattern()).To(Equal("2006-01-02"))
	})
})
