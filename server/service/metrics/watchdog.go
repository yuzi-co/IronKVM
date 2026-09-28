package metrics

import "NanoKVM-Server/service/watchdog"

// watchdogActions reads the host watchdog's counters, a variable so tests can
// stub it.
var watchdogActions = watchdog.ActionCounts

const helpWatchdogActions = "Presses the host watchdog made on a hung host, by action."

func collectWatchdog(w *Writer) {
	reset, power := watchdogActions()
	w.Counter("ironkvm_watchdog_actions_total", helpWatchdogActions, float64(reset), L("action", watchdog.ActionReset))
	w.Counter("ironkvm_watchdog_actions_total", helpWatchdogActions, float64(power), L("action", watchdog.ActionPower))
}
