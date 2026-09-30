package tailscale

import (
	"context"
	"os/exec"
	"time"

	"NanoKVM-Server/service/extensions/addon"
)

// Installed reports whether both binaries are in place. It is two stats.
func Installed() bool {
	return isInstalled()
}

// Running reports whether tailscaled runs now. It reads a pid file and a
// command line, and runs nothing.
func Running() bool {
	return addon.Running(addon.Tailscale)
}

// Connected asks the CLI whether this node is up on its tailnet. It runs the
// CLI, so a caller that asks often keeps the answer for a while.
func Connected(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, TailscalePath, "status", "--json")
	cmd.WaitDelay = time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	st, err := parseStatus(out)
	return err == nil && st.BackendState == "Running"
}
