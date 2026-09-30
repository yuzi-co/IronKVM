// pollWhileVisible calls fn every ms while the page is visible. A hidden tab
// skips its turns, and coming back to it calls fn at once, so what fn shows is
// current when it is looked at and the board is left alone when it is not.
// It returns a function that stops the polling.
//
// It is the plain version of usePoll, for pollers that live outside a
// component (a jotai atom's onMount). The timer and document are injectable
// so tests can drive it.
export type PollEnv = {
  isVisible: () => boolean;
  onVisibilityChange: (cb: () => void) => () => void;
  setInterval: (cb: () => void, ms: number) => () => void;
};

export const browserPollEnv: PollEnv = {
  isVisible: () => document.visibilityState === 'visible',
  onVisibilityChange: (cb) => {
    document.addEventListener('visibilitychange', cb);
    return () => document.removeEventListener('visibilitychange', cb);
  },
  setInterval: (cb, ms) => {
    const timer = window.setInterval(cb, ms);
    return () => window.clearInterval(timer);
  }
};

export function pollWhileVisible(fn: () => void, ms: number, env: PollEnv = browserPollEnv) {
  const tick = () => {
    if (env.isVisible()) fn();
  };
  const stopTimer = env.setInterval(tick, ms);
  const stopListening = env.onVisibilityChange(tick);
  return () => {
    stopTimer();
    stopListening();
  };
}
