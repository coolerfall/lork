package lork

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCompressedArchivePermissions(t *testing.T) {
	for _, suffix := range []string{".gz", ".zip"} {
		for _, mode := range []os.FileMode{0600, 0640} {
			t.Run(fmt.Sprintf("%s/%04o", suffix, mode), func(t *testing.T) {
				dir := t.TempDir()
				source := filepath.Join(dir, "active.log")
				if err := os.WriteFile(source, []byte("private messages"), mode); err != nil {
					t.Fatal(err)
				}
				checkCompressedArchivePermissions(t, source, filepath.Join(dir, "archive.log"+suffix))
			})
		}
	}
}

func TestCompressedArchivePermissionsAcrossFilesystems(t *testing.T) {
	// The worktree and temporary directory may be on different filesystems.
	file, err := os.CreateTemp(".", ".lork-archive-permissions-*.log")
	if err != nil {
		t.Fatal(err)
	}
	source := file.Name()
	t.Cleanup(func() { _ = os.Remove(source) })
	if _, err := file.WriteString("private messages"); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Link(source, filepath.Join(dir, "probe")); err == nil {
		t.Skip("worktree and temporary directory support hard links")
	}
	checkCompressedArchivePermissions(t, source, filepath.Join(dir, "archive.log.gz"))
}

func checkCompressedArchivePermissions(t *testing.T, source, archive string) {
	t.Helper()
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	want := info.Mode().Perm()
	origin, target, err := renameOpenFile(source, archive)
	if err != nil {
		t.Fatal(err)
	}
	defer origin.Close()
	defer target.Close()
	for _, file := range []*os.File{origin, target} {
		info, err := file.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s permissions = %04o, want %04o", file.Name(), got, want)
		}
	}
	data, err := io.ReadAll(origin)
	if err != nil || string(data) != "private messages" {
		t.Fatalf("retained contents = %q, %v", data, err)
	}
}

func TestCompressedArchiveMissingSourceLeavesNoFiles(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := renameOpenFile(filepath.Join(dir, "missing.log"), filepath.Join(dir, "archive.log.gz")); err == nil {
		t.Fatal("missing source was accepted")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("failed archive left %d files", len(entries))
	}
}
