package lork

import (
	"testing"
	"time"
)

func TestPatternPreservesCommas(t *testing.T) {
	now := time.Now()
	event := NewLogEvent()
	defer event.Recycle()
	event.unixNano = now.UnixNano()
	event.appendMessage("hello,a")

	for _, tt := range []struct {
		name    string
		pattern string
		want    string
	}{
		{"message", "#message", "hello,a\n"},
		{"literal", ",prefix #message, suffix,", ",prefix hello,a, suffix,\n"},
		{"consecutive", "#message,,suffix", "hello,a,,suffix\n"},
		{"date option", "#date{Mon, 02 Jan 2006}", now.Format("Mon, 02 Jan 2006") + "\n"},
		{"composite", "#color(#message, suffix){cyan}", "\x1b[36mhello,a, suffix\x1b[0m\n"},
		{"composite date option", "#color(#date{Mon, 02 Jan 2006}, #message){cyan}", "\x1b[36m" + now.Format("Mon, 02 Jan 2006") + ", hello,a\x1b[0m\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewPatternEncoder(func(o *PatternEncoderOption) { o.Pattern = tt.pattern })
			got, err := encoder.Encode(event)
			if err != nil || string(got) != tt.want {
				t.Fatalf("Encode() = %q, %v; want %q", got, err, tt.want)
			}
			if allocations := testing.AllocsPerRun(1000, func() {
				if _, err := encoder.Encode(event); err != nil {
					panic(err)
				}
			}); allocations != 0 {
				t.Fatalf("Encode() allocations = %v, want 0", allocations)
			}
		})
	}
}
