package ipmi

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/proto"
)

// The web UI's IPMI page. These handlers answer under /api/ipmi with the web
// UI's response envelope, behind its session check and the admin role.

// Settings is what the page shows about the service.
type Settings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	// PowerLED is whether the power LED is wired, which power on, off,
	// cycle and soft off need.
	PowerLED bool      `json:"powerLed"`
	Users    []UserRow `json:"users"`
}

// UserRow is one KVM account as the page lists it.
type UserRow struct {
	Username string     `json:"username"`
	Role     authn.Role `json:"role"`
	Enabled  bool       `json:"enabled"`
	// HasPassword is whether the account has an IPMI password.
	HasPassword bool `json:"hasPassword"`
	// NameFits is whether the name is short enough for IPMI.
	NameFits bool `json:"nameFits"`
}

type setSettingsRequest struct {
	Enabled *bool `json:"enabled" form:"enabled" validate:"required"`
}

type setPasswordRequest struct {
	Password string `json:"password" form:"password" validate:"required"`
}

// The limits on an IPMI password. 20 bytes is the IPMI 2.0 maximum. The
// minimum is higher than the web login's because RAKP message 2 lets anyone
// who knows the user name try passwords offline, as fast as they can
// compute HMACs.
const (
	minPasswordLen = 12
	maxPasswordLen = 20
)

var (
	errPasswordLength  = errors.New("the IPMI password must be 12 to 20 characters")
	errPasswordChars   = errors.New("the IPMI password may only contain printable ASCII characters")
	errSameAsWeb       = errors.New("the IPMI password must differ from the account's web password")
	errNameTooLong     = errors.New("IPMI user names are at most 16 characters")
	errNoSuchAccount   = errors.New("no such account")
	errCannotSealIPMI  = errors.New("failed to store the IPMI password")
	errCannotSaveIPMI  = errors.New("failed to save the setting")
	errCannotListUsers = errors.New("failed to read the accounts")
)

func validatePassword(password string) error {
	if len(password) < minPasswordLen || len(password) > maxPasswordLen {
		return errPasswordLength
	}
	for i := 0; i < len(password); i++ {
		if password[i] < 0x20 || password[i] > 0x7e {
			return errPasswordChars
		}
	}
	return nil
}

func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response

	users, err := s.deps.Accounts.List()
	if err != nil {
		log.Errorf("ipmi: list accounts: %s", err)
		rsp.ErrRsp(c, -2, errCannotListUsers.Error())
		return
	}

	rows := make([]UserRow, 0, len(users))
	for _, info := range users {
		row := UserRow{
			Username: info.Username,
			Role:     info.Role,
			Enabled:  info.Enabled,
			NameFits: len(info.Username) <= maxUsernameLen,
		}
		if user, err := s.deps.Accounts.Get(info.Username); err == nil {
			row.HasPassword = user.IPMIPassword != ""
		}
		rows = append(rows, row)
	}

	rsp.OkRspWithData(c, &Settings{
		Enabled:  s.deps.Enabled(),
		Port:     Port,
		PowerLED: s.deps.PowerLEDConnected(),
		Users:    rows,
	})
}

// SetSettings turns the service on or off. On opens the socket before the
// setting is saved, so a port that cannot be had is reported and not
// remembered. Off closes it and ends every session.
func (s *Service) SetSettings(c *gin.Context) {
	var rsp proto.Response
	var req setSettingsRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if s.deps.SetEnabled == nil {
		rsp.ErrRsp(c, -2, "the setting cannot be changed")
		return
	}

	s.switchMu.Lock()
	defer s.switchMu.Unlock()

	on := *req.Enabled
	wasOpen := s.LocalAddr() != nil
	if on {
		if err := s.open(); err != nil {
			log.Errorf("ipmi: listen on %s: %s", s.deps.Addr, err)
			rsp.ErrRsp(c, -3, "failed to open UDP port 623: "+err.Error())
			return
		}
	}
	if err := s.deps.SetEnabled(on); err != nil {
		log.Errorf("ipmi: save the enabled setting: %s", err)
		if on && !wasOpen {
			s.Stop()
		}
		rsp.ErrRsp(c, -2, errCannotSaveIPMI.Error())
		return
	}
	if !on {
		s.Stop()
	}

	log.Infof("ipmi enabled: %t", on)
	rsp.OkRsp(c)
}

// SetPassword gives an account an IPMI password, or a new one, and ends its
// IPMI sessions.
func (s *Service) SetPassword(c *gin.Context) {
	var rsp proto.Response
	var req setPasswordRequest

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	username := c.Param("username")
	if err := s.setPassword(username, req.Password); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	log.Infof("ipmi: IPMI password of %q set", username)
	rsp.OkRsp(c)
}

func (s *Service) setPassword(username, password string) error {
	if len(username) > maxUsernameLen {
		return errNameTooLong
	}
	if _, err := s.deps.Accounts.Get(username); err != nil {
		return errNoSuchAccount
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	// The web password is a bcrypt hash, so the only way to tell is to try
	// it. RAKP exposes the IPMI password to offline guessing, and the web
	// password must not be what is exposed.
	if _, same, err := s.deps.Accounts.Authenticate(username, password); err == nil && same {
		return errSameAsWeb
	}

	sealed, err := s.deps.Keyring.Seal(username, password)
	if err != nil {
		log.Errorf("ipmi: seal the IPMI password of %q: %s", username, err)
		return errCannotSealIPMI
	}
	if _, err := s.deps.Accounts.SetIPMIPassword(username, sealed); err != nil {
		if errors.Is(err, authn.ErrUserNotFound) {
			return errNoSuchAccount
		}
		log.Errorf("ipmi: save the IPMI password of %q: %s", username, err)
		return errCannotSealIPMI
	}
	s.EndUserSessions(username)
	return nil
}

// ClearPassword removes an account's IPMI password, which ends its IPMI
// access and its sessions.
func (s *Service) ClearPassword(c *gin.Context) {
	var rsp proto.Response

	username := c.Param("username")
	if _, err := s.deps.Accounts.SetIPMIPassword(username, ""); err != nil {
		if errors.Is(err, authn.ErrUserNotFound) {
			rsp.ErrRsp(c, -1, errNoSuchAccount.Error())
			return
		}
		log.Errorf("ipmi: remove the IPMI password of %q: %s", username, err)
		rsp.ErrRsp(c, -2, "failed to remove the IPMI password")
		return
	}
	s.EndUserSessions(username)

	log.Infof("ipmi: IPMI password of %q removed", username)
	rsp.OkRsp(c)
}
