package vm

import (
	"encoding/json"
	"testing"
)

// TestVirtualDeviceStateKeepsSipeedsBooleans pins the flat booleans of GET
// /api/vm/device/virtual. Clients written against Sipeed's API, Home
// Assistant's nanokvm integration among them, read "network" and "disk" as
// booleans and stop polling the whole device when they are anything else
// (ironkvm-dist#124).
func TestVirtualDeviceStateKeepsSipeedsBooleans(t *testing.T) {
	disk, ok := functionForDevice("disk")
	if !ok || len(disk.markers) == 0 {
		t.Fatal("the disk function has no marker")
	}
	present := func(marker string) bool {
		for _, m := range disk.markers {
			if m == marker {
				return true
			}
		}
		return false
	}

	body, err := json.Marshal(virtualDeviceState(present))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}

	for key, want := range map[string]bool{"network": false, "disk": true, "audio": false} {
		value, isBool := got[key].(bool)
		if !isBool {
			t.Errorf("%q is %T, want a boolean: %s", key, got[key], body)
			continue
		}
		if value != want {
			t.Errorf("%q is %v, want %v", key, value, want)
		}
	}

	functions, isObject := got["functions"].(map[string]any)
	if !isObject {
		t.Fatalf("functions is %T, want an object: %s", got["functions"], body)
	}
	for _, name := range []string{"console", "network", "disk", "audio"} {
		state, isObject := functions[name].(map[string]any)
		if !isObject {
			t.Errorf("functions.%s is %T, want an object", name, functions[name])
			continue
		}
		if enabled, _ := state["enabled"].(bool); enabled != (name == "disk") {
			t.Errorf("functions.%s.enabled is %v", name, state["enabled"])
		}
	}
}
