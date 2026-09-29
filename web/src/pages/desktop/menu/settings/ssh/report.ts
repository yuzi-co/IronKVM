import { message } from 'antd';

import { showFailure } from '@/lib/feedback.ts';

type Reply = { code: number; msg?: string };

// report shows the outcome of an SSH request. The server's codes that the
// page can explain get a translated line; anything else goes through the
// shared feedback, which shows the server's own reason.
export function report(rsp: Reply, known: Record<number, string>, success?: string): boolean {
  if (rsp.code === 0) {
    if (success) message.success(success);
    return true;
  }
  const text = known[rsp.code];
  if (text) message.error(text);
  else showFailure(rsp);
  return false;
}
