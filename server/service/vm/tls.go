package vm

import (
	"fmt"
	"os/exec"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
)

func (s *Service) SetTls(c *gin.Context) {
	var req proto.SetTlsReq
	var rsp proto.Response

	err := proto.ParseFormRequest(c, &req)
	if err != nil {
		rsp.ErrRsp(c, -1, fmt.Sprintf("invalid arguments: %s", err))
		return
	}

	if req.Enabled {
		err = enableTls()
	} else {
		err = disableTls()
	}

	if err != nil {
		log.Errorf("failed to set TLS: %s", err)
		rsp.ErrRsp(c, -2, "operation failed")
		return
	}

	rsp.OkRsp(c)

	_ = exec.Command("sh", "-c", "/etc/init.d/S95nanokvm restart").Run()
}

// ensureTlsCert is a variable so a test can prove that switching HTTPS on does
// not mint a certificate when the one on disk still answers for this device.
var ensureTlsCert = utils.EnsureCert

func enableTls() error {
	// EnsureCert, not GenerateCert. Switching HTTPS off leaves the certificate
	// on disk, so switching it back on can reuse the one the operator already
	// installed in their trust store. Generating unconditionally voided that
	// trust on every switch, and a self-signed certificate has to be trusted by
	// hand, so the cost of getting this wrong is a manual step every time.
	if err := ensureTlsCert(); err != nil {
		return err
	}

	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.Proto = "https"
	conf.Cert.Crt = utils.CertFile
	conf.Cert.Key = utils.KeyFile

	if err := config.Write(conf); err != nil {
		return err
	}

	return nil
}

func disableTls() error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.Proto = "http"

	if err := config.Write(conf); err != nil {
		return err
	}

	return nil
}
