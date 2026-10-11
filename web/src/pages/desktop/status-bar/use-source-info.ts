import { useEffect, useState } from 'react';

import { getStreamInfo, type StreamInfo } from '@/api/stream.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

// Every two seconds is enough for a mode that changes when the host changes
// it, and the answer is two small file reads and a counter on the board.
const SOURCE_POLL_MS = 2000;

// useSourceInfo is the HDMI source as the board reads it, asked for while the
// status bar is mounted and the tab is visible. Null until the first answer,
// and for good on a server that does not have the endpoint.
export function useSourceInfo(): StreamInfo | null {
  const [info, setInfo] = useState<StreamInfo | null>(null);

  useEffect(() => {
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
  }, []);

  return info;
}
