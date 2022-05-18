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
	"errors"
)

var (
	errLeadingInt = errors.New("bad [0-9]*")
	errAtoi       = errors.New("invalid number")
)

func atoi(s []byte) (x int, err error) {
	neg := false
	if s != nil && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		s = s[1:]
	}
	q, rem, err := leadingInt(s)
	x = int(q)
	if err != nil || rem == nil {
		return 0, errAtoi
	}
	if neg {
		x = -x
	}
	return x, nil
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s []byte) (x int64, rem []byte, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, nil, errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, nil, errLeadingInt
		}
	}
	return x, s[i:], nil
}
