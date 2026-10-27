package router

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// The Home Assistant nanokvm integration (github.com/Wouter0100/homeassistant-
// nanokvm, through the nanokvm client on PyPI) is written against Sipeed's API.
// These tests pin the part of ours it uses on a non-Pro board, so a change that
// would break it is noticed here first (ironkvm-dist#124). The integration
// skips the Pro-only routes once GET /api/vm/hardware answers a non-Pro board.

// haRoute is a route the integration calls. anyUser marks one it calls for an
// account that is not an administrator, which must therefore not be behind
// RequireRole.
type haRoute struct {
	method, path string
	anyUser      bool
}

var homeAssistantRoutes = []haRoute{
	{"POST", "/api/auth/login", true},
	{"POST", "/api/auth/logout", true},
	{"GET", "/api/auth/account", true},
	{"GET", "/api/vm/info", true},
	{"GET", "/api/vm/hardware", true},
	{"GET", "/api/vm/hostname", true},
	{"GET", "/api/vm/gpio", true},
	{"POST", "/api/vm/gpio", true},
	{"GET", "/api/hid/mode", true},
	{"GET", "/api/hid/leds", true},
	{"POST", "/api/hid/paste", true},
	{"GET", "/api/vm/hdmi", true},
	{"POST", "/api/network/wol", true},
	{"GET", "/api/stream/mjpeg", true},
	{"GET", "/api/stream/h264", true},

	{"GET", "/api/vm/ssh", false},
	{"POST", "/api/vm/ssh/enable", false},
	{"POST", "/api/vm/ssh/disable", false},
	{"GET", "/api/vm/mdns", false},
	{"POST", "/api/vm/mdns/enable", false},
	{"POST", "/api/vm/mdns/disable", false},
	{"GET", "/api/vm/oled", false},
	{"POST", "/api/vm/oled", false},
	{"GET", "/api/network/wifi", false},
	{"GET", "/api/vm/mouse-jiggler", false},
	{"POST", "/api/vm/mouse-jiggler/", false},
	{"GET", "/api/extensions/tailscale/status", false},
	{"GET", "/api/vm/device/virtual", false},
	{"POST", "/api/vm/device/virtual", false},
	{"GET", "/api/vm/swap", false},
	{"POST", "/api/vm/swap", false},
	{"POST", "/api/vm/hdmi/reset", false},
	{"POST", "/api/vm/hdmi/enable", false},
	{"POST", "/api/vm/hdmi/disable", false},
	{"POST", "/api/vm/hdmi/timeout", false},
	{"POST", "/api/hid/mode", false},
	{"POST", "/api/hid/reset", false},
	{"POST", "/api/vm/system/reboot", false},
	{"GET", "/api/application/version", false},
	{"POST", "/api/application/update", false},
	{"GET", "/api/storage/image", false},
	{"GET", "/api/storage/image/mounted", false},
	{"GET", "/api/storage/cdrom", false},
	{"GET", "/api/download/image/enabled", false},
	{"GET", "/api/download/image/status", false},
}

func TestHomeAssistantRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chains := make(map[string][]string)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if c.GetHeader(probeHeader) == "" {
			return
		}
		chains[c.Request.Method+" "+c.FullPath()] = c.HandlerNames()
		c.AbortWithStatus(http.StatusTeapot)
	})
	server(r)

	registered := make(map[string]bool)
	for _, route := range r.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, route := range homeAssistantRoutes {
		key := route.method + " " + route.path
		if !registered[key] {
			t.Errorf("%s: Home Assistant calls it, and it is not registered", key)
			continue
		}
		request := httptest.NewRequest(route.method, route.path, nil)
		request.Header.Set(probeHeader, "1")
		r.ServeHTTP(httptest.NewRecorder(), request)
		chain := chains[key]
		if route.anyUser && chainHas(chain, "RequireRole") {
			t.Errorf("%s: Home Assistant calls it for any account, and it is for administrators only", key)
		}
	}
}

// TestHomeAssistantResponseFields pins the fields the client reads, with the
// JSON kind it validates them as. A field may be added; one renamed, removed or
// changed in kind fails the device's whole poll in Home Assistant.
func TestHomeAssistantResponseFields(t *testing.T) {
	const (
		boolean = "bool"
		number  = "number"
		text    = "string"
		list    = "list"
		object  = "object"
	)
	cases := []struct {
		value  any
		fields map[string]string
	}{
		{proto.Response{}, map[string]string{"code": number, "msg": text}},
		{proto.GetAccountRsp{}, map[string]string{"username": text, "role": text}},
		{proto.GetInfoRsp{}, map[string]string{"ips": list, "mdns": text, "image": text, "application": text, "deviceKey": text}},
		{proto.IP{}, map[string]string{"name": text, "addr": text, "version": text, "type": text}},
		{proto.GetHardwareRsp{}, map[string]string{"version": text}},
		{proto.GetHostnameRsp{}, map[string]string{"hostname": text}},
		{proto.GetGpioRsp{}, map[string]string{"pwr": boolean, "hdd": boolean}},
		{proto.GetHidModeRsp{}, map[string]string{"mode": text}},
		{proto.GetKeyboardLedStatusRsp{}, map[string]string{"numLock": boolean, "capsLock": boolean, "scrollLock": boolean, "known": boolean, "updatedAt": text}},
		{proto.GetGetHdmiStateRsp{}, map[string]string{"enabled": boolean, "signal": boolean, "idleTimeout": number}},
		{proto.GetSSHStateRsp{}, map[string]string{"enabled": boolean}},
		{proto.GetMdnsStateRsp{}, map[string]string{"enabled": boolean}},
		{proto.GetOLEDRsp{}, map[string]string{"exist": boolean, "sleep": number}},
		{proto.GetWifiRsp{}, map[string]string{"supported": boolean, "apMode": boolean, "connected": boolean, "ssid": text}},
		{proto.GetMouseJigglerRsp{}, map[string]string{"enabled": boolean, "mode": text}},
		{proto.VpnStatus{}, map[string]string{"state": text, "name": text, "ip": text, "account": text}},
		{proto.GetVirtualDeviceRsp{}, map[string]string{"network": boolean, "disk": boolean}},
		{proto.GetSwapRsp{}, map[string]string{"size": number}},
		{proto.GetVersionRsp{}, map[string]string{"current": text, "latest": text}},
		{proto.GetImagesRsp{}, map[string]string{"files": list}},
		{proto.GetMountedImageRsp{}, map[string]string{"file": text}},
		{proto.GetCdRomRsp{}, map[string]string{"cdrom": number}},
		{proto.ImageEnabledRsp{}, map[string]string{"enabled": boolean}},
		{proto.StatusImageRsp{}, map[string]string{"status": text, "file": text, "percentage": text}},
	}

	for _, c := range cases {
		typ := reflect.TypeOf(c.value)
		got := jsonKinds(typ)
		names := make([]string, 0, len(c.fields))
		for name := range c.fields {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if kind, ok := got[name]; !ok {
				t.Errorf("%s has no %q", typ.Name(), name)
			} else if kind != c.fields[name] {
				t.Errorf("%s.%s is a %s, Home Assistant reads a %s", typ.Name(), name, kind, c.fields[name])
			}
		}
	}
}

// TestHomeAssistantHardwareVersions keeps the hardware names inside the set
// the client knows. Anything else fails GET /api/vm/hardware, which every
// poll starts with.
func TestHomeAssistantHardwareVersions(t *testing.T) {
	known := map[string]bool{"Alpha": true, "Beta": true, "PCIE": true, "Pro": true, "Unknown": true}
	for _, version := range []config.HWVersion{config.HWVersionAlpha, config.HWVersionBeta, config.HWVersionPcie, config.HWVersion(99)} {
		if name := version.String(); !known[name] {
			t.Errorf("hardware version %d is %q, which the client does not know", version, name)
		}
	}
}

// TestHomeAssistantStatusCodes pins the two codes the client tells apart: 401
// clears its session and logs in again, 403 marks the route as not allowed for
// the account and leaves the rest of the poll alone.
func TestHomeAssistantStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conf := config.GetInstance()
	original := conf.Authentication
	conf.Authentication = "enable"
	t.Cleanup(func() { conf.Authentication = original })

	r := gin.New()
	r.GET("/api/vm/info", middleware.CheckToken(), func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/vm/info", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("no session: %d, want 401", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/vm/info", nil)
	request.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: "not-a-token"})
	r.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("a token that is not a session: %d, want 401", recorder.Code)
	}

	if middleware.CookieName != "nano-kvm-token" {
		t.Errorf("the session cookie is %q; the client sends nano-kvm-token", middleware.CookieName)
	}
}

// jsonKinds maps each JSON field of a struct, embedded ones included, to the
// kind of JSON value it encodes as.
func jsonKinds(typ reflect.Type) map[string]string {
	kinds := make(map[string]string)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if field.Anonymous && tag == "" {
			for embedded, kind := range jsonKinds(field.Type) {
				kinds[embedded] = kind
			}
			continue
		}
		if name == "-" || !field.IsExported() {
			continue
		}
		if name == "" {
			name = field.Name
		}
		kinds[name] = jsonKind(field.Type)
	}
	return kinds
}

func jsonKind(typ reflect.Type) string {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		return "list"
	case reflect.Struct, reflect.Map:
		return "object"
	default:
		return typ.Kind().String()
	}
}
