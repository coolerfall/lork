// Copyright (c) 2019-2021 Vincent Cheung (coolingfall@gmail.com).
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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "filter test")
}

var _ = Describe("level filter", func() {
	var event = makeEvent([]byte(`{"level":"INFO","int":88}`))
	It("not filter", func() {
		filter := NewLevelFilter(InfoLevel)
		result := filter.Do(event)
		Expect(result).To(Equal(false))
	})
	It("filter", func() {
		filter := NewLevelFilter(ErrorLevel)
		result := filter.Do(event)
		Expect(result).To(Equal(true))
	})
})
var _ = Describe("keyword filter", func() {
	var event = makeEvent([]byte(`{"level":"INFO","int":88,"name":"key"}`))
	It("not filter", func() {
		filter := NewKeywordFilter("logger")
		result := filter.Do(event)
		Expect(result).To(Equal(false))
	})
	It("filter", func() {
		filter := NewKeywordFilter("name=key", "int")
		result := filter.Do(event)
		Expect(result).To(Equal(true))
	})
})
