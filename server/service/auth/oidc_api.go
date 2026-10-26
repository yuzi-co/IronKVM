package auth

import (
	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// oidcInfo is what the login page learns before anyone signs in.
type oidcInfo struct {
	Enabled           bool   `json:"enabled"`
	ProviderName      string `json:"providerName,omitempty"`
	HidePasswordLogin bool   `json:"hidePasswordLogin,omitempty"`
}

// oidcSettingsView is the settings as the API shows them: the client secret
// is never in it, only whether one is set.
type oidcSettingsView struct {
	Enabled           bool       `json:"enabled"`
	ProviderName      string     `json:"providerName"`
	Issuer            string     `json:"issuer"`
	ClientID          string     `json:"clientId"`
	ClientSecretSet   bool       `json:"clientSecretSet"`
	Scopes            []string   `json:"scopes"`
	UsernameClaim     string     `json:"usernameClaim"`
	EmailDomain       string     `json:"emailDomain"`
	AutoCreate        bool       `json:"autoCreate"`
	DefaultRole       authn.Role `json:"defaultRole"`
	GroupsClaim       string     `json:"groupsClaim"`
	AdminGroup        string     `json:"adminGroup"`
	UserGroup         string     `json:"userGroup"`
	ViewerGroup       string     `json:"viewerGroup"`
	HidePasswordLogin bool       `json:"hidePasswordLogin"`
	// RedirectURL is the callback address to register at the IdP, as this
	// request reached the board.
	RedirectURL string `json:"redirectUrl"`
}

// oidcSettingsReq is a save. An empty ClientSecret keeps the one on file;
// ClearClientSecret removes it, for a public client.
type oidcSettingsReq struct {
	Enabled           bool       `json:"enabled"`
	ProviderName      string     `json:"providerName"`
	Issuer            string     `json:"issuer"`
	ClientID          string     `json:"clientId"`
	ClientSecret      string     `json:"clientSecret"`
	ClearClientSecret bool       `json:"clearClientSecret"`
	Scopes            []string   `json:"scopes"`
	UsernameClaim     string     `json:"usernameClaim"`
	EmailDomain       string     `json:"emailDomain"`
	AutoCreate        bool       `json:"autoCreate"`
	DefaultRole       authn.Role `json:"defaultRole"`
	GroupsClaim       string     `json:"groupsClaim"`
	AdminGroup        string     `json:"adminGroup"`
	UserGroup         string     `json:"userGroup"`
	ViewerGroup       string     `json:"viewerGroup"`
	HidePasswordLogin bool       `json:"hidePasswordLogin"`
}

func viewOf(s *oidcSettings, c *gin.Context) oidcSettingsView {
	return oidcSettingsView{
		Enabled:           s.Enabled,
		ProviderName:      s.ProviderName,
		Issuer:            s.Issuer,
		ClientID:          s.ClientID,
		ClientSecretSet:   s.ClientSecret != "",
		Scopes:            s.Scopes,
		UsernameClaim:     s.UsernameClaim,
		EmailDomain:       s.EmailDomain,
		AutoCreate:        s.AutoCreate,
		DefaultRole:       s.DefaultRole,
		GroupsClaim:       s.GroupsClaim,
		AdminGroup:        s.AdminGroup,
		UserGroup:         s.UserGroup,
		ViewerGroup:       s.ViewerGroup,
		HidePasswordLogin: s.HidePasswordLogin,
		RedirectURL:       redirectURL(c),
	}
}

func (s *Service) GetOIDCSettings(c *gin.Context) {
	var rsp proto.Response
	settings, err := loadOIDCSettings()
	if err != nil {
		log.Errorf("read oidc settings: %s", err)
		rsp.ErrRsp(c, -1, "failed to read the OIDC settings")
		return
	}
	rsp.OkRspWithData(c, viewOf(&settings, c))
}

func (s *Service) SetOIDCSettings(c *gin.Context) {
	var rsp proto.Response
	var req oidcSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}

	oidcSettingsMu.Lock()
	defer oidcSettingsMu.Unlock()
	current, err := loadOIDCSettingsLocked()
	if err != nil {
		// A damaged file is replaced by the save, but its secret is lost.
		log.Warnf("read oidc settings before save: %s", err)
	}
	next := oidcSettings{
		Enabled:           req.Enabled,
		ProviderName:      req.ProviderName,
		Issuer:            req.Issuer,
		ClientID:          req.ClientID,
		ClientSecret:      current.ClientSecret,
		Scopes:            req.Scopes,
		UsernameClaim:     req.UsernameClaim,
		EmailDomain:       req.EmailDomain,
		AutoCreate:        req.AutoCreate,
		DefaultRole:       req.DefaultRole,
		GroupsClaim:       req.GroupsClaim,
		AdminGroup:        req.AdminGroup,
		UserGroup:         req.UserGroup,
		ViewerGroup:       req.ViewerGroup,
		HidePasswordLogin: req.HidePasswordLogin,
	}
	switch {
	case req.ClientSecret != "":
		next.ClientSecret = req.ClientSecret
	case req.ClearClientSecret:
		next.ClientSecret = ""
	}
	if err = next.normalize(); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	if err = saveOIDCSettingsLocked(next); err != nil {
		log.Errorf("save oidc settings: %s", err)
		rsp.ErrRsp(c, -3, "failed to save the OIDC settings")
		return
	}
	// Flows begun under the old settings would finish against the new ones.
	resetOIDCState()

	principal, _ := middleware.CurrentPrincipal(c)
	log.Infof("oidc settings changed by %s: enabled=%v issuer=%s", principal.Username, next.Enabled, next.Issuer)
	rsp.OkRspWithData(c, viewOf(&next, c))
}
