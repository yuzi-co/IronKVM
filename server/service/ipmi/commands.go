package ipmi

import (
	"encoding/binary"
	"regexp"
	"strconv"
)

// The network functions the service answers.
const (
	netFnChassis = 0x00
	netFnApp     = 0x06
)

// The commands, IPMI 2.0 appendix G.
const (
	cmdGetDeviceID           = 0x01
	cmdGetSystemGUID         = 0x37
	cmdGetChannelAuthCaps    = 0x38
	cmdGetSessionChallenge   = 0x39
	cmdActivateSession       = 0x3a
	cmdSetSessionPrivilege   = 0x3b
	cmdCloseSession          = 0x3c
	cmdGetChannelCipherSuite = 0x54

	cmdGetChassisStatus = 0x01
	cmdChassisControl   = 0x02
	cmdChassisIdentify  = 0x04
)

// The completion codes, IPMI 2.0 table 5-2.
const (
	ccOK                  = 0x00
	ccNodeBusy            = 0xc0
	ccInvalidCommand      = 0xc1
	ccRequestDataLength   = 0xc7
	ccInvalidDataField    = 0xcc
	ccInsufficientPriv    = 0xd4
	ccNotInPresentState   = 0xd5
	ccUnspecified         = 0xff
	ccInvalidSessionID    = 0x87 // Close Session
	ccLevelNotAvailable   = 0x80 // Set Session Privilege Level
	ccLevelExceedsLimit   = 0x81 // Set Session Privilege Level
	ccUnsupportedAuthType = ccInvalidDataField
)

// lanChannel is the channel the service is. 0x0e in a request means "the
// channel this request came in on".
const (
	lanChannel     = 0x01
	currentChannel = 0x0e
)

// presessionCommands answer outside a session. Everything else needs one.
func presessionCommand(netFn, cmd byte) bool {
	if netFn != netFnApp {
		return false
	}
	switch cmd {
	case cmdGetChannelAuthCaps, cmdGetChannelCipherSuite, cmdGetSystemGUID,
		cmdGetSessionChallenge, cmdActivateSession:
		return true
	}
	return false
}

// commandPrivilege is the privilege a command in a session needs, or 0
// for a command the service does not have.
func commandPrivilege(netFn, cmd byte) byte {
	switch netFn {
	case netFnApp:
		switch cmd {
		case cmdGetChannelAuthCaps, cmdGetChannelCipherSuite, cmdGetSystemGUID,
			cmdSetSessionPrivilege, cmdCloseSession:
			return privCallback
		case cmdGetDeviceID:
			return privUser
		}
	case netFnChassis:
		switch cmd {
		case cmdGetChassisStatus:
			return privUser
		case cmdChassisControl, cmdChassisIdentify:
			return privOperator
		}
	}
	return 0
}

// dispatch runs one command. sess is nil outside a session.
func (s *Service) dispatch(sess *session, req *request) (byte, []byte) {
	if sess == nil {
		if !presessionCommand(req.netFn, req.cmd) {
			return ccInsufficientPriv, nil
		}
	} else {
		need := commandPrivilege(req.netFn, req.cmd)
		if need == 0 {
			return ccInvalidCommand, nil
		}
		if sess.priv < need {
			return ccInsufficientPriv, nil
		}
	}

	switch req.netFn {
	case netFnApp:
		switch req.cmd {
		case cmdGetChannelAuthCaps:
			return getChannelAuthCaps(req.data)
		case cmdGetChannelCipherSuite:
			return getChannelCipherSuites(req.data)
		case cmdGetSystemGUID:
			return ccOK, s.deps.GUID[:]
		case cmdGetSessionChallenge, cmdActivateSession:
			// IPMI 1.5 sessions are off: no authentication type is
			// offered, so any the request names is unsupported.
			return ccUnsupportedAuthType, nil
		case cmdGetDeviceID:
			return ccOK, s.deviceID()
		case cmdSetSessionPrivilege:
			return setSessionPrivilege(sess, req.data)
		case cmdCloseSession:
			return closeSession(sess, req.data)
		}
	case netFnChassis:
		switch req.cmd {
		case cmdGetChassisStatus:
			return s.chassisStatus()
		case cmdChassisControl:
			return s.chassisControl(sess, req.data)
		case cmdChassisIdentify:
			// There is no light to blink. The command is acknowledged so
			// tools that identify a machine before working on it go on.
			return ccOK, nil
		}
	}
	return ccInvalidCommand, nil
}

// getChannelAuthCaps is Get Channel Authentication Capabilities, IPMI 2.0
// section 22.13. It offers no IPMI 1.5 authentication type, only RMCP+,
// and says KG is all zeros, so the SIK is keyed by the user's password.
func getChannelAuthCaps(data []byte) (byte, []byte) {
	if len(data) < 2 {
		return ccRequestDataLength, nil
	}
	channel := data[0] & 0x0f
	if channel != currentChannel && channel != lanChannel {
		return ccInvalidDataField, nil
	}
	v20 := data[0]&0x80 != 0

	var authTypes, extended byte
	if v20 {
		// Bit 7: the extended capabilities byte is valid. Bit 1 of that
		// byte: IPMI 2.0 connections.
		authTypes = 0x80
		extended = 0x02
	}
	// Non-null user names only: no null user, no anonymous login. Per
	// message and user level authentication are both on, and KG is the
	// default.
	const userBits = 0x04
	return ccOK, []byte{lanChannel, authTypes, userBits, extended, 0, 0, 0, 0}
}

// getChannelCipherSuites is Get Channel Cipher Suites, IPMI 2.0 section
// 22.15. The records are cut into 16 byte pieces by list index; a piece
// shorter than 16 bytes is the last.
func getChannelCipherSuites(data []byte) (byte, []byte) {
	if len(data) < 3 {
		return ccRequestDataLength, nil
	}
	channel := data[0] & 0x0f
	if channel != currentChannel && channel != lanChannel {
		return ccInvalidDataField, nil
	}
	index := int(data[2] & 0x3f)

	var records []byte
	if data[1]&0x3f == payloadIPMI {
		if data[2]&0x80 != 0 {
			// Listed by cipher suite: a start-of-record byte, the suite ID,
			// then each algorithm with its tag in the top two bits.
			for _, suite := range cipherSuites {
				records = append(records, 0xc0, suite.id, suite.auth, 0x40|suite.integrity, 0x80|suite.confidentiality)
			}
		} else {
			records = []byte{authRAKPHMACSHA1, authRAKPHMACSHA256,
				0x40 | integrityHMACSHA196, 0x40 | integrityHMACSHA256_128,
				0x80 | confidentialityAESCBC128}
		}
	}

	out := []byte{lanChannel}
	if start := index * 16; start < len(records) {
		out = append(out, records[start:min(start+16, len(records))]...)
	}
	return ccOK, out
}

var versionPattern = regexp.MustCompile(`(\d+)\.(\d+)`)

// deviceID is Get Device ID's answer, IPMI 2.0 section 20.1: a device
// without SDRs, the application's major and minor version, IPMI 2.0, and
// only the chassis device among the optional ones. There is no IANA
// manufacturer number to give, so it is 0.
func (s *Service) deviceID() []byte {
	var major, minor byte
	if m := versionPattern.FindStringSubmatch(s.deps.FirmwareVersion()); m != nil {
		a, _ := strconv.Atoi(m[1])
		b, _ := strconv.Atoi(m[2])
		major = byte(min(a, 0x7f))
		b = min(b, 99)
		minor = byte(b/10<<4 | b%10)
	}
	const (
		deviceID       = 0x20
		deviceRevision = 0x01
		ipmiVersion    = 0x02 // 2.0, minor in the high nibble
		chassisDevice  = 0x80
	)
	return []byte{deviceID, deviceRevision, major, minor, ipmiVersion, chassisDevice, 0, 0, 0, 0, 0}
}

// setSessionPrivilege is Set Session Privilege Level, IPMI 2.0 section
// 22.18. 0 asks for the present level.
func setSessionPrivilege(sess *session, data []byte) (byte, []byte) {
	if len(data) < 1 {
		return ccRequestDataLength, nil
	}
	level := data[0] & 0x0f
	switch {
	case level == 0:
	case level < privUser || level > privAdmin:
		return ccLevelNotAvailable, nil
	case level > sess.limit:
		return ccLevelExceedsLimit, nil
	default:
		sess.priv = level
	}
	return ccOK, []byte{sess.priv}
}

// closeSession is Close Session, IPMI 2.0 section 22.19. A session may
// close only itself.
func closeSession(sess *session, data []byte) (byte, []byte) {
	if len(data) < 4 {
		return ccRequestDataLength, nil
	}
	if binary.LittleEndian.Uint32(data[:4]) != sess.id {
		return ccInvalidSessionID, nil
	}
	sess.closing = true
	return ccOK, nil
}
