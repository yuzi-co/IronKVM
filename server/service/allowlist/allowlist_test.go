package allowlist

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// use puts list in force for one test, and Exempt back as it was after it.
func use(t *testing.T, entries ...string) {
	t.Helper()
	list, err := Parse(entries)
	if err != nil {
		t.Fatalf("Parse(%q): %s", entries, err)
	}
	previous := prefixes()
	exempt := Exempt
	set(list)
	t.Cleanup(func() {
		set(previous)
		Exempt = exempt
	})
}

func TestParseReadsCIDRsAndSingleAddresses(t *testing.T) {
	list, err := Parse([]string{" 10.0.0.0/8 ", "192.168.1.7", "2001:db8::/32", "", "10.1.2.3/16", "::ffff:172.16.0.0/108"})
	if err != nil {
		t.Fatal(err)
	}
	got := format(list)
	want := []string{"10.0.0.0/8", "192.168.1.7", "2001:db8::/32", "10.1.0.0/16", "172.16.0.0/12"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseRefusesWhatIsNotAnAddress(t *testing.T) {
	for _, entry := range []string{"10.0.0", "example.com", "10.0.0.0/33", "fe80::1%eth0", "::ffff:10.0.0.0/64"} {
		if _, err := Parse([]string{entry}); err == nil {
			t.Errorf("Parse accepted %q", entry)
		}
	}
	many := make([]string, maxEntries+1)
	for i := range many {
		many[i] = "10.0.0.1"
	}
	if _, err := Parse(many); err == nil {
		t.Errorf("Parse accepted %d entries", len(many))
	}
}

func TestFormatDropsDuplicates(t *testing.T) {
	list, _ := Parse([]string{"10.0.0.1", "10.0.0.1/32", "10.0.0.0/8", "10.9.9.9/8"})
	if got := format(list); strings.Join(got, ",") != "10.0.0.1,10.0.0.0/8" {
		t.Fatalf("got %v", got)
	}
}

func TestAnEmptyListAllowsEveryAddress(t *testing.T) {
	use(t)
	for _, a := range []string{"203.0.113.9", "2001:db8::1", "10.0.0.1"} {
		if !Allowed(netip.MustParseAddr(a)) {
			t.Errorf("%s refused by an empty list", a)
		}
	}
}

func TestTheListDecides(t *testing.T) {
	use(t, "10.0.0.0/24", "2001:db8::/64")
	Exempt = nil

	for a, want := range map[string]bool{
		"10.0.0.5":           true,
		"::ffff:10.0.0.5":    true,
		"10.0.1.5":           false,
		"2001:db8::9":        true,
		"2001:db8:1::9":      false,
		"127.0.0.1":          true,
		"::1":                true,
		"::ffff:127.0.0.1":   true,
		"192.168.1.1":        false,
		"::ffff:192.168.1.1": false,
	} {
		if got := Allowed(netip.MustParseAddr(a)); got != want {
			t.Errorf("Allowed(%s) = %v, want %v", a, got, want)
		}
	}
}

func TestExemptLetsInWhatTheListRefuses(t *testing.T) {
	use(t, "10.0.0.0/24")
	usb := netip.MustParsePrefix("172.31.255.0/30")
	Exempt = func(a netip.Addr) bool { return usb.Contains(a) }

	if !Allowed(netip.MustParseAddr("172.31.255.2")) {
		t.Fatal("the exempt address was refused")
	}
	if Allowed(netip.MustParseAddr("172.31.254.2")) {
		t.Fatal("an address outside both was allowed")
	}
}

func TestAllowedAddrReadsListenerAddresses(t *testing.T) {
	use(t, "10.0.0.0/24")
	Exempt = nil

	if !AllowedAddr(&net.TCPAddr{IP: net.ParseIP("10.0.0.9"), Port: 5900}) {
		t.Error("an allowed VNC client was refused")
	}
	if AllowedAddr(&net.TCPAddr{IP: net.ParseIP("10.0.1.9"), Port: 5900}) {
		t.Error("a refused VNC client was allowed")
	}
	if AllowedAddr(&net.UDPAddr{IP: net.ParseIP("10.0.1.9"), Port: 623}) {
		t.Error("a refused IPMI client was allowed")
	}
	if !AllowedAddr(&net.UnixAddr{Name: "/run/sock", Net: "unix"}) {
		t.Error("a unix socket peer was refused")
	}
}

func serve(t *testing.T, method, path, remote, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware())
	service := NewService()
	r.GET("/api/network/allowlist", service.GetSettings)
	r.POST("/api/network/allowlist", service.SetSettings)
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ui") })

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = remote
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestTheMiddlewareRefusesAddressesOffTheList(t *testing.T) {
	use(t, "10.0.0.0/24")
	Exempt = nil

	if w := serve(t, http.MethodGet, "/", "10.0.0.7:40000", ""); w.Code != http.StatusOK {
		t.Fatalf("an allowed client got %d", w.Code)
	}
	w := serve(t, http.MethodGet, "/", "10.0.1.7:40000", "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("a refused client got %d", w.Code)
	}
	var rsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil || rsp.Code == 0 || rsp.Msg == "" {
		t.Fatalf("the refusal is not the envelope: %q", w.Body.String())
	}
	if w := serve(t, http.MethodGet, "/", "[::1]:40000", ""); w.Code != http.StatusOK {
		t.Fatalf("loopback got %d", w.Code)
	}
}

type setReply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data State  `json:"data"`
}

func saveTo(t *testing.T, saved *[]string, fail error) {
	t.Helper()
	original := save
	save = func(networks []string) error {
		if fail != nil {
			return fail
		}
		*saved = append([]string(nil), networks...)
		return nil
	}
	t.Cleanup(func() { save = original })
}

func setList(t *testing.T, remote string, networks ...string) setReply {
	t.Helper()
	body, _ := json.Marshal(map[string][]string{"networks": networks})
	w := serve(t, http.MethodPost, "/api/network/allowlist", remote, string(body))
	var rsp setReply
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("answered %d %q", w.Code, w.Body.String())
	}
	return rsp
}

func TestASaveThatWouldLockOutTheCallerIsRefused(t *testing.T) {
	use(t)
	Exempt = nil
	var saved []string
	saveTo(t, &saved, nil)

	rsp := setList(t, "192.168.1.20:5000", "10.0.0.0/8")
	if rsp.Code == 0 {
		t.Fatal("a list without the caller was saved")
	}
	if !strings.Contains(rsp.Msg, "192.168.1.20") {
		t.Fatalf("the refusal should name the caller's address: %q", rsp.Msg)
	}
	if saved != nil || len(prefixes()) != 0 {
		t.Fatal("the refused list was saved or put in force")
	}
	// Still reachable.
	if w := serve(t, http.MethodGet, "/", "192.168.1.20:5000", ""); w.Code != http.StatusOK {
		t.Fatalf("the caller was locked out: %d", w.Code)
	}
}

func TestASaveThatKeepsTheCallerIsApplied(t *testing.T) {
	use(t)
	Exempt = nil
	var saved []string
	saveTo(t, &saved, nil)

	rsp := setList(t, "192.168.1.20:5000", "192.168.1.0/24", "10.0.0.1")
	if rsp.Code != 0 {
		t.Fatalf("refused: %+v", rsp)
	}
	if strings.Join(saved, ",") != "192.168.1.0/24,10.0.0.1" {
		t.Fatalf("saved %v", saved)
	}
	if rsp.Data.Client != "192.168.1.20" || len(rsp.Data.Networks) != 2 {
		t.Fatalf("the reply is %+v", rsp.Data)
	}
	if w := serve(t, http.MethodGet, "/", "10.0.0.2:5000", ""); w.Code != http.StatusForbidden {
		t.Fatalf("an address off the new list got %d", w.Code)
	}

	// Clearing it allows everyone again.
	if rsp := setList(t, "192.168.1.20:5000"); rsp.Code != 0 || len(saved) != 0 {
		t.Fatalf("clearing the list: %+v, saved %v", rsp, saved)
	}
	if w := serve(t, http.MethodGet, "/", "10.0.0.2:5000", ""); w.Code != http.StatusOK {
		t.Fatalf("an empty list still refused a client: %d", w.Code)
	}
}

func TestAnInvalidEntryIsRefusedByName(t *testing.T) {
	use(t)
	var saved []string
	saveTo(t, &saved, nil)

	rsp := setList(t, "192.168.1.20:5000", "192.168.1.0/24", "not-an-address")
	if rsp.Code == 0 || !strings.Contains(rsp.Msg, "not-an-address") {
		t.Fatalf("got %+v", rsp)
	}
}

func TestAFailedSaveChangesNothing(t *testing.T) {
	use(t)
	var saved []string
	saveTo(t, &saved, errors.New("read-only file system"))

	if rsp := setList(t, "192.168.1.20:5000", "192.168.1.0/24"); rsp.Code == 0 {
		t.Fatal("a failed save reported success")
	}
	if len(prefixes()) != 0 {
		t.Fatal("a list that was not saved was put in force")
	}
}
