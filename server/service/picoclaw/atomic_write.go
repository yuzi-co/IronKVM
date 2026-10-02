package picoclaw

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFileAtomic replaces path with data so that a reader, or a power cut,
// sees either the old file or the new one and never a truncated mix. The data
// goes to a temporary file in the same directory, is synced, and is renamed
// over path. Renaming over an existing file works on the exFAT /data partition
// under Linux as well as on ext4.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// exFAT has no permission bits, so a failure to set them there is not an
	// error. CreateTemp already made the file 0600.
	_ = os.Chmod(tmpPath, perm)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", filepath.Base(path), err)
	}
	renamed = true

	// Make the rename itself durable. Not every filesystem supports syncing a
	// directory, so this is best effort.
	if directory, err := os.Open(dir); err == nil {
		_ = directory.Sync()
		_ = directory.Close()
	}
	return nil
}
