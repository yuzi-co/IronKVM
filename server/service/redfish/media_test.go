package redfish

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/storage"
)

const (
	mediaURL   = "/redfish/v1/Managers/1/VirtualMedia"
	cdInsert   = mediaURL + "/Cd/Actions/VirtualMedia.InsertMedia"
	cdEject    = mediaURL + "/Cd/Actions/VirtualMedia.EjectMedia"
	diskInsert = mediaURL + "/Disk/Actions/VirtualMedia.InsertMedia"
)

// image puts an empty file in the harness's image directory and returns its
// path.
func (h *harness) image(name string) string {
	h.t.Helper()

	path := filepath.Join(h.imageDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		h.t.Fatal(err)
	}
	return path
}

func TestVirtualMediaListsTheDrivesTheGadgetHas(t *testing.T) {
	for _, tc := range []struct {
		drives []proto.DriveInfo
		want   string
	}{
		{[]proto.DriveInfo{{ID: "disk"}, {ID: "cdrom", Ro: true}}, mediaURL + "/Cd," + mediaURL + "/Disk"},
		{[]proto.DriveInfo{{ID: "disk"}}, mediaURL + "/Disk"},
		{[]proto.DriveInfo{}, ""},
	} {
		h := newHarness(t)
		h.host.drives = tc.drives

		body := decode(t, h.do(http.MethodGet, mediaURL, "", h.user()...))
		var got []string
		for _, member := range body["Members"].([]any) {
			got = append(got, member.(map[string]any)["@odata.id"].(string))
		}
		if strings.Join(got, ",") != tc.want {
			t.Fatalf("members %v, want %s", got, tc.want)
		}
	}
}

func TestAnEmptyCdReportsNoMedium(t *testing.T) {
	h := newHarness(t)

	w := h.do(http.MethodGet, mediaURL+"/Cd", "", h.user()...)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	body := decode(t, w)
	if body["Inserted"] != false || body["Image"] != nil || body["ImageName"] != nil {
		t.Fatalf("empty CD is %v", body)
	}
	if body["ConnectedVia"] != "Applet" || body["WriteProtected"] != true {
		t.Fatalf("empty CD is %v", body)
	}
	if w.Header().Get("ETag") == "" {
		t.Fatal("no ETag")
	}
}

func TestALoadedDiskReportsItsImage(t *testing.T) {
	h := newHarness(t)
	h.host.drives[0].File = "/data/tools.img"

	body := decode(t, h.do(http.MethodGet, mediaURL+"/Disk", "", h.user()...))
	if body["Inserted"] != true || body["Image"] != "/data/tools.img" || body["ImageName"] != "tools.img" {
		t.Fatalf("disk is %v", body)
	}
	if body["WriteProtected"] != false {
		t.Fatalf("WriteProtected is %v", body["WriteProtected"])
	}
}

func TestAMissingDriveIsNotFound(t *testing.T) {
	h := newHarness(t)
	h.host.drives = []proto.DriveInfo{{ID: "disk"}}

	expectError(t, h.do(http.MethodGet, mediaURL+"/Cd", "", h.user()...), http.StatusNotFound, "ResourceMissingAtURI")
	expectError(t, h.do(http.MethodGet, mediaURL+"/Floppy", "", h.user()...), http.StatusNotFound, "ResourceMissingAtURI")
	expectError(t, h.do(http.MethodPost, cdInsert, `{"Image":"a.iso"}`, h.admin()...), http.StatusNotFound, "ResourceMissingAtURI")
}

func TestInsertTakesABareNameOrAPathUnderTheImageDirectory(t *testing.T) {
	h := newHarness(t)
	path := h.image("win11.iso")

	for _, image := range []string{"win11.iso", path} {
		w := h.do(http.MethodPost, cdInsert, `{"Image":"`+image+`","Inserted":true}`, h.admin()...)
		if w.Code != http.StatusNoContent {
			t.Fatalf("%s: %d %s", image, w.Code, w.Body.String())
		}
	}
	want := "cdrom " + path + " ro=true"
	if len(h.host.inserts) != 2 || h.host.inserts[0] != want || h.host.inserts[1] != want {
		t.Fatalf("inserts %v, want %q twice", h.host.inserts, want)
	}
}

func TestWriteProtectedAppliesToTheDiskOnly(t *testing.T) {
	h := newHarness(t)
	path := h.image("tools.img")
	iso := h.image("win11.iso")

	h.do(http.MethodPost, diskInsert, `{"Image":"tools.img"}`, h.admin()...)
	h.do(http.MethodPost, diskInsert, `{"Image":"tools.img","WriteProtected":false}`, h.admin()...)
	h.do(http.MethodPost, cdInsert, `{"Image":"win11.iso","WriteProtected":false}`, h.admin()...)

	want := []string{
		"disk " + path + " ro=true",
		"disk " + path + " ro=false",
		"cdrom " + iso + " ro=true",
	}
	if strings.Join(h.host.inserts, "|") != strings.Join(want, "|") {
		t.Fatalf("inserts %v, want %v", h.host.inserts, want)
	}
}

func TestInsertRefusesAURL(t *testing.T) {
	h := newHarness(t)
	h.image("win11.iso")

	for _, image := range []string{"http://10.0.0.5/win11.iso", "https://example.com/win11.iso", "nfs://nas/win11.iso"} {
		w := h.do(http.MethodPost, cdInsert, `{"Image":"`+image+`"}`, h.admin()...)
		expectError(t, w, http.StatusBadRequest, "ActionParameterValueNotInList")
	}
	if len(h.host.inserts) != 0 {
		t.Fatalf("inserted %v", h.host.inserts)
	}
}

func TestInsertRefusesAnImageThatIsNotThere(t *testing.T) {
	h := newHarness(t)
	h.image("dir.iso/inner.iso")
	outside := filepath.Join(t.TempDir(), "outside.iso")
	if err := os.WriteFile(outside, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	for _, image := range []string{"missing.iso", "dir.iso", "../" + filepath.Base(filepath.Dir(outside)) + "/outside.iso", outside} {
		w := h.do(http.MethodPost, cdInsert, `{"Image":"`+image+`"}`, h.admin()...)
		expectError(t, w, http.StatusBadRequest, "ResourceNotFound")
	}
	if len(h.host.inserts) != 0 {
		t.Fatalf("inserted %v", h.host.inserts)
	}
}

func TestInsertChecksItsParameters(t *testing.T) {
	h := newHarness(t)
	h.image("win11.iso")

	for _, tc := range []struct {
		body string
		id   string
	}{
		{`{}`, "ActionParameterMissing"},
		{`{"Image":""}`, "ActionParameterValueTypeError"},
		{`{"Image":5}`, "ActionParameterValueTypeError"},
		{`{"Image":"win11.iso","Inserted":false}`, "ActionParameterValueNotInList"},
		{`{"Image":"win11.iso","TransferMethod":"Stream"}`, "ActionParameterNotSupported"},
		{`{"Image":"win11.iso","TransferProtocolType":"HTTP"}`, "ActionParameterNotSupported"},
		{`{"Image":"win11.iso","UserName":"u","Password":"p"}`, "ActionParameterNotSupported"},
	} {
		expectError(t, h.do(http.MethodPost, cdInsert, tc.body, h.admin()...), http.StatusBadRequest, tc.id)
	}
	if len(h.host.inserts) != 0 {
		t.Fatalf("inserted %v", h.host.inserts)
	}
}

// Some clients send every property and set the ones they do not use to null.
func TestInsertTakesNullTransferProperties(t *testing.T) {
	h := newHarness(t)
	h.image("win11.iso")

	w := h.do(http.MethodPost, cdInsert, `{"Image":"win11.iso","TransferMethod":null,"TransferProtocolType":null}`, h.admin()...)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
}

func TestStorageErrorsMapToRedfishMessages(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		id     string
	}{
		{storage.ErrInvalidImage, http.StatusBadRequest, "ResourceNotFound"},
		{storage.ErrInOtherDrive, http.StatusConflict, "ResourceInUse"},
		{storage.ErrMediumLocked, http.StatusBadRequest, "ActionNotSupported"},
		{storage.ErrNoDrive, http.StatusNotFound, "ResourceMissingAtURI"},
	} {
		h := newHarness(t)
		h.image("win11.iso")
		h.host.insertErr = tc.err
		h.host.ejectErr = tc.err

		expectError(t, h.do(http.MethodPost, cdInsert, `{"Image":"win11.iso"}`, h.admin()...), tc.status, tc.id)
		if tc.err != storage.ErrInvalidImage {
			expectError(t, h.do(http.MethodPost, cdEject, `{}`, h.admin()...), tc.status, tc.id)
		}
	}
}

func TestALockedMediumCarriesTheStorageText(t *testing.T) {
	h := newHarness(t)
	h.host.ejectErr = storage.ErrMediumLocked

	w := h.do(http.MethodPost, cdEject, "", h.admin()...)
	if message := decode(t, w)["error"].(map[string]any)["message"]; message != storage.ErrMediumLocked.Error() {
		t.Fatalf("message is %v", message)
	}
}

func TestEjectEmptiesTheDrive(t *testing.T) {
	h := newHarness(t)
	h.host.drives[1].File = "/data/win11.iso"

	// A bare POST, with no body and no Content-Type, as some clients send.
	w := h.do(http.MethodPost, cdEject, "", h.admin()...)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	if len(h.host.ejects) != 1 || h.host.ejects[0] != "cdrom" {
		t.Fatalf("ejects %v", h.host.ejects)
	}

	expectError(t, h.do(http.MethodPost, cdEject, `{"Force":true}`, h.admin()...), http.StatusBadRequest, "ActionParameterNotSupported")
}

func TestMediaActionsNeedAnAdmin(t *testing.T) {
	h := newHarness(t)
	h.image("win11.iso")

	expectError(t, h.do(http.MethodPost, cdInsert, `{"Image":"win11.iso"}`, h.user()...), http.StatusForbidden, "InsufficientPrivilege")
	expectError(t, h.do(http.MethodPost, cdEject, `{}`, h.user()...), http.StatusForbidden, "InsufficientPrivilege")
	if len(h.host.inserts) != 0 || len(h.host.ejects) != 0 {
		t.Fatal("a user changed the media")
	}
}
