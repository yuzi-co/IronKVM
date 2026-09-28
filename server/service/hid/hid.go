package hid

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"NanoKVM-Server/proto"

	log "github.com/sirupsen/logrus"
)

type Hid struct {
	g0                     *os.File
	g0Reader               *os.File
	g1                     *os.File
	g2                     *os.File
	ledReaderNotifyReader  *os.File
	ledReaderNotifyWriter  *os.File
	ledReaderNotifyReadFD  int
	ledReaderNotifyWriteFD int
	kbMutex                sync.Mutex
	mouseMutex             sync.Mutex
	ledReaderStartOnce     sync.Once

	// One health record per endpoint. They are separate because the failure
	// this reports is per-endpoint: a target can fetch keyboard and relative
	// mouse reports while ignoring the absolute mouse entirely.
	kbHealth  endpointHealth
	relHealth endpointHealth
	absHealth endpointHealth
	// absReportID is the ID put in front of every absolute pointer report, or
	// 0 for none. It is read from configfs each time /dev/hidg2 is opened,
	// because the gadget and this binary are deployed separately, and a
	// pointer that sends the wrong shape does not move: f_hid cuts a 7-byte
	// write to 6, and a host expecting an ID misreads a 6-byte report.
	// Guarded by mouseMutex.
	absReportID byte
	// absTouch says the gadget declares the touch screen under
	// TouchReportID. It is read with absReportID, for the same reason.
	// Guarded by mouseMutex.
	absTouch bool
	// absBuf holds one prefixed pointer report, so the prefix costs no
	// allocation on the hottest path in the server. Guarded by mouseMutex.
	absBuf [AbsoluteMouseReportLenWithID]byte
}

const (
	HID0 = "/dev/hidg0" // Keyboard
	HID1 = "/dev/hidg1" // Mouse (Relative Mode)
	HID2 = "/dev/hidg2" // Touchpad (Absolute Mode)
)

// Report lengths, one per endpoint. These were literals spread over seven
// files, and the literal did double duty: the mouse queue, the websocket and
// the cooldown all told a relative report from an absolute one by its length
// alone. That is workable while the numbers never move, and it is exactly what
// makes moving one dangerous.
//
// They must agree with report_length in kvmapp/system/init.d/S03usbdev and
// S03usbhid, which is what f_hid copies into the gadget. usb_report_test.go
// reads both scripts and fails the build if they disagree, the same way
// endpoints_shell_test.go holds the endpoint budget to its table.
const (
	// KeyboardReportLen is modifiers, one reserved byte, and six key codes.
	KeyboardReportLen = 8

	// RelativeMouseReportLen is buttons, X, Y, wheel, and horizontal wheel.
	// The fifth byte is AC Pan, from the Consumer page rather than Generic
	// Desktop, which is the usage a tilt wheel and a thumb wheel report.
	//
	// It is appended rather than inserted, so bytes 0 to 2 keep the USB boot
	// mouse layout. This interface claims the boot protocol, and a host reading
	// it in boot mode takes the first three bytes and ignores the rest.
	RelativeMouseReportLen = 5

	// AbsoluteMouseReportLen is buttons, a 16-bit X, a 16-bit Y, and wheel.
	AbsoluteMouseReportLen = 6

	// AbsoluteMouseReportLenWithID is what the host reads from the absolute
	// pointer in normal mode: one report ID byte, then the report.
	AbsoluteMouseReportLenWithID = AbsoluteMouseReportLen + 1
)

// Report IDs on the absolute pointer. Only S03usbdev declares them; the
// keyboard and the relative mouse claim the boot protocol, where a firmware
// reads reports without an ID, so they never carry one.
const (
	AbsolutePointerReportID byte = 1
	ConsumerReportID        byte = 2
	SystemReportID          byte = 3
)

const (
	hidWriteTimeout     = 50 * time.Millisecond
	hidWriteRetryDelay  = time.Millisecond
	hidReopenTimeout    = 2 * time.Second
	hidReopenRetryDelay = 100 * time.Millisecond
)

// absoluteReportLengthPath is where S03usbdev records the absolute pointer's
// report length. A variable so tests can point it at a scratch file.
var absoluteReportLengthPath = "/sys/kernel/config/usb_gadget/g0/functions/hid.GS2/report_length"

// readAbsoluteReportID returns the pointer's report ID if the gadget declares
// report IDs on hid.GS2, and 0 if it does not. The length is the tell: only
// the descriptor with IDs is AbsoluteMouseReportLenWithID long. A missing or
// unreadable file means no IDs, which is what an older gadget and hid-only
// mode both need.
func readAbsoluteReportID() byte {
	raw, err := os.ReadFile(absoluteReportLengthPath)
	if err != nil {
		return 0
	}
	if string(bytes.TrimSpace(raw)) == strconv.Itoa(AbsoluteMouseReportLenWithID) {
		return AbsolutePointerReportID
	}
	return 0
}

// RefreshAbsoluteReportID closes the pointer handle when the gadget's
// report_length no longer matches the ID read when the handle was opened, and
// reports whether it did. The next write reopens the handle and reads the ID
// again.
//
// Opening is not enough to keep the ID current. S03usbdev stop_start keeps the
// function directories, so /dev/hidg2 is not deleted and the handle stays
// valid while the descriptor under it changes. The server then sends plain
// reports to a gadget that declares IDs, and the host drops every one of them.
// The USB watchdog calls this on each poll.
func (h *Hid) RefreshAbsoluteReportID() bool {
	h.mouseMutex.Lock()
	defer h.mouseMutex.Unlock()

	if h.g2 == nil {
		return false
	}
	id := readAbsoluteReportID()
	touch := id != 0 && readAbsoluteTouch()
	if id == h.absReportID && touch == h.absTouch {
		return false
	}

	// The touch screen can come or go with report_length staying 7, so the
	// descriptor is compared as well as the length.
	log.Infof("%s: the pointer report ID changed from %d to %d and touch from %t to %t, reopening",
		HID2, h.absReportID, id, h.absTouch, touch)
	h.closeDeviceNoLock(h.absoluteMouseDevice(HID2))
	return true
}

type hidWriter interface {
	Write([]byte) (int, error)
}

// hidDevice points at one of the Hid handles rather than closing over it. A
// device is built for every HID report, and a pair of closures would be two
// heap allocations on the hottest path in the server.
type hidDevice struct {
	path   string
	name   string
	mu     *sync.Mutex
	file   **os.File
	health *endpointHealth
	// idPrefix asks writeHID to put Hid.absReportID in front of the report.
	// Only the pointer sets it; the keys on the same endpoint carry their own
	// ID in the report.
	idPrefix bool
	// needsID refuses the write when the open handle's gadget declares no
	// report IDs. The keys set it: without IDs their first byte would reach
	// the host as the pointer's buttons.
	needsID bool
	// needsTouch refuses the write when the open handle's gadget declares no
	// touch screen. The touch frames set it.
	needsTouch bool
}

func (d hidDevice) get() *os.File {
	return *d.file
}

func (d hidDevice) set(file *os.File) {
	*d.file = file
}

var (
	hid     *Hid
	hidOnce sync.Once
)

func GetHid() *Hid {
	hidOnce.Do(func() {
		hid = &Hid{}
	})
	return hid
}

func (h *Hid) Lock() {
	h.kbMutex.Lock()
	h.mouseMutex.Lock()
}

func (h *Hid) Unlock() {
	h.kbMutex.Unlock()
	h.mouseMutex.Unlock()
}

// The names are codes, not prose: the web UI translates them, and it needs a
// stable key to translate.
const (
	NameKeyboard      = "keyboard"
	NameRelativeMouse = "mouse-relative"
	NameAbsoluteMouse = "mouse-absolute"
)

func (h *Hid) keyboardDevice(path string) hidDevice {
	return hidDevice{path: path, name: NameKeyboard, mu: &h.kbMutex, file: &h.g0, health: &h.kbHealth}
}

func (h *Hid) relativeMouseDevice(path string) hidDevice {
	return hidDevice{path: path, name: NameRelativeMouse, mu: &h.mouseMutex, file: &h.g1, health: &h.relHealth}
}

func (h *Hid) absoluteMouseDevice(path string) hidDevice {
	return hidDevice{path: path, name: NameAbsoluteMouse, mu: &h.mouseMutex, file: &h.g2, health: &h.absHealth, idPrefix: true}
}

// extendedKeyDevice is the absolute pointer's endpoint seen by the Consumer and
// System Control keys. It shares the handle, the lock and the health record,
// and it writes reports as given.
func (h *Hid) extendedKeyDevice() hidDevice {
	return hidDevice{path: HID2, name: NameAbsoluteMouse, mu: &h.mouseMutex, file: &h.g2, health: &h.absHealth, needsID: true}
}

// Status reports what the target is doing with each endpoint. It takes none of
// the HID locks, so a caller cannot delay a keystroke by asking.
func (h *Hid) Status() []proto.HidDeviceStatus {
	now := time.Now()

	devices := h.devices()
	statuses := make([]proto.HidDeviceStatus, 0, len(devices))
	for _, device := range devices {
		status := device.health.snapshot(now)
		status.Name = device.name
		status.Path = device.path
		statuses = append(statuses, status)
	}

	return statuses
}

// WriteErrors reports how many writes to each endpoint have stalled or found
// the gadget detached since the server started. Like Status, it takes none of
// the HID locks. The Consumer and System Control keys share the absolute
// pointer's endpoint, so their failures count under mouse-absolute.
func (h *Hid) WriteErrors() []EndpointWriteErrors {
	devices := h.devices()
	counts := make([]EndpointWriteErrors, 0, len(devices))
	for _, device := range devices {
		counts = append(counts, EndpointWriteErrors{
			Name:     device.name,
			Stalled:  device.health.stalledWrites.Load(),
			Detached: device.health.detachedWrites.Load(),
		})
	}

	return counts
}

// linkFaultSince reports when the endpoints started failing with an error that
// means the gadget is not enumerated, and the zero time when none of them is.
//
// The supervisor in usb_watchdog.go reads this to break a tie it cannot break
// on its own. A wedged gadget and a perfectly healthy one in a switched-off
// computer both read "not attached" at the controller, and only one of them
// wants repairing. Somebody driving this KVM while every report is refused is
// not an idle host, so this evidence shortens how long the supervisor waits.
//
// The earliest of the endpoints wins. They fail together when the cause is the
// link, and the earliest is when the link went, which is the age the supervisor
// wants to compare against.
func (h *Hid) linkFaultSince() time.Time {
	var earliest time.Time

	for _, device := range h.devices() {
		since := device.health.linkFaultSince()
		if since.IsZero() {
			continue
		}
		if earliest.IsZero() || since.Before(earliest) {
			earliest = since
		}
	}

	return earliest
}

func (h *Hid) devices() []hidDevice {
	return []hidDevice{
		h.keyboardDevice(HID0),
		h.relativeMouseDevice(HID1),
		h.absoluteMouseDevice(HID2),
	}
}

func (h *Hid) OpenNoLock() error {
	h.CloseNoLock()

	var errs []error
	for _, device := range h.devices() {
		if err := h.openDeviceNoLock(device); err != nil {
			log.Errorf("open %s failed: %s", device.path, err)
			errs = append(errs, err)
		}
	}

	if h.g0 != nil {
		if err := h.openKeyboardLedReaderNoLock(); err != nil {
			log.Errorf("open keyboard LED reader failed: %s", err)
			errs = append(errs, err)
		} else {
			h.startKeyboardLedReader()
		}
	}

	return errors.Join(errs...)
}

func (h *Hid) OpenNoLockWithRetry(timeout, delay time.Duration) error {
	return openNoLockWithRetry(h.OpenNoLock, timeout, delay)
}

func openNoLockWithRetry(open func() error, timeout, delay time.Duration) error {
	if timeout <= 0 {
		return open()
	}
	if delay <= 0 {
		delay = hidReopenRetryDelay
	}

	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := open(); err != nil {
			lastErr = err
		} else {
			return nil
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if remaining > delay {
			remaining = delay
		}
		time.Sleep(remaining)
	}

	return fmt.Errorf("open HID devices within %s: %w", timeout, lastErr)
}

func (h *Hid) CloseNoLock() {
	h.closeKeyboardLedReaderNoLock()

	for _, device := range h.devices() {
		h.closeDeviceNoLock(device)
	}
}

// ForgetAcceptingNoLock clears every endpoint's record of having been polled.
// Call it with the HID locks held, around anything that makes the host
// enumerate the gadget afresh: a rebuild, a rebind or a PHY reset. The host
// decides again then which endpoints it polls, and a stall that follows is not
// evidence that anything stopped.
//
// It is not part of CloseNoLock, because every websocket client reopens the
// descriptors when it connects, and a browser reconnecting must not hide a
// real stall.
func (h *Hid) ForgetAcceptingNoLock() {
	for _, device := range h.devices() {
		device.health.forgetAccepting()
	}
}

func (h *Hid) openDeviceNoLock(device hidDevice) error {
	if device.get() != nil {
		return nil
	}

	file, err := os.OpenFile(device.path, os.O_WRONLY|syscall.O_NONBLOCK, 0o666)
	if err != nil {
		return fmt.Errorf("%s: %w", device.path, err)
	}

	device.set(file)
	if device.file == &h.g2 {
		h.absReportID = readAbsoluteReportID()
		h.absTouch = h.absReportID != 0 && readAbsoluteTouch()
	}
	return nil
}

func (h *Hid) closeDeviceNoLock(device hidDevice) {
	file := device.get()
	if file == nil {
		return
	}

	device.set(nil)
	if err := file.Close(); err != nil {
		log.Debugf("close %s failed: %s", device.path, err)
	}
}

func (h *Hid) openKeyboardLedReaderNoLock() error {
	if h.g0Reader != nil {
		return nil
	}

	if err := h.ensureKeyboardLedReaderNotifierNoLock(); err != nil {
		return err
	}

	// Keep this descriptor blocking. The LED reader waits for either an output
	// report or a lifecycle notification, so an idle host does not cause a
	// periodic EAGAIN retry loop.
	file, err := os.OpenFile(HID0, os.O_RDONLY, 0o666)
	if err != nil {
		return fmt.Errorf("%s: %w", HID0, err)
	}

	h.g0Reader = file
	h.notifyKeyboardLedReaderNoLock()
	return nil
}

func (h *Hid) closeKeyboardLedReaderNoLock() {
	if h.g0Reader == nil {
		return
	}

	file := h.g0Reader
	h.g0Reader = nil
	h.notifyKeyboardLedReaderNoLock()
	if err := file.Close(); err != nil {
		log.Debugf("close keyboard LED reader failed: %s", err)
	}
}

// writeWithTimeout bounds how long callers hold HID locks when writing to a
// nonblocking descriptor. EAGAIN means the host is not accepting HID reports
// yet, so retry until the caller's deadline expires.
func writeWithTimeout(writer hidWriter, data []byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		n, err := writer.Write(data)
		if err == nil {
			if n != len(data) {
				return io.ErrShortWrite
			}
			return nil
		}

		if n != 0 {
			return io.ErrShortWrite
		}
		if !isRetryableWriteError(err) {
			return err
		}

		remaining := time.Until(deadline)
		if timeout <= 0 || remaining <= 0 {
			return os.ErrDeadlineExceeded
		}
		if remaining > hidWriteRetryDelay {
			remaining = hidWriteRetryDelay
		}
		time.Sleep(remaining)
	}
}

func isRetryableWriteError(err error) bool {
	return errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK)
}

// hidFileWasDeleted reports whether the device node behind an open handle has
// been removed.
//
// Rebuilding the USB gadget - mounting an image, switching HID mode - deletes
// /dev/hidg*. A handle opened before that keeps accepting writes that reach
// nothing, so keyboard and mouse stop working with no error to say why. The
// kernel marks the link in /proc/self/fd for exactly this case.
//
// This runs on every report, so it builds the path and reads the link without
// allocating. fmt.Sprintf and os.Readlink together cost four allocations and
// 224 bytes per call, measured, which is small next to a syscall and still
// pointless on a board with one core and no memory to spare. The syscall and
// the test for " (deleted)" are unchanged: what the kernel is asked, and what
// counts as an answer, are the same as before.
const hidDeletedSuffix = " (deleted)"

func hidFileWasDeleted(file *os.File) bool {
	// Long enough for "/proc/self/fd/" and any descriptor number.
	var pathBuf [32]byte
	path := append(pathBuf[:0], "/proc/self/fd/"...)
	path = strconv.AppendInt(path, int64(file.Fd()), 10)

	// Long enough for the device paths this reads, and a short read only ever
	// means the answer is no.
	var target [128]byte

	n, err := syscall.Readlink(string(path), target[:])
	if err != nil || n <= 0 {
		return false
	}

	return bytes.HasSuffix(target[:n], []byte(hidDeletedSuffix))
}

func (h *Hid) Open() {
	h.kbMutex.Lock()
	defer h.kbMutex.Unlock()
	h.mouseMutex.Lock()
	defer h.mouseMutex.Unlock()

	h.OpenNoLock()
}

func (h *Hid) Close() {
	h.kbMutex.Lock()
	defer h.kbMutex.Unlock()
	h.mouseMutex.Lock()
	defer h.mouseMutex.Unlock()

	h.CloseNoLock()
}

func (h *Hid) WriteHid0(data []byte) {
	if err := h.WriteKeyboardReport(data); err != nil {
		reportWriteFailure("keyboard HID write failed", err)
	}
}

func (h *Hid) WriteHid1(data []byte) {
	if err := h.WriteRelativeMouseReport(data); err != nil {
		reportWriteFailure("relative mouse HID write failed", err)
	}
}

func (h *Hid) WriteHid2(data []byte) {
	if err := h.WriteAbsoluteMouseReport(data); err != nil {
		reportWriteFailure("absolute mouse HID write failed", err)
	}
}

func (h *Hid) WriteKeyboardReport(data []byte) error {
	if len(data) != KeyboardReportLen {
		return fmt.Errorf("invalid keyboard report length: %d", len(data))
	}
	return h.writeHID(h.keyboardDevice(HID0), data)
}

func (h *Hid) WriteRelativeMouseReport(data []byte) error {
	if len(data) != RelativeMouseReportLen {
		return fmt.Errorf("invalid relative mouse report length: %d", len(data))
	}
	return h.writeHID(h.relativeMouseDevice(HID1), data)
}

func (h *Hid) WriteAbsoluteMouseReport(data []byte) error {
	if len(data) != AbsoluteMouseReportLen {
		return fmt.Errorf("invalid absolute mouse report length: %d", len(data))
	}
	return h.writeHID(h.absoluteMouseDevice(HID2), data)
}

func (h *Hid) writeHIDReport(device hidDevice, data []byte) bool {
	if err := h.writeHID(device, data); err != nil {
		reportWriteFailure("write to "+device.path+" failed", err)
		return false
	}
	return true
}

// errRepeatedFailure marks a write failure that repeats one already reported for
// the same endpoint. A stalled endpoint fails every report - about twenty a
// second while the mouse moves - and each call site would otherwise print the
// same line every time. Call sites pass their failures through
// reportWriteFailure, which drops the repeats and keeps the first.
var errRepeatedFailure = errors.New("repeated failure")

// reportWriteFailure logs a failed HID operation unless the same endpoint has
// already reported this failure. The context describes the operation, because
// which write failed is worth knowing and the error alone does not say.
func reportWriteFailure(context string, err error) {
	if errors.Is(err, errRepeatedFailure) {
		return
	}

	log.Errorf("%s: %s", context, err)
}

func (h *Hid) writeHID(device hidDevice, data []byte) error {
	device.mu.Lock()
	defer device.mu.Unlock()

	err := h.writeHIDLocked(device, data)
	if errors.Is(err, errExtendedKeysUnavailable) || errors.Is(err, errTouchUnavailable) {
		// Nothing reached the endpoint, so it says nothing about its health.
		return err
	}
	return device.note(err)
}

// note records what one write did to the endpoint and decides whether the
// failure is worth reporting. It logs the two transitions no call site can see:
// an endpoint that has stopped accepting reports, and one that has started
// again.
func (d hidDevice) note(err error) error {
	transition := d.health.record(err, time.Now())

	if err == nil {
		if transition.Changed && transition.From != hidStateUnknown {
			log.Infof("%s is accepting reports again", d.path)
		}
		return nil
	}

	if !transition.Changed {
		return fmt.Errorf("%w: %w", errRepeatedFailure, err)
	}

	if transition.To == hidStateDetached {
		// Different from a stall, and worth different words. The target is not
		// refusing this endpoint: there is no link at all, so nothing typed or
		// clicked is reaching the host. Say so once, and leave the repair to
		// the supervisor, which is watching the controller and can tell an
		// unplugged cable from a wedged gadget.
		log.Errorf("%s: the gadget is not enumerated, so every report is lost. "+
			"Either no host is attached, or the USB link has failed", d.path)
	}

	if transition.To == hidStateStalled {
		// The failure that looks like nothing. The device node is present, the
		// gadget is bound, and the target is simply not fetching from this
		// endpoint - so every report to it is dropped while the others work.
		//
		// The message names the endpoint and stops there. Why a target stops
		// collecting from one interface is not knowable from this side, and a
		// guess printed as a fact would send the operator the wrong way.
		log.Errorf("%s: the target is not fetching reports from this endpoint, so reports to it are lost. "+
			"The other HID endpoints are unaffected. A USB reset, or a mouse mode that uses a different "+
			"endpoint, may recover it", d.path)
	}

	return err
}

// writeReport puts one report on the wire. It is a variable so tests can drive
// the outcomes writeHID has to react to: an endpoint the target is not fetching
// from cannot be built out of an ordinary file, and a pipe stands in for it only
// as long as the runtime keeps the descriptor pollable, which it does not
// promise.
var writeReport = func(path string, file *os.File, data []byte, timeout time.Duration) error {
	// Two mechanisms for one deadline, because only one of them is available
	// at a time. A pollable descriptor honours the deadline directly; one the
	// runtime declined to poll returns EAGAIN instead, and writeWithTimeout
	// bounds the retries.
	if err := file.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		log.Debugf("set write deadline for %s failed: %s", path, err)
	}

	return writeWithTimeout(file, data, timeout)
}

func (h *Hid) writeHIDLocked(device hidDevice, data []byte) error {
	// Drop a handle whose device node is gone before writing into it.
	if file := device.get(); file != nil && hidFileWasDeleted(file) {
		log.Debugf("%s was rebuilt underneath us, reopening", device.path)
		h.closeDeviceNoLock(device)
	}

	if err := h.openDeviceNoLock(device); err != nil {
		return err
	}
	if device.path == HID0 {
		if err := h.openKeyboardLedReaderNoLock(); err != nil {
			log.Debugf("open keyboard LED reader failed: %s", err)
		} else {
			h.startKeyboardLedReader()
		}
	}

	file := device.get()
	if file == nil {
		return fmt.Errorf("%s: hid handle is nil", device.path)
	}

	// absReportID was read when this handle was opened, so this check and
	// the write below see the same gadget: a mode switch that rebuilt it
	// deleted the node, and the reopen above read the ID again.
	if device.needsID && h.absReportID == 0 {
		return errExtendedKeysUnavailable
	}
	if device.needsTouch && !h.absTouch {
		return errTouchUnavailable
	}
	if device.idPrefix && h.absReportID != 0 {
		if len(data) != AbsoluteMouseReportLen {
			return fmt.Errorf("%s: pointer report of %d bytes, want %d", device.path, len(data), AbsoluteMouseReportLen)
		}
		h.absBuf[0] = h.absReportID
		copy(h.absBuf[1:], data)
		data = h.absBuf[:]
	}
	if err := writeReport(device.path, file, data, hidWriteTimeout); err != nil {
		// The LED reader shares the keyboard endpoint's descriptor, so a failed
		// keyboard write invalidates it too. Drop it before the device handle.
		if device.path == HID0 {
			h.closeKeyboardLedReaderNoLock()
		}

		h.closeDeviceNoLock(device)
		switch {
		case errors.Is(err, os.ErrClosed):
			return fmt.Errorf("hid already closed: %w", err)
		case errors.Is(err, os.ErrDeadlineExceeded):
			return fmt.Errorf("timeout after %s: %w", hidWriteTimeout, err)
		default:
			return err
		}
	}

	traceHIDWrite(device.path, data)
	return nil
}

// traceHIDWrite logs a report only when someone is listening. logrus skips the
// formatting on its own, but the variadic call still allocates a slice and
// boxes both arguments, once per report.
func traceHIDWrite(path string, data []byte) {
	if !log.IsLevelEnabled(log.DebugLevel) {
		return
	}

	log.Debugf("write to %s: %v", path, data)
}
