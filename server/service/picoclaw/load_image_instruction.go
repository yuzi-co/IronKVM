package picoclaw

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"NanoKVM-Server/service/agent"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	defaultImagePrompt = "Describe the image briefly."
)

func (s *Service) LoadImage(c *gin.Context) {
	var req LoadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writePicoclawError(c, newPicoclawError(CodeInvalidAction, "invalid load image payload"))
		return
	}

	sessionID, sessionErr := s.requireActiveGatewaySession(c)
	if sessionErr != nil {
		writePicoclawError(c, sessionErr)
		return
	}

	sourcePath, pathErr := normalizeLoadImagePath(req.Path)
	if pathErr != nil {
		writePicoclawError(c, pathErr)
		return
	}

	instruction := buildLoadImageInstruction(sourcePath, req.Prompt)
	messageID := uuid.NewString()
	if err := s.chat.Prompt(sessionID, agent.Prompt{RequestID: messageID, Text: instruction}); err != nil {
		writePicoclawError(c, newPicoclawError(CodeRuntimeUnavailable, "failed to deliver image instruction to picoclaw"))
		return
	}

	writeSuccess(c, gin.H{
		"accepted":      true,
		"session_id":    sessionID,
		"message_id":    messageID,
		"path":          sourcePath,
		"user_prompt":   normalizeImagePrompt(req.Prompt),
		"instructed_at": time.Now(),
	})
}

// requireActiveGatewaySession returns the chat session a loopback call is
// about: the one its header names, or the lock owner's. The session must be
// open.
func (s *Service) requireActiveGatewaySession(c *gin.Context) (string, *PicoclawError) {
	s.ensureDependencies()
	sessionID := strings.TrimSpace(c.GetHeader(sessionIDHeader))
	if sessionID == "" {
		sessionID = strings.TrimSpace(s.lock.Owner())
	}
	if sessionID == "" {
		return "", newPicoclawError(CodeSessionIDMissing, "missing X-PicoClaw-Session-ID")
	}
	if !s.chat.IsActive(sessionID) {
		err := newPicoclawError(CodeRuntimeUnavailable, "picoclaw session is not active")
		err.SessionID = sessionID
		return "", err
	}
	return sessionID, nil
}

func normalizeLoadImagePath(sourcePath string) (string, *PicoclawError) {
	sourcePath = strings.TrimSpace(sourcePath)
	if sourcePath == "" {
		return "", newPicoclawError(CodeInvalidAction, "image path is required")
	}

	return sourcePath, nil
}

func buildLoadImageInstruction(sourcePath string, prompt string) string {
	argsJSON, _ := json.Marshal(map[string]string{"path": sourcePath})
	return fmt.Sprintf(
		"Call `load_image` first with this exact argument object:\n```json\n%s\n```\n\nDo not ask the user to re-upload the image. Do not modify the path. Do not use `read_file` or any other tool to inspect this image directly.\n\nAfter `load_image` succeeds, treat the loaded image as the current task input and continue with this request:\n%s",
		string(argsJSON),
		normalizeImagePrompt(prompt),
	)
}

func normalizeImagePrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return defaultImagePrompt
	}
	return prompt
}
