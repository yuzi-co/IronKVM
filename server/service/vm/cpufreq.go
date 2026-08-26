package vm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// The SG2002 C906 boots at 850 MHz. The chip is rated for 1000 MHz, so 1000 is
// an in-spec change rather than an overclock. There is no voltage control and
// no cpufreq driver on this board: the clock is set once, by reprogramming a
// PLL and pointing the CPU mux at it. Switching the mux while the system runs
// glitches and can fault a running process, so this feature never switches the
// live clock. It records a target, and S00cpufreq applies it at the next boot,
// before the server and the capture pipeline start.
const stockFreqMHz = 850

// cpuFreqOptions are the frequencies the UI may select. Both are validated on
// SG2002 hardware; 1000 is the chip's rated ceiling. Nothing above rated spec
// is offered here.
var cpuFreqOptions = []int{850, 1000}

// errUnsupportedFreq means the request named a frequency that is not in
// cpuFreqOptions. The set is closed on purpose: an arbitrary value could ask
// the boot script to program a clock the silicon does not hold.
var errUnsupportedFreq = errors.New("unsupported cpu frequency")

// Clock registers on the SG2002. The CPU mux selects a parent PLL and divides
// it; each PLL has a CSR that encodes its rate. These addresses and the field
// layout are fixed by the SoC.
const (
	muxReg  = 0x03002130 // C906 clock mux: parent select and divider
	mpllReg = 0x03002908 // MPLL CSR, the stock CPU parent
	tpllReg = 0x0300290c // TPLL CSR, unused by anything else, used for the 1000 MHz target
)

// muxSel* are the parent-select values this feature produces in the mux
// register. Stock boots on MPLL (sel 3); S00cpufreq points the mux at TPLL
// (sel 0) for 1000 MHz. Any other selection is a state this feature did not
// create, so readRunningMHz reports it as unknown rather than guessing.
const (
	muxSelMpll = 3
	muxSelTpll = 0
)

// These are variables rather than constants so a test can point the reader and
// the config at temporary paths, the way swap_test.go and zram_test.go do.
var (
	cpuFreqConfigPath = "/etc/kvm/cpufreq"
	cpuFreqInitScript = "/etc/init.d/S00cpufreq"
	cpuFreqInitSource = "/kvmapp/system/init.d/S00cpufreq"
	cpuTempPath       = "/sys/class/thermal/thermal_zone0/temp"
)

// readRegister is the seam the tests replace. Production code reads the
// register with devmem; a test supplies the value, because devmem is not on a
// workstation and the addresses are device memory.
var readRegister = func(addr uint32) (uint32, bool) {
	out, err := exec.Command("devmem", fmt.Sprintf("0x%08x", addr)).Output()
	if err != nil {
		return 0, false
	}
	return parseDevmemValue(string(out))
}

// GetCpuFreq reports the running clock, the clock the next boot applies, and
// the CPU temperature. It always succeeds: a board whose registers it cannot
// read still has a valid target to report, and a diagnostic that fails the
// request would tell the operator less than one that reports "unknown".
func (s *Service) GetCpuFreq(c *gin.Context) {
	var rsp proto.Response

	running, measured := readRunningMHz()
	target := readTargetFreq()

	rsp.OkRspWithData(c, &proto.GetCpuFreqRsp{
		Running:        running,
		Measured:       measured,
		Target:         target,
		Temperature:    readCpuTempC(),
		Options:        cpuFreqOptions,
		RebootRequired: measured && running != target,
	})
}

// SetCpuFreq records the frequency the next boot applies. It does not switch
// the live clock: the switch is unsafe while the system runs, so it waits for
// the reboot that the response asks the operator to perform.
func (s *Service) SetCpuFreq(c *gin.Context) {
	var rsp proto.Response
	var req proto.SetCpuFreqReq

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if !isAllowedFreq(req.Target) {
		rsp.ErrRsp(c, -2, "unsupported frequency")
		return
	}

	if err := applyTargetFreq(req.Target); err != nil {
		log.Errorf("failed to set cpu frequency to %d: %s", req.Target, err)
		rsp.ErrRsp(c, -3, "set frequency failed")
		return
	}

	rsp.OkRsp(c)
}

// isAllowedFreq reports whether mhz is one of the offered frequencies.
func isAllowedFreq(mhz int) bool {
	for _, option := range cpuFreqOptions {
		if option == mhz {
			return true
		}
	}
	return false
}

// applyTargetFreq records the boot-time target.
//
// Stock is the pristine state: it removes the config and the boot script so
// the next boot follows the factory path with nothing extra to run. A non-stock
// target writes the config and installs the script that reads it. If the script
// cannot be installed, the config is removed again, because a config that no
// script reads would report a target the board would never apply.
func applyTargetFreq(mhz int) error {
	if !isAllowedFreq(mhz) {
		return errUnsupportedFreq
	}

	if mhz == stockFreqMHz {
		if err := os.Remove(cpuFreqInitScript); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.Remove(cpuFreqConfigPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	if err := writeTargetFreq(mhz); err != nil {
		return err
	}

	if err := installCpuFreqInitScript(); err != nil {
		_ = os.Remove(cpuFreqConfigPath)
		return err
	}

	return nil
}

// readTargetFreq returns the frequency the next boot applies. It reads the
// config, not the registers, so it reports the pending choice even before the
// reboot that makes it live. An absent or unrecognized file means stock, which
// is what the boot script also falls back to.
func readTargetFreq() int {
	data, err := os.ReadFile(cpuFreqConfigPath)
	if err != nil {
		return stockFreqMHz
	}

	mhz, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || !isAllowedFreq(mhz) {
		return stockFreqMHz
	}
	return mhz
}

// writeTargetFreq stores the target for S00cpufreq to read at boot.
func writeTargetFreq(mhz int) error {
	return os.WriteFile(cpuFreqConfigPath, []byte(strconv.Itoa(mhz)+"\n"), 0o644)
}

// readRunningMHz decodes the clock the core runs now from the mux and PLL
// registers. It returns ok=false when the registers cannot be read or when the
// mux selects a parent this feature never sets, because a wrong number would be
// worse than an honest "unknown".
func readRunningMHz() (int, bool) {
	mux, ok := readRegister(muxReg)
	if !ok {
		return 0, false
	}

	var csrAddr uint32
	switch (mux >> 8) & 0x3 {
	case muxSelMpll:
		csrAddr = mpllReg
	case muxSelTpll:
		csrAddr = tpllReg
	default:
		return 0, false
	}

	csr, ok := readRegister(csrAddr)
	if !ok {
		return 0, false
	}

	rate := pllRateMHz(csr)
	if rate == 0 {
		return 0, false
	}

	div := int((mux >> 16) & 0xF)
	if div == 0 {
		div = 1
	}

	return rate / div, true
}

// pllRateMHz computes a fractional PLL rate from its CSR. The rate is
// 25 MHz * div_sel / (pre_div_sel * post_div_sel), with the field positions
// fixed by the SoC. A zero divider field is treated as one so a malformed
// register cannot divide by zero; the caller still gets a plausible number
// rather than a panic.
func pllRateMHz(csr uint32) int {
	const oscMHz = 25

	preDiv := int(csr & 0x7F)
	postDiv := int((csr >> 8) & 0x7F)
	divSel := int((csr >> 17) & 0x7F)

	if preDiv == 0 {
		preDiv = 1
	}
	if postDiv == 0 {
		postDiv = 1
	}

	return oscMHz * divSel / (preDiv * postDiv)
}

// parseDevmemValue reads the hex word devmem prints, for example "0x00448101".
func parseDevmemValue(s string) (uint32, bool) {
	field := strings.TrimSpace(s)
	value, err := strconv.ParseUint(strings.TrimPrefix(field, "0x"), 16, 32)
	if err != nil {
		return 0, false
	}
	return uint32(value), true
}

// readCpuTempC reads the CPU temperature in degrees C. The sysfs value is in
// millidegrees. An unreadable sensor returns 0, which the UI shows as no
// reading rather than a fault.
func readCpuTempC() float64 {
	data, err := os.ReadFile(cpuTempPath)
	if err != nil {
		return 0
	}

	milli, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0
	}
	return float64(milli) / 1000.0
}

// installCpuFreqInitScript copies the packaged boot script into /etc/init.d.
//
// It mirrors installZramInitScript: the list of scripts kvm_system copies at
// boot is hard-coded C++, so the server installs this one instead, and its
// presence in /etc/init.d is the installed marker.
func installCpuFreqInitScript() error {
	content, err := os.ReadFile(cpuFreqInitSource)
	if err != nil {
		log.Errorf("failed to read %s: %s", cpuFreqInitSource, err)
		return err
	}

	// The mode matters: a script that is not executable never runs at boot,
	// and the failure only appears after a reboot.
	if err := os.WriteFile(cpuFreqInitScript, content, 0o755); err != nil {
		log.Errorf("failed to write %s: %s", cpuFreqInitScript, err)
		return err
	}

	// WriteFile does not change the mode of a file that already exists.
	if err := os.Chmod(cpuFreqInitScript, 0o755); err != nil {
		log.Errorf("failed to chmod %s: %s", cpuFreqInitScript, err)
		return err
	}

	return nil
}
