package vm

import (
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
)

// Applying the whole USB section at once.
//
// Each switch used to rebuild the gadget on its own, so turning the disk off
// and the network on cost the host its keyboard twice, and the pair could only
// be reached through whatever order the budget allowed one step at a time.
// ApplyVirtualDevice takes every optional function the gadget should carry,
// checks the set against the budget as a whole, writes every marker, and
// rebuilds once.

// Every switch rebuilds the gadget through S03usbdev, on both kernels. It is
// the script that reads the markers; a mainline slot runs it too (ironkvm-dist
// #86). hid.GadgetScript may answer S00aagadget on an older mainline trial
// slot, and that script reads no marker, so a switch routed there would report
// success and change nothing. These name S03usbdev, and fail where it cannot
// run.
const (
	usbGadgetStop  = hid.USBDevScript + " stop"
	usbGadgetStart = hid.USBDevScript + " start"
)

// usbSelection is the set of optional functions the gadget carries, or should.
// network is a mode: off, ncm, ecm or rndis.
type usbSelection struct {
	console bool
	disk    bool
	audio   bool
	network string
}

// currentSelection reads the selection from the markers.
func currentSelection(present func(string) bool) usbSelection {
	return usbSelection{
		console: present(virtualConsole),
		disk:    present(virtualDisk),
		audio:   present(virtualAudio),
		network: usbNetworkMode(present),
	}
}

// has reports whether the selection carries the function with this table name.
func (s usbSelection) has(name string) bool {
	switch name {
	case "console":
		return s.console
	case "disk":
		return s.disk
	case "audio":
		return s.audio
	case "network":
		return s.network != usbNetworkOff
	default:
		return false
	}
}

// markers answers for the selection the way present answers for the board:
// each function's marker exists exactly when the selection carries it. Every
// other marker, disable_hid among them, is read from the board, so the HID
// cost is charged the same way usedEndpoints charges it now.
func (s usbSelection) markers(present func(string) bool) func(string) bool {
	return func(marker string) bool {
		switch marker {
		case virtualConsole:
			return s.console
		case virtualDisk:
			return s.disk
		case virtualAudio:
			return s.audio
		}

		for _, candidate := range usbNetworkModes {
			if marker == candidate.marker {
				return s.network == candidate.mode
			}
		}

		return present(marker)
	}
}

// adds reports whether the selection turns on anything the board does not
// carry now. A switch between NCM and ECM adds nothing, because both cost the
// same.
func (s usbSelection) adds(current usbSelection) bool {
	for _, function := range usbFunctions {
		if s.has(function.name) && !current.has(function.name) {
			return true
		}
	}

	return false
}

// selectionFits checks the whole selection against the budget, through the
// same usedEndpoints the per-device switches use.
//
// A selection that only gives things up always fits, as turning one function
// off does in UpdateVirtualDevice: a board already over its budget has to be
// able to shed what it cannot carry.
func selectionFits(wanted usbSelection, current usbSelection, present func(string) bool) (bool, string) {
	used := usedEndpoints(wanted.markers(present))
	if used.fitsIn(endpointBudget()) || !wanted.adds(current) {
		return true, ""
	}

	return false, selectionRefusal(wanted, used, present)
}

// selectionRefusal says how far over the budget the selection is, in the
// direction that is short, and names each selected function that would bring
// it inside on its own, cheapest first, as refusalMessage does for one switch.
func selectionRefusal(wanted usbSelection, used endpointUse, present func(string) bool) string {
	budget := endpointBudget()

	direction, needs, total := "inbound", used.in, budget.in
	if used.in <= budget.in {
		direction, needs, total = "outbound", used.out, budget.out
	}

	message := fmt.Sprintf("these devices need %d %s USB endpoints, %d available", needs, direction, total)

	var relief []usbFunction
	for _, function := range usbFunctions {
		if !wanted.has(function.name) {
			continue
		}

		if used.sub(function.cost).fitsIn(budget) {
			relief = append(relief, function)
		}
	}

	if len(relief) == 0 {
		return message
	}

	sort.SliceStable(relief, func(i, j int) bool {
		return relief[i].cost.in+relief[i].cost.out < relief[j].cost.in+relief[j].cost.out
	})

	options := make([]string, 0, len(relief))
	for _, function := range relief {
		options = append(options, function.name)
	}

	return message + ": turn off " + strings.Join(options, " or ")
}

// markerSteps is a per-device command list without the stop and start around
// it: the marker and symlink work alone, for a rebuild that stops and starts
// once for every function it changes.
func markerSteps(commands []string) []string {
	steps := make([]string, 0, len(commands))
	for _, command := range commands {
		if command == usbGadgetStop || command == usbGadgetStart {
			continue
		}

		steps = append(steps, command)
	}

	return steps
}

// selectionCommands is the single rebuild that takes the board from current to
// wanted: one stop, the marker work for every function that changes, and one
// start. It reuses the per-device command lists, so a function goes on and off
// here exactly as its own switch takes it, down to removing only the config
// symlink and removing markers with rm -f.
//
// It returns nil when the markers already match and the subnet has not changed
// under a running link: re-enumerating costs the host its keyboard for a few
// seconds, so an apply that changes nothing the gadget carries must not.
func selectionCommands(wanted usbSelection, current usbSelection, subnetChanged bool) []string {
	var steps []string

	toggle := func(device string, want bool, have bool) {
		if want == have {
			return
		}

		_, mount, unmount, _ := commandsFor(device)
		if want {
			steps = append(steps, markerSteps(mount)...)
		} else {
			steps = append(steps, markerSteps(unmount)...)
		}
	}

	toggle("console", wanted.console, current.console)
	toggle("disk", wanted.disk, current.disk)
	toggle("audio", wanted.audio, current.audio)

	if wanted.network != current.network {
		if wanted.network == usbNetworkOff {
			_, _, unmount, _ := commandsFor("network")
			steps = append(steps, markerSteps(unmount)...)
		} else {
			steps = append(steps, markerSteps(usbNetworkCommands(wanted.network))...)
		}
	}

	if len(steps) == 0 && !(subnetChanged && wanted.network != usbNetworkOff) {
		return nil
	}

	commands := append([]string{usbGadgetStop}, steps...)

	return append(commands, usbGadgetStart)
}

// runRebuild runs one rebuild and says whether every command succeeded.
//
// A failed stop ends it there: nothing has been touched, and the gadget is as
// it was. After a successful stop, start always runs, even when a marker step
// failed before it, because a gadget left stopped takes the host's keyboard
// away with no way back from the web UI. The steps after a failed one are
// skipped, and the ones before it are not undone, the same as the per-device
// switches: the markers end half changed, the board carries whatever they now
// say, and the response is an error so the UI reads the state back.
//
// That differs from the per-device lists in one way. Those stop at the first
// failure, start included, and depend on every step being unable to fail
// (rm -f) to reach start. Here start runs regardless.
func runRebuild(commands []string, run func(string) error) error {
	var failed error

	for i, command := range commands {
		last := i == len(commands)-1
		if failed != nil && !last {
			continue
		}

		if err := run(command); err != nil {
			log.Errorf("usb apply: %s: %s", command, err)
			if i == 0 {
				return err
			}
			if failed == nil {
				failed = err
			}
		}
	}

	return failed
}

func runShell(command string) error {
	return exec.Command("sh", "-c", command).Run()
}

// usbApplyEnv is what applying needs from the board, so that the whole of it
// can run in a test with nothing touched.
type usbApplyEnv struct {
	present     func(string) bool
	read        func(string) ([]byte, error)
	writeSubnet func(usbSubnet) error
	local       func() []netip.Prefix
	// rebuild runs one gadget rebuild. It is called at most once per apply.
	rebuild func([]string) error
}

// applyFailure is a refusal or an error, with the code the response carries.
type applyFailure struct {
	code    int
	message string
}

// applyUSBSelection validates the request against the board and applies it.
// Nothing is written until every check has passed.
//
// The subnet file is written before the rebuild, as SetUSBNetwork writes it,
// and is not put back if the rebuild then fails.
func applyUSBSelection(req proto.ApplyVirtualDeviceReq, env usbApplyEnv) *applyFailure {
	current := currentSelection(env.present)
	wanted := usbSelection{
		console: req.Console,
		disk:    req.Disk,
		audio:   req.Audio,
		network: req.Network.Mode,
	}

	// RNDIS is read back from a board an older server set up, and nothing here
	// writes it. Asking for it keeps it, and only where it already runs.
	if wanted.network == usbNetworkRNDIS && current.network != usbNetworkRNDIS {
		return &applyFailure{-2, "rndis is no longer offered, choose ncm or ecm"}
	}

	currentSubnet := readUSBSubnet(env.read)
	subnet := currentSubnet

	if req.Network.Subnet != "" {
		parsed, err := parseUSBSubnet(req.Network.Subnet)
		if err != nil {
			return &applyFailure{-2, err.Error()}
		}

		if prefix, overlaps := usbSubnetOverlap(parsed, env.local()); overlaps {
			return &applyFailure{-2, fmt.Sprintf("%s overlaps %s, which the board already uses", parsed.prefix, prefix)}
		}

		subnet = parsed
	}

	// Refuse rather than drop something, as the per-device switches do.
	if ok, refusal := selectionFits(wanted, current, env.present); !ok {
		log.Infof("refused usb apply: %s", refusal)
		return &applyFailure{-4, refusal}
	}

	subnetChanged := subnet.prefix != currentSubnet.prefix
	if subnetChanged {
		if err := env.writeSubnet(subnet); err != nil {
			log.Errorf("write %s: %s", usbNetworkSubnetFile, err)
			return &applyFailure{-3, "operation failed"}
		}
	}

	commands := selectionCommands(wanted, current, subnetChanged)
	if commands == nil {
		return nil
	}

	if err := env.rebuild(commands); err != nil {
		return &applyFailure{-3, "operation failed"}
	}

	return nil
}

// rebuildWithHidClosed runs one rebuild with the HID devices closed around it,
// as UpdateVirtualDevice does.
func rebuildWithHidClosed(commands []string) error {
	h := hid.GetHid()
	h.Lock()
	h.CloseNoLock()
	h.ForgetAcceptingNoLock()
	defer func() {
		h.OpenNoLock()
		h.Unlock()
	}()

	return runRebuild(commands, runShell)
}

func (s *Service) ApplyVirtualDevice(c *gin.Context) {
	var req proto.ApplyVirtualDeviceReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid argument")
		return
	}

	failure := applyUSBSelection(req, usbApplyEnv{
		present:     markerPresent,
		read:        os.ReadFile,
		writeSubnet: writeUSBSubnet,
		local:       localPrefixes,
		rebuild:     rebuildWithHidClosed,
	})
	if failure != nil {
		rsp.ErrRsp(c, failure.code, failure.message)
		return
	}

	rsp.OkRspWithData(c, &proto.ApplyVirtualDeviceRsp{
		GetVirtualDeviceRsp: *virtualDeviceState(markerPresent),
		USBNetwork:          usbNetworkState(markerPresent, os.ReadFile),
	})

	log.Debugf("apply virtual devices success")
}
