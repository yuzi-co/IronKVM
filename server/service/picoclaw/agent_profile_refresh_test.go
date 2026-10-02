package picoclaw

import (
	"os"
	"path/filepath"
	"testing"
)

// The bundled profile files live in kvmapp/picoclaw at the repository root.
func bundledAgentProfile(t *testing.T, profile string) []byte {
	t.Helper()
	name := map[string]string{agentProfileDefault: "AGENT.md", agentProfileKVM: "AGENT_KVM.md"}[profile]
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "kvmapp", "picoclaw", name))
	if err != nil {
		t.Fatalf("read bundled %s: %v", name, err)
	}
	return content
}

func bundledAgentProfileSource(t *testing.T) func(string) ([]byte, error) {
	return func(profile string) ([]byte, error) {
		return bundledAgentProfile(t, profile), nil
	}
}

func TestShippedAgentProfilesIncludeBundledFiles(t *testing.T) {
	for _, profile := range []string{agentProfileDefault, agentProfileKVM} {
		digest := agentProfileDigest(bundledAgentProfile(t, profile))
		if got, ok := shippedAgentProfiles[digest]; !ok || got != profile {
			t.Errorf("bundled %s profile (sha256 %s) is not listed in shippedAgentProfiles", profile, digest)
		}
	}
}

func TestRefreshAgentProfileReplacesOlderShippedVersion(t *testing.T) {
	for _, profile := range []string{agentProfileDefault, agentProfileKVM} {
		t.Run(profile, func(t *testing.T) {
			workspace := t.TempDir()
			target := filepath.Join(workspace, agentProfileFile)
			// An older shipped version: the bundled file with different bytes
			// whose hash is registered for this profile.
			old := append([]byte("old shipped "+profile+"\n"), bundledAgentProfile(t, profile)...)
			shippedAgentProfiles[agentProfileDigest(old)] = profile
			t.Cleanup(func() { delete(shippedAgentProfiles, agentProfileDigest(old)) })
			if err := os.WriteFile(target, old, 0o644); err != nil {
				t.Fatal(err)
			}

			wrote, err := refreshAgentProfileIn(workspace, bundledAgentProfileSource(t))
			if err != nil || !wrote {
				t.Fatalf("refresh = %v, %v; want a write", wrote, err)
			}
			got, _ := os.ReadFile(target)
			if string(got) != string(bundledAgentProfile(t, profile)) {
				t.Fatalf("workspace AGENT.md is not the bundled %s profile", profile)
			}
		})
	}
}

func TestRefreshAgentProfileKeepsCustomisedFile(t *testing.T) {
	workspace := t.TempDir()
	target := filepath.Join(workspace, agentProfileFile)
	custom := append(bundledAgentProfile(t, agentProfileKVM), []byte("\nAlways answer in French.\n")...)
	if err := os.WriteFile(target, custom, 0o644); err != nil {
		t.Fatal(err)
	}

	wrote, err := refreshAgentProfileIn(workspace, bundledAgentProfileSource(t))
	if err != nil || wrote {
		t.Fatalf("refresh = %v, %v; want no write", wrote, err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != string(custom) {
		t.Fatal("customised AGENT.md was changed")
	}
}

func TestRefreshAgentProfileLeavesCurrentFileAlone(t *testing.T) {
	workspace := t.TempDir()
	target := filepath.Join(workspace, agentProfileFile)
	if err := os.WriteFile(target, bundledAgentProfile(t, agentProfileDefault), 0o644); err != nil {
		t.Fatal(err)
	}

	wrote, err := refreshAgentProfileIn(workspace, bundledAgentProfileSource(t))
	if err != nil || wrote {
		t.Fatalf("refresh = %v, %v; want no write", wrote, err)
	}
}

func TestRefreshAgentProfileWritesKVMProfileWhenMissing(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspace")

	wrote, err := refreshAgentProfileIn(workspace, bundledAgentProfileSource(t))
	if err != nil || !wrote {
		t.Fatalf("refresh = %v, %v; want a write", wrote, err)
	}
	got, err := os.ReadFile(filepath.Join(workspace, agentProfileFile))
	if err != nil || string(got) != string(bundledAgentProfile(t, agentProfileKVM)) {
		t.Fatalf("missing AGENT.md was not replaced by the KVM profile: %v", err)
	}
}
