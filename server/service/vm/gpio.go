package vm

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils/gpiocdev"
)

func (s *Service) SetGpio(c *gin.Context) {
	var req proto.SetGpioReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, fmt.Sprintf("invalid arguments: %s", err))
		return
	}

	if req.Type != ButtonPower && req.Type != ButtonReset {
		rsp.ErrRsp(c, -2, fmt.Sprintf("invalid power event: %s", req.Type))
		return
	}

	if err := PressButton(req.Type, pressDuration(req.Duration)); err != nil {
		rsp.ErrRsp(c, -3, fmt.Sprintf("operation failed: %s", err))
		return
	}

	log.Debugf("%s button pressed", req.Type)
	rsp.OkRsp(c)
}

func (s *Service) GetGpio(c *gin.Context) {
	var rsp proto.Response

	conf := config.GetInstance().Hardware

	// An LED that is not wired reads "off" whatever the host does, so it is
	// not read at all, and the UI shows the state as unknown.
	connected := PowerLEDConnected()
	pwr := false
	var err error
	if connected {
		pwr, err = PowerLED()
		if err != nil {
			rsp.ErrRsp(c, -2, fmt.Sprintf("failed to read power led: %s", err))
			return
		}
	}

	hdd := false
	hasHDD := hasHDDLed(conf)
	if hasHDD {
		hdd, err = readGpio(conf.GPIOHDDLed)
		if err != nil {
			rsp.ErrRsp(c, -2, fmt.Sprintf("failed to read hdd led: %s", err))
			return
		}
	}

	data := &proto.GetGpioRsp{
		PWR:          pwr,
		HDD:          hdd,
		HasHDD:       hasHDD,
		LEDConnected: connected,
	}
	rsp.OkRspWithData(c, data)
}

// hasHDDLed reports whether the board has an HDD LED input. Only the alpha
// board wires one to a GPIO; beta and PCIe leave the path empty.
func hasHDDLed(conf config.Hardware) bool {
	return conf.Version == config.HWVersionAlpha && conf.GPIOHDDLed != ""
}

const (
	defaultPressDuration = 800 * time.Millisecond
	maxPressDuration     = 10 * time.Second
)

// pressDuration is how long the line is held for a request.
//
// The value arrives from the client, and the line stays asserted for the whole
// of it, so an unbounded one holds the attached machine in reset for as long as
// the caller cares to name. Nobody presses a power button for more than a few
// seconds.
func pressDuration(milliseconds uint) time.Duration {
	if milliseconds == 0 {
		return defaultPressDuration
	}

	duration := time.Duration(milliseconds) * time.Millisecond
	if duration > maxPressDuration || duration < 0 {
		return maxPressDuration
	}

	return duration
}

// The two front-panel buttons the board is wired to.
const (
	ButtonPower = "power"
	ButtonReset = "reset"
)

var (
	errUnknownButton   = errors.New("unknown button")
	errPressOutOfRange = errors.New("press duration out of range")
)

// buttonMu is held for the whole of a press, whichever button it is and
// whoever asked for it. The UI and Redfish both press through PressButton,
// so a Redfish reset never lands in the middle of a power press from the UI,
// or the other way round.
var buttonMu sync.Mutex

// PressButton holds the power or the reset button for d, then releases it.
// It returns once the button is up again.
func PressButton(kind string, d time.Duration) error {
	conf := config.GetInstance().Hardware

	var device string
	switch kind {
	case ButtonPower:
		device = conf.GPIOPower
	case ButtonReset:
		device = conf.GPIOReset
	default:
		return fmt.Errorf("%w: %q", errUnknownButton, kind)
	}

	if d <= 0 || d > maxPressDuration {
		return fmt.Errorf("%w: %s", errPressOutOfRange, d)
	}

	buttonMu.Lock()
	defer buttonMu.Unlock()

	return writeGpio(device, d)
}

// PowerLED reports whether the host's power LED is lit. It fails on a board
// whose LED line cannot be read.
func PowerLED() (bool, error) {
	return readGpio(config.GetInstance().Hardware.GPIOPowerLED)
}

var (
	gpioLocksMutex sync.Mutex
	gpioLocks      = map[string]*sync.Mutex{}
)

// gpioLock returns the lock for one line. Presses of the same line have to be
// serialized: interleaved, one press releases the line while the other still
// believes it is holding it, so a reset becomes a no-op or the line is left
// asserted after both callers have gone. Separate lines are independent.
func gpioLock(device string) *sync.Mutex {
	gpioLocksMutex.Lock()
	defer gpioLocksMutex.Unlock()

	lock, ok := gpioLocks[device]
	if !ok {
		lock = &sync.Mutex{}
		gpioLocks[device] = lock
	}

	return lock
}

// cdevLines holds the lines taken through the GPIO character device, by name.
// An output stays requested once taken, so it stays driven low between presses
// rather than floating; the kernel releases it when the process exits.
var (
	cdevLinesMu sync.Mutex
	cdevLines   = map[string]*gpiocdev.Line{}
)

// cdevLine returns the character-device line to use in place of the sysfs
// path device, or nil when the sysfs path is the one to use.
//
// The sysfs path wins whenever it exists, which is always on the vendor kernel
// once the init script has exported the line, so nothing changes there. The
// mainline kernel has no /sys/class/gpio, and there the line is found by the
// name config.Hardware.GPIOLineName gives it. A name no chip carries (the
// vendor device tree names no lines) also returns nil, and the caller then
// fails on the sysfs path exactly as it did before.
func cdevLine(device string, dir gpiocdev.Direction) *gpiocdev.Line {
	if device == "" {
		return nil
	}
	if _, err := os.Stat(device); err == nil || !errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	name := config.GetInstance().Hardware.GPIOLineName(device)
	if name == "" {
		return nil
	}

	cdevLinesMu.Lock()
	defer cdevLinesMu.Unlock()

	if l, ok := cdevLines[name]; ok {
		return l
	}
	l, err := gpiocdev.Request(name, dir, 0)
	if err != nil {
		if !errors.Is(err, gpiocdev.ErrNotFound) {
			noteGpioRead(device, err)
		}
		return nil
	}
	log.Infof("gpio %s does not exist, using the gpio line %q", device, name)
	cdevLines[name] = l
	return l
}

func writeGpio(device string, duration time.Duration) error {
	lock := gpioLock(device)
	lock.Lock()
	defer lock.Unlock()

	if line := cdevLine(device, gpiocdev.Output); line != nil {
		return pressLine(line, duration)
	}

	if err := os.WriteFile(device, []byte("1"), 0o666); err != nil {
		log.Errorf("write gpio %s failed: %s", device, err)
		return err
	}

	time.Sleep(duration)

	if err := os.WriteFile(device, []byte("0"), 0o666); err != nil {
		log.Errorf("write gpio %s failed: %s", device, err)
		return err
	}

	return nil
}

// pressLine is writeGpio's press on a character-device line.
func pressLine(line *gpiocdev.Line, duration time.Duration) error {
	if err := line.Set(1); err != nil {
		log.Errorf("set gpio line %s failed: %s", line.Name(), err)
		return err
	}

	time.Sleep(duration)

	if err := line.Set(0); err != nil {
		log.Errorf("set gpio line %s failed: %s", line.Name(), err)
		return err
	}

	return nil
}

// gpioReadFailing holds the lines whose last read failed. A board without an
// LED wired fails every read, and Redfish reads on every GET, so a failure is
// logged when it starts, not each time it repeats.
var (
	gpioReadFailingMu sync.Mutex
	gpioReadFailing   = map[string]bool{}
)

// noteGpioRead logs err for device unless the previous read of it failed too.
// A nil err clears the note, so the next failure is logged again.
func noteGpioRead(device string, err error) {
	gpioReadFailingMu.Lock()
	defer gpioReadFailingMu.Unlock()

	if err == nil {
		delete(gpioReadFailing, device)
		return
	}
	if gpioReadFailing[device] {
		return
	}
	gpioReadFailing[device] = true
	log.Errorf("read gpio %s failed: %s", device, err)
}

// readGpio reports whether an active-low line is asserted. Content that is
// not a number is an error, not "released": the state is unknown.
func readGpio(device string) (bool, error) {
	if line := cdevLine(device, gpiocdev.Input); line != nil {
		value, err := line.Get()
		noteGpioRead(device, err)
		if err != nil {
			return false, err
		}
		return value == 0, nil
	}

	content, err := os.ReadFile(device)
	if err != nil {
		noteGpioRead(device, err)
		return false, err
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		err = fmt.Errorf("invalid gpio content %q", content)
		noteGpioRead(device, err)
		return false, err
	}

	noteGpioRead(device, nil)
	return value == 0, nil
}
