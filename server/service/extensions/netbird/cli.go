package netbird

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
// route's alone, but while it is on the copy in /etc/init.d is refreshed
// first, so a boot runs the script this server ships.
func (c *Cli) Start() error {
	if err := utils.EnsurePermission(NetbirdPath, 0o100); err != nil {
		return err
	}
	if err := addon.RefreshInitd(addon.NetBird); err != nil {
		return err
	}
	return vpn.Script(addon.NetBird.Script(), "start", LogFile)
}

func (c *Cli) Restart() error {
	return vpn.Script(addon.NetBird.Script(), "restart", LogFile)
}

// Stop fails only when the daemon is still there afterwards.
func (c *Cli) Stop() error {
	return vpn.StopDaemon(addon.NetBird)
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

// KeyDir is where a setup key is written for the CLI to read, mode 0600, and
// removed as soon as the CLI returns. /run is a tmpfs, so the key never reaches
// the SD card, and a file keeps it out of the process list, where an argument
// would show. A variable for the tests.
var KeyDir = "/run"

// setupKeyFormat is how NetBird's management server makes a setup key: an
// upper-case UUID (management/server/types/setupkey.go at v0.78.2).
var setupKeyFormat = regexp.MustCompile(`^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$`)

// ErrSetupKeyFormat does not repeat the key: what was pasted may be another
// secret altogether.
var ErrSetupKeyFormat = errors.New("the setup key is not in NetBird's format: " +
	"32 hexadecimal digits in groups of 8, 4, 4, 4 and 12, separated by dashes")

// NormalizeSetupKey trims the key and puts it in upper case, as the management
// server stores it, or refuses it when it is not a NetBird key.
func NormalizeSetupKey(key string) (string, error) {
	key = strings.ToUpper(strings.TrimSpace(key))
	if !setupKeyFormat.MatchString(key) {
		return "", ErrSetupKeyFormat
	}
	return key, nil
}

// JoinWithSetupKey runs `netbird up --setup-key-file`. The key is never
// logged, and a CLI that echoes it has it masked in its output before the
// output is cut to the tail that reaches the page and the server log.
func (c *Cli) JoinWithSetupKey(key string) error {
	key, err := NormalizeSetupKey(key)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(KeyDir, "netbird-setup-key-*")
	if err != nil {
		return fmt.Errorf("failed to write the setup key for the CLI: %w", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	// CreateTemp makes the file 0600 already; the umask cannot widen it.
	_, werr := f.WriteString(key + "\n")
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return fmt.Errorf("failed to write the setup key for the CLI: %w", werr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), upTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, NetbirdPath, "up", "--setup-key-file", f.Name())
	cmd.WaitDelay = waitDelay
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	masked := strings.NewReplacer(key, "***", strings.ToLower(key), "***").Replace(string(out))
	return &vpn.CmdError{Err: err, Tail: vpn.Tail([]byte(masked), vpn.TailLines)}
}

// LoginSSO starts an SSO login and returns the URL. NetBird prints it on
// stdout with --no-browser.
func (c *Cli) LoginSSO() (string, error) {
	cmd := exec.Command(NetbirdPath, "up", "--no-browser")
	return vpn.LoginURL(cmd, true, loginURLTimeout, ssoLife)
}
