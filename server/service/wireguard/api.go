package wireguard

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// The settings page's API. Every route is for administrators (router/
// wireguard.go), and the audit middleware records each change. No answer
// carries the private key or a preshared key.

// View is the stored configuration as the page sees it.
type View struct {
	// PublicKey is the board's, for the peer's side; empty without a key.
	PublicKey    string     `json:"publicKey"`
	Configured   bool       `json:"configured"`
	Addresses    []string   `json:"addresses"`
	DNS          []string   `json:"dns"`
	ListenPort   int        `json:"listenPort"`
	MTU          int        `json:"mtu"`
	Peers        []PeerView `json:"peers"`
	StartAtBoot  bool       `json:"startAtBoot"`
	DefaultRoute bool       `json:"defaultRoute"`
}

// PeerView is a peer without its preshared key.
type PeerView struct {
	PublicKey           string   `json:"publicKey"`
	HasPresharedKey     bool     `json:"hasPresharedKey"`
	Endpoint            string   `json:"endpoint"`
	AllowedIPs          []string `json:"allowedIPs"`
	PersistentKeepalive int      `json:"persistentKeepalive"`
}

// State is the answer of every route: the settings, the tunnel, and what the
// last change ignored.
type State struct {
	Settings View     `json:"settings"`
	Status   Status   `json:"status"`
	Warnings []string `json:"warnings,omitempty"`
}

func view(c Config) View {
	v := View{
		Configured:   c.Configured(),
		Addresses:    nonNil(c.Addresses),
		DNS:          nonNil(c.DNS),
		ListenPort:   c.ListenPort,
		MTU:          c.MTU,
		Peers:        []PeerView{},
		StartAtBoot:  c.StartAtBoot,
		DefaultRoute: c.DefaultRoute(),
	}
	if c.PrivateKey != "" {
		v.PublicKey, _ = PublicKey(c.PrivateKey)
	}
	for _, p := range c.Peers {
		v.Peers = append(v.Peers, PeerView{
			PublicKey:           p.PublicKey,
			HasPresharedKey:     p.PresharedKey != "",
			Endpoint:            p.Endpoint,
			AllowedIPs:          nonNil(p.AllowedIPs),
			PersistentKeepalive: p.PersistentKeepalive,
		})
	}
	return v
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

type peerRequest struct {
	PublicKey string `json:"publicKey"`
	// PresharedKey replaces the peer's; empty keeps the one stored for the
	// same public key, unless ClearPresharedKey.
	PresharedKey        string   `json:"presharedKey"`
	ClearPresharedKey   bool     `json:"clearPresharedKey"`
	Endpoint            string   `json:"endpoint"`
	AllowedIPs          []string `json:"allowedIPs"`
	PersistentKeepalive int      `json:"persistentKeepalive"`
}

type saveRequest struct {
	// PrivateKey replaces the board's; empty keeps it, and a board without
	// one makes one.
	PrivateKey string        `json:"privateKey"`
	Addresses  []string      `json:"addresses"`
	DNS        []string      `json:"dns"`
	ListenPort int           `json:"listenPort"`
	MTU        int           `json:"mtu"`
	Peers      []peerRequest `json:"peers"`
}

type importRequest struct {
	Config string `json:"config"`
}

type bootRequest struct {
	Enabled bool `json:"enabled"`
}

// Service holds the handlers.
type Service struct {
	// mu serialises the changes, which read and rewrite the file.
	mu sync.Mutex
	m  *Manager
}

// manager is the board's interface.
var manager = &Manager{sys: newSystem(), iface: interfaceName}

func NewService() *Service {
	return &Service{m: manager}
}

// Register adds the routes to an administrators' group under /api.
func (s *Service) Register(api gin.IRoutes) {
	api.GET("/network/wireguard", s.Get)                // settings and tunnel status
	api.POST("/network/wireguard", s.Save)              // save the fields, applied when up
	api.DELETE("/network/wireguard", s.Delete)          // take it down and forget it
	api.POST("/network/wireguard/import", s.Import)     // replace the settings with a wg-quick file
	api.POST("/network/wireguard/up", s.Up)             // bring the tunnel up
	api.POST("/network/wireguard/down", s.Down)         // take it down
	api.POST("/network/wireguard/boot", s.Boot)         // bring it up at boot, or not
	api.POST("/network/wireguard/key", s.RegenerateKey) // a new key pair, made on the board
}

func (s *Service) state(c Config, warnings []string) State {
	return State{Settings: view(c), Status: s.m.Status(c), Warnings: warnings}
}

func (s *Service) Get(c *gin.Context) {
	var rsp proto.Response
	conf, err := Load()
	if err != nil {
		log.Errorf("wireguard: reading the settings: %s", err)
		rsp.ErrRsp(c, -1, "failed to read the WireGuard settings")
		return
	}
	rsp.OkRspWithData(c, s.state(conf, nil))
}

// store saves conf and, when the tunnel is up, applies it. A refused or
// failed change answers with the error and keeps the file as it was.
func (s *Service) store(c *gin.Context, conf Config, warnings []string) {
	var rsp proto.Response
	if err := conf.Validate(); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	if err := Save(conf); err != nil {
		log.Errorf("wireguard: saving the settings: %s", err)
		rsp.ErrRsp(c, -4, "failed to save the WireGuard settings")
		return
	}
	if s.m.IsUp() {
		applied, err := s.m.Up(conf)
		if err != nil {
			rsp.ErrRsp(c, -3, "saved, but the tunnel failed to come up: "+err.Error())
			return
		}
		warnings = append(warnings, applied...)
	}
	rsp.OkRspWithData(c, s.state(conf, warnings))
}

func (s *Service) Save(c *gin.Context) {
	var rsp proto.Response
	var req saveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	old, err := Load()
	if err != nil {
		log.Warnf("wireguard: the stored settings are unreadable, replacing them: %s", err)
	}
	conf := Config{
		PrivateKey:  req.PrivateKey,
		Addresses:   req.Addresses,
		DNS:         req.DNS,
		ListenPort:  req.ListenPort,
		MTU:         req.MTU,
		StartAtBoot: old.StartAtBoot,
	}
	if conf.PrivateKey == "" {
		conf.PrivateKey = old.PrivateKey
	}
	if conf.PrivateKey == "" {
		if conf.PrivateKey, err = GeneratePrivateKey(); err != nil {
			rsp.ErrRsp(c, -4, "failed to make a key")
			return
		}
	}
	oldPSK := map[string]string{}
	for _, p := range old.Peers {
		oldPSK[p.PublicKey] = p.PresharedKey
	}
	for _, p := range req.Peers {
		peer := Peer{
			PublicKey:           p.PublicKey,
			PresharedKey:        p.PresharedKey,
			Endpoint:            p.Endpoint,
			AllowedIPs:          p.AllowedIPs,
			PersistentKeepalive: p.PersistentKeepalive,
		}
		if p.ClearPresharedKey {
			peer.PresharedKey = ""
		} else if peer.PresharedKey == "" {
			peer.PresharedKey = oldPSK[p.PublicKey]
		}
		conf.Peers = append(conf.Peers, peer)
	}
	s.store(c, conf, nil)
}

func (s *Service) Import(c *gin.Context) {
	var rsp proto.Response
	var req importRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	conf, warnings, err := ParseConf(req.Config)
	if err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if old, err := Load(); err == nil {
		conf.StartAtBoot = old.StartAtBoot
	}
	s.store(c, conf, warnings)
}

func (s *Service) Up(c *gin.Context) {
	var rsp proto.Response
	s.mu.Lock()
	defer s.mu.Unlock()

	conf, err := Load()
	if err != nil || !conf.Configured() {
		rsp.ErrRsp(c, -2, "WireGuard is not configured")
		return
	}
	warnings, err := s.m.Up(conf)
	if err != nil {
		rsp.ErrRsp(c, -3, err.Error())
		return
	}
	rsp.OkRspWithData(c, s.state(conf, warnings))
}

func (s *Service) Down(c *gin.Context) {
	var rsp proto.Response
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.m.Down(); err != nil {
		rsp.ErrRsp(c, -3, err.Error())
		return
	}
	conf, _ := Load()
	rsp.OkRspWithData(c, s.state(conf, nil))
}

func (s *Service) Boot(c *gin.Context) {
	var rsp proto.Response
	var req bootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	conf, err := Load()
	if err != nil || (req.Enabled && !conf.Configured()) {
		rsp.ErrRsp(c, -2, "WireGuard is not configured")
		return
	}
	conf.StartAtBoot = req.Enabled
	if err := Save(conf); err != nil {
		rsp.ErrRsp(c, -4, "failed to save the WireGuard settings")
		return
	}
	rsp.OkRspWithData(c, s.state(conf, nil))
}

// RegenerateKey replaces the key pair. The peer has to be given the new
// public key.
func (s *Service) RegenerateKey(c *gin.Context) {
	var rsp proto.Response
	s.mu.Lock()
	defer s.mu.Unlock()

	conf, err := Load()
	if err != nil {
		rsp.ErrRsp(c, -1, "failed to read the WireGuard settings")
		return
	}
	if conf.PrivateKey, err = GeneratePrivateKey(); err != nil {
		rsp.ErrRsp(c, -4, "failed to make a key")
		return
	}
	if !conf.Configured() {
		// Only a key so far, for the peer's side: nothing to check or apply.
		if err := Save(conf); err != nil {
			rsp.ErrRsp(c, -4, "failed to save the WireGuard settings")
			return
		}
		rsp.OkRspWithData(c, s.state(conf, nil))
		return
	}
	s.store(c, conf, nil)
}

func (s *Service) Delete(c *gin.Context) {
	var rsp proto.Response
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.m.Down(); err != nil {
		rsp.ErrRsp(c, -3, err.Error())
		return
	}
	if err := Remove(); err != nil {
		rsp.ErrRsp(c, -4, "failed to remove the WireGuard settings")
		return
	}
	rsp.OkRspWithData(c, s.state(Config{}, nil))
}

// Boot retries: the endpoint's name may not resolve until the network is up.
const (
	bootAttempts = 12
	bootRetry    = 10 * time.Second
	bootMark     = "/run/ironkvm-wireguard-restored"
)

// RestoreAtBoot brings the tunnel up when the server starts, if the owner
// asked for that.
//
// Once per boot, like the IPv6 setting: a server restarted by an update finds
// the tunnel as the owner left it, up or down.
func RestoreAtBoot() {
	if _, err := os.Stat(bootMark); err == nil {
		return
	}
	_ = os.WriteFile(bootMark, nil, 0o644)

	var lastErr error
	for attempt := 1; attempt <= bootAttempts; attempt++ {
		// Read again on each attempt: the owner may have changed the
		// settings, or brought the tunnel up, in the meantime.
		conf, err := Load()
		if err != nil {
			log.Errorf("wireguard: reading the settings: %s", err)
			return
		}
		if !conf.StartAtBoot || !conf.Configured() || (attempt > 1 && manager.IsUp()) {
			return
		}
		if _, lastErr = manager.Up(conf); lastErr == nil {
			return
		}
		log.Warnf("wireguard: bringing the tunnel up at boot (attempt %d): %s", attempt, lastErr)
		time.Sleep(bootRetry)
	}
	manager.SetError(errors.Join(errors.New("gave up at boot"), lastErr).Error())
}
