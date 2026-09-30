package vm

import (
	"testing"

	"NanoKVM-Server/proto"
)

func TestSetKeyJigglerReqValidation(t *testing.T) {
	valid := []proto.SetKeyJigglerReq{
		{Enabled: true, Key: "f15"},
		{Enabled: true, Key: "shift"},
		{Enabled: false, Key: "ctrl"},
		// No key keeps the one already chosen.
		{Enabled: true},
		{Enabled: false},
	}
	for _, req := range valid {
		if err := proto.ValidateRequest(&req); err != nil {
			t.Fatalf("%+v: unexpected error %s", req, err)
		}
	}

	invalid := []proto.SetKeyJigglerReq{
		{Enabled: true, Key: "alt"},
		{Enabled: true, Key: "F15"},
		{Enabled: false, Key: "mouse"},
	}
	for _, req := range invalid {
		if err := proto.ValidateRequest(&req); err == nil {
			t.Fatalf("%+v: expected a validation error", req)
		}
	}
}
