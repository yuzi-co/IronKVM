// Package apicontract is what this server promises a web UI: an API version
// and the features it has. GET /api/version reports both, and a UI package
// states the API versions it supports, so a UI can be installed apart from the
// server only when the two agree.
package apicontract

import (
	"sort"
	"sync"
)

// APIVersion is bumped on a breaking change to the API the web UI uses: a
// route removed or renamed, or a request or response changed in a way an
// existing UI cannot read. Adding a route or a field is not breaking; declare a
// feature instead.
//
// A server that predates GET /api/version speaks version 1, so the web UI
// treats a missing endpoint as version 1 with no declared features.
//
// Version 2: GET /api/vm/device/virtual and its apply answer moved each
// function's state under "functions" and gave "network", "disk" and "audio"
// back to Sipeed's booleans, which clients of Sipeed's API read
// (ironkvm-dist#124).
const APIVersion = 2

// features is the registry, one line per feature, kept in alphabetical order.
// A UI asks for a feature by this name (web/src/lib/server-features.ts) before
// it calls the routes the comment names. Add a line when a branch adds such a
// route; two branches that both add one conflict only on adjacent lines, and
// the resolution is to keep both.
var features = []string{
	"agent-tiers",         // /api/agent/tiers: what MCP and the on-board agent may do
	"api-key-scope",       // scope "view" on POST /api/auth/api-keys: a key that can only watch
	"audit-log",           // /api/audit: the audit log, its export and settings
	"diagnostics-bundle",  // GET /api/diagnostics/bundle: support bundle without secrets
	"direct-clock",        // clock request (byte 5) on the direct H.264 socket
	"edid-view",           // /api/edid: EDID presets and decoder, read only
	"hid-autotap",         // POST /api/hid/autotap, /api/hid/sequence: key auto-tap
	"hid-macros",          // /api/hid/macro(s), /api/hid/sequence: recorded macros
	"host-log",            // /api/hostlog: UDP syslog and netconsole receiver, live tail
	"info-latency",        // GET /api/stream/info boardLatency: the board's share of the latency
	"ip-allowlist",        // /api/network/allowlist: allowed client networks
	"ipv6-settings",       // /api/network/ipv6: SLAAC, DHCPv6, static or off
	"login-sessions",      // /api/auth/sessions: login sessions, listed and revoked one by one
	"notifications",       // /api/notify: webhook, ntfy and Telegram alerts with a snapshot
	"nut-client",          // /api/nut: NUT client, UPS state and settings
	"oidc-login",          // /api/auth/oidc: login through an OpenID Connect provider, its settings
	"role-viewer",         // role "viewer" in /api/auth/users and GET /api/auth/account: video only
	"settings-backup",     // /api/settings/export, /preview, /import: settings file without secrets
	"share-links",         // /api/auth/share, /api/auth/share-links: expiring view-only links
	"status-bar",          // GET /api/stream/info
	"stun-server",         // /api/network/webrtc/stun: the STUN server, or off
	"ui-update",           // /api/ui: separately installed UI packages
	"video-clock",         // /api/vm/video-clock: video clock profile (mainline kernel only, see SetProbe)
	"webrtc-clock",        // "clock" event on the WebRTC signalling socket
	"webrtc-port-forward", // /api/network/webrtc: fixed UDP port and public address
	"whep",                // /api/whep, /api/stream/whep: WHEP egress for players, and its switch
	"wol-devices",         // /api/network/wol/devices, /device, /check, /scan: managed Wake-on-LAN list
}

// probes decide, for a feature that depends on the board rather than on this
// server's code, whether it is there. A feature without a probe always is.
var (
	probesMu sync.Mutex
	probes   = map[string]func() bool{}
)

// SetProbe makes a registered feature depend on probe: Features lists it only
// while probe answers true. main sets the probes before the server answers.
func SetProbe(name string, probe func() bool) {
	probesMu.Lock()
	defer probesMu.Unlock()
	probes[name] = probe
}

// Features returns the registry, sorted, as a copy the caller may keep. A
// feature whose probe answers false is left out.
func Features() []string {
	probesMu.Lock()
	defer probesMu.Unlock()

	out := make([]string, 0, len(features))
	for _, f := range features {
		if probe, ok := probes[f]; ok && !probe() {
			continue
		}
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// Has reports whether name is registered.
func Has(name string) bool {
	for _, f := range features {
		if f == name {
			return true
		}
	}
	return false
}
