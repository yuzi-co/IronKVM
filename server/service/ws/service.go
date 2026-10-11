package ws

import (
	"time"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/audit"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Service struct{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     middleware.SameOrigin,
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Connect(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("create websocket failed: %s", err)
		return
	}
	stopSessionWatcher := middleware.WatchWebSocket(c.Request.Context(), ws)
	defer stopSessionWatcher()

	log.Debug("websocket connected")

	// The socket carries the keyboard and mouse of the browser, so its life is
	// a control session of the host.
	started := time.Now()
	audit.Request(c, audit.Event{Category: audit.CategoryControl, Action: "session-start", Outcome: audit.OutcomeOK})
	defer func() {
		audit.Request(c, audit.Event{
			Category: audit.CategoryControl,
			Action:   "session-stop",
			Outcome:  audit.OutcomeOK,
			Detail:   "duration " + time.Since(started).Round(time.Second).String(),
		})
	}()

	client := NewClient(ws)

	manager := GetManager()
	manager.AddClient(ws, client)
	defer manager.RemoveClient(ws)

	sendCaptureStatusSnapshot(client)
	sendH264ModeStatusSnapshot(client)
	sendKeyboardLedStatusSnapshot(client)

	client.Start()
}
