package utils

import (
	"errors"
	"os"
	"os/exec"
)

// IdentityScript carries the board's identity between slots: /etc/kvm and
// /root/.ssh are bound from /data, and /etc/shadow and the ssh host keys are
// copied there. It is absent on an upstream image.
const IdentityScript = "/etc/init.d/S02identity"

// SaveIdentity copies the credentials the board just changed to /data, so the
// next boot of any slot restores them rather than the ones they replaced.
//
// A board without the script is an upstream image with no slot layout, where
// there is nothing to write back to and nothing to lose. That is not an error.
func SaveIdentity() error {
	if _, err := os.Stat(IdentityScript); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return exec.Command(IdentityScript, "save").Run()
}
