package nut

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSplitReadsQuotedWords(t *testing.T) {
	cases := map[string][]string{
		`VAR ups ups.status "OL CHRG"`:      {"VAR", "ups", "ups.status", "OL CHRG"},
		`VAR ups ups.mfr "Back \"Ups\" \\"`: {"VAR", "ups", "ups.mfr", `Back "Ups" \`},
		`OK`:                                {"OK"},
		`VAR ups x ""`:                      {"VAR", "ups", "x", ""},
		`  END   LIST VAR ups `:             {"END", "LIST", "VAR", "ups"},
	}
	for line, want := range cases {
		got, err := split(line)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %q, %v; want %q", line, got, err, want)
		}
	}
	if _, err := split(`VAR ups x "open`); err == nil {
		t.Error("an unterminated quote was accepted")
	}
}

func TestQuoteRefusesAControlCharacter(t *testing.T) {
	if _, err := quote("a\nLOGOUT"); err == nil {
		t.Fatal("a newline went into a command")
	}
	if got, _ := quote(`p w"d`); got != `"p w\"d"` {
		t.Fatalf("quoted as %s", got)
	}
	if got, _ := quote("plain"); got != "plain" {
		t.Fatalf("quoted as %s", got)
	}
}

func dialFake(t *testing.T, s *fakeServer) *Client {
	t.Helper()
	c, err := Dial(context.Background(), s.addr(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestVarsReadsTheList(t *testing.T) {
	s := newFakeServer(t)
	vars, err := dialFake(t, s).Vars("ups")
	if err != nil {
		t.Fatal(err)
	}
	if vars["ups.status"] != "OL" || vars["battery.charge"] != "100" || vars["ups.mfr"] != `Back "Ups" \ Co` {
		t.Fatalf("%v", vars)
	}
}

func TestVarsOfAnUnknownUPS(t *testing.T) {
	s := newFakeServer(t)
	_, err := dialFake(t, s).Vars("nope")
	var se *ServerError
	if !errors.As(err, &se) || se.Code != "UNKNOWN-UPS" {
		t.Fatalf("%v", err)
	}
}

func TestLoginNeedsTheRightPassword(t *testing.T) {
	s := newFakeServer(t)
	s.user, s.password = "kvm", "secret"

	c := dialFake(t, s)
	if err := c.Authenticate("kvm", "wrong"); err != nil {
		t.Fatal(err)
	}
	var se *ServerError
	if err := c.Login("ups"); !errors.As(err, &se) || se.Code != "ACCESS-DENIED" {
		t.Fatalf("%v", err)
	}

	c = dialFake(t, s)
	if err := c.Authenticate("kvm", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := c.Login("ups"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the login", func() bool { return s.loggedIn() == 1 })

	if err := c.Logout(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the logout", func() bool { return s.loggedIn() == 0 })
}

func TestAPasswordWithASpaceCrossesTheWire(t *testing.T) {
	s := newFakeServer(t)
	s.user, s.password = "kvm", "pass word"

	c := dialFake(t, s)
	if err := c.Authenticate("kvm", "pass word"); err != nil {
		t.Fatal(err)
	}
	if err := c.Login("ups"); err != nil {
		t.Fatal(err)
	}
}

func TestAServerThatHangsUpIsAnError(t *testing.T) {
	s := newFakeServer(t)
	c := dialFake(t, s)
	s.close()
	if _, err := c.Vars("ups"); err == nil {
		t.Fatal("no error from a closed server")
	}
}

func TestReadingFromVars(t *testing.T) {
	r := readingFrom(map[string]string{"ups.status": "OB LB", "battery.charge": "42.5", "battery.runtime": "120"})
	if r.State() != UPSLowBattery || !r.Has("OB") || r.Charge != 42.5 || r.Runtime != 120 {
		t.Fatalf("%+v", r)
	}
	r = readingFrom(map[string]string{"ups.status": "OL"})
	if r.State() != UPSOnline || r.Charge != -1 || r.Runtime != -1 {
		t.Fatalf("%+v", r)
	}
	if readingFrom(map[string]string{"ups.status": "OB FSD"}).State() != UPSShutdown {
		t.Fatal("FSD is not a shutdown")
	}
}
