package vm

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// The video clocks of the mainline kernel (slot B): the codec core, the VPSS
// and capture, and the codec bus. Its sg2002_video_clk driver (ironkvm-dist
// patch 0949) takes a profile and changes a clock only while no video block has
// it on, so a new profile takes effect at once when the video is idle, and
// otherwise at the next stream start. The vendor kernel on slot A has no such
// driver, so there the feature is absent and the setting is ignored.
//
// Fast is what the device tree sets at boot (ironkvm-dist trial 82, #97), so it
// is the default: no config file means fast, and choosing fast removes the file.

// videoClockProfile is one selectable set of rates, in MHz.
type videoClockProfile struct {
	name  string
	codec int // vc-src0: Coda980, WAVE420L, JPEG unit
	vpss  int // vip-sys1: VPSS and capture
	bus   int // axi-video-codec
}

// videoClockProfiles is the closed set the kernel accepts, in UI order.
var videoClockProfiles = []videoClockProfile{
	{name: "standard", codec: 396, vpss: 300, bus: 360},
	{name: "fast", codec: 594, vpss: 396, bus: 450},
}

const defaultVideoClockProfile = "fast"

var errUnsupportedVideoClock = errors.New("unsupported video clock profile")

// These are variables so a test can point them at a temporary tree.
var (
	videoClockConfigPath = "/etc/kvm/video-clock"
	videoClockParamDir   = "/sys/module/sg2002_video_clk/parameters"
)

// VideoClockSupported reports whether the running kernel has the profile
// driver. The API contract advertises the feature only when it does.
func VideoClockSupported() bool {
	_, err := os.Stat(videoClockParamDir + "/profile")
	return err == nil
}

// ApplyVideoClock hands the stored profile to the kernel. The server calls it at
// start, before the video does, because the kernel boots with the device tree's
// rates and does not read /etc/kvm itself.
func ApplyVideoClock() {
	if !VideoClockSupported() {
		return
	}
	profile := readVideoClockProfile()
	if err := writeVideoClockParam(profile); err != nil {
		log.Errorf("failed to set the video clock profile %s: %s", profile, err)
	}
}

// GetVideoClock reports the chosen profile and the rates the clock framework
// has now, with the board temperature. The rates differ from the profile's
// while the video that started before the change still runs.
func (s *Service) GetVideoClock(c *gin.Context) {
	var rsp proto.Response

	if !VideoClockSupported() {
		rsp.ErrRsp(c, -1, "not supported by this kernel")
		return
	}

	profile := readVideoClockProfile()
	rates, measured := readVideoClockRates()

	options := make([]proto.VideoClockOption, 0, len(videoClockProfiles))
	for _, p := range videoClockProfiles {
		options = append(options, proto.VideoClockOption{
			Name:  p.name,
			Rates: proto.VideoClockRates{Codec: p.codec, Vpss: p.vpss, Bus: p.bus},
		})
	}

	want, _ := findVideoClockProfile(profile)
	rsp.OkRspWithData(c, &proto.GetVideoClockRsp{
		Profile:     profile,
		Options:     options,
		Rates:       rates,
		Measured:    measured,
		Pending:     measured && !ratesMatch(rates, want),
		Temperature: readCpuTempC(),
	})
}

// SetVideoClock stores the profile and hands it to the kernel, which applies it
// to each clock as soon as no video block uses it.
func (s *Service) SetVideoClock(c *gin.Context) {
	var rsp proto.Response
	var req proto.SetVideoClockReq

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if !VideoClockSupported() {
		rsp.ErrRsp(c, -2, "not supported by this kernel")
		return
	}

	if err := applyVideoClockProfile(req.Profile); err != nil {
		if errors.Is(err, errUnsupportedVideoClock) {
			rsp.ErrRsp(c, -3, "unsupported profile")
			return
		}
		log.Errorf("failed to set the video clock profile %s: %s", req.Profile, err)
		rsp.ErrRsp(c, -4, "set video clock profile failed")
		return
	}

	log.Infof("video clock profile set to %s", req.Profile)
	rsp.OkRsp(c)
}

// applyVideoClockProfile records the profile and passes it to the kernel. The
// default leaves no file behind, as the CPU frequency's stock setting does.
func applyVideoClockProfile(name string) error {
	if _, ok := findVideoClockProfile(name); !ok {
		return errUnsupportedVideoClock
	}

	if name == defaultVideoClockProfile {
		if err := os.Remove(videoClockConfigPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	} else if err := os.WriteFile(videoClockConfigPath, []byte(name+"\n"), 0o644); err != nil {
		return err
	}

	return writeVideoClockParam(name)
}

// readVideoClockProfile returns the stored profile. A missing or unknown file
// means the default, which is also what the kernel boots with.
func readVideoClockProfile() string {
	data, err := os.ReadFile(videoClockConfigPath)
	if err != nil {
		return defaultVideoClockProfile
	}
	name := strings.TrimSpace(string(data))
	if _, ok := findVideoClockProfile(name); !ok {
		return defaultVideoClockProfile
	}
	return name
}

func writeVideoClockParam(name string) error {
	return os.WriteFile(videoClockParamDir+"/profile", []byte(name), 0o644)
}

// readVideoClockRates reads the rates parameter, one name=Hz pair per clock,
// for example "vc-src0=594000000 vip-sys1=396000000 axi-video-codec=450000000".
// It says whether all three were there.
func readVideoClockRates() (proto.VideoClockRates, bool) {
	data, err := os.ReadFile(videoClockParamDir + "/rates")
	if err != nil {
		return proto.VideoClockRates{}, false
	}
	return parseVideoClockRates(string(data))
}

func parseVideoClockRates(s string) (proto.VideoClockRates, bool) {
	var rates proto.VideoClockRates
	seen := 0

	for _, field := range strings.Fields(s) {
		name, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		hz, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			continue
		}
		mhz := int((hz + 500000) / 1000000)
		switch name {
		case "vc-src0":
			rates.Codec = mhz
		case "vip-sys1":
			rates.Vpss = mhz
		case "axi-video-codec":
			rates.Bus = mhz
		default:
			continue
		}
		seen++
	}

	return rates, seen == 3
}

func findVideoClockProfile(name string) (videoClockProfile, bool) {
	for _, p := range videoClockProfiles {
		if p.name == name {
			return p, true
		}
	}
	return videoClockProfile{}, false
}

func ratesMatch(r proto.VideoClockRates, p videoClockProfile) bool {
	return r.Codec == p.codec && r.Vpss == p.vpss && r.Bus == p.bus
}
