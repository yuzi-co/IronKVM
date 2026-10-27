package picoclaw

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (s *Service) ReleaseRuntimeSession(c *gin.Context) {
	s.ensureDependencies()
	sessionID := namedSession(c.Request.Header)
	if sessionID == "" {
		writePicoclawError(c, newPicoclawError(CodeSessionIDMissing, "missing X-PicoClaw-Session-ID"))
		return
	}

	s.chat.CloseSession(sessionID, websocket.CloseNormalClosure, "session released")

	status := s.runtime.Get()
	status.CurrentSession = s.lock.Owner()
	s.runtime.Set(status)

	writeSuccess(c, gin.H{
		"released":        true,
		"current_session": status.CurrentSession,
	})
}
