package lork

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSizeAndTimeRollingAcrossDSTGap(t *testing.T) {
	// Put a spring-forward transition at the next real hour so the complete
	// writer path can run before the gap without changing the system clock.
	transition := time.Now().UTC().Truncate(time.Hour).Add(time.Hour)
	var tz bytes.Buffer
	tz.WriteString("TZif")
	tz.Write(make([]byte, 16))
	for _, count := range []int32{0, 0, 0, 1, 2, 8} {
		_ = binary.Write(&tz, binary.BigEndian, count)
	}
	_ = binary.Write(&tz, binary.BigEndian, int32(transition.Unix()))
	tz.WriteByte(1)
	_ = binary.Write(&tz, binary.BigEndian, int32(-5*3600))
	tz.Write([]byte{0, 0})
	_ = binary.Write(&tz, binary.BigEndian, int32(-4*3600))
	tz.Write([]byte{1, 4})
	tz.WriteString("STD\x00DST\x00")
	location, err := time.LoadLocationFromTZData("rotation-gap", tz.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	previous := time.Local
	time.Local = location
	t.Cleanup(func() { time.Local = previous })

	dir := t.TempDir()
	policy := NewSizeAndTimeBasedRollingPolicy(func(o *SizeAndTimeBasedRPOption) {
		o.FilenamePattern = filepath.Join(dir, "archive.#date{2006-01-02-15}.#index.log")
	}).(*sizeAndTimeBasedRollingPolicy)
	_ = NewFileWriter(func(o *FileWriterOption) {
		o.Filename = filepath.Join(dir, "active.log")
		o.RollingPolicy = policy
	})
	writer := policy.fileWriter
	writer.Start()
	t.Cleanup(writer.Stop)

	// Force one overdue rotation, then verify it does not keep rotating and
	// overwriting index zero while the spring-forward boundary is still ahead.
	policy.nextCheck = time.Now().Add(-time.Hour)
	for _, message := range []string{"first\n", "second\n", "third\n"} {
		if _, err := writer.Write([]byte(message)); err != nil {
			t.Fatal(err)
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var contents strings.Builder
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
		}
		contents.Write(data)
	}
	if got := contents.String(); got != "first\nsecond\nthird\n" {
		t.Fatalf("logged messages were not preserved: %q", got)
	}
	if !policy.nextCheck.After(time.Now()) {
		t.Fatalf("rotation left an expired deadline: %s", policy.nextCheck)
	}
}
