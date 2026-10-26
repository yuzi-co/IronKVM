package router

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/common"
	"NanoKVM-Server/common/version"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/application"
	"NanoKVM-Server/service/diagnostics"
	"NanoKVM-Server/service/failsafe"
	"NanoKVM-Server/service/ion"
	"NanoKVM-Server/service/settingsio"
	"NanoKVM-Server/utils"
)

// supportRouter mounts the two support features: the settings export and
// import, and the diagnostics bundle. Both need an administrator; see
// docs/superpowers/specs/2026-10-26-settings-export-diagnostics-design.md.
func supportRouter(r *gin.Engine) {
	settings := settingsio.NewService(boardSource)

	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings/export", settings.Export)
	admin.POST("/settings/preview", settings.Preview)
	admin.POST("/settings/import", settings.Import)

	collector := diagnostics.NewCollector()
	collector.Versions = func() map[string]string {
		return map[string]string{
			"application": application.InstalledVersion(),
			"server":      version.Decorate(application.InstalledVersion()),
			"build":       version.Build,
			"go":          runtime.Version(),
		}
	}
	collector.Settings = func() (any, error) {
		doc, _, err := settings.Device.Export(boardSource(), settings.Now())
		if err != nil {
			return nil, err
		}
		return doc.Sections, nil
	}
	collector.Video = diagnosticsVideo
	admin.GET("/diagnostics/bundle", diagnosticsBundle(collector))
}

// boardSource names the board in an exported file.
func boardSource() settingsio.Source {
	host, _ := os.Hostname()
	return settingsio.Source{Application: application.InstalledVersion(), Hostname: host}
}

// diagnosticsVideo is the video state for the bundle. It reads only what the
// server already holds and the ION dump; it does not ask libkvm for the HDMI
// signal, because the first call into libkvm builds the capture pipeline.
func diagnosticsVideo() any {
	state := map[string]any{
		"videoDisabled":          common.VideoDisabled(),
		"hdmiCaptureDisabled":    utils.IsHdmiDisabled(),
		"hdmiIdleTimeoutMinutes": utils.GetHDMIIdleTimeout(),
		"failsafe":               failsafe.Current(),
	}
	if ion.Present() {
		state["ion"] = ion.Read()
	} else {
		state["ion"] = "absent: no ION carveout on this kernel"
	}
	return state
}

// diagnosticsBundle answers GET /api/diagnostics/bundle. The bundle is built
// in memory first, so a failure is an error answer and never half an archive.
// It is small: the logs are cut at 256 KiB each before compression.
func diagnosticsBundle(collector *diagnostics.Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rsp proto.Response
		var buf bytes.Buffer

		if err := collector.Write(c.Request.Context(), &buf); err != nil {
			if errors.Is(err, diagnostics.ErrBusy) {
				rsp.ErrRsp(c, -2, err.Error())
				return
			}
			log.Errorf("diagnostics bundle: %s", err)
			rsp.ErrRsp(c, -1, "building the bundle failed")
			return
		}

		host, _ := os.Hostname()
		name := settingsio.FileName("ironkvm-diagnostics", host, collector.Now(), "tar.gz")
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
		c.Data(http.StatusOK, "application/gzip", buf.Bytes())
	}
}
