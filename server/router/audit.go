package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/vm"
)

// requestIdentity tells the audit log who a request came from, from what
// the authentication middleware left on it.
func requestIdentity(c *gin.Context) (audit.Identity, bool) {
	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return audit.Identity{}, false
	}
	return audit.Identity{Actor: principal.Username, Via: principal.Via, Key: principal.Key}, true
}

// auditRouter serves the audit log to administrators only: it names the other
// users and their addresses.
func auditRouter(r *gin.Engine) {
	service := audit.NewService()
	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)

	admin.GET("/audit", service.List)                 // a page of events, newest first, filtered
	admin.GET("/audit/export", service.Export)        // the matching events as JSON lines
	admin.GET("/audit/settings", service.GetSettings) // the syslog forward and the log's limits
	admin.POST("/audit/settings", service.SetSettings)
}

// auditedPress wraps vm.PressButton for the callers that press a button on
// their own account, rather than through a request that the middleware
// already records: IPMI, Redfish, the watchdog and macros. The press itself is
// the event, so it is recorded here, once, with the outcome.
func auditedPress(via string) func(kind string, d time.Duration) error {
	return func(kind string, d time.Duration) error {
		err := vm.PressButton(kind, d)
		e := audit.Event{
			Actor:    via,
			Via:      via,
			Category: audit.CategoryPower,
			Action:   "press",
			Target:   kind,
			Outcome:  audit.OutcomeOK,
			Detail:   d.String(),
		}
		if err != nil {
			e.Outcome = audit.OutcomeFail
			e.Detail = err.Error()
		}
		audit.Record(e)
		return err
	}
}
