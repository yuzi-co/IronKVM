package audit

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeniedMessage is what a caller is told when a tier is off.
func DeniedMessage(subject string, tier Tier) string {
	who := "MCP clients"
	if subject == SubjectAgent {
		who = "the on-board agent"
	}
	return fmt.Sprintf("the %s tier is off for %s; an administrator can turn it on under Settings > MCP", tier, who)
}

// Principal is who an on-board agent's call acts as: the agent, by name, and
// the chat session it acts for, empty when none holds device control.
type Principal struct {
	Actor   string
	Session string
}

// AgentGate guards a route the on-board agent calls over loopback. The route
// needs the tier; with it off the call is refused with 403, and either way it
// is recorded with the tier, the outcome and the principal who reports.
func AgentGate(who func(*gin.Context) Principal, tool string, tier Tier) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal := Principal{Actor: "agent"}
		if who != nil {
			principal = who(c)
		}
		e := Event{
			Actor:    principal.Actor,
			Via:      "agent",
			Session:  principal.Session,
			IP:       ClientIP(c),
			Category: CategoryAgent,
			Action:   tool,
			Tier:     string(tier),
		}
		if !Allowed(SubjectAgent, tier) {
			e.Outcome = OutcomeDenied
			Record(e)
			c.JSON(http.StatusForbidden, gin.H{"error": DeniedMessage(SubjectAgent, tier)})
			c.Abort()
			return
		}

		c.Next()

		e.Outcome = OutcomeOK
		if status := c.Writer.Status(); status >= 400 {
			e.Outcome = OutcomeFail
			e.Detail = fmt.Sprintf("HTTP %d", status)
		}
		Record(e)
	}
}

// UseStore makes s the log of this process and returns a function that puts
// the previous one back. Tests use it to keep their events off the disk.
func UseStore(s *Store) (restore func()) {
	old := std
	std = s
	return func() { std = old }
}
