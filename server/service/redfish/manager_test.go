package redfish

import (
	"net/http"
	"testing"
)

func TestManagerDescribesTheBoard(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/redfish/v1/Managers/1", "", h.user()...)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	body := decode(t, w)
	if body["ManagerType"] != "BMC" || body["FirmwareVersion"] != "2.3.0 (image v1.4.3)" {
		t.Fatalf("manager is %v", body)
	}
	if body["VirtualMedia"].(map[string]any)["@odata.id"] != "/redfish/v1/Managers/1/VirtualMedia" {
		t.Fatalf("VirtualMedia link is %v", body["VirtualMedia"])
	}
	if body["EthernetInterfaces"].(map[string]any)["@odata.id"] != "/redfish/v1/Managers/1/EthernetInterfaces" {
		t.Fatalf("EthernetInterfaces link is %v", body["EthernetInterfaces"])
	}
}

func TestEthernetInterfacesListTheBoardsInterfaces(t *testing.T) {
	h := newHarness(t)
	h.nics = []NIC{
		{ID: "eth0", MAC: "48:da:35:6e:00:01", IPv4: "10.0.0.222"},
		{ID: "wlan0", MAC: "48:da:35:6e:00:02"},
	}

	body := decode(t, h.do(http.MethodGet, "/redfish/v1/Managers/1/EthernetInterfaces", "", h.user()...))
	if body["Members@odata.count"] != float64(2) {
		t.Fatalf("collection is %v", body)
	}

	w := h.do(http.MethodGet, "/redfish/v1/Managers/1/EthernetInterfaces/eth0", "", h.user()...)
	if w.Code != http.StatusOK {
		t.Fatalf("eth0: %d", w.Code)
	}
	eth0 := decode(t, w)
	if eth0["MACAddress"] != "48:da:35:6e:00:01" {
		t.Fatalf("MACAddress is %v", eth0["MACAddress"])
	}
	addresses := eth0["IPv4Addresses"].([]any)
	if len(addresses) != 1 || addresses[0].(map[string]any)["Address"] != "10.0.0.222" {
		t.Fatalf("IPv4Addresses are %v", addresses)
	}

	wlan0 := decode(t, h.do(http.MethodGet, "/redfish/v1/Managers/1/EthernetInterfaces/wlan0", "", h.user()...))
	if addresses := wlan0["IPv4Addresses"].([]any); len(addresses) != 0 {
		t.Fatalf("wlan0 IPv4Addresses are %v, want none", addresses)
	}
}

func TestAnInterfaceWithoutAMACOmitsIt(t *testing.T) {
	h := newHarness(t)
	h.nics = []NIC{{ID: "eth0", IPv4: "10.0.0.222"}}

	body := decode(t, h.do(http.MethodGet, "/redfish/v1/Managers/1/EthernetInterfaces/eth0", "", h.user()...))
	if _, ok := body["MACAddress"]; ok {
		t.Fatalf("MACAddress is %v, want it absent", body["MACAddress"])
	}
}

func TestNoInterfacesIsAnEmptyCollection(t *testing.T) {
	h := newHarness(t)
	h.nics = nil

	body := decode(t, h.do(http.MethodGet, "/redfish/v1/Managers/1/EthernetInterfaces", "", h.user()...))
	if body["Members@odata.count"] != float64(0) || len(body["Members"].([]any)) != 0 {
		t.Fatalf("collection is %v", body)
	}
}

func TestAnUnknownInterfaceIsNotFound(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, "/redfish/v1/Managers/1/EthernetInterfaces/eth9", "", h.user()...)
	expectError(t, w, http.StatusNotFound, "ResourceMissingAtURI")
}
