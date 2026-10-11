package mcpservice

import (
	"context"
	"errors"
	"testing"
)

func TestWakeHandler(t *testing.T) {
	old := wakeOnLAN
	t.Cleanup(func() { wakeOnLAN = old })

	var gotMac, gotName string
	wakeOnLAN = func(mac, name string) (string, error) {
		gotMac, gotName = mac, name
		return "AA:BB:CC:DD:EE:01", nil
	}

	handler := wakeHandler()
	_, result, err := handler(context.Background(), nil, WakeOnLANParams{Name: "nas"})
	if err != nil || !result.OK || result.Mac != "AA:BB:CC:DD:EE:01" || gotName != "nas" || gotMac != "" {
		t.Fatalf("result = %+v, err = %v, mac=%q name=%q", result, err, gotMac, gotName)
	}

	if _, _, err := handler(context.Background(), nil, WakeOnLANParams{}); err == nil {
		t.Fatal("an empty request was accepted")
	}

	wakeOnLAN = func(string, string) (string, error) { return "", errors.New("no such device") }
	if _, _, err := handler(context.Background(), nil, WakeOnLANParams{Name: "tv"}); err == nil {
		t.Fatal("a failed wake returned no error")
	}
}
