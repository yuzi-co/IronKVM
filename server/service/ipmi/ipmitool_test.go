package ipmi

import (
	"context"
	"net"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

// These tests run the real ipmitool against the service, started in the
// test process with a fake host, so no button is ever pressed. They skip
// where ipmitool is not installed. To run them, use a container that has
// it, for example:
//
//	docker run --rm -v <repo>:/src -w /src/server -e CGO_ENABLED=0 golang:1.25 \
//	  sh -c 'apt-get update && apt-get install -y ipmitool && \
//	         go test -tags novision -count=1 -run Ipmitool -v ./service/ipmi/'

func ipmitool(t *testing.T, env *testEnv, args ...string) (string, error) {
	t.Helper()
	_, port, _ := net.SplitHostPort(env.service.LocalAddr().String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	full := append([]string{"-I", "lanplus", "-H", "127.0.0.1", "-p", port, "-N", "1", "-R", "2"}, args...)
	out, err := exec.CommandContext(ctx, "ipmitool", full...).CombinedOutput()
	t.Logf("ipmitool %s\n%s", strings.Join(full, " "), out)
	return string(out), err
}

func needIpmitool(t *testing.T) {
	if _, err := exec.LookPath("ipmitool"); err != nil {
		t.Skip("ipmitool is not installed")
	}
}

func TestIpmitoolReadsStatusAndPowersOn(t *testing.T) {
	needIpmitool(t)

	// "auto" leaves the choice to ipmitool, which asks for the cipher
	// suites and picks 17.
	for _, suite := range []string{"3", "17", "auto"} {
		t.Run("suite "+suite, func(t *testing.T) {
			env := newTestEnv(t)
			env.host.setLED(off())
			login := []string{"-U", adminName, "-P", adminIPMI}
			if suite != "auto" {
				login = append(login, "-C", suite)
			}
			run := func(args ...string) string {
				t.Helper()
				out, err := ipmitool(t, env, append(append([]string(nil), login...), args...)...)
				if err != nil {
					t.Fatalf("ipmitool %v: %v", args, err)
				}
				return out
			}

			if out := run("chassis", "status"); !strings.Contains(out, "System Power         : off") {
				t.Fatalf("chassis status:\n%s", out)
			}
			if out := run("mc", "info"); !strings.Contains(out, "IPMI Version              : 2.0") ||
				!strings.Contains(out, "Firmware Revision         : 2.04") {
				t.Fatalf("mc info:\n%s", out)
			}
			if out := run("power", "status"); !strings.Contains(out, "Chassis Power is off") {
				t.Fatalf("power status:\n%s", out)
			}
			run("power", "on")
			if got := env.host.waitForPresses(t, env.service, 1); !reflect.DeepEqual(got, []string{"power 800ms"}) {
				t.Fatalf("power on pressed %v", got)
			}
			env.host.setLED(on())
			if out := run("power", "status"); !strings.Contains(out, "Chassis Power is on") {
				t.Fatalf("power status:\n%s", out)
			}
			run("chassis", "identify")
			if got := env.host.pressed(); len(got) != 1 {
				t.Fatalf("identify pressed %v", got)
			}

			// The rest of the actions, with the host on.
			want := []string{"power 800ms"}
			for _, step := range []struct {
				action  string
				presses []string
			}{
				{"soft", []string{"power 800ms"}},
				{"off", []string{"power 5s"}},
				{"cycle", []string{"power 5s", "power 800ms"}},
				{"reset", []string{"reset 800ms"}},
			} {
				run("power", step.action)
				want = append(want, step.presses...)
				if got := env.host.waitForPresses(t, env.service, len(want)); !reflect.DeepEqual(got, want) {
					t.Fatalf("power %s: pressed %v, want %v", step.action, got, want)
				}
			}
		})
	}
}

func TestIpmitoolRefusals(t *testing.T) {
	needIpmitool(t)
	env := newTestEnv(t)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"a wrong password", []string{"-U", adminName, "-P", "wrong-password", "power", "status"}},
		{"the web password", []string{"-U", adminName, "-P", adminWeb, "power", "status"}},
		{"cipher suite 0", []string{"-U", adminName, "-P", adminIPMI, "-C", "0", "power", "status"}},
		{"a user account asking for ADMINISTRATOR", []string{"-U", userName, "-P", userIPMI, "power", "status"}},
		{"a user account pressing", []string{"-U", userName, "-P", userIPMI, "-L", "USER", "power", "reset"}},
	} {
		if out, err := ipmitool(t, env, tc.args...); err == nil {
			t.Errorf("%s worked:\n%s", tc.name, out)
		}
	}

	// IPMI 1.5, which is ipmitool's lan interface, is refused.
	_, port, _ := net.SplitHostPort(env.service.LocalAddr().String())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ipmitool", "-I", "lan", "-H", "127.0.0.1", "-p", port, "-N", "1", "-R", "1",
		"-U", adminName, "-P", adminIPMI, "power", "status").CombinedOutput()
	if err == nil {
		t.Errorf("IPMI 1.5 worked:\n%s", out)
	}

	if out, err := ipmitool(t, env, "-U", userName, "-P", userIPMI, "-L", "USER", "power", "status"); err != nil {
		t.Errorf("a user account cannot read the power state: %v\n%s", err, out)
	}
	if got := env.host.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v", got)
	}
}
