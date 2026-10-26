package agent_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/agent"
	"NanoKVM-Server/service/agent/agenttest"
	"NanoKVM-Server/service/controlmode"

	"github.com/gin-gonic/gin"
)

// These tests drive the Lifecycle with the agenttest agent: what a second
// agent gets by implementing agent.Runtime.

type fakeLock struct {
	mu    sync.Mutex
	owner string
}

func (l *fakeLock) Owner() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.owner
}

func (l *fakeLock) Release(string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.owner = ""
	return true
}

type fakeSessions struct {
	mu        sync.Mutex
	closed    []int
	events    []agent.Event
	openCount int
}

func (s *fakeSessions) Broadcast(event agent.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *fakeSessions) CloseAll(code int, _ string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = append(s.closed, code)
	n := s.openCount
	s.openCount = 0
	return n
}

func (s *fakeSessions) SessionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.openCount
}

type lifecycleHarness struct {
	t         *testing.T
	agent     *agenttest.Agent
	lifecycle *agent.Lifecycle
	control   *controlmode.Manager
	intent    *agent.IntentStore
	lock      *fakeLock
	sessions  *fakeSessions
	canceled  int
}

func newLifecycleHarness(t *testing.T, mode controlmode.Mode) *lifecycleHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	h := &lifecycleHarness{
		t:        t,
		agent:    agenttest.New(),
		control:  controlmode.NewManager(filepath.Join(tmp, "mode"), mode),
		intent:   agent.NewIntentStore(filepath.Join(tmp, "intent.json"), "AgentTest"),
		lock:     &fakeLock{},
		sessions: &fakeSessions{},
	}
	h.lifecycle = agent.NewLifecycle(agent.LifecycleConfig{
		ID:               "agenttest",
		Title:            "AgentTest",
		Runtime:          h.agent,
		Status:           agent.NewStatusStore(agenttest.InstallPath, agent.RuntimeStatus{Status: "checking"}),
		Intent:           h.intent,
		Control:          h.control,
		Lock:             h.lock,
		CancelOperations: func() int { h.canceled++; return 0 },
		Timing: agent.LifecycleTiming{
			StartTimeout:  50 * time.Millisecond,
			ReadyPoll:     5 * time.Millisecond,
			StopTimeout:   50 * time.Millisecond,
			ProbeInterval: time.Hour,
		},
	})
	h.lifecycle.SetChat(h.sessions)
	return h
}

type routeAnswer struct {
	Code    any             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (h *lifecycleHarness) call(handler gin.HandlerFunc) routeAnswer {
	h.t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/runtime", nil)
	handler(c)
	if recorder.Code != http.StatusOK {
		h.t.Fatalf("answered HTTP %d: %s", recorder.Code, recorder.Body.String())
	}
	var answer routeAnswer
	if err := json.Unmarshal(recorder.Body.Bytes(), &answer); err != nil {
		h.t.Fatalf("answer is not JSON: %v: %s", err, recorder.Body.String())
	}
	return answer
}

func (h *lifecycleHarness) intentNow() agent.RuntimeIntentStatus {
	h.t.Helper()
	intent, err := h.intent.Load()
	if err != nil {
		h.t.Fatal(err)
	}
	return intent
}

func wantCode(t *testing.T, answer routeAnswer, code any) {
	t.Helper()
	if answer.Code != code {
		t.Fatalf("code = %v (%s), want %v", answer.Code, answer.Message, code)
	}
}

func TestLifecycleStartNeedsDeviceControl(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModeOff)

	answer := h.call(h.lifecycle.Start)
	wantCode(t, answer, agent.CodeControlRequired)
	if answer.Message != "AgentTest device control is not enabled" {
		t.Fatalf("message = %q", answer.Message)
	}
	if h.agent.Launches != 0 {
		t.Fatalf("launched %d times without device control", h.agent.Launches)
	}
	var data struct {
		Status agent.RuntimeStatus `json:"status"`
	}
	if err := json.Unmarshal(answer.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Status.ControlMode != "off" || data.Status.InstallPath != agenttest.InstallPath {
		t.Fatalf("status = %+v", data.Status)
	}
}

func TestLifecycleStartLaunchesWaitsAndRecordsTheIntent(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)

	answer := h.call(h.lifecycle.Start)
	wantCode(t, answer, float64(0))
	var result agent.RuntimeStartResult
	if err := json.Unmarshal(answer.Data, &result); err != nil {
		t.Fatal(err)
	}
	if !result.Started || result.Command != "agenttest start" || result.Output != "started" {
		t.Fatalf("result = %+v", result)
	}
	if !result.Status.Ready || result.Status.Status != "ready" || !result.Status.Capabilities.DeviceWrite || !result.Status.Capabilities.Chat {
		t.Fatalf("status = %+v", result.Status)
	}
	if intent := h.intentNow(); !intent.DesiredRunning || intent.UpdatedBy != "web" {
		t.Fatalf("intent = %+v", intent)
	}

	// A second start finds the agent ready and does not launch it again.
	answer = h.call(h.lifecycle.Start)
	wantCode(t, answer, float64(0))
	if h.agent.Launches != 1 {
		t.Fatalf("launches = %d, want 1", h.agent.Launches)
	}
}

func TestLifecycleStartFailureIsRecorded(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	h.agent.LaunchErr = agent.NewError(agent.CodeRuntimeStartFailed, "launch broke")

	answer := h.call(h.lifecycle.Start)
	wantCode(t, answer, agent.CodeRuntimeStartFailed)
	if answer.Message != "launch broke" {
		t.Fatalf("message = %q", answer.Message)
	}
	if intent := h.intentNow(); intent.DesiredRunning || intent.LastError != "launch broke" {
		t.Fatalf("intent = %+v", intent)
	}
}

func TestLifecycleStopsAnAgentThatNeverBecomesReady(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	h.agent.NeverReady = true

	answer := h.call(h.lifecycle.Start)
	wantCode(t, answer, agent.CodeRuntimeStartFailed)
	if answer.Message != "agenttest is not answering" {
		t.Fatalf("message = %q", answer.Message)
	}
	if running, _ := h.agent.Running(); running || h.agent.Halts == 0 {
		t.Fatalf("running=%v halts=%d, want the half-started agent stopped", running, h.agent.Halts)
	}
	if status := h.lifecycle.RuntimeStatus(); status.Ready || status.Status != "unavailable" {
		t.Fatalf("status = %+v", status)
	}
}

func TestLifecycleStopReleasesControlClosesSessionsAndStops(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	wantCode(t, h.call(h.lifecycle.Start), float64(0))
	h.lock.owner = "s1"
	h.sessions.openCount = 2

	answer := h.call(h.lifecycle.Stop)
	wantCode(t, answer, float64(0))
	var result agent.RuntimeStartResult
	if err := json.Unmarshal(answer.Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Started || result.Status.Status != "stopped" || result.Status.ControlMode != "off" {
		t.Fatalf("result = %+v", result)
	}
	if running, _ := h.agent.Running(); running {
		t.Fatal("agent still runs")
	}
	if h.lock.Owner() != "" {
		t.Fatal("AI input lock kept")
	}
	if len(h.sessions.closed) != 1 || h.sessions.closed[0] != agent.CloseRuntimeStopped {
		t.Fatalf("closed = %v", h.sessions.closed)
	}
	if len(h.sessions.events) != 1 || h.sessions.events[0].Control.Source != "runtime_stop" {
		t.Fatalf("events = %+v", h.sessions.events)
	}
	if h.canceled == 0 {
		t.Fatal("operations were not canceled")
	}
	if intent := h.intentNow(); intent.DesiredRunning || intent.UpdatedBy != "web" {
		t.Fatalf("intent = %+v", intent)
	}
}

func TestLifecycleRestoresAnAgentThatShouldRun(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	if err := h.intent.SetDesiredRunning(true, "test"); err != nil {
		t.Fatal(err)
	}

	h.lifecycle.StartRestore()
	waitFor(t, "the lifecycle", func() bool {
		intent, _ := h.intent.Load()
		return intent.UpdatedBy == "restore"
	})
	if status := h.lifecycle.RuntimeStatus(); !status.Ready || status.Restoring {
		t.Fatalf("status = %+v", status)
	}
}

func TestLifecycleRestoreStopsAnAgentThatShouldNotRun(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	h.agent.SetRunning(true)

	h.lifecycle.StartRestore()
	waitFor(t, "the lifecycle", func() bool { return h.agent.Halts > 0 })
	waitFor(t, "the lifecycle", func() bool { return h.lifecycle.RuntimeStatus().Status == "stopped" })
}

func TestLifecycleRestoreWaitsForDeviceControl(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModeMCP)
	if err := h.intent.SetDesiredRunning(true, "test"); err != nil {
		t.Fatal(err)
	}

	h.lifecycle.StartRestore()
	waitFor(t, "the lifecycle", func() bool { return h.lifecycle.RuntimeStatus().Status == "blocked_by_mcp" })
	if h.agent.Launches != 0 {
		t.Fatal("restored while MCP owns device control")
	}
}

func TestLifecycleStopForMCP(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	wantCode(t, h.call(h.lifecycle.Start), float64(0))

	if err := h.lifecycle.StopForMCP(); err != nil {
		t.Fatal(err)
	}
	if running, _ := h.agent.Running(); running {
		t.Fatal("agent still runs")
	}
	if intent := h.intentNow(); intent.DesiredRunning || intent.UpdatedBy != "mcp_preempt" {
		t.Fatalf("intent = %+v", intent)
	}

	// PreemptForMCP ends the sessions and leaves the runtime alone.
	h.sessions.openCount = 1
	h.lock.owner = "s1"
	if err := h.lifecycle.PreemptForMCP(); err != nil {
		t.Fatal(err)
	}
	if h.lock.Owner() != "" || len(h.sessions.closed) == 0 || h.sessions.closed[len(h.sessions.closed)-1] != agent.CloseControlModeSwitched {
		t.Fatalf("lock=%q closed=%v", h.lock.Owner(), h.sessions.closed)
	}
}

func TestLifecycleInstallsInTheBackground(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModeOff)
	h.agent.SetInstalled(false)

	answer := h.call(h.lifecycle.Install)
	wantCode(t, answer, float64(0))
	var result agent.RuntimeInstallResult
	if err := json.Unmarshal(answer.Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Output != "agenttest installation started" || result.Binary != agenttest.InstallPath {
		t.Fatalf("result = %+v", result)
	}
	waitFor(t, "the lifecycle", func() bool { return h.lifecycle.RuntimeStatus().Status == "installed" })

	answer = h.call(h.lifecycle.Install)
	if err := json.Unmarshal(answer.Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Output != "agenttest is already installed" || !result.Installed {
		t.Fatalf("result = %+v", result)
	}

	// An update of a running agent is refused.
	h.agent.StaleInstall = true
	h.agent.SetRunning(true)
	answer = h.call(h.lifecycle.Install)
	wantCode(t, answer, agent.CodeRuntimeStartFailed)
	if answer.Message != "stop AgentTest before updating it" {
		t.Fatalf("message = %q", answer.Message)
	}
}

func TestLifecycleInstallFailureIsRecorded(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModeOff)
	h.agent.SetInstalled(false)
	h.agent.InstallErr = errors.New("download failed")

	wantCode(t, h.call(h.lifecycle.Install), float64(0))
	waitFor(t, "the lifecycle", func() bool { return h.lifecycle.RuntimeStatus().Status == "install_failed" })
	if status := h.lifecycle.RuntimeStatus(); status.LastError != "download failed" || status.Installed {
		t.Fatalf("status = %+v", status)
	}
}

func TestLifecycleUninstallStopsAndRemoves(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	wantCode(t, h.call(h.lifecycle.Start), float64(0))

	answer := h.call(h.lifecycle.Uninstall)
	wantCode(t, answer, float64(0))
	var result agent.RuntimeInstallResult
	if err := json.Unmarshal(answer.Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Installed || result.Output != "agenttest uninstalled successfully" || result.Status.Status != "not_installed" {
		t.Fatalf("result = %+v", result)
	}
	if h.agent.Uninstalled != 1 {
		t.Fatalf("uninstalled %d times", h.agent.Uninstalled)
	}
	if running, _ := h.agent.Running(); running {
		t.Fatal("agent still runs")
	}
}

func TestLifecycleStatusRouteAddsTheAgentsFields(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModeOff)
	h.agent.StaleInstall = true

	answer := h.call(h.lifecycle.Status)
	wantCode(t, answer, float64(0))
	var status agent.RuntimeStatus
	if err := json.Unmarshal(answer.Data, &status); err != nil {
		t.Fatal(err)
	}
	if status.AgentProfile != "test" || status.Version != agenttest.Version || !status.UpdateAvailable {
		t.Fatalf("status = %+v", status)
	}
	// The intent says stopped, so the status shows the model of a stopped
	// agent without checking it.
	if status.Status != "stopped" || !status.ModelConfigured || status.ModelName != "test-model" {
		t.Fatalf("status = %+v", status)
	}
}

func TestLifecycleReconfigureRestartsARunningAgent(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModePicoclaw)
	wantCode(t, h.call(h.lifecycle.Start), float64(0))

	applied := false
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	ok := h.lifecycle.Reconfigure(c, agent.Reconfiguration{
		Apply:   func() error { applied = true; return nil },
		Restart: true,
		Source:  "model_config",
	})
	if !ok || !applied || recorder.Body.Len() != 0 {
		t.Fatalf("ok=%v applied=%v body=%s", ok, applied, recorder.Body.String())
	}
	if h.agent.Launches != 2 || h.agent.Halts != 1 {
		t.Fatalf("launches=%d halts=%d, want a restart", h.agent.Launches, h.agent.Halts)
	}
	if intent := h.intentNow(); !intent.DesiredRunning || intent.UpdatedBy != "model_config" {
		t.Fatalf("intent = %+v", intent)
	}
}

func TestLifecycleReconfigureAnswersApplyErrors(t *testing.T) {
	h := newLifecycleHarness(t, controlmode.ModeOff)
	refreshed := false
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	ok := h.lifecycle.Reconfigure(c, agent.Reconfiguration{
		Apply:   func() error { return agent.NewError(agent.CodeInvalidRequest, "bad setting") },
		Refresh: func() { refreshed = true },
	})
	if ok || refreshed {
		t.Fatalf("ok=%v refreshed=%v", ok, refreshed)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"bad setting"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	ok = h.lifecycle.Reconfigure(c, agent.Reconfiguration{
		Apply:   func() error { return nil },
		Refresh: func() { refreshed = true },
	})
	if !ok || !refreshed || h.agent.Launches != 0 {
		t.Fatalf("ok=%v refreshed=%v launches=%d", ok, refreshed, h.agent.Launches)
	}
}
