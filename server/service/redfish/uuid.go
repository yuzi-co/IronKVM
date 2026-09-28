package redfish

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// UUIDFile keeps the service's UUID, so it stays the same across restarts.
const UUIDFile = "/etc/kvm/redfish-uuid"

// LoadUUID returns the UUID kept in path. The first call, or one that finds
// the file unreadable or holding something other than a UUID, generates a
// new one and writes it. If the write fails the new UUID is still returned,
// and lasts until the next restart.
func LoadUUID(path string) string {
	if data, err := os.ReadFile(path); err == nil {
		if id, err := uuid.Parse(strings.TrimSpace(string(data))); err == nil {
			return id.String()
		}
	}

	id := uuid.NewString()
	if err := writeFileAtomic(path, []byte(id+"\n")); err != nil {
		log.Warnf("redfish: keep the service UUID in %s: %s", path, err)
	}
	return id
}

// writeFileAtomic replaces path in one rename, so a power cut leaves either
// the old file or the new one.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
