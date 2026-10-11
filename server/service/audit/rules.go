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

	"POST /api/vm/gpio":          {CategoryPower, "press", false},
	"POST /api/network/wol":      {CategoryPower, "wake-on-lan", false},
	"POST /api/vm/system/reboot": {CategoryPower, "reboot", false},

	"POST /api/storage/image/mount":       {CategoryMedia, "mount", false},
	"POST /api/storage/drives/:id/insert": {CategoryMedia, "insert", false},
	"POST /api/storage/drives/:id/eject":  {CategoryMedia, "eject", false},
	"POST /api/storage/image/delete":      {CategoryMedia, "delete-image", false},
	"POST /api/download/image":            {CategoryMedia, "download-image", false},
	"POST /api/download/image/netboot":    {CategoryMedia, "download-image", false},
	"POST /api/download/file":             {CategoryMedia, "upload-image", false},
	"POST /api/vm/device/virtual":         {CategoryMedia, "virtual-device", false},
	"POST /api/vm/device/virtual/apply":   {CategoryMedia, "virtual-device", false},

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
}

// selfAudited are the routes whose handlers record their own events, with
// more than the route can tell: the MCP endpoint and the on-board agent's
// loopback tools.
var selfAudited = map[string]bool{
	"/api/mcp":                 true,
	"/api/picoclaw/mcp":        true,
	"/api/picoclaw/actions":    true,
	"/api/picoclaw/load-image": true,
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
