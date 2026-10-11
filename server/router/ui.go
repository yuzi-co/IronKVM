package router

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/webui"

	"github.com/gin-gonic/gin"
)

// uiStore decides which web UI is served: the one the application package
// carries, or one installed apart from it. web() creates it before server()
// registers the routes that change it.
var uiStore *webui.Store

func uiRouter(r *gin.Engine) {
	if uiStore == nil {
		return
	}
	service := webui.NewService(uiStore)
	api := r.Group("/api").Use(middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))

	api.GET("/ui", service.GetStatus)          // served UI, previous UI, signature check
	api.GET("/ui/latest", service.GetLatest)   // UI package in the update feed
	api.POST("/ui/update", service.Update)     // install the feed's UI package
	api.POST("/ui/upload", service.Upload)     // install an uploaded UI package
	api.POST("/ui/rollback", service.Rollback) // back to the previous UI
}

// staticRoot is one web directory and the handler that serves it.
type staticRoot struct {
	dir   string
	serve gin.HandlerFunc
}

// switchingStaticHandler serves whichever directory dir names at the time of
// the request. The handler for a directory is built once, when the directory
// first differs from the last one, so a switch costs one allocation and a
// request costs one atomic load.
func switchingStaticHandler(dir func() string) gin.HandlerFunc {
	var current atomic.Pointer[staticRoot]
	return func(c *gin.Context) {
		d := dir()
		root := current.Load()
		if root == nil || root.dir != d {
			root = &staticRoot{dir: d, serve: staticHandler(d)}
			current.Store(root)
		}
		root.serve(c)
	}
}

// probeWebDir serves the entry document from webDir through the same handler
// the server uses, and checks that what comes back is index.html. A UI that
// passes its file checks and still cannot be served is not switched to.
func probeWebDir(webDir string) error {
	want, err := os.ReadFile(filepath.Join(webDir, indexFileName))
	if err != nil {
		return err
	}
	engine := gin.New()
	engine.Use(staticHandler(webDir))
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		return fmt.Errorf("GET / answered %d", w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), want) {
		return errors.New("GET / did not answer index.html")
	}
	return nil
}
