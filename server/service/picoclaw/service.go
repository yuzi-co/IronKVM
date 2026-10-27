package picoclaw

import (
	"context"
	"sync"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/vm"

	"github.com/gin-gonic/gin"
)

// sessionIDHeader is agentSessionHeader's old name.
const sessionIDHeader = "X-PicoClaw-Session-ID"

// RuntimeIntentFile keeps whether PicoClaw should run, across reboots.
const RuntimeIntentFile = "/etc/kvm/picoclaw-runtime.json"

var (
	configStoreOnce        sync.Once
	configStore            *ConfigStore
	runtimeStoreOnce       sync.Once
	runtimeStore           *RuntimeStore
	runtimeIntentStoreOnce sync.Once
	runtimeIntentStore     *RuntimeIntentStore
)

// picoclawLifecycleTiming bounds PicoClaw's start and stop.
var picoclawLifecycleTiming = agent.LifecycleTiming{
	StartSettle:   picoclawStartWaitPeriod,
	StartTimeout:  picoclawStartTimeout,
	ReadyPoll:     picoclawReadyPollPeriod,
	StopTimeout:   picoclawStopTimeout,
	ProbeInterval: 30 * time.Second,
	StatusRefresh: 2 * time.Second,
}

func NewService(control *controlmode.Manager) *Service {
	if control == nil {
		control = controlmode.GetManager()
	}
	service := &Service{
		vision:             common.GetKvmVision(),
		hid:                hid.GetHid(),
		config:             getConfigStore(),
		lock:               GetSessionLock(),
		runtime:            getRuntimeStore(),
		runtimeIntent:      getRuntimeIntentStore(),
		control:            control,
		releaseHID:         hid.ReleaseAllHIDStateBestEffort,
		operations:         newControlOperationTracker(),
		acquireHDMIForRead: vm.AcquireHdmiCaptureLeaseForRead,
	}
	service.ensureDependencies()
	service.lifecycle.StartRestore()
	return service
}

func (s *Service) ensureDependencies() {
	if s == nil {
		return
	}
	if s.vision == nil {
		s.vision = common.GetKvmVision()
	}
	if s.hid == nil {
		s.hid = hid.GetHid()
	}
	if s.config == nil {
		s.config = getConfigStore()
	}
	if s.lock == nil {
		s.lock = GetSessionLock()
	}
	if s.runtime == nil {
		s.runtime = getRuntimeStore()
	}
	if s.runtimeIntent == nil {
		s.runtimeIntent = getRuntimeIntentStore()
	}
	if s.control == nil {
		s.control = controlmode.GetManager()
	}
	if s.releaseHID == nil {
		s.releaseHID = hid.ReleaseAllHIDStateBestEffort
	}
	if s.operations == nil {
		s.operations = newControlOperationTracker()
	}
	if s.acquireHDMIForRead == nil {
		s.acquireHDMIForRead = vm.AcquireHdmiCaptureLeaseForRead
	}
	if s.chat == nil {
		s.chat = noChat{}
	}
	if s.lifecycle == nil {
		s.lifecycle = agent.NewLifecycle(agent.LifecycleConfig{
			ID:               "picoclaw",
			Title:            "PicoClaw",
			Runtime:          s.Agent(),
			Status:           s.runtime,
			Intent:           s.runtimeIntent,
			Control:          s.control,
			Lock:             s.lock,
			BeginOperation:   s.beginRuntimeLifecycleOperation,
			CancelOperations: s.CancelActiveControlOperations,
			ReleaseHID:       func() error { return s.releaseHID() },
			Timing:           picoclawLifecycleTiming,
		})
		s.lifecycle.SetChat(s.chat)
	}
}

// Lifecycle is the agent-neutral lifecycle that runs PicoClaw: its runtime
// routes and what the AI control mode switch asks of it.
func (s *Service) Lifecycle() *agent.Lifecycle {
	s.ensureDependencies()
	return s.lifecycle
}

func getConfigStore() *ConfigStore {
	configStoreOnce.Do(func() {
		configStore = &ConfigStore{
			config: defaultConfig(),
		}
	})

	return configStore
}

// newRuntimeStore returns a status store for PicoClaw holding status.
func newRuntimeStore(status RuntimeStatus) *RuntimeStore {
	return agent.NewStatusStore(picoclawBinaryPath, status)
}

func getRuntimeStore() *RuntimeStore {
	runtimeStoreOnce.Do(func() {
		runtimeStore = newRuntimeStore(RuntimeStatus{
			Ready:           false,
			Installed:       false,
			Installing:      false,
			InstallProgress: 0,
			ModelConfigured: false,
			Status:          "checking",
		})
	})

	return runtimeStore
}

// NewRuntimeIntentStore returns a store that keeps PicoClaw's intent to run
// in path, RuntimeIntentFile when path is empty.
func NewRuntimeIntentStore(path string) *RuntimeIntentStore {
	if path == "" {
		path = RuntimeIntentFile
	}
	return agent.NewIntentStore(path, "PicoClaw")
}

func getRuntimeIntentStore() *RuntimeIntentStore {
	runtimeIntentStoreOnce.Do(func() {
		runtimeIntentStore = NewRuntimeIntentStore(RuntimeIntentFile)
	})

	return runtimeIntentStore
}

func (s *Service) GetRuntimeSession(c *gin.Context) {
	s.ensureDependencies()
	writeSuccess(c, gin.H{
		"current_session": s.lock.Owner(),
		"checked_at":      time.Now(),
	})
}

func (s *Service) acquireControlMode() (func(), *PicoclawError) {
	s.ensureDependencies()
	release, err := s.lifecycle.AcquireControl()
	if err != nil {
		return nil, picoclawErrorFrom(err)
	}
	return release, nil
}

// controlWriteError says why PicoClaw does not own device control.
func (s *Service) controlWriteError(_ controlmode.Mode, err error) *PicoclawError {
	s.ensureDependencies()
	return picoclawErrorFrom(s.lifecycle.ControlError(err))
}

// operationError is the lifecycle's error for a runtime lifecycle operation
// whose ctx ended, nil before it ends.
func (s *Service) operationError(ctx context.Context) *PicoclawError {
	s.ensureDependencies()
	if err := s.lifecycle.OperationError(ctx); err != nil {
		return picoclawErrorFrom(err)
	}
	return nil
}

func (s *Service) requireSessionID(c *gin.Context) (string, *PicoclawError) {
	sessionID := namedSession(c.Request.Header)
	if sessionID == "" {
		return "", newPicoclawError(CodeSessionIDMissing, "missing X-PicoClaw-Session-ID")
	}

	return sessionID, nil
}

func (s *ConfigStore) Get() Config {
	if s == nil {
		return defaultConfig()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *ConfigStore) Set(cfg Config) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

func picoclawInstalledState() (bool, bool) {
	installed, err := isPicoclawInstalled()
	if err != nil {
		return false, false
	}
	return installed, true
}
