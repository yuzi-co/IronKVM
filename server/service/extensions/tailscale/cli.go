package tailscale

import (
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	ScriptPath       = "/etc/init.d/S98tailscaled"
	ScriptBackupPath = "/kvmapp/system/init.d/S98tailscaled"

	// loginURLTimeout bounds how long the handler waits for the login URL.
	// The command itself runs for ten minutes waiting on the browser.
	loginURLTimeout = 60 * time.Second
)

type Cli struct{}

type TsStatus struct {
	BackendState string `json:"BackendState"`

	Self struct {
		HostName     string   `json:"HostName"`
		TailscaleIPs []string `json:"TailscaleIPs"`
	} `json:"Self"`

	CurrentTailnet struct {
		Name string `json:"Name"`
	} `json:"CurrentTailnet"`
}

func NewCli() *Cli {
	return &Cli{}
}

func (c *Cli) Start() error {
	for _, filePath := range []string{TailscalePath, TailscaledPath} {
		if err := utils.EnsurePermission(filePath, 0o100); err != nil {
			return err
		}
	}

	commands := []string{
		fmt.Sprintf("cp -f %s %s", ScriptBackupPath, ScriptPath),
		fmt.Sprintf("%s start", ScriptPath),
	}

	command := strings.Join(commands, " && ")
	if err := exec.Command("sh", "-c", command).Run(); err != nil {
		return err
	}
	return recordEnabled(true)
}

func (c *Cli) Restart() error {
	commands := []string{
		fmt.Sprintf("cp -f %s %s", ScriptBackupPath, ScriptPath),
		fmt.Sprintf("%s restart", ScriptPath),
	}

	command := strings.Join(commands, " && ")
	return exec.Command("sh", "-c", command).Run()
}

func (c *Cli) Stop() error {
	command := fmt.Sprintf("%s stop", ScriptPath)
	err := exec.Command("sh", "-c", command).Run()
	if err != nil {
		return err
	}

	if err := recordEnabled(false); err != nil {
		return err
	}
	return os.Remove(ScriptPath)
}

// recordEnabled keeps "start at boot" on /data on a distribution image, where
// /etc/init.d belongs to the slot and a new image would forget it. S04addons
// reads it at the next boot. Anywhere else the script in /etc/init.d is the
// whole record, as it always was.
func recordEnabled(on bool) error {
	if !addon.OnData() {
		return nil
	}
	return addon.SetEnabled(addonSpec().Name, on)
}

func (c *Cli) Up() error {
	command := "tailscale up --accept-dns=false"
	return exec.Command("sh", "-c", command).Run()
}

func (c *Cli) Down() error {
	command := "tailscale down"
	return exec.Command("sh", "-c", command).Run()
}

func (c *Cli) Status() (*TsStatus, error) {
	command := "tailscale status --json"
	cmd := exec.Command("sh", "-c", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// output is not in standard json format
	if outputStr := string(output); !strings.HasPrefix(outputStr, "{") {
		index := strings.Index(outputStr, "{")
		if index == -1 {
			return nil, errors.New("unknown output")
		}

		output = []byte(outputStr[index:])
	}

	var status TsStatus
	err = json.Unmarshal(output, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Cli) Login() (string, error) {
	// No shell: killing "sh -c tailscale ..." leaves tailscale holding the
	// stderr pipe, so the timeout could never take effect.
	cmd := exec.Command("tailscale", "login", "--accept-dns=false", "--timeout=10m")

	return vpn.LoginURL(cmd, false, loginURLTimeout, 0)
}

func (c *Cli) Logout() error {
	command := "tailscale logout"
	return exec.Command("sh", "-c", command).Run()
}
