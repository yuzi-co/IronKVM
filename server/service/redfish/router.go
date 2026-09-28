// Package redfish serves a fixed subset of the DMTF Redfish API: one system,
// one chassis and one manager, power control through the front-panel buttons,
// and the two virtual drives as virtual media.
package redfish

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/proto"
)

// The resource paths. The service has exactly one of each, with the ID 1.
const (
	rootPath           = "/redfish/v1/"
	systemsPath        = "/redfish/v1/Systems"
	systemPath         = "/redfish/v1/Systems/1"
	resetPath          = systemPath + "/Actions/ComputerSystem.Reset"
	chassisCollection  = "/redfish/v1/Chassis"
	chassisPath        = "/redfish/v1/Chassis/1"
	managersPath       = "/redfish/v1/Managers"
	managerPath        = "/redfish/v1/Managers/1"
	nicsPath           = managerPath + "/EthernetInterfaces"
	mediaPath          = managerPath + "/VirtualMedia"
	sessionServicePath = "/redfish/v1/SessionService"
	sessionsPath       = sessionServicePath + "/Sessions"
	metadataPath       = "/redfish/v1/$metadata"
	odataPath          = "/redfish/v1/odata"
)

// Deps is everything the service reaches outside this package. The router
// wires the real ones; tests pass fakes.
type Deps struct {
	// PressButton holds ButtonPower or ButtonReset for d.
	PressButton func(kind string, d time.Duration) error
	// PowerLED reports whether the host's power LED is lit.
	PowerLED func() (bool, error)
	// PowerLEDConnected reports whether the owner has said the LED header is
	// wired. When it is not, the LED is never read: the state is unknown.
	// Nil means wired.
	PowerLEDConnected func() bool

	ListDrives  func() ([]proto.DriveInfo, error)
	InsertDrive func(id string, file string, ro bool) error
	EjectDrive  func(id string) error
	// ImageDir is the directory InsertMedia resolves a bare file name in.
	ImageDir string

	Accounts Accounts
	// APIKeyUser returns the account an API key was issued to.
	APIKeyUser func(secret string) (username string, ok bool)
	Limiter    Limiter
	// AuthDisabled reports whether the board runs with authentication off.
	AuthDisabled func() bool
	// FailureDelay is how long a failed password check waits before it
	// answers.
	FailureDelay time.Duration

	FirmwareVersion func() string
	NICs            func() []NIC

	// UUID is the service's UUID, kept across restarts.
	UUID string
	// Now is the clock the session timeout runs on. Nil means time.Now.
	Now func() time.Time

	// Enabled reports the redfish.enabled setting. It is read on every
	// request, so the switch needs no restart. Nil means always on.
	Enabled func() bool
	// SetEnabled saves the setting and applies it: Enabled reports the new
	// value once it returns nil.
	SetEnabled func(on bool) error
	// HTTPS reports whether the board serves over HTTPS, which most Redfish
	// clients need.
	HTTPS func() bool
}

// The buttons PressButton accepts. They are the names service/vm uses.
const (
	ButtonPower = "power"
	ButtonReset = "reset"
)

// Accounts is the part of authn.Store the service uses.
type Accounts interface {
	Authenticate(username, password string) (*authn.User, bool, error)
	ValidateToken(username string, tokenVersion uint64) (*authn.User, error)
	Get(username string) (*authn.User, error)
}

// Limiter is the login brute-force limit, keyed by client address and
// account.
type Limiter interface {
	Locked(ip, username string) bool
	Failed(ip, username string)
	Succeeded(ip, username string)
}

// NIC is one of the board's network interfaces.
type NIC struct {
	ID   string
	MAC  string
	IPv4 string
}

// Service answers the /redfish routes.
type Service struct {
	deps        Deps
	sessions    *sessionStore
	credentials *credentialCache

	// resetMu serializes Reset actions from the LED read to the settle time
	// after the press.
	resetMu sync.Mutex
	// switchMu serializes changes to the enabled setting.
	switchMu sync.Mutex
}

func New(deps Deps) *Service {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.PowerLEDConnected == nil {
		deps.PowerLEDConnected = func() bool { return true }
	}
	if deps.Enabled == nil {
		deps.Enabled = func() bool { return true }
	}
	if deps.HTTPS == nil {
		deps.HTTPS = func() bool { return false }
	}

	return &Service{deps: deps, sessions: newSessionStore(deps.Now), credentials: newCredentialCache(deps.Now)}
}

// Register adds the /redfish routes to r. It also sets r's NoRoute handler,
// which answers unknown /redfish paths with a Redfish error and leaves every
// other path to gin's default 404. While the service is off, every /redfish
// path answers a plain 404.
func (s *Service) Register(r *gin.Engine) {
	r.NoRoute(s.notFoundUnderRedfish)

	g := r.Group("", s.whenEnabled, s.authenticate)

	route(g, "/redfish", map[string]gin.HandlerFunc{http.MethodGet: versions})
	route(g, "/redfish/v1", map[string]gin.HandlerFunc{http.MethodGet: s.serviceRoot})
	route(g, rootPath, map[string]gin.HandlerFunc{http.MethodGet: s.serviceRoot})
	route(g, odataPath, map[string]gin.HandlerFunc{http.MethodGet: serviceDocument})
	route(g, metadataPath, map[string]gin.HandlerFunc{http.MethodGet: metadata})

	route(g, sessionServicePath, map[string]gin.HandlerFunc{http.MethodGet: sessionService})
	route(g, sessionsPath, map[string]gin.HandlerFunc{
		http.MethodGet:  s.listSessions,
		http.MethodPost: s.createSession,
	})
	route(g, sessionsPath+"/:id", map[string]gin.HandlerFunc{
		http.MethodGet:    s.getSession,
		http.MethodDelete: s.deleteSession,
	})

	route(g, systemsPath, map[string]gin.HandlerFunc{http.MethodGet: systems})
	route(g, systemPath, map[string]gin.HandlerFunc{http.MethodGet: s.system})
	route(g, resetPath, map[string]gin.HandlerFunc{http.MethodPost: adminOnly(s.reset)})
	route(g, chassisCollection, map[string]gin.HandlerFunc{http.MethodGet: chassisList})
	route(g, chassisPath, map[string]gin.HandlerFunc{http.MethodGet: s.chassis})

	route(g, managersPath, map[string]gin.HandlerFunc{http.MethodGet: managers})
	route(g, managerPath, map[string]gin.HandlerFunc{http.MethodGet: s.manager})
	route(g, nicsPath, map[string]gin.HandlerFunc{http.MethodGet: s.nics})
	route(g, nicsPath+"/:id", map[string]gin.HandlerFunc{http.MethodGet: s.nic})

	route(g, mediaPath, map[string]gin.HandlerFunc{http.MethodGet: s.virtualMedia})
	route(g, mediaPath+"/:id", map[string]gin.HandlerFunc{http.MethodGet: s.medium})
	route(g, mediaPath+"/:id/Actions/"+insertAction, map[string]gin.HandlerFunc{http.MethodPost: adminOnly(s.insertMedia)})
	route(g, mediaPath+"/:id/Actions/"+ejectAction, map[string]gin.HandlerFunc{http.MethodPost: adminOnly(s.ejectMedia)})
}

// publicRoutes answer without credentials, as the Redfish specification
// requires: the version object, the service root, the OData documents, and
// the login itself. Keys are "METHOD route".
var publicRoutes = map[string]bool{
	"GET /redfish":         true,
	"HEAD /redfish":        true,
	"GET /redfish/v1":      true,
	"HEAD /redfish/v1":     true,
	"GET " + rootPath:      true,
	"HEAD " + rootPath:     true,
	"GET " + odataPath:     true,
	"HEAD " + odataPath:    true,
	"GET " + metadataPath:  true,
	"HEAD " + metadataPath: true,
	"POST " + sessionsPath: true,
}

var routeMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// route registers one resource's handlers. HEAD answers as GET does, and
// every method the resource lacks answers 405 with an Allow header.
func route(g *gin.RouterGroup, path string, handlers map[string]gin.HandlerFunc) {
	if get, ok := handlers[http.MethodGet]; ok {
		if _, ok := handlers[http.MethodHead]; !ok {
			handlers[http.MethodHead] = get
		}
	}

	var allowed []string
	for _, method := range routeMethods {
		if _, ok := handlers[method]; ok {
			allowed = append(allowed, method)
		}
	}
	allowed = append(allowed, http.MethodOptions)
	allow := strings.Join(allowed, ", ")

	g.Handle(http.MethodOptions, path, func(c *gin.Context) {
		c.Header("Allow", allow)
		c.Status(http.StatusNoContent)
	})

	for _, method := range routeMethods {
		handler, ok := handlers[method]
		if !ok {
			handler = methodNotAllowed(allow)
		}
		g.Handle(method, path, handler)
	}
}

func methodNotAllowed(allow string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Allow", allow)
		writeError(c, http.StatusMethodNotAllowed, "GeneralError", "the method is not allowed on this resource")
	}
}

// notFoundUnderRedfish is the engine's NoRoute handler. Only paths under
// /redfish are answered here; anything else falls through to gin's own 404.
// An unknown path needs credentials as a known one does, so a probe without
// them cannot tell which resources exist.
func (s *Service) notFoundUnderRedfish(c *gin.Context) {
	p := c.Request.URL.Path
	if p != "/redfish" && !strings.HasPrefix(p, "/redfish/") {
		return
	}
	s.whenEnabled(c)
	if c.IsAborted() {
		return
	}
	s.authenticate(c)
	if c.IsAborted() {
		return
	}
	notFound(c)
}

func (s *Service) enabled() bool {
	return s.deps.Enabled()
}

// whenEnabled answers 404 while the service is off, before any credential is
// looked at. The body is gin's own, so an off service looks like none.
func (s *Service) whenEnabled(c *gin.Context) {
	if s.enabled() {
		return
	}
	disabledNotFound(c)
}

func disabledNotFound(c *gin.Context) {
	c.String(http.StatusNotFound, "404 page not found")
	c.Abort()
}

func notFound(c *gin.Context) {
	writeError(c, http.StatusNotFound, "ResourceMissingAtURI", "", c.Request.URL.Path)
}

type object = map[string]any

func link(path string) object {
	return object{"@odata.id": path}
}

func links(paths ...string) []object {
	out := make([]object, 0, len(paths))
	for _, path := range paths {
		out = append(out, link(path))
	}
	return out
}

// newResource starts a resource body with its OData annotations. odataType
// is the full type, such as "ComputerSystem.v1_13_0.ComputerSystem".
func newResource(path string, odataType string) object {
	parts := strings.Split(odataType, ".")
	context := parts[0] + "." + parts[len(parts)-1]

	return object{
		"@odata.id":      path,
		"@odata.type":    "#" + odataType,
		"@odata.context": "/redfish/v1/$metadata#" + context,
	}
}

// newCollection builds a resource collection. odataType names the collection
// type once, such as "ComputerSystemCollection".
func newCollection(path string, odataType string, name string, members []string) object {
	body := newResource(path, odataType+"."+odataType)
	body["Name"] = name
	body["Members"] = links(members...)
	body["Members@odata.count"] = len(members)
	return body
}

func writeJSON(c *gin.Context, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		log.Errorf("redfish: encode %s: %s", c.Request.URL.Path, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("OData-Version", "4.0")
	c.Data(status, "application/json; charset=utf-8", data)
}

// writeTagged answers 200 with body and an ETag over its bytes.
func writeTagged(c *gin.Context, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		log.Errorf("redfish: encode %s: %s", c.Request.URL.Path, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	sum := sha256.Sum256(data)
	c.Header("ETag", `W/"`+hex.EncodeToString(sum[:8])+`"`)
	c.Header("OData-Version", "4.0")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

func versions(c *gin.Context) {
	writeJSON(c, http.StatusOK, object{"v1": rootPath})
}
