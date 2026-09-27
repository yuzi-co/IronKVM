package hid

import (
	"context"
	"errors"
	"fmt"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/inputcontrol"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Consumer and System Control keys share the absolute pointer's endpoint under
// their own report IDs. S03usbdev declares them; hid-only mode and an older
// gadget do not, and there a key report would reach the host as a garbled
// pointer report, so nothing is written.

const (
	extendedKeyHold = 50 * time.Millisecond

	consumerUsageMax = 0x3ff
	systemUsageMin   = 0x81
	systemUsageMax   = 0xb7
)

var errExtendedKeysUnavailable = errors.New("the USB gadget has no Consumer or System Control reports")

// extendedKeyReports builds the press and release reports for one key.
func extendedKeyReports(page string, usage int) (press, release []byte, err error) {
	switch page {
	case "consumer":
		if usage < 1 || usage > consumerUsageMax {
			return nil, nil, fmt.Errorf("consumer usage 0x%x is outside 0x1 to 0x%x", usage, consumerUsageMax)
		}
		return []byte{ConsumerReportID, byte(usage), byte(usage >> 8)}, []byte{ConsumerReportID, 0, 0}, nil
	case "system":
		if usage < systemUsageMin || usage > systemUsageMax {
			return nil, nil, fmt.Errorf("system usage 0x%x is outside 0x%x to 0x%x", usage, systemUsageMin, systemUsageMax)
		}
		return []byte{SystemReportID, byte(usage)}, []byte{SystemReportID, 0}, nil
	default:
		return nil, nil, fmt.Errorf("unknown key page %q", page)
	}
}

// ExtendedKeysAvailable reports whether the gadget declares the key reports.
// It reads configfs rather than the state kept for the open handle, because a
// caller may ask before anything has opened /dev/hidg2. It answers the UI and
// the route's early refusal; the write itself decides again under the lock.
func (h *Hid) ExtendedKeysAvailable() bool {
	return readAbsoluteReportID() != 0
}

// WriteExtendedKeyReport writes one Consumer or System Control report as given.
func (h *Hid) WriteExtendedKeyReport(report []byte) error {
	switch {
	case len(report) == 3 && report[0] == ConsumerReportID:
	case len(report) == 2 && report[0] == SystemReportID:
	default:
		return fmt.Errorf("not a Consumer or System Control report: % x", report)
	}
	return h.writeHID(h.extendedKeyDevice(), report)
}

// pressExtendedKey writes press, holds, and writes release. The release is
// written whatever happened before it, and tried a second time if it fails: a
// key left down on the host repeats, and for Power Down or Sleep a held key is
// worse than a lost one.
func pressExtendedKey(ctx context.Context, write func([]byte) error, press, release []byte, hold time.Duration) error {
	err := write(press)
	if err == nil {
		err = sleepPasteContext(ctx, hold)
	}
	releaseErr := write(release)
	if releaseErr != nil {
		releaseErr = write(release)
	}
	if err == nil {
		err = releaseErr
	}
	return err
}

// SendKey presses and releases one Consumer or System Control key.
func (s *Service) SendKey(c *gin.Context) {
	var req proto.SendHidKeyReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	press, release, err := extendedKeyReports(req.Page, req.Usage)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}

	if !s.hid.ExtendedKeysAvailable() {
		rsp.ErrRsp(c, -2, "extended keys are not available in this USB mode")
		return
	}

	manual := s.newManualSession()
	defer manual.Close()
	reservation, err := manual.Reserve(c.Request.Context(), inputcontrol.ManualAbsoluteMouse, false, nil)
	if err != nil {
		log.Errorf("hid key failed to acquire HID control: %v", err)
		rsp.ErrRsp(c, -3, "HID control is busy")
		return
	}

	write := func(report []byte) error {
		return manual.Execute(func() error {
			return s.hid.WriteExtendedKeyReport(report)
		})
	}

	err = pressExtendedKey(c.Request.Context(), write, press, release, extendedKeyHold)
	reservation.Complete(err == nil)
	if err != nil {
		reportWriteFailure("hid key failed", err)
		rsp.ErrRsp(c, -3, "HID key failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("hid key %s 0x%x sent", req.Page, req.Usage)
}
