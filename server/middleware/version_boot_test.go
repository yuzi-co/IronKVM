package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// withCmdline points VersionBoot at a command line of the test's own.
func withCmdline(t *testing.T, line string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cmdline")
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	old := cmdlinePath
	cmdlinePath = path
	t.Cleanup(func() { cmdlinePath = old })
}

func TestVersionBoot(t *testing.T) {
	cases := []struct {
		name, line, want string
	}{
		{"version", "root=/dev/mmcblk0p2 rootwait ro ironkvm.version=1.2.1-mainline psi=1\n", "1.2.1-mainline"},
		{"trial", "ironkvm.version=1.2.2-mainline ironkvm.trial=1 console=ttyGS0\n", "1.2.2-mainline"},
		{"slot", "root=/dev/mmcblk0p3 ironkvm.slot=B console=ttyGS0\n", ""},
		{"stock", "root=/dev/mmcblk0p2 rootwait rw console=ttyS0,115200\n", ""},
		{"prefix only", "xironkvm.version=1 ironkvm.versions=2\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withCmdline(t, tc.line)
			if got := VersionBoot(); got != tc.want {
				t.Fatalf("VersionBoot() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVersionBootUnreadable(t *testing.T) {
	old := cmdlinePath
	cmdlinePath = filepath.Join(t.TempDir(), "missing")
	t.Cleanup(func() { cmdlinePath = old })
	if got := VersionBoot(); got != "" {
		t.Fatalf("VersionBoot() = %q with no command line, want \"\"", got)
	}
}

func TestRefuseOnVersionBoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := func() (*gin.Engine, *bool) {
		ran := false
		r := gin.New()
		r.POST("/update", RefuseOnVersionBoot(), func(c *gin.Context) {
			ran = true
			c.JSON(http.StatusOK, gin.H{"code": 0})
		})
		return r, &ran
	}

	t.Run("version boot is refused", func(t *testing.T) {
		withCmdline(t, "ironkvm.version=1.2.1-mainline ro\n")
		r, ran := engine()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/update", nil))

		var rsp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
			t.Fatalf("body %q: %v", w.Body.String(), err)
		}
		if *ran {
			t.Fatal("the handler ran on a version boot")
		}
		if rsp.Code != -1 || !strings.Contains(rsp.Msg, "1.2.1-mainline") || !strings.Contains(rsp.Msg, "new version") {
			t.Fatalf("got code %d msg %q, want -1 naming the version and a new version", rsp.Code, rsp.Msg)
		}
	})

	t.Run("slot boot passes", func(t *testing.T) {
		withCmdline(t, "root=/dev/mmcblk0p3 ironkvm.slot=B\n")
		r, ran := engine()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/update", nil))
		if !*ran || w.Code != http.StatusOK {
			t.Fatalf("handler ran %v, status %d, want the handler to run", *ran, w.Code)
		}
	})
}
