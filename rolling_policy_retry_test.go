package lork

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type timeRollingCleanupRecorder struct {
	calls []time.Time
}

func (*timeRollingCleanupRecorder) MaxHistory(int) {}

func (r *timeRollingCleanupRecorder) CleanAsync(now time.Time) {
	r.calls = append(r.calls, now)
}

func TestTimeBasedRollingRetriesFailedArchive(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "active.log")
	archive := filepath.Join(dir, "archive.2000-01-01.log")
	previous := time.Date(2000, time.January, 1, 12, 0, 0, 0, time.Local)
	if err := os.WriteFile(active, []byte("previous\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(active, previous, previous); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(archive, 0755); err != nil {
		t.Fatal(err)
	}
	policy := NewTimeBasedRollingPolicy(func(o *TimeBasedRPOption) {
		o.FilenamePattern = filepath.Join(dir, "archive.#date{2006-01-02}.log")
	}).(*timeBasedRollingPolicy)
	_ = NewFileWriter(func(o *FileWriterOption) {
		o.Filename = active
		o.RollingPolicy = policy
	})
	writer := policy.fileWriter
	t.Cleanup(writer.Stop)
	if err := policy.Prepare(); err != nil {
		t.Fatal(err)
	}
	cleanup := &timeRollingCleanupRecorder{}
	policy.archiveRemover = cleanup
	period, deadline := policy.timeInCurrentPeriod, policy.nextCheck
	contents := "previous\n"
	for _, message := range []string{"first failure\n", "second failure\n"} {
		if n, err := writer.Write([]byte(message)); err != nil || n != len(message) {
			t.Fatalf("write during archive failure = %d, %v", n, err)
		}
		contents += message
		assertArchiveContents(t, active, contents)
		if policy.timeInCurrentPeriod != period || policy.nextCheck != deadline {
			t.Fatal("failed archive advanced rolling state")
		}
		if len(cleanup.calls) != 0 {
			t.Fatal("failed archive triggered retention cleanup")
		}
		if !policy.ShouldTrigger(0) || !policy.ShouldTrigger(0) {
			t.Fatal("failed archive no longer triggers a retry")
		}
		if policy.timeInCurrentPeriod != period || policy.nextCheck != deadline {
			t.Fatal("checking the trigger changed rolling state")
		}
	}
	if err := os.Remove(archive); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("recovered\n")); err != nil {
		t.Fatal(err)
	}
	assertArchiveContents(t, archive, contents)
	assertArchiveContents(t, active, "recovered\n")
	if !policy.timeInCurrentPeriod.After(period) || policy.nextCheck != policy.rollingDate.next(policy.timeInCurrentPeriod) {
		t.Fatal("successful archive did not advance rolling state")
	}
	if len(cleanup.calls) != 1 || cleanup.calls[0] != policy.timeInCurrentPeriod {
		t.Fatal("successful archive did not trigger cleanup with the committed period")
	}
	if _, err := writer.Write([]byte("same period\n")); err != nil {
		t.Fatal(err)
	}
	assertArchiveContents(t, archive, contents)
	assertArchiveContents(t, active, "recovered\nsame period\n")
	if len(cleanup.calls) != 1 {
		t.Fatal("writer rotated again within the same period")
	}
}
