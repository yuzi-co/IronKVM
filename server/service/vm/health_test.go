package vm

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/extensions/addon"
)

type healthFixture struct {
	base  string
	vpnA  addon.Daemon
	vpnB  addon.Daemon
	space struct {
		total, available uint64
		err              error
	}
}

// useHealth points every reading GetHealth takes at a scratch tree: the
// temperature file, the statfs answer, and two VPN daemons with their boot
// scripts and pid files.
func useHealth(t *testing.T) *healthFixture {
	t.Helper()

	f := &healthFixture{base: t.TempDir()}
	for _, dir := range []string{"init.d", "proc", "run"} {
		if err := os.MkdirAll(filepath.Join(f.base, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	savedTemp, savedStat, savedVpns := cpuTempPath, healthStatFS, healthVpns
	savedInitd, savedProc, savedMarker := addon.InitdDir, addon.ProcDir, addon.DistroMarker
	t.Cleanup(func() {
		cpuTempPath, healthStatFS, healthVpns = savedTemp, savedStat, savedVpns
		addon.InitdDir, addon.ProcDir, addon.DistroMarker = savedInitd, savedProc, savedMarker
	})

	cpuTempPath = filepath.Join(f.base, "temp")
	addon.InitdDir = filepath.Join(f.base, "init.d")
	addon.ProcDir = filepath.Join(f.base, "proc")
	// Off a distribution image, so start at boot is the script in InitdDir.
	addon.DistroMarker = filepath.Join(f.base, "no-deviceinfo")

	f.vpnA = addon.Daemon{Name: "a", Title: "VPN A", Initd: "S98a",
		PidFile: filepath.Join(f.base, "run", "a.pid"), Process: "a-daemon"}
	f.vpnB = addon.Daemon{Name: "b", Title: "VPN B", Initd: "S98b",
		PidFile: filepath.Join(f.base, "run", "b.pid"), Process: "b-daemon"}
	healthVpns = func() []addon.Daemon { return []addon.Daemon{f.vpnA, f.vpnB} }

	f.space.err = errors.New("not mounted")
	healthStatFS = func(string) (uint64, uint64, error) {
		return f.space.total, f.space.available, f.space.err
	}

	return f
}

func (f *healthFixture) bootEnabled(t *testing.T, d addon.Daemon) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(addon.InitdDir, d.Initd), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func (f *healthFixture) running(t *testing.T, d addon.Daemon, pid int) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(addon.ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte("/usr/bin/"+d.Process+"\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func healthEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/vm/health", NewService().GetHealth)
	return r
}

func getHealth(t *testing.T) map[string]any {
	t.Helper()
	rsp := serve(healthEngine(), http.MethodGet, "/api/vm/health", "")
	if rsp["code"] != float64(0) {
		t.Fatalf("GET /api/vm/health answered %v", rsp)
	}
	data, ok := rsp["data"].(map[string]any)
	if !ok {
		t.Fatalf("no data in %v", rsp)
	}
	return data
}

// A board with no sensor, no readable image filesystem and no VPN still
// answers, with nulls the UI reads as "nothing to say".
func TestGetHealthWithNothingToReadIsNotAnError(t *testing.T) {
	useHealth(t)

	data := getHealth(t)
	if data["temperature"] != nil {
		t.Errorf("temperature = %v, want null without a sensor", data["temperature"])
	}
	if data["storage"] != nil {
		t.Errorf("storage = %v, want null when statfs fails", data["storage"])
	}
	if vpn, ok := data["vpn"].([]any); !ok || len(vpn) != 0 {
		t.Errorf("vpn = %v, want an empty list", data["vpn"])
	}
}

func TestGetHealthReportsTemperatureAndSpace(t *testing.T) {
	f := useHealth(t)
	if err := os.WriteFile(cpuTempPath, []byte("71250\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f.space.total, f.space.available, f.space.err = 8<<30, 1<<30, nil

	data := getHealth(t)
	if data["temperature"] != 71.25 {
		t.Errorf("temperature = %v, want 71.25", data["temperature"])
	}
	st, _ := data["storage"].(map[string]any)
	if st["path"] != healthStoragePath || st["total"] != float64(8<<30) || st["available"] != float64(1<<30) {
		t.Errorf("storage = %v", data["storage"])
	}
}

// Zero degrees is a reading, not a missing sensor.
func TestGetHealthKeepsAZeroTemperature(t *testing.T) {
	useHealth(t)
	if err := os.WriteFile(cpuTempPath, []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if data := getHealth(t); data["temperature"] != float64(0) {
		t.Errorf("temperature = %v, want 0", data["temperature"])
	}
}

// Only a VPN set to start at boot is expected up. One that is merely
// installed, or started by hand and not at boot, is not listed.
func TestGetHealthListsOnlyVpnsThatStartAtBoot(t *testing.T) {
	f := useHealth(t)
	f.bootEnabled(t, f.vpnA)
	f.running(t, f.vpnB, 4242)

	vpn, _ := getHealth(t)["vpn"].([]any)
	if len(vpn) != 1 {
		t.Fatalf("vpn = %v, want only a", vpn)
	}
	entry, _ := vpn[0].(map[string]any)
	if entry["name"] != "a" || entry["title"] != "VPN A" || entry["running"] != false {
		t.Errorf("vpn[0] = %v, want a, not running", entry)
	}

	f.running(t, f.vpnA, 4343)
	vpn, _ = getHealth(t)["vpn"].([]any)
	entry, _ = vpn[0].(map[string]any)
	if entry["running"] != true {
		t.Errorf("vpn[0] = %v, want running", entry)
	}
}

// Only the alpha board has an HDD LED input, so only there does the answer
// say the hdd field means anything.
func TestGetGpioSaysWhetherTheBoardHasAnHDDLed(t *testing.T) {
	_, _, led := useButtons(t, "1\n")
	usePowerLEDSetting(t, false, nil)
	conf := config.GetInstance()

	conf.Hardware.Version = config.HWVersionBeta
	conf.Hardware.GPIOHDDLed = ""
	rsp := serve(gpioEngine(), http.MethodGet, "/api/vm/gpio", "")
	data, _ := rsp["data"].(map[string]any)
	if rsp["code"] != float64(0) || data["hasHdd"] != false {
		t.Fatalf("beta: GET answered %v", rsp)
	}

	hdd := led + ".hdd"
	if err := os.WriteFile(hdd, []byte("0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	conf.Hardware.Version = config.HWVersionAlpha
	conf.Hardware.GPIOHDDLed = hdd
	rsp = serve(gpioEngine(), http.MethodGet, "/api/vm/gpio", "")
	data, _ = rsp["data"].(map[string]any)
	if rsp["code"] != float64(0) || data["hasHdd"] != true || data["hdd"] != true {
		t.Fatalf("alpha: GET answered %v", rsp)
	}
}
