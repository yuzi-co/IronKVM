package audit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// Service holds the handlers of /api/audit and /api/agent/tiers. Every route
// is admin-only: the log names the other users and their addresses.
type Service struct {
	store *Store
}

func NewService() *Service {
	return &Service{store: std}
}

// queryFrom reads the filter and the paging from the query string.
func queryFrom(c *gin.Context) (Query, error) {
	q := Query{
		Category: strings.TrimSpace(c.Query("category")),
		Outcome:  strings.TrimSpace(c.Query("outcome")),
		Actor:    strings.TrimSpace(c.Query("actor")),
		IP:       strings.TrimSpace(c.Query("ip")),
		Text:     strings.TrimSpace(c.Query("q")),
	}
	for name, dst := range map[string]*time.Time{"since": &q.Since, "until": &q.Until} {
		raw := c.Query(name)
		if raw == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return q, fmt.Errorf("%s is not an RFC 3339 time", name)
		}
		*dst = t
	}
	for name, dst := range map[string]*int{"offset": &q.Offset, "limit": &q.Limit} {
		raw := c.Query(name)
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return q, fmt.Errorf("%s is not a number", name)
		}
		*dst = n
	}
	return q, nil
}

// List answers GET /api/audit with a page of events, newest first.
func (s *Service) List(c *gin.Context) {
	var rsp proto.Response
	c.Header("Cache-Control", "no-store")

	q, err := queryFrom(c)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	page, err := s.store.Query(q)
	if err != nil {
		log.Errorf("read audit log: %s", err)
		rsp.ErrRsp(c, -2, "read the audit log failed")
		return
	}
	if page.Events == nil {
		page.Events = []Event{}
	}
	rsp.OkRspWithData(c, page)
}

// Export answers GET /api/audit/export with every matching event as JSON
// lines, oldest first, the way the files hold them.
func (s *Service) Export(c *gin.Context) {
	q, err := queryFrom(c)
	if err != nil {
		var rsp proto.Response
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	events, err := s.store.collect(q)
	if err != nil {
		log.Errorf("read audit log: %s", err)
		var rsp proto.Response
		rsp.ErrRsp(c, -2, "read the audit log failed")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Content-Disposition", `attachment; filename="ironkvm-audit.jsonl"`)
	enc := json.NewEncoder(c.Writer)
	for i := len(events) - 1; i >= 0; i-- {
		if enc.Encode(events[i]) != nil {
			return
		}
	}
}

type settingsRsp struct {
	Settings
	// Directory is where the log is kept, and MaxBytes the most it holds.
	Directory string `json:"directory"`
	MaxBytes  int    `json:"maxBytes"`
}

func currentSettingsRsp() settingsRsp {
	return settingsRsp{Settings: CurrentSettings(), Directory: Dir, MaxBytes: segmentBytes * segments}
}

// GetSettings answers GET /api/audit/settings.
func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, currentSettingsRsp())
}

// SetSettings answers POST /api/audit/settings.
func (s *Service) SetSettings(c *gin.Context) {
	var req struct {
		Syslog *bool `json:"syslog"`
	}
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil || req.Syslog == nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if err := SaveSettings(Settings{Syslog: *req.Syslog}); err != nil {
		log.Errorf("save audit settings: %s", err)
		rsp.ErrRsp(c, -2, "save the settings failed")
		return
	}
	SetDetail(c, "syslog="+strconv.FormatBool(*req.Syslog))
	rsp.OkRspWithData(c, currentSettingsRsp())
}
