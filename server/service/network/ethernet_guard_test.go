package network

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runShell runs the detached revert's script in the foreground, which is the
// only difference between the test and the device. The shell is the same one:
// /bin/sh is busybox there and dash or bash here, and the script uses nothing
// the three disagree about.
func runShell(script string) error {
	output, err := exec.Command("sh", "-c", script).CombinedOutput()
	if err != nil {
		return &shellError{script: script, output: string(output), err: err}
	}

	return nil
}

type shellError struct {
	script string
	output string
	err    error
}

func (e *shellError) Error() string {
	return e.err.Error() + ": " + strings.TrimSpace(e.output) + " (" + e.script + ")"
}

// The property every case here is about: a trial that nobody confirms puts the
// address back even when the process that started it is gone.

func TestATrialRecordsItsDeadlineOutsideTheProcess(t *testing.T) {
	useTempConfig(t)
	reverts := clearTrial(t)

	config := ethernetConfig{Address: "10.0.0.99", Prefix: 24, Gateway: "10.0.0.1"}
	startTrial("token-outside", ethModeStatic, config, 90)

	state, ok := readTrialState()
	if !ok {
		t.Fatal("the trial was not recorded, so nothing outside this process could revert it")
	}
	if state.Token != "token-outside" {
		t.Errorf("token = %q, want %q", state.Token, "token-outside")
	}
	if state.Mode != ethModeStatic || state.Address != "10.0.0.99" || state.Prefix != 24 ||
		state.Gateway != "10.0.0.1" {
		t.Errorf("the record does not describe the trial: %+v", state)
	}

	calls := reverts.all()
	if len(calls) != 1 {
		t.Fatalf("started %d detached reverts, want 1", len(calls))
	}
	if calls[0].token != "token-outside" {
		t.Errorf("detached revert token = %q, want %q", calls[0].token, "token-outside")
	}

	// Past the deadline, so the in-process timer wins on a board that is alive.
	want := 90*time.Second + externalRevertGrace
	if calls[0].after != want {
		t.Errorf("detached revert waits %s, want %s", calls[0].after, want)
	}
}

func TestAConfirmationRetiresTheDetachedRevert(t *testing.T) {
	path := useTempConfig(t)
	clearTrial(t)
	recordCommands(t)

	startTrial("token-keep", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)
	if err := confirmTrial("token-keep"); err != nil {
		t.Fatalf("failed to confirm: %s", err)
	}

	if _, ok := readTrialState(); ok {
		t.Error("the record survived the confirmation, so the detached revert would still run")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the configuration was not saved: %s", err)
	}
}

func TestAnInProcessRevertRetiresTheDetachedOne(t *testing.T) {
	useTempConfig(t)
	clearTrial(t)
	recordCommands(t)

	startTrial("token-gone", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)
	revertNow()

	if _, ok := readTrialState(); ok {
		t.Error("the record survived the revert, so the boot script would run a second time")
	}
}

func TestASecondTrialReplacesTheRecord(t *testing.T) {
	useTempConfig(t)
	reverts := clearTrial(t)

	startTrial("token-first", ethModeStatic, ethernetConfig{Address: "10.0.0.98", Prefix: 24}, 60)
	startTrial("token-second", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)

	state, ok := readTrialState()
	if !ok {
		t.Fatal("no record after the second trial")
	}
	if state.Token != "token-second" {
		t.Errorf("token = %q, want the second trial's", state.Token)
	}

	// The first shell is still sleeping. It wakes, does not find its own token
	// in the record, and exits.
	if calls := reverts.all(); len(calls) != 2 {
		t.Fatalf("started %d detached reverts, want 2", len(calls))
	}
}

// A trial started by a server that has since restarted. The address is live,
// the detached revert is still counting, and this process knows none of it
// except what the file says.

func TestATrialFromAPreviousProcessIsReported(t *testing.T) {
	useTempConfig(t)
	clearTrial(t)

	deadline := time.Now().Add(2 * time.Minute)
	if err := writeTrialState(trialState{
		Token:    "token-adopted",
		Mode:     ethModeStatic,
		Address:  "10.0.0.99",
		Prefix:   24,
		Gateway:  "10.0.0.1",
		Deadline: deadline,
	}); err != nil {
		t.Fatalf("failed to write the record: %s", err)
	}

	described := describeTrial()
	if described == nil {
		t.Fatal("reported no trial, so the page would say the address is settled")
	}
	if described.Token != "token-adopted" || described.Address != "10.0.0.99" {
		t.Errorf("described %+v, want the recorded trial", described)
	}
	if described.RemainingSeconds < 100 || described.RemainingSeconds > 120 {
		t.Errorf("remaining = %d, want about 120", described.RemainingSeconds)
	}
}

func TestATrialFromAPreviousProcessCanBeConfirmed(t *testing.T) {
	path := useTempConfig(t)
	clearTrial(t)
	recordCommands(t)

	if err := writeTrialState(trialState{
		Token:    "token-adopted",
		Mode:     ethModeStatic,
		Address:  "10.0.0.99",
		Prefix:   24,
		Gateway:  "10.0.0.1",
		Deadline: time.Now().Add(2 * time.Minute),
	}); err != nil {
		t.Fatalf("failed to write the record: %s", err)
	}

	if err := confirmTrial("token-adopted"); err != nil {
		t.Fatalf("failed to confirm a trial this process did not start: %s", err)
	}

	saved, ok := readEthernetConfig()
	if !ok {
		t.Fatalf("nothing was saved to %s", path)
	}
	if saved.Address != "10.0.0.99" || saved.Prefix != 24 || saved.Gateway != "10.0.0.1" {
		t.Errorf("saved %+v, want the recorded trial", saved)
	}
	if _, ok := readTrialState(); ok {
		t.Error("the record survived the confirmation")
	}
}

func TestConfirmingATrialFromAPreviousProcessNeedsTheRightToken(t *testing.T) {
	useTempConfig(t)
	clearTrial(t)

	if err := writeTrialState(trialState{
		Token:    "token-adopted",
		Mode:     ethModeStatic,
		Address:  "10.0.0.99",
		Prefix:   24,
		Deadline: time.Now().Add(2 * time.Minute),
	}); err != nil {
		t.Fatalf("failed to write the record: %s", err)
	}

	if err := confirmTrial("token-wrong"); err == nil {
		t.Fatal("a wrong token confirmed a trial")
	}
	if _, ok := readTrialState(); !ok {
		t.Error("a refused confirmation removed the record, so nothing would revert")
	}
}

// A confirmation that cannot save must leave the deadline running. The address
// would otherwise work until the next reboot and be lost then.
func TestAConfirmationThatCannotSaveLeavesTheDeadlineRunning(t *testing.T) {
	useTempConfig(t)
	reverts := clearTrial(t)
	recordCommands(t)

	startTrial("token-unwritable", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)

	// A path whose parent does not exist, so the save fails and nothing else does.
	ethConfigFile = filepath.Join(t.TempDir(), "no-such-directory", "eth.nodhcp")

	if err := confirmTrial("token-unwritable"); err == nil {
		t.Fatal("a confirmation that could not save reported success")
	}

	if _, ok := readTrialState(); !ok {
		t.Error("the record was removed, so the detached revert would not run")
	}
	if describeTrial() == nil {
		t.Error("the trial was forgotten, so nothing in this process would revert either")
	}
	if len(reverts.all()) != 1 {
		t.Errorf("started %d detached reverts, want 1", len(reverts.all()))
	}
}

func TestATrialWithNoDetachedRevertKeepsNoRecord(t *testing.T) {
	useTempConfig(t)
	reverts := clearTrial(t)
	reverts.fail = true

	startTrial("token-nospawn", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)

	// The in-process timer still reverts, so the trial goes ahead. What must
	// not happen is a record claiming a revert that nothing is counting to.
	if _, ok := readTrialState(); ok {
		t.Error("kept a record after the detached revert failed to start")
	}
	if describeTrial() == nil {
		t.Error("the trial was abandoned; the in-process timer should still hold it")
	}
}

// The script is the whole mechanism, so it is checked rather than assumed.
func TestTheDetachedScriptChecksTheTokenBeforeItReverts(t *testing.T) {
	clearTrial(t)

	script := detachedRevertScript("token-abc", 75)

	for _, want := range []string{
		"sleep 75",
		"grep -q token-abc",
		trialStateFile,
		"rm -f " + trialStateFile,
		ethInitScript + " start",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("the script does not contain %q: %s", want, script)
		}
	}

	// The order is the contract: sleep, then check the token, then remove the
	// record, and only then run the boot script.
	sleepAt := strings.Index(script, "sleep 75")
	grepAt := strings.Index(script, "grep -q")
	removeAt := strings.Index(script, "rm -f")
	startAt := strings.Index(script, ethInitScript+" start")
	if !(sleepAt < grepAt && grepAt < removeAt && removeAt < startAt) {
		t.Errorf("the script does the right things in the wrong order: %s", script)
	}
}

// The script runs under busybox sh on the device, so it is run here too.
func TestTheDetachedScriptRunsTheBootScriptOnlyWhenTheRecordIsStillThere(t *testing.T) {
	clearTrial(t)

	tests := []struct {
		name       string
		record     string
		wantRevert bool
	}{
		{name: "the record still names this trial", record: `{"token":"token-abc"}`, wantRevert: true},
		{name: "a later trial replaced it", record: `{"token":"token-xyz"}`, wantRevert: false},
		{name: "the confirmation removed it", record: "", wantRevert: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			work := t.TempDir()

			original := trialStateFile
			trialStateFile = filepath.Join(work, "trial.json")
			t.Cleanup(func() { trialStateFile = original })

			if test.record != "" {
				if err := os.WriteFile(trialStateFile, []byte(test.record), 0o600); err != nil {
					t.Fatalf("failed to write the record: %s", err)
				}
			}

			// A stub for the boot script that records that it ran, in place of
			// the real one, which this machine does not have.
			marker := filepath.Join(work, "reverted")
			stub := filepath.Join(work, "S30eth")
			body := "#!/bin/sh\necho \"$1\" > " + marker + "\n"
			if err := os.WriteFile(stub, []byte(body), 0o755); err != nil {
				t.Fatalf("failed to write the stub: %s", err)
			}

			script := strings.Replace(
				detachedRevertScript("token-abc", 0), ethInitScript, stub, 1)
			if err := runShell(script); err != nil {
				t.Fatalf("the script failed: %s", err)
			}

			data, err := os.ReadFile(marker)
			reverted := err == nil
			if reverted != test.wantRevert {
				t.Fatalf("reverted = %v, want %v", reverted, test.wantRevert)
			}
			if reverted && strings.TrimSpace(string(data)) != "start" {
				t.Errorf("the boot script was run with %q, want \"start\"", strings.TrimSpace(string(data)))
			}

			_, recordLeft := readTrialState()
			if test.wantRevert && recordLeft {
				t.Error("the record was left behind, so a second shell would revert again")
			}
		})
	}
}
