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
	return audit.Identity{Actor: principal.Username, Via: principal.Via, Key: principal.Key, Role: string(principal.Role)}, true
}

// auditRouter serves the audit log and the capability tiers of MCP and the
// agent, both to administrators only: the log names the other users and their
// addresses, and the tiers decide what an AI may do.
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
	admin.GET("/agent/tiers", service.GetTiers) // what MCP and the agent may do
	admin.POST("/agent/tiers", service.SetTiers)
}

// auditedPress wraps vm.PressButton for the callers that press a button on
// their own account, rather than through a request that the middleware
// already records: IPMI, Redfish, the watchdog and macros. The press itself is
// the event, so it is recorded here, once, with the outcome.
func auditedPress(via string) func(kind string, d time.Duration) error {
	return auditedPressWith(via, vm.PressButton)
}

// auditedPressWith is auditedPress around another press function, such as
// vm.PressButtonQuiet for the watchdog, which reports its presses to the event
// bus itself. The bus carries no actor or outcome, so the audit log keeps its
// own record rather than subscribing to it.
func auditedPressWith(via string, press func(kind string, d time.Duration) error) func(kind string, d time.Duration) error {
	return func(kind string, d time.Duration) error {
		err := press(kind, d)
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
