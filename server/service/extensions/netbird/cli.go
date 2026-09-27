package netbird

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
)

// LogFile is where S98netbird points the daemon's log. A start that fails
// reports its last lines.
const LogFile = "/var/log/netbird.log"

const (
	// loginURLTimeout bounds how long the handler waits for the SSO URL.
	loginURLTimeout = 60 * time.Second
	// ssoLife bounds an SSO login nobody finishes: netbird up has no timeout
	// of its own and would wait for the browser forever.
	ssoLife = 10 * time.Minute
	// upTimeout bounds up and a join with a setup key, which talk to the
	// management server.
	upTimeout = 2 * time.Minute
	// cliTimeout bounds down, deregister and version.
	cliTimeout = 30 * time.Second
	// waitDelay bounds how long a command's pipes may outlive it.
	waitDelay = 2 * time.Second
)

// statusTimeout bounds `netbird status`, which the page runs on every visit.
// A variable for the tests.
var statusTimeout = 15 * time.Second

type Cli struct{}

func NewCli() *Cli {
	return &Cli{}
}

// run runs the CLI with a deadline. WaitDelay closes the pipes if the CLI
// leaves a child holding them, so the deadline holds.
func run(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, NetbirdPath, args...)
	cmd.WaitDelay = waitDelay
	return vpn.Run(cmd)
}

// Start runs the boot script from the package copy. Start at boot is the boot
// route's alone.
func (c *Cli) Start() error {
	if err := utils.EnsurePermission(NetbirdPath, 0o100); err != nil {
		return err
	}
	return vpn.Script(addon.NetBird.Script(), "start", LogFile)
}

func (c *Cli) Restart() error {
	return vpn.Script(addon.NetBird.Script(), "restart", LogFile)
}

func (c *Cli) Stop() error {
	return vpn.Script(addon.NetBird.Script(), "stop", "")
}

func (c *Cli) Up() error {
	_, err := run(upTimeout, "up")
	return err
}

func (c *Cli) Down() error {
	_, err := run(cliTimeout, "down")
	return err
}

// Deregister removes this peer from the NetBird account and deletes its
// configuration. It is NetBird's logout; joining again needs a new key or SSO.
func (c *Cli) Deregister() error {
	_, err := run(cliTimeout, "deregister")
	return err
}

func (c *Cli) Status() (*NbStatus, error) {
	out, err := run(statusTimeout, "status", "--json")
	if err != nil {
		return nil, err
	}
	return parseStatus(out)
}

// Version is the installed CLI's version, the first line of `netbird version`.
func (c *Cli) Version() (string, error) {
	out, err := run(cliTimeout, "version")
	if err != nil {
		return "", err
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(first), nil
}

// JoinWithSetupKey runs `netbird up --setup-key`. The key is never logged,
// and a CLI that echoes it has it replaced in the error, which reaches the
// page and the server log.
func (c *Cli) JoinWithSetupKey(key string) error {
	_, err := run(upTimeout, "up", "--setup-key", key)
	return redact(err, key)
}

// LoginSSO starts an SSO login and returns the URL. NetBird prints it on
// stdout with --no-browser.
func (c *Cli) LoginSSO() (string, error) {
	cmd := exec.Command(NetbirdPath, "up", "--no-browser")
	return vpn.LoginURL(cmd, true, loginURLTimeout, ssoLife)
}

func redact(err error, secret string) error {
	if err == nil || secret == "" {
		return err
	}
	var ce *vpn.CmdError
	if errors.As(err, &ce) {
		ce.Tail = strings.ReplaceAll(ce.Tail, secret, "***")
		ce.Err = errors.New(strings.ReplaceAll(ce.Err.Error(), secret, "***"))
		return ce
	}
	return errors.New(strings.ReplaceAll(err.Error(), secret, "***"))
}
