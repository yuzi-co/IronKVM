package nut

import (
	"math"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hostpower"
)

// The web UI's NUT page. The settings answer under /api/nut behind the admin
// role, and the state behind a login only, because the status bar shows it to
// every viewer.

// Settings is the client's settings as the page shows and saves them. The
// password never comes back: PasswordSet says whether there is one.
type Settings struct {
	Enabled             bool   `json:"enabled"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	UPS                 string `json:"ups"`
	Username            string `json:"username"`
	PasswordSet         bool   `json:"passwordSet"`
	OnBatteryMinutes    int    `json:"onBatteryMinutes"`
	ForceOffMinutes     int    `json:"forceOffMinutes"`
	PowerOnRestore      bool   `json:"powerOnRestore"`
	PowerOnMinCharge    int    `json:"powerOnMinCharge"`
	PowerOnDelaySeconds int    `json:"powerOnDelaySeconds"`
}

type setSettingsRequest struct {
	Enabled             *bool   `json:"enabled" form:"enabled" validate:"required"`
	Host                string  `json:"host" form:"host" validate:"max=253"`
	Port                int     `json:"port" form:"port" validate:"min=1,max=65535"`
	UPS                 string  `json:"ups" form:"ups" validate:"max=64"`
	Username            string  `json:"username" form:"username" validate:"max=64"`
	Password            *string `json:"password" form:"password" validate:"omitempty,max=128"`
	ClearPassword       bool    `json:"clearPassword" form:"clearPassword"`
	OnBatteryMinutes    int     `json:"onBatteryMinutes" form:"onBatteryMinutes" validate:"min=0,max=1440"`
	ForceOffMinutes     int     `json:"forceOffMinutes" form:"forceOffMinutes" validate:"min=0,max=1440"`
	PowerOnRestore      bool    `json:"powerOnRestore" form:"powerOnRestore"`
	PowerOnMinCharge    int     `json:"powerOnMinCharge" form:"powerOnMinCharge" validate:"min=0,max=100"`
	PowerOnDelaySeconds int     `json:"powerOnDelaySeconds" form:"powerOnDelaySeconds" validate:"min=10,max=3600"`
}

// State is what the client sees and does now. A time is absent until it
// happens, and the UPS fields are absent while the server cannot be read.
type State struct {
	Enabled bool `json:"enabled"`
	// Status is one of the Status constants.
	Status    string `json:"status"`
	Reachable bool   `json:"reachable"`

	// UPS is one of the UPS constants, and Flags the words of ups.status.
	UPS   string   `json:"ups"`
	Flags []string `json:"flags,omitempty"`
	// Charge is in percent, RuntimeSeconds is the UPS's estimate of the time
	// left on battery.
	Charge           *int `json:"charge,omitempty"`
	RuntimeSeconds   *int `json:"runtimeSeconds,omitempty"`
	OnBatterySeconds *int `json:"onBatterySeconds,omitempty"`

	// LoggedIn says the client is registered with the server as a client of
	// the UPS, which is what a primary waits for. LoginError is the server's
	// reason when it refused.
	LoggedIn   bool   `json:"loggedIn"`
	LoginError string `json:"loginError,omitempty"`

	// Host is the power LED's word: on, off or unknown. Phase is one of the
	// Phase constants.
	Host  string `json:"host"`
	Phase string `json:"phase"`
	// PowerOnInSeconds is how long until the host is switched on, present
	// while the client waits out the delay.
	PowerOnInSeconds *int `json:"powerOnInSeconds,omitempty"`

	LastPoll *time.Time `json:"lastPoll,omitempty"`
	// Events are the last things the client did or saw, newest first.
	Events []Event `json:"events"`
}

func (m *Monitor) GetSettings(c *gin.Context) {
	var rsp proto.Response

	s := m.deps.Settings()
	set := false
	if m.deps.Password != nil {
		if p, err := m.deps.Password(); err == nil {
			set = p != ""
		}
	}
	rsp.OkRspWithData(c, &Settings{
		Enabled:             s.Enabled,
		Host:                s.Host,
		Port:                s.Port,
		UPS:                 s.UPS,
		Username:            s.Username,
		PasswordSet:         set,
		OnBatteryMinutes:    s.OnBatteryMinutes,
		ForceOffMinutes:     s.ForceOffMinutes,
		PowerOnRestore:      s.PowerOnRestore,
		PowerOnMinCharge:    s.PowerOnMinCharge,
		PowerOnDelaySeconds: s.PowerOnDelaySeconds,
	})
}

// SetSettings saves the settings and the password. A new password is kept
// only when the request carries one; ClearPassword removes it.
func (m *Monitor) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if m.deps.SetSettings == nil || m.deps.SetPassword == nil {
		rsp.ErrRsp(c, -2, "the setting cannot be changed")
		return
	}

	if req.Host != "" && !ValidHost(req.Host) {
		rsp.ErrRsp(c, -4, "invalid server address")
		return
	}
	if req.UPS != "" && !upsNamePattern.MatchString(req.UPS) {
		rsp.ErrRsp(c, -4, "invalid UPS name")
		return
	}
	if req.Username != "" && !upsNamePattern.MatchString(req.Username) {
		rsp.ErrRsp(c, -4, "invalid user name")
		return
	}
	if req.Password != nil && *req.Password != "" {
		if err := validPassword(*req.Password); err != nil {
			rsp.ErrRsp(c, -4, err.Error())
			return
		}
	}

	if *req.Enabled {
		if req.Host == "" || req.UPS == "" {
			rsp.ErrRsp(c, -5, "the server address and the UPS name are needed")
			return
		}
		// Without the power LED the client cannot tell a host that is off from
		// one that is on, and a short press on a host that is off switches it
		// on.
		if connected, _, _ := m.deps.Power.LED(); !connected {
			rsp.ErrRsp(c, -3, "the NUT client needs the power LED connected")
			return
		}
	}

	s := config.NUT{
		Enabled:             *req.Enabled,
		Host:                req.Host,
		Port:                req.Port,
		UPS:                 req.UPS,
		Username:            req.Username,
		OnBatteryMinutes:    req.OnBatteryMinutes,
		ForceOffMinutes:     req.ForceOffMinutes,
		PowerOnRestore:      req.PowerOnRestore,
		PowerOnMinCharge:    req.PowerOnMinCharge,
		PowerOnDelaySeconds: req.PowerOnDelaySeconds,
	}

	switch {
	case req.ClearPassword:
		if err := m.deps.SetPassword(""); err != nil {
			log.Errorf("nut: remove the password: %s", err)
			rsp.ErrRsp(c, -2, "failed to save the setting")
			return
		}
	case req.Password != nil && *req.Password != "":
		if err := m.deps.SetPassword(*req.Password); err != nil {
			log.Errorf("nut: save the password: %s", err)
			rsp.ErrRsp(c, -2, "failed to save the setting")
			return
		}
	}
	if err := m.deps.SetSettings(s); err != nil {
		log.Errorf("nut: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the setting")
		return
	}
	m.Reconnect()

	// The password is not part of s.
	log.Infof("nut settings: %+v", s)
	rsp.OkRsp(c)
}

func (m *Monitor) GetState(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, m.State())
}

// State returns what the client sees and does now.
func (m *Monitor) State() State {
	s := m.deps.Settings()
	now := m.deps.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	st := State{
		Enabled:    s.Enabled,
		Status:     m.status,
		Reachable:  m.reachable,
		UPS:        UPSUnknown,
		LoggedIn:   m.loggedIn,
		LoginError: m.loginErr,
		Host:       m.host.String(),
		Phase:      PhaseIdle,
		Events:     []Event{},
	}
	if !s.Enabled {
		st.Status = StatusOff
	}
	if !m.lastPoll.IsZero() {
		t := m.lastPoll.UTC()
		st.LastPoll = &t
	}

	if r := m.reading; r != nil && m.status == StatusOK {
		st.UPS = r.State()
		st.Flags = r.Flags
		if r.Charge >= 0 {
			v := int(math.Round(r.Charge))
			st.Charge = &v
		}
		if r.Runtime >= 0 {
			v := r.Runtime
			st.RuntimeSeconds = &v
		}
	}
	if !m.onBatterySince.IsZero() {
		v := max(0, int(now.Sub(m.onBatterySince)/time.Second))
		st.OnBatterySeconds = &v
	}

	switch {
	case s.Enabled && m.host == hostpower.Unknown && m.status != StatusUnconfigured:
		st.Phase = PhaseNeedsLED
	case m.shutdown != "":
		st.Phase = PhaseShuttingDown
		if m.shutdownHostOff {
			st.Phase = PhaseWaitingMains
		}
	case m.restore && s.PowerOnRestore && m.host == hostpower.Off:
		st.Phase = PhaseWaitingMains
	}
	if !m.powerOnSince.IsZero() {
		left := time.Duration(s.PowerOnDelaySeconds)*time.Second - now.Sub(m.powerOnSince)
		v := max(0, int(left/time.Second))
		st.PowerOnInSeconds = &v
	}

	for i := len(m.events) - 1; i >= 0; i-- {
		st.Events = append(st.Events, m.events[i])
	}
	return st
}
