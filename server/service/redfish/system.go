package redfish

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	shortPress = 800 * time.Millisecond
	longPress  = 5 * time.Second

	resetAction = "ComputerSystem.Reset"
)

// resetTypes are the ResetType values the board can carry out. PowerCycle,
// Nmi and GracefulRestart are left out: two buttons and an LED cannot do them.
var resetTypes = []string{"On", "ForceOff", "GracefulShutdown", "ForceRestart", "PushPowerButton"}

// blindResetTypes press the same button whatever the host's state, so they
// are all a board without its power LED wired can offer.
var blindResetTypes = []string{"ForceRestart", "PushPowerButton"}

// offeredResetTypes are the ResetType values the system lists as allowed.
func (s *Service) offeredResetTypes() []string {
	if !s.deps.PowerLEDConnected() {
		return blindResetTypes
	}
	return resetTypes
}

var (
	errResetTypeNotAllowed = errors.New("reset type not allowed")
	errPowerStateUnknown   = errors.New("the power state is unknown")
)

// press is one button press.
type press struct {
	button string
	hold   time.Duration
}

// planReset maps a ResetType and the power LED to the press it needs. on is
// nil when the LED cannot be read. A nil press with no error means the host
// is already in the state asked for, and nothing is pressed.
func planReset(resetType string, on *bool) (*press, error) {
	switch resetType {
	case "ForceRestart":
		return &press{ButtonReset, shortPress}, nil
	case "PushPowerButton":
		return &press{ButtonPower, shortPress}, nil
	case "On", "ForceOff", "GracefulShutdown":
	default:
		return nil, errResetTypeNotAllowed
	}

	if on == nil {
		return nil, errPowerStateUnknown
	}

	switch {
	case resetType == "On" && !*on:
		return &press{ButtonPower, shortPress}, nil
	case resetType == "ForceOff" && *on:
		return &press{ButtonPower, longPress}, nil
	case resetType == "GracefulShutdown" && *on:
		return &press{ButtonPower, shortPress}, nil
	}
	return nil, nil
}

// powerLED reads the LED, or returns nil when the state is unknown: the
// owner has not said the LED is wired, or its line cannot be read. An LED
// that is not wired reads "off" whatever the host does.
func (s *Service) powerLED() *bool {
	if !s.deps.PowerLEDConnected() {
		return nil
	}
	on, err := s.deps.PowerLED()
	if err != nil {
		return nil
	}
	return &on
}

// powerState is the PowerState property: "On", "Off", or null when unknown.
func powerState(on *bool) any {
	if on == nil {
		return nil
	}
	if *on {
		return "On"
	}
	return "Off"
}

func systems(c *gin.Context) {
	writeJSON(c, http.StatusOK, newCollection(systemsPath, "ComputerSystemCollection", "Computer System Collection", []string{systemPath}))
}

func (s *Service) system(c *gin.Context) {
	body := newResource(systemPath, "ComputerSystem.v1_13_0.ComputerSystem")
	body["Id"] = "1"
	body["Name"] = "Managed host"
	body["SystemType"] = "Physical"
	body["PowerState"] = powerState(s.powerLED())
	body["VirtualMedia"] = link(systemMediaPath)
	body["Links"] = object{
		"Chassis":   links(chassisPath),
		"ManagedBy": links(managerPath),
	}
	body["Actions"] = object{
		"#" + resetAction: object{
			"target":                            resetPath,
			"ResetType@Redfish.AllowableValues": s.offeredResetTypes(),
		},
	}

	writeTagged(c, body)
}

// resetSettle is how long a reset keeps the lock after a press, so the next
// one reads an LED that has had time to follow. Tests shorten it.
var resetSettle = 2 * time.Second

// reset answers 204 once the press is done. It does not wait for the LED to
// follow: an ATX host can take seconds, and clients poll PowerState.
func (s *Service) reset(c *gin.Context) {
	params, ok := decodeParams(c)
	if !ok {
		return
	}
	for name := range params {
		if name != "ResetType" {
			writeError(c, http.StatusBadRequest, "ActionParameterNotSupported", "", name, resetAction)
			return
		}
	}
	if !present(params, "ResetType") {
		writeError(c, http.StatusBadRequest, "ActionParameterMissing", "", resetAction, "ResetType")
		return
	}

	var resetType string
	if err := json.Unmarshal(params["ResetType"], &resetType); err != nil {
		writeError(c, http.StatusBadRequest, "ActionParameterValueTypeError", "", string(params["ResetType"]), "ResetType", resetAction)
		return
	}

	// The LED is read, the press planned and made, and the host given time
	// to follow, all under one lock: two ForceOff calls at once would
	// otherwise both see the host on, and the second would turn it back on.
	s.resetMu.Lock()
	defer s.resetMu.Unlock()

	p, err := planReset(resetType, s.powerLED())
	if errors.Is(err, errPowerStateUnknown) && !s.deps.PowerLEDConnected() {
		writeError(c, http.StatusBadRequest, "ActionNotSupported",
			"the power LED is not connected, so the power state is unknown and "+resetType+
				" cannot tell whether to press; use PushPowerButton or ForceRestart, or, if the LED header is wired, "+
				"turn on hardware.powerLed (Power LED connected, in the web UI's power menu)", resetAction)
		return
	}
	switch {
	case errors.Is(err, errResetTypeNotAllowed):
		writeError(c, http.StatusBadRequest, "ActionParameterValueNotInList", "", resetType, "ResetType", resetAction)
		return
	case errors.Is(err, errPowerStateUnknown):
		writeError(c, http.StatusConflict, "ActionNotSupported",
			"the power LED cannot be read, so "+resetType+" cannot tell whether to press; use PushPowerButton or ForceRestart", resetAction)
		return
	}

	if p != nil {
		if err := s.deps.PressButton(p.button, p.hold); err != nil {
			log.Errorf("redfish: %s: press %s: %s", resetType, p.button, err)
			writeError(c, http.StatusInternalServerError, "GeneralError", "the button press failed: "+err.Error())
			return
		}
		time.Sleep(resetSettle)
	}

	c.Status(http.StatusNoContent)
}
