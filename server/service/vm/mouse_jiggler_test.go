package vm

import (
	"testing"

	"NanoKVM-Server/proto"
)

func TestJigglerKeyFromRequest(t *testing.T) {
	cases := []struct {
		method, key, current, want string
	}{
		// Clients that predate the method keep whatever the jiggler has.
		{"", "", "shift", "shift"},
		{"", "", "", ""},
		{"mouse", "f15", "shift", ""},
		{"key", "", "", "f15"},
		{"key", "ctrl", "", "ctrl"},
	}
	for _, c := range cases {
		if got := jigglerKey(c.method, c.key, c.current); got != c.want {
			t.Fatalf("jigglerKey(%q, %q, %q) = %q, want %q", c.method, c.key, c.current, got, c.want)
		}
	}
}

func TestSetMouseJigglerReqValidation(t *testing.T) {
	valid := []proto.SetMouseJigglerReq{
		{Enabled: true, Mode: "relative"},
		{Enabled: true, Mode: "relative", Method: "key", Key: "f15"},
		{Enabled: false, Method: "mouse"},
	}
	for _, req := range valid {
		if err := proto.ValidateRequest(&req); err != nil {
			t.Fatalf("%+v: unexpected error %s", req, err)
		}
	}

	invalid := []proto.SetMouseJigglerReq{
		{Enabled: true, Method: "keyboard"},
		{Enabled: true, Method: "key", Key: "alt"},
	}
	for _, req := range invalid {
		if err := proto.ValidateRequest(&req); err == nil {
			t.Fatalf("%+v: expected a validation error", req)
		}
	}
}
