package urlmedia

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/storage"
)

// events records the board-side calls in the order they happen.
type events struct {
	mu  sync.Mutex
	log []string
}

func (e *events) add(s string) {
	e.mu.Lock()
	e.log = append(e.log, s)
	e.mu.Unlock()
}

func (e *events) list() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string(nil), e.log...)
}

type fakeSource struct {
	*memSource
	ev *events
}

func (f fakeSource) BytesRead() int64 { return 0 }
func (f fakeSource) Close()           { f.ev.add("http-close") }

type fakeDevice struct {
	path   string
	ev     *events
	client net.Conn
}

func (d *fakeDevice) Path() string { return d.path }
func (d *fakeDevice) Disconnect() error {
	d.ev.add("nbd-disconnect " + d.path)
	return d.client.Close()
}

// fakeBoard replaces the board's side of a mount and records what it does.
type fakeBoard struct {
	ev         *events
	registered map[string]storage.Device
	insertErr  error
	ejectErr   error
	available  error
	opened     []string
	next       int
}

func newFakeBoard(t *testing.T) *fakeBoard {
	b := &fakeBoard{ev: &events{}, registered: map[string]storage.Device{}}
	var mu sync.Mutex
	saved := []any{nbdAvailable, openSource, attachDevice, registerDevice, unregisterDevice, insertDevice, ejectDrive}
	nbdAvailable = func() error { return b.available }
	openSource = func(_ context.Context, raw string) (source, error) {
		b.opened = append(b.opened, raw)
		return fakeSource{&memSource{data: testImage(1 << 20)}, b.ev}, nil
	}
	attachDevice = func(size int64) (device, net.Conn, error) {
		server, client := net.Pipe()
		mu.Lock()
		path := "/dev/nbd" + string(rune('0'+b.next))
		b.next++
		mu.Unlock()
		b.ev.add("nbd-attach " + path)
		return &fakeDevice{path: path, ev: b.ev, client: client}, server, nil
	}
	registerDevice = func(d storage.Device) {
		mu.Lock()
		b.registered[d.Path] = d
		mu.Unlock()
	}
	unregisterDevice = func(path string) {
		mu.Lock()
		delete(b.registered, path)
		mu.Unlock()
	}
	insertDevice = func(drive, path string) error {
		if b.insertErr != nil {
			return b.insertErr
		}
		b.ev.add("gadget-insert " + drive + " " + path)
		return nil
	}
	ejectDrive = func(drive string) error {
		if b.ejectErr != nil {
			return b.ejectErr
		}
		b.ev.add("gadget-eject " + drive)
		return nil
	}
	t.Cleanup(func() {
		nbdAvailable = saved[0].(func() error)
		openSource = saved[1].(func(context.Context, string) (source, error))
		attachDevice = saved[2].(func(int64) (device, net.Conn, error))
		registerDevice = saved[3].(func(storage.Device))
		unregisterDevice = saved[4].(func(string))
		insertDevice = saved[5].(func(string, string) error)
		ejectDrive = saved[6].(func(string) error)
	})
	return b
}

func (b *fakeBoard) device(path string) (storage.Device, bool) {
	d, ok := b.registered[path]
	return d, ok
}

func TestMountInsertsTheDeviceIntoTheDrive(t *testing.T) {
	b := newFakeBoard(t)
	s := New()

	st, err := s.Mount(context.Background(), storage.DriveCdrom, "https://user:pw@example.com/isos/x.iso?sig=1")
	if err != nil {
		t.Fatalf("Mount: %v", err)
	}
	want := []string{"nbd-attach /dev/nbd0", "gadget-insert cdrom /dev/nbd0"}
	if got := b.ev.list(); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	if _, ok := b.device("/dev/nbd0"); !ok {
		t.Fatal("the device was not registered with storage")
	}
	if st.Drive != storage.DriveCdrom || st.Device != "/dev/nbd0" || st.Size != 1<<20 {
		t.Fatalf("status = %+v", st)
	}
	if strings.Contains(st.URL, "pw") || strings.Contains(st.URL, "sig=1") || st.Host != "example.com" {
		t.Fatalf("status leaks the URL's secrets: %+v", st)
	}
	if mounts := s.Status().Mounts; len(mounts) != 1 {
		t.Fatalf("%d mounts listed, want 1", len(mounts))
	}
}

func TestEjectTearsDownGadgetThenNBDThenHTTP(t *testing.T) {
	b := newFakeBoard(t)
	s := New()
	if _, err := s.Mount(context.Background(), storage.DriveDisk, "http://example.com/x.img"); err != nil {
		t.Fatal(err)
	}

	if err := s.Eject(storage.DriveDisk); err != nil {
		t.Fatalf("Eject: %v", err)
	}
	want := []string{
		"nbd-attach /dev/nbd0",
		"gadget-insert disk /dev/nbd0",
		"gadget-eject disk",
		"nbd-disconnect /dev/nbd0",
		"http-close",
	}
	if got := b.ev.list(); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	if _, ok := b.device("/dev/nbd0"); ok {
		t.Fatal("the device is still registered after the eject")
	}
	if len(s.Status().Mounts) != 0 {
		t.Fatal("the mount is still listed after the eject")
	}
}

func TestEjectKeepsTheMountWhenTheGadgetRefuses(t *testing.T) {
	b := newFakeBoard(t)
	s := New()
	if _, err := s.Mount(context.Background(), storage.DriveCdrom, "http://example.com/x.iso"); err != nil {
		t.Fatal(err)
	}
	b.ejectErr = storage.ErrMediumLocked

	if err := s.Eject(storage.DriveCdrom); !errors.Is(err, storage.ErrMediumLocked) {
		t.Fatalf("err = %v, want ErrMediumLocked", err)
	}
	for _, e := range b.ev.list() {
		if strings.HasPrefix(e, "nbd-disconnect") || e == "http-close" {
			t.Fatalf("%q ran while the gadget still serves the device", e)
		}
	}
	if len(s.Status().Mounts) != 1 {
		t.Fatal("the mount is gone although the drive still holds it")
	}
}

func TestReleaseByTheDriveAPITearsDown(t *testing.T) {
	b := newFakeBoard(t)
	s := New()
	if _, err := s.Mount(context.Background(), storage.DriveCdrom, "http://example.com/x.iso"); err != nil {
		t.Fatal(err)
	}
	d, _ := b.device("/dev/nbd0")
	d.Released() // storage runs it after an eject or another insert

	deadline := time.Now().Add(2 * time.Second)
	for len(s.Status().Mounts) != 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	got := b.ev.list()
	if len(s.Status().Mounts) != 0 || got[len(got)-1] != "http-close" || got[len(got)-2] != "nbd-disconnect /dev/nbd0" {
		t.Fatalf("after a release: mounts %d, events %v", len(s.Status().Mounts), got)
	}
	d.Released() // a second release is harmless
}

func TestMountUndoesItselfWhenTheInsertFails(t *testing.T) {
	b := newFakeBoard(t)
	b.insertErr = errors.New("the kernel refused it")
	s := New()

	if _, err := s.Mount(context.Background(), storage.DriveDisk, "http://example.com/x.img"); err == nil {
		t.Fatal("Mount succeeded although the insert failed")
	}
	got := b.ev.list()
	want := []string{"nbd-attach /dev/nbd0", "nbd-disconnect /dev/nbd0", "http-close"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	if len(b.registered) != 0 || len(s.Status().Mounts) != 0 {
		t.Fatal("a failed mount left state behind")
	}
}

func TestMountNeedsNBD(t *testing.T) {
	b := newFakeBoard(t)
	b.available = errors.New("no nbd here")
	s := New()

	_, err := s.Mount(context.Background(), storage.DriveCdrom, "http://example.com/x.iso")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if len(b.opened) != 0 {
		t.Fatal("the URL was fetched although the kernel cannot serve it")
	}
	if st := s.Status(); st.Available || st.Reason == "" {
		t.Fatalf("status = %+v, want unavailable with a reason", st)
	}
}

func TestMountRefusesAnUnknownDrive(t *testing.T) {
	newFakeBoard(t)
	if _, err := New().Mount(context.Background(), "floppy", "http://example.com/x.iso"); err == nil {
		t.Fatal("Mount took an unknown drive")
	}
}

func TestMountReplacesTheDrivesURLMount(t *testing.T) {
	b := newFakeBoard(t)
	s := New()
	if _, err := s.Mount(context.Background(), storage.DriveCdrom, "http://example.com/a.iso"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Mount(context.Background(), storage.DriveCdrom, "http://example.com/b.iso"); err != nil {
		t.Fatal(err)
	}
	mounts := s.Status().Mounts
	if len(mounts) != 1 || !strings.HasSuffix(mounts[0].URL, "/b.iso") {
		t.Fatalf("mounts = %+v, want only b.iso", mounts)
	}
	// The first mount went with its drive: gadget first, then NBD and HTTP.
	got := strings.Join(b.ev.list(), ",")
	if !strings.Contains(got, "gadget-eject cdrom,nbd-disconnect /dev/nbd0,http-close,nbd-attach /dev/nbd1") {
		t.Fatalf("events = %s", got)
	}
}

// A server that stopped while serving a URL left /dev/nbdN in a drive with
// nothing behind it: the next run ejects it, and leaves images alone.
func TestStartEjectsDevicesTheLastRunLeft(t *testing.T) {
	b := newFakeBoard(t)
	saved := listDrives
	listDrives = func() ([]storage.Drive, error) {
		return []storage.Drive{
			{ID: storage.DriveDisk, File: "/data/rescue.img"},
			{ID: storage.DriveCdrom, File: "/dev/nbd3"},
		}, nil
	}
	t.Cleanup(func() { listDrives = saved })

	New().Start()
	if got := b.ev.list(); !reflect.DeepEqual(got, []string{"gadget-eject cdrom"}) {
		t.Fatalf("events = %v, want only the CD ejected", got)
	}
}
