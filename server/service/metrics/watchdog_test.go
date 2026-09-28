package metrics

import "testing"

func TestWatchdogCountsEachAction(t *testing.T) {
	setVar(t, &watchdogActions, func() (uint64, uint64) { return 2, 1 })

	want := `# HELP ironkvm_watchdog_actions_total Presses the host watchdog made on a hung host, by action.
# TYPE ironkvm_watchdog_actions_total counter
ironkvm_watchdog_actions_total{action="reset"} 2
ironkvm_watchdog_actions_total{action="power"} 1
`
	assertText(t, render(t, collectWatchdog), want)
}
