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
)

var _ = ginkgo.Describe("threshold filter", func() {
	var event = MakeEvent([]byte(`{"level":"INFO","int":88}`))
	ginkgo.It("not filter", func() {
		filter := NewThresholdFilter(InfoLevel)
		result := filter.Do(event)
		Expect(result).To(Equal(Accept))
	})
	ginkgo.It("filter", func() {
		filter := NewThresholdFilter(ErrorLevel)
		result := filter.Do(event)
		Expect(result).To(Equal(Deny))
	})
})
var _ = ginkgo.Describe("keyword filter", func() {
	var event = MakeEvent([]byte(`{"level":"INFO","int":88,"name":"key"}`))
	ginkgo.It("not filter", func() {
		filter := NewKeywordFilter("logger")
		result := filter.Do(event)
		Expect(result).To(Equal(Deny))
	})
	ginkgo.It("filter", func() {
		filter := NewKeywordFilter("name=key", "int")
		result := filter.Do(event)
		Expect(result).To(Equal(Accept))
	})
})
