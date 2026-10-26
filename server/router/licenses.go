package router

import (
	"errors"
	"io"
	"net/http"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/licenses"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// licensesRouter registers the read-only view of the image's license
// manifest (service/licenses).
//
// Any signed-in principal may read it, a viewer included: it is the same kind
// of information as the versions on the About page (GET /api/vm/info), and the
// licenses of the software on the device are owed to whoever uses it.
func licensesRouter(r *gin.Engine) {
	registerLicenses(r, licenses.DefaultDir)
}

// registerLicenses serves the manifest in dir, so a test can point it at a
// directory of its own.
func registerLicenses(r *gin.Engine, dir string) {
	manifest := licenses.New(dir)
	api := r.Group("/api").Use(middleware.CheckToken())

	api.GET("/licenses", listLicenses(manifest))          // list the manifest's files
	api.GET("/licenses/file", serveLicenseFile(manifest)) // one file, as plain text
}

type licensesRsp struct {
	Files []licenses.File `json:"files"`
}

// listLicenses answers the envelope, except that a missing directory is a
// 404, so the UI can tell an image without a manifest (or a server without
// this route) from a failure.
func listLicenses(manifest *licenses.Manifest) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rsp proto.Response

		files, err := manifest.List()
		if errors.Is(err, licenses.ErrNotInstalled) {
			rsp.Err(-1, err.Error())
			c.JSON(http.StatusNotFound, rsp)
			return
		}
		if err != nil {
			log.Errorf("list licenses: %s", err)
			rsp.ErrRsp(c, -2, "failed to list the license files")
			return
		}

		rsp.OkRspWithData(c, licensesRsp{Files: files})
	}
}

// serveLicenseFile answers the file named by the path query parameter as
// plain text, so a browser shows it in a tab rather than rendering it.
func serveLicenseFile(manifest *licenses.Manifest) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, size, err := manifest.Open(c.Query("path"))
		if err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, licenses.ErrBadPath):
				status = http.StatusBadRequest
			case errors.Is(err, licenses.ErrNotInstalled), errors.Is(err, licenses.ErrNotFound):
				status = http.StatusNotFound
			case errors.Is(err, licenses.ErrTooLarge):
			default:
				log.Errorf("open license file: %s", err)
				err = errors.New("failed to read the license file")
			}
			c.String(status, err.Error())
			return
		}
		defer file.Close()

		c.DataFromReader(http.StatusOK, size, "text/plain; charset=utf-8", io.LimitReader(file, size), map[string]string{
			"X-Content-Type-Options":  "nosniff",
			"Content-Security-Policy": "default-src 'none'",
			"Cache-Control":           "no-cache",
		})
	}
}
