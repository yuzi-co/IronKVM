package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/network"
)

func networkRouter(r *gin.Engine) {
	service := network.NewService()

	r.POST("/api/network/wifi", service.ConnectWifiNoAuth)    // connect Wi-Fi without auth (only available in ap mode)
	r.POST("/api/network/wifi/verify", service.VerifyApLogin) // verify ap login

	api := r.Group("/api").Use(middleware.CheckToken())

	api.POST("/network/wol", service.WakeOnLAN)            // wake on lan
	api.GET("/network/wol/mac", service.GetMac)            // get mac list
	api.GET("/network/wol/devices", service.GetWolDevices) // device list with options
	api.GET("/network/wol/check", service.CheckWolDevice)  // is the device up

	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.DELETE("/network/wol/mac", service.DeleteMac)      // delete mac
	admin.POST("/network/wol/mac/name", service.SetMacName)  // set mac name
	admin.POST("/network/wol/device", service.SaveWolDevice) // add or edit a device
	admin.GET("/network/wol/scan", service.ScanWol)          // hosts on the LAN

	admin.GET("/network/wifi", service.GetWifi)                    // get Wi-Fi information
	admin.POST("/network/wifi/connect", service.ConnectWifi)       // connect Wi-Fi
	admin.POST("/network/wifi/disconnect", service.DisconnectWifi) // disconnect Wi-Fi
	admin.GET("/network/dns", service.GetDNS)                      // get DNS configuration
	admin.POST("/network/dns", service.SetDNS)                     // set DNS configuration

	// Remote syslog: the collector S01syslogd forwards the system log to.
	admin.GET("/network/syslog", service.GetSyslog)        // get the remote syslog setting and what runs
	admin.POST("/network/syslog", service.SetSyslog)       // save the collector and restart syslogd
	admin.POST("/network/syslog/test", service.TestSyslog) // send one message through the local syslogd

	// The address routes change the interface the caller is talking over, so
	// a change is applied on trial and reverts unless it is confirmed from
	// the new address.
	admin.GET("/network/ethernet", service.GetEthernet)              // get the wired address configuration
	admin.POST("/network/ethernet", service.SetEthernet)             // apply a wired address configuration on trial
	admin.POST("/network/ethernet/confirm", service.ConfirmEthernet) // keep the applied configuration

	// IPv6 leaves the IPv4 address alone, so it needs no trial: it is saved after
	// it applied, and an apply that fails puts the saved setting back.
	admin.GET("/network/ipv6", service.GetIPv6)  // get the IPv6 configuration
	admin.POST("/network/ipv6", service.SetIPv6) // apply an IPv6 configuration
}
