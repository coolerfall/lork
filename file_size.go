// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
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
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	_KB = 1024
	_MB = 1024 * _KB
	_GB = 1024 * _MB
)

// parseFileSize parse file size string to byte length.
func parseFileSize(fileSizeStr string) (int64, error) {
	sizeRegex, err := regexp.Compile(`([0-9]+)\s*(?i)(kb|mb|gb)s?`)
	if err != nil {
		return 0, err
	}
	result := sizeRegex.FindStringSubmatch(fileSizeStr)
	if len(result) != 3 {
		return 0, errors.New("not a valid file size string")
	}

	var coefficient int64
	lenVal, err := strconv.ParseInt(strings.TrimSpace(result[1]), 10, 64)
	if err != nil {
		return 0, err
	}
	unit := strings.ToUpper(result[2])
	switch unit {
	case "":
		coefficient = 1
	case "KB":
		coefficient = _KB
	case "MB":
		coefficient = _MB
	case "GB":
		coefficient = _GB
	default:
		return 0, fmt.Errorf("unkown unit: %s", unit)
	}

	return lenVal * coefficient, nil
}
