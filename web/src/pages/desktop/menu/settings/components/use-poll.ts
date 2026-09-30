import { useEffect } from 'react';

import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

// usePoll calls fn every ms while active and the tab is visible. A hidden tab
// skips its turns, and coming back to it polls at once, so the page is current
// when it is looked at and quiet when it is not.
export function usePoll(fn: () => void, ms: number, active = true) {
  const tick = useStableCallback(fn);

  useEffect(() => {
    if (!active) return;
    return pollWhileVisible(tick, ms);
  }, [tick, ms, active]);
}
