package vm

import (
	"fmt"
	"os/exec"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
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
		err = enableTlsConfig()
	} else {
		err = disableTlsConfig()
	}

	if err != nil {
		log.Errorf("failed to set TLS: %s", err)
		rsp.ErrRsp(c, -2, "operation failed")
		return
	}

	if !req.Enabled {
		// This response is the last one the browser gets over HTTPS, and it is
		// the only chance to take the session cookie back.
		//
		// The cookie was set with Secure while HTTPS was on, so the browser
		// will not send it over plain HTTP and will not let a plain HTTP
		// response replace or delete it either. Leaving it in place logs the
		// operator into nothing: the login form posts, the server answers with
		// a new cookie the browser refuses to store over the old Secure one,
		// and the page returns to the login form with nothing in the log to
		// say why. Deleting it here, while this connection is still encrypted,
		// costs one deliberate log in and avoids that.
		middleware.ClearSessionCookie(c, true)
	}

	rsp.OkRsp(c)

	restartServer()
}

// restartServer is a variable so a test can drive SetTls without shelling out
// to an init script that is not there.
var restartServer = func() {
	_ = exec.Command("sh", "-c", "/etc/init.d/S95nanokvm restart").Run()
}

// ensureTlsCert is a variable so a test can prove that switching HTTPS on does
// not mint a certificate when the one on disk still answers for this device.
var ensureTlsCert = utils.EnsureCert

// enableTlsConfig and disableTlsConfig are variables for the same reason: both
// write the real configuration file, and what SetTls does around them is
// testable without that.
var (
	enableTlsConfig  = enableTls
	disableTlsConfig = disableTls
)

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
