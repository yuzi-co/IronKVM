package redfish

import (
	"os"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/utils/atomicfile"
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
	if err := atomicfile.WriteFile(path, []byte(id+"\n"), 0o644); err != nil {
		log.Warnf("redfish: keep the service UUID in %s: %s", path, err)
	}
	return id
}
