package utils

import (
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GoMemLimitFile is a var so tests can point it somewhere writable.
var GoMemLimitFile = "/etc/kvm/GOMEMLIMIT"

func InitGoMemLimit() {
	if !IsGoMemLimitExist() {
		return
	}

	limit, err := GetGoMemLimit()
	if err != nil {
		return
	}

	debug.SetMemoryLimit(limit * 1024 * 1024)
	log.Debugf("set GOMEMLIMIT to %d MB", limit)
}

func SetGoMemLimit(limit int64) error {
	memoryLimit := max(limit, 50)
	debug.SetMemoryLimit(memoryLimit * 1024 * 1024)

	log.Debugf("set GOMEMLIMIT to %d MB", limit)

	data := []byte(fmt.Sprintf("%d", limit))
	err := os.WriteFile(GoMemLimitFile, data, 0o644)
	if err != nil {
		log.Errorf("failed to write GOMEMLIMIT: %s", err)
		return err
	}

	return nil
}

func GetGoMemLimit() (int64, error) {
	data, err := os.ReadFile(GoMemLimitFile)
	if err != nil {
		log.Errorf("failed to read GOMEMLIMIT: %s", err)
		return 0, err
	}

	content := strings.TrimSpace(string(data))
	limit, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		log.Errorf("failed to parse GOMEMLIMIT: %s", err)
		return 0, err
	}

	return limit, nil
}

func DelGoMemLimit() error {
	// math.MaxInt64 is Go's "no limit" value. A literal 1GB is a real cap on
	// the 1GB boards, which is the opposite of turning the limit off.
	debug.SetMemoryLimit(math.MaxInt64)

	err := os.Remove(GoMemLimitFile)
	if err != nil {
		log.Errorf("failed to delete GOMEMLIMIT: %s", err)
		return err
	}

	return nil
}

func IsGoMemLimitExist() bool {
	_, err := os.Stat(GoMemLimitFile)
	return err == nil
}

// oldTailscaleLimit is what the Tailscale page's memory switch wrote to
// /etc/kvm/GOMEMLIMIT, byte for byte, before the init scripts derived the
// daemons' limit from the addons group.
const oldTailscaleLimit = "75"

// MigrateGoMemLimit removes a /etc/kvm/GOMEMLIMIT that holds exactly the old
// Tailscale value, and reports whether it did. Nothing writes that file for
// Tailscale any more, so a leftover would stay forever and cap this server at
// 75 MiB through InitGoMemLimit. S98tailscaled and S98netbird then fall back
// to seven eighths of the addons group's memory.high. Any other content is an
// owner's setting and stays. It runs at every start and is a no-op once the
// file is gone, so it logs once.
func MigrateGoMemLimit() bool {
	data, err := os.ReadFile(GoMemLimitFile)
	if err != nil || string(data) != oldTailscaleLimit {
		return false
	}
	if err := os.Remove(GoMemLimitFile); err != nil {
		log.Errorf("failed to remove the old Tailscale GOMEMLIMIT: %s", err)
		return false
	}
	log.Infof("removed %s, which held the old Tailscale memory limit of %s MiB; the init scripts derive it now",
		GoMemLimitFile, oldTailscaleLimit)
	return true
}
