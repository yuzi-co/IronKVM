import { useEffect, useState } from 'react';
import { Switch, Tooltip } from 'antd';
import { CircleAlertIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { logout } from '@/api/auth.ts';
import * as api from '@/api/vm.ts';

// The device restarts its server to change scheme, and on this hardware that
// takes about two minutes. The switch used to wait a flat 30 seconds and then
// reload, which lands on a server that is still stopped.
const PROBE_INTERVAL_MS = 2000;
const GOING_DOWN_MS = 90_000;
const COMING_UP_MS = 300_000;

// After the https listener stops there is no way to watch the http one from
// this page: a secure page may not fetch plain http, so the check is blocked
// rather than failed. This is the one wait that stays a guess, and the message
// says so.
const HTTP_GRACE_MS = 45_000;

// targetUrl names where the operator has to end up. Default ports only, which
// is what the device's own redirect uses, and the page cannot read the
// configured ports anyway.
function targetUrl(enable: boolean): string {
  return `${enable ? 'https:' : 'http:'}//${window.location.hostname}/`;
}

// reachable answers whether this origin is still serving. redirect: 'manual'
// keeps it from following the device's own 307 into a scheme whose certificate
// the browser has not been shown yet.
async function reachable(): Promise<boolean> {
  try {
    await fetch(`${window.location.origin}/`, { cache: 'no-store', redirect: 'manual' });
    return true;
  } catch {
    return false;
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// waitUntil polls for the origin to reach a state, and gives up rather than
// waiting for ever. Giving up still navigates: being on the right URL and
// having to reload beats sitting on a page that will never work again.
async function waitUntil(serving: boolean, budgetMs: number): Promise<void> {
  const deadline = Date.now() + budgetMs;

  while (Date.now() < deadline) {
    if ((await reachable()) === serving) return;
    await sleep(PROBE_INTERVAL_MS);
  }
}

export const Tls = () => {
  const { t } = useTranslation();

  const [isEnabled, setIsEnabled] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState('');

  useEffect(() => {
    setIsEnabled(window.location.protocol === 'https:');
  }, []);

  async function update() {
    if (isLoading) return;
    setIsLoading(true);

    const enable = !isEnabled;

    let ok = false;
    try {
      const rsp = await api.setTLS(enable);
      ok = rsp.code === 0;
    } catch (err) {
      console.log(err);
    }

    if (!ok) {
      setIsLoading(false);
      return;
    }

    setIsEnabled(enable);

    // Before the scheme changes, and only here. A cookie written over https
    // carries Secure, and a page served over plain http may neither read nor
    // delete it. Leaving it behind means every later login writes a cookie the
    // browser discards, and the operator is locked out of their own KVM with
    // nothing in the interface able to clear it. The cookie is HttpOnly now, so
    // the page asks the server to clear it while this origin can still do it.
    try {
      await logout();
    } catch {
      // A session that cannot be ended is one the next login replaces anyway.
    }

    setStatus(t('settings.network.tls.restarting'));
    await waitUntil(false, GOING_DOWN_MS);

    if (enable) {
      // Port 80 keeps serving under HTTPS, answering a redirect, so this origin
      // returning is a genuine signal that the server is back.
      setStatus(t('settings.network.tls.waiting'));
      await waitUntil(true, COMING_UP_MS);
    } else {
      setStatus(t('settings.network.tls.waitingHttp'));
      await sleep(HTTP_GRACE_MS);
    }

    window.location.replace(targetUrl(enable));
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center space-x-2">
          <span>HTTPS</span>

          <Tooltip
            title={t('settings.network.tls.tip')}
            className="cursor-pointer"
            placement="right"
            styles={{ root: { maxWidth: '400px' } }}
          >
            <CircleAlertIcon className="text-neutral-500" size={14} />
          </Tooltip>
        </div>
        <span className="text-xs text-neutral-500">
          {status || t('settings.network.tls.description')}
        </span>
      </div>

      <Switch checked={isEnabled} loading={isLoading} disabled={isLoading} onChange={update} />
    </div>
  );
};
