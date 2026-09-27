package redfish

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func managers(c *gin.Context) {
	writeJSON(c, http.StatusOK, newCollection(managersPath, "ManagerCollection", "Manager Collection", []string{managerPath}))
}

// manager is the board itself.
func (s *Service) manager(c *gin.Context) {
	body := newResource(managerPath, "Manager.v1_10_0.Manager")
	body["Id"] = "1"
	body["Name"] = "IronKVM"
	body["ManagerType"] = "BMC"
	body["FirmwareVersion"] = s.deps.FirmwareVersion()
	body["UUID"] = s.deps.UUID
	body["VirtualMedia"] = link(mediaPath)
	body["EthernetInterfaces"] = link(nicsPath)
	body["Links"] = object{
		"ManagerForServers": links(systemPath),
		"ManagerForChassis": links(chassisPath),
	}

	writeJSON(c, http.StatusOK, body)
}

func (s *Service) nics(c *gin.Context) {
	var members []string
	for _, nic := range s.deps.NICs() {
		members = append(members, nicsPath+"/"+nic.ID)
	}

	writeJSON(c, http.StatusOK, newCollection(nicsPath, "EthernetInterfaceCollection", "Manager Ethernet Interfaces", members))
}

func (s *Service) nic(c *gin.Context) {
	id := c.Param("id")
	for _, nic := range s.deps.NICs() {
		if nic.ID != id {
			continue
		}

		body := newResource(nicsPath+"/"+nic.ID, "EthernetInterface.v1_6_0.EthernetInterface")
		body["Id"] = nic.ID
		body["Name"] = "Manager Ethernet Interface " + nic.ID
		body["InterfaceEnabled"] = true
		body["LinkStatus"] = "LinkUp"
		if nic.MAC != "" {
			body["MACAddress"] = nic.MAC
		}
		addresses := []object{}
		if nic.IPv4 != "" {
			addresses = append(addresses, object{"Address": nic.IPv4})
		}
		body["IPv4Addresses"] = addresses

		writeJSON(c, http.StatusOK, body)
		return
	}

	notFound(c)
}
