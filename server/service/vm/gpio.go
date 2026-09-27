package vm

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
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

	pwr, err := PowerLED()
	if err != nil {
		rsp.ErrRsp(c, -2, fmt.Sprintf("failed to read power led: %s", err))
		return
	}

	hdd := false
	if conf.Version == config.HWVersionAlpha {
		hdd, err = readGpio(conf.GPIOHDDLed)
		if err != nil {
			rsp.ErrRsp(c, -2, fmt.Sprintf("failed to read hdd led: %s", err))
			return
		}
	}

	data := &proto.GetGpioRsp{
		PWR: pwr,
		HDD: hdd,
	}
	rsp.OkRspWithData(c, data)
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

func writeGpio(device string, duration time.Duration) error {
	lock := gpioLock(device)
	lock.Lock()
	defer lock.Unlock()

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

func readGpio(device string) (bool, error) {
	content, err := os.ReadFile(device)
	if err != nil {
		log.Errorf("read gpio %s failed: %s", device, err)
		return false, err
	}

	contentStr := string(content)
	if len(contentStr) > 1 {
		contentStr = contentStr[:len(contentStr)-1]
	}

	value, err := strconv.Atoi(contentStr)
	if err != nil {
		log.Errorf("invalid gpio content: %s", content)
		return false, nil
	}

	return value == 0, nil
}
