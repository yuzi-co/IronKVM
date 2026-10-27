package audit

import "strings"

// Rule says how a route is recorded.
type Rule struct {
	Category string
	Action   string
	// Anonymous routes are recorded although nobody is signed in: they are
	// the ones that do the signing in.
	Anonymous bool
}

// named routes have a category and a stable action name. The key is
// "METHOD /api/route" as gin spells the route.
var named = map[string]Rule{
	"POST /api/auth/login":      {CategoryAuth, "login", true},
	"POST /api/auth/login/totp": {CategoryAuth, "login-2fa", true},
	"POST /api/auth/logout":     {CategoryAuth, "logout", false},
	// The OIDC login is a GET, recorded by its handler as login-oidc.
	"POST /api/auth/oidc/settings": {CategoryAuth, "oidc-settings", false},
	// The target is the account, the route's last parameter.
	"DELETE /api/auth/users/:username/oidc": {CategoryAuth, "oidc-unlink", false},
	"PUT /api/auth/users/:username/oidc":    {CategoryAuth, "oidc-prelink", false},
	// The user confirms a single sign-on's link with the account's password.
	"POST /api/auth/oidc/link": {CategoryAuth, "login-oidc-link", true},

	"DELETE /api/auth/sessions/:id":         {CategoryAuth, "revoke-session", false},
	"POST /api/auth/sessions/revoke-others": {CategoryAuth, "revoke-other-sessions", false},

	"POST /api/vm/gpio":          {CategoryPower, "press", false},
	"POST /api/network/wol":      {CategoryPower, "wake-on-lan", false},
	"POST /api/vm/system/reboot": {CategoryPower, "reboot", false},
	// A person's answer to a press the on-board agent asked for. The target
	// is the button; the detail names the press, the agent and its session.
	"POST /api/agent/power/confirmations/:id/confirm": {CategoryPower, "agent-press-confirm", false},
	"POST /api/agent/power/confirmations/:id/deny":    {CategoryPower, "agent-press-deny", false},

	"POST /api/storage/image/mount":       {CategoryMedia, "mount", false},
	"POST /api/storage/drives/:id/insert": {CategoryMedia, "insert", false},
	"POST /api/storage/drives/:id/eject":  {CategoryMedia, "eject", false},
	"POST /api/storage/image/delete":      {CategoryMedia, "delete-image", false},
	"POST /api/storage/image/create":      {CategoryMedia, "create-image", false},
	"POST /api/storage/files/upload":      {CategoryMedia, "file-upload", false},
	"POST /api/storage/files/delete":      {CategoryMedia, "file-delete", false},
	"POST /api/download/image":            {CategoryMedia, "download-image", false},
	"POST /api/download/image/netboot":    {CategoryMedia, "download-image", false},
	"POST /api/download/file":             {CategoryMedia, "upload-image", false},
	"POST /api/vm/device/virtual":         {CategoryMedia, "virtual-device", false},
	"POST /api/vm/device/virtual/apply":   {CategoryMedia, "virtual-device", false},
	// An image served from a URL: the target is the drive and the URL's
	// host, never the URL, which may carry credentials.
	"POST /api/storage/url/mount": {CategoryMedia, "mount-url", false},
	"POST /api/storage/url/eject": {CategoryMedia, "eject-url", false},

	"POST /api/hid/autotap":       {CategoryMacro, "autotap", false},
	"POST /api/hid/macro/play":    {CategoryMacro, "play", false},
	"POST /api/hid/sequence/stop": {CategoryMacro, "stop", false},
	"POST /api/hid/macro":         {CategoryMacro, "save", false},
	"DELETE /api/hid/macro":       {CategoryMacro, "delete", false},

	"POST /api/hid/paste": {CategoryControl, "paste", false},
	"POST /api/hid/key":   {CategoryControl, "key", false},

	"POST /api/application/update":         {CategorySystem, "update", false},
	"POST /api/application/update/offline": {CategorySystem, "update-offline", false},
	"POST /api/ui/update":                  {CategorySystem, "ui-update", false},
	"POST /api/ui/upload":                  {CategorySystem, "ui-upload", false},
	"POST /api/ui/rollback":                {CategorySystem, "ui-rollback", false},
	// An install is recorded when it starts, with the slot and the image's
	// host and path, and its end again as slot-install-result.
	"POST /api/slot/install": {CategorySystem, "slot-install", false},
	"POST /api/slot/try":     {CategorySystem, "slot-try", false},
	"POST /api/slot/confirm": {CategorySystem, "slot-confirm", false},
	"POST /api/slot/revert":  {CategorySystem, "slot-revert", false},

	"POST /api/settings/import": {CategorySettings, "settings-import", false},

	// A clip of the last minute, asked for by hand or by a script, and its
	// deletion. The target is the clip. A clip an event saved is recorded by
	// the dashcam itself.
	"POST /api/dashcam/save":          {CategoryMedia, "dashcam-save", false},
	"DELETE /api/dashcam/clips/:name": {CategoryMedia, "dashcam-delete", false},
}

// skipped routes change nothing worth a line: reads that use POST, progress
// the UI sends while it works, and the loopback routes of the board's own
// scripts.
var skipped = map[string]bool{
	"POST /api/hid/paste/check":          true,
	"POST /api/hid/paste/cancel":         true,
	"POST /api/stream/mjpeg/detect":      true,
	"POST /api/stream/mjpeg/detect/stop": true,
	"POST /api/vm/input-region":          true,
	"POST /api/vm/screen":                true,
	"POST /api/vm/tls/cert/check":        true,
	"POST /api/network/syslog/test":      true,
	"POST /api/network/wifi/verify":      true,
	"POST /api/download/image/cancel":    true,
	"POST /api/internal/usb/recover":     true,
	// The preview writes nothing; the import that follows it is recorded.
	"POST /api/settings/preview": true,
}

// selfAudited are the routes whose handlers record their own events, with
// more than the route can tell: the MCP endpoint and the on-board agent's
// loopback tools.
var selfAudited = map[string]bool{
	"/api/mcp":                 true,
	"/api/agent/mcp":           true,
	"/api/agent/actions":       true,
	"/api/agent/load-image":    true,
	"/api/picoclaw/mcp":        true,
	"/api/picoclaw/actions":    true,
	"/api/picoclaw/load-image": true,
	// WHEP records each session's start and end, with who opened it.
	"/api/whep":     true,
	"/api/whep/:id": true,
}

// Classify returns the rule for a route, if it is audited. Every other
// change under /api, whatever its handler, is a settings change.
func Classify(method, route string) (Rule, bool) {
	if !strings.HasPrefix(route, "/api/") {
		return Rule{}, false
	}
	key := method + " " + route
	if skipped[key] {
		return Rule{}, false
	}
	if r, ok := named[key]; ok {
		return r, true
	}
	if selfAudited[route] {
		return Rule{}, false
	}
	category := CategorySettings
	if strings.HasPrefix(route, "/api/auth/") {
		category = CategoryAuth
	}
	return Rule{Category: category, Action: strings.ToLower(method) + " " + strings.TrimPrefix(route, "/api")}, true
}
