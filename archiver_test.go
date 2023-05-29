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
	"io"
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("archiver", func() {
	ginkgo.It("rename and open file", func() {
		fn := "/tmp/test.log"
		of, _ := os.Create(fn)
		_, _ = of.WriteString("ABC")
		origin, target, err := renameOpenFile(fn, "/tmp/test.log.gz")
		Expect(err).To(BeNil())
		c, err := io.ReadAll(origin)
		Expect(err).To(BeNil())
		Expect(string(c)).To(Equal("ABC"))
		Expect(strings.HasPrefix(origin.Name(), "/tmp/test.log.gz")).To(BeTrue())
		Expect(target.Name()).To(Equal("/tmp/test.log.gz"))
		err = os.Remove(origin.Name())
		Expect(err).To(BeNil())
	})
})
