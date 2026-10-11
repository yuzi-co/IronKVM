package notify

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// The web UI's Notifications page. These handlers answer under /api/notify
// with the web UI's response envelope, behind its session check and admin
// role. A secret never leaves the server: the page is told whether one is set.

// View is the settings as the page reads them.
type View struct {
	Enabled            bool            `json:"enabled"`
	Snapshot           bool            `json:"snapshot"`
	Events             map[string]Rule `json:"events"`
	LoginThreshold     int             `json:"loginThreshold"`
	LoginWindowSeconds int             `json:"loginWindowSeconds"`

	Webhook struct {
		Enabled         bool   `json:"enabled"`
		URLSet          bool   `json:"urlSet"`
		URLHint         string `json:"urlHint,omitempty"`
		IncludeSnapshot bool   `json:"includeSnapshot"`
	} `json:"webhook"`
	Ntfy struct {
		Enabled  bool   `json:"enabled"`
		Server   string `json:"server"`
		Topic    string `json:"topic"`
		TokenSet bool   `json:"tokenSet"`
	} `json:"ntfy"`
	Telegram struct {
		Enabled     bool   `json:"enabled"`
		BotTokenSet bool   `json:"botTokenSet"`
		ChatID      string `json:"chatId"`
	} `json:"telegram"`
}

func viewOf(s Settings) View {
	v := View{
		Enabled:            s.Enabled,
		Snapshot:           s.Snapshot,
		Events:             s.Events,
		LoginThreshold:     s.LoginThreshold,
		LoginWindowSeconds: s.LoginWindowSeconds,
	}
	v.Webhook.Enabled = s.Webhook.Enabled
	v.Webhook.URLSet = s.Webhook.URL != ""
	v.Webhook.URLHint = hint(s.Webhook.URL)
	v.Webhook.IncludeSnapshot = s.Webhook.IncludeSnapshot
	v.Ntfy.Enabled = s.Ntfy.Enabled
	v.Ntfy.Server = s.Ntfy.Server
	v.Ntfy.Topic = s.Ntfy.Topic
	v.Ntfy.TokenSet = s.Ntfy.Token != ""
	v.Telegram.Enabled = s.Telegram.Enabled
	v.Telegram.BotTokenSet = s.Telegram.BotToken != ""
	v.Telegram.ChatID = s.Telegram.ChatID
	return v
}

// setRequest is a save. A secret is a pointer: absent keeps the stored one, an
// empty string clears it, anything else replaces it.
type setRequest struct {
	Enabled            bool            `json:"enabled"`
	Snapshot           bool            `json:"snapshot"`
	Events             map[string]Rule `json:"events"`
	LoginThreshold     int             `json:"loginThreshold"`
	LoginWindowSeconds int             `json:"loginWindowSeconds"`

	Webhook struct {
		Enabled         bool    `json:"enabled"`
		URL             *string `json:"url"`
		IncludeSnapshot bool    `json:"includeSnapshot"`
	} `json:"webhook"`
	Ntfy struct {
		Enabled bool    `json:"enabled"`
		Server  string  `json:"server"`
		Topic   string  `json:"topic"`
		Token   *string `json:"token"`
	} `json:"ntfy"`
	Telegram struct {
		Enabled  bool    `json:"enabled"`
		BotToken *string `json:"botToken"`
		ChatID   string  `json:"chatId"`
	} `json:"telegram"`
}

func (r setRequest) apply(s *Settings) {
	s.Enabled = r.Enabled
	s.Snapshot = r.Snapshot
	s.Events = r.Events
	s.LoginThreshold = r.LoginThreshold
	s.LoginWindowSeconds = r.LoginWindowSeconds

	s.Webhook.Enabled = r.Webhook.Enabled
	s.Webhook.IncludeSnapshot = r.Webhook.IncludeSnapshot
	if r.Webhook.URL != nil {
		s.Webhook.URL = *r.Webhook.URL
	}
	s.Ntfy.Enabled = r.Ntfy.Enabled
	s.Ntfy.Server = r.Ntfy.Server
	s.Ntfy.Topic = r.Ntfy.Topic
	if r.Ntfy.Token != nil {
		s.Ntfy.Token = *r.Ntfy.Token
	}
	s.Telegram.Enabled = r.Telegram.Enabled
	s.Telegram.ChatID = r.Telegram.ChatID
	if r.Telegram.BotToken != nil {
		s.Telegram.BotToken = *r.Telegram.BotToken
	}
}

func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response

	set, err := s.deps.Store.Load()
	if err != nil {
		log.Warnf("notify: load the settings: %s", err)
		rsp.ErrRsp(c, -2, "failed to read the settings")
		return
	}
	rsp.OkRspWithData(c, viewOf(set))
}

func (s *Service) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	set, err := s.deps.Store.Update(func(cur *Settings) error {
		req.apply(cur)
		return nil
	})
	if err != nil {
		// The validation errors name a field and never a value, so they are
		// safe to show. A failed write is not the owner's to fix.
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	log.Infof("notify settings saved: enabled=%t webhook=%t ntfy=%t telegram=%t",
		set.Enabled, set.Webhook.Enabled, set.Ntfy.Enabled, set.Telegram.Enabled)
	rsp.OkRspWithData(c, viewOf(set))
}

func (s *Service) GetState(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.State())
}

type testRequest struct {
	Channel string `json:"channel"`
}

// Test sends a test message and answers with what each channel said.
func (s *Service) PostTest(c *gin.Context) {
	var rsp proto.Response
	var req testRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	results, err := s.Test(req.Channel)
	if errors.Is(err, ErrNoChannel) {
		rsp.ErrRsp(c, -3, "no channel is set up")
		return
	}
	if err != nil {
		log.Warnf("notify: test: %s", err)
		rsp.ErrRsp(c, -2, "failed to send the test")
		return
	}
	rsp.OkRspWithData(c, results)
}
