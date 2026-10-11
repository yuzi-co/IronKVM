package vm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
	"NanoKVM-Server/utils"
)

type certFixture struct {
	t      *testing.T
	dir    string
	conf   config.Config
	router *gin.Engine
}

// selfSigned makes a throwaway P-256 server certificate for name.
func selfSigned(t *testing.T, name string) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: name},
		DNSNames:     []string{name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(90 * 24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
}

func newCertFixture(t *testing.T) *certFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	f := &certFixture{t: t, dir: t.TempDir()}
	f.conf.Proto = "https"

	saved := []any{uploadedCertFile, uploadedKeyFile, generatedCert, generatedKey, readTLSConfig, writeTLSConfig, ensureSelfCert, certIdentity, servingCert}
	t.Cleanup(func() {
		uploadedCertFile = saved[0].(string)
		uploadedKeyFile = saved[1].(string)
		generatedCert = saved[2].(string)
		generatedKey = saved[3].(string)
		readTLSConfig = saved[4].(func() (*config.Config, error))
		writeTLSConfig = saved[5].(func(*config.Config) error)
		ensureSelfCert = saved[6].(func() error)
		certIdentity = saved[7].(func() utils.CertIdentity)
		servingCert = saved[8].(*utils.ServingCert)
	})

	uploadedCertFile = filepath.Join(f.dir, "uploaded.crt")
	uploadedKeyFile = filepath.Join(f.dir, "uploaded.key")
	generatedCert = filepath.Join(f.dir, "server.crt")
	generatedKey = filepath.Join(f.dir, "server.key")
	readTLSConfig = func() (*config.Config, error) {
		copy := f.conf
		return &copy, nil
	}
	writeTLSConfig = func(conf *config.Config) error {
		f.conf = *conf
		return nil
	}
	ensureSelfCert = func() error {
		if _, err := os.Stat(generatedCert); err == nil {
			return nil
		}
		certPEM, keyPEM := selfSigned(t, "generated.invalid")
		return utils.WriteCertPair(generatedCert, generatedKey, certPEM, keyPEM)
	}
	certIdentity = func() utils.CertIdentity {
		return utils.CertIdentity{DNS: []string{"localhost", "kvm.test"}, IPs: []net.IP{net.ParseIP("127.0.0.1")}}
	}
	servingCert = &utils.ServingCert{}

	if err := ensureSelfCert(); err != nil {
		t.Fatal(err)
	}
	f.conf.Cert.Crt, f.conf.Cert.Key = generatedCert, generatedKey
	if err := servingCert.Load(generatedCert, generatedKey); err != nil {
		t.Fatal(err)
	}

	service := NewService()
	f.router = gin.New()
	f.router.GET("/cert", service.GetTlsCert)
	f.router.POST("/cert/check", service.CheckTlsCert)
	f.router.POST("/cert", service.SetTlsCert)
	f.router.DELETE("/cert", service.RestoreTlsCert)
	return f
}

type certReply struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (f *certFixture) call(method, path string, body any) certReply {
	f.t.Helper()
	var reader *strings.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = strings.NewReader(string(data))
	} else {
		reader = strings.NewReader("")
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		f.t.Fatalf("%s %s: status %d", method, path, recorder.Code)
	}
	var r certReply
	if err := json.Unmarshal(recorder.Body.Bytes(), &r); err != nil {
		f.t.Fatal(err)
	}
	return r
}

func servedName(t *testing.T) string {
	t.Helper()
	cert := servingCert.Current()
	if cert == nil {
		t.Fatal("nothing served")
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	return leaf.Subject.CommonName
}

func TestCertCheckDoesNotInstall(t *testing.T) {
	f := newCertFixture(t)
	certPEM, keyPEM := selfSigned(t, "kvm.test")
	r := f.call(http.MethodPost, "/cert/check", map[string]string{"cert": string(certPEM), "key": string(keyPEM)})
	if r.Code != 0 || !strings.Contains(string(r.Data), `"subject":"CN=kvm.test"`) {
		t.Fatalf("check: %d %s %s", r.Code, r.Msg, r.Data)
	}
	if _, err := os.Stat(uploadedCertFile); !os.IsNotExist(err) {
		t.Fatal("check wrote the certificate")
	}
	if f.conf.Cert.Crt != generatedCert || servedName(t) != "generated.invalid" {
		t.Fatal("check changed what is served")
	}
}

func TestCertUploadAndRestore(t *testing.T) {
	f := newCertFixture(t)
	certPEM, keyPEM := selfSigned(t, "kvm.test")

	r := f.call(http.MethodPost, "/cert", map[string]string{"cert": string(certPEM), "key": string(keyPEM)})
	if r.Code != 0 {
		t.Fatalf("upload: %d %s", r.Code, r.Msg)
	}
	if f.conf.Cert.Crt != uploadedCertFile || f.conf.Cert.Key != uploadedKeyFile {
		t.Fatalf("config not pointed at the upload: %+v", f.conf.Cert)
	}
	if servedName(t) != "kvm.test" {
		t.Fatal("the upload is not served without a restart")
	}
	info, err := os.Stat(uploadedKeyFile)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("uploaded key: %v %v", info, err)
	}
	r = f.call(http.MethodGet, "/cert", nil)
	if r.Code != 0 || !strings.Contains(string(r.Data), `"uploaded":true`) || !strings.Contains(string(r.Data), "kvm.test") {
		t.Fatalf("get after upload: %s", r.Data)
	}
	if !uploadedPairUsable(&f.conf) {
		t.Fatal("HTTPS on would drop the upload")
	}

	r = f.call(http.MethodDelete, "/cert", nil)
	if r.Code != 0 {
		t.Fatalf("restore: %d %s", r.Code, r.Msg)
	}
	if f.conf.Cert.Crt != generatedCert || servedName(t) != "generated.invalid" {
		t.Fatal("not back on the generated certificate")
	}
	for _, path := range []string{uploadedCertFile, uploadedKeyFile} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s left behind", path)
		}
	}
	r = f.call(http.MethodGet, "/cert", nil)
	if !strings.Contains(string(r.Data), `"uploaded":false`) {
		t.Fatalf("get after restore: %s", r.Data)
	}
}

func TestCertUploadRefusedLeavesEverything(t *testing.T) {
	f := newCertFixture(t)
	certPEM, _ := selfSigned(t, "kvm.test")
	_, otherKey := selfSigned(t, "other.test")

	r := f.call(http.MethodPost, "/cert", map[string]string{"cert": string(certPEM), "key": string(otherKey)})
	if r.Code != -2 || !strings.Contains(string(r.Data), `"kind":"keyMismatch"`) {
		t.Fatalf("mismatch: %d %s %s", r.Code, r.Msg, r.Data)
	}
	if strings.Contains(r.Msg, "PRIVATE") {
		t.Fatal("the reply quotes the key")
	}
	if _, err := os.Stat(uploadedCertFile); !os.IsNotExist(err) {
		t.Fatal("a refused upload was written")
	}
	if f.conf.Cert.Crt != generatedCert || servedName(t) != "generated.invalid" {
		t.Fatal("a refused upload changed what is served")
	}
	if r := f.call(http.MethodPost, "/cert", map[string]string{"cert": string(certPEM)}); r.Code != -1 {
		t.Fatalf("missing key: %d", r.Code)
	}
}

// Switching HTTPS on keeps an uploaded pair only while it loads.
func TestUploadedPairUsable(t *testing.T) {
	f := newCertFixture(t)
	if uploadedPairUsable(&f.conf) {
		t.Fatal("the generated pair counted as uploaded")
	}
	certPEM, keyPEM := selfSigned(t, "kvm.test")
	if err := utils.WriteCertPair(uploadedCertFile, uploadedKeyFile, certPEM, keyPEM); err != nil {
		t.Fatal(err)
	}
	conf := f.conf
	conf.Cert.Crt, conf.Cert.Key = uploadedCertFile, uploadedKeyFile
	if !uploadedPairUsable(&conf) {
		t.Fatal("a good upload not usable")
	}
	if err := os.Remove(uploadedKeyFile); err != nil {
		t.Fatal(err)
	}
	if uploadedPairUsable(&conf) {
		t.Fatal("an upload without its key counted as usable")
	}
}
