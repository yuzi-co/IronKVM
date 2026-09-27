package redfish

import (
	"net/http"
	"sort"
	"strings"
	"testing"
)

// collectLinks gathers every @odata.id below the top level of a resource.
func collectLinks(value any, top bool, out *[]string) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			if key == "@odata.id" && !top {
				if s, ok := child.(string); ok {
					*out = append(*out, s)
				}
				continue
			}
			collectLinks(child, false, out)
		}
	case []any:
		for _, child := range v {
			collectLinks(child, false, out)
		}
	}
}

// Every resource reachable from the service root must answer, carry its
// OData annotations, name itself by the path it was fetched at, and use a
// type whose schema $metadata references. A validator walks the service the
// same way, so a dangling link or a missing reference fails here first.
func TestEveryLinkResolvesToAnAnnotatedResource(t *testing.T) {
	h := newHarness(t)
	h.host.drives[1].File = "/data/win11.iso"
	h.nics = []NIC{{ID: "eth0", MAC: "48:da:35:6e:00:01", IPv4: "10.0.0.222"}}
	token := h.admin()

	metadataBody := h.do(http.MethodGet, "/redfish/v1/$metadata", "").Body.String()

	seen := map[string]bool{}
	queue := []string{"/redfish/v1/"}
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		if seen[path] {
			continue
		}
		seen[path] = true

		w := h.do(http.MethodGet, path, "", token...)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s: %d %s", path, w.Code, w.Body.String())
		}
		if w.Header().Get("OData-Version") != "4.0" {
			t.Fatalf("GET %s: no OData-Version header", path)
		}

		body := decode(t, w)
		if body["@odata.id"] != path {
			t.Fatalf("GET %s: @odata.id is %v", path, body["@odata.id"])
		}
		odataType, _ := body["@odata.type"].(string)
		if !strings.HasPrefix(odataType, "#") {
			t.Fatalf("GET %s: @odata.type is %v", path, body["@odata.type"])
		}
		if context, _ := body["@odata.context"].(string); !strings.HasPrefix(context, "/redfish/v1/$metadata#") {
			t.Fatalf("GET %s: @odata.context is %v", path, body["@odata.context"])
		}
		if _, ok := body["Name"]; !ok {
			t.Fatalf("GET %s: no Name", path)
		}

		// "#ComputerSystem.v1_13_0.ComputerSystem" needs the namespace
		// "ComputerSystem.v1_13_0"; a collection type needs its own name.
		namespace := odataType[1:strings.LastIndex(odataType, ".")]
		if !strings.Contains(metadataBody, `Namespace="`+namespace+`"`) {
			t.Fatalf("GET %s: $metadata does not include %s", path, namespace)
		}

		var found []string
		collectLinks(body, true, &found)
		sort.Strings(found)
		for _, link := range found {
			if !strings.HasPrefix(link, "/redfish/v1/") {
				t.Fatalf("GET %s: link %q leaves the service", path, link)
			}
			queue = append(queue, link)
		}
	}

	for _, want := range []string{
		"/redfish/v1/Systems/1",
		"/redfish/v1/Chassis/1",
		"/redfish/v1/Managers/1",
		"/redfish/v1/Managers/1/EthernetInterfaces/eth0",
		"/redfish/v1/Managers/1/VirtualMedia/Cd",
		"/redfish/v1/Managers/1/VirtualMedia/Disk",
		"/redfish/v1/SessionService",
	} {
		if !seen[want] {
			t.Fatalf("%s is not reachable from the service root", want)
		}
	}
}
