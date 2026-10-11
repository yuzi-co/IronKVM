package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DefaultBodyLimit is the most a request may send unless its route says
// otherwise. Every JSON route fits in a fraction of it: the largest is a paste
// of 20,000 characters, which escaped as \uXXXX is about 120KB.
//
// The board has no swap and no working OOM killer. Below about 30MB free, fork
// fails and only a power cycle clears it, so a body nothing bounds is not a
// spike and a restart here, it is a visit to the board.
const DefaultBodyLimit = 1 << 20

// NoBodyLimit leaves a route's body to its handler, which must then bound it
// itself. It is for uploads whose size is checked against something better
// than a constant, such as the free space on /data.
const NoBodyLimit = -1

// LimitBody caps the request body of every route at def, or at the cap limits
// holds for it. limits is keyed by method and route pattern as gin registers
// them, for example "POST /api/ui/upload".
//
// A body over the cap is refused with 413 before the handler runs, with the
// usual {code, msg, data} envelope so the web UI shows the reason. A body
// whose length is declared is refused on the header alone. A chunked one has
// no length, so it is read into memory up to the cap, which is small enough
// to hold, and the handler reads the copy.
//
// It goes on the engine, before any route is registered, so that it covers
// every route, including the ones with no session such as the login.
func LimitBody(def int64, limits map[string]int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := def
		if n, ok := limits[c.Request.Method+" "+c.FullPath()]; ok {
			limit = n
		}

		body := c.Request.Body
		// A websocket upgrade and every GET arrive with no body at all.
		if limit == NoBodyLimit || body == nil || body == http.NoBody {
			c.Next()
			return
		}

		switch length := c.Request.ContentLength; {
		case length > limit:
			bodyTooLarge(c)
			return
		case length >= 0:
			// net/http already stops at the declared length; this keeps the
			// cap true if that ever changes.
			c.Request.Body = http.MaxBytesReader(c.Writer, body, limit)
		default:
			data, err := io.ReadAll(io.LimitReader(body, limit+1))
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if int64(len(data)) > limit {
				bodyTooLarge(c)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(data))
		}

		c.Next()
	}
}

func bodyTooLarge(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
		"code": -1,
		"msg":  "request body too large",
		"data": nil,
	})
}
