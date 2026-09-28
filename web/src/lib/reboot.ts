import * as api from '@/api/vm.ts';

import { REBOOT_DOWN_MS, REBOOT_UP_MS, reloadAfterRestart } from './wait-server.ts';

// rebootAndReload asks the board to reboot and reloads the page once it is
// back. The board often goes down before it answers, so a request that fails
// is no reason to stop watching; only an answer that says it refused is, and
// onFailed then gets the server's message.
export function rebootAndReload(onFailed: (msg: string) => void) {
  const abort = new AbortController();

  api
    .reboot()
    .then((rsp) => {
      if (rsp.code !== 0) {
        abort.abort();
        onFailed(rsp.msg);
      }
    })
    .catch((err) => {
      console.log(err);
    });

  reloadAfterRestart(REBOOT_DOWN_MS, REBOOT_UP_MS, abort.signal);
}
