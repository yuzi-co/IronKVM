package vm

import (
	"NanoKVM-Server/proto"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	AvahiDaemonPid          = "/run/avahi-daemon/pid"
	AvahiDaemonScript       = "/etc/init.d/S50avahi-daemon"
	AvahiDaemonBackupScript = "/kvmapp/system/init.d/S50avahi-daemon"

	// MdnsDisabledFile records that the owner turned mDNS off.
	//
	// Deleting AvahiDaemonScript is not enough on its own. That script lives
	// on the root slot, and on an A/B board the next image puts it back, so
	// mDNS came back on after every update. /etc/kvm is bound from /data, so
	// this file survives an image, and S50avahi-daemon does not start while
	// it exists.
	MdnsDisabledFile = "/etc/kvm/mdns_disabled"
)

func (s *Service) GetMdnsState(c *gin.Context) {
	var rsp proto.Response

	pid := getAvahiDaemonPid()

	rsp.OkRspWithData(c, &proto.GetMdnsStateRsp{
		Enabled: pid != "",
	})
}

func (s *Service) EnableMdns(c *gin.Context) {
	var rsp proto.Response

	// Before the running check, so a board whose daemon was started by hand
	// still loses a marker that would keep it off at the next boot.
	if err := recordMdnsState(MdnsDisabledFile, true); err != nil {
		log.Errorf("failed to clear %s: %s", MdnsDisabledFile, err)
		rsp.ErrRsp(c, -1, "failed to enable mdns")
		return
	}

	pid := getAvahiDaemonPid()
	if pid != "" {
		rsp.OkRsp(c)
		return
	}

	commands := []string{
		fmt.Sprintf("cp -f %s %s", AvahiDaemonBackupScript, AvahiDaemonScript),
		fmt.Sprintf("%s start", AvahiDaemonScript),
	}

	command := strings.Join(commands, " && ")
	err := exec.Command("sh", "-c", command).Run()
	if err != nil {
		log.Errorf("failed to start avahi-daemon: %s", err)
		rsp.ErrRsp(c, -1, "failed to enable mdns")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("avahi-daemon started")
}

func (s *Service) DisableMdns(c *gin.Context) {
	var rsp proto.Response

	// First, and whether or not the daemon runs now. A daemon that is down
	// for some other reason would otherwise come back at the next boot.
	if err := recordMdnsState(MdnsDisabledFile, false); err != nil {
		log.Errorf("failed to write %s: %s", MdnsDisabledFile, err)
		rsp.ErrRsp(c, -1, "failed to disable mdns")
		return
	}

	pid := getAvahiDaemonPid()
	if pid != "" {
		command := fmt.Sprintf("kill -9 %s", pid)
		err := exec.Command("sh", "-c", command).Run()
		if err != nil {
			log.Errorf("failed to stop avahi-daemon: %s", err)
			rsp.ErrRsp(c, -1, "failed to disable mdns")
			return
		}
	}

	_ = os.Remove(AvahiDaemonPid)
	_ = os.Remove(AvahiDaemonScript)

	rsp.OkRsp(c)
	log.Debugf("avahi-daemon stopped")
}

// recordMdnsState writes the marker for off and removes it for on.
func recordMdnsState(marker string, enabled bool) error {
	if enabled {
		if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return os.WriteFile(marker, nil, 0o644)
}

func getAvahiDaemonPid() string {
	if _, err := os.Stat(AvahiDaemonPid); err != nil {
		return ""
	}

	content, err := os.ReadFile(AvahiDaemonPid)
	if err != nil {
		log.Errorf("failed to read mdns pid: %s", err)
		return ""
	}

	return strings.ReplaceAll(string(content), "\n", "")
}
