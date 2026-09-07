package network

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// The trial timer in ethernet.go lives in the server process, and it is the
// only thing that puts a wrong address back. That is enough for the case it
// was written for, a person who applies a change and then cannot reach the
// board, because the board keeps running and the timer keeps counting.
//
// It is not enough for a server that stops counting. If the process dies
// between the apply and the confirmation, the interface stays on the trial
// address and nothing is left to revert it. The supervisor does not help: it
// probes http://127.0.0.1/ and sees a healthy server, so it never restarts and
// never escalates to a reboot. A reboot would fix it, because a trial writes
// nothing to /boot, but somebody has to reach the board to cause one. This is
// out-of-band management hardware in an enclosure with no remote power cycle,
// so "somebody walks to it" is the whole cost of the failure.
//
// So the deadline is also held outside the process. Starting a trial writes
// the state to /run and starts a detached shell that sleeps past the deadline
// and runs the boot script if the state is still there. Confirming removes the
// state, and so does the in-process revert, and the detached shell then wakes
// up and exits without doing anything.
//
// /run is tmpfs, so a reboot clears the state on its own. That is the right
// behaviour rather than a limitation: the trial address was never written to
// /boot, so a board that reboots comes back on the saved configuration and
// there is nothing left to revert.

const (
	// externalRevertGrace is how long the detached revert waits past the
	// deadline. The in-process timer should win whenever the server is alive,
	// because it reverts and clears the state in one place. The grace is what
	// keeps the two from racing on a board that is merely busy.
	externalRevertGrace = 15 * time.Second
)

// trialStateFile is a var so the tests can point it somewhere they may write.
var trialStateFile = "/run/kvm-ethernet-trial.json"

// trialState is what survives the process. It carries enough to answer
// GetEthernet and to accept a confirmation, so a server that restarts during a
// trial does not report that nothing is pending while a detached revert is
// still counting down.
type trialState struct {
	Token    string    `json:"token"`
	Mode     string    `json:"mode"`
	Address  string    `json:"address"`
	Prefix   int       `json:"prefix"`
	Gateway  string    `json:"gateway"`
	Deadline time.Time `json:"deadline"`
	// Applied says the interface took the change. The in-process flag of the
	// same name does not survive the process, and a trial adopted from this
	// record would otherwise be assumed to have been applied: a server that
	// died between writing this record and finishing the apply would let the
	// next request confirm a change the interface never took.
	Applied bool `json:"applied"`
}

func writeTrialState(state trialState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Written whole. A detached revert reads this file, and a half-written one
	// would read as a trial that does not match its own token, which is the
	// one case where the revert must not be skipped.
	temporary := trialStateFile + ".new"
	if err := os.WriteFile(temporary, append(data, '\n'), 0o600); err != nil {
		return err
	}

	return os.Rename(temporary, trialStateFile)
}

func readTrialState() (trialState, bool) {
	data, err := os.ReadFile(trialStateFile)
	if err != nil {
		return trialState{}, false
	}

	var state trialState
	if err := json.Unmarshal(data, &state); err != nil {
		return trialState{}, false
	}
	if state.Token == "" {
		return trialState{}, false
	}

	return state, true
}

func clearTrialState() {
	_ = os.Remove(trialStateFile)
	_ = os.Remove(trialStateFile + ".new")
}

// startExternalRevert leaves a process behind that outlives this one.
//
// Setsid is what makes it outlive the server rather than merely outlive the
// request. The child gets its own session, so it is not in the server's
// process group and a signal sent to that group does not reach it, and
// S95NanoKVM's stop path kills the server by name.
//
// It is a var so the tests can record the arguments without leaving a sleeping
// shell behind on the machine that runs them.
var startExternalRevert = func(token string, after time.Duration) error {
	seconds := int(after.Seconds())
	if seconds < 1 {
		seconds = 1
	}

	command := exec.Command("sh", "-c", detachedRevertScript(token, seconds))
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return err
	}

	// Reaped here so a long-running server does not collect one zombie per
	// trial. If the server exits first the child is reparented to init, which
	// reaps it instead, and the sleep is unaffected either way.
	go func() {
		_ = command.Wait()
	}()

	log.Infof("ethernet trial %s has a detached revert in %ds", token, seconds)

	return nil
}

// detachedRevertScript is what the detached shell runs.
//
// The token is in the test and not only in the file name, so a second trial
// that rewrites the record retires the first revert: the older shell wakes up,
// does not find its own token, and exits. The record is removed before the
// boot script runs, so two shells that somehow wake together still produce one
// revert.
func detachedRevertScript(token string, seconds int) string {
	return fmt.Sprintf(
		"sleep %d; grep -q %s %s 2>/dev/null || exit 0; rm -f %s; %s start",
		seconds, token, trialStateFile, trialStateFile, ethInitScript,
	)
}
