package vnc

import (
	"bufio"
	"crypto/des"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

// maxCredentialLength bounds the username and the password a client may send.
// An account name is at most 32 characters.
const maxCredentialLength = 1024

// vncPasswordLength is how much of a VNC password DES uses.
const vncPasswordLength = 8

// errAuthFailed is the reason a client is given for any failed login, so the
// answer does not tell an account that exists from one that does not.
var errAuthFailed = errors.New("authentication failed")

// silentError is a handshake failure after which the protocol has no place
// for a SecurityResult: VeNCrypt refuses its version or subtype with a byte of
// its own, and after a failed TLS handshake the stream is neither plain nor
// encrypted. The server closes the connection without another word.
type silentError struct{ err error }

func (e silentError) Error() string { return e.err.Error() }
func (e silentError) Unwrap() error { return e.err }

// authResult is who a handshake authenticated.
type authResult struct {
	user   string
	method string
}

// Authentication methods, as the state reports them.
const (
	methodVeNCrypt = "vencrypt"
	methodVNC      = "vnc"
)

// vncAuthKey turns a VNC password into its DES key: the first eight bytes,
// padded with zeros, with the bits of each byte reversed. The reversal is a
// quirk of the original implementation that every client repeats.
func vncAuthKey(password string) []byte {
	key := make([]byte, vncPasswordLength)
	copy(key, password)
	for i, b := range key {
		var r byte
		for bit := 0; bit < 8; bit++ {
			if b&(1<<bit) != 0 {
				r |= 1 << (7 - bit)
			}
		}
		key[i] = r
	}
	return key
}

// vncAuthResponse is the answer a client with the right password gives to the
// challenge: the challenge's two blocks encrypted with DES under the key.
func vncAuthResponse(password string, challenge []byte) ([]byte, error) {
	block, err := des.NewCipher(vncAuthKey(password))
	if err != nil {
		return nil, err
	}
	response := make([]byte, 16)
	block.Encrypt(response[:8], challenge[:8])
	block.Encrypt(response[8:], challenge[8:16])
	return response, nil
}

// vncAuth runs plain VNC authentication: a random challenge and the client's
// DES answer.
func (s *session) vncAuth() (authResult, error) {
	password, ok, err := s.srv.deps.Password.Get()
	if err != nil || !ok {
		return authResult{}, errors.New("no VNC password is set")
	}

	challenge := make([]byte, 16)
	if _, err := rand.Read(challenge); err != nil {
		return authResult{}, err
	}
	if _, err := s.conn.Write(challenge); err != nil {
		return authResult{}, err
	}
	response := make([]byte, 16)
	if _, err := io.ReadFull(s.reader, response); err != nil {
		return authResult{}, err
	}

	// The limiter counts VNC password guesses under a name no account can
	// have, so they share the address's limit with the web login but lock no
	// account.
	const limiterName = ":vnc"
	if s.srv.deps.Limiter.Locked(s.ip, limiterName) {
		return authResult{}, errAuthFailed
	}
	expected, err := vncAuthResponse(password, challenge)
	if err != nil {
		return authResult{}, err
	}
	if subtle.ConstantTimeCompare(expected, response) != 1 {
		s.srv.deps.Limiter.Failed(s.ip, limiterName)
		s.srv.sleep(s.srv.deps.FailureDelay)
		return authResult{}, errAuthFailed
	}
	s.srv.deps.Limiter.Succeeded(s.ip, limiterName)
	return authResult{method: methodVNC}, nil
}

// vencrypt runs VeNCrypt 0.2 with the one subtype X509Plain: version
// agreement, the subtype, TLS with the server's certificate, and then a
// username and a password checked against the KVM accounts.
func (s *session) vencrypt() (authResult, error) {
	if _, err := s.conn.Write([]byte{0, 2}); err != nil {
		return authResult{}, err
	}
	version := make([]byte, 2)
	if _, err := io.ReadFull(s.reader, version); err != nil {
		return authResult{}, err
	}
	if version[0] != 0 || version[1] != 2 {
		_, _ = s.conn.Write([]byte{1})
		return authResult{}, silentError{fmt.Errorf("unsupported VeNCrypt version %d.%d", version[0], version[1])}
	}

	offer := []byte{0, 1}
	offer = binary.BigEndian.AppendUint32(offer, vencryptX509Plain)
	if _, err := s.conn.Write(offer); err != nil {
		return authResult{}, err
	}
	var subtype uint32
	if err := binary.Read(s.reader, binary.BigEndian, &subtype); err != nil {
		return authResult{}, err
	}
	if subtype != vencryptX509Plain {
		_, _ = s.conn.Write([]byte{0})
		return authResult{}, silentError{fmt.Errorf("unsupported VeNCrypt subtype %d", subtype)}
	}
	if _, err := s.conn.Write([]byte{1}); err != nil {
		return authResult{}, err
	}

	cert, err := s.srv.deps.Certificate()
	if err != nil {
		return authResult{}, silentError{fmt.Errorf("load the TLS certificate: %w", err)}
	}
	tlsConn := tls.Server(&bufferedConn{Conn: s.conn, r: s.reader}, &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MinVersion:   tls.VersionTLS12,
	})
	if err := tlsConn.Handshake(); err != nil {
		return authResult{}, silentError{fmt.Errorf("TLS handshake: %w", err)}
	}
	s.conn = tlsConn
	s.reader = bufio.NewReader(tlsConn)

	var lengths [2]uint32
	if err := binary.Read(s.reader, binary.BigEndian, &lengths); err != nil {
		return authResult{}, err
	}
	if lengths[0] > maxCredentialLength || lengths[1] > maxCredentialLength {
		return authResult{}, errors.New("credentials too long")
	}
	credentials := make([]byte, lengths[0]+lengths[1])
	if _, err := io.ReadFull(s.reader, credentials); err != nil {
		return authResult{}, err
	}
	username := string(credentials[:lengths[0]])
	password := string(credentials[lengths[0]:])

	if err := s.checkPassword(username, password); err != nil {
		return authResult{}, err
	}
	return authResult{user: username, method: methodVeNCrypt}, nil
}

// checkPassword checks an account's credentials under the web login's
// brute-force limit, and waits as long as the web login after a failure.
func (s *session) checkPassword(username, password string) error {
	deps := s.srv.deps
	if deps.Limiter.Locked(s.ip, username) {
		return errAuthFailed
	}

	user, ok, err := deps.Accounts.Authenticate(username, password)
	if err != nil {
		log.Errorf("vnc: load accounts: %s", err)
		return errors.New("authentication unavailable")
	}
	if !ok {
		deps.Limiter.Failed(s.ip, username)
		s.srv.sleep(deps.FailureDelay)
		return errAuthFailed
	}

	// Guesses sent at once all pass the first check before any of them has
	// failed, as in the Redfish login.
	if deps.Limiter.Locked(s.ip, username) {
		return errAuthFailed
	}
	deps.Limiter.Succeeded(s.ip, username)

	// The web UI makes such an account choose a password before it may do
	// anything else, and a VNC client has no way to.
	if user.MustChangePassword {
		return errors.New("change the password in the web UI first")
	}
	return nil
}
