package hid

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

func (h *Hid) Mouse(queue <-chan []byte) {
	legacy := make(chan QueuedReport)
	go func() {
		defer close(legacy)
		for report := range queue {
			legacy <- QueuedReport{Data: report}
		}
	}()
	h.MouseReports(legacy)
}

func (h *Hid) MouseReports(queue <-chan QueuedReport) {
	h.mouseReports(queue, HID1, HID2)
}

func (h *Hid) mouseReports(queue <-chan QueuedReport, relativePath string, absolutePath string) {
	var execute func(func() error) error
	var resetRelativeMouse func()
	var resetAbsoluteMouse func()
	relativeButtonsActive := false
	absoluteButtonsActive := false
	absoluteReleaseReport := absoluteMouseReleaseReport(nil)
	// heldContacts are the touch contacts the host may still see down. They
	// are lifted wherever held buttons are released.
	var heldContacts []TouchContact
	liftTouches := func() error {
		for _, frame := range touchLiftFrames(heldContacts) {
			err := runCleanup(execute, func() error {
				return h.writeTouchFrame(h.touchDeviceAt(absolutePath), frame)
			})
			if errors.Is(err, errTouchUnavailable) {
				// The gadget was rebuilt without the touch screen, and the
				// host lost the contacts with it.
				break
			}
			if err != nil {
				return err
			}
		}
		heldContacts = nil
		if resetAbsoluteMouse != nil {
			resetAbsoluteMouse()
		}
		return nil
	}
	defer func() {
		if len(heldContacts) > 0 {
			if err := liftTouches(); err != nil {
				reportWriteFailure("lift touch contacts on queue close failed", err)
			}
		}
		if relativeButtonsActive {
			if err := runCleanup(execute, func() error {
				return h.writeHID(h.relativeMouseDevice(relativePath), relativeMouseReleaseReport())
			}); err != nil {
				reportWriteFailure("release relative mouse on queue close failed", err)
			} else if resetRelativeMouse != nil {
				resetRelativeMouse()
			}
		}
		if absoluteButtonsActive {
			if err := runCleanup(execute, func() error {
				return h.writeHID(h.absoluteMouseDevice(absolutePath), absoluteReleaseReport)
			}); err != nil {
				reportWriteFailure("release absolute mouse on queue close failed", err)
			} else if resetAbsoluteMouse != nil {
				resetAbsoluteMouse()
			}
		}
	}()

	for event := range queue {
		execute = event.Execute
		resetRelativeMouse = event.ResetRelativeMouse
		resetAbsoluteMouse = event.ResetAbsoluteMouse

		cleanupFailure := func(writeErr error) {
			reportWriteFailure("mouse HID write failed", writeErr)
			if dropped := drainHIDQueue(queue); dropped > 0 {
				log.Debugf("dropped %d stale mouse HID reports after write failure", dropped)
			}

			if len(event.Data) == RelativeMouseReportLen && event.Data[0] != 0 {
				relativeButtonsActive = true
			}
			if len(event.Data) == AbsoluteMouseReportLen && event.Data[0] != 0 {
				absoluteButtonsActive = true
				absoluteReleaseReport = absoluteMouseReleaseReport(event.Data)
			}
			if event.Touch != nil {
				heldContacts = heldTouches(heldContacts, event.Touch)
			}
			if len(heldContacts) > 0 {
				if err := liftTouches(); err != nil {
					reportWriteFailure("lift touch contacts after write failure failed", err)
				}
			}

			if relativeButtonsActive || len(event.Data) == RelativeMouseReportLen {
				if err := runCleanup(execute, func() error {
					return h.writeHID(h.relativeMouseDevice(relativePath), relativeMouseReleaseReport())
				}); err != nil {
					reportWriteFailure("release relative mouse after write failure failed", err)
				} else {
					relativeButtonsActive = false
					if resetRelativeMouse != nil {
						resetRelativeMouse()
					}
				}
			}

			if absoluteButtonsActive || len(event.Data) == 6 {
				releaseReport := absoluteReleaseReport
				if len(event.Data) == 6 {
					releaseReport = absoluteMouseReleaseReport(event.Data)
				}
				if err := runCleanup(execute, func() error {
					return h.writeHID(h.absoluteMouseDevice(absolutePath), releaseReport)
				}); err != nil {
					reportWriteFailure("release absolute mouse after write failure failed", err)
				} else {
					absoluteButtonsActive = false
					if resetAbsoluteMouse != nil {
						resetAbsoluteMouse()
					}
				}
			}

			event.complete(false)
		}

		// A touch frame goes to the absolute pointer's endpoint like an
		// absolute report, so buttons held on either mouse are released
		// first, the same way a switch between the two mice does.
		if event.Touch != nil {
			if relativeButtonsActive {
				if err := runCleanup(execute, func() error {
					return h.writeHID(h.relativeMouseDevice(relativePath), relativeMouseReleaseReport())
				}); err != nil {
					cleanupFailure(err)
					continue
				}
				relativeButtonsActive = false
				if resetRelativeMouse != nil {
					resetRelativeMouse()
				}
			}
			if absoluteButtonsActive {
				if err := runCleanup(execute, func() error {
					return h.writeHID(h.absoluteMouseDevice(absolutePath), absoluteReleaseReport)
				}); err != nil {
					cleanupFailure(err)
					continue
				}
				absoluteButtonsActive = false
			}

			if err := event.run(func() error {
				return h.writeTouchFrame(h.touchDeviceAt(absolutePath), event.Touch)
			}); err != nil {
				cleanupFailure(err)
				continue
			}
			heldContacts = heldTouches(heldContacts, event.Touch)
			event.complete(true)
			continue
		}

		// A mouse report after touch lifts the fingers first. A finger left
		// down on the host would hold its touch while the pointer moves.
		if len(heldContacts) > 0 && isMouseReportLen(len(event.Data)) {
			if err := liftTouches(); err != nil {
				cleanupFailure(err)
				continue
			}
		}

		switch len(event.Data) {
		case RelativeMouseReportLen:
			if absoluteButtonsActive {
				if err := runCleanup(execute, func() error {
					return h.writeHID(h.absoluteMouseDevice(absolutePath), absoluteReleaseReport)
				}); err != nil {
					cleanupFailure(err)
					continue
				}
				absoluteButtonsActive = false
				if resetAbsoluteMouse != nil {
					resetAbsoluteMouse()
				}
			}

			if err := event.run(func() error {
				return h.writeHID(h.relativeMouseDevice(relativePath), event.Data)
			}); err != nil {
				cleanupFailure(err)
				continue
			}
			relativeButtonsActive = event.Data[0] != 0
			event.complete(true)
		case AbsoluteMouseReportLen:
			if relativeButtonsActive {
				if err := runCleanup(execute, func() error {
					return h.writeHID(h.relativeMouseDevice(relativePath), relativeMouseReleaseReport())
				}); err != nil {
					cleanupFailure(err)
					continue
				}
				relativeButtonsActive = false
				if resetRelativeMouse != nil {
					resetRelativeMouse()
				}
			}

			if err := event.run(func() error {
				return h.writeHID(h.absoluteMouseDevice(absolutePath), event.Data)
			}); err != nil {
				cleanupFailure(err)
				continue
			}
			absoluteReleaseReport = absoluteMouseReleaseReport(event.Data)
			absoluteButtonsActive = event.Data[0] != 0
			event.complete(true)
		default:
			event.complete(false)
			log.Debugf("invalid mouse event: %v", event.Data)
		}
	}
}

func isMouseReportLen(length int) bool {
	return length == RelativeMouseReportLen || length == AbsoluteMouseReportLen
}

func relativeMouseReleaseReport() []byte {
	return make([]byte, RelativeMouseReportLen)
}

func absoluteMouseReleaseReport(positionReport []byte) []byte {
	report := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if len(positionReport) >= 5 {
		copy(report[1:5], positionReport[1:5])
	}
	return report
}
