package vm

import (
	"errors"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
)

const (
	virtualDisk  = "/boot/usb.disk0"
	virtualAudio = "/boot/usb.uac"
)

var (
	mountConsoleCommands = []string{
		"touch /boot/usb.acm",
		usbGadgetStop,
		usbGadgetStart,
	}

	// The function directory stays. `rmdir functions/acm.GS0` blocks forever:
	// /etc/inittab respawns a getty on /dev/ttyGS0, so the character device
	// always has a holder, and recovering from that wedge needs a full teardown
	// of the gadget. Remove the config symlink and nothing else - the same rule
	// the audio function follows.
	//
	// The marker goes with rm -f. /boot is vfat on the SD card and turns
	// read-only after an error; the loop that runs these stops at the first
	// nonzero exit, so a bare rm that failed would skip `start` and leave the
	// board in host mode with no HID and no way back from the UI.
	unmountConsoleCommands = []string{
		usbGadgetStop,
		"rm -rf /sys/kernel/config/usb_gadget/g0/configs/c.1/acm.GS0",
		"rm -f /boot/usb.acm",
		usbGadgetStart,
	}

	// The switch turns the network on as NCM, the mode the USB network
	// section defaults to. It used to write the RNDIS marker, which the web UI
	// no longer offers. See usb-network.go for the section itself.
	mountNetworkCommands = []string{
		"touch /boot/usb.ncm",
		usbGadgetStop,
		usbGadgetStart,
	}

	// The network has three markers - usb.ncm, usb.ecm and usb.rndis0 - because
	// S03usbdev builds whichever ranks first. Clearing only one leaves the
	// others on disk, and the function comes straight back at the next boot.
	// The config symlinks are removed the same way: the ones that did not bind
	// are simply absent, and `rm -rf` does not error on a path that is not there.
	unmountNetworkCommands = []string{
		usbGadgetStop,
		"rm -rf /sys/kernel/config/usb_gadget/g0/configs/c.1/ncm.usb0",
		"rm -rf /sys/kernel/config/usb_gadget/g0/configs/c.1/ecm.usb0",
		"rm -rf /sys/kernel/config/usb_gadget/g0/configs/c.1/rndis.usb0",
		"rm -f /boot/usb.ncm",
		"rm -f /boot/usb.ecm",
		"rm -f /boot/usb.rndis0",
		usbGadgetStart,
	}

	mountDiskCommands = []string{
		"touch /boot/usb.disk0",
		usbGadgetStop,
		usbGadgetStart,
	}

	// The marker goes with rm -f, for the reason the console list records.
	unmountDiskCommands = []string{
		usbGadgetStop,
		"rm -rf /sys/kernel/config/usb_gadget/g0/configs/c.1/mass_storage.disk0",
		"rm -f /boot/usb.disk0",
		usbGadgetStart,
	}

	mountAudioCommands = []string{
		"touch /boot/usb.uac",
		usbGadgetStop,
		usbGadgetStart,
	}

	// The function directory stays. Removing it blocks until every holder of
	// its character device closes it, and recovering from that needs a full
	// teardown of the gadget. The marker goes with rm -f, for the reason the
	// console list records.
	unmountAudioCommands = []string{
		usbGadgetStop,
		"rm -rf /sys/kernel/config/usb_gadget/g0/configs/c.1/uac1.usb0",
		"rm -f /boot/usb.uac",
		usbGadgetStart,
	}
)

// enabledForToggle decides whether a device is already on, for the purpose of
// picking mount or unmount. It goes through functionForDevice and checks every
// marker the function declares, not just the single marker its own mount
// command creates.
//
// The network has three markers - usb.ncm, usb.ecm and usb.rndis0 - and
// commandsFor's device name resolves only to the one its own mount command
// touches (usb.ncm). A board enabled through another marker alone would check
// that single marker as false, take the mount branch, touch usb.ncm, and
// restart the gadget with the network already on: it stays on, and a stray
// second marker is left behind. Checking every marker through the function's
// enabled method is what keeps that from happening.
func enabledForToggle(device string, present func(string) bool) bool {
	function, ok := functionForDevice(device)
	return ok && function.enabled(present)
}

// commandsFor maps a device name onto its marker and the two command lists.
// It exists so that the mapping can be tested without running anything.
func commandsFor(device string) (marker string, mount []string, unmount []string, ok bool) {
	switch device {
	case "console":
		return virtualConsole, mountConsoleCommands, unmountConsoleCommands, true
	case "network":
		return virtualNetworkNCM, mountNetworkCommands, unmountNetworkCommands, true
	case "disk":
		return virtualDisk, mountDiskCommands, unmountDiskCommands, true
	case "audio":
		return virtualAudio, mountAudioCommands, unmountAudioCommands, true
	default:
		return "", nil, nil, false
	}
}

func (s *Service) GetVirtualDevice(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, virtualDeviceState(markerPresent))

	log.Debugf("get virtual device success")
}

// virtualDeviceState is what GET /api/vm/device/virtual reports, and what an
// apply returns beside the USB network link.
func virtualDeviceState(present func(string) bool) *proto.GetVirtualDeviceRsp {
	state := func(device string) proto.VirtualDeviceState {
		function, ok := functionForDevice(device)
		if !ok {
			return proto.VirtualDeviceState{}
		}

		return proto.VirtualDeviceState{
			Enabled: function.enabled(present),
			Active:  isFunctionActive(function.name),
			Cost:    function.cost.in,
		}
	}

	return &proto.GetVirtualDeviceRsp{
		Console: state("console"),
		Network: state("network"),
		Disk:    state("disk"),
		Audio:   state("audio"),
		// The panel reports the inbound direction. It is the one that runs
		// out: with every function enabled the gadget asks for eight inbound
		// endpoints of six, and seven outbound of seven.
		Used:  usedEndpoints(present).in,
		Total: endpointBudget().in,
		Fits:  fittingSets(present),
	}
}

func (s *Service) UpdateVirtualDevice(c *gin.Context) {
	var req proto.UpdateVirtualDeviceReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid argument")
		return
	}

	device, mount, unmount, ok := commandsFor(req.Device)
	if !ok {
		rsp.ErrRsp(c, -2, "invalid arguments")
		return
	}

	present := func(marker string) bool {
		exist, _ := isDeviceExist(marker)
		return exist
	}

	commands := mount
	if enabledForToggle(req.Device, present) {
		// Turning a function off always fits, so it is never checked.
		commands = unmount
	} else if ok, free, relief := canEnable(req.Device, present); !ok {
		// Refuse rather than drop something. A person is here to be told, and
		// silently switching off what they configured earlier is worse than
		// declining what they asked for now.
		log.Infof("refused %s: %d endpoints free", req.Device, free)
		rsp.ErrRsp(c, -4, refusalMessage(req.Device, free, relief))
		return
	}

	h := hid.GetHid()
	h.Lock()
	h.CloseNoLock()
	h.ForgetAcceptingNoLock()
	defer func() {
		h.OpenNoLock()
		h.Unlock()
	}()

	for _, command := range commands {
		err := exec.Command("sh", "-c", command).Run()
		if err != nil {
			rsp.ErrRsp(c, -3, "operation failed")
			return
		}
	}

	on, _ := isDeviceExist(device)
	rsp.OkRspWithData(c, &proto.UpdateVirtualDeviceRsp{
		On: on,
	})

	log.Debugf("update virtual device %s success", req.Device)
}

func isDeviceExist(device string) (bool, error) {
	_, err := os.Stat(device)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	log.Errorf("check file %s err: %s", device, err)
	return false, err
}
