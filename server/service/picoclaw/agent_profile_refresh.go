package picoclaw

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// shippedAgentProfiles maps the SHA-256 of every AGENT.md and AGENT_KVM.md
// this firmware has ever shipped to the profile it belongs to. On runtime
// start a workspace AGENT.md whose content is one of these is replaced by the
// bundled file for the same profile, so new agent instructions reach existing
// installs after an update.
//
// A workspace file that matches none of them was edited by someone (by hand,
// or by the agent through its workspace file tools) and is left alone. Hashes
// of exact shipped bytes are used rather than a marker in the file, because a
// marker survives a user's edits and would let the refresh overwrite them.
//
// When kvmapp/picoclaw/AGENT.md or AGENT_KVM.md changes, add the hash of the
// new content here (TestShippedAgentProfilesIncludeBundledFiles fails until
// it is listed). Never remove an entry: an install may still carry it.
var shippedAgentProfiles = map[string]string{
	// AGENT.md
	"ca31d9cac638e2f3c2a2eb6e80553bdaec940914baba491cd073ea1866b80b76": agentProfileDefault,
	"9cb21b64069bb53b2ed7834c465dcafcfe85697399b67bdb7e90cebf9a1f97a7": agentProfileDefault,
	// AGENT_KVM.md
	"cdf20efa3f7e4f60ad685731795e0045f72cf54f455202a141ac6af0e16cc230": agentProfileKVM,
	"64e1acd35a8ea960097848afb96f2b8955259e8bd8537469439836644c36270d": agentProfileKVM,
	"de3e1f8a1840dd2692b622ae8af3cb6fbdca08d70d4ca5a70e513dc5231d8c5f": agentProfileKVM,
	"d67bc0dc27f84df68011bb1f47c6041a9e2a4ace9826ca8f59ee7f61eaee07f4": agentProfileKVM,
}

func agentProfileDigest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// refreshPicoclawAgentProfile brings the workspace AGENT.md up to date with
// the bundled profile files. See refreshAgentProfileIn for the rule.
func refreshPicoclawAgentProfile() (bool, error) {
	workspacePath, err := resolvePicoclawWorkspacePath()
	if err != nil {
		return false, err
	}

	return refreshAgentProfileIn(workspacePath, func(profile string) ([]byte, error) {
		sourcePath, err := resolveAgentProfileSourcePath(profile)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(sourcePath)
	})
}

// refreshAgentProfileIn writes the bundled profile into workspace/AGENT.md
// when the workspace has no AGENT.md (the KVM profile, which is what the UI
// reports for a missing file) or when the file holds an older shipped
// version of a profile (the current version of that same profile, so the
// user's profile choice is kept). Anything else is a customised file and
// stays as it is. It reports whether it wrote the file.
func refreshAgentProfileIn(workspacePath string, readSource func(profile string) ([]byte, error)) (bool, error) {
	targetPath := filepath.Join(workspacePath, agentProfileFile)

	profile := agentProfileKVM
	current, err := os.ReadFile(targetPath)
	switch {
	case err == nil:
		shippedProfile, shipped := shippedAgentProfiles[agentProfileDigest(current)]
		if !shipped {
			return false, nil
		}
		profile = shippedProfile
	case errors.Is(err, os.ErrNotExist):
	default:
		return false, fmt.Errorf("failed to read picoclaw AGENT.md: %w", err)
	}

	bundled, err := readSource(profile)
	if err != nil {
		return false, fmt.Errorf("failed to read agent profile source: %w", err)
	}
	if current != nil && agentProfileDigest(current) == agentProfileDigest(bundled) {
		return false, nil
	}

	if err := os.MkdirAll(workspacePath, 0o755); err != nil {
		return false, fmt.Errorf("failed to create picoclaw workspace: %w", err)
	}
	if err := writeFileAtomic(targetPath, bundled, 0o644); err != nil {
		return false, fmt.Errorf("failed to write picoclaw AGENT.md: %w", err)
	}

	return true, nil
}
