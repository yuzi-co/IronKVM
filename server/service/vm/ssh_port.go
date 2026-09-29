package vm

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/netboot"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// The SSH port works like the keys-only setting beside it. /etc/kvm/ssh_port
// holds the owner's choice on /data, and is absent for the default. The
// drop-in is on the slot's root filesystem, so S50sshd writes it again from
// the file at every start, and the server writes the same line when the owner
// changes the port.
//
// Port is the one keyword sshd does not take the first value of: every Port
// line adds a listener. The image's sshd_config leaves "#Port 22" commented
// out, so the drop-in is the only one, and setSSHPort asks sshd to prove it
// before the change is kept.
var (
	sshPortFile   = "/etc/kvm/ssh_port"
	sshPortDropIn = "/etc/ssh/sshd_config.d/ironkvm-port.conf"

	// tcpListeners names the TCP ports something on the board listens on.
	tcpListeners = readTCPListeners
	// reservedPorts are the ports IronKVM itself listens on, or will once a
	// setting is turned on. sshd must not take one of them.
	reservedPorts = ironkvmPorts
)

const defaultSSHPort = 22

var procNetTCP = []string{"/proc/net/tcp", "/proc/net/tcp6"}

var (
	errPortInvalid     = errors.New("the port must be a number from 1 to 65535")
	errPortReserved    = errors.New("IronKVM uses this port")
	errPortInUse       = errors.New("another program listens on this port")
	errPortNotHonoured = errors.New("sshd on this image does not take the port setting")
)

// portConfig must match what S50sshd writes at boot.
func portConfig(port int) string {
	return fmt.Sprintf("# Written by IronKVM: Settings > SSH > Port.\nPort %d\n", port)
}

func validSSHPort(port int) bool {
	return port >= 1 && port <= 65535
}

func ironkvmPorts() []int {
	conf := config.GetInstance()
	return []int{80, 443, conf.Port.Http, conf.Port.Https, conf.VNC.WithDefaults().Port, netboot.HTTPPort}
}

// sshdPorts lists every port sshd would listen on, in its order.
func sshdPorts(config map[string]string) []int {
	var ports []int
	for _, value := range strings.Split(config["port"], ",") {
		if port, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && port > 0 {
			ports = append(ports, port)
		}
	}
	return ports
}

// listenPorts lists the port of each ListenAddress sshd reports, such as
// 0.0.0.0:2222 or [::]:2222. A ListenAddress with its own port overrides Port,
// so a board configured that way would ignore the drop-in.
func listenPorts(config map[string]string) []int {
	var ports []int
	for _, value := range strings.Split(config["listenaddress"], ",") {
		address, _, _ := strings.Cut(strings.TrimSpace(value), " ")
		i := strings.LastIndex(address, ":")
		if i < 0 {
			continue
		}
		if port, err := strconv.Atoi(address[i+1:]); err == nil {
			ports = append(ports, port)
		}
	}
	return ports
}

// honoursPort says whether sshd's effective configuration listens on port
// and nothing else.
func honoursPort(config map[string]string, port int) bool {
	ports := sshdPorts(config)
	if len(ports) != 1 || ports[0] != port {
		return false
	}
	for _, listen := range listenPorts(config) {
		if listen != port {
			return false
		}
	}
	return true
}

// parseTCPListeners reads /proc/net/tcp or tcp6 and returns the local ports
// of the sockets in LISTEN state (st 0A).
func parseTCPListeners(data []byte, into map[int]bool) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || fields[3] != "0A" {
			continue
		}
		i := strings.LastIndex(fields[1], ":")
		if i < 0 {
			continue
		}
		if port, err := strconv.ParseUint(fields[1][i+1:], 16, 16); err == nil {
			into[int(port)] = true
		}
	}
}

func readTCPListeners() (map[int]bool, error) {
	ports := map[int]bool{}
	for _, path := range procNetTCP {
		data, err := os.ReadFile(path)
		if err != nil {
			// A kernel without IPv6 has no tcp6.
			if errors.Is(err, os.ErrNotExist) && path != procNetTCP[0] {
				continue
			}
			return nil, err
		}
		parseTCPListeners(data, ports)
	}
	return ports, nil
}

// checkPortFree refuses a port IronKVM uses, or one another program listens
// on. The ports sshd already has are its own and stay allowed.
func checkPortFree(port int, sshdHas []int) error {
	if slices.Contains(sshdHas, port) {
		return nil
	}
	if slices.Contains(reservedPorts(), port) {
		return errPortReserved
	}
	listening, err := tcpListeners()
	if err != nil {
		return err
	}
	if listening[port] {
		return errPortInUse
	}
	return nil
}

// restoreFile puts back what a file held before a change: its bytes, or its
// absence.
func restoreFile(path string, data []byte, existed bool) {
	if !existed {
		_ = os.Remove(path)
		return
	}
	if err := writeFileAtomic(path, data, 0o644); err != nil {
		log.Errorf("failed to restore %s: %s", path, err)
	}
}

func setSSHPort(port int) error {
	if !validSSHPort(port) {
		return errPortInvalid
	}

	sshMu.Lock()
	defer sshMu.Unlock()

	var current []int
	if config, err := sshdEffectiveConfig(); err == nil {
		current = sshdPorts(config)
	}
	if err := checkPortFree(port, current); err != nil {
		return err
	}

	previous, readErr := os.ReadFile(sshPortDropIn)
	existed := readErr == nil
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return readErr
	}

	if port == defaultSSHPort {
		if err := os.Remove(sshPortDropIn); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(sshPortDropIn), 0o755); err != nil {
			return err
		}
		if err := writeFileAtomic(sshPortDropIn, []byte(portConfig(port)), 0o644); err != nil {
			return err
		}
	}

	// As with keys-only: sshd -T proves the configuration parses before
	// SIGHUP makes sshd re-execute with it, and says which ports it would
	// listen on. Anything but the one asked for goes back.
	config, err := sshdEffectiveConfig()
	if err != nil || !honoursPort(config, port) {
		restoreFile(sshPortDropIn, previous, existed)
		if err != nil {
			log.Errorf("sshd -T failed with the port drop-in: %s", err)
		}
		return errPortNotHonoured
	}

	// The file is the setting the next boot reads.
	if port == defaultSSHPort {
		err = os.Remove(sshPortFile)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	} else {
		err = writeFileAtomic(sshPortFile, []byte(strconv.Itoa(port)+"\n"), 0o644)
	}
	if err != nil {
		restoreFile(sshPortDropIn, previous, existed)
		return err
	}

	// The listener re-executes on the new port. Open sessions are their own
	// processes and keep running.
	if err := reloadSSHD(); err != nil {
		log.Errorf("failed to reload sshd: %s", err)
		return errReloadFailed
	}
	advertiseSSHPort(port)
	return nil
}

// The avahi package ships _ssh._tcp and _sftp-ssh._tcp services with port 22
// written into them, so mDNS would keep sending clients to the old port.
// S50sshd rewrites them the same way at boot.
var (
	avahiSSHServices = []string{"/etc/avahi/services/ssh.service", "/etc/avahi/services/sftp-ssh.service"}
	reloadAvahi      = func() error { return exec.Command("avahi-daemon", "-r").Run() }
	avahiPortRe      = regexp.MustCompile(`<port>[0-9]+</port>`)
)

// withAdvertisedPort returns the service file with every port set to port.
func withAdvertisedPort(data []byte, port int) []byte {
	return avahiPortRe.ReplaceAll(data, []byte(fmt.Sprintf("<port>%d</port>", port)))
}

// advertiseSSHPort points the avahi SSH services at port. A missing file
// (mDNS not installed) is skipped, and a failure only costs the announcement,
// so it is logged rather than returned.
func advertiseSSHPort(port int) {
	changed := false
	for _, path := range avahiSSHServices {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		updated := withAdvertisedPort(data, port)
		if bytes.Equal(updated, data) {
			continue
		}
		if err := writeFileAtomic(path, updated, 0o644); err != nil {
			log.Errorf("failed to update %s: %s", path, err)
			continue
		}
		changed = true
	}
	if changed {
		// avahi-daemon -r fails when mDNS is off, which is fine: it reads the
		// files when it starts.
		_ = reloadAvahi()
	}
}

func (s *Service) SetSSHPort(c *gin.Context) {
	var rsp proto.Response
	var req proto.SetSSHPortReq
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	err := setSSHPort(req.Port)
	switch {
	case errors.Is(err, errPortInvalid):
		rsp.ErrRsp(c, -2, err.Error())
	case errors.Is(err, errPortReserved):
		rsp.ErrRsp(c, -3, err.Error())
	case errors.Is(err, errPortInUse):
		rsp.ErrRsp(c, -4, err.Error())
	case errors.Is(err, errPortNotHonoured):
		rsp.ErrRsp(c, -5, err.Error())
	case errors.Is(err, errReloadFailed):
		rsp.ErrRsp(c, -6, err.Error())
	case err != nil:
		log.Errorf("failed to change the ssh port: %s", err)
		rsp.ErrRsp(c, -7, fmt.Sprintf("failed to change the ssh port: %s", err))
	default:
		log.Infof("ssh port set to %d", req.Port)
		rsp.OkRsp(c)
	}
}
