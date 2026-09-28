// Package netboot boots the host from the network: PXE, iPXE and a menu of
// the images on /data over the USB network link, and proxy DHCP with
// netboot.xyz on the LAN.
//
// The server writes dnsmasq's two configuration files and serves the menu
// and the images; S85netboot runs dnsmasq, and S03usbdev decides whether
// dnsmasq or its udhcpd serves the link. See
// docs/superpowers/specs/2026-09-28-network-boot-design.md.
package netboot

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
)

// Paths shared with the scripts, and variables so the tests can point them at
// a scratch tree. S85netboot and S03usbdev hold the same defaults, and a test
// holds the copies together.
var (
	// ConfDir is where the two dnsmasq configuration files go. /etc/kvm is
	// on /data, so the link comes up with network boot at the next boot,
	// before the server runs.
	ConfDir = "/etc/kvm/netboot"
	// RunDir is where S85netboot keeps the pid files, the link's leases
	// and both logs.
	RunDir = "/tmp/netboot"
	// USBScript rebuilds the link's DHCP server with its dhcp action.
	USBScript = "/etc/init.d/S03usbdev"
	// runScript runs a script with its arguments. Tests replace it.
	runScript = func(script string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), scriptTimeout)
		defer cancel()
		_, err := vpn.Run(exec.CommandContext(ctx, "sh", append([]string{script}, args...)...))
		return err
	}
)

const (
	scriptTimeout = 60 * time.Second
	// reconcileEvery is how often the listener and the LAN side are checked
	// against the link and the LAN. Both change without telling this
	// package: the operator changes the link's subnet on another page, and
	// the LAN's DHCP server can move the board.
	reconcileEvery = 10 * time.Second
)

var (
	errNotInstalled = errors.New("install network boot first: it needs dnsmasq and the boot files")
	errBusy         = errors.New("network boot is busy with another change")
)

// Link is the USB network link as network boot needs it.
type Link struct {
	Mode   string // off, ncm, ecm or rndis
	Prefix netip.Prefix
	Board  netip.Addr
	Host   netip.Addr
	Active bool
}

func (l Link) on() bool { return l.Mode != "" && l.Mode != "off" }

// Deps is what the service needs from the rest of the server.
type Deps struct {
	Settings     func() config.NetBoot
	SaveSettings func(config.NetBoot) error
	Link         func() Link
	LAN          func() (lanNetwork, error)
	ImageDir     string
}

// Service owns network boot: the configuration files, the link's HTTP server
// and the add-on.
type Service struct {
	deps  Deps
	boots bootLog

	// busy keeps install, uninstall and a settings change apart. They share
	// the files, and an install takes minutes.
	busy sync.Mutex

	mu       sync.Mutex
	server   *http.Server
	serving  netip.AddrPort
	subnet   netip.Prefix
	lanError string
}

func New(deps Deps) *Service {
	if deps.LAN == nil {
		deps.LAN = currentLAN
	}
	return &Service{deps: deps}
}

// Start brings the files in line with the settings, and keeps the link's
// HTTP server and the LAN side in line with the link and the LAN.
func (s *Service) Start() {
	go func() {
		s.sync(true)
		ticker := time.NewTicker(reconcileEvery)
		defer ticker.Stop()
		for range ticker.C {
			s.sync(false)
		}
	}()
}

// sync rewrites the files from the settings, restarts the side whose file
// changed, and moves the listener. At the server's start it also starts a
// LAN side that is on and not running: S85netboot ran it at boot, but a
// failed start there is not retried by anything else.
func (s *Service) sync(starting bool) {
	if !s.busy.TryLock() {
		return
	}
	defer s.busy.Unlock()

	settings := s.deps.Settings()
	if !installed() {
		settings = config.NetBoot{}
	}

	usbChanged, lanChanged, err := s.writeConfs(settings)
	if err != nil {
		log.Errorf("netboot: %s", err)
	}
	if usbChanged && s.deps.Link().on() {
		if err := runScript(USBScript, "dhcp"); err != nil {
			log.Errorf("netboot: restart the link's DHCP server: %s", err)
		}
	}
	if lanChanged || (starting && settings.LAN && !running("lan")) {
		s.restartLAN(settings.LAN)
	}

	s.reconcileListener(settings)
}

// writeConfs writes the file of each side that is on and removes the file of
// each side that is off, and says which changed. The files are S85netboot's
// switches: it serves a side only while its file exists.
func (s *Service) writeConfs(settings config.NetBoot) (usbChanged, lanChanged bool, err error) {
	usbFile := filepath.Join(ConfDir, "usb.conf")
	lanFile := filepath.Join(ConfDir, "lan.conf")

	if settings.USB {
		usbChanged, err = replaceFile(usbFile, linkConf(tftpRoot()))
	} else {
		usbChanged, err = removeFile(usbFile)
	}
	if err != nil {
		return usbChanged, false, err
	}

	lanError := ""
	if settings.LAN {
		var lan lanNetwork
		var conf string
		lan, err = s.deps.LAN()
		if err == nil {
			conf, err = lanConf(lan, tftpRoot())
		}
		if err == nil {
			lanChanged, err = replaceFile(lanFile, conf)
		} else {
			// The LAN is gone for now: a cable out, or no lease yet.
			// The old file stays, as does the old instance, until the
			// LAN comes back; the page says why.
			lanError = err.Error()
			err = nil
		}
	} else {
		lanChanged, err = removeFile(lanFile)
	}

	s.mu.Lock()
	s.lanError = lanError
	s.mu.Unlock()

	return usbChanged, lanChanged, err
}

func (s *Service) restartLAN(on bool) {
	action := "stop"
	if on {
		action = "restart"
	}
	if err := runScript(filepath.Join(addon.PkgInitdDir, Initd), action); err != nil {
		log.Errorf("netboot: %s the LAN side: %s", action, err)
	}
}

// apply saves new settings and makes them so. Turning a side on needs the
// add-on; turning one off never does.
func (s *Service) apply(next config.NetBoot) error {
	if !s.busy.TryLock() {
		return errBusy
	}
	defer s.busy.Unlock()

	return s.applyLocked(next)
}

func (s *Service) applyLocked(next config.NetBoot) error {
	if (next.USB || next.LAN) && !installed() {
		return errNotInstalled
	}
	if next.LAN {
		if _, err := s.deps.LAN(); err != nil {
			return fmt.Errorf("proxy DHCP needs the LAN: %w", err)
		}
	}

	prev := s.deps.Settings()
	if err := s.deps.SaveSettings(next); err != nil {
		return err
	}

	usbChanged, lanChanged, err := s.writeConfs(next)
	if err != nil {
		return err
	}

	if addon.OnData() && installed() {
		if err := addon.SetEnabled(AddonName, next.LAN); err != nil {
			log.Errorf("netboot: record the LAN side at boot: %s", err)
		}
	}

	if lanChanged || prev.LAN != next.LAN {
		s.restartLAN(next.LAN)
	}
	// The link's DHCP server changes hands without a new enumeration, so
	// the host keeps its keyboard. With the link off, S03usbdev reads the
	// file at its next start.
	if usbChanged && s.deps.Link().on() {
		if err := runScript(USBScript, "dhcp"); err != nil {
			return errors.New(vpn.Message("the link's DHCP server did not restart", err))
		}
	}

	s.reconcileListener(next)
	return nil
}

// reconcileListener serves the link while network boot is on for it and the
// board holds its address there, and follows a change of subnet. A listen that
// fails because the address is not up yet is tried again at the next sync.
func (s *Service) reconcileListener(settings config.NetBoot) {
	var want netip.AddrPort
	link := s.deps.Link()
	if settings.USB && link.on() && link.Board.IsValid() {
		want = netip.AddrPortFrom(link.Board, HTTPPort)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serving == want && s.subnet == link.Prefix {
		return
	}

	if s.server != nil {
		_ = s.server.Close()
		s.server = nil
		s.serving = netip.AddrPort{}
	}
	if !want.IsValid() {
		return
	}

	ln, err := net.Listen("tcp4", want.String())
	if err != nil {
		log.Debugf("netboot: not serving the link yet: %s", err)
		return
	}

	handler := &linkServer{
		subnet:   link.Prefix,
		base:     fmt.Sprintf("http://%s", want),
		imageDir: s.deps.ImageDir,
		tftpRoot: tftpRoot(),
		images:   func() []string { return listImages(s.deps.ImageDir) },
		boots:    &s.boots,
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("netboot: the link's HTTP server stopped: %s", err)
		}
	}()

	s.server = server
	s.serving = want
	s.subnet = link.Prefix
	log.Infof("netboot: serving the menu on http://%s/menu.ipxe", want)
}

// running says whether the instance named usb or lan runs now.
func running(name string) bool {
	return addon.Running(addon.Daemon{
		PidFile: filepath.Join(RunDir, name+".pid"),
		Process: "dnsmasq",
	})
}

// install and uninstall hold busy for their whole length. An uninstall turns
// both sides off first, so the link goes back to udhcpd before dnsmasq goes.
func (s *Service) install() error {
	if !s.busy.TryLock() {
		return errBusy
	}
	defer s.busy.Unlock()

	if err := installAddon(); err != nil {
		return err
	}

	// A reinstall over settings that were left on puts the files back.
	return s.applyLocked(s.deps.Settings())
}

func (s *Service) uninstall() error {
	if !s.busy.TryLock() {
		return errBusy
	}
	defer s.busy.Unlock()

	if err := s.applyLocked(config.NetBoot{}); err != nil {
		return err
	}
	if running("usb") || running("lan") {
		return errors.New("dnsmasq is still running, so it is not removed")
	}
	return uninstallAddon()
}

// The add-on's own install and uninstall, variables for the tests.
var (
	installAddon   = install
	uninstallAddon = uninstall
)

// Status is what the Network boot page shows.
type Status struct {
	USB bool `json:"usb"`
	LAN bool `json:"lan"`

	// OnData is whether the add-on can be installed at all.
	OnData    bool   `json:"onData"`
	Installed bool   `json:"installed"`
	Version   string `json:"version"`

	Link struct {
		Mode   string `json:"mode"`
		Board  string `json:"board"`
		Host   string `json:"host"`
		Active bool   `json:"active"`
	} `json:"link"`
	USBRunning bool   `json:"usbRunning"`
	MenuURL    string `json:"menuUrl"`

	LANRunning   bool   `json:"lanRunning"`
	LANInterface string `json:"lanInterface"`
	LANNetwork   string `json:"lanNetwork"`
	LANError     string `json:"lanError"`

	Images []string `json:"images"`
	Leases []Lease  `json:"leases"`
	Boots  []Boot   `json:"boots"`
	USBLog []string `json:"usbLog"`
	LANLog []string `json:"lanLog"`
}

const statusLogLines = 10

func (s *Service) status() Status {
	var st Status

	settings := s.deps.Settings()
	st.USB = settings.USB
	st.LAN = settings.LAN
	st.OnData = addon.OnData()
	st.Installed = installed()
	if v, err := os.ReadFile(versionPath()); err == nil {
		st.Version = strings.TrimSpace(string(v))
	}

	link := s.deps.Link()
	st.Link.Mode = link.Mode
	if link.Board.IsValid() {
		st.Link.Board = link.Board.String()
	}
	if link.Host.IsValid() {
		st.Link.Host = link.Host.String()
	}
	st.Link.Active = link.Active
	st.USBRunning = running("usb")
	st.LANRunning = running("lan")

	if lan, err := s.deps.LAN(); err == nil {
		st.LANInterface = lan.Interface
		st.LANNetwork = lan.Prefix.String()
	}

	s.mu.Lock()
	if s.serving.IsValid() {
		st.MenuURL = fmt.Sprintf("http://%s/menu.ipxe", s.serving)
	}
	st.LANError = s.lanError
	s.mu.Unlock()

	st.Images = listImages(s.deps.ImageDir)
	if content, err := os.ReadFile(filepath.Join(RunDir, "usb.leases")); err == nil {
		st.Leases = parseLeases(content)
	}
	st.Boots = s.boots.list()
	st.USBLog = tailLines(filepath.Join(RunDir, "usb.log"), statusLogLines)
	st.LANLog = tailLines(filepath.Join(RunDir, "lan.log"), statusLogLines)

	return st
}
