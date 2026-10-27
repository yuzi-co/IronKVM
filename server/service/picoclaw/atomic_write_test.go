package picoclaw

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPicoclawConfigDocumentSavesAtomically(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PICOCLAW_HOME", home)
	configPath := filepath.Join(home, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"agents":{"defaults":{"model_name":"a"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	doc, err := loadPicoclawConfigDocument()
	if err != nil {
		t.Fatal(err)
	}
	doc.raw["gateway"] = map[string]any{"port": 18790}
	doc.security.ChannelList = map[string]picoclawChannelSecurityEntry{
		"pico": {Settings: &picoclawChannelSecuritySettings{Token: "placeholder"}},
	}
	if err := doc.saveConfig(); err != nil {
		t.Fatal(err)
	}
	if err := doc.saveSecurity(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := loadPicoclawConfigDocument()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.config.Gateway.Port != 18790 {
		t.Fatalf("gateway port = %d, want 18790", reloaded.config.Gateway.Port)
	}
	if reloaded.resolvedGatewayToken() != "placeholder" {
		t.Fatalf("token = %q", reloaded.resolvedGatewayToken())
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(home, ".security.yml"))
		if err != nil {
			t.Fatal(err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Fatalf(".security.yml mode = %o, want 600", mode)
		}
	}
	assertOnlyEntries(t, home, ".security.yml", "config.json")
}

func assertOnlyEntries(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("entries = %v, want %v", names, want)
	}
}
