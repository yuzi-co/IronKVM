import { useEffect, useState } from 'react';
import { Button, message } from 'antd';
import { CopyIcon, ExternalLinkIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { writeClipboardText } from '@/lib/clipboard.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import type { Rsp } from './types.ts';

// How often the page asks whether the operator has finished signing in.
const POLL_MS = 4000;

type LoginUrlProps = {
  url: string;
  // A line saying how long the link stays valid.
  period: string;
  getStatus: () => Promise<Rsp>;
  onSuccess: () => void;
};

// LoginUrl shows the sign-in link a VPN add-on hands out, and moves on by
// itself once the operator has signed in.
//
// The link arrives after a request, and a window opened after an await is a
// popup the browser blocks, so the page does not try to open it. The link is
// shown in full, with Copy for a sign-in on another device and Open for this
// one.
export const LoginUrl = ({ url, period, getStatus, onSuccess }: LoginUrlProps) => {
  const { t } = useTranslation();

  const [isChecking, setIsChecking] = useState(false);

  useEffect(() => {
    let stopped = false;

    const stopPoll = pollWhileVisible(() => {
      getStatus()
        .then((rsp) => {
          if (!stopped && rsp.code === 0 && rsp.data?.state !== 'notLogin') {
            stopped = true;
            onSuccess();
          }
        })
        .catch(() => {
          // The next poll tries again.
        });
    }, POLL_MS);

    return () => {
      stopped = true;
      stopPoll();
    };
  }, [getStatus, onSuccess]);

  async function copy() {
    try {
      await writeClipboardText(url);
      message.success(t('settings.vpn.copied'));
    } catch {
      message.error(t('settings.vpn.copyFailed'));
    }
  }

  function check() {
    if (isChecking) return;
    setIsChecking(true);

    getStatus()
      .then((rsp) => {
        if (rsp.code === 0 && rsp.data?.state !== 'notLogin') {
          onSuccess();
          return;
        }
        message.info(t('settings.vpn.notSignedIn'));
      })
      .catch(() => message.error(t('settings.vpn.checkFailed')))
      .finally(() => setIsChecking(false));
  }

  return (
    <div className="flex w-full max-w-[460px] flex-col items-center space-y-4">
      <div className="w-full rounded bg-neutral-800/60 p-3 font-mono text-xs break-all select-all">
        {url}
      </div>

      <div className="flex items-center space-x-2">
        <Button icon={<CopyIcon size={15} />} onClick={copy}>
          {t('settings.vpn.copy')}
        </Button>
        <Button
          type="primary"
          icon={<ExternalLinkIcon size={15} />}
          href={url}
          target="_blank"
          rel="noopener noreferrer"
        >
          {t('settings.vpn.open')}
        </Button>
      </div>

      <span className="text-center text-xs text-neutral-400">
        {period} {t('settings.vpn.loginWaiting')}
      </span>

      <Button shape="round" loading={isChecking} onClick={check}>
        {t('settings.vpn.checkAgain')}
      </Button>
    </div>
  );
};
