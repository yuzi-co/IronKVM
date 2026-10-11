import { useEffect, useState } from 'react';

import { getNutState, type NutState } from '@/api/nut.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

// The server polls the UPS every five seconds, so asking more often only
// repeats its last answer.
const UPS_POLL_MS = 10 * 1000;

// useUps is the NUT client's state, asked for while the status bar is mounted
// and the tab is visible. Null until the first answer, and for good on a
// server without the client.
export function useUps(): NutState | null {
  const [state, setState] = useState<NutState | null>(null);
  const available = useServerFeature('nut-client');

  useEffect(() => {
    if (!available) return;
    let alive = true;
    const read = () => {
      getNutState()
        .then((rsp) => {
          if (alive && rsp.code === 0 && rsp.data) setState(rsp.data as NutState);
        })
        .catch(() => {
          // The next poll tries again. The bar keeps the last state meanwhile.
        });
    };
    read();
    const stop = pollWhileVisible(read, UPS_POLL_MS);
    return () => {
      alive = false;
      stop();
    };
  }, [available]);

  return available ? state : null;
}
