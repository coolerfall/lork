// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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

func TestPatternParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pattern parser test")
}

var _ = Describe("pattern", func() {
	It("parse pattern", func() {
		parser := NewPatternParser(
			`archive-#color(#date{2016-01-02 15:04:05.000}){cyan}.#index.log`)
		node, err := parser.Parse()
		Expect(err).To(BeNil())
		Expect("archive").To(Equal(node.value))
		node = node.next
		Expect(typeLiteral).To(Equal(node._type))
		Expect("-").To(Equal(node.value))
		node = node.next
		Expect(typeComposite).To(Equal(node._type))
		Expect("color").To(Equal(node.value))
		Expect("cyan").To(Equal(node.options[0]))
		child := node.child
		Expect(typeSingle).To(Equal(child._type))
		Expect("date").To(Equal(child.value))
		Expect("2016-01-02 15:04:05.000").To(Equal(child.options[0]))
		node = node.next
		Expect(".").To(Equal(node.value))
		node = node.next
		Expect("index").To(Equal(node.value))
		node = node.next
		Expect(".").To(Equal(node.value))
		node = node.next
		Expect("log").To(Equal(node.value))
	})
})
