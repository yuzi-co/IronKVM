package ipmi

import (
	"errors"
	"reflect"
	"testing"
)

func TestPlanControl(t *testing.T) {
	const (
		power5s  = "power 5s"
		power800 = "power 800ms"
		reset800 = "reset 800ms"
	)
	cases := []struct {
		name   string
		action byte
		led    *bool
		want   []string
		cc     byte
	}{
		{"power down, on", controlPowerDown, on(), []string{power5s}, ccOK},
		{"power down, off", controlPowerDown, off(), nil, ccOK},
		{"power down, unknown", controlPowerDown, nil, nil, ccNotInPresentState},
		{"power up, off", controlPowerUp, off(), []string{power800}, ccOK},
		{"power up, on", controlPowerUp, on(), nil, ccOK},
		{"power up, unknown", controlPowerUp, nil, nil, ccNotInPresentState},
		{"cycle, on", controlPowerCycle, on(), []string{power5s, power800}, ccOK},
		{"cycle, off", controlPowerCycle, off(), nil, ccNotInPresentState},
		{"cycle, unknown", controlPowerCycle, nil, nil, ccNotInPresentState},
		{"hard reset, on", controlHardReset, on(), []string{reset800}, ccOK},
		{"hard reset, off", controlHardReset, off(), []string{reset800}, ccOK},
		{"hard reset, unknown", controlHardReset, nil, []string{reset800}, ccOK},
		{"diagnostic interrupt", controlDiagInt, on(), nil, ccInvalidDataField},
		{"soft off, on", controlSoftOff, on(), []string{power800}, ccOK},
		{"soft off, off", controlSoftOff, off(), nil, ccOK},
		{"soft off, unknown", controlSoftOff, nil, nil, ccNotInPresentState},
		{"reserved", 0x06, on(), nil, ccInvalidDataField},
	}
	for _, tc := range cases {
		presses, cc := planControl(tc.action, tc.led)
		var got []string
		for _, p := range presses {
			got = append(got, p.button+" "+p.hold.String())
		}
		if cc != tc.cc || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: %v cc %#x, want %v cc %#x", tc.name, got, cc, tc.want, tc.cc)
		}
	}
}

// adminSession logs in as the admin and takes ADMINISTRATOR.
func adminSession(t *testing.T, env *testEnv) *testClient {
	t.Helper()
	c := dial(t, env.service)
	c.login(17, adminName, adminIPMI, privAdmin|lookupNameOnly)
	c.mustCommand(netFnApp, cmdSetSessionPrivilege, []byte{privAdmin}, ccOK)
	return c
}

func TestChassisControlPressesThroughPressButton(t *testing.T) {
	env := newTestEnv(t)
	c := adminSession(t, env)

	env.host.setLED(off())
	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlPowerUp}, ccOK)
	if got := env.host.waitForPresses(t, env.service, 1); !reflect.DeepEqual(got, []string{"power 800ms"}) {
		t.Fatalf("power up pressed %v", got)
	}

	env.host.setLED(on())
	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlPowerCycle}, ccOK)
	want := []string{"power 800ms", "power 5s", "power 800ms"}
	if got := env.host.waitForPresses(t, env.service, 3); !reflect.DeepEqual(got, want) {
		t.Fatalf("power cycle pressed %v", got)
	}

	c.mustCommand(netFnChassis, cmdChassisIdentify, []byte{15}, ccOK)
	if got := env.host.pressed(); len(got) != 3 {
		t.Fatalf("identify pressed something: %v", got)
	}
}

func TestWithoutThePowerLEDOnlyResetIsOffered(t *testing.T) {
	env := newTestEnv(t)
	env.host.setLED(nil)
	c := adminSession(t, env)

	c.mustCommand(netFnChassis, cmdGetChassisStatus, nil, ccNotInPresentState)
	for _, action := range []byte{controlPowerDown, controlPowerUp, controlPowerCycle, controlSoftOff} {
		c.mustCommand(netFnChassis, cmdChassisControl, []byte{action}, ccNotInPresentState)
	}
	if got := env.host.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v without knowing the power state", got)
	}

	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlHardReset}, ccOK)
	if got := env.host.waitForPresses(t, env.service, 1); !reflect.DeepEqual(got, []string{"reset 800ms"}) {
		t.Fatalf("hard reset pressed %v", got)
	}
}

func TestAnUnreadableLEDFailsStatusAndPowerButNotReset(t *testing.T) {
	env := newTestEnv(t)
	env.host.mu.Lock()
	env.host.ledErr = errors.New("gpio gone")
	env.host.mu.Unlock()
	c := adminSession(t, env)

	c.mustCommand(netFnChassis, cmdGetChassisStatus, nil, ccUnspecified)
	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlPowerUp}, ccUnspecified)
	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlHardReset}, ccOK)
	if got := env.host.waitForPresses(t, env.service, 1); !reflect.DeepEqual(got, []string{"reset 800ms"}) {
		t.Fatalf("pressed %v", got)
	}
}

func TestAControlWhilePressesRunIsBusy(t *testing.T) {
	env := newTestEnv(t)
	c := adminSession(t, env)

	env.service.powerMu.Lock()
	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlHardReset}, ccNodeBusy)
	env.service.powerMu.Unlock()
	if got := env.host.pressed(); len(got) != 0 {
		t.Fatalf("pressed %v while busy", got)
	}
}

func TestChassisStatusFollowsTheLED(t *testing.T) {
	env := newTestEnv(t)
	c := adminSession(t, env)

	env.host.setLED(off())
	if got := c.mustCommand(netFnChassis, cmdGetChassisStatus, nil, ccOK); got[0]&0x01 != 0 || got[0]&0x60 != 0x60 {
		t.Fatalf("off: %x", got)
	}
	env.host.setLED(on())
	if got := c.mustCommand(netFnChassis, cmdGetChassisStatus, nil, ccOK); got[0]&0x01 != 1 {
		t.Fatalf("on: %x", got)
	}
}
