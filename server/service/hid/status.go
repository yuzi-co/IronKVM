package hid

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/inputcontrol"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	ModeNormal  = "normal"
	ModeHidOnly = "hid-only"
	ModeFlag    = "/sys/kernel/config/usb_gadget/g0/bcdDevice"

	ModeNormalScript  = "/kvmapp/system/init.d/S03usbdev"
	ModeHidOnlyScript = "/kvmapp/system/init.d/S03usbhid"

	USBDevScript = "/etc/init.d/S03usbdev"
)

var modeMap = map[string]string{
	"0x0510": ModeNormal,
	"0x0623": ModeHidOnly,
}

func (s *Service) GetHidMode(c *gin.Context) {
	var rsp proto.Response

	mode, err := GetMode()
	if err != nil {
		rsp.ErrRsp(c, -1, "get HID mode failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetHidModeRsp{
		Mode: mode,
	})
	log.Debugf("get hid mode: %s", mode)
}

func (s *Service) GetKeyboardLedStatus(c *gin.Context) {
	var rsp proto.Response
	status := GetKeyboardLedStatus()
	updatedAt := ""
	if !status.UpdatedAt.IsZero() {
		updatedAt = status.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}

	rsp.OkRspWithData(c, &proto.GetKeyboardLedStatusRsp{
		NumLock:    status.NumLock,
		CapsLock:   status.CapsLock,
		ScrollLock: status.ScrollLock,
		Known:      status.Known,
		UpdatedAt:  updatedAt,
	})
}

// GetHidStatus reports what the target is doing with each HID endpoint.
//
// The web UI needs this because the interesting failure is invisible from every
// other angle: the device node is present, the gadget is bound, and the target
// is simply not fetching from one endpoint. Absolute mouse reports then vanish
// while the keyboard keeps working, and the operator has no way to know that
// relative mouse mode would work.
func (s *Service) GetHidStatus(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, &proto.GetHidStatusRsp{
		Devices: GetHid().Status(),
	})
}

func (s *Service) SetHidMode(c *gin.Context) {
	var req proto.SetHidModeReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if req.Mode != ModeNormal && req.Mode != ModeHidOnly {
		rsp.ErrRsp(c, -2, "invalid arguments")
		return
	}

	if mode, _ := GetMode(); req.Mode == mode {
		rsp.OkRsp(c)
		return
	}

	srcScript := ModeNormalScript
	if req.Mode == ModeHidOnly {
		srcScript = ModeHidOnlyScript
	}

	if err := copyModeFile(srcScript); err != nil {
		rsp.ErrRsp(c, -3, "operation failed")
		return
	}

	if err := applyHidMode(req.Mode); err != nil {
		log.Errorf("failed to switch to HID mode %s: %s", req.Mode, err)
		rsp.ErrRsp(c, -4, "failed to apply hid mode")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("hid mode is now %s", req.Mode)
}

// usbDevCommand runs one action of the gadget script that is installed at
// USBDevScript. It is a variable so the switch can be tested without a device.
var usbDevCommand = func(action string) error {
	return exec.Command("sh", "-c", fmt.Sprintf("%s %s", USBDevScript, action)).Run()
}

// applyHidMode rebuilds the USB gadget from the script SetHidMode has just
// installed, so that changing mode does not need a reboot.
//
// This used to call reboot. A restart of this board costs about two minutes
// under the graceful stop, and it takes the video, the web session and any
// mounted image with it. The gadget needs none of that: `stop` writes an empty
// UDC and `start` builds the configuration again from whichever script is now
// at /etc/init.d/S03usbdev. It is the same sequence the virtual disk, network,
// console and speaker toggles already use.
//
// What makes it work is in the scripts rather than here. Both of them take
// their HID functions out of the configuration before they write the
// descriptors, because f_hid copies subclass, protocol, report_length and the
// report descriptor into the instance at the moment the link is made and
// refuses every write to them while it exists. A rebuild that skipped that
// would rebind carrying the descriptors of the mode it left, while the mode
// flag said otherwise.
func applyHidMode(mode string) error {
	// Same reason as ResetUSBPHY: switchGadget unbinds the UDC, and the
	// supervisor would otherwise see that as the link dying.
	NoteUSBGadgetMutated()

	h := GetHid()
	h.Lock()
	h.CloseNoLock()
	defer h.Unlock()

	err := switchGadget(mode, usbDevCommand, GetMode)

	// Reopen whatever the rebuild produced, even after it reported a failure.
	// Leaving the descriptors closed would take the keyboard away for a reason
	// that has nothing to do with which mode is installed, and the operator
	// would have no way back through the UI.
	if openErr := h.OpenNoLockWithRetry(hidReopenTimeout, hidReopenRetryDelay); openErr != nil && err == nil {
		err = fmt.Errorf("reopen the HID devices: %w", openErr)
	}

	return err
}

// switchGadget rebuilds the gadget and confirms it came back in the mode that
// was asked for. Everything it touches is passed in, so it can be tested off a
// device.
//
// The check at the end is not ceremony. The mode is read from bcdDevice on the
// live gadget, so a rebuild that ran but left the old descriptor in place would
// otherwise be reported to the operator as a success.
func switchGadget(mode string, run func(string) error, readMode func() (string, error)) error {
	if err := run("stop_start"); err != nil {
		return fmt.Errorf("rebuild the usb gadget: %w", err)
	}

	got, err := readMode()
	if err != nil {
		return fmt.Errorf("read the HID mode back: %w", err)
	}

	if got != mode {
		return fmt.Errorf("the gadget reports %s after switching to %s", got, mode)
	}

	return nil
}

func (s *Service) ResetHid(c *gin.Context) {
	var rsp proto.Response

	manual := s.newManualSession()
	defer manual.Close()
	reservation, err := manual.Reserve(c.Request.Context(), inputcontrol.ManualRelativeMouse, false, nil)
	if err != nil {
		log.Errorf("failed to acquire manual control for HID reset: %v", err)
		rsp.ErrRsp(c, -1, "HID control is busy")
		return
	}
	err = manual.Execute(ResetUSBPHY)
	reservation.Complete(err == nil)
	if err != nil {
		log.Errorf("failed to reset hid: %v", err)
		rsp.ErrRsp(c, -1, "failed to reset hid")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("reset hid success")
}

func (s *Service) RecoverUSB(c *gin.Context) {
	var rsp proto.Response

	if err := ResetUSBPHY(); err != nil {
		log.Errorf("failed to recover usb: %v", err)
		rsp.ErrRsp(c, -1, "failed to recover usb")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("recover usb success")
}

func ResetUSBPHY() error {
	// The supervisor must not read this operation's own disconnect as a fault
	// and start rebinding underneath it.
	NoteUSBGadgetMutated()

	h := GetHid()
	h.Lock()
	h.CloseNoLock()
	defer h.Unlock()

	if err := usbDevCommand("restart_phy"); err != nil {
		return fmt.Errorf("restart usb phy: %w", err)
	}

	if err := h.OpenNoLockWithRetry(hidReopenTimeout, hidReopenRetryDelay); err != nil {
		return fmt.Errorf("reopen HID devices after usb phy reset: %w", err)
	}

	return nil
}

func copyModeFile(srcScript string) error {
	// open the source file
	srcFile, err := os.Open(srcScript)
	if err != nil {
		log.Errorf("failed to open %s: %s", srcScript, err)
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		log.Errorf("failed to get %s info: %s", srcScript, err)
		return err
	}

	// create and copy to temporary file
	tmpFile, err := os.CreateTemp("/etc/init.d/", ".S03usbdev-")
	if err != nil {
		log.Errorf("failed to create temp %s: %s", USBDevScript, err)
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	log.Debugf("create temporary file: %s", tmpPath)

	if err := tmpFile.Chmod(srcInfo.Mode()); err != nil {
		_ = tmpFile.Close()
		log.Errorf("failed to set %s mode: %s", tmpPath, err)
		return err
	}

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		_ = tmpFile.Close()
		log.Errorf("failed to copy %s: %s", srcScript, err)
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		log.Errorf("failed to sync %s: %s", tmpPath, err)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		log.Errorf("failed to close %s: %s", tmpPath, err)
		return err
	}

	// replace the target file with the temporary file
	if err := os.Rename(tmpPath, USBDevScript); err != nil {
		log.Errorf("failed to rename %s: %s", tmpPath, err)
		return err
	}

	log.Debugf("copy %s to %s successful", srcScript, USBDevScript)
	return nil
}

func GetMode() (string, error) {
	data, err := os.ReadFile(ModeFlag)
	if err != nil {
		log.Errorf("failed to read %s: %s", ModeFlag, err)
		return "", err
	}

	key := strings.TrimSpace(string(data))
	mode, ok := modeMap[key]
	if !ok {
		log.Errorf("invalid mode flag: %s", key)
		return "", errors.New("invalid mode flag")
	}

	return mode, nil
}
