package rtsp

import (
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/stream/egress"
)

// settingsMu guards config's RTSP block, which the owner can change while a
// request reads it.
var settingsMu sync.RWMutex

// CurrentSettings returns the running settings, with the defaults filled in.
func CurrentSettings() config.RTSP {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return config.GetInstance().RTSP.WithDefaults()
}

// saveSettings writes the settings to server.yaml. Tests replace it.
var saveSettings = func(r config.RTSP) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.RTSP = r

	return config.Write(conf)
}

// ApplySettings saves the settings to server.yaml and applies them to the
// running configuration. The caller then calls Server.Apply.
func ApplySettings(r config.RTSP) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	if err := saveSettings(r); err != nil {
		return err
	}
	config.GetInstance().RTSP = r
	return nil
}

// Settings is what the settings page shows: the settings, and what the
// server is doing with them.
type Settings struct {
	Enabled    bool   `json:"enabled"`
	Port       int    `json:"port"`
	UDP        bool   `json:"udp"`
	UDPPort    int    `json:"udpPort"`
	Path       string `json:"path"`
	MaxClients int    `json:"maxClients"`

	Listening bool     `json:"listening"`
	Error     string   `json:"error,omitempty"`
	Clients   []Client `json:"clients"`
}

// Client is a player.
type Client struct {
	Address   string `json:"address"`
	User      string `json:"user,omitempty"`
	Transport string `json:"transport,omitempty"`
	Since     string `json:"since"`
	Playing   bool   `json:"playing"`
}

type setSettingsRequest struct {
	Enabled *bool `json:"enabled" form:"enabled" validate:"required"`
	Port    int   `json:"port" form:"port" validate:"min=1,max=65535"`
	UDP     bool  `json:"udp" form:"udp"`
	// RTCP takes the port above the RTP one, so the RTP one stops short of
	// the top.
	UDPPort int `json:"udpPort" form:"udpPort" validate:"min=1,max=65534"`
}

// State is the settings page's view of the server.
func (s *Server) State() Settings {
	settings := s.deps.Settings()

	s.mu.Lock()
	defer s.mu.Unlock()

	state := Settings{
		Enabled:    settings.Enabled,
		Port:       settings.Port,
		UDP:        settings.UDP,
		UDPPort:    settings.UDPPort,
		Path:       Path,
		MaxClients: egress.MaxClients,
		Listening:  s.rtsp != nil,
		Error:      s.listenErr,
		Clients:    []Client{},
	}
	for _, sess := range s.sessions {
		client := Client{
			Address: sess.client,
			User:    sess.viewer.Username,
			Since:   sess.since.Format(time.RFC3339),
			Playing: sess.playing,
		}
		if t := sess.session.Transport(); t != nil {
			client.Transport = transportName(t.Protocol)
		}
		state.Clients = append(state.Clients, client)
	}
	sort.Slice(state.Clients, func(i, j int) bool { return state.Clients[i].Since < state.Clients[j].Since })
	return state
}

func transportName(p gortsplib.Protocol) string {
	if p == gortsplib.ProtocolTCP {
		return "tcp"
	}
	return "udp"
}

// GetSettings answers the settings page.
func (s *Server) GetSettings(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, s.State())
}

// SetSettings saves the settings and applies them at once. Any change ends
// the players there are.
func (s *Server) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if s.deps.ReservedPorts != nil && slices.Contains(s.deps.ReservedPorts(), req.Port) {
		rsp.ErrRsp(c, -1, "the port is used by the web server")
		return
	}

	err := s.deps.SetSettings(config.RTSP{
		Enabled: *req.Enabled,
		Port:    req.Port,
		UDP:     req.UDP,
		UDPPort: req.UDPPort,
	})
	if err != nil {
		log.Errorf("rtsp: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the settings")
		return
	}

	if err := s.Apply(); err != nil {
		rsp.ErrRsp(c, -3, "the settings were saved, but the server cannot listen: "+err.Error())
		return
	}

	log.Infof("rtsp: settings saved, enabled %t, port %d, udp %t", *req.Enabled, req.Port, req.UDP)
	rsp.OkRspWithData(c, s.State())
}
