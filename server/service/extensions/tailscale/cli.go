package tailscale

import (
	"os/exec"
	"strings"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
)

// loginURLTimeout bounds how long the handler waits for the login URL. The
// command itself runs for ten minutes waiting on the browser.
const loginURLTimeout = 60 * time.Second

type Cli struct{}

func NewCli() *Cli {
	return &Cli{}
}

// Start runs the boot script from the package copy. It no longer touches
// start at boot, which is the boot switch's, through addon.SetBoot, but while
// start at boot is on it refreshes the copy in /etc/init.d first.
func (c *Cli) Start() error {
	for _, filePath := range []string{TailscalePath, TailscaledPath} {
		if err := utils.EnsurePermission(filePath, 0o100); err != nil {
			return err
		}
	}
	if err := addon.RefreshInitd(addon.Tailscale); err != nil {
		return err
	}
	return vpn.Script(addon.Tailscale.Script(), "start", "")
}

func (c *Cli) Restart() error {
	return vpn.Script(addon.Tailscale.Script(), "restart", "")
}

// Stop fails only when the daemon is still there afterwards. S98tailscaled
// exits 0 either way and says FAIL, which vpn.Script reads.
func (c *Cli) Stop() error {
	return vpn.StopDaemon(addon.Tailscale)
}

func (c *Cli) Up() error {
	_, err := vpn.Run(exec.Command(TailscalePath, "up", "--accept-dns=false"))
	return err
}

func (c *Cli) Down() error {
	_, err := vpn.Run(exec.Command(TailscalePath, "down"))
	return err
}

func (c *Cli) Status() (*TsStatus, error) {
	output, err := exec.Command(TailscalePath, "status", "--json").CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseStatus(output)
}

func (c *Cli) Login() (string, error) {
	// No shell: killing "sh -c tailscale ..." leaves tailscale holding the
	// stderr pipe, so the timeout could never take effect.
	cmd := exec.Command(TailscalePath, "login", "--accept-dns=false", "--timeout=10m")

	return vpn.LoginURL(cmd, false, loginURLTimeout, 0)
}

func (c *Cli) Logout() error {
	_, err := vpn.Run(exec.Command(TailscalePath, "logout"))
	return err
}

// Version is the installed CLI's version: the first line of `tailscale version`.
func (c *Cli) Version() (string, error) {
	out, err := vpn.Run(exec.Command(TailscalePath, "version"))
	if err != nil {
		return "", err
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(first), nil
}
