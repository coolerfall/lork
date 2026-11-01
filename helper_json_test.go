package lork

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestReplaceJSONPreservesEscapedFields(t *testing.T) {
	input := []byte(`{"a\"b":"say \"hello\" at C:\\logs\nnext","line\n\t\u0000":"中文😀","back\\slash":{"nested":"a\"b"},"level":"info","number":42,"array":["x\\y",null,true]}`)
	for _, newKey := range []string{"level", "level\"name\\\n"} {
		t.Run(newKey, func(t *testing.T) {
			var want map[string]interface{}
			if err := json.Unmarshal(input, &want); err != nil {
				t.Fatal(err)
			}
			delete(want, "level")
			want[newKey] = "INFO"
			var buf bytes.Buffer
			for i := 0; i < 2; i++ {
				buf.Reset()
				err := ReplaceJson(input, &buf, "level", func(_, _ []byte) ([]byte, []byte, error) {
					return []byte(newKey), []byte("INFO"), nil
				})
				if err != nil {
					t.Fatal(err)
				}
				var got map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
					t.Fatalf("invalid JSON %q: %v", buf.String(), err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			}
		})
	}
}

func TestReplaceJSONReturnsErrors(t *testing.T) {
	transformError := errors.New("transform failed")
	for _, tc := range []struct {
		name  string
		input string
		want  error
	}{
		{"transform", `{"first":1,"level":"info"}`, transformError},
		{"parse", `{"first":1,"level":`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBufferString("prefix")
			err := ReplaceJson([]byte(tc.input), buf, "level", func(k, v []byte) ([]byte, []byte, error) {
				return k, v, transformError
			})
			if err == nil {
				t.Fatal("expected an error")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if buf.String() != "prefix" {
				t.Fatalf("failed transformation changed buffer: %q", buf.String())
			}
		})
	}
}

func BenchmarkReplaceJSON(b *testing.B) {
	upper := []byte("INFO")
	for _, tc := range []struct {
		name  string
		input string
	}{
		{"plain", `{"level":"info","message":"hello","count":42}`},
		{"escaped", `{"a\"b":"say \"hello\"","level":"info","path\\key":"C:\\logs"}`},
	} {
		b.Run(tc.name, func(b *testing.B) {
			input := []byte(tc.input)
			var buf bytes.Buffer
			buf.Grow(1024)
			transform := func(k, _ []byte) ([]byte, []byte, error) { return k, upper, nil }
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := ReplaceJson(input, &buf, "level", transform); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
