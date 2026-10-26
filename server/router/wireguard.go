package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/wireguard"
)

// wireguardRouter is the built-in WireGuard tunnel, for administrators only:
// it changes the board's routes and holds a private key.
func wireguardRouter(r *gin.Engine) {
	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	wireguard.NewService().Register(admin)
}
