package edid

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"math"
	"os"
	"testing"
)

// go test ./service/edid -run TestGeneratedPresetsMatchTheirSpec -update
// rewrites the generated presets from their specs.
var update = flag.Bool("update", false, "rewrite the generated preset files")

func TestGeneratedPresetsMatchTheirSpec(t *testing.T) {
	for _, p := range generated {
		t.Run(p.ID, func(t *testing.T) {
			built := Build(p.spec)
			path := "presets/" + p.ID + ".bin"
			if *update {
				if err := os.WriteFile(path, built, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			committed, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(built, committed) {
				t.Fatalf("%s differs from Build(spec); run with -update if the spec changed on purpose", path)
			}
		})
	}
}

func TestEveryPresetDecodesClean(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range Presets() {
		t.Run(p.ID, func(t *testing.T) {
			if seen[p.ID] {
				t.Fatalf("duplicate id %s", p.ID)
			}
			seen[p.ID] = true

			data, err := PresetData(p.ID)
			if err != nil {
				t.Fatal(err)
			}
			info, err := Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			if !info.Valid {
				t.Fatalf("problems: %v", info.Problems)
			}
			sum := sha256.Sum256(data)
			if p.SHA256 != hex.EncodeToString(sum[:]) || p.Size != len(data) {
				t.Errorf("sha256 %s size %d, listed %s %d", hex.EncodeToString(sum[:]), len(data), p.SHA256, p.Size)
			}
			if p.Name == "" || p.Source == "" || p.Licence == "" {
				t.Errorf("name %q, source %q, licence %q: each preset says where it came from", p.Name, p.Source, p.Licence)
			}
			switch p.Origin {
			case OriginGenerated, OriginSipeed, OriginBridge:
			default:
				t.Errorf("origin %q", p.Origin)
			}
			for _, slot := range p.CapturedOn {
				if slot != "a" && slot != "b" {
					t.Errorf("slot %q", slot)
				}
			}
			if p.Preferred == nil || *p.Preferred != *info.Preferred {
				t.Errorf("listed preferred %+v, decoded %+v", p.Preferred, info.Preferred)
			}
		})
	}
	if !seen["sipeed-nanokvm"] || !seen["lt6911uxc-builtin"] {
		t.Errorf("presets %v lack the stock or the built-in EDID", seen)
	}
}

// The stock preset is the file nanokvm_update_edid ships, byte for byte.
func TestSipeedPresetIsTheShippedFile(t *testing.T) {
	shipped, err := os.ReadFile("../../../tools/nanokvm_update_edid/E21_NanoKVM.bin")
	if err != nil {
		t.Fatal(err)
	}
	data, err := PresetData("sipeed-nanokvm")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(shipped, data) {
		t.Fatal("presets/sipeed-nanokvm.bin differs from tools/nanokvm_update_edid/E21_NanoKVM.bin")
	}
}

// The single-rate rule (ironkvm-dist #69, after NanoKVM-OS): a generated
// preset lists no mode as wide or as tall as its preferred one other than the
// preferred mode itself, because a GPU driver may pick such a mode over the
// preferred one.
func TestGeneratedPresetsKeepTheSingleRateRule(t *testing.T) {
	for _, p := range generated {
		t.Run(p.ID, func(t *testing.T) {
			info, err := Parse(Build(p.spec))
			if err != nil {
				t.Fatal(err)
			}
			pref := info.Preferred
			if pref == nil {
				t.Fatal("no preferred timing")
			}
			want := p.spec.Timings[0]
			if pref.HActive != want.HActive || pref.VActive != want.VActive || pref.PixelClockKHz != want.PixelClockKHz {
				t.Fatalf("preferred %+v, spec %+v", pref, want)
			}
			for _, m := range AllModes(info) {
				if m.Width < pref.HActive && m.Height < pref.VActive {
					continue
				}
				if m.Width == pref.HActive && m.Height == pref.VActive && math.Abs(m.RefreshHz-pref.RefreshHz) < 0.5 {
					continue
				}
				t.Errorf("mode %dx%d@%v competes with the preferred %dx%d@%v",
					m.Width, m.Height, m.RefreshHz, pref.HActive, pref.VActive, pref.RefreshHz)
			}
			if info.RangeLimits == nil {
				t.Fatal("no range limits")
			}
			for _, m := range AllModes(info) {
				if m.RefreshHz < float64(info.RangeLimits.MinVHz)-0.1 || m.RefreshHz > float64(info.RangeLimits.MaxVHz) {
					t.Errorf("mode %+v outside the range limits %+v", m, info.RangeLimits)
				}
			}
			if info.CTA == nil || info.CTA.HDMI == nil || float64(info.CTA.HDMI.MaxTMDSMHz)*1000 < float64(pref.PixelClockKHz) {
				t.Errorf("HDMI VSDB %+v does not cover %d kHz", info.CTA.HDMI, pref.PixelClockKHz)
			}
		})
	}
}

func TestUnknownPreset(t *testing.T) {
	if _, err := PresetData("nope"); !errors.Is(err, ErrUnknownPreset) {
		t.Fatalf("err = %v", err)
	}
	if _, err := PresetData("../presets_test.go"); !errors.Is(err, ErrUnknownPreset) {
		t.Fatalf("err = %v", err)
	}
}

// The board cannot read the EDID back without driving the bridge's I2C bus,
// so the answer is "not readable" and points at the built-in one.
func TestCurrentIsNotReadable(t *testing.T) {
	cur := Current()
	if cur.Readable || cur.Reason != ReasonBridgeOnly || cur.BuiltinPreset != "lt6911uxc-builtin" {
		t.Fatalf("current = %+v", cur)
	}
	if _, err := PresetData(cur.BuiltinPreset); err != nil {
		t.Fatal(err)
	}
}
