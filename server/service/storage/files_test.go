package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/service/storage/diskfs"
)

// filesFixture is an image directory with one image, x.img, whose "mount" is
// a plain directory: the handlers' policy is under test here, and diskfs has
// its own tests for real images.
type filesFixture struct {
	gadget  string
	image   string
	content string
	mounts  []bool // the writable flag of every mount asked for
	free    uint64
	kind    diskfs.Kind
}

type fakeImage struct {
	diskfs.Image
	free uint64
}

func (f fakeImage) StatFS() (uint64, uint64, error) { return 1 << 40, f.free, nil }

func newFilesFixture(t *testing.T) *filesFixture {
	t.Helper()
	f := &filesFixture{
		gadget:  fakeGadget(t, "lun.0", "lun.1"),
		content: t.TempDir(),
		free:    1 << 30,
		kind:    diskfs.KindFAT16,
	}
	root := t.TempDir()
	oldRoot := imageRoot
	imageRoot = root
	t.Cleanup(func() { imageRoot = oldRoot })

	f.image = filepath.Join(root, "x.img")
	if err := os.WriteFile(f.image, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldMount := mountImage
	mountImage = func(path string, writable bool) (diskfs.Image, error) {
		if path != f.image {
			t.Errorf("mounted %s, want %s", path, f.image)
		}
		f.mounts = append(f.mounts, writable)
		if writable && !f.kind.Writable() {
			return nil, diskfs.ErrReadOnly
		}
		image, err := diskfs.OpenDir(f.content, f.kind, writable)
		if err != nil {
			return nil, err
		}
		return fakeImage{Image: image, free: f.free}, nil
	}
	t.Cleanup(func() { mountImage = oldMount })

	oldJobs := transfers
	transfers = &jobSlot{}
	t.Cleanup(func() { transfers = oldJobs })
	return f
}

// load puts the image into a drive by hand, as the host would see it.
func (f *filesFixture) load(t *testing.T, lun string, ro bool) {
	t.Helper()
	flag := "0"
	if ro {
		flag = "1"
	}
	if err := os.WriteFile(filepath.Join(f.gadget, lun, "ro"), []byte(flag), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.gadget, lun, "file"), []byte(f.image+"\n"), 0o666); err != nil {
		t.Fatal(err)
	}
}

type apiRsp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func filesRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := NewService()
	r.GET("/files", s.ListFiles)
	r.GET("/files/download", s.DownloadFile)
	r.POST("/files/upload", s.UploadFile)
	r.POST("/files/delete", s.DeleteFile)
	r.GET("/files/status", s.FileJobStatus)
	return r
}

func call(t *testing.T, req *http.Request) (*httptest.ResponseRecorder, apiRsp) {
	t.Helper()
	rec := httptest.NewRecorder()
	filesRouter().ServeHTTP(rec, req)
	var rsp apiRsp
	if strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(rec.Body.Bytes(), &rsp); err != nil {
			t.Fatalf("response %q: %s", rec.Body.String(), err)
		}
	}
	return rec, rsp
}

func query(file, path string) string {
	return "file=" + url.QueryEscape(file) + "&path=" + url.QueryEscape(path)
}

func list(t *testing.T, file, path string) apiRsp {
	t.Helper()
	_, rsp := call(t, httptest.NewRequest(http.MethodGet, "/files?"+query(file, path), nil))
	return rsp
}

func uploadFile(t *testing.T, file, dir, name, body string, contentLength int64) apiRsp {
	t.Helper()
	var form bytes.Buffer
	w := multipart.NewWriter(&form)
	part, err := w.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(body))
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/files/upload?"+query(file, dir), &form)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if contentLength > 0 {
		req.ContentLength = contentLength
	}
	_, rsp := call(t, req)
	return rsp
}

func deleteFile(t *testing.T, file, path string) apiRsp {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"file": file, "path": path})
	req := httptest.NewRequest(http.MethodPost, "/files/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	_, rsp := call(t, req)
	return rsp
}

func TestListFilesInAnImage(t *testing.T) {
	f := newFilesFixture(t)
	mustWriteFile(t, filepath.Join(f.content, "docs", "a.txt"), "aaa")
	mustWriteFile(t, filepath.Join(f.content, "readme.txt"), "r")

	rsp := list(t, f.image, "/")
	if rsp.Code != 0 {
		t.Fatalf("code %d: %s", rsp.Code, rsp.Msg)
	}
	var data struct {
		Path     string `json:"path"`
		FS       string `json:"fs"`
		Writable bool   `json:"writable"`
		Drive    string `json:"drive"`
		Entries  []struct {
			Name  string `json:"name"`
			Size  int64  `json:"size"`
			IsDir bool   `json:"isDir"`
		} `json:"entries"`
		Free uint64 `json:"free"`
	}
	if err := json.Unmarshal(rsp.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Path != "/" || data.FS != "fat16" || !data.Writable || data.Drive != "" || data.Free != 1<<30 {
		t.Errorf("data = %+v", data)
	}
	if len(data.Entries) != 2 || data.Entries[0].Name != "docs" || !data.Entries[0].IsDir ||
		data.Entries[1].Name != "readme.txt" || data.Entries[1].Size != 1 {
		t.Errorf("entries = %+v", data.Entries)
	}
	// Browsing never mounts for writing.
	if len(f.mounts) != 1 || f.mounts[0] {
		t.Errorf("mounts = %v", f.mounts)
	}
}

// The image is named the way the drive API names it, and only an image under
// the image directory is one.
func TestFilesRefuseAnythingButAnImage(t *testing.T) {
	f := newFilesFixture(t)
	outside := filepath.Join(t.TempDir(), "y.img")
	mustWriteFile(t, outside, "y")
	if err := os.Symlink(outside, filepath.Join(imageRoot, "link.img")); err != nil {
		t.Fatal(err)
	}

	for _, file := range []string{
		"",
		"/etc/shadow",
		outside,
		filepath.Join(imageRoot, "..", "y.img"),
		filepath.Join(imageRoot, "link.img"),
		filepath.Join(imageRoot, "notes.txt"),
	} {
		if rsp := list(t, file, "/"); rsp.Code != codeInvalidArgument {
			t.Errorf("list %q: code %d (%s), want %d", file, rsp.Code, rsp.Msg, codeInvalidArgument)
		}
	}
	if len(f.mounts) != 0 {
		t.Errorf("mounted %d times", len(f.mounts))
	}
}

func TestFilesRefuseAPathThatLeavesTheImage(t *testing.T) {
	f := newFilesFixture(t)
	for _, p := range []string{"..", "../x.img", "docs/../../..", "a\x00b", "con"} {
		if rsp := list(t, f.image, p); rsp.Code != codeInvalidArgument {
			t.Errorf("list %q: code %d, want %d", p, rsp.Code, codeInvalidArgument)
		}
		if rsp := deleteFile(t, f.image, p); rsp.Code != codeInvalidArgument {
			t.Errorf("delete %q: code %d, want %d", p, rsp.Code, codeInvalidArgument)
		}
	}
	if len(f.mounts) != 0 {
		t.Errorf("mounted %d times", len(f.mounts))
	}
}

// The host and the board must never both write one filesystem. An image the
// host can write is not even read here, since its view changes under us; one
// it reads only can be browsed, but not changed.
func TestAServedImageIsReadOnlyHereOrNotOpenedAtAll(t *testing.T) {
	t.Run("in the cd drive", func(t *testing.T) {
		f := newFilesFixture(t)
		f.load(t, "lun.1", true)

		rsp := list(t, f.image, "/")
		if rsp.Code != 0 {
			t.Fatalf("list: code %d: %s", rsp.Code, rsp.Msg)
		}
		var data struct {
			Writable bool   `json:"writable"`
			Drive    string `json:"drive"`
		}
		_ = json.Unmarshal(rsp.Data, &data)
		if data.Writable || data.Drive != "cdrom" {
			t.Errorf("data = %+v", data)
		}

		if rsp := uploadFile(t, f.image, "/", "new.txt", "x", 0); rsp.Code != codeMediumBusy {
			t.Errorf("upload: code %d (%s), want %d", rsp.Code, rsp.Msg, codeMediumBusy)
		}
		mustWriteFile(t, filepath.Join(f.content, "keep.txt"), "k")
		if rsp := deleteFile(t, f.image, "keep.txt"); rsp.Code != codeMediumBusy {
			t.Errorf("delete: code %d, want %d", rsp.Code, codeMediumBusy)
		}
		if _, err := os.Stat(filepath.Join(f.content, "new.txt")); err == nil {
			t.Error("the upload landed")
		}
		if _, err := os.Stat(filepath.Join(f.content, "keep.txt")); err != nil {
			t.Error("the delete happened")
		}
		for _, w := range f.mounts {
			if w {
				t.Error("mounted for writing")
			}
		}
	})

	t.Run("in a writable disk", func(t *testing.T) {
		f := newFilesFixture(t)
		f.load(t, "lun.0", false)

		if rsp := list(t, f.image, "/"); rsp.Code != codeMediumBusy {
			t.Errorf("list: code %d (%s), want %d", rsp.Code, rsp.Msg, codeMediumBusy)
		}
		req := httptest.NewRequest(http.MethodGet, "/files/download?"+query(f.image, "a.txt"), nil)
		if _, rsp := call(t, req); rsp.Code != codeMediumBusy {
			t.Errorf("download: code %d, want %d", rsp.Code, codeMediumBusy)
		}
		if len(f.mounts) != 0 {
			t.Errorf("mounted %d times", len(f.mounts))
		}
	})
}

func TestUploadWritesTheFileAndReportsTheJob(t *testing.T) {
	f := newFilesFixture(t)
	if err := os.Mkdir(filepath.Join(f.content, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	rsp := uploadFile(t, f.image, "/docs", "../../report.txt", "the report", 0)
	if rsp.Code != 0 {
		t.Fatalf("code %d: %s", rsp.Code, rsp.Msg)
	}
	data, err := os.ReadFile(filepath.Join(f.content, "docs", "report.txt"))
	if err != nil || string(data) != "the report" {
		t.Fatalf("stored %q, %v", data, err)
	}
	if len(f.mounts) != 1 || !f.mounts[0] {
		t.Errorf("mounts = %v", f.mounts)
	}

	_, st := call(t, httptest.NewRequest(http.MethodGet, "/files/status", nil))
	var status struct {
		Status string `json:"status"`
		Name   string `json:"name"`
	}
	_ = json.Unmarshal(st.Data, &status)
	if status.Status != "success" || !strings.HasSuffix(status.Name, "docs/report.txt") {
		t.Errorf("status = %+v", status)
	}
}

func TestUploadRefusesBadNamesAndFullImages(t *testing.T) {
	f := newFilesFixture(t)

	if rsp := uploadFile(t, f.image, "/", "a:b.txt", "x", 0); rsp.Code != codeInvalidArgument {
		t.Errorf("a bad name: code %d, want %d", rsp.Code, codeInvalidArgument)
	}

	f.free = 100
	if rsp := uploadFile(t, f.image, "/", "big.bin", strings.Repeat("b", 500), 0); rsp.Code != codeNoSpace {
		t.Errorf("an upload larger than the free space: code %d (%s), want %d", rsp.Code, rsp.Msg, codeNoSpace)
	}

	// FAT holds no file of 4 GiB or more, whatever the free space says.
	f.free = 1 << 40
	if rsp := uploadFile(t, f.image, "/", "huge.bin", "h", 5<<30); rsp.Code != codeNoSpace {
		t.Errorf("a 5 GiB upload to FAT: code %d (%s), want %d", rsp.Code, rsp.Msg, codeNoSpace)
	}

	entries, _ := os.ReadDir(f.content)
	if len(entries) != 0 {
		t.Errorf("left %d entries", len(entries))
	}
}

func TestUploadIntoAReadOnlyFilesystemIsRefused(t *testing.T) {
	f := newFilesFixture(t)
	f.kind = diskfs.KindISO9660
	if rsp := uploadFile(t, f.image, "/", "a.txt", "x", 0); rsp.Code != codeReadOnly {
		t.Errorf("code %d (%s), want %d", rsp.Code, rsp.Msg, codeReadOnly)
	}
}

func TestDeleteRemovesAFileOrADirectory(t *testing.T) {
	f := newFilesFixture(t)
	mustWriteFile(t, filepath.Join(f.content, "a.txt"), "a")
	mustWriteFile(t, filepath.Join(f.content, "d", "b.txt"), "b")

	for _, p := range []string{"a.txt", "/d"} {
		if rsp := deleteFile(t, f.image, p); rsp.Code != 0 {
			t.Fatalf("delete %s: code %d: %s", p, rsp.Code, rsp.Msg)
		}
	}
	if rsp := deleteFile(t, f.image, "/"); rsp.Code != codeInvalidArgument {
		t.Errorf("delete of the root: code %d", rsp.Code)
	}
	entries, _ := os.ReadDir(f.content)
	if len(entries) != 0 {
		t.Errorf("left %d entries", len(entries))
	}
}

func TestDownloadStreamsTheFile(t *testing.T) {
	f := newFilesFixture(t)
	mustWriteFile(t, filepath.Join(f.content, "docs", "Résumé 1.txt"), "content")

	req := httptest.NewRequest(http.MethodGet, "/files/download?"+query(f.image, "docs/Résumé 1.txt"), nil)
	rec, _ := call(t, req)
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "content" {
		t.Fatalf("body %q", body)
	}
	want := `attachment; filename="R_sum_ 1.txt"; filename*=UTF-8''R%C3%A9sum%C3%A9%201.txt`
	if got := rec.Header().Get("Content-Disposition"); got != want {
		t.Errorf("Content-Disposition %q, want %q", got, want)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/download?"+query(f.image, "docs"), nil)
	if _, rsp := call(t, req); rsp.Code == 0 {
		t.Error("downloaded a directory")
	}
}

// While the board has an image open for writing, the host may not have it at
// all; while it reads it, the host may only read it too.
func TestInsertRespectsTheBrowser(t *testing.T) {
	f := newFilesFixture(t)
	open, writable := false, false
	oldState := openHere
	openHere = func(path string) (bool, bool) {
		if realPath(path) == realPath(f.image) {
			return open, writable
		}
		return false, false
	}
	t.Cleanup(func() { openHere = oldState })

	open, writable = true, true
	for _, tc := range []struct {
		id string
		ro bool
	}{{DriveDisk, true}, {DriveDisk, false}, {DriveCdrom, true}} {
		if _, err := insertLocked(tc.id, f.image, tc.ro, resolveAny); !errors.Is(err, errImageOpen) {
			t.Errorf("insert into %s (ro %v) while open for writing: %v", tc.id, tc.ro, err)
		}
	}

	open, writable = true, false
	if _, err := insertLocked(DriveDisk, f.image, false, resolveAny); !errors.Is(err, errImageOpen) {
		t.Errorf("a writable insert while read here: %v", err)
	}
	if _, err := insertLocked(DriveCdrom, f.image, true, resolveAny); err != nil {
		t.Errorf("a read-only insert while read here: %v", err)
	}
}

func TestRemoveImageRefusesAnImageOpenHere(t *testing.T) {
	f := newFilesFixture(t)
	oldState := openHere
	openHere = func(string) (bool, bool) { return true, false }
	t.Cleanup(func() { openHere = oldState })

	if err := removeImage(f.image); !errors.Is(err, errImageOpen) {
		t.Fatalf("got %v, want errImageOpen", err)
	}
	if _, err := os.Stat(f.image); err != nil {
		t.Fatal("the image was removed")
	}
}

func resolveAny(file string) (string, error) { return filepath.Clean(file), nil }

func mustWriteFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
