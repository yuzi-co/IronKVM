package edid

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func readPreset(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("presets/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// The EDID Sipeed ships for nanokvm_update_edid (tools/nanokvm_update_edid).
// Every value below was read off the bytes by hand, not from this parser.
func TestParseSipeedStock(t *testing.T) {
	info, err := Parse(readPreset(t, "sipeed-nanokvm.bin"))
	if err != nil {
		t.Fatal(err)
	}

	if !info.Valid || len(info.Problems) != 0 {
		t.Fatalf("valid = %v, problems = %v", info.Valid, info.Problems)
	}
	if info.Size != 256 || info.Blocks != 2 || info.Version != "1.3" {
		t.Errorf("size %d, blocks %d, version %q", info.Size, info.Blocks, info.Version)
	}
	if info.Manufacturer != "SPD" || info.ProductCode != 0x3301 || info.SerialNumber != 21 {
		t.Errorf("manufacturer %q, product %#x, serial %d", info.Manufacturer, info.ProductCode, info.SerialNumber)
	}
	if info.Name != "NanoKVM" || info.SerialText != "NanoKVM" {
		t.Errorf("name %q, serial text %q", info.Name, info.SerialText)
	}
	if info.Week != 30 || info.Year != 2025 || info.ModelYear {
		t.Errorf("week %d, year %d, model year %v", info.Week, info.Year, info.ModelYear)
	}
	if !info.Digital || info.WidthCm != 48 || info.HeightCm != 27 || info.Gamma != 2.2 {
		t.Errorf("digital %v, %dx%d cm, gamma %v", info.Digital, info.WidthCm, info.HeightCm, info.Gamma)
	}

	want1080 := Timing{
		PixelClockKHz: 148500,
		HActive:       1920, HBlank: 280, HFrontPorch: 88, HSync: 44,
		VActive: 1080, VBlank: 45, VFrontPorch: 4, VSync: 5,
		WidthMm: 476, HeightMm: 268,
		HSyncPositive: true, VSyncPositive: true,
		RefreshHz: 60, HFreqKHz: 67.5,
	}
	if info.Preferred == nil || *info.Preferred != want1080 {
		t.Fatalf("preferred = %+v, want %+v", info.Preferred, want1080)
	}
	if len(info.DetailedTimings) != 3 {
		t.Fatalf("detailed timings = %d, want 3 (one base, two CTA)", len(info.DetailedTimings))
	}
	cta720 := info.DetailedTimings[2]
	if cta720.HActive != 1280 || cta720.VActive != 720 || cta720.PixelClockKHz != 74250 ||
		cta720.HFrontPorch != 110 || cta720.HSync != 40 || cta720.VFrontPorch != 5 || cta720.VSync != 5 ||
		cta720.RefreshHz != 60 || cta720.Block != 1 {
		t.Errorf("CTA 720p timing = %+v", cta720)
	}

	wantEstablished := []Mode{
		{Width: 720, Height: 400, RefreshHz: 70},
		{Width: 640, Height: 480, RefreshHz: 60},
		{Width: 640, Height: 480, RefreshHz: 75},
		{Width: 800, Height: 600, RefreshHz: 60},
		{Width: 800, Height: 600, RefreshHz: 75},
		{Width: 1024, Height: 768, RefreshHz: 60},
		{Width: 1024, Height: 768, RefreshHz: 75},
		{Width: 1280, Height: 1024, RefreshHz: 75},
	}
	if !reflect.DeepEqual(info.EstablishedTimings, wantEstablished) {
		t.Errorf("established = %+v", info.EstablishedTimings)
	}
	wantStandard := []Mode{
		{Width: 1920, Height: 1080, RefreshHz: 60},
		{Width: 1280, Height: 1024, RefreshHz: 60},
	}
	if !reflect.DeepEqual(info.StandardTimings, wantStandard) {
		t.Errorf("standard = %+v", info.StandardTimings)
	}

	wantRange := &RangeLimits{MinVHz: 56, MaxVHz: 76, MinHKHz: 30, MaxHKHz: 83, MaxPixelClockMHz: 170}
	if !reflect.DeepEqual(info.RangeLimits, wantRange) {
		t.Errorf("range limits = %+v", info.RangeLimits)
	}

	if len(info.Extensions) != 1 || info.Extensions[0].Tag != 0x02 || info.Extensions[0].Name != "CTA-861" {
		t.Fatalf("extensions = %+v", info.Extensions)
	}
	cta := info.CTA
	if cta == nil {
		t.Fatal("no CTA block decoded")
	}
	if cta.Revision != 3 || cta.BasicAudio || !cta.YCbCr444 || !cta.YCbCr422 || cta.Underscan || cta.NativeDTDs != 0 {
		t.Errorf("CTA header = %+v", cta)
	}
	var vics []int
	for _, v := range cta.VICs {
		vics = append(vics, v.Code)
	}
	if !reflect.DeepEqual(vics, []int{4, 31, 20, 19, 1, 16}) {
		t.Errorf("VICs = %v", vics)
	}
	if v := cta.VICs[2]; v.Width != 1920 || v.Height != 1080 || !v.Interlaced || v.RefreshHz != 50 {
		t.Errorf("VIC 20 = %+v", v)
	}
	if v := cta.VICs[0]; v.Width != 1280 || v.Height != 720 || v.Interlaced || v.RefreshHz != 60 || v.Aspect != "16:9" {
		t.Errorf("VIC 4 = %+v", v)
	}
	if len(cta.Audio) != 0 {
		t.Errorf("audio = %+v, want none", cta.Audio)
	}
	if cta.HDMI == nil || cta.HDMI.PhysicalAddress != "1.0.0.0" || cta.HDMI.MaxTMDSMHz != 0 {
		t.Errorf("HDMI VSDB = %+v", cta.HDMI)
	}

	if len(info.Checksums) != 2 || !info.Checksums[0].Valid || info.Checksums[0].Stored != 0xf5 ||
		!info.Checksums[1].Valid || info.Checksums[1].Stored != 0xda {
		t.Errorf("checksums = %+v", info.Checksums)
	}
}

// The EDID the LT6911UXC presents with an empty EDID slot, as a host read it
// (ironkvm-dist trial 79). Unlike the stock one it declares audio and
// speakers.
func TestParseBridgeBuiltin(t *testing.T) {
	info, err := Parse(readPreset(t, "lt6911uxc-builtin.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.Valid {
		t.Fatalf("problems: %v", info.Problems)
	}
	if info.Manufacturer != "VCS" || info.ProductCode != 0x1145 || info.Name != "Connector" {
		t.Errorf("manufacturer %q, product %#x, name %q", info.Manufacturer, info.ProductCode, info.Name)
	}
	if info.Preferred == nil || info.Preferred.HActive != 1920 || info.Preferred.VActive != 1080 || info.Preferred.RefreshHz != 60 {
		t.Errorf("preferred = %+v", info.Preferred)
	}
	if info.CTA == nil || !info.CTA.BasicAudio {
		t.Fatalf("CTA = %+v", info.CTA)
	}
	if len(info.CTA.Audio) != 1 {
		t.Fatalf("audio = %+v", info.CTA.Audio)
	}
	audio := info.CTA.Audio[0]
	if audio.Format != "LPCM" || audio.Channels != 2 ||
		!reflect.DeepEqual(audio.SampleRatesKHz, []float64{32, 44.1, 48}) ||
		!reflect.DeepEqual(audio.BitDepths, []int{16, 20}) {
		t.Errorf("audio = %+v", audio)
	}
	if !reflect.DeepEqual(info.CTA.Speakers, []string{"FL/FR"}) {
		t.Errorf("speakers = %v", info.CTA.Speakers)
	}
	if info.CTA.HDMI == nil || info.CTA.HDMI.PhysicalAddress != "1.0.0.0" {
		t.Errorf("HDMI = %+v", info.CTA.HDMI)
	}
	// Bit 7 of a VIC byte marks a native mode for codes 129 to 192.
	native := 0
	for _, v := range info.CTA.VICs {
		if v.Native {
			native++
			if v.Code != 16 {
				t.Errorf("native VIC %d, want 16", v.Code)
			}
		}
	}
	if native != 1 {
		t.Errorf("native VICs = %d, want 1", native)
	}
}

func TestChecksumMismatchIsReportedNotFatal(t *testing.T) {
	data := readPreset(t, "sipeed-nanokvm.bin")
	data[0x7F]++ // base block checksum
	data[0x90] ^= 0x01

	info, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if info.Valid {
		t.Fatal("an EDID with bad checksums decoded as valid")
	}
	if c := info.Checksums[0]; c.Valid || c.Stored != 0xf6 || c.Computed != 0xf5 {
		t.Errorf("block 0 checksum = %+v", c)
	}
	if c := info.Checksums[1]; c.Valid || c.Stored != 0xda {
		t.Errorf("block 1 checksum = %+v", c)
	}
	if len(info.Problems) != 2 || !strings.Contains(info.Problems[0], "block 0") || !strings.Contains(info.Problems[1], "block 1") {
		t.Errorf("problems = %v", info.Problems)
	}
	// The rest still decodes, so a broken file can be looked at.
	if info.Preferred == nil || info.Preferred.HActive != 1920 {
		t.Errorf("preferred = %+v", info.Preferred)
	}
}

func TestParseRejectsWhatIsNotAnEDID(t *testing.T) {
	stock := readPreset(t, "sipeed-nanokvm.bin")
	badHeader := append([]byte(nil), stock...)
	badHeader[3] = 0

	for name, tc := range map[string]struct {
		data []byte
		want error
	}{
		"empty":      {nil, ErrTruncated},
		"short":      {stock[:100], ErrTruncated},
		"bad header": {badHeader, ErrNotEDID},
		"too large":  {make([]byte, MaxSize+128), ErrTooLarge},
	} {
		t.Run(name, func(t *testing.T) {
			info, err := Parse(tc.data)
			if !errors.Is(err, tc.want) || info != nil {
				t.Fatalf("Parse = %v, %v; want %v", info, err, tc.want)
			}
		})
	}
}

func TestMissingExtensionIsAProblem(t *testing.T) {
	stock := readPreset(t, "sipeed-nanokvm.bin")

	info, err := Parse(stock[:128])
	if err != nil {
		t.Fatal(err)
	}
	if info.Valid || info.Blocks != 1 || info.CTA != nil {
		t.Fatalf("valid %v, blocks %d, CTA %v", info.Valid, info.Blocks, info.CTA)
	}
	if len(info.Problems) != 1 || !strings.Contains(info.Problems[0], "1 extension") {
		t.Errorf("problems = %v", info.Problems)
	}

	// A partial block at the end is dropped and named.
	info, err = Parse(stock[:200])
	if err != nil {
		t.Fatal(err)
	}
	if info.Valid || info.Blocks != 1 {
		t.Fatalf("valid %v, blocks %d", info.Valid, info.Blocks)
	}
	if !strings.Contains(strings.Join(info.Problems, "; "), "72 trailing bytes") {
		t.Errorf("problems = %v", info.Problems)
	}
}

// A CTA block whose data block runs past the DTD offset must not read out of
// bounds or loop.
func TestMalformedCTADataBlockIsAProblem(t *testing.T) {
	data := readPreset(t, "sipeed-nanokvm.bin")
	data[0x84] = 0x5f // video data block, length 31: past the DTD offset 17
	fixChecksum(data[128:256])

	info, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if info.Valid {
		t.Fatal("decoded as valid")
	}
	if !strings.Contains(strings.Join(info.Problems, "; "), "data block") {
		t.Errorf("problems = %v", info.Problems)
	}
}

func fixChecksum(block []byte) {
	var sum byte
	for _, b := range block[:127] {
		sum += b
	}
	block[127] = -sum
}

func TestDecodeInput(t *testing.T) {
	stock := readPreset(t, "sipeed-nanokvm.bin")

	var hexText strings.Builder
	for i, b := range stock {
		if i > 0 && i%16 == 0 {
			hexText.WriteString("\r\n")
		} else if i > 0 {
			hexText.WriteString(" ")
		}
		hexText.WriteString(strings.ToUpper(string("0123456789abcdef"[b>>4]) + string("0123456789abcdef"[b&15])))
	}

	for name, input := range map[string][]byte{
		"binary":            stock,
		"hex with spaces":   []byte(hexText.String()),
		"hex in one run":    []byte("  " + strings.ReplaceAll(strings.ReplaceAll(hexText.String(), " ", ""), "\r\n", "") + "\n"),
		"xrandr tab indent": []byte("\t\t" + strings.ReplaceAll(strings.ReplaceAll(hexText.String(), " ", ""), "\r\n", "\n\t\t")),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeInput(input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, stock) {
				t.Fatalf("got %d bytes, differs from the input EDID", len(got))
			}
		})
	}

	for name, input := range map[string][]byte{
		"odd hex digits": []byte("00ffffffffffff0"),
		"not hex":        []byte("hello, world"),
		"empty":          nil,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeInput(input); !errors.Is(err, ErrNotEDID) && !errors.Is(err, ErrTruncated) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}
