package lork

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type archiveListingFunc func(time.Time) []string

func (f archiveListingFunc) listFilesInPeriod(t time.Time) []string { return f(t) }

func cleanupPeriod(t *testing.T, periods <-chan time.Time) time.Time {
	t.Helper()
	select {
	case period := <-periods:
		return period
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup did not reach the expected period")
		return time.Time{}
	}
}

func TestArchiveCleanupProgress(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	previous := time.Date(2026, 10, 31, 23, 59, 50, 0, location)
	r := &timeBasedArchiveRemover{rd: newRollingDate("2006-01-02"), lastClean: previous, maxHistory: 7}
	periods := make(chan time.Time, 8)
	lister := archiveListingFunc(func(t time.Time) []string { periods <- t; return nil })
	for i := 0; i < 9; i++ {
		r.cleanAsync(previous.Add(time.Duration(i)*time.Second), lister)
	}
	if !r.lastClean.Equal(previous) {
		t.Fatal("same-period rotations advanced the cleanup cursor")
	}
	now := previous.Add(11 * time.Second)
	r.cleanAsync(now, lister)
	if got := cleanupPeriod(t, periods).Format("2006-01-02"); got != "2026-10-25" {
		t.Fatalf("first expired period = %s, want 2026-10-25", got)
	}
	if !r.lastClean.Equal(now) {
		t.Fatalf("cleanup cursor = %s, want supplied time %s", r.lastClean, now)
	}
	later := now.AddDate(0, 0, 2)
	r.cleanAsync(later, lister)
	want := map[string]bool{"2026-10-26": true, "2026-10-27": true}
	for range 2 {
		date := cleanupPeriod(t, periods).Format("2006-01-02")
		if !want[date] {
			t.Fatalf("unexpected or duplicate expired period %s", date)
		}
		delete(want, date)
	}
	// Older requests and wall-clock rollback must not regress the cursor.
	r.cleanAsync(now, lister)
	r.cleanAsync(later.Add(time.Hour), lister)
	if !r.lastClean.Equal(later) {
		t.Fatal("late or same-period requests changed the cleanup cursor")
	}
	select {
	case extra := <-periods:
		t.Fatalf("scheduled redundant cleanup for %s", extra)
	case <-time.After(10 * time.Millisecond):
	}
}

func TestArchiveCleanupConcurrentRanges(t *testing.T) {
	previous := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	r := &timeBasedArchiveRemover{rd: newRollingDate("2006-01-02"), lastClean: previous, maxHistory: 7}
	periods := make(chan time.Time, 512)
	lister := archiveListingFunc(func(t time.Time) []string { periods <- t; return nil })
	start := make(chan struct{})
	var callers sync.WaitGroup
	for i := 0; i < 128; i++ {
		callers.Add(1)
		go func(day int) {
			defer callers.Done()
			<-start
			r.cleanAsync(previous.AddDate(0, 0, day%8+1), lister)
		}(i)
	}
	close(start)
	callers.Wait()
	seen := make(map[string]bool)
	for range 8 {
		date := cleanupPeriod(t, periods).Format("2006-01-02")
		if seen[date] {
			t.Fatalf("duplicate cleanup for %s", date)
		}
		seen[date] = true
	}
	for i := range 8 {
		date := previous.AddDate(0, 0, i+1-7).Format("2006-01-02")
		if !seen[date] {
			t.Fatalf("cleanup omitted %s", date)
		}
	}
	if !r.lastClean.Equal(previous.AddDate(0, 0, 8)) {
		t.Fatal("out-of-order calls regressed the cleanup cursor")
	}
	select {
	case extra := <-periods:
		t.Fatalf("scheduled redundant cleanup for %s", extra)
	case <-time.After(10 * time.Millisecond):
	}
}

func TestArchiveCleanupInitialScanBound(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		pattern string
		want    int
	}{
		{"2006-01-02-15-04-05", maxInitialCleanupPeriods},
		{"2006-01-02-15-04", maxInitialCleanupPeriods},
		{"2006-01-02-15", maxInitialCleanupPeriods},
		{"2006-01-02", 32},
		{"2006-01", 1},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			r := &timeBasedArchiveRemover{rd: newRollingDate(tc.pattern), maxHistory: 7}
			periods := make(chan time.Time, maxInitialCleanupPeriods+1)
			lister := archiveListingFunc(func(t time.Time) []string { periods <- t; return nil })
			r.cleanAsync(now, lister)
			for range tc.want {
				cleanupPeriod(t, periods)
			}
			if !r.lastClean.Equal(now) {
				t.Fatal("first cleanup did not use the supplied time")
			}
			select {
			case extra := <-periods:
				t.Fatalf("initial scan exceeded its bound at %s", extra)
			case <-time.After(10 * time.Millisecond):
			}
		})
	}
}

func TestArchiveCleanupDisabled(t *testing.T) {
	for _, history := range []int{0, -1} {
		r := &timeBasedArchiveRemover{rd: newRollingDate("2006-01-02"), maxHistory: history}
		r.cleanAsync(time.Now(), archiveListingFunc(func(time.Time) []string { panic("disabled cleanup was scheduled") }))
		if !r.lastClean.IsZero() {
			t.Fatal("disabled cleanup advanced its cursor")
		}
	}
}

func TestArchiveCleanupHistoryConcurrentUpdate(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	r := &timeBasedArchiveRemover{rd: newRollingDate("2006-01-02"), lastClean: now, maxHistory: 7}
	var callers sync.WaitGroup
	callers.Add(2)
	go func() {
		defer callers.Done()
		for i := range 100 {
			r.MaxHistory(i % 8)
		}
	}()
	go func() {
		defer callers.Done()
		for range 100 {
			r.cleanAsync(now, archiveListingFunc(func(time.Time) []string { panic("same-period cleanup was scheduled") }))
		}
	}()
	callers.Wait()
}

func TestArchiveCleanupFiles(t *testing.T) {
	for _, indexed := range []bool{false, true} {
		for _, suffix := range []string{"", ".gz", ".zip"} {
			t.Run(fmt.Sprintf("indexed=%v%s", indexed, suffix), func(t *testing.T) {
				dir := t.TempDir()
				pattern := "archive[app].#date{2006-01-02}"
				if indexed {
					pattern += ".#index"
				}
				fp, err := newFilenamePattern(filepath.Join(dir, pattern+".log"+suffix))
				if err != nil {
					t.Fatal(err)
				}
				rd := newRollingDate("2006-01-02")
				base := newTimeBasedArchiveRemover(fp, rd).(*timeBasedArchiveRemover)
				var lister archiveFileLister = base
				if indexed {
					lister = &sizeAndTimeArchiveRemover{timeBasedArchiveRemover: base}
				}
				now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
				var expired, retained []string
				for _, age := range []int{0, 1, 2, 3, 4} {
					indices := []int{0}
					if indexed {
						indices = append(indices, 1, 1000)
					}
					for _, index := range indices {
						filename := fp.convert(now.AddDate(0, 0, -age), index)
						names := []string{filename}
						if suffix != "" {
							names = append(names, uncompressedArchiveFilename(filename))
						}
						for _, name := range names {
							if err := os.WriteFile(name, []byte("keep original contents"), 0600); err != nil {
								t.Fatal(err)
							}
							if age >= 2 {
								expired = append(expired, name)
							} else {
								retained = append(retained, name)
							}
						}
					}
				}
				unrelated := filepath.Join(dir, "other.2026-10-30.0.log"+suffix)
				if err := os.WriteFile(unrelated, []byte("keep original contents"), 0600); err != nil {
					t.Fatal(err)
				}
				retained = append(retained, unrelated)
				base.cleanPeriods(now, 3, 2, lister)
				for _, name := range expired {
					if _, err := os.Stat(name); !os.IsNotExist(err) {
						t.Fatalf("expired archive %s remains: %v", name, err)
					}
				}
				for _, name := range retained {
					data, err := os.ReadFile(name)
					if err != nil || string(data) != "keep original contents" {
						t.Fatalf("retained file %s changed: %q, %v", name, data, err)
					}
				}
			})
		}
	}
}

func TestArchiveCleanupSamePeriodAllocations(t *testing.T) {
	for _, indexed := range []bool{false, true} {
		t.Run(fmt.Sprint(indexed), func(t *testing.T) {
			remover, now := samePeriodRemover(indexed)
			if got := testing.AllocsPerRun(1000, func() { remover.CleanAsync(now) }); got != 0 {
				t.Fatalf("same-period cleanup allocated %v objects", got)
			}
		})
	}
}

func TestArchiveCleanupAmbiguousDates(t *testing.T) {
	for _, layout := range []string{"2006", "02", "15", "2006-01-02-03", "literal"} {
		for _, indexed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/indexed=%v", layout, indexed), func(t *testing.T) {
				pattern := "app.#date{" + layout + "}"
				if indexed {
					pattern += ".#index"
				}
				fp, err := newFilenamePattern(filepath.Join(t.TempDir(), pattern+".log.gz"))
				if err != nil {
					t.Fatal(err)
				}
				base := newTimeBasedArchiveRemover(fp, newRollingDate(layout)).(*timeBasedArchiveRemover)
				base.MaxHistory(7)
				var lister archiveFileLister = base
				if indexed {
					lister = &sizeAndTimeArchiveRemover{timeBasedArchiveRemover: base}
				}
				now := time.Date(2026, 11, 5, 17, 43, 51, 0, time.UTC)
				base.CleanAsync(now)
				if !base.lastClean.IsZero() {
					t.Fatal("ambiguous date scheduled automatic cleanup")
				}
				retained := fp.convert(now.AddDate(0, 0, -1), 1000)
				if layout != "02" {
					retained = fp.convert(now, 1000)
				}
				original := uncompressedArchiveFilename(retained)
				for _, filename := range []string{retained, original} {
					if err := os.WriteFile(filename, []byte("retained"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				base.cleanPeriods(now, maxInitialCleanupPeriods, 7, lister)
				for _, filename := range []string{retained, original} {
					if data, err := os.ReadFile(filename); err != nil || string(data) != "retained" {
						t.Fatalf("ambiguous date deleted retained file %s: %q, %v", filename, data, err)
					}
				}
			})
		}
	}
}

func TestArchiveCleanupHistoricalTimezone(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, indexed := range []bool{false, true} {
		t.Run(fmt.Sprint(indexed), func(t *testing.T) {
			dir := t.TempDir()
			pattern := "app.#date{2006-01-02-Z07:00}"
			index := ""
			if indexed {
				pattern += ".#index"
				index = ".1000"
			}
			fp, err := newFilenamePattern(filepath.Join(dir, pattern+".log.gz"))
			if err != nil {
				t.Fatal(err)
			}
			base := newTimeBasedArchiveRemover(fp, newRollingDate(fp.datePattern())).(*timeBasedArchiveRemover)
			var lister archiveFileLister = base
			if indexed {
				lister = &sizeAndTimeArchiveRemover{timeBasedArchiveRemover: base}
			}
			now := time.Date(2026, 11, 2, 12, 0, 0, 0, location)
			expired := filepath.Join(dir, "app.2026-10-31--04:00"+index+".log.gz")
			retained := filepath.Join(dir, "app.2026-11-02--05:00"+index+".log.gz")
			for _, filename := range []string{expired, uncompressedArchiveFilename(expired), retained} {
				if err := os.WriteFile(filename, []byte("contents"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			base.cleanPeriods(now, 1, 2, lister)
			for _, filename := range []string{expired, uncompressedArchiveFilename(expired)} {
				if _, err := os.Stat(filename); !os.IsNotExist(err) {
					t.Fatalf("expired archive remains: %s: %v", filename, err)
				}
			}
			if data, err := os.ReadFile(retained); err != nil || string(data) != "contents" {
				t.Fatalf("retained archive changed: %q, %v", data, err)
			}
		})
	}
}

func TestArchiveCleanupAsyncDispatch(t *testing.T) {
	for _, indexed := range []bool{false, true} {
		t.Run(fmt.Sprint(indexed), func(t *testing.T) {
			pattern := "app.#date{2006-01-02}"
			if indexed {
				pattern += ".#index"
			}
			fp, err := newFilenamePattern(filepath.Join(t.TempDir(), pattern+".log"))
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 11, 1, 0, 0, 1, 0, time.UTC)
			var remover ArchiveRemover
			var base *timeBasedArchiveRemover
			if indexed {
				r := newSizeAndTimeArchiveRemover(fp, newRollingDate("2006-01-02")).(*sizeAndTimeArchiveRemover)
				remover, base = r, r.timeBasedArchiveRemover
			} else {
				base = newTimeBasedArchiveRemover(fp, newRollingDate("2006-01-02")).(*timeBasedArchiveRemover)
				remover = base
			}
			remover.MaxHistory(1)
			base.lastClean = now.Add(-2 * time.Second)
			expired := fp.convert(now.AddDate(0, 0, -1), 5)
			retained := fp.convert(now, 5)
			for _, filename := range []string{expired, retained} {
				if err := os.WriteFile(filename, []byte("contents"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			// Repeated size rotations in the previous day must not postpone expiry.
			for range 100 {
				remover.CleanAsync(base.lastClean)
			}
			remover.CleanAsync(now)
			deadline := time.Now().Add(2 * time.Second)
			for {
				if _, err := os.Stat(expired); os.IsNotExist(err) {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("expired archive was not deleted across midnight")
				}
				time.Sleep(time.Millisecond)
			}
			if data, err := os.ReadFile(retained); err != nil || string(data) != "contents" {
				t.Fatalf("retained archive changed: %q, %v", data, err)
			}
		})
	}
}

func TestArchiveCleanupDateDirectories(t *testing.T) {
	for _, otherFile := range []bool{false, true} {
		t.Run(fmt.Sprint(otherFile), func(t *testing.T) {
			fp, err := newFilenamePattern(filepath.Join(t.TempDir(), "#date{2006/01/02}", "app.#index.log.gz"))
			if err != nil {
				t.Fatal(err)
			}
			r := newSizeAndTimeArchiveRemover(fp, newRollingDate("2006/01/02")).(*sizeAndTimeArchiveRemover)
			now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
			expired := fp.convert(now.AddDate(0, 0, -2), 1)
			retained := fp.convert(now.AddDate(0, 0, -1), 1)
			for _, filename := range []string{expired, uncompressedArchiveFilename(expired), retained} {
				if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filename, []byte("contents"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			foreign := filepath.Join(filepath.Dir(expired), "keep.txt")
			if otherFile {
				if err := os.WriteFile(foreign, []byte("contents"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			r.cleanPeriods(now, 1, 2, r)
			if _, err := os.Stat(expired); !os.IsNotExist(err) {
				t.Fatalf("expired archive remains: %v", err)
			}
			if _, err := os.Stat(uncompressedArchiveFilename(expired)); !os.IsNotExist(err) {
				t.Fatalf("expired original remains: %v", err)
			}
			if otherFile {
				if data, err := os.ReadFile(foreign); err != nil || string(data) != "contents" {
					t.Fatalf("unrelated file changed: %q, %v", data, err)
				}
			} else if _, err := os.Stat(filepath.Dir(expired)); !os.IsNotExist(err) {
				t.Fatalf("empty archive directory remains: %v", err)
			}
			if data, err := os.ReadFile(retained); err != nil || string(data) != "contents" {
				t.Fatalf("retained archive changed: %q, %v", data, err)
			}
		})
	}
}

func samePeriodRemover(indexed bool) (ArchiveRemover, time.Time) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	base := &timeBasedArchiveRemover{rd: newRollingDate("2006-01-02"), lastClean: now, maxHistory: 7}
	if indexed {
		return &sizeAndTimeArchiveRemover{timeBasedArchiveRemover: base}, now.Add(time.Second)
	}
	return base, now.Add(time.Second)
}

func BenchmarkArchiveCleanupSamePeriod(b *testing.B) {
	for _, indexed := range []bool{false, true} {
		b.Run(fmt.Sprint(indexed), func(b *testing.B) {
			remover, now := samePeriodRemover(indexed)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				remover.CleanAsync(now)
			}
		})
	}
}
