package lork

import (
	"testing"
	"time"
)

func TestWeekdayFormattingMatchesGo(t *testing.T) {
	// Use the formatter's current local offset without changing time.Local.
	name, offset := time.Now().Zone()
	location := time.FixedZone(name, offset)
	start := time.Date(2026, time.November, 1, 12, 34, 56, 123000000, location)
	dates := []time.Time{
		time.Date(1969, time.December, 31, 12, 0, 0, 0, location),
		time.Date(2000, time.February, 29, 12, 0, 0, 0, location),
		time.Date(2026, time.December, 31, 23, 59, 59, 0, location),
		time.Date(2027, time.January, 1, 0, 0, 0, 0, location),
	}
	for day := 0; day < 7; day++ {
		dates = append(dates, start.AddDate(0, 0, day))
	}
	for _, layout := range []string{
		"Mon",
		"Monday",
		"Monday Mon Monday",
		"Mon Jan 2 15:04:05 MST 2006",
		"Monday 2006-01-02 15:04:05.000",
		"Month Monsoon Mon Monday",
	} {
		t.Run(layout, func(t *testing.T) {
			encoder := NewPatternEncoder(func(o *PatternEncoderOption) {
				o.Pattern = "#date{" + layout + "}"
			})
			event := NewLogEvent()
			defer event.Recycle()
			for _, date := range dates {
				want := date.Format(layout)
				got, err := appendFormat(nil, date, layout)
				if err != nil || string(got) != want {
					t.Fatalf("format %s = %q, %v; want %q", date, got, err, want)
				}
				got, err = convertFormat(nil, []byte(date.Format(time.RFC3339Nano)), time.RFC3339Nano, layout)
				if err != nil || string(got) != want {
					t.Fatalf("convert %s = %q, %v; want %q", date, got, err, want)
				}
				event.unixNano = date.UnixNano()
				got, err = encoder.Encode(event)
				if err != nil || string(got) != want+"\n" {
					t.Fatalf("encode %s = %q, %v; want %q", date, got, err, want+"\n")
				}
			}
		})
	}
}

func TestWeekdayFormattingZeroAllocations(t *testing.T) {
	const layout = "Mon Monday 2006-01-02 15:04:05.000"
	event := NewLogEvent()
	defer event.Recycle()
	event.unixNano = time.Now().UnixNano()
	buf := make([]byte, 0, 128)
	if allocs := testing.AllocsPerRun(1000, func() {
		var err error
		buf, err = appendFormatUnix(buf[:0], event.unixNano, layout)
		if err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("weekday formatter allocations = %v; want 0", allocs)
	}
	encoder := NewPatternEncoder(func(o *PatternEncoderOption) {
		o.Pattern = "#date{" + layout + "}"
	})
	if allocs := testing.AllocsPerRun(1000, func() {
		if _, err := encoder.Encode(event); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatalf("weekday pattern encoder allocations = %v; want 0", allocs)
	}
}
