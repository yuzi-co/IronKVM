package network

import (
	"errors"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/wol"
)

// The device list itself, its files and the wake packets are in service/wol.
// These handlers keep the routes and answers they always had, and add the
// ones the managed list needs.

func (s *Service) WakeOnLAN(c *gin.Context) {
	var req proto.WakeOnLANReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	name := wol.SanitizeName(req.Name)
	if name == "" && strings.TrimSpace(req.Mac) == "" {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	device, output, err := wol.Default.Wake(req.Mac, name)
	switch {
	case errors.Is(err, wol.ErrInvalidMAC):
		rsp.ErrRsp(c, -2, "invalid MAC address")
	case errors.Is(err, wol.ErrNotFound):
		rsp.ErrRsp(c, -4, "no device with this name")
	case err != nil && output == "":
		// Never the device: its options include the SecureOn password.
		log.Errorf("failed to wake on lan: %s", err)
		rsp.ErrRsp(c, -3, err.Error())
	case err != nil:
		log.Errorf("failed to wake on lan: %s", err)
		rsp.ErrRsp(c, -3, output)
	default:
		rsp.OkRsp(c)
		log.Debugf("wake on lan: %s", device.Mac)
	}
}

func (s *Service) GetMac(c *gin.Context) {
	var rsp proto.Response

	devices, err := wol.Default.List()
	if err != nil {
		rsp.ErrRsp(c, -2, "open file error")
		return
	}

	macs := make([]string, 0, len(devices))
	for _, d := range devices {
		line := d.Mac
		if d.Name != "" {
			line += " " + d.Name
		}
		macs = append(macs, line)
	}

	rsp.OkRspWithData(c, &proto.GetMacRsp{Macs: macs})
}

func (s *Service) SetMacName(c *gin.Context) {
	var req proto.SetMacNameReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	err := wol.Default.Rename(req.Mac, req.Name)
	switch {
	case errors.Is(err, wol.ErrInvalidMAC):
		rsp.ErrRsp(c, -2, "invalid MAC address")
	case errors.Is(err, wol.ErrInvalidName):
		rsp.ErrRsp(c, -1, "invalid arguments")
	case err != nil:
		log.Errorf("failed to rename wol mac %s: %s", req.Mac, err)
		rsp.ErrRsp(c, -3, "write failed")
	default:
		rsp.OkRsp(c)
		log.Debugf("set wol mac name: %s %s", req.Mac, req.Name)
	}
}

func (s *Service) DeleteMac(c *gin.Context) {
	var req proto.DeleteMacReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	err := wol.Default.Delete(req.Mac)
	switch {
	case errors.Is(err, wol.ErrInvalidMAC):
		rsp.ErrRsp(c, -2, "invalid MAC address")
	case err != nil:
		log.Errorf("failed to delete wol mac %s: %s", req.Mac, err)
		rsp.ErrRsp(c, -3, "write failed")
	default:
		rsp.OkRsp(c)
		log.Debugf("delete wol mac: %s", req.Mac)
	}
}

// GetWolDevices lists the devices with their options. The SecureOn password
// stays on the board.
func (s *Service) GetWolDevices(c *gin.Context) {
	var rsp proto.Response

	devices, err := wol.Default.List()
	if err != nil {
		rsp.ErrRsp(c, -2, "open file error")
		return
	}

	out := make([]proto.WolDevice, 0, len(devices))
	for _, d := range devices {
		out = append(out, proto.WolDevice{
			Mac:         d.Mac,
			Name:        d.Name,
			Broadcast:   d.Broadcast,
			Port:        d.Port,
			HasSecureOn: d.SecureOn != "",
			IP:          d.IP,
		})
	}

	rsp.OkRspWithData(c, &proto.GetWolDevicesRsp{Devices: out})
}

// SaveWolDevice adds a device, or edits one, without sending a packet.
func (s *Service) SaveWolDevice(c *gin.Context) {
	var req proto.SaveWolDeviceReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	device := wol.Device{
		Mac:  req.Mac,
		Name: req.Name,
		Options: wol.Options{
			Broadcast: req.Broadcast,
			Port:      req.Port,
			SecureOn:  req.SecureOn,
			IP:        req.IP,
		},
	}

	var inputErr *wol.InputError
	_, err := wol.Default.Save(req.OldMac, device, !req.ClearSecureOn)
	switch {
	case errors.Is(err, wol.ErrInvalidMAC):
		rsp.ErrRsp(c, -2, "invalid MAC address")
	case errors.Is(err, wol.ErrDuplicate):
		rsp.ErrRsp(c, -4, err.Error())
	case errors.Is(err, wol.ErrNotFound):
		rsp.ErrRsp(c, -5, err.Error())
	case errors.Is(err, wol.ErrInvalidName):
		rsp.ErrRsp(c, -1, err.Error())
	case errors.As(err, &inputErr):
		rsp.ErrRsp(c, -1, err.Error())
	case err != nil:
		log.Errorf("failed to save wol device %s: %s", req.Mac, err)
		rsp.ErrRsp(c, -3, "write failed")
	default:
		rsp.OkRsp(c)
		log.Debugf("saved wol device: %s", req.Mac)
	}
}

// CheckWolDevice says whether a listed device answers on its host address.
func (s *Service) CheckWolDevice(c *gin.Context) {
	var rsp proto.Response

	device, err := wol.Default.Get(c.Query("mac"))
	if err != nil {
		rsp.ErrRsp(c, -2, "no such device")
		return
	}
	if device.IP == "" {
		rsp.OkRspWithData(c, &proto.CheckWolDeviceRsp{})
		return
	}

	rsp.OkRspWithData(c, &proto.CheckWolDeviceRsp{Up: wol.IsUp(device.IP)})
}

var scanMutex sync.Mutex

// ScanWol lists the hosts the board knows on the LAN, to pick a MAC from.
func (s *Service) ScanWol(c *gin.Context) {
	var rsp proto.Response

	// avahi-browse takes seconds on a single core; one scan at a time.
	scanMutex.Lock()
	neighbors := wol.Scan()
	scanMutex.Unlock()

	listed := map[string]bool{}
	if devices, err := wol.Default.List(); err == nil {
		for _, d := range devices {
			listed[d.Mac] = true
		}
	}

	out := make([]proto.WolNeighbor, 0, len(neighbors))
	for _, n := range neighbors {
		out = append(out, proto.WolNeighbor{Mac: n.Mac, IP: n.IP, Name: n.Name, Listed: listed[n.Mac]})
	}

	rsp.OkRspWithData(c, &proto.ScanWolRsp{Neighbors: out})
}
