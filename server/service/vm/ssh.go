package vm

import (
	"NanoKVM-Server/proto"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	SSHScript   = "/etc/init.d/S50sshd"
	SSHStopFlag = "/etc/kvm/ssh_stop"
)

func (s *Service) GetSSHState(c *gin.Context) {
	var rsp proto.Response

	state := &proto.GetSSHStateRsp{
		Enabled:      isSSHEnabled(),
		Running:      sshdPid() != 0,
		Port:         22,
		KeysOnly:     keysOnlyEnabled(),
		HostKeys:     readHostKeys(),
		RootPassword: rootPasswordState(),
	}
	if config, err := sshdEffectiveConfig(); err == nil {
		state.Port = sshPort(config)
		state.PasswordAuth = config["passwordauthentication"]
	}
	if lines, err := readAuthorizedKeys(); err == nil {
		state.KeyCount = len(keysOf(lines))
	}

	rsp.OkRspWithData(c, state)
}

func (s *Service) EnableSSH(c *gin.Context) {
	var rsp proto.Response

	command := fmt.Sprintf("%s permanent_on", SSHScript)
	err := exec.Command("sh", "-c", command).Run()
	if err != nil {
		log.Errorf("failed to run SSH script: %s", err)
		rsp.ErrRsp(c, -1, "operation failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("SSH enabled")
}

func (s *Service) DisableSSH(c *gin.Context) {
	var rsp proto.Response

	command := fmt.Sprintf("%s permanent_off", SSHScript)
	err := exec.Command("sh", "-c", command).Run()
	if err != nil {
		log.Errorf("failed to run SSH script: %s", err)
		rsp.ErrRsp(c, -1, "operation failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("SSH disabled")
}

func isSSHEnabled() bool {
	_, err := os.Stat(SSHStopFlag)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true
		}
	}

	return false
}
