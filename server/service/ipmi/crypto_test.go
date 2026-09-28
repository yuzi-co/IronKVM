package ipmi

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"
)

// vectorInput is the handshake the vectors below were computed for. They
// were computed outside this package, with Python's hmac and hashlib, from
// the formulas of IPMI 2.0 sections 13.31 and 13.32 as ipmitool implements
// them: the password zero-padded to 20 bytes, session IDs little-endian,
// the whole RAKP 1 role byte with its name-only lookup bit, and K1 and K2
// over 20 constant bytes for both suites.
func vectorInput() *rakpInput {
	in := &rakpInput{
		consoleID: 0xa4a3a2a1,
		systemID:  0x12345678,
		role:      0x14,
		username:  []byte("admin"),
	}
	for i := range 16 {
		in.rm[i] = byte(i)
		in.rc[i] = byte(16 + i)
		in.guid[i] = byte(32 + i)
	}
	return in
}

var vectorPassword = []byte("p4ssw0rd-ipmi!")

type rakpVectors struct {
	suite                            byte
	rakp2, rakp3, sik, k1, k2, rakp4 string
}

var vectors = []rakpVectors{
	{
		suite: 3,
		rakp2: "ce644fa9359f6c326266bcdce8757eefca6984f4",
		rakp3: "073f4bad7ec1a0046ac35b6cf138569165f3afbf",
		sik:   "78e580ecd2a6b74f924896aa436ffdfae9067762",
		k1:    "aa9cc4092c5403d421b76da98bb91dbf2714e7f2",
		k2:    "4aa9e9d68223856bf7d50a27339fea87660f07e2",
		rakp4: "4c043669d7f4f3e18f1c3abc",
	},
	{
		suite: 17,
		rakp2: "80035eee68177fbc967fcdf5448026009661e0c67d9dc64d6638cec5b1aae61d",
		rakp3: "32ce5b383970b7930402454342d6fce33f30fa4bbc9143234819705a87f41925",
		sik:   "65006b5beec0eabea2ffd19ec2824152fb011c462662a29e3c28283643b313cd",
		k1:    "0ac59c944b6ccbbea877290af9023e4d3a09ad50469cd06fe57fd821383f7f7c",
		k2:    "3442d1109ce4b8923190e9590e8ec73ca5b9a577796fb5b94fda1bb97befd6e5",
		rakp4: "77d0526e1f0e7f1d9e139d88809a6fb5",
	},
}

func suiteByID(t *testing.T, id byte) *cipherSuite {
	t.Helper()
	for _, suite := range cipherSuites {
		if suite.id == id {
			return suite
		}
	}
	t.Fatalf("no cipher suite %d", id)
	return nil
}

func TestTheRAKPMathMatchesIndependentVectors(t *testing.T) {
	for _, v := range vectors {
		suite := suiteByID(t, v.suite)
		in := vectorInput()

		sik := suite.sik(vectorPassword, in)
		got := map[string][]byte{
			"rakp2": suite.rakp2AuthCode(vectorPassword, in),
			"rakp3": suite.rakp3AuthCode(vectorPassword, in),
			"sik":   sik,
			"k1":    suite.k1(sik),
			"k2":    suite.k2(sik),
			"rakp4": suite.rakp4ICV(sik, in),
		}
		want := map[string]string{
			"rakp2": v.rakp2, "rakp3": v.rakp3, "sik": v.sik,
			"k1": v.k1, "k2": v.k2, "rakp4": v.rakp4,
		}
		for name, value := range got {
			if hex.EncodeToString(value) != want[name] {
				t.Errorf("suite %d %s = %x, want %s", v.suite, name, value, want[name])
			}
		}
	}
}

// ipmitool pads the password with zeros to 20 bytes before it keys an
// HMAC. HMAC pads its key with zeros anyway, so the service, which keys
// with the bare password, must get the same answer.
func TestAZeroPaddedPasswordGivesTheSameHMAC(t *testing.T) {
	padded := make([]byte, 20)
	copy(padded, vectorPassword)
	for _, suite := range cipherSuites {
		in := vectorInput()
		if !bytes.Equal(suite.rakp2AuthCode(padded, in), suite.rakp2AuthCode(vectorPassword, in)) {
			t.Errorf("suite %d: the padded password gives another HMAC", suite.id)
		}
	}
}

func TestOnlySuites3And17AreOffered(t *testing.T) {
	if s := findCipherSuite(authRAKPHMACSHA1, integrityHMACSHA196, confidentialityAESCBC128); s == nil || s.id != 3 {
		t.Errorf("suite 3 not found: %+v", s)
	}
	if s := findCipherSuite(authRAKPHMACSHA256, integrityHMACSHA256_128, confidentialityAESCBC128); s == nil || s.id != 17 {
		t.Errorf("suite 17 not found: %+v", s)
	}
	for _, combo := range [][3]byte{
		{0, 0, 0}, // cipher suite 0
		{authRAKPHMACSHA1, 0, 0},
		{authRAKPHMACSHA1, integrityHMACSHA196, 0}, // integrity without encryption
		{authRAKPHMACSHA256, integrityHMACSHA256_128, 0},
		{0x02, 0x02, 0x01},                            // HMAC-MD5
		{authRAKPHMACSHA1, integrityHMACSHA196, 0x02}, // xRC4
	} {
		if s := findCipherSuite(combo[0], combo[1], combo[2]); s != nil {
			t.Errorf("algorithms %v were accepted as suite %d", combo, s.id)
		}
	}
}

func TestEncryptPayloadLaysOutIVDataPadAndPadLength(t *testing.T) {
	k2 := bytes.Repeat([]byte{0x5a}, 20)
	for size := 0; size <= 40; size++ {
		plain := bytes.Repeat([]byte{0xa5}, size)
		out, err := encryptPayload(k2, plain)
		if err != nil {
			t.Fatal(err)
		}
		if (len(out)-aes.BlockSize)%aes.BlockSize != 0 || len(out) < 2*aes.BlockSize {
			t.Fatalf("size %d: %d bytes out", size, len(out))
		}

		// Decrypt with the standard library alone and look at the layout.
		block, _ := aes.NewCipher(k2[:16])
		raw := make([]byte, len(out)-aes.BlockSize)
		cipher.NewCBCDecrypter(block, out[:aes.BlockSize]).CryptBlocks(raw, out[aes.BlockSize:])
		padLen := int(raw[len(raw)-1])
		if size+padLen+1 != len(raw) || padLen >= aes.BlockSize {
			t.Fatalf("size %d: pad length %d in %d bytes", size, padLen, len(raw))
		}
		if !bytes.Equal(raw[:size], plain) {
			t.Fatalf("size %d: data changed", size)
		}
		for i := 0; i < padLen; i++ {
			if raw[size+i] != byte(i+1) {
				t.Fatalf("size %d: pad byte %d is %d", size, i, raw[size+i])
			}
		}

		back, err := decryptPayload(k2, out)
		if err != nil || !bytes.Equal(back, plain) {
			t.Fatalf("size %d: round trip gave %x, %v", size, back, err)
		}
	}
}

func TestDecryptPayloadRefusesABadPad(t *testing.T) {
	k2 := bytes.Repeat([]byte{0x33}, 20)
	block, _ := aes.NewCipher(k2[:16])
	iv := make([]byte, aes.BlockSize)

	for _, plain := range [][]byte{
		append(bytes.Repeat([]byte{0}, 15), 16),                 // pad longer than a block
		append(bytes.Repeat([]byte{0}, 13), 1, 3, 2),            // pad bytes out of order
		append(append(bytes.Repeat([]byte{7}, 12), 1, 2, 9), 3), // wrong pad byte
	} {
		ct := make([]byte, len(plain))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(ct, plain)
		if _, err := decryptPayload(k2, append(append([]byte(nil), iv...), ct...)); err == nil {
			t.Errorf("pad %x was accepted", plain[len(plain)-4:])
		}
	}
	if _, err := decryptPayload(k2, make([]byte, 20)); err == nil {
		t.Error("a payload that is not whole blocks was accepted")
	}
}

func TestTheIntegrityTrailerRoundTrips(t *testing.T) {
	for _, suite := range cipherSuites {
		k1 := bytes.Repeat([]byte{0x11}, 20)
		for size := 0; size < 9; size++ {
			packet := buildV20(v20Header{payloadType: 0xc0, sessionID: 7, seq: 9}, bytes.Repeat([]byte{1}, size), suite, k1)
			if (len(packet)-rmcpHeaderLen-suite.authCodeLen)%4 != 0 {
				t.Fatalf("suite %d size %d: the AuthCode input is %d bytes", suite.id, size, len(packet)-rmcpHeaderLen-suite.authCodeLen)
			}
			p, err := parseV20(packet)
			if err != nil {
				t.Fatal(err)
			}
			if !checkTrailer(p, suite, k1) {
				t.Fatalf("suite %d size %d: own trailer refused", suite.id, size)
			}
			packet[v20PayloadAt] ^= 1
			if size > 0 && checkTrailer(p, suite, k1) {
				t.Fatalf("suite %d size %d: a changed payload passed", suite.id, size)
			}
		}
	}
}
