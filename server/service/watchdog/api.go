package watchdog

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
)

// The web UI's Watchdog page. These handlers answer under /api/watchdog with
// the web UI's response envelope, behind its session check and admin role.

// Settings is the watchdog's settings as the page shows and saves them.
type Settings struct {
	Enabled         bool   `json:"enabled"`
	TimeoutMinutes  int    `json:"timeoutMinutes"`
	Action          string `json:"action"`
	CooldownMinutes int    `json:"cooldownMinutes"`
	MaxPerHour      int    `json:"maxPerHour"`
	PingHost        string `json:"pingHost"`
}

type setSettingsRequest struct {
	Enabled         *bool  `json:"enabled" form:"enabled" validate:"required"`
	TimeoutMinutes  int    `json:"timeoutMinutes" form:"timeoutMinutes" validate:"min=1,max=1440"`
	Action          string `json:"action" form:"action" validate:"oneof=reset power"`
	CooldownMinutes int    `json:"cooldownMinutes" form:"cooldownMinutes" validate:"min=1,max=1440"`
	MaxPerHour      int    `json:"maxPerHour" form:"maxPerHour" validate:"min=1,max=20"`
	PingHost        string `json:"pingHost" form:"pingHost" validate:"omitempty,ip"`
}

// State is what the detector sees now. A time is absent until it happens.
type State struct {
	Status       string `json:"status"`
	Signal       bool   `json:"signal"`
	LEDConnected bool   `json:"ledConnected"`
	LEDOn        bool   `json:"ledOn"`
	// PingHost is empty when no ping is set; PingOK is the last reply.
	PingHost string `json:"pingHost"`
	PingOK   bool   `json:"pingOK"`

	LastSample *time.Time `json:"lastSample,omitempty"`
	LastChange *time.Time `json:"lastChange,omitempty"`
	LastPing   *time.Time `json:"lastPing,omitempty"`
	LastAction *time.Time `json:"lastAction,omitempty"`

	// ActsInSeconds is how long the host may still show no sign of life
	// before the watchdog acts. It is present only while it is watching.
	ActsInSeconds *int `json:"actsInSeconds,omitempty"`
	// ActionsLastHour counts toward MaxPerHour.
	ActionsLastHour int `json:"actionsLastHour"`
}

func (w *Watchdog) GetSettings(c *gin.Context) {
	var rsp proto.Response

	s := w.deps.Settings()
	rsp.OkRspWithData(c, &Settings{
		Enabled:         s.Enabled,
		TimeoutMinutes:  s.TimeoutMinutes,
		Action:          s.Action,
		CooldownMinutes: s.CooldownMinutes,
		MaxPerHour:      s.MaxPerHour,
		PingHost:        s.PingHost,
	})
}

// SetSettings saves the settings. Any change starts the timer again, so a new
// timeout or a watchdog just turned on never acts at once.
func (w *Watchdog) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if w.deps.SetSettings == nil {
		rsp.ErrRsp(c, -2, "the setting cannot be changed")
		return
	}

	s := config.Watchdog{
		Enabled:         *req.Enabled,
		TimeoutMinutes:  req.TimeoutMinutes,
		Action:          req.Action,
		CooldownMinutes: req.CooldownMinutes,
		MaxPerHour:      req.MaxPerHour,
		PingHost:        req.PingHost,
	}
	if err := w.deps.SetSettings(s); err != nil {
		log.Errorf("watchdog: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the setting")
		return
	}
	w.rearm()

	log.Infof("watchdog settings: %+v", s)
	rsp.OkRsp(c)
}

func (w *Watchdog) GetState(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, w.State())
}

// State returns what the detector sees now.
func (w *Watchdog) State() State {
	s := w.deps.Settings()
	now := w.deps.Now()

	w.mu.Lock()
	defer w.mu.Unlock()

	st := State{
		Status:       w.status,
		Signal:       w.signal,
		LEDConnected: w.ledConnected,
		LEDOn:        w.ledOn,
		PingOK:       w.pingOK,
		LastSample:   timePtr(w.lastSample),
		LastChange:   timePtr(w.lastChange),
		LastPing:     timePtr(w.lastPing),
		LastAction:   timePtr(w.lastAction),
	}
	// Off, the watchdog samples nothing, so the last sample says nothing about
	// now. The signal and the LED are cheap to read and are read here instead.
	if !s.Enabled {
		st.Status = StatusOff
		st.Signal = w.deps.Signal()
		st.LEDConnected = w.deps.PowerLEDConnected()
		st.LEDOn = false
		if st.LEDConnected {
			on, err := w.deps.PowerLED()
			st.LEDOn = err == nil && on
		}
	}
	if s.PingHost != "" {
		st.PingHost = s.PingHost
	}
	if st.Status == StatusWatching {
		left := time.Duration(s.TimeoutMinutes)*time.Minute - now.Sub(w.lastAlive)
		seconds := max(0, int(left/time.Second))
		st.ActsInSeconds = &seconds
	}
	for _, t := range w.actions {
		if now.Sub(t) < time.Hour {
			st.ActionsLastHour++
		}
	}
	return st
}

// GetLog lists the actions, newest first.
func (w *Watchdog) GetLog(c *gin.Context) {
	var rsp proto.Response

	entries := []Entry{}
	if w.deps.Log != nil {
		entries = w.deps.Log.Entries()
	}
	rsp.OkRspWithData(c, entries)
}

// GetScreenshot serves the screenshot taken before one action.
func (w *Watchdog) GetScreenshot(c *gin.Context) {
	var rsp proto.Response

	if w.deps.Log == nil {
		rsp.ErrRsp(c, -1, "no such screenshot")
		return
	}
	path, ok := w.deps.Log.Screenshot(c.Param("id"))
	if !ok {
		rsp.ErrRsp(c, -1, "no such screenshot")
		return
	}

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "private, max-age=86400")
	c.File(path)
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	u := t.UTC()
	return &u
}
