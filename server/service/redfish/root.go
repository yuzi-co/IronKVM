package redfish

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// redfishVersion is the protocol version the service claims.
const redfishVersion = "1.6.0"

func (s *Service) serviceRoot(c *gin.Context) {
	body := newResource(rootPath, "ServiceRoot.v1_5_0.ServiceRoot")
	body["Id"] = "RootService"
	body["Name"] = "IronKVM Redfish Service"
	body["RedfishVersion"] = redfishVersion
	body["UUID"] = s.deps.UUID
	body["Systems"] = link(systemsPath)
	body["Chassis"] = link(chassisCollection)
	body["Managers"] = link(managersPath)
	body["SessionService"] = link(sessionServicePath)
	body["Links"] = object{"Sessions": link(sessionsPath)}

	writeJSON(c, http.StatusOK, body)
}

func serviceDocument(c *gin.Context) {
	entry := func(name, url string) object {
		return object{"name": name, "kind": "Singleton", "url": url}
	}

	writeJSON(c, http.StatusOK, object{
		"@odata.context": "/redfish/v1/$metadata",
		"value": []object{
			entry("Service", rootPath),
			entry("Systems", systemsPath),
			entry("Chassis", chassisCollection),
			entry("Managers", managersPath),
			entry("SessionService", sessionServicePath),
			entry("Sessions", sessionsPath),
		},
	})
}

// metadataDocument references every DMTF schema a resource here uses. The
// versioned namespaces must match the @odata.type values the handlers write.
const metadataDocument = `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/ServiceRoot_v1.xml">
    <edmx:Include Namespace="ServiceRoot"/>
    <edmx:Include Namespace="ServiceRoot.v1_5_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/ComputerSystemCollection_v1.xml">
    <edmx:Include Namespace="ComputerSystemCollection"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/ComputerSystem_v1.xml">
    <edmx:Include Namespace="ComputerSystem"/>
    <edmx:Include Namespace="ComputerSystem.v1_13_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/ChassisCollection_v1.xml">
    <edmx:Include Namespace="ChassisCollection"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/Chassis_v1.xml">
    <edmx:Include Namespace="Chassis"/>
    <edmx:Include Namespace="Chassis.v1_14_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/ManagerCollection_v1.xml">
    <edmx:Include Namespace="ManagerCollection"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/Manager_v1.xml">
    <edmx:Include Namespace="Manager"/>
    <edmx:Include Namespace="Manager.v1_10_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/EthernetInterfaceCollection_v1.xml">
    <edmx:Include Namespace="EthernetInterfaceCollection"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/EthernetInterface_v1.xml">
    <edmx:Include Namespace="EthernetInterface"/>
    <edmx:Include Namespace="EthernetInterface.v1_6_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/VirtualMediaCollection_v1.xml">
    <edmx:Include Namespace="VirtualMediaCollection"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/VirtualMedia_v1.xml">
    <edmx:Include Namespace="VirtualMedia"/>
    <edmx:Include Namespace="VirtualMedia.v1_3_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/SessionService_v1.xml">
    <edmx:Include Namespace="SessionService"/>
    <edmx:Include Namespace="SessionService.v1_1_8"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/SessionCollection_v1.xml">
    <edmx:Include Namespace="SessionCollection"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/Session_v1.xml">
    <edmx:Include Namespace="Session"/>
    <edmx:Include Namespace="Session.v1_3_0"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/Message_v1.xml">
    <edmx:Include Namespace="Message"/>
    <edmx:Include Namespace="Message.v1_1_1"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/Resource_v1.xml">
    <edmx:Include Namespace="Resource"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/IPAddresses_v1.xml">
    <edmx:Include Namespace="IPAddresses"/>
  </edmx:Reference>
  <edmx:Reference Uri="http://redfish.dmtf.org/schemas/v1/RedfishExtensions_v1.xml">
    <edmx:Include Namespace="RedfishExtensions.v1_0_0" Alias="Redfish"/>
  </edmx:Reference>
  <edmx:DataServices>
    <Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Service">
      <EntityContainer Name="Service" Extends="ServiceRoot.v1_5_0.ServiceContainer"/>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>
`

func metadata(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(metadataDocument))
}
