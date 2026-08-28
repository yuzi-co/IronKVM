package vm

import (
	"reflect"
	"strings"
	"testing"
)

// presence builds the marker probe the budget functions take, so no test
// touches the filesystem.
func presence(markers ...string) func(string) bool {
	set := make(map[string]bool, len(markers))
	for _, marker := range markers {
		set[marker] = true
	}

	return func(path string) bool { return set[path] }
}

func TestUsedEndpointsCountsHidAndEachFunctionOnce(t *testing.T) {
	for _, test := range []struct {
		name    string
		markers []string
		want    endpointUse
	}{
		{"nothing but hid", nil, endpointUse{in: 3, out: 3}},
		{"hid and the console", []string{virtualConsole}, endpointUse{in: 5, out: 4}},
		{"hid, console and audio", []string{virtualConsole, virtualAudio}, endpointUse{in: 5, out: 5}},
		{"hid, console and disk", []string{virtualConsole, virtualDisk}, endpointUse{in: 6, out: 5}},
		{"disk and network without the console", []string{virtualDisk, virtualNetwork}, endpointUse{in: 6, out: 5}},
		{"hid, console and network", []string{virtualConsole, virtualNetwork}, endpointUse{in: 7, out: 5}},
		{"everything at once", []string{virtualConsole, virtualDisk, virtualNetwork, virtualAudio}, endpointUse{in: 8, out: 7}},
	} {
		if got := usedEndpoints(presence(test.markers...)); got != test.want {
			t.Errorf("%s: used %+v endpoints, want %+v", test.name, got, test.want)
		}
	}
}

// The shipped gadget sits exactly on the inbound ceiling, and the ceiling was
// read from the controller. If either number moves without the other, the
// board runs one endpoint short and says nothing about it.
func TestTheShippedConfigurationIsExactlyTheInboundBudget(t *testing.T) {
	used := usedEndpoints(presence(virtualConsole, virtualDisk, virtualAudio))

	if used.in != DefaultInEndpointBudget {
		t.Errorf("console + disk + speaker uses %d inbound endpoints, want exactly %d", used.in, DefaultInEndpointBudget)
	}

	if used.out > DefaultOutEndpointBudget {
		t.Errorf("console + disk + speaker uses %d outbound endpoints of %d", used.out, DefaultOutEndpointBudget)
	}
}

// usb.ncm and usb.rndis0 are alternatives for one function. Counting both
// would reserve three endpoints that nothing uses, and the guard would then
// refuse a function that fits.
func TestUsedEndpointsCountsTheNetworkOnce(t *testing.T) {
	both := usedEndpoints(presence(virtualNetworkNCM, virtualNetwork))
	one := usedEndpoints(presence(virtualNetwork))

	if both != one {
		t.Errorf("two network markers cost %d, one costs %d; want the same", both, one)
	}
}

// A board with HID disabled has three more endpoints to spend. Charging for
// hardware that is not there would drop functions that fit.
func TestHidCostsNothingWhenItIsDisabled(t *testing.T) {
	if got := usedEndpoints(presence(disableHid)); got != (endpointUse{}) {
		t.Errorf("used %+v endpoints with hid disabled, want none", got)
	}
}

func TestCanEnableAllowsWhatFits(t *testing.T) {
	// hid(3 in) + console(2 in) is 5 of 6, and the speaker costs nothing
	// inbound.
	ok, free, _ := canEnable("audio", presence(virtualConsole))

	if !ok {
		t.Error("refused the speaker, which costs no inbound endpoint at all")
	}

	if free != (endpointUse{in: 1, out: 3}) {
		t.Errorf("reported %+v free, want {in:1 out:3}", free)
	}
}

// The speaker is the one function that can always be switched on. It is an
// isochronous OUT stream and nothing else, so it never competes for the
// direction that runs out.
func TestCanEnableAlwaysAllowsTheSpeaker(t *testing.T) {
	for _, markers := range [][]string{
		nil,
		{virtualConsole},
		{virtualConsole, virtualDisk},
		{virtualDisk, virtualNetwork},
	} {
		if ok, free, _ := canEnable("audio", presence(markers...)); !ok {
			t.Errorf("refused the speaker with %v enabled and %+v free", markers, free)
		}
	}
}

// The console and the network are two inbound endpoints each, HID is three,
// and the controller has six dedicated transmit FIFOs. The pair cannot work.
//
// A flat budget of nine allowed it, because nine is what the pair costs when
// both directions are added together. Nothing complains at boot: endpoint
// structures are handed out at bind time and there are seven of those, so the
// gadget binds and the UDC reports "configured". The shortage only appears
// when the host sets the configuration, and one interrupt IN endpoint fails to
// enable with nothing said about it.
func TestCanEnableRefusesTheConsoleAndNetworkTogether(t *testing.T) {
	ok, free, relief := canEnable("network", presence(virtualConsole))

	if ok {
		t.Fatal("allowed the network alongside the console, which needs a seventh inbound endpoint")
	}

	if free.in != 1 {
		t.Errorf("reported %d inbound free, want 1", free.in)
	}

	if !reflect.DeepEqual(relief, []string{"console"}) {
		t.Errorf("suggested %v, want [console]", relief)
	}
}

func TestCanEnableRefusesWhatDoesNotFit(t *testing.T) {
	// hid(3 in) + console(2 in) + disk(1 in) is 6 of 6, so nothing inbound is
	// left and the network needs two.
	ok, free, relief := canEnable("network", presence(virtualConsole, virtualDisk))

	if ok {
		t.Error("allowed the network with no inbound endpoint free")
	}

	if free != (endpointUse{in: 0, out: 2}) {
		t.Errorf("reported %+v free, want {in:0 out:2}", free)
	}

	// Naming something that would not free enough is worse than naming
	// nothing: the operator turns it off and is refused again. The disk frees
	// one inbound endpoint of the two that are short, so it is not named.
	if !reflect.DeepEqual(relief, []string{"console"}) {
		t.Errorf("suggested %v, want [console]", relief)
	}
}

// Every suggestion has to actually make room. Naming one that does not is worse
// than naming none: the operator turns it off and is refused again, and learns
// the rule by exhaustion.
func TestCanEnableOnlySuggestsFunctionsThatFreeEnough(t *testing.T) {
	_, free, relief := canEnable("network", presence(virtualConsole, virtualDisk))

	wanted, ok := endpointCost("network")
	if !ok {
		t.Fatal("the network is missing from the table")
	}

	needed := shortfall(wanted, free)

	if len(relief) == 0 {
		t.Fatal("refused the network without suggesting anything")
	}

	for _, name := range relief {
		cost, ok := endpointCost(name)
		if !ok {
			t.Fatalf("suggested %q, which is not a function", name)
		}

		if cost.in < needed.in || cost.out < needed.out {
			t.Errorf("suggested %q, which frees %+v of the %+v needed", name, cost, needed)
		}
	}
}

// console(2 in) + disk(1 in) + audio(0 in) + hid(3 in) is exactly six, and the
// network needs both of its inbound endpoints back. The disk frees one and the
// speaker frees none, so only the console can supply them on its own. A build
// with no filter would answer [audio disk console] and send the operator to
// turn off a speaker that frees nothing at all in the direction that is short.
//
// This filter has already been deleted once during this task. It stays tested.
func TestCanEnableWillNotSuggestAFunctionThatIsTooSmall(t *testing.T) {
	ok, free, relief := canEnable("network", presence(virtualConsole, virtualDisk, virtualAudio))

	if ok {
		t.Fatal("allowed the network at a full budget")
	}

	if free.in != 0 {
		t.Fatalf("reported %d inbound free, want 0", free.in)
	}

	if !reflect.DeepEqual(relief, []string{"console"}) {
		t.Errorf("suggested %v, want [console] - the only one that frees two inbound", relief)
	}
}

func TestEndpointCostRejectsUnknownNames(t *testing.T) {
	if _, ok := endpointCost("speaker"); ok {
		t.Error("endpointCost accepted a name it does not know")
	}
}

// The console claims two of the six inbound endpoints - the largest single
// share after HID - so it needs a switch like the rest. A budget display that shows
// the operator a full bar while offering no way to free the biggest consumer
// states the problem and withholds the answer.
func TestConsoleIsTogglableLikeTheOthers(t *testing.T) {
	if _, ok := endpointCost("console"); !ok {
		t.Error("the console is missing from the table")
	}

	function, ok := functionForDevice("console")
	if !ok {
		t.Fatal("no function answers to the device name \"console\"")
	}

	if function.markers[0] != virtualConsole {
		t.Errorf("the console is gated on %q, want %q", function.markers[0], virtualConsole)
	}
}

func TestPriorityOrderIsAudioFirstConsoleLast(t *testing.T) {
	var order []string
	for _, function := range dropOrder() {
		order = append(order, function.name)
	}

	want := []string{"audio", "network", "disk", "console"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("drop order is %v, want %v", order, want)
	}
}

// The refusal is the whole interactive experience of this feature. "Operation
// failed" would leave the operator exactly where they were before it existed:
// switching things at random and losing HID.
func TestRefusalMessageNamesTheNumbersAndTheWayOut(t *testing.T) {
	message := refusalMessage("network", endpointUse{in: 0, out: 1}, []string{"console"})

	// The direction is part of the answer. A refusal that reports the wrong
	// one cannot be checked against the controller.
	for _, want := range []string{"network", "2", "inbound", "0 free", "console"} {
		if !strings.Contains(message, want) {
			t.Errorf("refusal %q does not mention %q", message, want)
		}
	}
}

// A function that fits inbound and not outbound has to be told so, or the
// operator reads the inbound bar, sees room, and cannot explain the refusal.
func TestRefusalMessageNamesTheOutboundDirectionWhenThatIsWhatIsShort(t *testing.T) {
	message := refusalMessage("audio", endpointUse{in: 2, out: 0}, nil)

	if !strings.Contains(message, "outbound") {
		t.Errorf("refusal %q does not say which direction is short", message)
	}
}

func TestRefusalMessageWithNothingToSuggest(t *testing.T) {
	message := refusalMessage("network", endpointUse{}, nil)

	if strings.Contains(message, "turn off") {
		t.Errorf("refusal %q offers a way out when there is none", message)
	}

	if !strings.Contains(message, "network") {
		t.Errorf("refusal %q does not name the device", message)
	}
}

// Every entry in the table that has a device name must be reachable through the
// toggle, and every name the toggle accepts must be in the table. A name in one
// and not the other is a switch that reports success and changes nothing, or a
// function the budget cannot see.
//
// This lives here rather than with the table in Task 1 because it asserts an
// agreement between two files, and the second half of that agreement - the
// console's entry in commandsFor - is added by Step 5 below.
func TestEveryTogglableFunctionHasCommands(t *testing.T) {
	for _, function := range usbFunctions {
		if function.device == "" {
			t.Errorf("%s has no device name, so nothing can switch it", function.name)
			continue
		}

		if _, _, _, ok := commandsFor(function.device); !ok {
			t.Errorf("commandsFor does not know %q", function.device)
		}
	}
}

// The gadget path is what the API reports as active, so a wrong name would
// report every function dead and the UI would warn about all of them. The
// network's names come straight from the `ln -s` targets in S03usbdev: NCM is
// the primary and RNDIS is the fallback S03usbdev builds when only the RNDIS
// marker is set.
func TestEveryFunctionNamesItsGadgetDirectory(t *testing.T) {
	want := map[string]string{
		"console": "acm.GS0",
		"disk":    "mass_storage.disk0",
		"network": "ncm.usb0",
		"audio":   "uac1.usb0",
	}
	wantAlt := map[string]string{
		"network": "rndis.usb0",
	}

	seen := make(map[string]bool, len(usbFunctions))

	for _, function := range usbFunctions {
		seen[function.name] = true

		gadget, ok := want[function.name]
		if !ok {
			continue
		}

		if function.gadget != gadget {
			t.Errorf("%s links %q, want %q", function.name, function.gadget, gadget)
		}

		if alt, ok := wantAlt[function.name]; ok && function.gadgetAlt != alt {
			t.Errorf("%s falls back to %q, want %q", function.name, function.gadgetAlt, alt)
		}
	}

	// A function dropped from usbFunctions would otherwise leave its entry in
	// want unconsulted, and the test above would pass without ever noticing.
	for name := range want {
		if !seen[name] {
			t.Errorf("%q is missing from usbFunctions", name)
		}
	}
}

// active takes an injected predicate specifically so it can be tested without
// the real configfs tree. These three cases are the ones that matter: the
// primary directory alone, only the alternate, and neither - which is the one
// a deleted or mis-wired gadgetAlt branch would get wrong.
func TestActiveIsTrueWithOnlyThePrimaryDirectoryLinked(t *testing.T) {
	function := usbFunction{gadget: "ncm.usb0", gadgetAlt: "rndis.usb0"}

	if !function.active(presence("ncm.usb0")) {
		t.Error("not active with the primary gadget directory linked")
	}
}

func TestActiveIsTrueWithOnlyTheAlternateDirectoryLinked(t *testing.T) {
	function := usbFunction{gadget: "ncm.usb0", gadgetAlt: "rndis.usb0"}

	if !function.active(presence("rndis.usb0")) {
		t.Error("not active with only the alternate gadget directory linked")
	}
}

func TestActiveIsFalseWithNeitherDirectoryLinked(t *testing.T) {
	function := usbFunction{gadget: "ncm.usb0", gadgetAlt: "rndis.usb0"}

	if function.active(presence()) {
		t.Error("active with neither gadget directory linked")
	}
}
