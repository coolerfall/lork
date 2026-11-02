package bridge

import (
	"testing"

	"github.com/coolerfall/lork"
)

func TestBridgeParseLevel(t *testing.T) {
	bridges := []struct {
		name   string
		bridge lork.Bridge
	}{
		{"logrus", &logrusBridge{}},
		{"zap", &zapBridge{}},
		{"zerolog", &zerologBridge{}},
	}
	levels := []struct {
		text string
		want lork.Level
	}{
		{"debug", lork.DebugLevel},
		{"info", lork.InfoLevel},
		{"warn", lork.WarnLevel},
		{"error", lork.ErrorLevel},
		{"fatal", lork.FatalLevel},
		{"panic", lork.PanicLevel},
		{"PANIC", lork.PanicLevel},
	}
	for _, bridge := range bridges {
		t.Run(bridge.name, func(t *testing.T) {
			for _, level := range levels {
				t.Run(level.text, func(t *testing.T) {
					if got := bridge.bridge.ParseLevel(level.text); got != level.want {
						t.Errorf("ParseLevel(%q) = %v, want %v", level.text, got, level.want)
					}
				})
			}
		})
	}
	for _, level := range []string{"dpanic", "DPANIC"} {
		t.Run("zap/"+level, func(t *testing.T) {
			if got := (&zapBridge{}).ParseLevel(level); got != lork.PanicLevel {
				t.Errorf("ParseLevel(%q) = %v, want %v", level, got, lork.PanicLevel)
			}
		})
	}
}
