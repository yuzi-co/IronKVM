package audit

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
)

// std is the log of this process.
var std = NewStore(Dir)

// Default returns the log of this process.
func Default() *Store { return std }

// Record adds an event to the log.
func Record(e Event) { std.Record(e) }

// Init connects the syslog forward. forward sends one line to syslog; the
// setting decides, event by event, whether it is called.
func Init(forward func(line string)) {
	if forward == nil {
		return
	}
	std.SetSink(func(e Event) {
		if CurrentSettings().Syslog {
			forward(Syslog(e))
		}
	})
}

// Identity is who a request came from.
type Identity struct {
	Actor string
	Via   string
	// Key names the API key, when one authenticated the request.
	Key string
}

// identify reads the identity the authentication middleware left on the
// request. The router sets it: this package must not import middleware, which
// the services it is called from already import.
var identify func(*gin.Context) (Identity, bool)

// SetIdentity sets how a request's identity is read.
func SetIdentity(f func(*gin.Context) (Identity, bool)) { identify = f }

// ClientIP is the address a request came from. It is the peer address, never
// X-Forwarded-For, which the client writes: a log that believed it would let
// anyone put another address in it.
func ClientIP(c *gin.Context) string {
	if ip := c.RemoteIP(); ip != "" {
		return ip
	}
	return c.ClientIP()
}

const noteKey = "audit.note"

// note is what a handler adds to the event the middleware records for it.
type note struct {
	actor, target, detail, outcome, action string
}

func noteOf(c *gin.Context) *note {
	if v, ok := c.Get(noteKey); ok {
		if n, ok := v.(*note); ok {
			return n
		}
	}
	n := &note{}
	c.Set(noteKey, n)
	return n
}

// SetActor names the actor of a request that has no authenticated identity,
// such as a login attempt.
func SetActor(c *gin.Context, actor string) { noteOf(c).actor = actor }

// SetTarget names what the request acted on: an image, a macro, a button.
func SetTarget(c *gin.Context, target string) { noteOf(c).target = target }

// SetDetail adds a short remark. It is never the place for a secret or for
// text the user typed.
func SetDetail(c *gin.Context, detail string) { noteOf(c).detail = detail }

// SetOutcome overrides the outcome the middleware worked out from the
// response.
func SetOutcome(c *gin.Context, outcome string) { noteOf(c).outcome = outcome }

// SetAction overrides the action the route's rule names.
func SetAction(c *gin.Context, action string) { noteOf(c).action = action }

// Request records an event for the request now, with its identity and
// address. The middleware does this for ordinary routes; a handler that
// outlives its request, such as a websocket, calls it itself.
func Request(c *gin.Context, e Event) {
	if e.Actor == "" {
		if id, ok := identityOf(c); ok {
			e.Actor, e.Via, e.Key = id.Actor, id.Via, id.Key
		}
	}
	if e.IP == "" {
		e.IP = ClientIP(c)
	}
	Record(e)
}

func identityOf(c *gin.Context) (Identity, bool) {
	if identify == nil {
		return Identity{}, false
	}
	return identify(c)
}

// Middleware records the changes a request made. It runs after the handler,
// so it knows the outcome. A route is audited when Classify has a rule for
// it; reads are not, and neither are the routes Classify skips.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			return
		}
		rule, ok := Classify(c.Request.Method, c.FullPath())
		if !ok {
			return
		}

		n := &note{}
		if v, ok := c.Get(noteKey); ok {
			if stored, ok := v.(*note); ok {
				n = stored
			}
		}

		id, known := identityOf(c)
		if n.actor != "" {
			id = Identity{Actor: n.actor, Via: id.Via}
			known = true
		}
		// A request nobody authenticated is a probe, not an action. The
		// routes that take the credentials themselves are the exception.
		if !known && !rule.Anonymous {
			return
		}

		e := Event{
			Actor:    id.Actor,
			Via:      id.Via,
			Key:      id.Key,
			IP:       ClientIP(c),
			Category: rule.Category,
			Action:   rule.Action,
			Target:   n.target,
			Outcome:  OutcomeOK,
			Detail:   n.detail,
		}
		if n.action != "" {
			e.Action = n.action
		}
		if len(c.Params) > 0 && e.Target == "" {
			e.Target = c.Params[len(c.Params)-1].Value
		}

		status := c.Writer.Status()
		switch {
		case status == http.StatusUnauthorized || status == http.StatusForbidden:
			e.Outcome = OutcomeDenied
		case status >= 400:
			e.Outcome = OutcomeFail
		default:
			if code, msg := ResponseResult(c); code != 0 {
				e.Outcome = OutcomeFail
				if e.Detail == "" {
					e.Detail = msg
				}
			}
		}
		if n.outcome != "" {
			e.Outcome = n.outcome
		}
		Record(e)
	}
}

// ResponseResult returns the error code and message a handler answered with,
// or 0 and "" when it answered with success.
func ResponseResult(c *gin.Context) (int, string) {
	code, _ := c.Get(proto.ResponseCodeKey)
	msg, _ := c.Get(proto.ResponseMsgKey)
	n, _ := code.(int)
	s, _ := msg.(string)
	return n, strings.TrimSpace(s)
}

// ClientHeader carries the peer address from the gin layer to an MCP tool
// handler, which sees only the headers of the request. The gin middleware
// overwrites whatever the client sent.
const ClientHeader = "X-Nanokvm-Audit-Client"

type clientKey struct{}

// WithClient returns a context that carries the client address.
func WithClient(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientKey{}, ip)
}

// ClientFrom returns the address WithClient stored.
func ClientFrom(ctx context.Context) string {
	ip, _ := ctx.Value(clientKey{}).(string)
	return ip
}
