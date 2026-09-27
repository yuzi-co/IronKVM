package redfish

import (
	"net/http"
	"strings"
	"testing"
)

func TestVersionsObjectAnswersWithoutCredentials(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/redfish", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	if got := decode(t, w)["v1"]; got != "/redfish/v1/" {
		t.Fatalf("v1 is %v", got)
	}
	if w.Header().Get("OData-Version") != "4.0" {
		t.Fatal("no OData-Version header")
	}
}

func TestServiceRootAnswersWithoutCredentialsAtBothSpellings(t *testing.T) {
	h := newHarness(t)

	for _, path := range []string{"/redfish/v1/", "/redfish/v1"} {
		w := h.do(http.MethodGet, path, "")
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status %d: %s", path, w.Code, w.Body.String())
		}
		body := decode(t, w)
		if body["UUID"] != testUUID {
			t.Fatalf("%s: UUID is %v", path, body["UUID"])
		}
		if body["@odata.type"] != "#ServiceRoot.v1_5_0.ServiceRoot" {
			t.Fatalf("%s: @odata.type is %v", path, body["@odata.type"])
		}
		sessions := body["Links"].(map[string]any)["Sessions"].(map[string]any)["@odata.id"]
		if sessions != "/redfish/v1/SessionService/Sessions" {
			t.Fatalf("%s: Links.Sessions is %v", path, sessions)
		}
	}
}

func TestHeadAnswersLikeGet(t *testing.T) {
	h := newHarness(t)

	if w := h.do(http.MethodHead, "/redfish/v1/", ""); w.Code != http.StatusOK {
		t.Fatalf("HEAD status %d", w.Code)
	}
}

func TestODataDocumentsAnswerWithoutCredentials(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/redfish/v1/$metadata", "")
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/xml") {
		t.Fatalf("$metadata: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), `Namespace="ServiceRoot.v1_5_0"`) {
		t.Fatal("$metadata does not reference the ServiceRoot schema")
	}

	w = h.do(http.MethodGet, "/redfish/v1/odata", "")
	if w.Code != http.StatusOK {
		t.Fatalf("odata: %d", w.Code)
	}
	if values, ok := decode(t, w)["value"].([]any); !ok || len(values) == 0 {
		t.Fatalf("odata has no value list: %s", w.Body.String())
	}
}

func TestUnknownRedfishPathsAnswerARedfishNotFound(t *testing.T) {
	h := newHarness(t)

	for _, path := range []string{"/redfish/v1/Nope", "/redfish/v1/Systems/2", "/redfish/v2"} {
		w := h.do(http.MethodGet, path, "")
		expectError(t, w, http.StatusNotFound, "ResourceMissingAtURI")
	}
}

// NoRoute belongs to the whole engine, so a path outside /redfish must still
// get gin's own 404.
func TestPathsOutsideRedfishKeepTheDefaultNotFound(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/nope", "")
	if w.Code != http.StatusNotFound || w.Body.String() != "404 page not found" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}

func TestAMethodAResourceLacksIsNotAllowed(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodDelete, "/redfish/v1/", "", h.admin()...)
	expectError(t, w, http.StatusMethodNotAllowed, "GeneralError")
	if got := w.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow is %q", got)
	}
}

func TestErrorsNameTheirRegistryMessage(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/redfish/v1/Nope", "")
	e := decode(t, w)["error"].(map[string]any)
	info := e["@Message.ExtendedInfo"].([]any)[0].(map[string]any)
	if info["Message"] != "The resource at the URI '/redfish/v1/Nope' was not found." {
		t.Fatalf("Message is %v", info["Message"])
	}
	if args := info["MessageArgs"].([]any); len(args) != 1 || args[0] != "/redfish/v1/Nope" {
		t.Fatalf("MessageArgs is %v", info["MessageArgs"])
	}
}
