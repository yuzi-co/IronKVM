package netbird

import (
	"context"
	"os/exec"

	"NanoKVM-Server/service/extensions/addon"
)

// Installed reports whether the binary is in place. It is one stat.
func Installed() bool {
	return isInstalled()
}

// Running reports whether the daemon runs now. It reads a pid file and a
// command line, and runs nothing.
func Running() bool {
	return addon.Running(addon.NetBird)
}

// Connected asks the CLI whether the daemon is connected to its network, not
// merely connecting. It runs the CLI, so a caller that asks often keeps the
// answer for a while.
func Connected(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, NetbirdPath, "status", "--json")
	cmd.WaitDelay = waitDelay
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	st, err := parseStatus(out)
	return err == nil && st.DaemonStatus == "Connected"
}
