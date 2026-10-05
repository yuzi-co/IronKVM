package hid

// usb_events.go hears every change of the controller's state attribute, not
// only the ones a two-second sample happens to land on.
//
// A host that resets the port takes the gadget from configured through "not
// attached", "default" and "addressed" back to configured in well under a
// second. In trial 30 on the mainline slot the host did that 22 times in half an
// hour, and the watchdog's samples caught 4 of them (#70). The gadget core
// calls sysfs_notify on the state attribute after every change, so a poll for
// POLLPRI on it wakes for each one. The notification is sent from a work item,
// so several changes close together can arrive as one wakeup, and the state
// read then is already the last of them: a wakeup that reads configured right
// after configured is a whole reset seen at once.
//
// The events only feed usbWatchdog.track, which counts and logs
// re-enumerations. Recovery still decides on the samples, so this cannot make
// the watchdog act on something a sample would not see.

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// usbLinkEvent is the link as read right after one wakeup.
type usbLinkEvent struct {
	at   time.Time
	link usbLink
}

// usbLinkEvents carries wakeups to the watchdog, which drains it on every
// poll. A reset makes at most four; 64 holds a burst of them between polls.
var (
	usbLinkEvents        = make(chan usbLinkEvent, 64)
	usbLinkEventsDropped atomic.Uint64
)

func sendUSBLinkEvent(ev usbLinkEvent) {
	select {
	case usbLinkEvents <- ev:
	default:
		usbLinkEventsDropped.Add(1)
	}
}

// watchUSBLinkEvents runs for the life of the server. When the controller goes
// away (restart_phy unbinds it from its driver) the wait fails, and it opens
// the attribute again once there is one.
func watchUSBLinkEvents() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("usb watchdog: the link event reader panicked, samples only from now: %v", r)
		}
	}()

	lastErr := ""
	for {
		err := waitUSBLinkEvents()
		if msg := err.Error(); msg != lastErr {
			log.Infof("usb watchdog: link events: %s; retrying", msg)
			lastErr = msg
		}
		time.Sleep(usbPollInterval)
	}
}

// waitUSBLinkEvents returns only with an error.
func waitUSBLinkEvents() error {
	udc, err := firstUDC()
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(udcClassDir, udc, "state"))
	if err != nil {
		return err
	}
	defer f.Close()

	// kernfs reports a change to an open file only after that file has read
	// the attribute once, so this read arms the first wait.
	if _, err := readStateFile(f); err != nil {
		return err
	}

	fds := []unix.PollFd{{Fd: int32(f.Fd()), Events: unix.POLLPRI | unix.POLLERR}}
	for {
		if _, err := unix.Poll(fds, -1); err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return err
		}

		state, err := readStateFile(f)
		if err != nil {
			return err
		}
		speed, _ := readUDCAttr(udc, "current_speed")
		sendUSBLinkEvent(usbLinkEvent{at: usbWatchdogNow(), link: usbLink{State: state, Speed: speed}})
	}
}

func readStateFile(f *os.File) (string, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	buf := make([]byte, 64)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(string(buf[:n])), nil
}

// firstUDC names the controller readUSBLink would read.
func firstUDC() (string, error) {
	entries, err := udcReadDir(udcClassDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if _, err := readUDCAttr(entry.Name(), "state"); err == nil {
			return entry.Name(), nil
		}
	}
	return "", os.ErrNotExist
}
