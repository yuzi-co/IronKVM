import { useEffect, useState } from 'react';

import { getStreamInfo, type StreamInfo } from '@/api/stream.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

// Every two seconds is enough for a mode that changes when the host changes
// it, and the answer is two small file reads and a counter on the board.
const SOURCE_POLL_MS = 2000;

// useSourceInfo is the HDMI source as the board reads it, asked for while the
// status bar is mounted and the tab is visible. Null until the first answer,
// and for good on a server that does not have the endpoint: one that says so
// in its contract is never asked, and an older one is asked until it answers
// 404.
export function useSourceInfo(): StreamInfo | null {
  const [info, setInfo] = useState<StreamInfo | null>(null);
  const available = useServerFeature('status-bar');

  useEffect(() => {
    if (!available) return;
    let alive = true;
    let missing = false;
    const read = () => {
      if (missing) return;
      getStreamInfo()
        .then((rsp) => {
          if (alive && rsp.code === 0 && rsp.data) setInfo(rsp.data as StreamInfo);
        })
        .catch((error: { response?: { status?: number } }) => {
          // An older server has no such endpoint, and will not grow one
          // while the page is open. Anything else may pass.
          if (error?.response?.status === 404) missing = true;
        });
    };
    read();
    const stop = pollWhileVisible(read, SOURCE_POLL_MS);
    return () => {
      alive = false;
      stop();
    };
  }, [available]);

  return available ? info : null;
}
