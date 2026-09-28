package vm

import (
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
)

// powerLEDMu guards config's HardwareSettings.PowerLED, which the owner can
// change while Redfish and the UI read it.
var powerLEDMu sync.RWMutex

// PowerLEDConnected reports whether the owner has said the host's power LED
// header is wired to the board. Without it the LED line reads "off" whatever
// the host does, so nothing may decide from it.
func PowerLEDConnected() bool {
	powerLEDMu.RLock()
	defer powerLEDMu.RUnlock()

	return config.GetInstance().HardwareSettings.PowerLED
}

// savePowerLEDSetting writes the setting to server.yaml. Tests replace it.
var savePowerLEDSetting = func(connected bool) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.HardwareSettings.PowerLED = connected

	return config.Write(conf)
}

func (s *Service) GetPowerLED(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, &proto.PowerLEDSetting{Connected: PowerLEDConnected()})
}

// SetPowerLED records whether the power LED is wired, in server.yaml and in
// the running server.
func (s *Service) SetPowerLED(c *gin.Context) {
	var rsp proto.Response
	var req proto.PowerLEDSetting

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	powerLEDMu.Lock()
	defer powerLEDMu.Unlock()

	if err := savePowerLEDSetting(req.Connected); err != nil {
		log.Errorf("save the power LED setting: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the setting")
		return
	}
	config.GetInstance().HardwareSettings.PowerLED = req.Connected

	log.Infof("power LED connected: %t", req.Connected)
	rsp.OkRsp(c)
}
