package slot

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
)

// Error codes beside -1, so the page can word the common refusals itself.
const (
	codeFailed      = -1
	codeUnavailable = -2
	codeBusy        = -3
	codeInvalid     = -4
)

// Service answers the /api/slot routes.
type Service struct {
	install *installer
}

func NewService() *Service {
	in := &installer{}
	in.onExit = recordResult
	return &Service{install: in}
}

// recordResult puts the end of an install in the audit log. The request that
// started it was recorded when it answered, before a byte was written, so
// this is the only line that says whether the slot now holds the image.
func recordResult(job Job, ok bool, reason string) {
	e := audit.Event{
		Actor:    job.actor,
		Via:      job.via,
		Key:      job.key,
		Role:     job.role,
		Category: audit.CategorySystem,
		Action:   "slot-install-result",
		Target:   target(job.Slot, job.Source),
		Outcome:  audit.OutcomeOK,
	}
	if !ok {
		e.Outcome = audit.OutcomeFail
		e.Detail = reason
	}
	audit.Record(e)
}

func target(slot, source string) string {
	return fmt.Sprintf("slot %s from %s", slot, source)
}

// GetStatus answers GET /api/slot: `slot status`, read into fields.
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response
	st, err := readStatus(c.Request.Context())
	if err != nil {
		rsp.ErrRsp(c, codeOf(err), err.Error())
		return
	}
	rsp.OkRspWithData(c, st)
}

// GetImages answers GET /api/slot/images: the slot images already on /data.
func (s *Service) GetImages(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, listImages())
}

// GetProgress answers GET /api/slot/install: the install that runs, or ran
// last. It reads two small files and runs nothing, so the page can ask every
// second.
func (s *Service) GetProgress(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.install.progress())
}

type installRequest struct {
	Slot string `json:"slot" validate:"required,oneof=a b recovery"`
	// One of the two: an https URL, or the path of a file on /data.
	URL  string `json:"url"`
	File string `json:"file"`
	// Force writes an image the slot already holds.
	Force bool `json:"force"`
}

// Install answers POST /api/slot/install. It answers once the tool has
// started; the page follows the install through GetProgress.
func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response
	var req installRequest
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, codeInvalid, "invalid arguments")
		return
	}

	var source string
	var err error
	switch {
	case req.URL != "" && req.File == "":
		source, err = validateURL(req.URL)
	case req.File != "" && req.URL == "":
		source, err = resolveFile(req.File)
	default:
		err = errBadSource
	}
	if err != nil {
		rsp.ErrRsp(c, codeInvalid, err.Error())
		return
	}
	audit.SetTarget(c, target(req.Slot, displaySource(source)))

	if !Available() {
		rsp.ErrRsp(c, codeUnavailable, errUnavailable.Error())
		return
	}

	// The tool refuses all of these too. Asking first answers before a
	// process is started, and the records also show an install started from
	// a shell, which this server's own lock cannot see.
	st, err := readStatus(c.Request.Context())
	if err != nil {
		rsp.ErrRsp(c, codeOf(err), err.Error())
		return
	}
	if name := st.Installing(); name != "" {
		rsp.ErrRsp(c, codeBusy, fmt.Sprintf("slot %s is being installed right now", name))
		return
	}
	if _, ok := st.Slot(req.Slot); !ok {
		rsp.ErrRsp(c, codeInvalid, fmt.Sprintf("this board has no slot %s", req.Slot))
		return
	}
	if req.Slot == st.Running {
		rsp.ErrRsp(c, codeInvalid, fmt.Sprintf("slot %s is running; install onto another slot", req.Slot))
		return
	}

	who := identity(c)
	job, err := s.install.start(req.Slot, source, req.Force, who)
	if err != nil {
		rsp.ErrRsp(c, codeOf(err), err.Error())
		return
	}
	rsp.OkRspWithData(c, job)
}

type tryRequest struct {
	Slot string `json:"slot" validate:"required,oneof=a b"`
}

// Try answers POST /api/slot/try: the slot is booted once, on the next boot,
// and the board returns to the trusted slot if it does not come up.
func (s *Service) Try(c *gin.Context) {
	var rsp proto.Response
	var req tryRequest
	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, codeInvalid, "invalid arguments")
		return
	}
	audit.SetTarget(c, "slot "+req.Slot)
	s.decide(c, tryTimeout, "try", req.Slot)
}

// Confirm answers POST /api/slot/confirm: the running slot becomes trusted.
func (s *Service) Confirm(c *gin.Context) {
	s.decide(c, decideTimeout, "confirm")
}

// Revert answers POST /api/slot/revert: the slot trusted before the last
// confirm is trusted again, from the next boot.
func (s *Service) Revert(c *gin.Context) {
	s.decide(c, decideTimeout, "revert")
}

// decide runs one of the commands that change what the board boots, and
// answers with the status after it.
func (s *Service) decide(c *gin.Context, timeout time.Duration, args ...string) {
	var rsp proto.Response
	if _, err := run(c.Request.Context(), timeout, args...); err != nil {
		rsp.ErrRsp(c, codeOf(err), err.Error())
		return
	}
	st, err := readStatus(c.Request.Context())
	if err != nil {
		// The decision was made; only the answer about it failed.
		rsp.OkRsp(c)
		return
	}
	rsp.OkRspWithData(c, st)
}

func codeOf(err error) int {
	switch {
	case errors.Is(err, errUnavailable):
		return codeUnavailable
	case errors.Is(err, errBusy):
		return codeBusy
	default:
		return codeFailed
	}
}

// identity is who asked for an install, kept for the line that records its
// end.
func identity(c *gin.Context) audit.Identity {
	id, _ := audit.IdentityOf(c)
	return id
}
