package unixcrypt

import (
	"errors"
	"testing"
)

// Vectors from Drepper's specification where it has them, and otherwise from
// `openssl passwd -1/-5/-6 -salt <salt> <password>`, which implements the same
// schemes independently.
var vectors = []struct {
	password string
	hash     string
}{
	{"Hello world!", "$6$saltstring$svn8UoSVapNtMuq1ukKS4tPQd8iKwSMHWjl/O817G3uBnIFNjnQJuesI68u4OTLiBFdcbYEdFCoEOfaS35inz1"},
	{"Hello world!", "$6$rounds=10000$saltstringsaltst$OW1/O6BYHV6BcXZu8QVeXbDWra3Oeqh0sbHbbMCVNSnCM/UrjmM0Dp8vOuZeHBy/YTBmSK6H9qs/y3RnOaw5v."},
	{"x", "$6$toolongsaltstrin$m5UCdGkMg16fnZB/afayhHcFYEQvnTRoKod8GIJKB0rGGl9IQGUPKeGHHw5xPOFVPjhTAcuzioT6ZcDqsMKJ80"},
	{"Hello world!", "$5$saltstring$5B8vYYiY.CVt1RlTTf8KbXBH3hsxY/GNooZaBBGWEc5"},
	{"Hello world!", "$5$rounds=10000$saltstringsaltst$3xv.VbSHBb41AL9AvLeujZkZRBAwqFMz2.opqey6IcA"},
	{"the minimum number is still observed", "$5$rounds=1000$roundstoolow$yfvwcWrQ8l/K0DAWyuPMDNHpIVlTQebY9l/gL972bIC"},
	{"a much longer password that exceeds thirty two bytes easily yes", "$5$ab$EpR/XjNWlnIXOw.eYDLK.RGZZGAh99ak82GUm2IjIO9"},
	{"Hello world!", "$1$saltstri$YMyguxXMBpd2TEZ.vS/3q1"},
	{"", "$1$abcdefgh$M55TzYaaccxVGbptZWaxX/"},
	{"a password longer than sixteen bytes", "$1$x$HsMtwpZ51PhVKsQCJ8q181"},
}

func TestVerifyAcceptsKnownVectors(t *testing.T) {
	for _, v := range vectors {
		ok, err := Verify(v.password, v.hash)
		if err != nil || !ok {
			got, _ := Crypt(v.password, v.hash)
			t.Errorf("Verify(%q, %s) = %v, %v; crypt gave %s", v.password, v.hash, ok, err, got)
		}
	}
}

func TestVerifyRejectsAWrongPassword(t *testing.T) {
	for _, v := range vectors {
		if ok, err := Verify(v.password+"!", v.hash); err != nil || ok {
			t.Errorf("Verify accepted a wrong password for %s (err %v)", v.hash, err)
		}
	}
}

// Rounds below the minimum are raised to it, as the specification says; the
// vector above shows the output then names 1000, not what was asked.
func TestCryptRaisesRoundsToTheMinimum(t *testing.T) {
	got, err := Crypt("the minimum number is still observed", "$5$rounds=10$roundstoolow")
	if err != nil {
		t.Fatal(err)
	}
	if want := "$5$rounds=1000$roundstoolow$yfvwcWrQ8l/K0DAWyuPMDNHpIVlTQebY9l/gL972bIC"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// A scheme this package cannot compute must say so rather than answer "no",
// or the caller would report a password as changed that may be the default.
func TestVerifyReportsSchemesItCannotCompute(t *testing.T) {
	for _, hash := range []string{
		"$y$j9T$abcdefgh$0123456789abcdefghijklmnopqrstuvwxyzABCDEF",
		"$2b$10$abcdefghijklmnopqrstuu",
		"abJnggxhB/yWI",
		"",
		"!",
	} {
		if _, err := Verify("root", hash); !errors.Is(err, ErrUnsupported) {
			t.Errorf("Verify(%q) err = %v, want ErrUnsupported", hash, err)
		}
	}
}
