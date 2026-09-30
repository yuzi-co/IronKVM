//go:build linux

package netboot

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/extensions/addon"
)

type serviceFixture struct {
	svc      *Service
	settings *memSettings
	link     Link
	lan      lanNetwork
	lanErr   error
	calls    *[]string
}

func newServiceFixture(t *testing.T) *serviceFixture {
	t.Helper()
	scratchImage(t, true)

	f := &serviceFixture{
		settings: &memSettings{},
		link: Link{
			Mode:   "ncm",
			Prefix: netip.MustParsePrefix("172.31.255.0/30"),
			Board:  netip.MustParseAddr("172.31.255.1"),
			Host:   netip.MustParseAddr("172.31.255.2"),
			Active: true,
		},
		lan: lanNetwork{Interface: "eth0", Prefix: netip.MustParsePrefix("192.168.1.0/24")},
	}
	f.calls = recordScripts(t, nil)
	f.svc = New(Deps{
		Settings:     f.settings.get,
		SaveSettings: f.settings.save,
		Link:         func() Link { return f.link },
		LAN: func() (lanNetwork, error) {
			return f.lan, f.lanErr
		},
		ImageDir: t.TempDir(),
	})
	t.Cleanup(func() { f.svc.reconcileListener(config.NetBoot{}) })
	return f
}

func (f *serviceFixture) takeCalls() []string {
	calls := *f.calls
	*f.calls = nil
	return calls
}

func confExists(name string) bool {
	_, err := os.Stat(filepath.Join(ConfDir, name))
	return err == nil
}

func TestNothingTurnsOnWithoutTheAddon(t *testing.T) {
	f := newServiceFixture(t)

	for _, next := range []config.NetBoot{{USB: true}, {LAN: true}} {
		if err := f.svc.apply(next); !errors.Is(err, errNotInstalled) {
			t.Errorf("%+v without the add-on: %v", next, err)
		}
	}
	if f.settings.saves != 0 || confExists("usb.conf") || confExists("lan.conf") {
		t.Fatal("a refused change was saved or written")
	}
	if err := f.svc.apply(config.NetBoot{}); err != nil {
		t.Fatalf("turning off never needs the add-on: %v", err)
	}
}

// Turning the link side on writes its file and hands the link's DHCP server
// to dnsmasq through S03usbdev dhcp, which does not re-enumerate the gadget.
func TestTheLinkSideSwitchesTheLinksDHCPServer(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)

	if err := f.svc.apply(config.NetBoot{USB: true}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(ConfDir, "usb.conf"))
	if err != nil || string(b) != linkConf(StagedTFTPRoot) {
		t.Fatalf("usb.conf holds %q (%v)", b, err)
	}
	if got := f.takeCalls(); !slices.Equal(got, []string{"S03usbdev dhcp"}) {
		t.Fatalf("calls %q", got)
	}
	if !f.settings.current.USB {
		t.Fatal("the setting was not saved")
	}

	// The same settings again change nothing, so the host keeps its lease.
	if err := f.svc.apply(config.NetBoot{USB: true}); err != nil {
		t.Fatal(err)
	}
	if got := f.takeCalls(); len(got) != 0 {
		t.Fatalf("an unchanged save ran %q", got)
	}

	if err := f.svc.apply(config.NetBoot{}); err != nil {
		t.Fatal(err)
	}
	if confExists("usb.conf") {
		t.Fatal("usb.conf is left, so S85netboot would keep serving the link")
	}
	if got := f.takeCalls(); !slices.Equal(got, []string{"S03usbdev dhcp"}) {
		t.Fatalf("turning off: calls %q", got)
	}
}

// With the link off there is no DHCP server to switch. S03usbdev reads the file
// when the link comes on.
func TestTheLinkSideWithTheLinkOffOnlyWritesTheFile(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	f.link = Link{Mode: "off"}

	if err := f.svc.apply(config.NetBoot{USB: true}); err != nil {
		t.Fatal(err)
	}
	if !confExists("usb.conf") {
		t.Fatal("no usb.conf")
	}
	if got := f.takeCalls(); len(got) != 0 {
		t.Fatalf("calls %q", got)
	}
}

func TestTheLANSideRunsProxyDHCPAndStartsAtBoot(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)

	if err := f.svc.apply(config.NetBoot{LAN: true}); err != nil {
		t.Fatal(err)
	}
	want, _ := lanConf(f.lan, StagedTFTPRoot)
	if b, err := os.ReadFile(filepath.Join(ConfDir, "lan.conf")); err != nil || string(b) != want {
		t.Fatalf("lan.conf holds %q (%v)", b, err)
	}
	if got := f.takeCalls(); !slices.Equal(got, []string{"S85netboot restart"}) {
		t.Fatalf("calls %q", got)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir(AddonName), "enabled")); err != nil {
		t.Fatal("S04addons is not told to put S85netboot in /etc/init.d")
	}

	// The LAN moves: the file follows and dnsmasq restarts, at the next sync.
	f.lan.Prefix = netip.MustParsePrefix("10.1.0.0/16")
	f.svc.sync(false)
	b, _ := os.ReadFile(filepath.Join(ConfDir, "lan.conf"))
	if !strings.Contains(string(b), "dhcp-range=10.1.0.0,proxy,255.255.0.0\n") {
		t.Fatalf("lan.conf did not follow the LAN:\n%s", b)
	}
	if got := f.takeCalls(); !slices.Equal(got, []string{"S85netboot restart"}) {
		t.Fatalf("calls %q", got)
	}

	// The LAN goes away: the file and dnsmasq stay, and the page says why.
	f.lanErr = errors.New("the board has no default route")
	f.svc.sync(false)
	if got := f.takeCalls(); len(got) != 0 {
		t.Fatalf("a lost LAN ran %q", got)
	}
	if !confExists("lan.conf") || f.svc.status().LANError == "" {
		t.Fatal("a lost LAN removed the file or said nothing")
	}
	f.lanErr = nil

	if err := f.svc.apply(config.NetBoot{}); err != nil {
		t.Fatal(err)
	}
	if confExists("lan.conf") {
		t.Fatal("lan.conf is left")
	}
	if got := f.takeCalls(); !slices.Equal(got, []string{"S85netboot stop"}) {
		t.Fatalf("turning off: calls %q", got)
	}
	if _, err := os.Stat(filepath.Join(addon.Dir(AddonName), "enabled")); err == nil {
		t.Fatal("S85netboot would still start at boot")
	}
}

func TestTheLANSideNeedsALAN(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	f.lanErr = errors.New("the board has no default route")

	if err := f.svc.apply(config.NetBoot{LAN: true}); err == nil {
		t.Fatal("proxy DHCP was turned on with no LAN")
	}
	if f.settings.saves != 0 || confExists("lan.conf") {
		t.Fatal("the refused change was saved or written")
	}
}

func TestAFailedSaveChangesNothing(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	f.settings.fail = errors.New("read-only")

	if err := f.svc.apply(config.NetBoot{USB: true, LAN: true}); err == nil {
		t.Fatal("a failed save was reported as done")
	}
	if confExists("usb.conf") || confExists("lan.conf") || len(f.takeCalls()) != 0 {
		t.Fatal("a change that was not saved was applied")
	}
}

// The menu is served on the link's address while the link side is on, and
// the listener follows the address when the subnet changes.
func TestTheListenerFollowsTheLink(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	f.link.Board = netip.MustParseAddr("127.0.0.1")
	f.link.Prefix = netip.MustParsePrefix("127.0.0.0/8")

	if err := f.svc.apply(config.NetBoot{USB: true}); err != nil {
		t.Fatal(err)
	}
	menuURL := fmt.Sprintf("http://127.0.0.1:%d/menu.ipxe", HTTPPort)
	if got := f.svc.status().MenuURL; got != menuURL {
		t.Fatalf("menu URL %q", got)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	rsp, err := client.Get(menuURL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK || !strings.Contains(string(body), "set base http://127.0.0.1:8069") {
		t.Fatalf("status %d:\n%s", rsp.StatusCode, body)
	}

	// A new subnet the board does not hold yet: the old listener goes and
	// the new one waits for the address.
	f.link.Board = netip.MustParseAddr("172.31.254.1")
	f.link.Prefix = netip.MustParsePrefix("172.31.254.0/30")
	f.svc.sync(false)
	if got := f.svc.status().MenuURL; got != "" {
		t.Fatalf("still serving on %q", got)
	}
	if _, err := client.Get(menuURL); err == nil {
		t.Fatal("the old address still answers")
	}

	f.link.Board = netip.MustParseAddr("127.0.0.1")
	f.link.Prefix = netip.MustParsePrefix("127.0.0.0/8")
	f.svc.sync(false)
	if err := f.svc.apply(config.NetBoot{}); err != nil {
		t.Fatal(err)
	}
	if got := f.svc.status().MenuURL; got != "" {
		t.Fatalf("serving with the link side off: %q", got)
	}
}

func TestUninstallGivesTheLinkBackAndRemovesTheAddon(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	if err := f.svc.apply(config.NetBoot{USB: true, LAN: true}); err != nil {
		t.Fatal(err)
	}
	f.takeCalls()

	if err := f.svc.uninstall(); err != nil {
		t.Fatal(err)
	}
	if got := f.takeCalls(); !slices.Equal(got, []string{"S85netboot stop", "S03usbdev dhcp"}) {
		t.Fatalf("calls %q", got)
	}
	if f.settings.current.USB || f.settings.current.LAN {
		t.Fatal("the settings are left on")
	}
	if _, err := os.Stat(addon.Dir(AddonName)); err == nil {
		t.Fatal("the add-on is left on /data")
	}
	if confExists("usb.conf") || confExists("lan.conf") {
		t.Fatal("a configuration file is left")
	}
}

func TestInstallAppliesTheSettingsItFinds(t *testing.T) {
	f := newServiceFixture(t)
	f.settings.current = config.NetBoot{USB: true}

	saved := installAddon
	t.Cleanup(func() { installAddon = saved })
	installAddon = func() error {
		fakeInstall(t)
		return nil
	}

	if err := f.svc.install(); err != nil {
		t.Fatal(err)
	}
	if !confExists("usb.conf") {
		t.Fatal("a reinstall over settings left on did not put usb.conf back")
	}
}

// A change while an install runs is refused rather than queued behind it.
func TestChangesWaitForNothing(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)

	f.svc.busy.Lock()
	defer f.svc.busy.Unlock()
	if err := f.svc.apply(config.NetBoot{USB: true}); !errors.Is(err, errBusy) {
		t.Fatalf("apply while busy: %v", err)
	}
	if err := f.svc.install(); !errors.Is(err, errBusy) {
		t.Fatalf("install while busy: %v", err)
	}
}

// countDeps wraps the fixture's Settings and Link and counts calls to either.
func (f *serviceFixture) countDeps() *int {
	calls := 0
	settings, link := f.svc.deps.Settings, f.svc.deps.Link
	f.svc.deps.Settings = func() config.NetBoot { calls++; return settings() }
	f.svc.deps.Link = func() Link { calls++; return link() }
	return &calls
}

// Without the add-on, the first sync cleans up and later ones do nothing: no
// settings read, no link read, no file removed.
func TestSyncWithoutTheAddonIdlesAfterCleaningUp(t *testing.T) {
	f := newServiceFixture(t)
	calls := f.countDeps()

	// A file left behind by an earlier install is removed once.
	if err := os.MkdirAll(ConfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ConfDir, "usb.conf"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	f.svc.sync(true)
	if confExists("usb.conf") {
		t.Fatal("the first sync left usb.conf behind")
	}
	if *calls == 0 {
		t.Fatal("the first sync did not look at the settings or the link")
	}

	*calls = 0
	f.takeCalls()
	f.svc.sync(false)
	f.svc.sync(false)
	if *calls != 0 {
		t.Fatalf("idle syncs made %d dependency calls, want 0", *calls)
	}
	if got := f.takeCalls(); len(got) != 0 {
		t.Fatalf("idle syncs ran scripts: %v", got)
	}
}

// An install that appears after the service went idle is picked up by the
// next sync, and a settings change always gets a full sync after it.
func TestSyncLeavesIdleWhenTheAddonAppears(t *testing.T) {
	f := newServiceFixture(t)
	f.link.Mode = "off"
	f.svc.sync(true)
	if !f.svc.idle {
		t.Fatal("the service did not go idle without the add-on")
	}

	fakeInstall(t)
	f.settings.current = config.NetBoot{USB: true}
	f.svc.sync(false)
	if !confExists("usb.conf") {
		t.Fatal("the sync after an install did not write usb.conf")
	}
	if f.svc.idle {
		t.Fatal("the service stayed idle with the add-on installed")
	}
}
