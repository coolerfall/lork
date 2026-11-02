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
	"encoding/json"
	"errors"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var logEvent = MakeEvent([]byte(
	`{"level":"INFO","time":"2019-12-27T10:40:14.465199844+08:00","key":"value"}`,
))
var _ = ginkgo.Describe("json encoder", func() {
	var data []byte
	rt, _ := appendFormatUnix(data, logEvent.Timestamp(), TimeFormatRFC3339)
	ginkgo.It("encode", func() {
		result := []byte(`{"time":"` + string(rt) + `","level":"INFO","logger_name":"","message":"","key":"value"}` + "\n")
		je := NewJsonEncoder()
		out, err := je.Encode(logEvent)
		Expect(err).To(BeNil())
		Expect(out).To(Equal(result))
	})

	ginkgo.It("escapes strings and preserves their values", func() {
		controls := make([]byte, 32)
		for i := range controls {
			controls[i] = byte(i)
		}
		values := []string{
			"plain", `say "hello"`, `C:\new\logs`, "hello\nworld",
			string(controls), "中文😀\u2028\u2029", "bad\xff\xfeutf8",
		}
		encoder := NewJsonEncoder()
		for _, value := range values {
			event := NewLogEvent()
			event.appendLogger([]byte(value))
			event.appendMessage(value)
			event.appendString(value, value)
			event.appendStrings("strings", []string{value, ""})
			event.appendErrors("errors", []error{errors.New(value)})
			event.appendInt("number", 42)
			out, err := encoder.Encode(event)
			event.Recycle()
			Expect(err).To(Succeed())
			var decoded map[string]interface{}
			Expect(json.Unmarshal(out, &decoded)).To(Succeed())
			// Match encoding/json's replacement of malformed UTF-8 bytes.
			canonical, err := json.Marshal(value)
			Expect(err).To(Succeed())
			var want string
			Expect(json.Unmarshal(canonical, &want)).To(Succeed())
			Expect(decoded[LoggerNameFieldKey]).To(Equal(want))
			Expect(decoded[MessageFieldKey]).To(Equal(strings.TrimRight(want, "\n")))
			Expect(decoded[want]).To(Equal(want))
			Expect(decoded["strings"]).To(Equal([]interface{}{want, ""}))
			Expect(decoded["errors"]).To(Equal([]interface{}{want}))
			Expect(decoded["number"]).To(Equal(float64(42)))
		}
	})

	ginkgo.It("decodes JSON strings once before encoding", func() {
		input := []byte(`{"time":"2026-11-01T10:00:00Z","level":"\u0049NFO","logger_name":"a\\b\"c","message":"say \"hello\" at C:\\new\nline","a\"b":"c\\d\n","nested":{"text":"a\"b"},"array":["c\\d",1,true]}`)
		var want map[string]interface{}
		Expect(json.Unmarshal(input, &want)).To(Succeed())
		encoder := NewJsonEncoder()
		for i := 0; i < 3; i++ {
			event := MakeEvent(input)
			Expect(string(event.LoggerName())).To(Equal(want[LoggerNameFieldKey]))
			Expect(string(event.Message())).To(Equal(want[MessageFieldKey]))
			Expect(event.Fields(func(k, v []byte, isString bool) error {
				if string(k) == `a"b` {
					Expect(isString).To(BeTrue())
					Expect(string(v)).To(Equal(want[`a"b`]))
				}
				return nil
			})).To(Succeed())
			out, err := encoder.Encode(event)
			event.Recycle()
			Expect(err).To(Succeed())
			var decoded map[string]interface{}
			Expect(json.Unmarshal(out, &decoded)).To(Succeed())
			// Timestamp formatting is independent of string escaping.
			delete(decoded, TimestampFieldKey)
			delete(want, TimestampFieldKey)
			Expect(decoded).To(Equal(want))
			input = append(input[:0], out...)
		}
	})
})

var _ = ginkgo.Describe("pattern encoder", func() {
	var data []byte
	rt, _ := appendFormatUnix(data, logEvent.Timestamp(), "2006-01-02 15:04:05")
	ginkgo.It("encode", func() {
		result := []byte(string(rt) + ` INFO - key=value` + "\n")
		pe := NewPatternEncoder(func(o *PatternEncoderOption) {
			o.Pattern = "#date{2006-01-02 15:04:05} #level #message #fields"
		})
		out, err := pe.Encode(logEvent)
		Expect(err).To(BeNil())
		Expect(out).To(Equal(result))
	})

	ginkgo.It("preserves preceding content when fields are empty", func() {
		for _, tc := range []struct {
			pattern string
			fields  bool
			want    string
		}{
			{"#fields", false, "\n"},
			{"#fields #message", false, " hello\n"},
			{"prefix#fields", false, "prefix\n"},
			{"#message #fields", false, "hello \n"},
			{"#fields #message", true, "key=value count=2 hello\n"},
			{"#fields #fields", true, "key=value count=2 key=value count=2\n"},
		} {
			event := NewLogEvent()
			event.appendMessage("hello")
			if tc.fields {
				event.appendString("key", "value")
				event.appendInt("count", 2)
			}
			encoder := NewPatternEncoder(func(o *PatternEncoderOption) {
				o.Pattern = tc.pattern
			})
			for i := 0; i < 2; i++ {
				out, err := encoder.Encode(event)
				Expect(err).To(Succeed())
				Expect(string(out)).To(Equal(tc.want), tc.pattern)
			}
			event.Recycle()
		}
	})
})
