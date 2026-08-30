package vm

import (
	"NanoKVM-Server/proto"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	OLEDExistFile      = "/etc/kvm/oled_exist"
	OLEDSleepFile      = "/etc/kvm/oled_sleep"
	OLEDBrightnessFile = "/etc/kvm/oled_contrast"

	// OLEDFeatureFile is written by kvm_system at start and lists what that
	// build does with the panel.
	//
	// It exists because a release carries Sipeed's kvm_system, not the fork's,
	// and that build reads the sleep file and nothing else. Without this check
	// the brightness control would write a setting nothing acts on, and report
	// success while the panel stayed as it was.
	OLEDFeatureFile = "/tmp/kvm/oled_features"
)

const (
	// minSleepSeconds mirrors OLED_SLEEP_DELAY_MIN in kvm_system. Anything
	// below it disables the screen saver instead of shortening it.
	minSleepSeconds = 10

	// maxSleepSeconds is the largest value kvm_system can hold: it parses the
	// file into a uint16_t, so 65536 wraps back to "never".
	maxSleepSeconds = 65535
)

const (
	// The drive current for SSD1306 command 0x81, mirroring OLED_DRIVE_MIN and
	// OLED_DRIVE_DEFAULT in oled_ctrl.h.
	//
	// The floor is not zero. A drive of zero is a legal command and a screen
	// nobody can read, and this setting is reachable from a browser, so the
	// floor is what stops a slip of the hand from blanking the panel with no
	// way to see how to put it back.
	minBrightness     = 0x10
	maxBrightness     = 0xFF
	defaultBrightness = 0xCF
)

// clampSleepSeconds keeps a requested delay inside the range the firmware can
// act on. Out-of-range values are not rejected outright because the firmware
// reads them as "never sleep", which is the opposite of what was asked for.
func clampSleepSeconds(seconds int) int {
	switch {
	case seconds <= 0:
		return 0 // the one deliberate way to keep the screen on
	case seconds < minSleepSeconds:
		return minSleepSeconds
	case seconds > maxSleepSeconds:
		return maxSleepSeconds
	default:
		return seconds
	}
}

func clampBrightness(level int) int {
	switch {
	case level < minBrightness:
		return minBrightness
	case level > maxBrightness:
		return maxBrightness
	default:
		return level
	}
}

func writeSleepSetting(path string, seconds int) error {
	return os.WriteFile(path, []byte(strconv.Itoa(clampSleepSeconds(seconds))), 0o644)
}

func writeBrightnessSetting(path string, level int) error {
	return os.WriteFile(path, []byte(strconv.Itoa(clampBrightness(level))), 0o644)
}

// readIntSetting returns the value in a setting file, or fallback when the file
// is absent. A file that exists and does not parse is a fault worth reporting,
// so it comes back as an error rather than as the default.
func readIntSetting(path string, fallback int) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fallback, nil
		}

		return fallback, err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return fallback, nil
	}

	return strconv.Atoi(content)
}

// oledFeatures reports what the running kvm_system does with the panel.
func oledFeatures() []string {
	data, err := os.ReadFile(OLEDFeatureFile)
	if err != nil {
		return nil
	}

	return strings.Fields(string(data))
}

func hasOLEDFeature(name string) bool {
	for _, feature := range oledFeatures() {
		if feature == name {
			return true
		}
	}

	return false
}

func (s *Service) SetOLED(c *gin.Context) {
	var req proto.SetOledReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if req.Sleep != nil {
		if err := writeSleepSetting(OLEDSleepFile, *req.Sleep); err != nil {
			rsp.ErrRsp(c, -2, "failed to write data")
			return
		}
		log.Debugf("set OLED sleep: %d", clampSleepSeconds(*req.Sleep))
	}

	if req.Brightness != nil {
		// The setting is written whether or not the running firmware reads it.
		// It is a file on the device, the next kvm_system to start may well be
		// one that acts on it, and refusing the write would lose the value.
		if err := writeBrightnessSetting(OLEDBrightnessFile, *req.Brightness); err != nil {
			rsp.ErrRsp(c, -2, "failed to write data")
			return
		}
		log.Debugf("set OLED brightness: %d", clampBrightness(*req.Brightness))
	}

	rsp.OkRsp(c)
}

func (s *Service) GetOLED(c *gin.Context) {
	var rsp proto.Response

	if _, err := os.Stat(OLEDExistFile); err != nil {
		rsp.OkRspWithData(c, &proto.GetOLEDRsp{
			Exist: false,
			Sleep: 0,
		})
		return
	}

	sleep, err := readIntSetting(OLEDSleepFile, 0)
	if err != nil {
		log.Errorf("failed to parse OLED: %s", err)
		rsp.ErrRsp(c, -1, "failed to parse OLED config")
		return
	}

	brightness, err := readIntSetting(OLEDBrightnessFile, defaultBrightness)
	if err != nil {
		log.Errorf("failed to parse OLED brightness: %s", err)
		rsp.ErrRsp(c, -1, "failed to parse OLED config")
		return
	}

	rsp.OkRspWithData(c, &proto.GetOLEDRsp{
		Exist:               true,
		Sleep:               sleep,
		Brightness:          brightness,
		BrightnessSupported: hasOLEDFeature("brightness"),
	})
	log.Debugf("get OLED config successful, sleep %d, brightness %d", sleep, brightness)
}
