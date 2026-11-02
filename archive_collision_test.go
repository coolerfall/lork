package lork

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestUncompressedArchiveRejectsOccupiedNames(t *testing.T) {
	for _, kind := range []string{"file", "empty file", "directory", "symlink", "same path", "hard link"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			source, target := filepath.Join(dir, "active.log"), filepath.Join(dir, "archive.log")
			if err := os.WriteFile(source, []byte("active messages"), 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "file":
				err = os.WriteFile(target, []byte("previous archive"), 0600)
			case "empty file":
				err = os.WriteFile(target, nil, 0600)
			case "directory":
				err = os.Mkdir(target, 0700)
			case "symlink":
				err = os.Symlink("missing.log", target)
			case "same path":
				target = source
			case "hard link":
				err = os.Link(source, target)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := (&noneArchiver{}).Archive(source, target); err == nil {
				t.Fatal("occupied archive name was accepted")
			}
			assertArchiveContents(t, source, "active messages")
			switch kind {
			case "file":
				assertArchiveContents(t, target, "previous archive")
			case "empty file":
				assertArchiveContents(t, target, "")
			case "symlink":
				if got, err := os.Readlink(target); err != nil || got != "missing.log" {
					t.Fatalf("archive symlink changed: %q, %v", got, err)
				}
			case "directory":
				if info, err := os.Stat(target); err != nil || !info.IsDir() {
					t.Fatalf("archive directory changed: %v", err)
				}
			}
		})
	}
}

func TestUncompressedArchiveMovesWithoutOverwriting(t *testing.T) {
	for _, symlink := range []bool{false, true} {
		t.Run(fmt.Sprintf("symlink=%t", symlink), func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "active.log")
			original := filepath.Join(dir, "original.log")
			target := filepath.Join(dir, "nested", "archive.log")
			if err := os.WriteFile(original, []byte("messages"), 0600); err != nil {
				t.Fatal(err)
			}
			if symlink {
				if err := os.Symlink("original.log", source); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Rename(original, source); err != nil {
				t.Fatal(err)
			}
			if err := (&noneArchiver{}).Archive(source, target); err != nil {
				t.Fatal(err)
			}
			assertArchiveContents(t, target, "messages")
			if _, err := os.Lstat(source); !os.IsNotExist(err) {
				t.Fatalf("active name still exists: %v", err)
			}
			info, err := os.Lstat(target)
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
				t.Fatalf("archive type or permissions changed: %v, %v", info, err)
			}
			if symlink {
				assertArchiveContents(t, original, "messages")
			}
		})
	}
}

func TestUncompressedArchiveConcurrentClaims(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "archive.log")
	const count = 8
	var workers sync.WaitGroup
	errors := make([]error, count)
	for i := 0; i < count; i++ {
		source := filepath.Join(dir, fmt.Sprintf("active-%d.log", i))
		if err := os.WriteFile(source, []byte(fmt.Sprintf("messages-%d", i)), 0600); err != nil {
			t.Fatal(err)
		}
		workers.Add(1)
		go func(i int, source string) {
			defer workers.Done()
			errors[i] = (&noneArchiver{}).Archive(source, target)
		}(i, source)
	}
	workers.Wait()
	winners := 0
	for i, err := range errors {
		source := filepath.Join(dir, fmt.Sprintf("active-%d.log", i))
		if err == nil {
			winners++
			assertArchiveContents(t, target, fmt.Sprintf("messages-%d", i))
			if _, err := os.Stat(source); !os.IsNotExist(err) {
				t.Fatalf("successful source remains: %v", err)
			}
		} else {
			assertArchiveContents(t, source, fmt.Sprintf("messages-%d", i))
		}
	}
	if winners != 1 {
		t.Fatalf("successful archives = %d, want 1", winners)
	}
}

func TestUncompressedArchiveAcrossFilesystems(t *testing.T) {
	// The working directory and system temp directory may use different mounts.
	sourceDir, err := os.MkdirTemp(".", ".lork-archive-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(sourceDir) })
	source := filepath.Join(sourceDir, "active.log")
	target := filepath.Join(t.TempDir(), "archive.log")
	if err := os.WriteFile(source, []byte("cross-filesystem messages"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(source, target); err == nil {
		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		t.Skip("working and temp directories support hard links; copy fallback is covered by symlink sources")
	}
	if err := (&noneArchiver{}).Archive(source, target); err != nil {
		t.Fatal(err)
	}
	assertArchiveContents(t, target, "cross-filesystem messages")
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source remains after copying: %v", err)
	}
	if err := os.WriteFile(source, []byte("later messages"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := (&noneArchiver{}).Archive(source, target); err == nil {
		t.Fatal("copy fallback replaced an existing archive")
	}
	assertArchiveContents(t, source, "later messages")
	assertArchiveContents(t, target, "cross-filesystem messages")
}

func TestRollingSamePathPreservesMessages(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "archive.2000-01-01.log")
	if err := os.WriteFile(active, []byte("previous messages\n"), 0600); err != nil {
		t.Fatal(err)
	}
	previous := time.Date(2000, 1, 1, 12, 0, 0, 0, time.Local)
	if err := os.Chtimes(active, previous, previous); err != nil {
		t.Fatal(err)
	}
	policy := NewTimeBasedRollingPolicy(func(o *TimeBasedRPOption) {
		o.FilenamePattern = filepath.Join(dir, "archive.#date{2006-01-02}.log")
	}).(*timeBasedRollingPolicy)
	_ = NewFileWriter(func(o *FileWriterOption) { o.Filename = active; o.RollingPolicy = policy })
	if err := policy.Prepare(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = policy.fileWriter.Close() })
	if _, err := policy.fileWriter.Write([]byte("next message\n")); err != nil {
		t.Fatal(err)
	}
	assertArchiveContents(t, active, "previous messages\nnext message\n")
}

func TestRollingRevisitedPeriodPreservesArchives(t *testing.T) {
	for _, suffix := range []string{".log", ".log.gz", ".log.zip"} {
		t.Run(suffix, func(t *testing.T) {
			dir := t.TempDir()
			now := time.Now()
			active := filepath.Join(dir, "active.log")
			previous := time.Date(2000, 1, 1, 12, 0, 0, 0, time.Local)
			if err := os.WriteFile(active, []byte("previous period"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(active, previous, previous); err != nil {
				t.Fatal(err)
			}
			currentPrefix := filepath.Join(dir, "archive."+now.Format("2006-01-02"))
			for _, index := range []int{0, 3} {
				if err := os.WriteFile(fmt.Sprintf("%s.%d%s", currentPrefix, index, suffix), []byte("existing archive"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			policy := NewSizeAndTimeBasedRollingPolicy(func(o *SizeAndTimeBasedRPOption) {
				o.FilenamePattern = filepath.Join(dir, "archive.#date{2006-01-02}.#index"+suffix)
				o.MaxFileSize = "1KB"
			}).(*sizeAndTimeBasedRollingPolicy)
			_ = NewFileWriter(func(o *FileWriterOption) { o.Filename = active; o.RollingPolicy = policy })
			if err := policy.Prepare(); err != nil {
				t.Fatal(err)
			}
			// Keep this regression synchronous; compressed index discovery is the same path.
			policy.archiver = &noneArchiver{}
			t.Cleanup(func() { _ = policy.fileWriter.Close() })
			if _, err := policy.fileWriter.Write([]byte(strings.Repeat("x", 1024))); err != nil {
				t.Fatal(err)
			}
			if policy.index != 4 {
				t.Fatalf("new period index = %d, want 4", policy.index)
			}
			if _, err := policy.fileWriter.Write([]byte("next chunk")); err != nil {
				t.Fatal(err)
			}
			for _, index := range []int{0, 3} {
				assertArchiveContents(t, fmt.Sprintf("%s.%d%s", currentPrefix, index, suffix), "existing archive")
			}
			assertArchiveContents(t, currentPrefix+".4"+suffix, strings.Repeat("x", 1024))
			assertArchiveContents(t, active, "next chunk")
		})
	}
}

func TestRollingIndexScanFailurePreservesState(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "active.log")
	if err := os.WriteFile(active, []byte("previous messages"), 0600); err != nil {
		t.Fatal(err)
	}
	previous := time.Date(2000, 1, 1, 12, 0, 0, 0, time.Local)
	if err := os.Chtimes(active, previous, previous); err != nil {
		t.Fatal(err)
	}
	policy := NewSizeAndTimeBasedRollingPolicy(func(o *SizeAndTimeBasedRPOption) {
		o.FilenamePattern = filepath.Join(dir, "#date{2006-01-02}", "archive.#index.log")
	}).(*sizeAndTimeBasedRollingPolicy)
	_ = NewFileWriter(func(o *FileWriterOption) { o.Filename = active; o.RollingPolicy = policy })
	if err := policy.Prepare(); err != nil {
		t.Fatal(err)
	}
	obstacle := filepath.Join(dir, time.Now().Format("2006-01-02"))
	if err := os.WriteFile(obstacle, []byte("obstacle"), 0600); err != nil {
		t.Fatal(err)
	}
	period, deadline, index := policy.timeInCurrentPeriod, policy.nextCheck, policy.index
	if err := policy.Rotate(); err == nil {
		t.Fatal("expected new-period directory scan to fail")
	}
	if policy.timeInCurrentPeriod != period || policy.nextCheck != deadline || policy.index != index {
		t.Fatal("failed index scan advanced rolling state")
	}
	assertArchiveContents(t, active, "previous messages")
	if err := os.Remove(obstacle); err != nil {
		t.Fatal(err)
	}
	if err := policy.Rotate(); err != nil {
		t.Fatal(err)
	}
	assertArchiveContents(t, filepath.Join(dir, "2000-01-01", "archive.0.log"), "previous messages")
}

func TestFilenamePatternRejectsRepeatedIndex(t *testing.T) {
	for _, pattern := range []string{
		"archive.#date{2006-01-02}.#index.#index.log",
		"#date{2006-01-02}/#index/archive.#index.log",
	} {
		if _, err := newFilenamePattern(pattern); err == nil {
			t.Errorf("accepted repeated index in %q", pattern)
		}
	}
}

func assertArchiveContents(t *testing.T, filename, want string) {
	t.Helper()
	got, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s contains %q, want %q", filename, got, want)
	}
}
