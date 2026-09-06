package network

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// ethInitScriptSource is the boot script this package has to agree with. The
// test reads the repository copy, because the device copy is installed from it.
const ethInitScriptSource = "../../../kvmapp/system/init.d/S30eth"

// A DHCP trial and a boot must ask the DHCP server for the same things.
// Otherwise a board that took its lease from a trial is configured differently
// from the same board after a reboot, and the difference appears later as a
// route that quietly went missing.
//
// The arguments drifted apart once already: the script gained `-O 121` and the
// Go copy of the argument list did not, while the comment above it claimed the
// two matched.
func TestTheTrialAsksForWhatTheBootScriptAsksFor(t *testing.T) {
	script, err := os.ReadFile(ethInitScriptSource)
	if err != nil {
		t.Skipf("cannot read %s: %s", ethInitScriptSource, err)
	}

	options := udhcpcOptionsIn(t, string(script))

	commands := recordCommands(t)
	if err := applyDHCP(); err != nil {
		t.Fatalf("failed to apply dhcp: %s", err)
	}

	var applied string
	for _, line := range commands.all() {
		if strings.HasPrefix(line, "udhcpc ") {
			applied = line
		}
	}
	if applied == "" {
		t.Fatalf("no udhcpc command ran: %v", commands.all())
	}

	for _, option := range options {
		if !strings.Contains(applied, option) {
			t.Errorf("the boot script passes %q to udhcpc and the trial does not: %q", option, applied)
		}
	}
}

// udhcpcOptionsIn collects the request options the script passes, which are the
// arguments that change what the lease carries. The timeouts and the pid file
// are deliberately not compared: the script uses a shorter retry count on its
// fallback path, and matching those exactly would make the test fail on a
// change that does not matter.
func udhcpcOptionsIn(t *testing.T, script string) []string {
	t.Helper()

	lines := regexp.MustCompile(`(?m)^.*\budhcpc\b.*$`).FindAllString(script, -1)
	if len(lines) == 0 {
		t.Fatalf("%s does not call udhcpc any more, so this test is measuring nothing", ethInitScriptSource)
	}

	seen := make(map[string]bool)
	var options []string

	for _, line := range lines {
		fields := strings.Fields(line)
		for i, field := range fields {
			if field != "-O" || i+1 >= len(fields) {
				continue
			}

			option := "-O " + fields[i+1]
			if !seen[option] {
				seen[option] = true
				options = append(options, option)
			}
		}
	}

	if len(options) == 0 {
		t.Skip("the boot script requests no dhcp options, so there is nothing to match")
	}

	return options
}
