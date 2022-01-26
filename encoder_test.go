// Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
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

func TestEncoder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "encoder test")
}

var logEvent = makeEvent([]byte(
	`{"level":"INFO","time":"2019-12-27T10:40:14.465199844+08:00","key":"value"}`,
))
var _ = Describe("json encoder", func() {
	var data []byte
	rt, _ := convertFormat(data,
		[]byte("2019-12-27T10:40:14.465199844+08:00"), TimestampFormat, jsonTimeFormat)
	It("encode", func() {
		result := []byte(`{"time":"` + string(rt) + `","level":"INFO","logger_name":"","message":"","key":"value"}` + "\n")
		je := NewJsonEncoder()
		out, err := je.Encode(logEvent)
		Expect(err).To(BeNil())
		Expect(out).To(Equal(result))
	})
})

var _ = Describe("pattern encoder", func() {
	var data []byte
	rt, _ := convertFormat(data,
		[]byte("2019-12-27T10:40:14.465199844+08:00"), TimestampFormat, "2006-01-02 15:04:05")
	It("encode", func() {
		result := []byte(string(rt) + ` INFO - key=value` + "\n")
		pe := NewPatternEncoder(func(o *PatternEncoderOption) {
			o.Layout = "#date{2006-01-02 15:04:05} #level #message #fields"
		})
		out, err := pe.Encode(logEvent)
		Expect(err).To(BeNil())
		Expect(out).To(Equal(result))
	})
})
