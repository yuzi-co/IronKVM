package middleware

import (
	"os"
	"strings"

	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// cmdlinePath is the kernel command line. Tests point it at a file of their own.
var cmdlinePath = "/proc/cmdline"

// VersionBoot returns the version the board booted from the version store
// (ironkvm-dist docs/specs/2026-10-29-version-store-design.md), or "" on a
// slot boot or stock. U-Boot names it on the command line as
// ironkvm.version=<id>.
func VersionBoot() string {
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return ""
	}
	for _, word := range strings.Fields(string(data)) {
		if id, ok := strings.CutPrefix(word, "ironkvm.version="); ok {
			return id
		}
	}
	return ""
}

// RefuseOnVersionBoot refuses a route that installs into /kvmapp when the
// board booted a version.
//
// A version's root is a read-only squashfs under an overlay whose upper layer
// is tmpfs. An application or web UI package unpacked there lands in RAM,
// where a 60MB application package and its cache can take most of what is
// free, and it is gone at the next boot, leaving the board on the version's
// own files with nothing said. Updates on a version boot come as a new
// version, so these routes answer with the reason instead.
func RefuseOnVersionBoot() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := VersionBoot()
		if id == "" {
			c.Next()
			return
		}

		var rsp proto.Response
		rsp.ErrRsp(c, -1, "this board runs version "+id+" from the version store, whose root is read-only: "+
			"an update installed here would last only until the next reboot. Install a new version instead")
		c.Abort()
	}
}
