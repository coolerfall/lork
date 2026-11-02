package lork

import (
	"regexp"
	"testing"
	"time"
)

func TestArchiveFilenameTimezone(t *testing.T) {
	const layout = "2006-01-02-15-04-05-Z07:00"
	fp, err := newFilenamePattern("app.#date{" + layout + "}.#index.log.gz")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		zone, timestamp string
	}{
		{"UTC", "2026-11-01T23:59:59.123Z"},
		{"America/New_York", "2026-07-01T23:59:59.123-04:00"},
		{"America/New_York", "2026-12-01T23:59:59.123-05:00"},
		{"Australia/Lord_Howe", "2026-10-04T02:30:00.123+11:00"},
	} {
		t.Run(tc.timestamp, func(t *testing.T) {
			location, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			timestamp, err := time.Parse(time.RFC3339Nano, tc.timestamp)
			if err != nil {
				t.Fatal(err)
			}
			timestamp = timestamp.In(location)
			want := "app." + timestamp.Format(layout) + ".7.log.gz"
			if got := fp.convert(timestamp, 7); got != want {
				t.Fatalf("archive filename = %s, want %s", got, want)
			}
			wantRegex := "app." + timestamp.Format(layout) + ".(\\d{1,3}).log.gz"
			if got := fp.toFilenameRegexForFixed(timestamp); got != wantRegex {
				t.Fatalf("filename regex = %s, want %s", got, wantRegex)
			}
			pattern := fp.toArchiveFilenameRegexForFixed(timestamp)
			matched, err := regexp.MatchString("^"+pattern+"$", want)
			if err != nil || !matched {
				t.Fatalf("archive regex %s does not match %s: %v", pattern, want, err)
			}
		})
	}
}
