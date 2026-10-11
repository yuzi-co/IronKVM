package router

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/common"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/events"
	"NanoKVM-Server/service/failsafe"
	kvmcapture "NanoKVM-Server/service/mcp/capture/kvm"
	"NanoKVM-Server/service/notify"
	"NanoKVM-Server/utils"
)

// notifyFile is a variable so tests can keep the tokens off /etc/kvm.
var notifyFile = notify.File

// notifySnapshotQuality is the JPEG quality of a notification's picture. It is
// for a person to read, as the watchdog's screenshot is.
const notifySnapshotQuality = 60

// notifyTimeout bounds one send on one channel.
const notifyTimeout = 20 * time.Second

func notifyRouter(r *gin.Engine) {
	snapshotter := watchdogCapture(kvmcapture.New())

	service := notify.New(notify.Deps{
		Store: notify.NewStore(notifyFile),
		Bus:   events.Default,

		Snapshot: func(ctx context.Context) ([]byte, error) {
			return snapshotter(ctx, notifySnapshotQuality)
		},
		Signal: func() (bool, bool) {
			if utils.IsHdmiDisabled() || common.VideoDisabled() {
				return false, false
			}
			return common.GetKvmVision().HasHDMISignal(), true
		},
		Client: func() *http.Client { return utils.OutboundClient(notifyTimeout) },
		Device: func() string {
			if name, err := os.Hostname(); err == nil && name != "" {
				return name
			}
			return "IronKVM"
		},
	})
	service.Start()

	// A server that started in failsafe mode says so once, as it is up. The
	// event cannot come from failsafe.Init: that runs before anything listens.
	if st := failsafe.Current(); st.Active {
		events.Publish(events.Event{
			Type:       events.FailsafeEntered,
			Title:      "Failsafe mode: the server started without video",
			Message:    st.Reason,
			NoSnapshot: true,
		})
	}

	// The web UI's Notifications page.
	admin := r.Group("/api/notify").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/settings", service.GetSettings)
	admin.POST("/settings", service.SetSettings)
	admin.GET("/state", service.GetState)
	admin.POST("/test", service.PostTest)
}
