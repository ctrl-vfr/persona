package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic_WritesAndOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := WriteFileAtomic(path, []byte("first"), 0o600); err != nil {
		t.Fatalf("write 1: %v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != "first" {
		t.Errorf("after write 1 got %q", got)
	}

	if err := WriteFileAtomic(path, []byte("second"), 0o600); err != nil {
		t.Fatalf("write 2: %v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != "second" {
		t.Errorf("after write 2 got %q", got)
	}
}

func TestWriteFileAtomic_AppliesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "private.txt")

	if err := WriteFileAtomic(path, []byte("secret"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("got mode %o, want 0600", info.Mode().Perm())
	}
}

func TestWriteFileAtomic_NoLeftoverTmpFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := WriteFileAtomic(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "data.txt" {
			t.Errorf("unexpected leftover file: %s", e.Name())
		}
	}
}

func TestWriteFileAtomic_FailsOnNonexistentDir(t *testing.T) {
	err := WriteFileAtomic("/nonexistent-dir/data.txt", []byte("x"), 0o600)
	if err == nil {
		t.Fatal("expected error writing to nonexistent dir")
	}
}
