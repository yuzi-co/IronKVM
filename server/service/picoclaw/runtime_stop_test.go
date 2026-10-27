package picoclaw

import (
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/agent"
)

func TestPreemptControlLeasesForMCPDoesNotStopRuntimeOrChangeIntent(t *testing.T) {
	t.Setenv("PICOCLAW_HOME", t.TempDir())
	intentStore := NewRuntimeIntentStore(filepath.Join(t.TempDir(), "picoclaw-runtime.json"))
	if err := intentStore.SetDesiredRunning(true, "test"); err != nil {
		t.Fatal(err)
	}

	lock := &agent.SessionLock{}
	lock.ForceTakeover("session-1")
	service := &Service{
		lock:          lock,
		runtime:       newRuntimeStore(RuntimeStatus{Ready: true, Status: "ready"}),
		runtimeIntent: intentStore,
		operations:    newControlOperationTracker(),
	}

	if err := service.Lifecycle().PreemptForMCP(); err != nil {
		t.Fatal(err)
	}
	if owner := lock.Owner(); owner != "" {
		t.Fatalf("lock owner = %q, want released", owner)
	}
	status := service.runtime.Get()
	if !status.Ready || status.Status != "ready" {
		t.Fatalf("runtime status changed during soft preempt: %+v", status)
	}
	intent, err := intentStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !intent.DesiredRunning || intent.UpdatedBy != "test" {
		t.Fatalf("runtime intent changed during soft preempt: %+v", intent)
	}
}

func TestStopRuntimeForMCPStopsRuntimeAndDisablesIntent(t *testing.T) {
	t.Setenv("PICOCLAW_HOME", t.TempDir())
	intentStore := NewRuntimeIntentStore(filepath.Join(t.TempDir(), "picoclaw-runtime.json"))
	if err := intentStore.SetDesiredRunning(true, "test"); err != nil {
		t.Fatal(err)
	}

	service := &Service{
		runtime:       newRuntimeStore(RuntimeStatus{Ready: true, Status: "ready"}),
		runtimeIntent: intentStore,
	}

	if err := service.Lifecycle().StopForMCP(); err != nil {
		t.Fatal(err)
	}
	status := service.runtime.Get()
	if status.Ready || status.Status != "stopped" {
		t.Fatalf("runtime status = %+v, want stopped", status)
	}
	intent, err := intentStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if intent.DesiredRunning || intent.UpdatedBy != "mcp_preempt" {
		t.Fatalf("runtime intent = %+v, want disabled by mcp_preempt", intent)
	}
}

func TestEnsureDependenciesInitializesRuntimeFields(t *testing.T) {
	service := &Service{}
	service.ensureDependencies()

	if service.config == nil {
		t.Fatal("config store was not initialized")
	}
	if service.lock == nil {
		t.Fatal("session lock was not initialized")
	}
	if service.runtime == nil {
		t.Fatal("runtime store was not initialized")
	}
	if service.control == nil {
		t.Fatal("control manager was not initialized")
	}
	if service.operations == nil {
		t.Fatal("control operation tracker was not initialized")
	}
	if service.releaseHID == nil {
		t.Fatal("HID release callback was not initialized")
	}
	if service.lifecycle == nil {
		t.Fatal("lifecycle was not initialized")
	}
}

func TestIsPicoclawGatewayCmdlineOnlyMatchesGatewaySubcommand(t *testing.T) {
	if !isPicoclawGatewayCmdline([]byte("/usr/bin/picoclaw\x00gateway\x00")) {
		t.Fatal("gateway command was not recognized")
	}
	if isPicoclawGatewayCmdline([]byte("/usr/bin/picoclaw\x00agent\x00")) {
		t.Fatal("agent command was recognized as gateway")
	}
	if isPicoclawGatewayCmdline([]byte("/usr/bin/other\x00gateway\x00")) {
		t.Fatal("other binary was recognized as picoclaw gateway")
	}
}
