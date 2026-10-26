package urlmedia

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/storage"
)

var (
	// ErrUnavailable wraps the reason the board cannot mount from a URL.
	ErrUnavailable = errors.New("mounting from a URL is not available")
	errNoNBD       = errors.New("this kernel has no network block device (NBD) support; it needs the mainline kernel (slot B)")
	errNoDrive     = errors.New("the drive must be cdrom or disk")
	errNotMounted  = errors.New("the drive holds no image from a URL")
	// The kernel ends the connection itself only after its own timeout.
	errConnectionEnded = errors.New("the kernel ended the connection: eject the image and mount it again")
)

// source is an image read over the network.
type source interface {
	backend
	BytesRead() int64
	Close()
}

// device is a connected nbd device.
type device interface {
	Path() string
	// Disconnect ends the kernel's connection and waits for it to end.
	Disconnect() error
}

// The board's side, as variables so the tests can replace it.
var (
	nbdAvailable = kernelNBDAvailable
	openSource   = func(ctx context.Context, raw string) (source, error) {
		return Open(ctx, raw, Options{})
	}
	attachDevice     = attachNBD
	registerDevice   = storage.RegisterDevice
	unregisterDevice = storage.UnregisterDevice
	insertDevice     = storage.InsertDeviceInto
	ejectDrive       = storage.EjectDrive
	listDrives       = storage.Drives
)

// MountStatus is one drive's image from a URL. URL has no credentials or
// query.
type MountStatus struct {
	Drive     string     `json:"drive"`
	Device    string     `json:"device"`
	URL       string     `json:"url"`
	Host      string     `json:"host"`
	Size      int64      `json:"size"`
	BytesRead int64      `json:"bytesRead"`
	Cache     CacheStats `json:"cache"`
	// Error is the last read's failure, empty once a read works again.
	Error string    `json:"error,omitempty"`
	Since time.Time `json:"since"`
}

// Status says whether the board can mount from a URL, and what it serves.
type Status struct {
	Available bool          `json:"available"`
	Reason    string        `json:"reason,omitempty"`
	Mounts    []MountStatus `json:"mounts"`
}

type mount struct {
	drive, url, host string
	since            time.Time
	src              source
	cache            *Cache
	dev              device
	conn             net.Conn
	served           chan struct{}
	cancel           context.CancelFunc

	once     sync.Once
	errMu    sync.Mutex
	lastErr  string
	onRemove func(*mount)
}

// tracked notes the last read's failure for the status.
type tracked struct {
	backend
	m *mount
}

func (t tracked) ReadAt(ctx context.Context, p []byte, off int64) error {
	err := t.backend.ReadAt(ctx, p, off)
	t.m.errMu.Lock()
	if err != nil {
		t.m.lastErr = err.Error()
	} else {
		t.m.lastErr = ""
	}
	t.m.errMu.Unlock()
	return err
}

// teardown takes the mount down once: the kernel's connection, then the
// network. The gadget has let go of the device before it runs.
func (m *mount) teardown() {
	m.once.Do(func() {
		if err := m.dev.Disconnect(); err != nil {
			log.Warnf("urlmedia: %s", err)
		}
		// A read the kernel no longer waits for may still be retrying:
		// closing the cache ends it, so the server's loop can return.
		m.cancel()
		m.cache.Close()
		_ = m.conn.Close()
		<-m.served
		m.src.Close()
		unregisterDevice(m.dev.Path())
		m.onRemove(m)
		log.Infof("urlmedia: %s no longer serves %s", m.dev.Path(), m.host)
	})
}

func (m *mount) status() MountStatus {
	m.errMu.Lock()
	lastErr := m.lastErr
	m.errMu.Unlock()
	if e, ok := m.dev.(interface{ Ended() bool }); ok && e.Ended() {
		lastErr = errConnectionEnded.Error()
	}
	return MountStatus{
		Drive:     m.drive,
		Device:    m.dev.Path(),
		URL:       m.url,
		Host:      m.host,
		Size:      m.src.Size(),
		BytesRead: m.src.BytesRead(),
		Cache:     m.cache.Stats(),
		Error:     lastErr,
		Since:     m.since,
	}
}

// Service mounts images from URLs into the virtual drives.
type Service struct {
	op sync.Mutex // one mount or eject at a time

	mu     sync.Mutex
	mounts map[string]*mount

	availOnce sync.Once
	availErr  error
}

func New() *Service {
	return &Service{mounts: map[string]*mount{}}
}

var (
	defaultOnce sync.Once
	defaultSvc  *Service
)

// Default is the server's one service, which the HTTP API and MCP share.
func Default() *Service {
	defaultOnce.Do(func() { defaultSvc = New() })
	return defaultSvc
}

// available checks the kernel once: it does not gain NBD while running.
func (s *Service) available() error {
	s.availOnce.Do(func() { s.availErr = nbdAvailable() })
	return s.availErr
}

// Start ejects the nbd devices a previous run of the server left in a
// drive: nothing serves them any more, so the host gets only errors.
func (s *Service) Start() {
	drives, err := listDrives()
	if err != nil {
		return
	}
	for _, d := range drives {
		if strings.HasPrefix(d.File, "/dev/nbd") {
			log.Infof("urlmedia: eject %s from the %s drive, left by the last run", d.File, d.ID)
			if err := ejectDrive(d.ID); err != nil {
				log.Warnf("urlmedia: eject %s: %s", d.ID, err)
			}
		}
	}
}

// Status lists the mounts.
func (s *Service) Status() Status {
	st := Status{Available: true, Mounts: []MountStatus{}}
	if err := s.available(); err != nil {
		st.Available = false
		st.Reason = err.Error()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range []string{storage.DriveCdrom, storage.DriveDisk} {
		if m, ok := s.mounts[id]; ok {
			st.Mounts = append(st.Mounts, m.status())
		}
	}
	return st
}

// Mount serves the image at raw in the drive: the CD, or the disk,
// read-only. An image from a URL the drive already has is ejected first.
func (s *Service) Mount(ctx context.Context, drive, raw string) (MountStatus, error) {
	if drive != storage.DriveCdrom && drive != storage.DriveDisk {
		return MountStatus{}, errNoDrive
	}
	if err := s.available(); err != nil {
		return MountStatus{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}

	s.op.Lock()
	defer s.op.Unlock()

	if s.current(drive) != nil {
		if err := s.ejectLocked(drive); err != nil {
			return MountStatus{}, err
		}
	}

	src, err := openSource(ctx, raw)
	if err != nil {
		return MountStatus{}, err
	}
	dev, conn, err := attachDevice(src.Size())
	if err != nil {
		src.Close()
		return MountStatus{}, err
	}

	serveCtx, cancel := context.WithCancel(context.Background())
	m := &mount{
		drive:    drive,
		url:      Redact(raw),
		host:     HostOf(raw),
		since:    time.Now(),
		src:      src,
		dev:      dev,
		conn:     conn,
		served:   make(chan struct{}),
		cancel:   cancel,
		onRemove: s.remove,
	}
	m.cache = NewCache(tracked{src, m}, DefaultBlockSize, DefaultBlocks, DefaultReadAhead)
	go func() {
		defer close(m.served)
		if err := serveNBD(serveCtx, conn, m.cache); err != nil && serveCtx.Err() == nil {
			log.Warnf("urlmedia: %s: %s", dev.Path(), err)
		}
		// The kernel's receive thread ends when this end closes.
		_ = conn.Close()
	}()

	s.mu.Lock()
	s.mounts[drive] = m
	s.mu.Unlock()

	registerDevice(storage.Device{
		Path:     dev.Path(),
		Name:     m.host,
		Released: m.teardown,
	})
	if err := insertDevice(drive, dev.Path()); err != nil {
		m.teardown()
		return MountStatus{}, err
	}
	log.Infof("urlmedia: the %s drive serves %s through %s", drive, m.url, dev.Path())
	return m.status(), nil
}

// Eject takes the drive's image from a URL out: the gadget lets go of the
// device, then the device's connection ends, then the network's.
func (s *Service) Eject(drive string) error {
	s.op.Lock()
	defer s.op.Unlock()
	return s.ejectLocked(drive)
}

func (s *Service) ejectLocked(drive string) error {
	m := s.current(drive)
	if m == nil {
		return errNotMounted
	}
	if err := ejectDrive(drive); err != nil {
		return err
	}
	m.teardown()
	return nil
}

func (s *Service) current(drive string) *mount {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mounts[drive]
}

func (s *Service) remove(m *mount) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mounts[m.drive] == m {
		delete(s.mounts, m.drive)
	}
}
