package zap

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/coolerfall/lork"
	uberzap "go.uber.org/zap"
)

type zapRecordCapture struct {
	encoder lork.Encoder
	events  []map[string]json.RawMessage
}

func (*zapRecordCapture) Name() string { return "zap-record-capture" }

func (w *zapRecordCapture) DoWrite(event *lork.LogEvent) error {
	encoded, err := w.encoder.Encode(event)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return err
	}
	w.events = append(w.events, fields)
	return nil
}

func TestUintPreservesFields(t *testing.T) {
	previous := uberzap.L()
	t.Cleanup(func() { uberzap.ReplaceGlobals(previous) })
	multi := lork.NewMultiWriter()
	t.Cleanup(multi.ResetWriter)
	capture := &zapRecordCapture{encoder: lork.NewJsonEncoder()}
	multi.AddWriter(capture)
	logger := NewZapLogger("uint", multi)
	for _, value := range []uint{0, 7, ^uint(0)} {
		want := strconv.FormatUint(uint64(value), 10)
		t.Run(want, func(t *testing.T) {
			capture.events = nil
			logger.Info().Str("before", "retained").Uint("count", value).Int("after", 42).Msg("operation")
			logger.Info().Msg("next record")
			if len(capture.events) != 2 {
				t.Fatalf("captured %d events, want 2", len(capture.events))
			}
			got := capture.events[0]
			if string(got["count"]) != want {
				t.Errorf("count = %s, want %s", got["count"], want)
			}
			if string(got["before"]) != `"retained"` || string(got["after"]) != "42" ||
				string(got[lork.MessageFieldKey]) != `"operation"` {
				t.Errorf("message or adjacent fields changed: %s", got)
			}
			next := capture.events[1]
			if string(next[lork.MessageFieldKey]) != `"next record"` ||
				next["count"] != nil || next["before"] != nil || next["after"] != nil {
				t.Errorf("fields leaked into next record: %s", next)
			}
		})
	}
}
