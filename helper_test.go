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
	"bytes"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("helper", func() {
	ginkgo.It("replace json", func() {
		var buf = new(bytes.Buffer)
		var json = []byte(`{"level":"INFO","int":88}`)
		var result = `{"level":"INFO","ints":99}` + "\n"
		err := ReplaceJson(json, buf, "int", func(k, v []byte) (nk, kv []byte, e error) {
			return []byte("ints"), []byte("99"), nil
		})
		Expect(err).To(BeNil())
		Expect(buf.String()).To(Equal(result))
	})
})
