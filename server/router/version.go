package router

import (
	"sync"

	"NanoKVM-Server/common/apicontract"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/application"

	"github.com/gin-gonic/gin"
)

// VersionRsp is the contract a web UI reads before it calls anything else.
//
// It is answered without authentication, because the login page is a UI too
// and has to know which server it talks to. It therefore carries nothing a
// visitor could not already learn from the login page and the static files: the
// release number, the API version and the names of the features. It carries no
// build stamp (which names a commit), no host name, no configuration and
// nothing about the session.
type VersionRsp struct {
	// Version is the installed application release, as /kvmapp/version says.
	Version string `json:"version"`
	// API is apicontract.APIVersion.
	API int `json:"api"`
	// Features is the registry in apicontract, sorted.
	Features []string `json:"features"`
}

// installedVersion is a variable so a test need not read /kvmapp/version.
var installedVersion = application.InstalledVersion

var (
	versionOnce sync.Once
	versionRsp  VersionRsp
)

// versionRouter registers GET /api/version.
func versionRouter(r *gin.Engine) {
	r.GET("/api/version", getVersion)
}

// getVersion answers the contract. Every field is fixed for the life of the
// process (an application update restarts the server), so it is read once.
func getVersion(c *gin.Context) {
	versionOnce.Do(func() {
		versionRsp = VersionRsp{
			Version:  installedVersion(),
			API:      apicontract.APIVersion,
			Features: apicontract.Features(),
		}
	})

	var rsp proto.Response
	rsp.OkRspWithData(c, versionRsp)
}
