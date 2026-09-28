package ipmi

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	shortPress = 800 * time.Millisecond
	longPress  = 5 * time.Second
)

// The press timings are variables so tests can shorten them. pressSettle
// is how long Chassis Control keeps the lock after its last press, so the
// next command reads an LED that has had time to follow. cyclePause is the
// gap between the two halves of a power cycle, as the watchdog uses.
var (
	pressSettle = 2 * time.Second
	cyclePause  = 5 * time.Second
)

// The Chassis Control actions, IPMI 2.0 table 28-4.
const (
	controlPowerDown  = 0x00
	controlPowerUp    = 0x01
	controlPowerCycle = 0x02
	controlHardReset  = 0x03
	controlDiagInt    = 0x04
	controlSoftOff    = 0x05
)

// press is one button press, and pause the wait after it.
type press struct {
	button string
	hold   time.Duration
	pause  time.Duration
}

// powerLED reads the LED, or returns nil when the state is unknown: the
// owner has not said the LED is wired, or its line cannot be read. An LED
// that is not wired reads "off" whatever the host does.
func (s *Service) powerLED() (*bool, error) {
	if !s.deps.PowerLEDConnected() {
		return nil, nil
	}
	on, err := s.deps.PowerLED()
	if err != nil {
		return nil, err
	}
	return &on, nil
}

// chassisStatus is Get Chassis Status, IPMI 2.0 section 28.2. IPMI has no
// way to say the power state is unknown, and "off" would be a lie a tool
// might act on, so without the LED the command fails.
func (s *Service) chassisStatus() (byte, []byte) {
	on, err := s.powerLED()
	if err != nil {
		log.Errorf("ipmi: read the power LED: %s", err)
		return ccUnspecified, nil
	}
	if on == nil {
		return ccNotInPresentState, nil
	}

	// Bits 6:5 of the power state are the restore policy, 11b unknown: the
	// board does not control what the host does when power returns.
	power := byte(0x60)
	if *on {
		power |= 0x01
	}
	// Bit 6 of the third byte says Chassis Identify is supported, with its
	// state, off, in bits 5:4.
	const identifySupported = 0x40
	return ccOK, []byte{power, 0x00, identifySupported}
}

// planControl maps a Chassis Control action and the power LED to the
// presses it needs. on is nil when the state is unknown. No presses and
// ccOK means the host is already in the state asked for.
func planControl(action byte, on *bool) ([]press, byte) {
	switch action {
	case controlHardReset:
		return []press{{ButtonReset, shortPress, 0}}, ccOK
	case controlPowerDown, controlPowerUp, controlPowerCycle, controlSoftOff:
	case controlDiagInt:
		// There is no NMI line.
		return nil, ccInvalidDataField
	default:
		return nil, ccInvalidDataField
	}

	if on == nil {
		return nil, ccNotInPresentState
	}

	switch action {
	case controlPowerDown:
		if *on {
			return []press{{ButtonPower, longPress, 0}}, ccOK
		}
	case controlPowerUp:
		if !*on {
			return []press{{ButtonPower, shortPress, 0}}, ccOK
		}
	case controlSoftOff:
		if *on {
			return []press{{ButtonPower, shortPress, 0}}, ccOK
		}
	case controlPowerCycle:
		// The specification leaves a cycle of a host that is off to the
		// BMC. Turning it on would surprise whoever asked for a cycle.
		if !*on {
			return nil, ccNotInPresentState
		}
		return []press{{ButtonPower, longPress, cyclePause}, {ButtonPower, shortPress, 0}}, ccOK
	}
	return nil, ccOK
}

// chassisControl is Chassis Control, IPMI 2.0 section 28.3. It answers as
// soon as the presses are planned and makes them after, because a 5 second
// hold is longer than ipmitool waits for an answer. A command that arrives
// while presses are running answers "node busy".
func (s *Service) chassisControl(sess *session, data []byte) (byte, []byte) {
	if len(data) < 1 {
		return ccRequestDataLength, nil
	}
	action := data[0] & 0x0f

	if !s.powerMu.TryLock() {
		return ccNodeBusy, nil
	}
	on, err := s.powerLED()
	if err != nil && action != controlHardReset {
		s.powerMu.Unlock()
		log.Errorf("ipmi: read the power LED: %s", err)
		return ccUnspecified, nil
	}
	presses, cc := planControl(action, on)
	if cc != ccOK || len(presses) == 0 {
		s.powerMu.Unlock()
		return cc, nil
	}

	username := sess.username
	go func() {
		defer s.powerMu.Unlock()
		for _, p := range presses {
			log.Infof("ipmi: %q: chassis control %d: press %s for %s", username, action, p.button, p.hold)
			if err := s.deps.PressButton(p.button, p.hold); err != nil {
				log.Errorf("ipmi: chassis control %d: press %s: %s", action, p.button, err)
				return
			}
			time.Sleep(p.pause)
		}
		time.Sleep(pressSettle)
	}()
	return ccOK, nil
}
