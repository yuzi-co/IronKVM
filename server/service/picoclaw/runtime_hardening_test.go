package picoclaw

import (
	"encoding/json"
	"testing"
)

// PicoClaw runs as root on the KVM and reads text from the controlled host's
// screen. The startup defaults are what keep it from having a shell, web
// access, skills, subagents or scheduled work, and they are the only thing
// that changes installs which already have a config.json from an older
// server that turned exec and remote access on.

// hardenedValues lists the settings the startup defaults must force.
var hardenedValues = map[string]any{
	"agents.defaults.restrict_to_workspace":        true,
	"agents.defaults.allow_read_outside_workspace": false,
	"tools.exec.enabled":                           false,
	"tools.exec.allow_remote":                      false,
	"tools.exec.enable_deny_patterns":              true,
	"tools.web.enabled":                            false,
	"tools.web_fetch.enabled":                      false,
	"tools.skills.enabled":                         false,
	"tools.find_skills.enabled":                    false,
	"tools.install_skill.enabled":                  false,
	"tools.spawn.enabled":                          false,
	"tools.spawn_status.enabled":                   false,
	"tools.subagent.enabled":                       false,
	"tools.cron.enabled":                           false,
	"tools.cron.allow_command":                     false,
	"heartbeat.enabled":                            false,
	// The KVM control path: MCP carries kvm_screenshot and kvm_actions.
	"tools.mcp.enabled": true,
}

// permissiveConfig is a config.json as an older server left it, with the
// unsafe values it used to write and a few the user set.
const permissiveConfig = `{
  "version": 3,
  "agents": {"defaults": {
    "restrict_to_workspace": false,
    "allow_read_outside_workspace": true,
    "model_name": "qwen"
  }},
  "tools": {
    "exec": {"enabled": true, "allow_remote": true, "enable_deny_patterns": false, "timeout_seconds": 90},
    "web": {"enabled": true, "prefer_native": true},
    "web_fetch": {"enabled": true},
    "skills": {"enabled": true},
    "find_skills": {"enabled": true},
    "install_skill": {"enabled": true},
    "spawn": {"enabled": true},
    "subagent": {"enabled": true},
    "cron": {"enabled": true, "allow_command": true},
    "mcp": {"enabled": true}
  },
  "heartbeat": {"enabled": true, "interval": 30}
}`

func parseRawConfig(t *testing.T, data string) map[string]any {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	return raw
}

func lookupDotted(raw map[string]any, dotted string) (any, bool) {
	var current any = raw
	start := 0
	for i := 0; i <= len(dotted); i++ {
		if i < len(dotted) && dotted[i] != '.' {
			continue
		}
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = obj[dotted[start:i]]
		if !ok {
			return nil, false
		}
		start = i + 1
	}
	return current, true
}

func TestStartupDefaultsHardenAnExistingPermissiveConfig(t *testing.T) {
	editor := &picoclawConfigEditor{raw: parseRawConfig(t, permissiveConfig)}

	applyPicoclawNanoKVMDefaults(editor)

	for path, want := range hardenedValues {
		got, ok := lookupDotted(editor.raw, path)
		if !ok || got != want {
			t.Errorf("%s = %v (present %v), want %v", path, got, ok, want)
		}
	}
	if !editor.changed {
		t.Fatal("the editor did not record a change, so the hardened config is never saved")
	}

	// Settings the defaults do not own survive.
	if got, _ := lookupDotted(editor.raw, "tools.exec.timeout_seconds"); got != float64(90) {
		t.Errorf("tools.exec.timeout_seconds = %v, want the user's 90", got)
	}
	if got, _ := lookupDotted(editor.raw, "agents.defaults.model_name"); got != "qwen" {
		t.Errorf("agents.defaults.model_name = %v, want the user's qwen", got)
	}
}

func TestStartupDefaultsHardenAFreshConfig(t *testing.T) {
	editor := &picoclawConfigEditor{raw: map[string]any{}}

	applyPicoclawNanoKVMDefaults(editor)

	for path, want := range hardenedValues {
		if got, ok := lookupDotted(editor.raw, path); !ok || got != want {
			t.Errorf("%s = %v (present %v), want %v", path, got, ok, want)
		}
	}
}

func TestStartupDefaultsKeepKVMControlTools(t *testing.T) {
	// kvm_screenshot results arrive as media and load_image/read_file read
	// them; the defaults must never switch those tools off.
	for _, entry := range picoclawNanoKVMDefaults {
		if len(entry.path) != 3 || entry.path[0] != "tools" || entry.path[2] != "enabled" {
			continue
		}
		switch entry.path[1] {
		case "mcp", "load_image", "read_file", "list_dir":
			if entry.value != true {
				t.Errorf("tools.%s.enabled is forced to %v", entry.path[1], entry.value)
			}
		}
	}
}

func TestStartupDefaultsLeaveAHardenedConfigUnchanged(t *testing.T) {
	// The config lives on the SD card; a start with nothing to change must
	// not rewrite it. The second pass sees the config as json.Unmarshal
	// returns it, with every number a float64.
	first := &picoclawConfigEditor{raw: parseRawConfig(t, permissiveConfig)}
	applyPicoclawNanoKVMDefaults(first)

	data, err := json.Marshal(first.raw)
	if err != nil {
		t.Fatal(err)
	}
	second := &picoclawConfigEditor{raw: parseRawConfig(t, string(data))}
	applyPicoclawNanoKVMDefaults(second)

	if second.changed {
		t.Fatal("a second start rewrote a config that already had the defaults")
	}
}

func TestSetValueTreatsAnUnmarshalledNumberAsEqual(t *testing.T) {
	editor := &picoclawConfigEditor{raw: parseRawConfig(t, `{"gateway": {"port": 18790}}`)}

	editor.setValue(18790, "gateway", "port")

	if editor.changed {
		t.Fatal("float64(18790) from the file was treated as different from int 18790")
	}

	editor.setValue(18791, "gateway", "port")
	if !editor.changed {
		t.Fatal("a real port change was not recorded")
	}
}
