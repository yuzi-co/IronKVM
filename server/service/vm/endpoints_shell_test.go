package vm

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const initScript = "../../../kvmapp/system/init.d/S03usbdev"

func readInitScript(t *testing.T) string {
	t.Helper()

	body, err := os.ReadFile(filepath.FromSlash(initScript))
	if err != nil {
		t.Fatalf("cannot read %s: %s", initScript, err)
	}

	return string(body)
}

// shellCosts pulls the case arms out of usb_cost:
//
//	console) echo 3 ;;
func shellCosts(t *testing.T, script string, fn string) map[string]int {
	t.Helper()

	start := strings.Index(script, fn+"() {")
	if start < 0 {
		t.Fatalf("S03usbdev has no %s function", fn)
	}

	end := strings.Index(script[start:], "\n}")
	if end < 0 {
		t.Fatalf("%s is not terminated", fn)
	}

	arm := regexp.MustCompile(`(?m)^\s*([a-z]+}?\))\s*echo\s+(\d+)`)
	costs := make(map[string]int)

	for _, match := range arm.FindAllStringSubmatch(script[start:start+end], -1) {
		name := strings.TrimSuffix(match[1], ")")
		if name == "*" {
			continue
		}

		value, err := strconv.Atoi(match[2])
		if err != nil {
			t.Fatalf("%s gives %q a non-numeric cost %q", fn, name, match[2])
		}

		costs[name] = value
	}

	return costs
}

// The costs decide whether a set fits. If the two copies disagree the guard
// permits a set the boot script then refuses to build, or the other way round,
// and the symptom is the silent HID loss this whole feature exists to prevent.
func TestShellAndGoAgreeOnEveryCost(t *testing.T) {
	script := readInitScript(t)

	for _, direction := range []struct {
		fn   string
		want func(usbFunction) int
	}{
		{"usb_in_cost", func(f usbFunction) int { return f.cost.in }},
		{"usb_out_cost", func(f usbFunction) int { return f.cost.out }},
	} {
		costs := shellCosts(t, script, direction.fn)

		if len(costs) != len(usbFunctions) {
			t.Errorf("%s names %d functions, the Go table has %d: %v vs %v",
				direction.fn, len(costs), len(usbFunctions), costs, usbFunctions)
		}

		for _, function := range usbFunctions {
			shell, ok := costs[function.name]
			if !ok {
				t.Errorf("%s has no arm for %q", direction.fn, function.name)
				continue
			}

			if shell != direction.want(function) {
				t.Errorf("%s costs %d in %s and %d in Go", function.name, shell, direction.fn, direction.want(function))
			}
		}
	}
}

// The order decides what is given up. Disagreement here loses the wrong
// function, and the one that matters is the console: it is the only way into a
// board whose network is gone.
func TestShellAndGoAgreeOnTheDropOrder(t *testing.T) {
	script := readInitScript(t)

	order := regexp.MustCompile(`usb_drop_order\(\) \{\s*echo "([^"]+)"`).FindStringSubmatch(script)
	if order == nil {
		t.Fatal("S03usbdev has no usb_drop_order function")
	}

	shell := strings.Fields(order[1])

	var goOrder []string
	for _, function := range dropOrder() {
		goOrder = append(goOrder, function.name)
	}

	if !reflect.DeepEqual(shell, goOrder) {
		t.Errorf("S03usbdev drops %v, Go drops %v", shell, goOrder)
	}
}

// usb_resolve - the function that actually decides what gets built at boot -
// reads usb_keep_order, not usb_drop_order. dropOrder() is otherwise used
// only by tests, so pinning just the drop order leaves the order that decides
// boot behaviour checked only transitively, through
// tools/service/test-usb-endpoints.sh asserting that the keep order and the
// drop order are each other's reverse. This test pins usb_keep_order directly
// against the Go table - highest priority first, the reverse of dropOrder() -
// so drift in the boot-critical order is caught here even if that other
// script is ever deleted.
func TestShellAndGoAgreeOnTheKeepOrder(t *testing.T) {
	script := readInitScript(t)

	order := regexp.MustCompile(`usb_keep_order\(\) \{\s*echo "([^"]+)"`).FindStringSubmatch(script)
	if order == nil {
		t.Fatal("S03usbdev has no usb_keep_order function")
	}

	shell := strings.Fields(order[1])

	drop := dropOrder()
	goOrder := make([]string, len(drop))
	for i, function := range drop {
		goOrder[len(drop)-1-i] = function.name
	}

	if !reflect.DeepEqual(shell, goOrder) {
		t.Errorf("S03usbdev keeps %v, Go's dropOrder() reversed keeps %v", shell, goOrder)
	}
}

// usb_hid_cost has two arms and only one of them is hidCost: the disable_hid
// arm must charge nothing, and the other arm must charge exactly what Go
// charges. A substring search over the whole function body would stay green if
// the two arms were swapped - HID would then cost nothing while still being
// built, the guard would approve the console and the network together on top
// of the three inbound endpoints HID actually uses, and the gadget would ask
// for seven inbound of six with nothing said about it.
var hidCostShape = regexp.MustCompile(
	`(?s)if\s+usb_marker\s+disable_hid\s*\n\s*then\s*\n\s*echo\s+(\d+)\s*\n\s*else\s*\n\s*echo\s+(\d+)`)

func TestShellAndGoAgreeOnWhatHidCosts(t *testing.T) {
	script := readInitScript(t)

	for _, direction := range []struct {
		fn   string
		want int
	}{
		{"usb_hid_in_cost", hidInCost},
		{"usb_hid_out_cost", hidOutCost},
	} {
		start := strings.Index(script, direction.fn+"() {")
		if start < 0 {
			t.Fatalf("S03usbdev has no %s function", direction.fn)
		}

		end := strings.Index(script[start:], "\n}")
		if end < 0 {
			t.Fatalf("%s is not terminated", direction.fn)
		}

		body := script[start : start+end]

		arms := hidCostShape.FindStringSubmatch(body)
		if arms == nil {
			t.Fatalf("%s does not have the if usb_marker disable_hid / then / else shape this test parses:\n%s", direction.fn, body)
		}

		disabled, err := strconv.Atoi(arms[1])
		if err != nil {
			t.Fatalf("%s's disable_hid arm echoes %q, which is not a number", direction.fn, arms[1])
		}

		built, err := strconv.Atoi(arms[2])
		if err != nil {
			t.Fatalf("%s's else arm echoes %q, which is not a number", direction.fn, arms[2])
		}

		if disabled != 0 {
			t.Errorf("%s charges %d when disable_hid is present, want 0", direction.fn, disabled)
		}

		if built != direction.want {
			t.Errorf("%s charges %d when HID is built, Go charges %d", direction.fn, built, direction.want)
		}
	}
}

// shellGadgetDirs pulls the case arms out of usb_gadget_dirs:
//
//	network) echo "ncm.usb0 rndis.usb0" ;;
func shellGadgetDirs(t *testing.T, script string) map[string][]string {
	t.Helper()

	start := strings.Index(script, "usb_gadget_dirs() {")
	if start < 0 {
		t.Fatal("S03usbdev has no usb_gadget_dirs function")
	}

	end := strings.Index(script[start:], "\n}")
	if end < 0 {
		t.Fatal("usb_gadget_dirs is not terminated")
	}

	arm := regexp.MustCompile(`(?m)^\s*([a-z]+)\)\s*echo\s+"([^"]*)"`)
	dirs := make(map[string][]string)

	for _, match := range arm.FindAllStringSubmatch(script[start:start+end], -1) {
		dirs[match[1]] = strings.Fields(match[2])
	}

	return dirs
}

// usb_gadget_dirs is what the boot script's prune reads to decide which config
// symlinks to remove, and the Go table's gadget/gadgetAlt is what the UI reads
// to decide whether a function is actually running. They are two hand-kept
// copies of the same names. If the shell copy loses one, the prune stops
// removing it: a function the budget dropped stays linked from an earlier
// start, the total goes back over 9, and the gadget refuses to bind with every
// /dev/hidg* gone - the exact failure this feature exists to prevent.
func TestShellAndGoAgreeOnEveryGadgetDirectory(t *testing.T) {
	dirs := shellGadgetDirs(t, readInitScript(t))

	if len(dirs) != len(usbFunctions) {
		t.Errorf("usb_gadget_dirs names %d functions, the Go table has %d: %v",
			len(dirs), len(usbFunctions), dirs)
	}

	for _, function := range usbFunctions {
		want := []string{function.gadget}
		if function.gadgetAlt != "" {
			want = append(want, function.gadgetAlt)
		}

		shell, ok := dirs[function.name]
		if !ok {
			t.Errorf("usb_gadget_dirs has no arm for %q", function.name)
			continue
		}

		if !reflect.DeepEqual(shell, want) {
			t.Errorf("%s is %v in S03usbdev and %v in Go", function.name, shell, want)
		}
	}

	// HID is gated on /boot/disable_hid alone and never reaches the keep set,
	// so a hid.GS* arm here would hand the prune the keyboard and both mice.
	for name, shell := range dirs {
		for _, dir := range shell {
			if strings.HasPrefix(dir, "hid.") {
				t.Errorf("usb_gadget_dirs maps %q to %q - the prune would unlink a HID function", name, dir)
			}
		}
	}
}

// Both copies carry the same two constants, read from the controller rather
// than derived at runtime. This test is what notices if a future edit moves
// only one of them.
func TestShellAndGoAgreeOnTheBudget(t *testing.T) {
	script := readInitScript(t)

	for _, direction := range []struct {
		fn   string
		want int
	}{
		{"usb_in_budget", DefaultInEndpointBudget},
		{"usb_out_budget", DefaultOutEndpointBudget},
	} {
		budget := regexp.MustCompile(direction.fn + `\(\) \{\s*echo (\d+)`).FindStringSubmatch(script)
		if budget == nil {
			t.Fatalf("S03usbdev has no %s function, or it does not echo a constant", direction.fn)
		}

		shell, err := strconv.Atoi(budget[1])
		if err != nil {
			t.Fatalf("%s echoes %q, which is not a number", direction.fn, budget[1])
		}

		if shell != direction.want {
			t.Errorf("S03usbdev budgets %d endpoints in %s, Go budgets %d", shell, direction.fn, direction.want)
		}
	}
}

// The two directions are not the same number. A change that collapsed them
// back into one would allow the console and the network together again.
func TestTheTwoBudgetsAreNotOneNumber(t *testing.T) {
	if DefaultInEndpointBudget == DefaultOutEndpointBudget {
		t.Error("both directions budget the same count, so the direction is being ignored")
	}
}

// The budgets are read from the controller, and the kernel's own summary line
// disagrees with both. A later change that "fixes" a constant by parsing dmesg
// would take 8 for a number that means neither thing.
func TestTheShellBudgetIsNotReadFromDmesg(t *testing.T) {
	script := readInitScript(t)

	for _, fn := range []string{"usb_in_budget", "usb_out_budget"} {
		start := strings.Index(script, fn+"() {")
		if start < 0 {
			t.Fatalf("S03usbdev has no %s function", fn)
		}

		end := strings.Index(script[start:], "\n}")
		if end < 0 {
			t.Fatalf("%s is not terminated", fn)
		}

		if strings.Contains(script[start:start+end], "dmesg") {
			t.Errorf("%s consults dmesg, which reports a number that is neither budget", fn)
		}
	}
}
