package srt

import (
	"sort"
	"sync"
	"time"

	gosrt "github.com/datarhei/gosrt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/stream/egress"
)

// settingsMu guards config's SRT block, which the owner can change while a
// request reads it.
var settingsMu sync.RWMutex

// CurrentSettings returns the running settings, with the defaults filled in.
func CurrentSettings() config.SRT {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return config.GetInstance().SRT.WithDefaults()
}

// saveSettings writes the settings to server.yaml. Tests replace it.
var saveSettings = func(v config.SRT) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.SRT = v

	return config.Write(conf)
}

// ApplySettings saves the settings to server.yaml and applies them to the
// running configuration. The caller then calls Server.Apply.
func ApplySettings(v config.SRT) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	if err := saveSettings(v); err != nil {
		return err
	}
	config.GetInstance().SRT = v
	return nil
}

// Settings is what the settings page shows: the settings, and what the
// listener is doing with them. The passphrase is never sent back:
// PassphraseSet says whether there is one.
type Settings struct {
	Enabled       bool `json:"enabled"`
	Port          int  `json:"port"`
	LatencyMs     int  `json:"latencyMs"`
	PassphraseSet bool `json:"passphraseSet"`
	MaxClients    int  `json:"maxClients"`

	Listening bool     `json:"listening"`
	Error     string   `json:"error,omitempty"`
	Clients   []Client `json:"clients"`
}

// Client is a caller, with SRT's own view of its link.
type Client struct {
	Address string `json:"address"`
	User    string `json:"user,omitempty"`
	Since   string `json:"since"`
	// RTTMs is the smoothed round trip.
	RTTMs float64 `json:"rttMs"`
	// SendRateMbps is what the caller is being sent now.
	SendRateMbps float64 `json:"sendRateMbps"`
	// Sent, Lost, Retransmitted and Dropped are packet counts since it
	// connected. Lost were reported missing by the caller, and Dropped never
	// reached it in time.
	Sent          uint64 `json:"sent"`
	Lost          uint64 `json:"lost"`
	Retransmitted uint64 `json:"retransmitted"`
	Dropped       uint64 `json:"dropped"`
}

type setSettingsRequest struct {
	Enabled   *bool `json:"enabled" form:"enabled" validate:"required"`
	Port      int   `json:"port" form:"port" validate:"min=1,max=65535"`
	LatencyMs int   `json:"latencyMs" form:"latencyMs" validate:"min=20,max=8000"`
	// Passphrase replaces the passphrase when it is not empty.
	Passphrase string `json:"passphrase" form:"passphrase"`
}

// State is the settings page's view of the listener.
func (s *Server) State() Settings {
	settings := s.deps.Settings()
	passphraseSet := false
	if s.deps.Passphrase != nil {
		_, passphraseSet, _ = s.deps.Passphrase.Get()
	}

	s.mu.Lock()
	state := Settings{
		Enabled:       settings.Enabled,
		Port:          settings.Port,
		LatencyMs:     settings.LatencyMs,
		PassphraseSet: passphraseSet,
		MaxClients:    egress.MaxClients,
		Listening:     s.listener != nil,
		Error:         s.listenErr,
		Clients:       []Client{},
	}
	clients := make([]*client, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.Unlock()

	for _, c := range clients {
		var stats gosrt.Statistics
		c.conn.Stats(&stats)
		state.Clients = append(state.Clients, Client{
			Address:       c.address,
			User:          c.viewer.Username,
			Since:         c.since.Format(time.RFC3339),
			RTTMs:         stats.Instantaneous.MsRTT,
			SendRateMbps:  stats.Instantaneous.MbpsSentRate,
			Sent:          stats.Accumulated.PktSent,
			Lost:          stats.Accumulated.PktSendLoss,
			Retransmitted: stats.Accumulated.PktRetrans,
			Dropped:       stats.Accumulated.PktSendDrop,
		})
	}
	sort.Slice(state.Clients, func(i, j int) bool { return state.Clients[i].Since < state.Clients[j].Since })
	return state
}

// GetSettings answers the settings page.
func (s *Server) GetSettings(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, s.State())
}

// SetSettings saves the settings and applies them at once. Any change ends
// the callers there are.
func (s *Server) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if req.Passphrase != "" {
		if err := ValidatePassphrase(req.Passphrase); err != nil {
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}
	if *req.Enabled && req.Passphrase == "" {
		if _, ok, _ := s.deps.Passphrase.Get(); !ok {
			rsp.ErrRsp(c, -1, "SRT needs a passphrase")
			return
		}
	}

	if req.Passphrase != "" {
		if err := s.deps.Passphrase.Set(req.Passphrase); err != nil {
			log.Errorf("srt: save the passphrase: %s", err)
			rsp.ErrRsp(c, -2, "failed to save the passphrase")
			return
		}
	}

	err := s.deps.SetSettings(config.SRT{
		Enabled:   *req.Enabled,
		Port:      req.Port,
		LatencyMs: req.LatencyMs,
	})
	if err != nil {
		log.Errorf("srt: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the settings")
		return
	}

	if err := s.Apply(); err != nil {
		rsp.ErrRsp(c, -3, "the settings were saved, but the server cannot listen: "+err.Error())
		return
	}

	log.Infof("srt: settings saved, enabled %t, port %d, latency %d ms", *req.Enabled, req.Port, req.LatencyMs)
	rsp.OkRspWithData(c, s.State())
}
