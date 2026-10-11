package webui

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func newTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/ui", svc.GetStatus)
	r.GET("/api/ui/latest", svc.GetLatest)
	r.POST("/api/ui/update", svc.Update)
	r.POST("/api/ui/upload", svc.Upload)
	r.POST("/api/ui/rollback", svc.Rollback)
	return r
}

func do(t *testing.T, r *gin.Engine, req *http.Request) envelope {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var env envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("not an envelope: %s", w.Body.String())
	}
	return env
}

func uploadRequest(t *testing.T, pkg string, withSig bool) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	add := func(field, path, name string) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		fw, err := mw.CreateFormFile(field, name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write(data)
	}
	if withSig {
		// The signature first: the order of the parts must not matter.
		add("signature", pkg+".minisig", "whatever.minisig")
	}
	add("package", pkg, filepath.Base(pkg))
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/ui/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestUploadInstallsAndRollbackRestores(t *testing.T) {
	e := newTestEnv(t)
	r := newTestRouter(NewService(e.store))

	env := do(t, r, uploadRequest(t, e.signedPackage(t, "1.1.0"), true))
	if env.Code != 0 {
		t.Fatalf("upload: %s", env.Msg)
	}
	var res InstallResult
	_ = json.Unmarshal(env.Data, &res)
	if res.Installed.Version != "1.1.0" || res.Verdict.Signer != "test.pub" {
		t.Fatalf("upload result %+v", res)
	}

	env = do(t, r, httptest.NewRequest(http.MethodGet, "/api/ui", nil))
	var st Status
	_ = json.Unmarshal(env.Data, &st)
	if st.Active.Version != "1.1.0" || !st.CanRollback || st.SignatureCheck != "on" || st.ServerAPI != 1 {
		t.Fatalf("status %+v", st)
	}

	env = do(t, r, httptest.NewRequest(http.MethodPost, "/api/ui/rollback", nil))
	if env.Code != 0 || servedIndex(t, e.store) != "bundled" {
		t.Fatalf("rollback: %s", env.Msg)
	}
	left, _ := filepath.Glob(filepath.Join(e.store.Root, uploadPrefix+"*"))
	if len(left) != 0 {
		t.Fatalf("upload directories left: %v", left)
	}
}

func TestUploadRefusals(t *testing.T) {
	e := newTestEnv(t)
	r := newTestRouter(NewService(e.store))
	pkg := e.signedPackage(t, "1.1.0")

	if env := do(t, r, uploadRequest(t, pkg, false)); env.Code == 0 || !strings.Contains(env.Msg, "no signature") {
		t.Fatalf("upload without a signature: %+v", env)
	}

	renamed := filepath.Join(e.pkgs, "evil.tar.gz")
	data, _ := os.ReadFile(pkg)
	_ = os.WriteFile(renamed, data, 0o644)
	_ = os.WriteFile(renamed+".minisig", []byte("x"), 0o644)
	if env := do(t, r, uploadRequest(t, renamed, true)); env.Code == 0 || !strings.Contains(env.Msg, "not a UI package name") {
		t.Fatalf("badly named upload: %+v", env)
	}
	if servedIndex(t, e.store) != "bundled" {
		t.Fatal("a refused upload changed the served UI")
	}
}

func fakeFeed(t *testing.T, e *testEnv, pkg string, api APIRange) *Service {
	t.Helper()
	data, _ := os.ReadFile(pkg)
	sum := sha512.Sum512(data)
	latest := &Latest{
		Format: 1, Version: strings.TrimSuffix(strings.TrimPrefix(filepath.Base(pkg), "ironkvm-ui_"), ".tar.gz"),
		Name: filepath.Base(pkg), SizeBytes: uint64(len(data)),
		Sha512: base64.StdEncoding.EncodeToString(sum[:]), API: api,
	}
	if err := latest.validate(); err != nil {
		t.Fatal(err)
	}
	svc := NewService(e.store)
	svc.fetchFeed = func() (*Latest, string, error) { return latest, "https://feed.example/" + latest.Name, nil }
	svc.download = func(url, target string, max int64) error {
		src := strings.Replace(url, "https://feed.example/", e.pkgs+"/", 1)
		f, err := os.Open(src)
		if err != nil {
			return errors.New("404")
		}
		defer f.Close()
		return writeLimited(f, target, max)
	}
	return svc
}

func TestUpdateFromTheFeed(t *testing.T) {
	e := newTestEnv(t)
	r := newTestRouter(fakeFeed(t, e, e.signedPackage(t, "1.4.0"), APIRange{1, 1}))

	env := do(t, r, httptest.NewRequest(http.MethodGet, "/api/ui/latest", nil))
	var l LatestRsp
	_ = json.Unmarshal(env.Data, &l)
	if env.Code != 0 || l.Version != "1.4.0" || !l.Compatible || l.Active {
		t.Fatalf("latest %+v %s", l, env.Msg)
	}

	if env := do(t, r, httptest.NewRequest(http.MethodPost, "/api/ui/update", nil)); env.Code != 0 {
		t.Fatalf("update: %s", env.Msg)
	}
	if !strings.Contains(servedIndex(t, e.store), "index-1.4.0.js") {
		t.Fatal("the feed's UI is not served")
	}
	env = do(t, r, httptest.NewRequest(http.MethodGet, "/api/ui/latest", nil))
	_ = json.Unmarshal(env.Data, &l)
	if !l.Active {
		t.Fatal("the installed feed version should read as active")
	}
}

func TestUpdateRefusesAnIncompatibleFeedEntryBeforeDownloading(t *testing.T) {
	e := newTestEnv(t)
	svc := fakeFeed(t, e, e.signedPackage(t, "2.0.0"), APIRange{2, 2})
	downloaded := false
	svc.download = func(string, string, int64) error { downloaded = true; return nil }
	r := newTestRouter(svc)

	env := do(t, r, httptest.NewRequest(http.MethodGet, "/api/ui/latest", nil))
	var l LatestRsp
	_ = json.Unmarshal(env.Data, &l)
	if l.Compatible {
		t.Fatal("an API 2 package read as compatible with an API 1 server")
	}
	if env := do(t, r, httptest.NewRequest(http.MethodPost, "/api/ui/update", nil)); env.Code == 0 {
		t.Fatal("an incompatible feed entry was installed")
	}
	if downloaded {
		t.Fatal("an incompatible package was downloaded")
	}
}

func TestUpdateRefusesAPackageThatDiffersFromTheFeed(t *testing.T) {
	e := newTestEnv(t)
	pkg := e.signedPackage(t, "1.4.0")
	svc := fakeFeed(t, e, pkg, APIRange{1, 1})
	// The file on the "server" changes after the feed was written.
	f, _ := os.OpenFile(pkg, os.O_APPEND|os.O_WRONLY, 0)
	_, _ = io.WriteString(f, "x")
	_ = f.Close()
	if env := do(t, newTestRouter(svc), httptest.NewRequest(http.MethodPost, "/api/ui/update", nil)); env.Code == 0 {
		t.Fatal("a package that differs from the feed was installed")
	}
}

func TestFeedValidation(t *testing.T) {
	good := Latest{Format: 1, Version: "1.0.0", Name: "ironkvm-ui_1.0.0.tar.gz", SizeBytes: 1,
		Sha512: base64.StdEncoding.EncodeToString(make([]byte, 64)), API: APIRange{1, 1}}
	if err := good.validate(); err != nil {
		t.Fatal(err)
	}
	bad := []func(l *Latest){
		func(l *Latest) { l.Format = 2 },
		func(l *Latest) { l.Name = "../ironkvm-ui_1.0.0.tar.gz" },
		func(l *Latest) { l.Name = "nanokvm_1.0.0.tar.gz" },
		func(l *Latest) { l.SizeBytes = 0 },
		func(l *Latest) { l.Sha512 = "abc" },
		func(l *Latest) { l.API = APIRange{2, 1} },
	}
	for i, mutate := range bad {
		l := good
		mutate(&l)
		if l.validate() == nil {
			t.Errorf("case %d: a bad feed entry validated: %+v", i, l)
		}
	}
}
