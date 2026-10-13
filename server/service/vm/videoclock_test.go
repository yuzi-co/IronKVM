package vm

import (
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/proto"
)

// useVideoClockPaths points the config and the kernel's parameter directory at
// a temporary tree. With kernel set it creates the parameters the driver
// would, profile holding the boot default and rates the given text.
func useVideoClockPaths(t *testing.T, kernel bool, rates string) (config string, params string) {
	t.Helper()

	dir := t.TempDir()
	config = filepath.Join(dir, "video-clock")
	params = filepath.Join(dir, "parameters")

	if kernel {
		if err := os.MkdirAll(params, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(params, "profile"), []byte("fast\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(params, "rates"), []byte(rates), 0o444); err != nil {
			t.Fatal(err)
		}
	}

	origConfig, origParams := videoClockConfigPath, videoClockParamDir
	t.Cleanup(func() {
		videoClockConfigPath, videoClockParamDir = origConfig, origParams
	})
	videoClockConfigPath, videoClockParamDir = config, params

	return config, params
}

func readParam(t *testing.T, params string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(params, "profile"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestVideoClockSupportedFollowsTheDriver(t *testing.T) {
	useVideoClockPaths(t, false, "")
	if VideoClockSupported() {
		t.Fatal("supported without the driver's parameters (slot A)")
	}

	useVideoClockPaths(t, true, "")
	if !VideoClockSupported() {
		t.Fatal("not supported although the driver's parameters exist")
	}
}

func TestParseVideoClockRates(t *testing.T) {
	got, ok := parseVideoClockRates("vc-src0=594000000 vip-sys1=396000000 axi-video-codec=450000000\n")
	want := proto.VideoClockRates{Codec: 594, Vpss: 396, Bus: 450}
	if !ok || got != want {
		t.Fatalf("got %+v, %v; want %+v, true", got, ok, want)
	}

	// An odd rate rounds to the nearest MHz rather than truncating.
	got, ok = parseVideoClockRates("vc-src0=395999999 vip-sys1=300000000 axi-video-codec=360000000")
	if !ok || got.Codec != 396 {
		t.Fatalf("got %+v, %v; want codec 396", got, ok)
	}

	// A driver bound to no clocks prints an empty line.
	if _, ok := parseVideoClockRates("\n"); ok {
		t.Fatal("an empty rates line was read as measured")
	}
	if _, ok := parseVideoClockRates("vc-src0=594000000 vip-sys1=x axi-video-codec=450000000"); ok {
		t.Fatal("a line with an unreadable rate was read as measured")
	}
}

func TestReadVideoClockProfileDefaultsToFast(t *testing.T) {
	config, _ := useVideoClockPaths(t, true, "")

	if got := readVideoClockProfile(); got != "fast" {
		t.Fatalf("no file: got %q, want fast", got)
	}
	if err := os.WriteFile(config, []byte("turbo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readVideoClockProfile(); got != "fast" {
		t.Fatalf("unknown profile: got %q, want fast", got)
	}
	if err := os.WriteFile(config, []byte("standard\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readVideoClockProfile(); got != "standard" {
		t.Fatalf("got %q, want standard", got)
	}
}

func TestApplyVideoClockProfileStandardWritesConfigAndKernel(t *testing.T) {
	config, params := useVideoClockPaths(t, true, "")

	if err := applyVideoClockProfile("standard"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(config)
	if err != nil || string(data) != "standard\n" {
		t.Fatalf("config %q, %v; want standard", data, err)
	}
	if got := readParam(t, params); got != "standard" {
		t.Fatalf("kernel got %q, want standard", got)
	}
}

func TestApplyVideoClockProfileFastIsPristine(t *testing.T) {
	config, params := useVideoClockPaths(t, true, "")

	if err := os.WriteFile(config, []byte("standard\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := applyVideoClockProfile("fast"); err != nil {
		t.Fatal(err)
	}
	if pathExists(config) {
		t.Fatal("the default left a config file behind")
	}
	if got := readParam(t, params); got != "fast" {
		t.Fatalf("kernel got %q, want fast", got)
	}
}

func TestApplyVideoClockProfileRejectsUnknown(t *testing.T) {
	config, params := useVideoClockPaths(t, true, "")

	if err := applyVideoClockProfile("turbo"); err != errUnsupportedVideoClock {
		t.Fatalf("got %v, want errUnsupportedVideoClock", err)
	}
	if pathExists(config) {
		t.Fatal("a rejected profile was stored")
	}
	if got := readParam(t, params); got != "fast\n" {
		t.Fatalf("a rejected profile reached the kernel: %q", got)
	}
}

func TestApplyVideoClockHandsTheStoredProfileToTheKernel(t *testing.T) {
	config, params := useVideoClockPaths(t, true, "")

	if err := os.WriteFile(config, []byte("standard\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ApplyVideoClock()
	if got := readParam(t, params); got != "standard" {
		t.Fatalf("kernel got %q, want standard", got)
	}
}

func TestApplyVideoClockWithoutTheDriverWritesNothing(t *testing.T) {
	config, params := useVideoClockPaths(t, false, "")

	if err := os.WriteFile(config, []byte("standard\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ApplyVideoClock()
	if pathExists(params) {
		t.Fatal("ApplyVideoClock created the parameter directory on a kernel without the driver")
	}
}

func TestRatesMatch(t *testing.T) {
	fast, _ := findVideoClockProfile("fast")
	standard, _ := findVideoClockProfile("standard")
	r := proto.VideoClockRates{Codec: 594, Vpss: 396, Bus: 450}

	if !ratesMatch(r, fast) {
		t.Fatal("fast rates did not match fast")
	}
	if ratesMatch(r, standard) {
		t.Fatal("fast rates matched standard")
	}
	// A profile half applied (the VPSS still busy) is not a match.
	if ratesMatch(proto.VideoClockRates{Codec: 396, Vpss: 396, Bus: 360}, standard) {
		t.Fatal("a half-applied profile matched")
	}
}
