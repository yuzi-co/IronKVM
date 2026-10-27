package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/storage/diskfs"
	"NanoKVM-Server/service/storage/mkfs"
)

// createFixture points the image directory at a temporary one with 4 GiB
// free, as statfs would report it.
func createFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	oldRoot := imageRoot
	imageRoot = root
	oldFree := freeBytes
	freeBytes = func(path string) (int64, error) {
		if path != root {
			t.Errorf("free space asked of %s, want %s", path, root)
		}
		return 4 << 30, nil
	}
	t.Cleanup(func() {
		imageRoot = oldRoot
		freeBytes = oldFree
	})
	return root
}

func createRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := NewService()
	r.POST("/image/create", s.CreateImage)
	r.GET("/image/space", s.GetImageSpace)
	return r
}

func createCall(t *testing.T, name string, size int64, fs string) apiRsp {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"name": name, "size": size, "fs": fs})
	req := httptest.NewRequest(http.MethodPost, "/image/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	createRouter().ServeHTTP(rec, req)
	var rsp apiRsp
	if err := json.Unmarshal(rec.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %s", rec.Body.String(), err)
	}
	return rsp
}

// A new exchange image lands in the image directory, in the list, with the
// filesystem asked for.
func TestCreateImageAddsAnExchangeImage(t *testing.T) {
	root := createFixture(t)

	for _, c := range []struct {
		name, fs, file string
		kind           diskfs.Kind
	}{
		{"exchange", "", "exchange.img", diskfs.KindExFAT},
		{"old host.IMG", "fat32", "old host.IMG", diskfs.KindFAT32},
		{"  spaced ", "exfat", "spaced.img", diskfs.KindExFAT},
	} {
		rsp := createCall(t, c.name, 64<<20, c.fs)
		if rsp.Code != 0 {
			t.Fatalf("create %q: %d %s", c.name, rsp.Code, rsp.Msg)
		}
		var data proto.CreateImageRsp
		if err := json.Unmarshal(rsp.Data, &data); err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(root, c.file)
		if data.File != want {
			t.Errorf("created %s, want %s", data.File, want)
		}
		info, err := os.Stat(want)
		if err != nil || info.Size() != 64<<20 {
			t.Fatalf("%s: %v, %v", want, info, err)
		}
		if vol, err := diskfs.Detect(want); err != nil || vol.Kind != c.kind {
			t.Errorf("%s holds %+v, %v; want %s", want, vol, err, c.kind)
		}
	}

	images, _, err := listImages(root)
	if err != nil || len(images) != 3 {
		t.Errorf("the list holds %v, %v", images, err)
	}
}

func TestCreateImageRefusesAnExistingName(t *testing.T) {
	root := createFixture(t)
	path := filepath.Join(root, "exchange.img")
	mustWriteFile(t, path, "a disk somebody made")

	if rsp := createCall(t, "exchange", 64<<20, "exfat"); rsp.Code == 0 {
		t.Fatal("an existing image was overwritten")
	}
	if data, _ := os.ReadFile(path); string(data) != "a disk somebody made" {
		t.Errorf("the image now holds %d bytes", len(data))
	}
}

func TestCreateImageValidatesTheRequest(t *testing.T) {
	root := createFixture(t)

	for _, c := range []struct {
		why  string
		name string
		size int64
		fs   string
		code int
	}{
		{"no name", "", 64 << 20, "exfat", codeInvalidArgument},
		{"a path", "../escape", 64 << 20, "exfat", codeInvalidArgument},
		{"a subdirectory", "dir/x", 64 << 20, "exfat", codeInvalidArgument},
		{"a backslash", `dir\x`, 64 << 20, "exfat", codeInvalidArgument},
		{"a Windows device name", "con", 64 << 20, "exfat", codeInvalidArgument},
		{"a forbidden character", "what?", 64 << 20, "exfat", codeInvalidArgument},
		{"a trailing dot", "name.", 64 << 20, "exfat", codeInvalidArgument},
		{"a CD image", "disc.iso", 64 << 20, "exfat", codeInvalidArgument},
		{"a hidden name", ".hidden", 64 << 20, "exfat", codeInvalidArgument},
		{"an unknown filesystem", "x", 64 << 20, "ntfs", codeInvalidArgument},
		{"too small", "x", 1 << 20, "exfat", codeInvalidArgument},
		{"no size", "x", 0, "exfat", codeInvalidArgument},
		{"not whole MiB", "x", 64<<20 + 512, "exfat", codeInvalidArgument},
		{"more than the card has", "x", 4 << 30, "exfat", codeNoSpace},
		{"into the reserve", "x", 4<<30 - 128<<20, "exfat", codeNoSpace},
		{"a floppy of a disk's size", "x", 64 << 20, "fat12", codeInvalidArgument},
		{"a floppy of no floppy size", "x", 1440000, "fat12", codeInvalidArgument},
		{"a disk of a floppy's size", "x", 1474560, "fat32", codeInvalidArgument},
	} {
		rsp := createCall(t, c.name, c.size, c.fs)
		if rsp.Code != c.code {
			t.Errorf("%s: code %d (%s), want %d", c.why, rsp.Code, rsp.Msg, c.code)
		}
	}

	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Errorf("refused requests left %d files", len(entries))
	}
}

// A floppy image is a bare FAT12 volume of one of the floppy sizes, made
// like any other exchange image and listed with them.
func TestCreateImageMakesAFloppy(t *testing.T) {
	root := createFixture(t)

	for _, size := range []int64{737280, 1474560, 2949120} {
		name := fmt.Sprintf("floppy%d", size)
		rsp := createCall(t, name, size, "fat12")
		if rsp.Code != 0 {
			t.Fatalf("create %d: %d %s", size, rsp.Code, rsp.Msg)
		}
		path := filepath.Join(root, name+".img")
		info, err := os.Stat(path)
		if err != nil || info.Size() != size {
			t.Fatalf("%s: %v, %v", path, info, err)
		}
		if vol, err := diskfs.Detect(path); err != nil || vol.Kind != diskfs.KindFAT12 {
			t.Errorf("%s holds %+v, %v", path, vol, err)
		}
	}
	if images, _, err := listImages(root); err != nil || len(images) != 3 {
		t.Errorf("the list holds %v, %v", images, err)
	}
}

// A floppy still leaves the card its reserve.
func TestCreateFloppyKeepsTheReserve(t *testing.T) {
	createFixture(t)
	freeBytes = func(string) (int64, error) { return reservedFreeBytes + 1<<20, nil }

	if rsp := createCall(t, "x", 1474560, "fat12"); rsp.Code != codeNoSpace {
		t.Errorf("code %d (%s), want %d", rsp.Code, rsp.Msg, codeNoSpace)
	}
}

// The dialog offers sizes up to what the card can spare.
func TestImageSpaceReportsTheLimits(t *testing.T) {
	createFixture(t)

	rec := httptest.NewRecorder()
	createRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/image/space", nil))
	var rsp apiRsp
	if err := json.Unmarshal(rec.Body.Bytes(), &rsp); err != nil || rsp.Code != 0 {
		t.Fatalf("%s: %v", rec.Body.String(), err)
	}
	var data proto.ImageSpaceRsp
	if err := json.Unmarshal(rsp.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Min != mkfs.MinImageSize || data.Max != 4<<30-reservedFreeBytes || data.Free != 4<<30 {
		t.Errorf("space %+v", data)
	}
}
