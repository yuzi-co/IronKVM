// Package apicontract is what this server promises a web UI: an API version
// and the features it has. GET /api/version reports both, and a UI package
// states the API versions it supports, so a UI can be installed apart from the
// server only when the two agree.
package apicontract

import "sort"

// APIVersion is bumped on a breaking change to the API the web UI uses: a
// route removed or renamed, or a request or response changed in a way an
// existing UI cannot read. Adding a route or a field is not breaking; declare a
// feature instead.
//
// A server that predates GET /api/version speaks version 1, so the web UI
// treats a missing endpoint as version 1 with no declared features.
const APIVersion = 1

// features is the registry, one line per feature, kept in alphabetical order.
// A UI asks for a feature by this name (web/src/lib/server-features.ts) before
// it calls the routes the comment names. Add a line when a branch adds such a
// route; two branches that both add one conflict only on adjacent lines, and
// the resolution is to keep both.
var features = []string{
	"agent-tiers",         // /api/agent/tiers: what MCP and the on-board agent may do
	"audit-log",           // /api/audit: the audit log, its export and settings
	"direct-clock",        // clock request (byte 5) on the direct H.264 socket
	"hid-autotap",         // POST /api/hid/autotap, /api/hid/sequence: key auto-tap
	"hid-macros",          // /api/hid/macro(s), /api/hid/sequence: recorded macros
	"host-log",            // /api/hostlog: UDP syslog and netconsole receiver, live tail
	"info-latency",        // GET /api/stream/info boardLatency: the board's share of the latency
	"ip-allowlist",        // /api/network/allowlist: allowed client networks
	"ipv6-settings",       // /api/network/ipv6: SLAAC, DHCPv6, static or off
	"notifications",       // /api/notify: webhook, ntfy and Telegram alerts with a snapshot
	"nut-client",          // /api/nut: NUT client, UPS state and settings
	"role-viewer",          // role "viewer" in /api/auth/users and GET /api/auth/account: video only
	"share-links",          // /api/auth/share, /api/auth/share-links: expiring view-only links
	"status-bar",          // GET /api/stream/info
	"ui-update",           // /api/ui: separately installed UI packages
	"webrtc-clock",        // "clock" event on the WebRTC signalling socket
	"webrtc-port-forward", // /api/network/webrtc: fixed UDP port and public address
	"wol-devices",         // /api/network/wol/devices, /device, /check, /scan: managed Wake-on-LAN list
}

// Features returns the registry, sorted, as a copy the caller may keep.
func Features() []string {
	out := append([]string(nil), features...)
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
