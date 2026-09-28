import { useEffect, useState } from 'react';
import { Button, Progress } from 'antd';
import { useAtom } from 'jotai';
import { XIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { cancelPaste } from '@/api/hid.ts';
import { pasteStatusAtom } from '@/jotai/paste.ts';

// How long a finished paste stays on screen. A failure stays until closed.
const finishedVisibleMs = 3000;

export const PasteProgress = () => {
  const { t } = useTranslation();
  const [status, setStatus] = useAtom(pasteStatusAtom);
  const [isCanceling, setIsCanceling] = useState(false);

  useEffect(() => {
    if (status?.status !== 'done' && status?.status !== 'canceled') return;

    const timer = setTimeout(() => setStatus(null), finishedVisibleMs);
    return () => clearTimeout(timer);
  }, [status, setStatus]);

  if (!status || status.status === 'idle') return null;

  function cancel() {
    if (isCanceling) return;
    setIsCanceling(true);

    cancelPaste()
      .then((rsp) => {
        if (rsp.data) setStatus(rsp.data);
      })
      .catch(() => {
        // The next status poll tells what happened.
      })
      .finally(() => setIsCanceling(false));
  }

  const isTyping = status.status === 'typing';
  const percent = status.total > 0 ? Math.floor((status.typed * 100) / status.total) : 0;

  const titles = {
    typing: t('keyboard.pasting.typing'),
    done: t('keyboard.pasting.done'),
    canceled: t('keyboard.pasting.canceled'),
    failed: t('keyboard.pasting.failed')
  };
  const errors = {
    control_busy: t('keyboard.pasting.controlBusy'),
    hid_error: t('keyboard.pasting.hidError')
  };

  return (
    <div className="fixed bottom-6 left-1/2 z-1000 w-[340px] max-w-[calc(100vw-32px)] -translate-x-1/2 rounded-lg bg-neutral-800/95 px-4 py-3 text-sm text-white shadow-lg">
      <div className="flex items-center justify-between gap-3">
        <span>{titles[status.status]}</span>
        {isTyping ? (
          <Button size="small" loading={isCanceling} onClick={cancel}>
            {t('keyboard.pasting.cancel')}
          </Button>
        ) : (
          <XIcon
            size={16}
            className="cursor-pointer text-neutral-400 hover:text-white"
            onClick={() => setStatus(null)}
          />
        )}
      </div>

      <Progress
        percent={percent}
        size="small"
        showInfo={false}
        status={
          status.status === 'failed'
            ? 'exception'
            : status.status === 'done'
              ? 'success'
              : isTyping
                ? 'active'
                : 'normal'
        }
      />

      <div className="text-xs text-neutral-400">
        {status.typed} / {status.total}
        {status.error && <span className="pl-2">{errors[status.error]}</span>}
      </div>
    </div>
  );
};
