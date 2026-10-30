package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// SameOrigin reports whether the request may act on the device.
//
// Browsers always attach an Origin header to cross-site POST requests and to
// websocket handshakes, so comparing it against the host the request was sent
// to is what stops a hostile page from riding the session cookie (CSRF and
// cross-site websocket hijacking). Requests without an Origin header come from
// non-browser clients, which cannot be driven by a web page, and are allowed.
// Only the hostname is compared: the device serves the same UI over both http
// and https, so an origin arriving on the other port is still the same device.
func SameOrigin(r *http.Request) bool {
	if r == nil {
		return false
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}

	return strings.EqualFold(u.Hostname(), hostname(r.Host))
}

// allowByOrigin applies SameOrigin unless authentication is switched off, a
// mode in which the API is deliberately open (main.go also enables permissive
// CORS there) and origin enforcement would contradict the operator's choice.
func allowByOrigin(r *http.Request) bool {
	if config.GetInstance().Authentication == "disable" {
		return true
	}

	return SameOrigin(r)
}

func hostname(hostport string) string {
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return strings.Trim(host, "[]")
	}

	return strings.Trim(hostport, "[]")
}

// ForbiddenOrigin is the data of a 403 that refused a request's origin. Both
// names are ones the requester already holds: the Origin it sent and the Host
// the device saw. A page behind a reverse proxy that rewrites Host reads them
// to tell the operator which names disagree (ironkvm-dist#178).
type ForbiddenOrigin struct {
	Origin string `json:"origin"`
	Host   string `json:"host"`
}

// originHost is the host part of an Origin header, or the header as sent when
// it does not parse.
func originHost(origin string) string {
	if u, err := url.Parse(origin); err == nil && u.Host != "" {
		return u.Host
	}
	return origin
}

const originLogInterval = time.Minute

var (
	originLogMu   sync.Mutex
	originLogLast time.Time
	// originNow and originWarn are vars so a test can drive the clock and
	// watch the log.
	originNow  = time.Now
	originWarn = func(origin, host string) {
		log.Warnf("rejected request from origin %s to host %s: a reverse proxy in front of the device must pass the original Host header", origin, host)
	}
)

// logForbiddenOrigin warns about a refused origin at most once a minute. A page
// whose socket keeps retrying would otherwise fill the log.
func logForbiddenOrigin(origin, host string) {
	originLogMu.Lock()
	now := originNow()
	if !originLogLast.IsZero() && now.Sub(originLogLast) < originLogInterval {
		originLogMu.Unlock()
		return
	}
	originLogLast = now
	originLogMu.Unlock()

	originWarn(origin, host)
}

// OriginCheck answers a page that asks whether the device accepts its origin.
// The route sits behind CheckToken, which refuses a foreign origin with a 403
// that names both hosts, so reaching here means origin and session both pass.
func OriginCheck(c *gin.Context) {
	(&proto.Response{}).OkRsp(c)
}
