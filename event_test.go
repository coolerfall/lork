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

var jsonData = []byte(`{"time":"2022-10-08T15:13:25.320+08:00","level":"INFO",
"logger_name":"github.com/coolerfall/lork","message":"hello","int":88,"lork":"val"}`)

var _ = ginkgo.Describe("event", func() {
	ginkgo.It("make event from json", func() {
		event := MakeEvent(jsonData)
		Expect(string(event.Message())).To(Equal("hello"))
		Expect(string(event.LoggerName())).To(Equal("github.com/coolerfall/lork"))
		Expect(string(event.Level())).To(Equal("INFO"))
		Expect(event.Timestamp()).To(Equal(int64(1665213205320000000)))
	})
	ginkgo.It("copy event", func() {
		event := MakeEvent(jsonData)
		eventCopy := event.Copy()
		Expect(event.Message()).To(Equal(eventCopy.Message()))
		Expect(event.LoggerName()).To(Equal(eventCopy.LoggerName()))
		Expect(event.Level()).To(Equal(eventCopy.Level()))
		Expect(event.Timestamp()).To(Equal(eventCopy.Timestamp()))
	})
})
