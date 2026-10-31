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
	"os"
	"path/filepath"
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

	ginkgo.DescribeTable("preserves archives across periods", func(failOnce bool) {
		dir := ginkgo.GinkgoT().TempDir()
		filename := filepath.Join(dir, "app.log")
		firstArchive := filepath.Join(dir, "archive.2000-12-31.0.log")
		secondArchive := filepath.Join(dir, "archive.2000-12-31.1.log")
		lastArchive := filepath.Join(dir, "archive.2000-12-31.2.log")
		Expect(os.WriteFile(firstArchive, []byte("first chunk"), 0600)).To(Succeed())
		Expect(os.WriteFile(secondArchive, []byte("second chunk"), 0600)).To(Succeed())
		Expect(os.WriteFile(filename, []byte("last chunk"), 0600)).To(Succeed())
		previousPeriod := time.Date(2000, time.December, 31, 12, 0, 0, 0, time.Local)
		Expect(os.Chtimes(filename, previousPeriod, previousPeriod)).To(Succeed())

		policy := NewSizeAndTimeBasedRollingPolicy(func(o *SizeAndTimeBasedRPOption) {
			o.FilenamePattern = filepath.Join(dir, "archive.#date{2006-01-02}.#index.log")
		}).(*sizeAndTimeBasedRollingPolicy)
		_ = NewFileWriter(func(o *FileWriterOption) {
			o.Filename = filename
			o.RollingPolicy = policy
		})
		ginkgo.DeferCleanup(policy.fileWriter.Close)
		Expect(policy.Prepare()).To(Succeed())
		policy.index = 2 // Two chunks have already been archived in this period.

		if failOnce {
			period, deadline := policy.timeInCurrentPeriod, policy.nextCheck
			Expect(os.Mkdir(lastArchive, 0755)).To(Succeed())
			_, err := policy.fileWriter.Write([]byte("new period"))
			Expect(err).To(HaveOccurred())
			Expect(policy.index).To(Equal(2))
			Expect(policy.timeInCurrentPeriod).To(Equal(period))
			Expect(policy.nextCheck).To(Equal(deadline))
			Expect(os.Remove(lastArchive)).To(Succeed())
		}

		_, err := policy.fileWriter.Write([]byte("new period"))
		Expect(err).NotTo(HaveOccurred())
		contents, err := os.ReadFile(firstArchive)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("first chunk"))
		contents, err = os.ReadFile(secondArchive)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("second chunk"))
		contents, err = os.ReadFile(lastArchive)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("last chunk"))
		contents, err = os.ReadFile(filename)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("new period"))
		Expect(policy.index).To(BeZero())
		Expect(policy.nextCheck).To(Equal(policy.rollingDate.next(policy.timeInCurrentPeriod)))

		currentArchive := filepath.Join(dir,
			"archive."+policy.timeInCurrentPeriod.Format("2006-01-02")+".0.log")
		policy.nextCheck = time.Now().Add(24 * time.Hour)
		policy.triggerSize = 1
		_, err = policy.fileWriter.Write([]byte("next chunk"))
		Expect(err).NotTo(HaveOccurred())
		contents, err = os.ReadFile(currentArchive)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("new period"))
		Expect(policy.index).To(Equal(1))
	},
		ginkgo.Entry("after successful archival", false),
		ginkgo.Entry("after retrying failed archival", true),
	)
})
