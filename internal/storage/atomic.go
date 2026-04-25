package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// File modes used across the persona data directory.
// User-private (0600) for content that may include personal prompts,
// transcripts or PIDs; standard 0755 for directories.
const (
	UserFileMode os.FileMode = 0o600
	UserDirMode  os.FileMode = 0o755
)

// WriteFileAtomic writes data to path through a sibling temporary file
// then renames it into place. Prevents truncated files on crash mid-write.
// fsync is best-effort: errors from Sync are returned because losing them
// silently defeats the point of doing the dance.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename %s to %s: %w", tmpName, path, err)
	}
	return nil
}
