package lork

import (
	"encoding/json"
	"errors"
	"testing"
)

type errorRecordWriter func(*LogEvent) error

func (errorRecordWriter) Name() string { return "error-record" }

func (w errorRecordWriter) DoWrite(event *LogEvent) error { return w(event) }

func TestClassicRecordErrors(t *testing.T) {
	first := errors.New("first")
	last := errors.New("last\n\"quoted\"")
	for _, tc := range []struct {
		name       string
		fields     func(Record) Record
		wantError  string
		wantErrors string
	}{
		{
			name:   "nil error is omitted",
			fields: func(r Record) Record { return r.Err(nil) },
		},
		{
			name:      "non-nil error is preserved",
			fields:    func(r Record) Record { return r.Err(first) },
			wantError: `"first"`,
		},
		{
			name:      "nil error preserves an earlier error",
			fields:    func(r Record) Record { return r.Err(first).Err(nil) },
			wantError: `"first"`,
		},
		{
			name:       "nil slice is empty array",
			fields:     func(r Record) Record { return r.Errs("errors", nil) },
			wantErrors: `[]`,
		},
		{
			name:       "empty slice is empty array",
			fields:     func(r Record) Record { return r.Errs("errors", []error{}) },
			wantErrors: `[]`,
		},
		{
			name:       "nil element is null",
			fields:     func(r Record) Record { return r.Errs("errors", []error{nil}) },
			wantErrors: `[null]`,
		},
		{
			name:       "nil elements preserve array positions",
			fields:     func(r Record) Record { return r.Errs("errors", []error{nil, first, nil, last, nil}) },
			wantErrors: `[null,"first",null,"last\n\"quoted\"",null]`,
		},
		{
			name:       "non-nil errors are preserved",
			fields:     func(r Record) Record { return r.Errs("errors", []error{first, last}) },
			wantErrors: `["first","last\n\"quoted\""]`,
		},
		{
			name:       "any error slice handles nil elements",
			fields:     func(r Record) Record { return r.Any("errors", []error{nil, first, nil}) },
			wantErrors: `[null,"first",null]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := NewLoggerContext(NewClassicLogger).RealLogger("errors")
			encoder := NewJsonEncoder()
			writes := 0
			logger.AddWriter(errorRecordWriter(func(event *LogEvent) error {
				writes++
				encoded, err := encoder.Encode(event)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(encoded, &fields); err != nil {
					t.Fatalf("invalid JSON %q: %v", encoded, err)
				}
				for key, want := range map[string]string{
					ErrorFieldKey:   tc.wantError,
					"errors":        tc.wantErrors,
					MessageFieldKey: `"operation completed"`,
					"after":         `"preserved"`,
				} {
					if got := string(fields[key]); got != want {
						t.Errorf("field %q = %q, want %q", key, got, want)
					}
				}
				return nil
			}))
			for i := 0; i < 2; i++ {
				tc.fields(logger.Info()).Str("after", "preserved").Msg("operation completed")
			}
			if writes != 2 {
				t.Fatalf("writes = %d, want 2", writes)
			}
		})
	}
}
