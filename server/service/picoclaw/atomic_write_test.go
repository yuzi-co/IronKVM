package picoclaw

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteFileAtomicReplacesFileAndLeavesNoTemporary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := writeFileAtomic(path, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("content = %q, want new", data)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Fatalf("mode = %o, want 600", mode)
		}
	}
	assertOnlyEntries(t, dir, "config.json")
}

func TestWriteFileAtomicKeepsOldFileWhenReplaceFails(t *testing.T) {
	dir := t.TempDir()
	// A directory in the target's place makes the final rename fail.
	target := filepath.Join(dir, "config.json")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "keep"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := writeFileAtomic(target, []byte("new"), 0o600); err == nil {
		t.Fatal("expected an error when the target is a non-empty directory")
	}
	if _, err := os.Stat(filepath.Join(target, "keep")); err != nil {
		t.Fatalf("existing target was disturbed: %v", err)
	}
	assertOnlyEntries(t, dir, "config.json")
}

func TestWriteFileAtomicFailsForMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "config.json")
	if err := writeFileAtomic(path, []byte("new"), 0o600); err == nil {
		t.Fatal("expected an error for a missing directory")
	}
}

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
