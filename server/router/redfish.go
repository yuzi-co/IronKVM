package router

import (
	"net"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/apikey"
	"NanoKVM-Server/service/redfish"
	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/service/vm"
)

// redfishUUIDFile is a variable so tests can keep the UUID out of /etc/kvm.
var redfishUUIDFile = redfish.UUIDFile

func redfishRouter(r *gin.Engine) {
	service := redfish.New(redfish.Deps{
		PressButton: vm.PressButton,
		PowerLED:    vm.PowerLED,

		ListDrives:  storage.ListDrives,
		InsertDrive: storage.InsertDrive,
		EjectDrive:  storage.EjectDrive,
		ImageDir:    storage.ImageDirectory,

		Accounts:   authn.DefaultStore,
		APIKeyUser: redfishAPIKeyUser,
		Limiter:    redfish.LoginLimiter{},
		AuthDisabled: func() bool {
			return config.GetInstance().Authentication == "disable"
		},
		// The web login waits as long after a wrong password.
		FailureDelay: 2 * time.Second,

		FirmwareVersion: vm.FirmwareVersion,
		NICs:            redfishNICs,

		UUID: redfish.LoadUUID(redfishUUIDFile),
	})

	service.Register(r)
}

// redfishAPIKeyUser returns the account an API key belongs to. A key issued
// before keys carried an account belongs to nobody, as middleware/apikey.go
// treats it.
func redfishAPIKeyUser(secret string) (string, bool) {
	key, ok := apikey.Verify(secret)
	if !ok || key.Username == "" {
		return "", false
	}
	return key.Username, true
}

// redfishNICs lists the board's up interfaces with an address, with their MAC
// addresses, which vm.GetInterfaceInfos does not carry.
func redfishNICs() []redfish.NIC {
	infos, err := vm.GetInterfaceInfos()
	if err != nil {
		return nil
	}

	nics := make([]redfish.NIC, 0, len(infos))
	for _, info := range infos {
		nic := redfish.NIC{ID: info.Name}
		if iface, err := net.InterfaceByName(info.Name); err == nil {
			nic.MAC = iface.HardwareAddr.String()
		}
		if ip := info.IP.To4(); ip != nil {
			nic.IPv4 = ip.String()
		}
		nics = append(nics, nic)
	}
	return nics
}
