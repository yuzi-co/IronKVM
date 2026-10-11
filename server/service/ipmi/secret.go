package ipmi

import "NanoKVM-Server/authn"

// KeyFile holds the key IPMI passwords are sealed under.
const KeyFile = "/etc/kvm/ipmi.key"

var errNoKey = authn.ErrNoKey

// Keyring seals IPMI passwords for the account file and opens them again for
// the RAKP handshake. It is authn's keyring under its own key file; see
// authn.Keyring for what the sealing does and does not protect against.
type Keyring = authn.Keyring

func NewKeyring(path string) *Keyring {
	return authn.NewKeyring(path)
}
