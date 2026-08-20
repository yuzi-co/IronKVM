package vm

import (
	"errors"
	"testing"
)

// Switching HTTPS on used to mint a certificate every time, unconditionally.
//
// That is not a small waste. A self-signed certificate has to be trusted by
// hand, and replacing it voids the trust the operator installed, so every
// switch cost them a manual reinstall on every machine that talks to the board.
// Switching HTTPS off leaves the certificate on disk, so the one they already
// trust is sitting right there to be reused.
//
// utils.EnsureCert keeps a certificate that still answers for the device and
// replaces one that does not. These tests pin that enableTls goes through it,
// because reverting to GenerateCert would compile, pass every certificate test
// in utils, and quietly bring the manual step back.

func TestEnablingTlsGoesThroughTheCertificateThatMayBeKept(t *testing.T) {
	original := ensureTlsCert
	t.Cleanup(func() { ensureTlsCert = original })

	called := 0
	ensureTlsCert = func() error {
		called++
		return nil
	}

	// config.Write reaches the real configuration file, so this stops at the
	// first step. The certificate decision is what is under test and it happens
	// before anything is written.
	_ = enableTls()

	if called != 1 {
		t.Fatalf("enableTls consulted the certificate %d time(s), want 1", called)
	}
}

func TestEnablingTlsStopsWhenTheCertificateCannotBeMade(t *testing.T) {
	original := ensureTlsCert
	t.Cleanup(func() { ensureTlsCert = original })

	want := errors.New("no certificate")
	ensureTlsCert = func() error { return want }

	// Writing proto: https with no usable certificate leaves a board that
	// starts a TLS listener it cannot serve from, reachable on neither scheme.
	if err := enableTls(); !errors.Is(err, want) {
		t.Fatalf("enableTls returned %v, want %v", err, want)
	}
}
