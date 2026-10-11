package vm

import (
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/failsafe"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetFailsafe reports whether the server runs without video after a video
// crash loop, why, and the exits that led there. It always succeeds.
func (s *Service) GetFailsafe(c *gin.Context) {
	var rsp proto.Response

	st := failsafe.Current()
	rsp.OkRspWithData(c, &st)
}

// restartDelay lets the answer reach the browser before the restart closes
// its connection.
var restartDelay = 500 * time.Millisecond

// RetryFailsafe forgets the crash records and starts the server again, with
// video. A crash after that counts from one again, so the server needs a whole
// new loop before it gives up on video a second time.
func (s *Service) RetryFailsafe(c *gin.Context) {
	var rsp proto.Response

	if !failsafe.Current().Active {
		rsp.ErrRsp(c, -1, "the server is not in video failsafe mode")
		return
	}
	if failsafe.Restart == nil {
		rsp.ErrRsp(c, -2, "the server cannot restart itself")
		return
	}
	if err := failsafe.Retry(); err != nil {
		log.Errorf("failsafe: clearing the crash records failed: %s", err)
		rsp.ErrRsp(c, -3, "clearing the crash records failed: "+err.Error())
		return
	}

	log.Warnf("failsafe: video retry requested, restarting")
	rsp.OkRsp(c)

	restart := failsafe.Restart
	go func() {
		time.Sleep(restartDelay)
		restart()
	}()
}
