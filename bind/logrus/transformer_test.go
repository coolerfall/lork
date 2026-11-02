package logrus

import (
	"encoding/json"
	"testing"

	"github.com/coolerfall/lork"
	"github.com/sirupsen/logrus"
)

type transformerCapture struct {
	encoder lork.Encoder
	events  []map[string]interface{}
}

func (*transformerCapture) Name() string { return "transformer-capture" }

func (w *transformerCapture) DoWrite(event *lork.LogEvent) error {
	encoded, err := w.encoder.Encode(event)
	if err != nil {
		return err
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return err
	}
	w.events = append(w.events, fields)
	return nil
}

func TestLogrusEscapedFieldNames(t *testing.T) {
	standard := logrus.StandardLogger()
	formatter, output, level := standard.Formatter, standard.Out, standard.GetLevel()
	t.Cleanup(func() {
		logrus.SetFormatter(formatter)
		logrus.SetOutput(output)
		logrus.SetLevel(level)
	})
	multi := lork.NewMultiWriter()
	capture := &transformerCapture{encoder: lork.NewJsonEncoder()}
	multi.AddWriter(capture)
	logger := NewLogrusLogger("escaped", multi)
	message := "say \"hello\"\nnext line"
	value := "C:\\logs\n中文😀"
	keys := []string{"plain", "a\"b", "path\\key", "line\n\t\x00", "中文😀"}
	for _, key := range keys {
		logger.Info().Str(key, value).Int("count", 42).Msg(message)
	}
	if len(capture.events) != len(keys) {
		t.Fatalf("captured %d events, want %d", len(capture.events), len(keys))
	}
	for i, key := range keys {
		got := capture.events[i]
		if got[key] != value || got[lork.MessageFieldKey] != message ||
			got[lork.LevelFieldKey] != "INFO" || got["count"] != float64(42) {
			t.Errorf("field %q: unexpected event %#v", key, got)
		}
	}
}

func TestTransformerRejectsInvalidInputAndRecovers(t *testing.T) {
	for _, input := range []string{
		`{"message":"partial","level":`,
		`{"message":"partial","level":"invalid-level"}`,
	} {
		t.Run(input, func(t *testing.T) {
			multi := lork.NewMultiWriter()
			capture := &transformerCapture{encoder: lork.NewJsonEncoder()}
			multi.AddWriter(capture)
			writer := newTransformer(multi)
			if n, err := writer.Write([]byte(input)); err == nil || n != 0 {
				t.Fatalf("invalid write = (%d, %v), want (0, error)", n, err)
			}
			if len(capture.events) != 0 {
				t.Fatal("invalid input was forwarded")
			}
			if writer.buf.Len() != 0 {
				t.Fatal("failed write left buffered data")
			}
			valid := []byte(`{"level":"info","message":"after failure"}`)
			if n, err := writer.Write(valid); err != nil || n != len(valid) {
				t.Fatalf("valid write = (%d, %v), want (%d, nil)", n, err, len(valid))
			}
			if len(capture.events) != 1 || capture.events[0][lork.MessageFieldKey] != "after failure" {
				t.Fatalf("unexpected events after recovery: %#v", capture.events)
			}
		})
	}
}
