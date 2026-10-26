package hostlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
)

// The web UI's Host log page. These handlers answer under /api/hostlog with
// the web UI's response envelope, behind its session check and admin role.

// Settings is the receiver's settings as the page shows and saves them.
type Settings struct {
	Syslog     config.HostLogListener `json:"syslog"`
	Netconsole config.HostLogListener `json:"netconsole"`
	Bind       string                 `json:"bind"`
	Sources    []string               `json:"sources"`
	MemoryKB   int                    `json:"memoryKb"`
	Disk       bool                   `json:"disk"`
	DiskKB     int                    `json:"diskKb"`
	Patterns   []string               `json:"patterns"`
}

func settingsOf(h config.HostLog) Settings {
	sources := h.Sources
	if sources == nil {
		sources = []string{}
	}
	return Settings{
		Syslog:     h.Syslog,
		Netconsole: h.Netconsole,
		Bind:       h.Bind,
		Sources:    sources,
		MemoryKB:   h.MemoryKB,
		Disk:       h.Disk,
		DiskKB:     h.DiskKB,
		Patterns:   h.Patterns,
	}
}

// Limits the server enforces on a save.
const (
	minMemoryKB = 64
	maxMemoryKB = 4096
	minDiskKB   = 256
	maxDiskKB   = 65536
	maxSources  = 32
	maxPatterns = 16
	maxPattern  = 200
)

// validate checks a save and returns the settings to store.
func (s Settings) validate() (config.HostLog, error) {
	for _, p := range []int{s.Syslog.Port, s.Netconsole.Port} {
		if p < 1 || p > 65535 {
			return config.HostLog{}, fmt.Errorf("a port must be 1 to 65535")
		}
	}
	if s.Syslog.Enabled && s.Netconsole.Enabled && s.Syslog.Port == s.Netconsole.Port {
		return config.HostLog{}, fmt.Errorf("the two listeners need different ports")
	}
	if s.Bind != "" && s.Bind != "usb" {
		return config.HostLog{}, fmt.Errorf("bind must be empty or usb")
	}
	if len(s.Sources) > maxSources {
		return config.HostLog{}, fmt.Errorf("at most %d sources", maxSources)
	}
	if _, err := parseSources(s.Sources); err != nil {
		return config.HostLog{}, err
	}
	if s.MemoryKB < minMemoryKB || s.MemoryKB > maxMemoryKB {
		return config.HostLog{}, fmt.Errorf("the memory buffer is %d to %d KB", minMemoryKB, maxMemoryKB)
	}
	if s.DiskKB < minDiskKB || s.DiskKB > maxDiskKB {
		return config.HostLog{}, fmt.Errorf("the disk log is %d to %d KB", minDiskKB, maxDiskKB)
	}
	if len(s.Patterns) > maxPatterns {
		return config.HostLog{}, fmt.Errorf("at most %d phrases", maxPatterns)
	}
	for _, p := range s.Patterns {
		if p == "" || len(p) > maxPattern {
			return config.HostLog{}, fmt.Errorf("a phrase is 1 to %d characters", maxPattern)
		}
	}
	if _, err := patternsOf(s.Patterns); err != nil {
		return config.HostLog{}, err
	}

	return config.HostLog{
		Syslog:     s.Syslog,
		Netconsole: s.Netconsole,
		Bind:       s.Bind,
		Sources:    s.Sources,
		MemoryKB:   s.MemoryKB,
		Disk:       s.Disk,
		DiskKB:     s.DiskKB,
		Patterns:   s.Patterns,
	}, nil
}

func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, settingsOf(s.deps.Settings()))
}

func (s *Service) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req Settings

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	next, err := req.validate()
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	if s.deps.SetSettings == nil {
		rsp.ErrRsp(c, -2, "the setting cannot be changed")
		return
	}
	if err := s.deps.SetSettings(next); err != nil {
		log.Errorf("hostlog: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the setting")
		return
	}
	s.Reconcile()

	log.Infof("hostlog settings: syslog=%+v netconsole=%+v bind=%q disk=%t", next.Syslog, next.Netconsole, next.Bind, next.Disk)
	rsp.OkRspWithData(c, settingsOf(s.deps.Settings()))
}

func (s *Service) GetState(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.State())
}

// Query limits.
const (
	defaultLines = 500
	maxLines     = 5000
)

func filterOf(c *gin.Context) Filter {
	f := Filter{
		Words:       ParseWords(c.Query("q")),
		Source:      c.Query("source"),
		MaxSeverity: -1,
	}
	if v, err := strconv.Atoi(c.Query("severity")); err == nil && v >= 0 && v <= 7 {
		f.MaxSeverity = v
	}
	if v, err := strconv.ParseUint(c.Query("after"), 10, 64); err == nil {
		f.After = v
	}
	return f
}

// GetLines returns the newest lines that pass the filter, oldest first.
func (s *Service) GetLines(c *gin.Context) {
	var rsp proto.Response

	limit := defaultLines
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, maxLines)
	}
	lines := s.ring.Query(filterOf(c), limit)
	if lines == nil {
		lines = []Entry{}
	}
	rsp.OkRspWithData(c, lines)
}

// Download sends the lines that pass the filter as a text file.
func (s *Service) Download(c *gin.Context) {
	lines := s.ring.Query(filterOf(c), 0)

	var b strings.Builder
	for _, e := range lines {
		b.WriteString(Format(e))
		b.WriteByte('\n')
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="hostlog.txt"`)
	c.String(http.StatusOK, "%s", b.String())
}

// Format is a line of the downloaded text.
func Format(e Entry) string {
	var b strings.Builder
	b.WriteString(e.Time.UTC().Format(time.RFC3339))
	b.WriteByte(' ')
	b.WriteString(e.Source)
	if e.Host != "" {
		b.WriteByte(' ')
		b.WriteString(e.Host)
	}
	if e.Tag != "" {
		b.WriteByte(' ')
		b.WriteString(e.Tag)
		b.WriteByte(':')
	}
	b.WriteByte(' ')
	b.WriteString(e.Msg)
	return b.String()
}

func (s *Service) ClearLog(c *gin.Context) {
	var rsp proto.Response

	if err := s.Clear(); err != nil {
		log.Warnf("hostlog: clear the disk log: %s", err)
		rsp.ErrRsp(c, -2, "failed to clear the log")
		return
	}
	rsp.OkRsp(c)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 16 * 1024,
	CheckOrigin:     middleware.SameOrigin,
}

const (
	streamBacklog = 200
	writeWait     = 10 * time.Second
	pingEvery     = 30 * time.Second
)

// streamMessage is one frame of the live tail. The first is the backlog, then
// each new line follows as an entry.
type streamMessage struct {
	Type    string  `json:"type"`
	Entries []Entry `json:"entries"`
}

// Stream is the live tail: the newest lines that pass the filter, then every
// new one. The filter comes from the query string, as for GetLines; to change
// it the page opens a new socket.
func (s *Service) Stream(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("hostlog: failed to init websocket: %s", err)
		return
	}
	stopWatch := middleware.WatchWebSocket(c.Request.Context(), ws)
	defer stopWatch()
	defer func() { _ = ws.Close() }()

	filter := filterOf(c)
	live, cancel := s.ring.Subscribe()
	defer cancel()

	backlog := s.ring.Query(filter, streamBacklog)
	if backlog == nil {
		backlog = []Entry{}
	}
	if !writeJSON(ws, streamMessage{Type: "backlog", Entries: backlog}) {
		return
	}
	if n := len(backlog); n > 0 {
		filter.After = backlog[n-1].Seq
	}

	// The reader exists to notice a close; the page sends nothing.
	gone := make(chan struct{})
	go func() {
		defer close(gone)
		ws.SetReadLimit(1024)
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ping := time.NewTicker(pingEvery)
	defer ping.Stop()
	for {
		select {
		case <-gone:
			return
		case <-ping.C:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			if ws.WriteMessage(websocket.PingMessage, nil) != nil {
				return
			}
		case e := <-live:
			batch := []Entry{}
			if filter.Match(e) {
				batch = append(batch, e)
			}
			// Take what is already waiting, so a burst is one frame.
		drain:
			for len(batch) < 200 {
				select {
				case more := <-live:
					if filter.Match(more) {
						batch = append(batch, more)
					}
				default:
					break drain
				}
			}
			if len(batch) == 0 {
				continue
			}
			if !writeJSON(ws, streamMessage{Type: "entries", Entries: batch}) {
				return
			}
		}
	}
}

func writeJSON(ws *websocket.Conn, v any) bool {
	data, err := json.Marshal(v)
	if err != nil {
		return false
	}
	_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.WriteMessage(websocket.TextMessage, data) == nil
}

// ValidateSettings checks h against the limits the settings page applies and
// returns it as the page would save it. h must have its defaults filled in.
func ValidateSettings(h config.HostLog) (config.HostLog, error) {
	return settingsOf(h).validate()
}
