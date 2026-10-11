package vm

import (
	"crypto/tls"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
)

// The HTTPS certificate: an uploaded one, or the generated self-signed one.
//
// The uploaded pair is written to utils.UploadedCertFile and UploadedKeyFile
// in /etc/kvm, and server.yaml's cert paths are pointed at it. /etc/kvm is
// bound from /data/identity on a board with slots, so the pair lives on /data
// and is the same on both slots. The listener serves from memory
// (utils.ActiveCert), so installing or restoring takes effect on the next
// handshake, without a restart. Neither the key nor its contents are logged.

// The paths and hooks below are variables so a test can run the handlers
// against a scratch directory.
var (
	uploadedCertFile = utils.UploadedCertFile
	uploadedKeyFile  = utils.UploadedKeyFile
	generatedCert    = utils.CertFile
	generatedKey     = utils.KeyFile

	readTLSConfig  = config.Read
	writeTLSConfig = config.Write
	ensureSelfCert = utils.EnsureCert
	certIdentity   = utils.CurrentCertIdentity
	servingCert    = utils.ActiveCert
)

// certMutex serialises the handlers that change the certificate. Each one
// writes files and the configuration file and then reloads, and two at once
// could leave the configuration naming one pair and memory serving another.
var certMutex sync.Mutex

type tlsCertRsp struct {
	// Uploaded is true when an uploaded certificate is configured, false
	// for the generated self-signed one.
	Uploaded bool `json:"uploaded"`
	// HTTPS is whether the HTTPS listener is on. An uploaded certificate is
	// kept while it is off, and served once it is on again.
	HTTPS bool            `json:"https"`
	Info  *utils.CertInfo `json:"info"`
}

// certFailure answers a certificate the validation refused. The page
// translates by kind and falls back to the message.
func certFailure(c *gin.Context, err error) {
	var rsp proto.Response
	var certErr *utils.CertError
	if errors.As(err, &certErr) {
		rsp.Data = gin.H{"kind": certErr.Kind}
		rsp.ErrRsp(c, -2, certErr.Message)
		return
	}
	rsp.ErrRsp(c, -2, "invalid certificate")
}

// GetTlsCert describes the configured certificate.
func (s *Service) GetTlsCert(c *gin.Context) {
	var rsp proto.Response
	conf, err := readTLSConfig()
	if err != nil {
		rsp.ErrRsp(c, -1, "failed to read the TLS setting")
		return
	}
	uploaded := conf.Cert.Crt == uploadedCertFile
	path := generatedCert
	if uploaded {
		path = uploadedCertFile
	}
	data := &tlsCertRsp{Uploaded: uploaded, HTTPS: conf.Proto == "https"}
	if info, err := utils.DescribeCertFile(path, certIdentity()); err == nil {
		data.Info = info
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Warnf("failed to read the TLS certificate %s: %s", path, err)
	}
	rsp.OkRspWithData(c, data)
}

// CheckTlsCert validates a certificate and key without installing them, so
// the page can show what would be served before the operator commits.
func (s *Service) CheckTlsCert(c *gin.Context) {
	var req proto.TlsCertReq
	var rsp proto.Response
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "a certificate and a key are both required")
		return
	}
	info, err := utils.ValidateCertPair([]byte(req.Cert), []byte(req.Key), time.Now(), certIdentity())
	if err != nil {
		certFailure(c, err)
		return
	}
	rsp.OkRspWithData(c, info)
}

// SetTlsCert installs an uploaded certificate and key and serves them from
// the next handshake on.
func (s *Service) SetTlsCert(c *gin.Context) {
	var req proto.TlsCertReq
	var rsp proto.Response
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "a certificate and a key are both required")
		return
	}
	certPEM, keyPEM := []byte(req.Cert), []byte(req.Key)
	info, err := utils.ValidateCertPair(certPEM, keyPEM, time.Now(), certIdentity())
	if err != nil {
		certFailure(c, err)
		return
	}

	certMutex.Lock()
	defer certMutex.Unlock()

	if err := utils.WriteCertPair(uploadedCertFile, uploadedKeyFile, certPEM, keyPEM); err != nil {
		log.Errorf("failed to store the uploaded certificate: %s", err)
		rsp.ErrRsp(c, -3, "failed to store the certificate")
		return
	}
	if err := pointConfigAt(uploadedCertFile, uploadedKeyFile); err != nil {
		log.Errorf("failed to configure the uploaded certificate: %s", err)
		rsp.ErrRsp(c, -3, "failed to store the certificate")
		return
	}
	if err := servingCert.Load(uploadedCertFile, uploadedKeyFile); err != nil {
		log.Errorf("failed to load the uploaded certificate: %s", err)
		rsp.ErrRsp(c, -4, "the certificate was stored but could not be loaded")
		return
	}

	rsp.OkRspWithData(c, info)
	log.Infof("uploaded TLS certificate installed: %s, SHA-256 %s", info.Subject, info.Fingerprint)
}

// RestoreTlsCert goes back to the generated self-signed certificate and
// deletes the uploaded pair.
func (s *Service) RestoreTlsCert(c *gin.Context) {
	var rsp proto.Response

	certMutex.Lock()
	defer certMutex.Unlock()

	if err := ensureSelfCert(); err != nil {
		log.Errorf("failed to generate the self-signed certificate: %s", err)
		rsp.ErrRsp(c, -1, "failed to generate the self-signed certificate")
		return
	}
	if err := pointConfigAt(generatedCert, generatedKey); err != nil {
		log.Errorf("failed to configure the self-signed certificate: %s", err)
		rsp.ErrRsp(c, -2, "failed to restore the self-signed certificate")
		return
	}
	if err := servingCert.Load(generatedCert, generatedKey); err != nil {
		log.Errorf("failed to load the self-signed certificate: %s", err)
		rsp.ErrRsp(c, -3, "failed to load the self-signed certificate")
		return
	}
	for _, path := range []string{uploadedKeyFile, uploadedCertFile} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Warnf("failed to remove %s: %s", path, err)
		}
	}

	rsp.OkRsp(c)
	log.Info("restored the self-signed TLS certificate")
}

func pointConfigAt(certPath, keyPath string) error {
	conf, err := readTLSConfig()
	if err != nil {
		return err
	}
	if conf.Cert.Crt == certPath && conf.Cert.Key == keyPath {
		return nil
	}
	conf.Cert.Crt = certPath
	conf.Cert.Key = keyPath
	return writeTLSConfig(conf)
}

// uploadedPairUsable reports whether the configuration names the uploaded
// pair and that pair still loads, so switching HTTPS on keeps it.
func uploadedPairUsable(conf *config.Config) bool {
	if conf.Cert.Crt != uploadedCertFile || conf.Cert.Key != uploadedKeyFile {
		return false
	}
	_, err := tls.LoadX509KeyPair(uploadedCertFile, uploadedKeyFile)
	return err == nil
}
