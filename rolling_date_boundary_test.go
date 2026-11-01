package lork

import (
	"testing"
	"time"
)

func TestRollingDateNext(t *testing.T) {
	for _, tc := range []struct {
		name, pattern, now, zone, want string
	}{
		{"second", "2006-01-02-15-04-05", "2026-03-08T12:34:56.789Z", "UTC", "2026-03-08T12:34:57Z"},
		{"minute", "2006-01-02-15-04", "2026-03-08T12:34:56.789Z", "UTC", "2026-03-08T12:35:00Z"},
		{"local_hour", "2006-01-02-15", "2026-03-08T12:34:56+05:45", "Asia/Kathmandu", "2026-03-08T13:00:00+05:45"},
		{"spring_hour", "2006-01-02-15", "2026-03-08T01:30:00-05:00", "America/New_York", "2026-03-08T03:00:00-04:00"},
		{"spring_minute", "2006-01-02-15-04", "2026-03-08T01:59:30-05:00", "America/New_York", "2026-03-08T03:00:00-04:00"},
		{"spring_second", "2006-01-02-15-04-05", "2026-03-08T01:59:59.789-05:00", "America/New_York", "2026-03-08T03:00:00-04:00"},
		{"fall_first_hour", "2006-01-02-15", "2026-11-01T01:30:00-04:00", "America/New_York", "2026-11-01T02:00:00-05:00"},
		{"fall_second_hour", "2006-01-02-15", "2026-11-01T01:30:00-05:00", "America/New_York", "2026-11-01T02:00:00-05:00"},
		{"fall_first_minute", "2006-01-02-15-04", "2026-11-01T01:30:45-04:00", "America/New_York", "2026-11-01T01:31:00-04:00"},
		{"fall_second_minute", "2006-01-02-15-04", "2026-11-01T01:30:45-05:00", "America/New_York", "2026-11-01T01:31:00-05:00"},
		{"fall_second_second", "2006-01-02-15-04-05", "2026-11-01T01:30:45.789-05:00", "America/New_York", "2026-11-01T01:30:46-05:00"},
		{"fall_berlin_minute", "2006-01-02-15-04", "2026-10-25T02:30:45+01:00", "Europe/Berlin", "2026-10-25T02:31:00+01:00"},
		{"spring_day", "2006-01-02", "2026-03-08T12:00:00-04:00", "America/New_York", "2026-03-09T00:00:00-04:00"},
		{"fall_day", "2006-01-02", "2026-11-01T12:00:00-05:00", "America/New_York", "2026-11-02T00:00:00-05:00"},
		{"missing_midnight", "2006-01-02", "2018-11-03T12:00:00-03:00", "America/Sao_Paulo", "2018-11-04T01:00:00-02:00"},
		{"before_missing_midnight", "2006-01-02", "2018-11-03T23:30:00-03:00", "America/Sao_Paulo", "2018-11-04T01:00:00-02:00"},
		{"partial_hour", "2006-01-02-15", "2026-10-04T01:45:00+10:30", "Australia/Lord_Howe", "2026-10-04T02:30:00+11:00"},
		{"missing_day", "2006-01-02", "2011-12-29T12:00:00-10:00", "Pacific/Apia", "2011-12-31T00:00:00+14:00"},
		{"leap_month", "2006-01", "2024-02-29T23:59:59+08:00", "Asia/Shanghai", "2024-03-01T00:00:00+08:00"},
		{"year", "2006-01", "2026-12-31T23:59:59Z", "UTC", "2027-01-01T00:00:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			location, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			now, err := time.Parse(time.RFC3339Nano, tc.now)
			if err != nil {
				t.Fatal(err)
			}
			want, err := time.Parse(time.RFC3339, tc.want)
			if err != nil {
				t.Fatal(err)
			}
			next := newRollingDate(tc.pattern).next(now.In(location))
			if !next.Equal(want) || !next.After(now) {
				t.Fatalf("next rotation = %s, want %s after %s", next, want, now)
			}
			if next.Location() != location {
				t.Fatalf("next rotation location = %s, want %s", next.Location(), location)
			}
		})
	}
}

func TestRollingDateCleanupSupported(t *testing.T) {
	for _, tc := range []struct {
		layout string
		want   bool
	}{
		{"2006-01-02-15-04-05", true},
		{"2006-01-02-15-04", true},
		{"2006-01-02-15", true},
		{"2006-01-02", true},
		{"2006-01", true},
		{"20060102", true},
		{"2006/Jan/_2", true},
		{"2006-002", true},
		{"2006-01-02-03-PM", true},
		{"2006-01-02-15-Z07:00", true},
		{"2006", false},
		{"06-01-02", false},
		{"02", false},
		{"15", false},
		{"2006-01-02-03", false},
		{"2006-01-15", false},
		{"literal", false},
	} {
		t.Run(tc.layout, func(t *testing.T) {
			if got := newRollingDate(tc.layout).cleanupSupported; got != tc.want {
				t.Fatalf("cleanup supported = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRollingDatePeriodBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, pattern, start, end, zone string
		want                            int
	}{
		{"second", "2006-01-02-15-04-05", "2026-11-01T12:00:00Z", "2026-11-01T12:00:01Z", "UTC", 1},
		{"minute", "2006-01-02-15-04", "2026-11-01T12:00:59Z", "2026-11-01T12:01:01Z", "UTC", 1},
		{"hour", "2006-01-02-15", "2026-11-01T12:59:59Z", "2026-11-01T13:00:01Z", "UTC", 1},
		{"day", "2006-01-02", "2026-10-31T23:59:59Z", "2026-11-01T00:00:01Z", "UTC", 1},
		{"month", "2006-01", "2026-12-31T23:59:59Z", "2027-01-01T00:00:01Z", "UTC", 1},
		{"leap_day", "2006-01-02", "2024-02-28T23:59:59Z", "2024-03-01T00:00:01Z", "UTC", 2},
		{"same_period", "2006-01-02", "2026-11-01T00:00:01Z", "2026-11-01T23:59:59Z", "UTC", 0},
		{"clock_backward", "2006-01-02", "2026-11-01T00:00:01Z", "2026-10-31T23:59:59Z", "UTC", 0},
		{"unix_epoch", "2006-01-02", "1969-12-31T23:59:59Z", "1970-01-01T00:00:01Z", "UTC", 1},
		{"local_hour", "2006-01-02-15", "2026-11-01T12:59:59+05:45", "2026-11-01T13:00:01+05:45", "Asia/Kathmandu", 1},
		{"local_month", "2006-01", "2026-10-31T23:59:59+08:00", "2026-11-01T00:00:01+08:00", "Asia/Shanghai", 1},
		{"spring_day", "2006-01-02", "2026-03-08T00:00:00-05:00", "2026-03-09T00:00:00-04:00", "America/New_York", 1},
		{"fall_day", "2006-01-02", "2026-11-01T00:00:00-04:00", "2026-11-02T00:00:00-05:00", "America/New_York", 1},
		{"repeated_hour", "2006-01-02-15", "2026-11-01T01:59:59-04:00", "2026-11-01T01:00:01-05:00", "America/New_York", 0},
		{"repeated_minutes", "2006-01-02-15-04", "2026-11-01T01:59:59-04:00", "2026-11-01T01:00:01-05:00", "America/New_York", 0},
		{"after_repeated_hour", "2006-01-02-15", "2026-11-01T01:59:59-04:00", "2026-11-01T02:00:01-05:00", "America/New_York", 1},
		{"spring_hours", "2006-01-02-15", "2026-03-08T01:59:59-05:00", "2026-03-08T03:00:01-04:00", "America/New_York", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			location, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			start, err := time.Parse(time.RFC3339, tc.start)
			if err != nil {
				t.Fatal(err)
			}
			end, err := time.Parse(time.RFC3339, tc.end)
			if err != nil {
				t.Fatal(err)
			}
			rd := newRollingDate(tc.pattern)
			if got := rd.periodCrossed(start, end.In(location)); got != tc.want {
				t.Fatalf("crossed %d periods, want %d", got, tc.want)
			}
		})
	}
}

func TestRollingDateCleanupAcrossDST(t *testing.T) {
	for _, tc := range []struct {
		name, pattern, now, zone, want string
		offset                         int
		exists                         bool
	}{
		{"missing_hour", "2006-01-02-15", "2026-03-08T03:30:00-04:00", "America/New_York", "", -1, false},
		{"partial_hour", "2006-01-02-15", "2026-10-04T04:30:00+11:00", "Australia/Lord_Howe", "2026-10-04-02", -2, true},
		{"missing_midnight", "2006-01-02", "2018-11-06T12:00:00-02:00", "America/Sao_Paulo", "2018-11-04", -2, true},
		{"missing_day", "2006-01-02", "2011-12-31T12:00:00+14:00", "Pacific/Apia", "", -1, false},
		{"month", "2006-01", "2027-01-01T00:00:01+08:00", "Asia/Shanghai", "2026-12", -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			location, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			now, err := time.Parse(time.RFC3339, tc.now)
			if err != nil {
				t.Fatal(err)
			}
			period, exists := newRollingDate(tc.pattern).periodForCleanup(now.In(location), tc.offset)
			if exists != tc.exists {
				t.Fatalf("period exists = %v, want %v; got %s", exists, tc.exists, period)
			}
			if exists && period.Format(tc.pattern) != tc.want {
				t.Fatalf("cleanup period = %s, want %s", period.Format(tc.pattern), tc.want)
			}
		})
	}
}
