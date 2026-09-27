package vpn

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"

	"github.com/gin-gonic/gin"
)

// Refuse answers the request with the reason when the other VPN runs or
// starts at boot, and reports whether it did. Install, start, up, login and
// boot on both add-ons call it before they do anything.
func Refuse(c *gin.Context, d addon.Daemon) bool {
	err := addon.CheckExclusive(d.Name)
	if err == nil {
		return false
	}
	var rsp proto.Response
	rsp.ErrRsp(c, -1, err.Error())
	return true
}

// Boot handles POST boot for either add-on: {enabled} turns start at boot on
// or off. Turning it on is refused while the other VPN runs or starts at boot;
// turning it off never is.
func Boot(c *gin.Context, d addon.Daemon, installed bool) {
	var req proto.VpnBootReq
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if req.Enabled {
		if !installed {
			rsp.ErrRsp(c, -1, d.Title+" is not installed")
			return
		}
		if Refuse(c, d) {
			return
		}
	}
	if err := addon.SetBoot(d, req.Enabled); err != nil {
		rsp.ErrRsp(c, -1, Message("start at boot failed", err))
		return
	}
	rsp.OkRsp(c)
}
