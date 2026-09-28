package router

import (
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/ipmi"
	"NanoKVM-Server/service/redfish"
	"NanoKVM-Server/service/vm"
	"NanoKVM-Server/utils"
)

// ipmiKeyFile is a variable so tests can keep the key out of /etc/kvm.
var ipmiKeyFile = ipmi.KeyFile

func ipmiRouter(r *gin.Engine) {
	service := ipmi.New(ipmi.Deps{
		PressButton:       vm.PressButton,
		PowerLED:          vm.PowerLED,
		PowerLEDConnected: vm.PowerLEDConnected,

		Accounts: authn.DefaultStore,
		Keyring:  ipmi.NewKeyring(ipmiKeyFile),
		// The same record the Redfish login counts failures under, per
		// address and account.
		Limiter: redfish.LoginLimiter{},

		FirmwareVersion: vm.FirmwareVersion,
		// The Redfish service UUID, so both services name the same system.
		GUID: ipmiGUID(redfish.LoadUUID(redfishUUIDFile)),

		Addr:       utils.ListenAddr(config.GetInstance().Host, strconv.Itoa(ipmi.Port)),
		Enabled:    ipmiEnabled,
		SetEnabled: setIPMIEnabled,
	})
	service.Start()

	// The web UI's IPMI page.
	admin := r.Group("/api/ipmi").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
	admin.POST("/users/:username/password", service.SetPassword)
	admin.DELETE("/users/:username/password", service.ClearPassword)
}

// ipmiGUID encodes a UUID the way SMBIOS and most BMCs send a GUID: the
// first three fields little-endian, the rest as they are.
func ipmiGUID(id string) [16]byte {
	var out [16]byte
	u, err := uuid.Parse(id)
	if err != nil {
		return out
	}
	copy(out[:], u[:])
	out[0], out[1], out[2], out[3] = u[3], u[2], u[1], u[0]
	out[4], out[5] = u[5], u[4]
	out[6], out[7] = u[7], u[6]
	return out
}

// ipmiSettingMu guards config's IPMI.Enabled, which the owner can change
// while the service reads it.
var ipmiSettingMu sync.RWMutex

func ipmiEnabled() bool {
	ipmiSettingMu.RLock()
	defer ipmiSettingMu.RUnlock()

	return config.GetInstance().IPMI.Enabled
}

// saveIPMISetting writes the setting to server.yaml. Tests replace it.
var saveIPMISetting = func(enabled bool) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}

	conf.IPMI.Enabled = enabled

	return config.Write(conf)
}

// setIPMIEnabled saves the setting to server.yaml and applies it to the
// running configuration.
func setIPMIEnabled(enabled bool) error {
	ipmiSettingMu.Lock()
	defer ipmiSettingMu.Unlock()

	if err := saveIPMISetting(enabled); err != nil {
		return err
	}
	config.GetInstance().IPMI.Enabled = enabled
	return nil
}
