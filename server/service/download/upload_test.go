package download

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// upload posts body as an image named name and returns the response code.
//
// The handler takes the transfer sentinel in /run, so the test runs only where
// it can create that file and nothing else holds it: a build container, not a
// board with a transfer running.
func upload(t *testing.T, name string, body []byte) int {
	t.Helper()

	if _, err := os.Stat(transferSentinelPath); err == nil {
		t.Skipf("%s exists, a transfer is running here", transferSentinelPath)
	}
	probe, err := os.OpenFile(transferSentinelPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		t.Skipf("cannot take the transfer sentinel here: %s", err)
	}
	_ = probe.Close()
	_ = os.Remove(transferSentinelPath)

	oldDir := imageDir
	t.Cleanup(func() { imageDir = oldDir })
	imageDir = t.TempDir()

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/upload", NewService().DownloadImageFile)

	request := httptest.NewRequest(http.MethodPost, "/upload", &form)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var rsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %s", recorder.Body.String(), err)
	}

	return rsp.Code
}

// imageFiles lists what the upload left in the image directory.
func imageFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(imageDir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}

// The header is read before the file is created and written back in front of
// the rest, so the stored image has to be the upload byte for byte.
func TestAnUploadedImageIsStoredWhole(t *testing.T) {
	body := headWithDescriptor("CD001")
	body = append(body, bytes.Repeat([]byte{0x5A}, 100_000)...)

	if code := upload(t, "a.iso", body); code != 0 {
		t.Fatalf("a valid image was refused, code %d", code)
	}

	stored, err := os.ReadFile(filepath.Join(imageDir, "a.iso"))
	if err != nil {
		t.Fatalf("the image was not stored: %s", err)
	}
	if !bytes.Equal(stored, body) {
		t.Fatalf("stored %d bytes that differ from the %d uploaded", len(stored), len(body))
	}
	if names := imageFiles(t); len(names) != 1 {
		t.Fatalf("the image directory holds %v, want the image alone", names)
	}
}

// An image shorter than the volume descriptor sector is judged on what there
// is: a raw disk image with an MBR passes.
func TestAShortDiskImageIsStoredWhole(t *testing.T) {
	body := make([]byte, 4096)
	body[0x1FE], body[0x1FF] = 0x55, 0xAA

	if code := upload(t, "b.iso", body); code != 0 {
		t.Fatalf("a short disk image was refused, code %d", code)
	}

	stored, err := os.ReadFile(filepath.Join(imageDir, "b.iso"))
	if err != nil || !bytes.Equal(stored, body) {
		t.Fatalf("the short image was not stored intact (%v)", err)
	}
}

// Anything else is refused on its first sectors, and nothing is left behind.
func TestAnUploadThatIsNoImageLeavesNothing(t *testing.T) {
	body := bytes.Repeat([]byte{0xA5}, imageHeaderSize+100_000)

	if code := upload(t, "c.iso", body); code == 0 {
		t.Fatal("a file that is no image was accepted")
	}
	if names := imageFiles(t); len(names) != 0 {
		t.Fatalf("a refused upload left %v", names)
	}
}
