package ipmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"NanoKVM-Server/authn"
)

func TestChannelAuthCapabilitiesOfferRMCPPlusOnly(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)

	cc, data, err := c.v15Command(netFnApp, cmdGetChannelAuthCaps, []byte{0x8e, privAdmin})
	if err != nil || cc != ccOK {
		t.Fatalf("cc %#x, %v", cc, err)
	}
	want := []byte{lanChannel, 0x80, 0x04, 0x02, 0, 0, 0, 0}
	if !bytes.Equal(data, want) {
		t.Fatalf("answer %x, want %x", data, want)
	}

	// A client that asks without the IPMI 2.0 bit is an IPMI 1.5 client,
	// and is offered nothing.
	_, data, _ = c.v15Command(netFnApp, cmdGetChannelAuthCaps, []byte{0x0e, privAdmin})
	if data[1] != 0 || data[3] != 0 {
		t.Fatalf("a 1.5 client was offered %x", data)
	}
}

func TestIPMI15SessionsAreRefused(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)

	// Get Session Challenge for MD5 and for no authentication.
	for _, authType := range []byte{0x00, 0x02} {
		data := append([]byte{authType}, []byte(adminName)...)
		data = append(data, make([]byte, 16-len(adminName))...)
		cc, _, err := c.v15Command(netFnApp, cmdGetSessionChallenge, data)
		if err != nil || cc != ccInvalidDataField {
			t.Fatalf("auth type %d: cc %#x, %v", authType, cc, err)
		}
	}

	// A packet with an IPMI 1.5 AuthCode gets no answer at all.
	msg := buildRequest(netFnApp, cmdGetDeviceID, 1, nil)
	packet := []byte{rmcpVersion, 0, rmcpNoAck, rmcpClassIPMI, 0x02, 1, 0, 0, 0, 1, 0, 0, 0}
	packet = append(packet, make([]byte, 16)...)
	packet = append(packet, byte(len(msg)))
	if _, err := c.roundTrip(append(packet, msg...)); !errors.Is(err, errNoAnswer) {
		t.Fatalf("an MD5 session packet was answered: %v", err)
	}
}

func TestCipherSuitesAreListed(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)

	cc, data, err := c.presession(netFnApp, cmdGetChannelCipherSuite, []byte{0x0e, 0x00, 0x80})
	if err != nil || cc != ccOK {
		t.Fatalf("cc %#x, %v", cc, err)
	}
	want := []byte{lanChannel, 0xc0, 3, 0x01, 0x41, 0x81, 0xc0, 17, 0x03, 0x44, 0x81}
	if !bytes.Equal(data, want) {
		t.Fatalf("records %x, want %x", data, want)
	}

	// The next index is past the end.
	_, data, _ = c.presession(netFnApp, cmdGetChannelCipherSuite, []byte{0x0e, 0x00, 0x81})
	if !bytes.Equal(data, []byte{lanChannel}) {
		t.Fatalf("index 1 gave %x", data)
	}
}

func TestNothingElseAnswersOutsideASession(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)

	for _, cmd := range [][2]byte{{netFnApp, cmdGetDeviceID}, {netFnChassis, cmdGetChassisStatus}, {netFnChassis, cmdChassisControl}} {
		cc, _, err := c.presession(cmd[0], cmd[1], []byte{controlPowerUp})
		if err != nil || cc != ccInsufficientPriv {
			t.Fatalf("netFn %#x cmd %#x: cc %#x, %v", cmd[0], cmd[1], cc, err)
		}
	}
	if len(env.host.pressed()) != 0 {
		t.Fatalf("pressed %v", env.host.pressed())
	}
}

func TestCipherSuite0AndOtherCombinationsAreRefused(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)

	for _, algs := range [][3]byte{
		{0, 0, 0},
		{authRAKPHMACSHA1, integrityHMACSHA196, 0},
		{authRAKPHMACSHA256, integrityHMACSHA256_128, 0},
		{0x02, 0x02, 0x01},
	} {
		status, err := c.openSession(algs[0], algs[1], algs[2], 0)
		if err != nil || status != statusNoCipherSuiteMatch {
			t.Fatalf("algorithms %v: status %#x, %v", algs, status, err)
		}
	}
	if n := len(env.service.sessions); n != 0 {
		t.Fatalf("%d sessions left behind", n)
	}
}

func TestALoginWorksWithBothSuites(t *testing.T) {
	for _, suite := range []byte{3, 17} {
		env := newTestEnv(t)
		c := dial(t, env.service)
		c.login(suite, adminName, adminIPMI, privAdmin|lookupNameOnly)

		id := c.mustCommand(netFnApp, cmdGetDeviceID, nil, ccOK)
		if len(id) < 11 || id[2] != 2 || id[3] != 0x04 || id[4] != 0x02 || id[5] != 0x80 {
			t.Fatalf("suite %d: device ID %x", suite, id)
		}
		if got := c.mustCommand(netFnApp, cmdGetSystemGUID, nil, ccOK); string(got) != testGUIDValue {
			t.Fatalf("GUID %q", got)
		}

		// The session starts at USER.
		c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlPowerUp}, ccInsufficientPriv)
		if got := c.mustCommand(netFnApp, cmdSetSessionPrivilege, []byte{privAdmin}, ccOK); !bytes.Equal(got, []byte{privAdmin}) {
			t.Fatalf("set privilege gave %x", got)
		}

		status := c.mustCommand(netFnChassis, cmdGetChassisStatus, nil, ccOK)
		if len(status) != 3 || status[0]&0x01 != 1 {
			t.Fatalf("chassis status %x", status)
		}

		c.mustCommand(netFnApp, cmdCloseSession, binary.LittleEndian.AppendUint32(nil, c.systemID), ccOK)
		if _, _, err := c.command(netFnApp, cmdGetDeviceID, nil); !errors.Is(err, errNoAnswer) {
			t.Fatalf("suite %d: the closed session still answers: %v", suite, err)
		}
		if len(env.limiter.succeeded) != 1 {
			t.Fatalf("limiter successes %v", env.limiter.succeeded)
		}
	}
}

func TestAWrongPasswordFailsAtRAKP3(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)
	c.ignoreRAKP2 = true

	suite := suiteByID(t, 3)
	if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err != nil || status != statusOK {
		t.Fatal(status, err)
	}
	status, err := c.rakp(adminName, "not-the-password", privAdmin|lookupNameOnly)
	if err != nil || status != statusInvalidIntegrityChk {
		t.Fatalf("status %#x, %v", status, err)
	}
	if len(env.limiter.failed) != 1 || env.limiter.failed[0] != "127.0.0.1/"+adminName {
		t.Fatalf("limiter failures %v", env.limiter.failed)
	}
	if n := len(env.service.sessions); n != 0 {
		t.Fatalf("%d sessions left after a failed login", n)
	}
}

// The web password is not the IPMI password.
func TestTheWebPasswordDoesNotLogIn(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)

	suite := suiteByID(t, 17)
	_, _ = c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0)
	if _, err := c.rakp(adminName, adminWeb, privAdmin|lookupNameOnly); err == nil {
		t.Fatal("the web password logged in")
	}
}

func TestAccountsThatCannotUseIPMIAreRefusedByName(t *testing.T) {
	env := newTestEnv(t)
	suite := suiteByID(t, 3)

	for _, name := range []string{"nobody", noIPMIName, disabledName, ""} {
		c := dial(t, env.service)
		_, _ = c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0)
		status, err := c.rakp(name, "whatever-password", privUser|lookupNameOnly)
		if err != nil || status != statusUnauthorizedName {
			t.Fatalf("%q: status %#x, %v", name, status, err)
		}
	}

	c := dial(t, env.service)
	_, _ = c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0)
	status, err := c.rakp("abcdefghijklmnopq", "whatever-password", privUser)
	if err != nil || status != statusInvalidNameLength {
		t.Fatalf("a 17 character name: status %#x, %v", status, err)
	}
}

func TestALockedOutAccountIsRefused(t *testing.T) {
	env := newTestEnv(t)
	env.limiter.locked = true
	c := dial(t, env.service)

	suite := suiteByID(t, 3)
	_, _ = c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0)
	status, err := c.rakp(adminName, adminIPMI, privAdmin|lookupNameOnly)
	if err != nil || status != statusUnauthorizedName {
		t.Fatalf("status %#x, %v", status, err)
	}
}

func TestAUserAccountGetsUSEROnly(t *testing.T) {
	env := newTestEnv(t)
	suite := suiteByID(t, 17)

	// ipmitool's default is ADMINISTRATOR, which a user account does not
	// get.
	c := dial(t, env.service)
	_, _ = c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0)
	if status, err := c.rakp(userName, userIPMI, privAdmin|lookupNameOnly); err != nil || status != statusUnauthorizedRole {
		t.Fatalf("status %#x, %v", status, err)
	}

	c = dial(t, env.service)
	c.login(17, userName, userIPMI, privUser|lookupNameOnly)
	c.mustCommand(netFnChassis, cmdGetChassisStatus, nil, ccOK)
	c.mustCommand(netFnApp, cmdSetSessionPrivilege, []byte{privOperator}, ccLevelExceedsLimit)
	c.mustCommand(netFnChassis, cmdChassisControl, []byte{controlHardReset}, ccInsufficientPriv)
	if len(env.host.pressed()) != 0 {
		t.Fatalf("a user account pressed %v", env.host.pressed())
	}
}

func TestReplayedAndTamperedPacketsAreDropped(t *testing.T) {
	env := newTestEnv(t)
	c := dial(t, env.service)
	c.login(3, adminName, adminIPMI, privAdmin|lookupNameOnly)

	packet := c.sessionPacket(10, 1, netFnApp, cmdGetDeviceID, nil)
	if _, err := c.roundTrip(packet); err != nil {
		t.Fatal(err)
	}
	if _, err := c.roundTrip(packet); !errors.Is(err, errNoAnswer) {
		t.Fatalf("a replayed packet was answered: %v", err)
	}

	// Far behind the highest sequence number.
	if _, err := c.roundTrip(c.sessionPacket(100, 2, netFnApp, cmdGetDeviceID, nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := c.roundTrip(c.sessionPacket(50, 3, netFnApp, cmdGetDeviceID, nil)); !errors.Is(err, errNoAnswer) {
		t.Fatalf("a packet 50 behind was answered: %v", err)
	}
	// A little behind, and not seen before, is fine: UDP reorders.
	if _, err := c.roundTrip(c.sessionPacket(90, 4, netFnApp, cmdGetDeviceID, nil)); err != nil {
		t.Fatalf("a packet 10 behind was dropped: %v", err)
	}

	tampered := c.sessionPacket(101, 5, netFnApp, cmdGetDeviceID, nil)
	tampered[v20PayloadAt+3] ^= 0x01
	if _, err := c.roundTrip(tampered); !errors.Is(err, errNoAnswer) {
		t.Fatalf("a tampered packet was answered: %v", err)
	}

	// The same request without encryption and authentication.
	plain := buildV20(v20Header{payloadType: payloadIPMI, sessionID: c.systemID, seq: 102},
		buildRequest(netFnApp, cmdGetDeviceID, 6, nil), nil, nil)
	if _, err := c.roundTrip(plain); !errors.Is(err, errNoAnswer) {
		t.Fatalf("a plain packet in the session was answered: %v", err)
	}
}

// ipmitool sends a request whose answer was lost again, with a new session
// sequence number and the same request sequence number. It must not press
// twice.
func TestARequestSentAgainIsNotRunTwice(t *testing.T) {
	env := newTestEnv(t)
	env.host.setLED(off())
	c := dial(t, env.service)
	c.login(3, adminName, adminIPMI, privAdmin|lookupNameOnly)
	c.mustCommand(netFnApp, cmdSetSessionPrivilege, []byte{privAdmin}, ccOK)

	for seq := uint32(20); seq < 23; seq++ {
		reply, err := c.roundTrip(c.sessionPacket(seq, 9, netFnChassis, cmdChassisControl, []byte{controlPowerUp}))
		if err != nil {
			t.Fatal(err)
		}
		if cc, _, err := c.readSessionAnswer(reply); err != nil || cc != ccOK {
			t.Fatalf("cc %#x, %v", cc, err)
		}
	}
	if got := env.host.waitForPresses(t, env.service, 1); len(got) != 1 {
		t.Fatalf("presses %v", got)
	}
}

func TestSessionsExpire(t *testing.T) {
	env := newTestEnv(t)
	var now atomic.Int64
	now.Store(time.Now().UnixNano())
	env.service.mu.Lock()
	env.service.deps.Now = func() time.Time { return time.Unix(0, now.Load()) }
	env.service.mu.Unlock()

	c := dial(t, env.service)
	c.login(3, adminName, adminIPMI, privAdmin|lookupNameOnly)
	c.mustCommand(netFnApp, cmdGetDeviceID, nil, ccOK)

	now.Add(int64(sessionTimeout))
	if _, _, err := c.command(netFnApp, cmdGetDeviceID, nil); !errors.Is(err, errNoAnswer) {
		t.Fatalf("an expired session answered: %v", err)
	}
}

func TestTheSessionCountIsBounded(t *testing.T) {
	env := newTestEnv(t)
	suite := suiteByID(t, 3)
	c := dial(t, env.service)

	for i := 0; i < maxSessions; i++ {
		if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err != nil || status != statusOK {
			t.Fatalf("session %d: status %#x, %v", i, status, err)
		}
	}
	if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err != nil || status != statusNoResources {
		t.Fatalf("one too many: status %#x, %v", status, err)
	}
}

func TestAChangedOrRemovedAccountEndsItsSession(t *testing.T) {
	env := newTestEnv(t)

	c := dial(t, env.service)
	c.login(3, adminName, adminIPMI, privAdmin|lookupNameOnly)
	env.setIPMIPassword(t, adminName, "another-ipmi-pw")
	if _, _, err := c.command(netFnApp, cmdGetDeviceID, nil); !errors.Is(err, errNoAnswer) {
		t.Fatalf("a session outlived its password: %v", err)
	}

	c = dial(t, env.service)
	c.login(3, userName, userIPMI, privUser|lookupNameOnly)
	if _, err := env.accounts.Update(adminName, userName, authn.UserPatch{Enabled: off()}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.command(netFnApp, cmdGetDeviceID, nil); !errors.Is(err, errNoAnswer) {
		t.Fatalf("a session outlived its account: %v", err)
	}
}

func TestTurningTheServiceOffClosesTheSocket(t *testing.T) {
	env := newTestEnv(t)
	env.service.Stop()
	if env.service.LocalAddr() != nil {
		t.Fatal("the socket is still open")
	}
	if n := len(env.service.sessions); n != 0 {
		t.Fatalf("%d sessions after stop", n)
	}
}
