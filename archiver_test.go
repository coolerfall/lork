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
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("archiver", func() {
	ginkgo.It("rename and open file", func() {
		dir := ginkgo.GinkgoT().TempDir()
		fn := filepath.Join(dir, "active.log")
		archiveFilename := filepath.Join(dir, "archive.log.gz")
		of, _ := os.Create(fn)
		_, _ = of.WriteString("ABC")
		Expect(of.Close()).To(Succeed())
		origin, target, err := renameOpenFile(fn, archiveFilename)
		Expect(err).To(BeNil())
		defer origin.Close()
		defer target.Close()
		c, err := io.ReadAll(origin)
		Expect(err).To(BeNil())
		Expect(string(c)).To(Equal("ABC"))
		Expect(origin.Name()).To(Equal(filepath.Join(dir, "archive.log")))
		Expect(target.Name()).To(Equal(archiveFilename))
		err = os.Remove(origin.Name())
		Expect(err).To(BeNil())
	})
})
