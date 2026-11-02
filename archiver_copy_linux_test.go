package lork

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestUncompressedArchiveCopyFailure(t *testing.T) {
	if os.Getenv("LORK_ARCHIVE_COPY_FAILURE") == "1" {
		// Limit only this subprocess to force a partial write without filling a disk.
		signal.Ignore(syscall.SIGXFSZ)
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 128, Max: 128}); err != nil {
			t.Fatal(err)
		}
		source, target := os.Getenv("LORK_ARCHIVE_SOURCE"), os.Getenv("LORK_ARCHIVE_TARGET")
		if err := (&noneArchiver{}).Archive(source, target); err == nil {
			t.Fatal("expected copy to fail after a partial write")
		}
		assertArchiveContents(t, source, strings.Repeat("x", 1024))
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			t.Fatalf("partial archive remains: %v", err)
		}
		return
	}

	dir := t.TempDir()
	original, source, target := filepath.Join(dir, "original.log"), filepath.Join(dir, "active.log"), filepath.Join(dir, "archive.log")
	if err := os.WriteFile(original, []byte(strings.Repeat("x", 1024)), 0600); err != nil {
		t.Fatal(err)
	}
	// A symlink source selects the same exclusive-copy fallback as a cross-device move.
	if err := os.Symlink("original.log", source); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestUncompressedArchiveCopyFailure$")
	cmd.Env = append(os.Environ(), "LORK_ARCHIVE_COPY_FAILURE=1", "LORK_ARCHIVE_SOURCE="+source, "LORK_ARCHIVE_TARGET="+target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("copy failure subprocess: %v\n%s", err, output)
	}
	assertArchiveContents(t, original, strings.Repeat("x", 1024))
}
