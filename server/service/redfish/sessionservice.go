package redfish

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func sessionService(c *gin.Context) {
	body := newResource(sessionServicePath, "SessionService.v1_1_8.SessionService")
	body["Id"] = "SessionService"
	body["Name"] = "Session Service"
	body["ServiceEnabled"] = true
	body["SessionTimeout"] = int(sessionIdleTimeout.Seconds())
	body["Sessions"] = link(sessionsPath)

	writeJSON(c, http.StatusOK, body)
}

func sessionBody(sess session) object {
	body := newResource(sessionsPath+"/"+sess.id, "Session.v1_3_0.Session")
	body["Id"] = sess.id
	body["Name"] = "User Session"
	body["UserName"] = sess.username
	body["Password"] = nil
	return body
}

func (s *Service) listSessions(c *gin.Context) {
	var members []string
	for _, sess := range s.sessions.list() {
		members = append(members, sessionsPath+"/"+sess.id)
	}

	writeJSON(c, http.StatusOK, newCollection(sessionsPath, "SessionCollection", "Session Collection", members))
}

func (s *Service) getSession(c *gin.Context) {
	sess, ok := s.sessions.get(c.Param("id"))
	if !ok {
		notFound(c)
		return
	}

	writeJSON(c, http.StatusOK, sessionBody(sess))
}

// createSession is the login. It is public, and runs the same password check
// and brute-force limit as Basic credentials do.
func (s *Service) createSession(c *gin.Context) {
	params, ok := decodeParams(c)
	if !ok {
		return
	}

	var username, password string
	for _, field := range []struct {
		name   string
		target *string
	}{{"UserName", &username}, {"Password", &password}} {
		if !present(params, field.name) {
			writeError(c, http.StatusBadRequest, "PropertyMissing", "", field.name)
			return
		}
		if err := json.Unmarshal(params[field.name], field.target); err != nil {
			writeError(c, http.StatusBadRequest, "MalformedJSON", field.name+" must be a string")
			return
		}
	}

	var tokenVersion uint64
	if !s.deps.AuthDisabled() {
		user, ok := s.checkPassword(c, username, password)
		if !ok {
			return
		}
		username = user.Username
		tokenVersion = user.TokenVersion
	}

	sess, token, err := s.sessions.create(username, tokenVersion)
	if errors.Is(err, errSessionLimit) {
		writeError(c, http.StatusServiceUnavailable, "SessionLimitExceeded", "")
		return
	}
	if err != nil {
		log.Errorf("redfish: create a session: %s", err)
		writeError(c, http.StatusInternalServerError, "GeneralError", "could not create a session")
		return
	}

	c.Header("X-Auth-Token", token)
	c.Header("Location", sessionsPath+"/"+sess.id)
	writeJSON(c, http.StatusCreated, sessionBody(sess))
}

// deleteSession is the logout. Anyone may end their own session; ending
// someone else's needs admin.
func (s *Service) deleteSession(c *gin.Context) {
	sess, ok := s.sessions.get(c.Param("id"))
	if !ok {
		notFound(c)
		return
	}

	p := currentPrincipal(c)
	if !p.admin && sess.username != p.username {
		writeError(c, http.StatusForbidden, "InsufficientPrivilege", "")
		return
	}

	s.sessions.remove(sess.id)
	c.Status(http.StatusNoContent)
}
