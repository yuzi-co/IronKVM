package agent

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultSessionPageSize = 20

// ListSessionsHandler answers with one page of the agent's stored sessions.
// The query parameters offset and limit select the page.
func ListSessionsHandler(history History) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := history.ListSessions(c.Request.Context())
		if err != nil {
			writeError(c, asError(err, CodeRuntimeUnavailable))
			return
		}
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultSessionPageSize)))
		writeSuccess(c, page(items, offset, limit))
	}
}

// ReadSessionHandler answers with one stored session.
func ReadSessionHandler(history History) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := sessionIDParam(c)
		if !ok {
			return
		}
		detail, err := history.ReadSession(c.Request.Context(), id)
		if err != nil {
			writeError(c, asError(err, CodeRuntimeUnavailable))
			return
		}
		writeSuccess(c, detail)
	}
}

// DeleteSessionHandler deletes one stored session.
func DeleteSessionHandler(history History) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := sessionIDParam(c)
		if !ok {
			return
		}
		if err := history.DeleteSession(c.Request.Context(), id); err != nil {
			writeError(c, asError(err, CodeRuntimeUnavailable))
			return
		}
		writeSuccess(c, gin.H{"id": id, "deleted": true})
	}
}

func sessionIDParam(c *gin.Context) (string, bool) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		writeError(c, NewError(CodeInvalidRequest, "missing session id"))
		return "", false
	}
	return id, true
}

func page(items []SessionSummary, offset, limit int) []SessionSummary {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultSessionPageSize
	}
	if offset >= len(items) {
		return []SessionSummary{}
	}
	end := min(offset+limit, len(items))
	return items[offset:end]
}

// writeSuccess and writeError write the envelope the PicoClaw routes always
// used, so the browser reads these routes unchanged.
func writeSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

func writeError(c *gin.Context, err *Error) {
	status := err.Status
	if status == 0 {
		status = http.StatusOK
	}
	c.AbortWithStatusJSON(status, gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
}
