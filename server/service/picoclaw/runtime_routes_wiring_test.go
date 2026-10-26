package picoclaw

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/service/controlmode"
)

// runtimeRoutes are the handlers TestRuntimeRoutesAnswerAsRecorded calls.
type runtimeRoutes struct {
	status, start, stop, install, uninstall gin.HandlerFunc
	modelConfig, agentProfile               gin.HandlerFunc
	releaseSession                          gin.HandlerFunc
}

func newRuntimeRoutesService(control *controlmode.Manager, intentPath string) *Service {
	return &Service{
		control:       control,
		runtime:       newRuntimeStore(RuntimeStatus{}),
		runtimeIntent: NewRuntimeIntentStore(intentPath),
		releaseHID:    func() error { return nil },
		lock:          &SessionLock{},
	}
}

func runtimeRoutesUnderTest(s *Service) runtimeRoutes {
	lifecycle := s.Lifecycle()
	return runtimeRoutes{
		status:         lifecycle.Status,
		start:          lifecycle.Start,
		stop:           lifecycle.Stop,
		install:        lifecycle.Install,
		uninstall:      lifecycle.Uninstall,
		modelConfig:    s.UpdateModelConfig,
		agentProfile:   s.UpdateAgentProfile,
		releaseSession: s.ReleaseRuntimeSession,
	}
}
