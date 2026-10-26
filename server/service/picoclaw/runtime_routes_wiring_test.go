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
		runtime:       &RuntimeStore{},
		runtimeIntent: NewRuntimeIntentStore(intentPath),
		releaseHID:    func() error { return nil },
		lock:          &SessionLock{},
	}
}

func runtimeRoutesUnderTest(s *Service) runtimeRoutes {
	return runtimeRoutes{
		status:         s.GetRuntimeStatus,
		start:          s.StartRuntime,
		stop:           s.StopRuntime,
		install:        s.InstallRuntime,
		uninstall:      s.UninstallRuntime,
		modelConfig:    s.UpdateModelConfig,
		agentProfile:   s.UpdateAgentProfile,
		releaseSession: s.ReleaseRuntimeSession,
	}
}
