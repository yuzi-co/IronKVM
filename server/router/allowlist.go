package router

import (
	"net/netip"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/allowlist"
	"NanoKVM-Server/service/vm"
)

func allowlistRouter(r *gin.Engine) {
	// The host on the USB network link is always let in: it is the machine
	// the board is plugged into, and the way back from a list that shuts
	// out the owner's network.
	allowlist.Exempt = usbLinkHost

	service := allowlist.NewService()
	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/network/allowlist", service.GetSettings)  // get the allowed client networks
	admin.POST("/network/allowlist", service.SetSettings) // set them, refusing a list without the caller
}

// usbLinkHost reports whether addr is on the USB network link while the link
// is up.
func usbLinkHost(addr netip.Addr) bool {
	link := vm.CurrentUSBLink()
	return link.Active && link.Prefix.IsValid() && link.Prefix.Contains(addr)
}
