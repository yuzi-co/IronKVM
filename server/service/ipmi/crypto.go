package ipmi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
)

// The algorithm numbers of the Open Session payloads, IPMI 2.0 table 13-17
// to 13-19. Only the ones the two cipher suites use are named.
const (
	authRAKPHMACSHA1   = 0x01
	authRAKPHMACSHA256 = 0x03

	integrityHMACSHA196     = 0x01
	integrityHMACSHA256_128 = 0x04

	confidentialityAESCBC128 = 0x01
)

// cipherSuite is one of the two combinations the service accepts. The
// specification lets a client pick each algorithm on its own, but a BMC only
// has to honour the combinations it lists, and these are the two that have
// no MD5, no RC4 and no missing piece.
type cipherSuite struct {
	id              byte
	auth            byte
	integrity       byte
	confidentiality byte

	// hash is the RAKP HMAC's hash. The integrity HMAC uses the same one in
	// both suites.
	hash func() hash.Hash
	// icvLen is how much of the RAKP 4 HMAC goes on the wire.
	icvLen int
	// authCodeLen is how much of the integrity HMAC goes on the wire.
	authCodeLen int
}

var cipherSuites = []*cipherSuite{
	{
		id: 3, auth: authRAKPHMACSHA1, integrity: integrityHMACSHA196, confidentiality: confidentialityAESCBC128,
		hash: sha1.New, icvLen: 12, authCodeLen: 12,
	},
	{
		id: 17, auth: authRAKPHMACSHA256, integrity: integrityHMACSHA256_128, confidentiality: confidentialityAESCBC128,
		hash: sha256.New, icvLen: 16, authCodeLen: 16,
	},
}

// findCipherSuite returns the suite with exactly these three algorithms, or
// nil.
func findCipherSuite(auth, integrity, confidentiality byte) *cipherSuite {
	for _, suite := range cipherSuites {
		if suite.auth == auth && suite.integrity == integrity && suite.confidentiality == confidentiality {
			return suite
		}
	}
	return nil
}

func (cs *cipherSuite) mac(key []byte, parts ...[]byte) []byte {
	h := hmac.New(cs.hash, key)
	for _, part := range parts {
		h.Write(part)
	}
	return h.Sum(nil)
}

// rakpInput is what the RAKP messages and the SIK have in common: the
// random numbers, the session IDs and the user, each as it went over the
// wire. Session IDs are little-endian there, as every IPMI integer is.
type rakpInput struct {
	consoleID uint32   // SIDm, the remote console's session ID
	systemID  uint32   // SIDc, the managed system's session ID
	rm        [16]byte // the remote console's random number
	rc        [16]byte // the managed system's random number
	guid      [16]byte // GUIDc, the managed system's GUID
	role      byte     // ROLEm, the whole byte from RAKP 1, lookup bit included
	username  []byte
}

func le32(v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return b[:]
}

func (in *rakpInput) roleAndName() []byte {
	out := make([]byte, 0, 2+len(in.username))
	out = append(out, in.role, byte(len(in.username)))
	return append(out, in.username...)
}

// rakp2AuthCode is the key exchange authentication code of RAKP message 2:
// HMAC under the user's password of SIDm, SIDc, Rm, Rc, GUIDc, ROLEm,
// ULENGTHm and UNAMEm (IPMI 2.0 section 13.31).
func (cs *cipherSuite) rakp2AuthCode(kuid []byte, in *rakpInput) []byte {
	return cs.mac(kuid, le32(in.consoleID), le32(in.systemID), in.rm[:], in.rc[:], in.guid[:], in.roleAndName())
}

// rakp3AuthCode is the key exchange authentication code the remote console
// sends in RAKP message 3: HMAC under the user's password of Rc, SIDm,
// ROLEm, ULENGTHm and UNAMEm.
func (cs *cipherSuite) rakp3AuthCode(kuid []byte, in *rakpInput) []byte {
	return cs.mac(kuid, in.rc[:], le32(in.consoleID), in.roleAndName())
}

// sik is the session integrity key: HMAC under KG of Rm, Rc, ROLEm,
// ULENGTHm and UNAMEm. KG is never set on this board, and the
// specification says to use the user's password in its place then.
func (cs *cipherSuite) sik(kg []byte, in *rakpInput) []byte {
	return cs.mac(kg, in.rm[:], in.rc[:], in.roleAndName())
}

// rakp4ICV is the integrity check value of RAKP message 4: HMAC under the
// SIK of Rm, SIDc and GUIDc, cut to the suite's length.
func (cs *cipherSuite) rakp4ICV(sik []byte, in *rakpInput) []byte {
	return cs.mac(sik, in.rm[:], le32(in.systemID), in.guid[:])[:cs.icvLen]
}

// The constants K1 and K2 are derived from. The specification and ipmitool
// use 20 bytes for both suites, not the hash's length.
var (
	const1 = []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	const2 = []byte{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
)

// k1 keys the integrity HMAC of every message in the session.
func (cs *cipherSuite) k1(sik []byte) []byte {
	return cs.mac(sik, const1)
}

// k2 keys the encryption. AES-CBC-128 takes its first 16 bytes.
func (cs *cipherSuite) k2(sik []byte) []byte {
	return cs.mac(sik, const2)
}

// authCode is the integrity trailer's AuthCode over data, which runs from
// the session header's first byte to the Next Header byte.
func (cs *cipherSuite) authCode(k1, data []byte) []byte {
	return cs.mac(k1, data)[:cs.authCodeLen]
}

var errBadCiphertext = errors.New("bad ciphertext")

// encryptPayload is AES-CBC-128 as IPMI 2.0 section 13.29 lays it out: a
// random IV, then the payload, pad bytes counting up from 1, and the pad
// length, all in whole blocks.
func encryptPayload(k2, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k2[:16])
	if err != nil {
		return nil, err
	}

	padLen := 0
	if mod := (len(plaintext) + 1) % aes.BlockSize; mod != 0 {
		padLen = aes.BlockSize - mod
	}
	padded := make([]byte, 0, len(plaintext)+padLen+1)
	padded = append(padded, plaintext...)
	for i := 1; i <= padLen; i++ {
		padded = append(padded, byte(i))
	}
	padded = append(padded, byte(padLen))

	out := make([]byte, aes.BlockSize+len(padded))
	if _, err := rand.Read(out[:aes.BlockSize]); err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(block, out[:aes.BlockSize]).CryptBlocks(out[aes.BlockSize:], padded)
	return out, nil
}

// decryptPayload undoes encryptPayload. The pad has to be the one the
// specification describes; anything else is refused.
func decryptPayload(k2, payload []byte) ([]byte, error) {
	if len(payload) < 2*aes.BlockSize || len(payload)%aes.BlockSize != 0 {
		return nil, errBadCiphertext
	}
	block, err := aes.NewCipher(k2[:16])
	if err != nil {
		return nil, err
	}

	plain := make([]byte, len(payload)-aes.BlockSize)
	cipher.NewCBCDecrypter(block, payload[:aes.BlockSize]).CryptBlocks(plain, payload[aes.BlockSize:])

	padLen := int(plain[len(plain)-1])
	if padLen >= aes.BlockSize || padLen+1 > len(plain) {
		return nil, errBadCiphertext
	}
	end := len(plain) - 1 - padLen
	for i := 0; i < padLen; i++ {
		if plain[end+i] != byte(i+1) {
			return nil, errBadCiphertext
		}
	}
	return plain[:end], nil
}
