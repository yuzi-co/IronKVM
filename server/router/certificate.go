package router

import (
	"crypto/tls"
	"os"

	"NanoKVM-Server/config"
	"NanoKVM-Server/utils"
)

// servingCertificate is the certificate the HTTPS listener serves, including
// one uploaded since the server started. VNC's TLS layer and RTSPS serve it
// too. A board that serves plain HTTP has none in memory, so it loads the
// configured pair, and when there is none yet the shipped one is made the
// same way the HTTPS switch makes it.
func servingCertificate() (*tls.Certificate, error) {
	if cert := utils.ActiveCert.Current(); cert != nil {
		return cert, nil
	}
	conf := config.GetInstance()
	if cert, err := tls.LoadX509KeyPair(conf.Cert.Crt, conf.Cert.Key); err == nil {
		return &cert, nil
	}

	if err := utils.EnsureCert(); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(utils.CertFile, utils.KeyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// hasServingCertificate reports whether servingCertificate has a pair to give
// without making one. A settings page polls it, so it only looks for files.
func hasServingCertificate() bool {
	if utils.ActiveCert.Current() != nil {
		return true
	}
	conf := config.GetInstance()
	return pairExists(conf.Cert.Crt, conf.Cert.Key) || pairExists(utils.CertFile, utils.KeyFile)
}

func pairExists(certPath, keyPath string) bool {
	if certPath == "" || keyPath == "" {
		return false
	}
	for _, path := range []string{certPath, keyPath} {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			return false
		}
	}
	return true
}
