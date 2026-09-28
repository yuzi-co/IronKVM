package redfish

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func chassisList(c *gin.Context) {
	writeJSON(c, http.StatusOK, newCollection(chassisCollection, "ChassisCollection", "Chassis Collection", []string{chassisPath}))
}

// chassis is the host's case. The board knows nothing about it beyond the
// power LED, but sushy and the validator expect a system to have one.
func (s *Service) chassis(c *gin.Context) {
	body := newResource(chassisPath, "Chassis.v1_14_0.Chassis")
	body["Id"] = "1"
	body["Name"] = "Managed host chassis"
	body["ChassisType"] = "Other"
	body["PowerState"] = powerState(s.powerLED())
	body["Links"] = object{
		"ComputerSystems": links(systemPath),
		"ManagedBy":       links(managerPath),
	}

	writeJSON(c, http.StatusOK, body)
}
