import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, message } from 'antd';
import { useAtomValue, useSetAtom } from 'jotai';
import { PowerIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/power-confirm.ts';
import type { PowerConfirmation } from '@/api/power-confirm.ts';
import { showFailure } from '@/lib/feedback.ts';
import { picoclawChatOpenAtom } from '@/jotai/picoclaw.ts';
import {
  dropPowerConfirmationAtom,
  pendingPowerConfirmationsAtom,
  refreshPowerConfirmations
} from '@/jotai/power-confirm.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { pressKeys, secondsLeft, visiblePresses } from './model.ts';

// PowerConfirmPrompt shows the power and reset presses the on-board agent
// asked for, which the server holds until a person confirms one here. The
// chat panel shows them inline while it is open; otherwise they float over
// the desktop, since the user may be talking to the agent through another
// channel, such as Telegram, that cannot confirm a press. Only the roles that
// may press power see them; a viewer never does.
export const PowerConfirmPrompt = ({ inline = false }: { inline?: boolean }) => {
  const role = useAuth().account.role;
  const supported = useServerFeature('agent-power-confirm');
  if (role === 'viewer' || supported !== true) return null;
  return <PendingPresses inline={inline} />;
};

const PendingPresses = ({ inline }: { inline: boolean }) => {
  const list = useAtomValue(pendingPowerConfirmationsAtom);
  const chatOpen = useAtomValue(picoclawChatOpenAtom);
  const [now, setNow] = useState(() => Date.now());
  const shown = visiblePresses(list, now);
  const waiting = list.length > 0;

  // The countdown, while anything waits.
  useEffect(() => {
    if (!waiting) return;
    const tick = () => setNow(Date.now());
    const first = window.setTimeout(tick, 0);
    const timer = window.setInterval(tick, 1000);
    return () => {
      window.clearTimeout(first);
      window.clearInterval(timer);
    };
  }, [waiting]);

  if (!inline && chatOpen) return null;
  if (shown.length === 0) return null;

  const cards = shown.map((c) => <PressCard key={c.id} confirmation={c} now={now} />);
  if (inline) {
    return <div className="flex flex-col gap-2 px-3 pt-3">{cards}</div>;
  }
  return (
    <div className="fixed top-14 left-1/2 z-1000 flex w-[min(92vw,380px)] -translate-x-1/2 flex-col gap-2">
      {cards}
    </div>
  );
};

const PressCard = ({ confirmation, now }: { confirmation: PowerConfirmation; now: number }) => {
  const { t } = useTranslation();
  const drop = useSetAtom(dropPowerConfirmationAtom);
  const [busy, setBusy] = useState<'confirm' | 'deny' | null>(null);
  const keys = pressKeys(confirmation);

  function answer(confirm: boolean) {
    if (busy) return;
    setBusy(confirm ? 'confirm' : 'deny');
    const call = confirm ? api.confirmPowerConfirmation : api.denyPowerConfirmation;
    call(confirmation.id)
      .then((rsp) => {
        // Answered or not, the press no longer waits: the server pressed,
        // refused it, or it had already ended.
        drop(confirmation.id);
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        message.success(t(confirm ? 'power.agentConfirm.pressed' : 'power.agentConfirm.denied'));
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setBusy(null);
        refreshPowerConfirmations();
      });
  }

  return (
    <div
      role="alertdialog"
      aria-label={t('power.agentConfirm.title')}
      className="rounded-lg border border-amber-500/60 bg-neutral-900/95 p-3 text-sm text-neutral-200 shadow-lg"
    >
      <div className="flex items-center gap-2 font-medium">
        <PowerIcon size={16} className="shrink-0 text-amber-400" />
        <span>{t('power.agentConfirm.title')}</span>
      </div>
      <div className="mt-2 font-semibold">{t(keys.label)}</div>
      <div className="text-xs text-neutral-400">{t(keys.desc)}</div>
      <div className="mt-2 text-xs text-neutral-400">{t('power.agentConfirm.hint')}</div>
      <div className="mt-3 flex items-center justify-between gap-2">
        <span className="text-xs text-neutral-500">
          {t('power.agentConfirm.expires', {
            seconds: secondsLeft(confirmation.expiresAt, now)
          })}
        </span>
        <div className="flex gap-2">
          <Button
            size="small"
            disabled={busy !== null}
            loading={busy === 'deny'}
            onClick={() => answer(false)}
          >
            {t('power.agentConfirm.deny')}
          </Button>
          <Button
            size="small"
            type="primary"
            danger
            disabled={busy !== null}
            loading={busy === 'confirm'}
            onClick={() => answer(true)}
          >
            {t('power.agentConfirm.confirm')}
          </Button>
        </div>
      </div>
    </div>
  );
};
