// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package slago

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFindValue(t *testing.T) {
	jsonBytes := []byte(
		`{"level":"INFO","int":88,"time":"2019-08-26T11:06:09.855+08:00","message":"lazy dog"}`)
	buf := &bytes.Buffer{}
	findValue(jsonBytes, "time", buf)
	fmt.Printf("values: %v\n", buf.String())
}
