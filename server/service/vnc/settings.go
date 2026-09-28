package vnc

import (
	"slices"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
)

// settingsMu guards config's VNC block, which the owner can change while a
// handshake reads it.
var settingsMu sync.RWMutex

// CurrentSettings returns the running VNC settings, with the defaults filled
// in.
func CurrentSettings() config.VNC {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return config.GetInstance().VNC.WithDefaults()
}

// saveSettings writes the settings to server.yaml. Tests replace it.
var saveSettings = func(v config.VNC) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.VNC = v

	return config.Write(conf)
}

// ApplySettings saves the settings to server.yaml and applies them to the
// running configuration. The caller then calls Server.Apply.
func ApplySettings(v config.VNC) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	if err := saveSettings(v); err != nil {
		return err
	}
	config.GetInstance().VNC = v
	return nil
}

// The web UI's VNC page. These handlers answer under /api/vnc, behind the
// session check and the admin role.

// Settings is what the page shows and saves. The password is never sent back:
// PasswordSet says whether there is one.
type Settings struct {
	Enabled     bool `json:"enabled"`
	Port        int  `json:"port"`
	MaxFPS      int  `json:"maxFps"`
	VNCAuth     bool `json:"vncAuth"`
	PasswordSet bool `json:"passwordSet"`
}

type setSettingsRequest struct {
	Enabled *bool `json:"enabled" form:"enabled" validate:"required"`
	Port    int   `json:"port" form:"port" validate:"min=1,max=65535"`
	MaxFPS  int   `json:"maxFps" form:"maxFps" validate:"min=1,max=60"`
	VNCAuth *bool `json:"vncAuth" form:"vncAuth" validate:"required"`
	// Password replaces the VNC password when it is not empty.
	Password string `json:"password" form:"password"`
}

func (s *Server) GetSettings(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, s.settings())
}

func (s *Server) settings() Settings {
	settings := s.deps.Settings()
	_, passwordSet, err := s.deps.Password.Get()
	if err != nil {
		log.Errorf("vnc: read the VNC password: %s", err)
	}
	return Settings{
		Enabled:     settings.Enabled,
		Port:        settings.Port,
		MaxFPS:      settings.MaxFPS,
		VNCAuth:     settings.VNCAuth,
		PasswordSet: passwordSet,
	}
}

// SetSettings saves the settings and applies them at once. A change of port,
// or turning the server off, ends the open session.
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
	if req.Password != "" {
		if err := validatePassword(req.Password); err != nil {
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}
	if *req.VNCAuth && req.Password == "" {
		if _, ok, _ := s.deps.Password.Get(); !ok {
			rsp.ErrRsp(c, -1, "plain VNC authentication needs a VNC password")
			return
		}
	}

	if req.Password != "" {
		if err := s.deps.Password.Set(req.Password); err != nil {
			log.Errorf("vnc: save the VNC password: %s", err)
			rsp.ErrRsp(c, -2, "failed to save the password")
			return
		}
	}

	err := s.deps.SetSettings(config.VNC{
		Enabled: *req.Enabled,
		Port:    req.Port,
		MaxFPS:  req.MaxFPS,
		VNCAuth: *req.VNCAuth,
	})
	if err != nil {
		log.Errorf("vnc: save the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the settings")
		return
	}

	if err := s.Apply(); err != nil {
		rsp.ErrRsp(c, -3, "the settings were saved, but the server cannot listen: "+err.Error())
		return
	}

	log.Infof("vnc: settings saved, enabled %t, port %d", *req.Enabled, req.Port)
	rsp.OkRsp(c)
}

// GetState reports the listener and the open session.
func (s *Server) GetState(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, s.State())
}

// EndSession ends the open session.
func (s *Server) EndSession(c *gin.Context) {
	var rsp proto.Response

	if !s.Disconnect() {
		rsp.ErrRsp(c, -1, "no open session")
		return
	}
	rsp.OkRsp(c)
}
