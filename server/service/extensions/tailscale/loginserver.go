package tailscale

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// DefaultLoginServer is Tailscale's own control server, which tailscaled
// uses when no other is set.
const DefaultLoginServer = "https://controlplane.tailscale.com"

// defaultSynonyms are the names tailscaled treats as its own control server
// (ipn.IsLoginServerSynonym). An empty ControlURL is a daemon that has never
// been told one, which is the same server.
var defaultSynonyms = map[string]bool{
	"":                                    true,
	DefaultLoginServer:                    true,
	"https://login.tailscale.com":         true,
	"https://controlplane.tailscale.com/": true,
	"https://login.tailscale.com/":        true,
}

// IsDefaultLoginServer reports whether u is Tailscale's own control server.
func IsDefaultLoginServer(u string) bool {
	return defaultSynonyms[strings.TrimSpace(u)]
}

// ErrLoginServer is the page's answer for a value that is not a control
// server's address.
var ErrLoginServer = errors.New("the login server must be an https URL, for example https://headscale.example.com")

// NormalizeLoginServer checks a login server the page sent and returns it as
// it is given to tailscale. Empty means Tailscale's own server. Only https
// is accepted: the node's keys and its traffic policy come from this server,
// and a plain http one hands both to anyone on the path.
func NormalizeLoginServer(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return DefaultLoginServer, nil
	}
	if len(s) > 512 || strings.ContainsAny(s, " \t\r\n") {
		return "", ErrLoginServer
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.Opaque != "" || u.User != nil ||
		u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", ErrLoginServer
	}
	host := u.Hostname()
	if host == "" || strings.HasPrefix(host, "-") {
		return "", ErrLoginServer
	}
	if port := u.Port(); port != "" {
		if _, err := net.LookupPort("tcp", port); err != nil {
			return "", ErrLoginServer
		}
	}
	return strings.TrimRight(u.String(), "/"), nil
}

// loginServerArgs is what up and login add so that tailscale keeps a custom
// login server. Both refuse to run when a setting that differs from its
// default goes unmentioned ("changing settings via 'tailscale up' requires
// mentioning all non-default flags"), so a node on Headscale could not be
// brought up by the page at all without this. Nothing is added for
// Tailscale's own server, so the command stays what it always was there.
func loginServerArgs(current string) []string {
	if IsDefaultLoginServer(current) {
		return nil
	}
	return []string{"--login-server=" + current}
}

// prefs is the part of `tailscale debug prefs` this reads.
type prefs struct {
	ControlURL string `json:"ControlURL"`
}

// LoginServer is the control server tailscaled uses now: its ControlURL
// preference.
func (c *Cli) LoginServer() (string, error) {
	out, err := vpn.Run(exec.Command(TailscalePath, "debug", "prefs"))
	if err != nil {
		return "", err
	}
	i := bytes.IndexByte(out, '{')
	if i < 0 {
		return "", errors.New("no preferences in the output")
	}
	var p prefs
	if err := json.Unmarshal(out[i:], &p); err != nil {
		return "", err
	}
	return p.ControlURL, nil
}

// SetHasLoginServer reports whether the installed CLI's `tailscale set`
// takes --login-server. The releases this was written against (1.9x) do not:
// the login server is an up and login flag only, and changing it means a new
// login.
func (c *Cli) SetHasLoginServer() bool {
	out, _ := exec.Command(TailscalePath, "set", "--help").CombinedOutput()
	return bytes.Contains(out, []byte("-login-server"))
}

// SetLoginServer changes the server with `tailscale set`, where it can.
func (c *Cli) SetLoginServer(server string) error {
	_, err := vpn.Run(exec.Command(TailscalePath, "set", "--login-server="+server))
	return err
}

// LoginWith starts a login against server and returns the URL the user has
// to visit. Starting it already stores server as the node's ControlURL.
func (c *Cli) LoginWith(server string, timeout, life time.Duration) (string, error) {
	cmd := exec.Command(TailscalePath, "login", "--accept-dns=false", "--timeout=10m", "--login-server="+server)
	return vpn.LoginURL(cmd, false, timeout, life)
}

// The page asks for the status every ten seconds, and each answer would cost
// another run of the CLI for a value that only changes through this file. So
// it is kept, for a minute at most, and forgotten on every change.
const loginServerTTL = time.Minute

var loginServerCache struct {
	sync.Mutex
	value string
	at    time.Time
}

func cachedLoginServer(cli *Cli) string {
	loginServerCache.Lock()
	defer loginServerCache.Unlock()

	if !loginServerCache.at.IsZero() && time.Since(loginServerCache.at) < loginServerTTL {
		return loginServerCache.value
	}
	v, err := cli.LoginServer()
	if err != nil {
		log.Debugf("failed to read the tailscale login server: %s", err)
		return loginServerCache.value
	}
	loginServerCache.value, loginServerCache.at = v, time.Now()
	return v
}

func forgetLoginServer() {
	loginServerCache.Lock()
	loginServerCache.at = time.Time{}
	loginServerCache.Unlock()
}

// Variables so the tests can shorten them.
var (
	// applyURLTimeout bounds the wait for the login URL from the new server.
	// A server that does not answer in this time is a failed change.
	applyURLTimeout = 30 * time.Second

	// rollbackLife ends the login that puts the old server back once it has
	// printed its URL: nobody is going to visit it.
	rollbackLife = 2 * time.Second
)

// ErrLoggedIn refuses a change that would end the node's current login.
var ErrLoggedIn = errors.New("log out of Tailscale first: changing the login server needs a new login, and this device is logged in now")

// applyLoginServer moves the node to server, and puts the old one back when
// that fails. It returns the URL to log in at when the change started one.
//
// With a CLI whose set takes --login-server the change is that one command.
// Without it, the only way is a login against the new server, so the change
// is refused while the node is logged in: it would end the login the operator
// may be using to reach this page. A logged-out node starts the login, and a
// server that prints no login URL in applyURLTimeout is a failed change.
//
// A failed change is checked against what tailscaled holds afterwards, and
// put back with the same command when it moved, because the login command
// stores the new server before it learns whether the server answers.
func applyLoginServer(cli *Cli, state, old, server string) (string, error) {
	viaSet := cli.SetHasLoginServer()

	if !viaSet && (state == "Running" || state == "Stopped") {
		return "", ErrLoggedIn
	}

	var loginURL string
	var err error
	if viaSet {
		err = cli.SetLoginServer(server)
	} else {
		loginURL, err = cli.LoginWith(server, applyURLTimeout, 0)
	}
	if err == nil {
		now, readErr := cli.LoginServer()
		switch {
		case readErr != nil:
			err = readErr
		case strings.TrimRight(now, "/") != server:
			err = fmt.Errorf("tailscale still uses %q", now)
		default:
			return loginURL, nil
		}
	}

	// Put the old one back, when the failed command moved it. A node that
	// never had one gets Tailscale's.
	back := old
	if back == "" {
		back = DefaultLoginServer
	}
	var rbErr error
	if now, readErr := cli.LoginServer(); readErr != nil || now != old {
		if viaSet {
			rbErr = cli.SetLoginServer(back)
		} else {
			_, rbErr = cli.LoginWith(back, applyURLTimeout, rollbackLife)
		}
	}
	if rbErr == nil {
		if now, readErr := cli.LoginServer(); readErr != nil {
			rbErr = readErr
		} else if now != old && !(IsDefaultLoginServer(now) && IsDefaultLoginServer(old)) {
			rbErr = fmt.Errorf("tailscale uses %q", now)
		}
	}
	if rbErr != nil {
		log.Errorf("tailscale: the login server change failed and the old one could not be put back: %s", rbErr)
		return "", &changeError{cause: err, note: "putting " + back + " back also failed: " + rbErr.Error()}
	}
	return "", &changeError{cause: err, note: back + " is in use again"}
}

// SetLoginServer handles POST login-server: {loginServer}, empty for
// Tailscale's own server. The answer carries the server now in use and, when
// the change started a login, the URL to finish it at.
func (s *Service) SetLoginServer(c *gin.Context) {
	var req proto.TailscaleLoginServerReq
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	server, err := NormalizeLoginServer(req.LoginServer)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	if !isInstalled() {
		rsp.ErrRsp(c, -1, "tailscale is not installed")
		return
	}

	defer addon.LockVPN()()
	if vpn.Refuse(c, addon.Tailscale) {
		return
	}
	defer forgetLoginServer()

	cli := NewCli()
	status, err := cli.Status()
	if err != nil {
		if err := cli.Start(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
			return
		}
		status, err = cli.Status()
	}
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("unknown status", err))
		return
	}

	old, err := cli.LoginServer()
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("reading the login server failed", err))
		return
	}
	if strings.TrimRight(old, "/") == server || (IsDefaultLoginServer(old) && IsDefaultLoginServer(server)) {
		rsp.OkRspWithData(c, &proto.TailscaleLoginServerRsp{LoginServer: server})
		return
	}

	loginURL, err := applyLoginServer(cli, status.BackendState, old, server)
	if errors.Is(err, ErrLoggedIn) {
		rsp.ErrRsp(c, -3, err.Error())
		return
	}
	var ce *changeError
	if errors.As(err, &ce) {
		rsp.ErrRsp(c, -2, vpn.Message("changing the login server failed", ce.cause)+"\n"+ce.note)
		return
	}
	if err != nil {
		rsp.ErrRsp(c, -2, vpn.Message("changing the login server failed", err))
		return
	}

	log.Infof("tailscale: login server is now %s", server)
	rsp.OkRspWithData(c, &proto.TailscaleLoginServerRsp{LoginServer: server, Url: loginURL})
}

// changeError is a failed change and what became of the old server after it.
// The two are kept apart because vpn.Message replaces an error's text with the
// command's last lines, and the page needs both.
type changeError struct {
	cause error
	note  string
}

func (e *changeError) Error() string { return e.cause.Error() + "; " + e.note }

func (e *changeError) Unwrap() error { return e.cause }
