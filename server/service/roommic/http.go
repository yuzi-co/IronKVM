package roommic

import (
	"net/http"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// SetRequest changes the administrator's settings. A field left out keeps its
// value.
type SetRequest struct {
	Allowed *bool `json:"allowed"`
	Gain    *int  `json:"gain"`
}

// GetStatus answers every signed-in viewer: the page needs it to decide what
// to show, and the live flag is what the indicator shows to everyone.
func GetStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, Shared.Status())
}

// SetSettings is for administrators. The router puts it behind RequireRole
// already; the check here keeps that true if it is ever mounted elsewhere.
func SetSettings(c *gin.Context) {
	var rsp proto.Response

	principal, ok := middleware.CurrentPrincipal(c)
	if !ok || principal.Role != authn.RoleAdmin {
		c.JSON(http.StatusForbidden, "forbidden")
		c.Abort()
		return
	}

	var req SetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	settings := Shared.Settings()
	if req.Allowed != nil {
		settings.Allowed = *req.Allowed
	}
	if req.Gain != nil {
		settings.Gain = *req.Gain
	}
	if err := settings.validate(); err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}

	// Allowing it on a kernel without the card would only show viewers a
	// switch that cannot work. Turning it off is always accepted.
	if settings.Allowed && !Shared.available() {
		rsp.ErrRsp(c, -3, ErrUnavailable.Error())
		return
	}

	by := principal.Username
	if by == "" {
		by = "an administrator"
	}

	if err := Shared.SetSettings(settings, by); err != nil {
		rsp.ErrRsp(c, -2, "operation failed")
		return
	}

	rsp.OkRspWithData(c, Shared.Status())
}

// UserOf names the viewer behind a request for the log, or says it could not.
func UserOf(c *gin.Context) string {
	if principal, ok := middleware.CurrentPrincipal(c); ok && principal.Username != "" {
		return principal.Username
	}

	return "unknown user"
}
