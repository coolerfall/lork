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

func TestFileSize(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "file size test")
}

var _ = Describe("file size", func() {
	It("parse file size", func() {
		size, err := parseFileSize("10Kb")
		Expect(err).To(BeNil())
		Expect(size).To(Equal(int64(10240)))
	})
	It("parse file size error", func() {
		_, err := parseFileSize("10K")
		Expect(err).NotTo(BeNil())
	})
})
