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

// Neither the trial nor the boot script may flush IPv6.
//
// `ip addr flush dev eth0` takes every family with it, the IPv6 link-local
// included. On the device on 2026-09-07 a revert left eth0 with no IPv6 at all
// for several minutes. The global address came back later on its own and the
// link-local did not, and it took disabling and re-enabling IPv6 on the
// interface to get both. An address change should not cost an address family
// it does not manage, however long it takes to come back.
//
// Both copies are checked here, because the revert runs the boot script and
// the trial runs the Go code, so a fix to one of them alone leaves the other
// stripping the interface.
func TestNeitherTheTrialNorTheBootScriptFlushesIPv6(t *testing.T) {
	t.Run("the trial", func(t *testing.T) {
		for _, apply := range []struct {
			name string
			run  func() error
		}{
			{name: "static", run: func() error {
				return applyStatic(ethernetConfig{Address: "10.0.0.99", Prefix: 24})
			}},
			{name: "dhcp", run: applyDHCP},
		} {
			t.Run(apply.name, func(t *testing.T) {
				commands := recordCommands(t)
				if err := apply.run(); err != nil {
					t.Fatalf("failed to apply: %s", err)
				}

				assertFlushesIPv4Only(t, commands.all())
			})
		}
	})

	t.Run("the boot script", func(t *testing.T) {
		script, err := os.ReadFile(ethInitScriptSource)
		if err != nil {
			t.Skipf("cannot read %s: %s", ethInitScriptSource, err)
		}

		assertFlushesIPv4Only(t, strings.Split(string(script), "\n"))
	})
}

// assertFlushesIPv4Only requires at least one flush, so a copy that stopped
// flushing does not pass by having nothing to check, and requires every flush
// it finds to name the family.
func assertFlushesIPv4Only(t *testing.T, lines []string) {
	t.Helper()

	found := 0
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "ip" {
			continue
		}
		if !strings.Contains(line, "flush") {
			continue
		}

		found++
		if fields[1] != "-4" {
			t.Errorf("this flush takes the IPv6 addresses with it: %q", strings.TrimSpace(line))
		}
	}

	if found == 0 {
		t.Error("nothing here flushes the interface, so this test is measuring nothing")
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
