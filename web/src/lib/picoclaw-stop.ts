// Stopping a running PicoClaw turn from the sidebar.
//
// The Stop button sends session/cancel. The agent ends the turn with a
// turn_done whose stopReason is "cancelled", and that turn_done is what puts
// the sidebar back to idle, as for any other turn. The stopper only covers the
// two ways that can fail to happen: the command cannot be sent because the
// chat socket is not open, or the agent never confirms. In both cases it
// reports the failure so the sidebar can show an error and leave the running
// state instead of spinning forever.
//
// The timer is injectable so a test can drive it.

export const PICOCLAW_STOP_TIMEOUT_MS = 15000;

export type TurnStopFailure = 'not_sent' | 'timeout';

export type TurnStopperOptions = {
  // sendCancel sends session/cancel and reports whether it went out.
  sendCancel: () => boolean;
  // onSent runs once the command went out.
  onSent: () => void;
  // onFailed runs when the command could not be sent, or when the turn was
  // still running timeoutMs after it was.
  onFailed: (reason: TurnStopFailure) => void;
  // onPendingChange reports whether a stop is waiting for the agent.
  onPendingChange?: (pending: boolean) => void;
  timeoutMs?: number;
  setTimer?: (cb: () => void, ms: number) => () => void;
};

const defaultSetTimer = (cb: () => void, ms: number) => {
  const timer = setTimeout(cb, ms);
  return () => clearTimeout(timer);
};

export function createTurnStopper(options: TurnStopperOptions) {
  const timeoutMs = options.timeoutMs ?? PICOCLAW_STOP_TIMEOUT_MS;
  const setTimer = options.setTimer ?? defaultSetTimer;
  let clearTimer: (() => void) | null = null;

  const setPending = (next: (() => void) | null) => {
    const was = clearTimer !== null;
    clearTimer = next;
    if (was !== (next !== null)) {
      options.onPendingChange?.(next !== null);
    }
  };

  return {
    // stop sends the cancel. It does nothing while an earlier stop is still
    // waiting, and returns whether a cancel went out.
    stop() {
      if (clearTimer) {
        return false;
      }
      if (!options.sendCancel()) {
        options.onFailed('not_sent');
        return false;
      }
      options.onSent();
      setPending(
        setTimer(() => {
          setPending(null);
          options.onFailed('timeout');
        }, timeoutMs)
      );
      return true;
    },

    // settle is called when the turn is no longer running, however it ended.
    settle() {
      if (!clearTimer) {
        return;
      }
      clearTimer();
      setPending(null);
    },

    isPending() {
      return clearTimer !== null;
    }
  };
}

export type TurnStopper = ReturnType<typeof createTurnStopper>;
