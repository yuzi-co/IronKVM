package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The USB controller limits the two directions separately, and the inbound
// direction is the one that runs out.
//
// dwc2 announces
//
//	dwc2 4340000.usb: EPs: 8, dedicated fifos, 3072 entries in SPRAM
//
// and neither number is a budget. The real limits are readable from the
// controller, and were read from the running board on 2026-08-27:
//
//	# cat /sys/kernel/debug/usb/4340000.usb/hw_params
//	num_dev_ep                    : 7
//	total_fifo_size               : 3072
//	# cat /sys/kernel/debug/usb/4340000.usb/fifo
//	RXFIFO: Size 536
//	NPTXFIFO: Size 32, Start 0x00000218
//	Periodic TXFIFOs:
//		DPTXFIFO 1: Size 768, Start 0x00000238
//		DPTXFIFO 2: Size 512, Start 0x00000538
//		DPTXFIFO 3: Size 512, Start 0x00000738
//		DPTXFIFO 4: Size 384, Start 0x00000938
//		DPTXFIFO 5: Size 128, Start 0x00000ab8
//		DPTXFIFO 6: Size 128, Start 0x00000b38
//
// So the controller has seven endpoint pairs beside ep0, and six dedicated
// transmit FIFOs. In dedicated FIFO mode dwc2_hsotg_ep_enable refuses an IN
// endpoint it cannot give a FIFO of its own to, so six is the inbound ceiling
// and seven is the outbound one. There is no room for a seventh FIFO either:
// 536 + 32 + 768 + 512 + 512 + 384 + 128 + 128 is 3000 of the 3072 words the
// controller has.
//
// The flat budget of nine that stood here before was the sum of both
// directions, and it was fitted to three configurations that were tried on
// hardware. It got two of them right by arithmetic accident. The one it got
// wrong is the console and the network together, which asks for seven inbound
// endpoints and can only be given six.
//
// That case does not announce itself where the old rule was tested.
// usb_ep_autoconfig hands out endpoint structures at bind time and there are
// seven of those, so the gadget binds. The FIFO shortage appears later, when
// the host sets the configuration and dwc2 has no free transmit FIFO left for
// the seventh IN endpoint.
//
// Built on the board on 2026-08-28, console and network together with the
// three HID functions:
//
//	dwc2 4340000.usb: bound driver configfs-gadget
//	dwc2 4340000.usb: new device is high-speed
//	dwc2 4340000.usb: new address 6
//	dwc2 4340000.usb: dwc2_hsotg_ep_enable: No suitable fifo found
//	dwc2 4340000.usb: dwc2_hsotg_ep_enable: No suitable fifo found
//
// Nothing else showed it. The UDC reported "configured", /dev/hidg0 through
// /dev/hidg2 and /dev/ttyGS0 were all present, and usb0 was up. DIEPCTL had six
// IN endpoints active holding all six FIFOs, and a seventh carrying a
// programmed maximum packet size with its USBACTEP bit clear.
//
// Which endpoint loses is whichever is enabled seventh, not a fixed one. In
// that run the HID endpoints kept their places and a bulk pair did not.
//
// Read on the running board with the shipped configuration - console, disk,
// speaker and all three HID functions - every one of the six FIFOs is seated
// and ep7in is inactive. The shipped gadget sits exactly on the ceiling.
const (
	DefaultInEndpointBudget  = 6
	DefaultOutEndpointBudget = 7
)

// hidInCost and hidOutCost are what the keyboard, the relative mouse and the
// absolute pointer cost together.
//
// Each one is an interrupt IN endpoint and an interrupt OUT endpoint. The OUT
// half is easy to overlook because nothing here writes to it: f_hid on 5.10
// allocates it whether or not the report descriptor has output reports, and it
// was read as active on the board.
const (
	hidInCost  = 3
	hidOutCost = 3
)

// endpointUse is a count of endpoints in each direction. Budgets and costs are
// both this shape, because a plan fits only when both directions do.
type endpointUse struct {
	in  int
	out int
}

func (u endpointUse) add(v endpointUse) endpointUse {
	return endpointUse{in: u.in + v.in, out: u.out + v.out}
}

func (u endpointUse) sub(v endpointUse) endpointUse {
	return endpointUse{in: u.in - v.in, out: u.out - v.out}
}

// fitsIn reports whether this much use stays inside the limit.
func (u endpointUse) fitsIn(limit endpointUse) bool {
	return u.in <= limit.in && u.out <= limit.out
}

// The markers that decide which functions the gadget carries. usb.ncm and
// usb.rndis0 are alternatives for one function - S03usbdev prefers NCM - so
// they belong to a single entry and are counted once.
const (
	virtualConsole    = "/boot/usb.acm"
	virtualNetworkNCM = "/boot/usb.ncm"
	disableHid        = "/boot/disable_hid"
)

// usbFunction is one optional gadget function.
//
// device is the name the API accepts to switch it. gadget is the configfs
// directory that proves the function actually linked; gadgetAlt is a second
// accepted directory for a function with two forms (the network's NCM and
// RNDIS). priority decides what survives when more is enabled than fits:
// higher survives longer.
//
// cost is what the function takes in each direction. The inbound half decides
// nearly everything, because the controller has six inbound endpoints and
// seven outbound ones.
type usbFunction struct {
	name      string
	device    string
	markers   []string
	gadget    string
	gadgetAlt string
	cost      endpointUse
	priority  int
}

// The console outranks everything except HID because it is the only way into a
// board whose network is gone. Audio is last because it is the only entry that
// costs nothing to lose.
var usbFunctions = []usbFunction{
	// f_acm and the two network functions each take a bulk pair and an
	// interrupt IN for notifications, so two inbound and one outbound.
	// f_mass_storage takes a bulk pair. The speaker is a playback stream
	// alone, which is one isochronous OUT and no inbound endpoint at all, so
	// it never competes for the scarce direction.
	{name: "console", device: "console", markers: []string{virtualConsole}, gadget: "acm.GS0", cost: endpointUse{in: 2, out: 1}, priority: 40},
	{name: "disk", device: "disk", markers: []string{virtualDisk}, gadget: "mass_storage.disk0", cost: endpointUse{in: 1, out: 1}, priority: 30},
	{name: "network", device: "network", markers: []string{virtualNetworkNCM, virtualNetwork}, gadget: "ncm.usb0", gadgetAlt: "rndis.usb0", cost: endpointUse{in: 2, out: 1}, priority: 20},
	{name: "audio", device: "audio", markers: []string{virtualAudio}, gadget: "uac1.usb0", cost: endpointUse{in: 0, out: 1}, priority: 10},
}

// endpointBudget is what the controller fits, in each direction.
func endpointBudget() endpointUse {
	return endpointUse{in: DefaultInEndpointBudget, out: DefaultOutEndpointBudget}
}

// enabled reports whether any of the function's markers is present.
func (f usbFunction) enabled(present func(string) bool) bool {
	for _, marker := range f.markers {
		if present(marker) {
			return true
		}
	}

	return false
}

// hidEndpointCost charges nothing when HID is switched off, because those
// endpoints are then genuinely free. Charging for them would refuse a function
// that fits.
func hidEndpointCost(present func(string) bool) endpointUse {
	if present(disableHid) {
		return endpointUse{}
	}

	return endpointUse{in: hidInCost, out: hidOutCost}
}

// usedEndpoints totals HID and every enabled function.
func usedEndpoints(present func(string) bool) endpointUse {
	used := hidEndpointCost(present)

	for _, function := range usbFunctions {
		if function.enabled(present) {
			used = used.add(function.cost)
		}
	}

	return used
}

// endpointCost reports what one function costs, by its table name.
func endpointCost(name string) (endpointUse, bool) {
	for _, function := range usbFunctions {
		if function.name == name {
			return function.cost, true
		}
	}

	return endpointUse{}, false
}

// functionForDevice finds the table entry an API device name refers to.
func functionForDevice(device string) (usbFunction, bool) {
	if device == "" {
		return usbFunction{}, false
	}

	for _, function := range usbFunctions {
		if function.device == device {
			return function, true
		}
	}

	return usbFunction{}, false
}

// dropOrder lists the optional functions lowest priority first, which is the
// order they are given up in.
func dropOrder() []usbFunction {
	order := make([]usbFunction, len(usbFunctions))
	copy(order, usbFunctions)

	sort.SliceStable(order, func(i, j int) bool {
		return order[i].priority < order[j].priority
	})

	return order
}

// shortfall is how much of each direction a plan is over by, never negative.
func shortfall(cost endpointUse, free endpointUse) endpointUse {
	short := endpointUse{in: cost.in - free.in, out: cost.out - free.out}
	if short.in < 0 {
		short.in = 0
	}
	if short.out < 0 {
		short.out = 0
	}

	return short
}

// canEnable reports whether one more function fits, how many endpoints are
// free in each direction, and which enabled functions would free enough on
// their own.
//
// The suggestions are the point: an operator told only "no room" turns
// something off, is refused again, and learns the rule by exhaustion. Only
// functions that would actually make room are named, cheapest first, so the
// operator gives up as little as possible. A function has to cover the
// shortfall in both directions to be worth naming, because turning off
// something that only frees the direction that was not short refuses them
// twice.
func canEnable(device string, present func(string) bool) (bool, endpointUse, []string) {
	wanted, ok := functionForDevice(device)
	if !ok {
		return false, endpointUse{}, nil
	}

	free := endpointBudget().sub(usedEndpoints(present))

	if wanted.enabled(present) || wanted.cost.fitsIn(free) {
		return true, free, nil
	}

	needed := shortfall(wanted.cost, free)

	candidates := make([]usbFunction, 0, len(usbFunctions))
	for _, function := range usbFunctions {
		if function.name == wanted.name || !function.enabled(present) {
			continue
		}

		if function.cost.in >= needed.in && function.cost.out >= needed.out {
			candidates = append(candidates, function)
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].cost.in+candidates[i].cost.out < candidates[j].cost.in+candidates[j].cost.out
	})

	relief := make([]string, 0, len(candidates))
	for _, function := range candidates {
		relief = append(relief, function.name)
	}

	return false, free, relief
}

// gadgetConfigPath is where configfs records the functions this gadget carries.
// A symlink here means the function was built; a marker only means it was
// asked for, and the two differ whenever the budget dropped something.
const gadgetConfigPath = "/sys/kernel/config/usb_gadget/g0/configs/c.1"

// active reports whether the function is linked into the running gadget.
func (f usbFunction) active(linked func(string) bool) bool {
	if linked(f.gadget) {
		return true
	}

	return f.gadgetAlt != "" && linked(f.gadgetAlt)
}

// isFunctionActive answers the same question against the real configfs.
func isFunctionActive(name string) bool {
	for _, function := range usbFunctions {
		if function.name != name {
			continue
		}

		return function.active(func(dir string) bool {
			_, err := os.Lstat(filepath.Join(gadgetConfigPath, dir))
			return err == nil
		})
	}

	return false
}

// refusalMessage tells the operator what was refused, how short the budget is,
// and what would make room. Naming a function that would not free enough is
// worse than naming none: they turn it off and are refused again.
//
// It names the direction, because the two limits differ and a refusal that
// reports the wrong one cannot be checked against anything.
func refusalMessage(device string, free endpointUse, relief []string) string {
	wanted, ok := functionForDevice(device)
	if !ok {
		return "unknown device"
	}

	direction, needs, spare := "inbound", wanted.cost.in, free.in
	if wanted.cost.in <= free.in {
		direction, needs, spare = "outbound", wanted.cost.out, free.out
	}

	message := fmt.Sprintf("%s needs %d %s USB endpoints, %d free", device, needs, direction, spare)

	if len(relief) == 0 {
		return message
	}

	options := make([]string, 0, len(relief))
	for _, name := range relief {
		cost, _ := endpointCost(name)
		frees := cost.in
		if direction == "outbound" {
			frees = cost.out
		}
		options = append(options, fmt.Sprintf("%s (%d)", name, frees))
	}

	return message + " — turn off " + strings.Join(options, " or ") + " first"
}
