package metrics

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ContentType is the text exposition format, version 0.0.4, which every
// Prometheus release reads.
const ContentType = "text/plain; version=0.0.4"

// Handler serves GET /api/metrics. The router puts the token check in front of
// it; any role may read it, because it holds no secret.
func Handler(c *gin.Context) {
	var body bytes.Buffer
	if err := Collect(&body); err != nil {
		c.String(http.StatusInternalServerError, "metrics: %s", err)
		return
	}

	c.Data(http.StatusOK, ContentType, body.Bytes())
}
