// Package atomicfile replaces files whole.
package atomicfile

import (
	"os"
	"path/filepath"
)

// WriteFile replaces path with data so that a reader, or a power cut,
// sees either the old file or the new one and never a truncated mix. The data
// goes to a hidden temporary file in the same directory, which is synced and
// renamed over path, and the directory is then synced so the rename itself
// survives a power cut. The temporary file is removed on any failure.
//
// The file gets mode exactly, whatever the umask. The temporary file is
// created 0600 and changed to mode before anything is written, so a secret is
// never readable by others, not even for a moment. A filesystem without
// permission bits, such as exFAT on /data, refuses the change; that is not an
// error, and the file then has whatever the mount gives it.
//
// The directory must exist. Renaming over an existing file works on exFAT
// under Linux as well as on ext4.
func WriteFile(path string, data []byte, mode os.FileMode) error {
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

	_ = tmp.Chmod(mode)
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}
	renamed = true

	// Not every filesystem can sync a directory, so this is best effort.
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}
