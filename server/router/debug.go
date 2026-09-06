package router

import (
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
)

// maxProfileSeconds caps the two profiles that run for a caller-chosen
// duration. The board has one usable core and capture already takes most of
// it, so a request for an hour of CPU profiling would hold a sampling timer
// and a growing buffer for that hour, and nothing in the profiler would stop
// it. Thirty seconds is the pprof default and long enough to see a steady
// state.
const maxProfileSeconds = 30

// debugRouter mounts the standard pprof handlers.
//
// They are always registered rather than hidden behind a configuration flag.
// Turning a flag on means editing /etc/kvm/server.yaml and restarting, the
// restart costs about two minutes, and it throws away the process state that
// the profile was meant to explain. A profiler that cannot look at the
// incident in front of you is not worth shipping.
//
// What keeps that safe is the role check, not obscurity: every route below
// needs an admin session or an admin API key, and an admin can already reboot
// the board and cut power to the host.
//
// The prefix is /api/debug/pprof/ so the static middleware returns early and
// no file on the SD card can shadow these paths. `go tool pprof` takes the
// full URL, so the prefix costs nothing:
//
//	go tool pprof 'http://<device>/api/debug/pprof/profile?seconds=20'
func debugRouter(r *gin.Engine) {
	group := r.Group("/api/debug/pprof", middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))

	group.GET("", gin.WrapF(pprof.Index))
	group.GET("/", gin.WrapF(pprof.Index))
	group.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	group.GET("/symbol", gin.WrapF(pprof.Symbol))
	group.POST("/symbol", gin.WrapF(pprof.Symbol))

	group.GET("/profile", clampSeconds(pprof.Profile))
	group.GET("/trace", clampSeconds(pprof.Trace))

	// The named profiles cost nothing until they are read. Two of them are
	// inert on top of that: block and mutex stay empty unless something has
	// called runtime.SetBlockProfileRate or runtime.SetMutexProfileFraction,
	// and this server never does, because both rates charge every goroutine
	// in the process for the whole time they are set.
	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		group.GET("/"+name, gin.WrapH(pprof.Handler(name)))
	}
}

// clampSeconds holds the `seconds` parameter down to maxProfileSeconds. The
// value is rewritten in the query rather than rejected, so an over-long
// request still returns a usable profile instead of an error the caller has
// to read.
func clampSeconds(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()

		if raw := query.Get("seconds"); raw != "" {
			seconds, err := strconv.ParseFloat(raw, 64)
			if err == nil && seconds > maxProfileSeconds {
				query.Set("seconds", strconv.Itoa(maxProfileSeconds))
				c.Request.URL.RawQuery = query.Encode()
			}
		}

		handler(c.Writer, c.Request)
	}
}
