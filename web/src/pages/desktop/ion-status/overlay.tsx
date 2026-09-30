import { useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { AlertCircle, AlertTriangle, LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { rebootAndReload } from '@/lib/reboot.ts';

import type { IonStatus } from './model';

// Neutral, not a warning: this is the state before the ION check has an answer,
// not a verdict about the answer. It must read as distinct from both the plain
// black panel (nothing shown) and the red critical gate (something is wrong).
export function IonCheckingIndicator() {
  const { t } = useTranslation();

  return (
    <div className="absolute inset-0 z-10 flex flex-col items-center justify-center gap-2.5 bg-black/60 text-neutral-400">
      <LoaderCircleIcon size={20} className="shrink-0 animate-spin" />
      <span className="text-sm">{t('ion.checking')}</span>
    </div>
  );
}

export function IonWarningBadge({ status }: { status: IonStatus | null }) {
  const { t } = useTranslation();

  if (status?.verdict !== 'warn') {
    return null;
  }

  return (
    <div className="pointer-events-none absolute top-4 left-1/2 z-10 flex max-w-[calc(100%-2rem)] -translate-x-1/2 items-center gap-2.5 rounded-lg border border-amber-500/50 bg-neutral-900/90 px-4 py-2.5 text-sm font-medium text-amber-400 shadow-xl shadow-amber-900/20 backdrop-blur-md">
      <AlertTriangle size={20} className="shrink-0" />
      <span className="min-w-0">{t('ion.warn')}</span>
    </div>
  );
}

export function IonCriticalGate({ onContinue }: { onContinue: () => void }) {
  const { t } = useTranslation();

  const { account } = useAuth();
  // Rebooting the board is an admin action on the server.
  const isAdmin = account.role === 'admin';
  const [isRebooting, setIsRebooting] = useState(false);

  // The board stops answering the moment it reboots, so the request may never
  // resolve and nothing else would tell the operator the click landed. Hold
  // the button in its pending state and reload once the board is back, the
  // same way Settings > Device does it.
  function reboot() {
    if (isRebooting) return;
    setIsRebooting(true);

    rebootAndReload((msg) => {
      console.log(msg);
      setIsRebooting(false);
    });
  }

  return (
    <div className="absolute inset-0 z-10 flex items-center justify-center bg-black/80 p-6">
      <div className="flex max-w-md flex-col gap-4 rounded-lg border border-red-500/50 bg-neutral-900/95 p-6 shadow-xl">
        <div className="flex items-center gap-2.5 text-red-400">
          <AlertCircle size={20} className="shrink-0" />
          <span className="font-medium">{t('ion.criticalTitle')}</span>
        </div>

        <p className="text-sm text-neutral-300">{t('ion.criticalBody')}</p>

        <div className="flex justify-end gap-3">
          <button
            className="rounded px-3 py-1.5 text-sm text-neutral-400 hover:text-neutral-200 disabled:opacity-50"
            disabled={isRebooting}
            onClick={onContinue}
          >
            {t('ion.criticalContinue')}
          </button>
          {isAdmin && (
            <button
              className="flex items-center gap-2 rounded bg-red-600 px-3 py-1.5 text-sm text-white hover:bg-red-500 disabled:cursor-default disabled:bg-red-600/60"
              disabled={isRebooting}
              onClick={reboot}
            >
              {isRebooting && <LoaderCircleIcon size={16} className="shrink-0 animate-spin" />}
              {isRebooting ? t('ion.criticalRebooting') : t('ion.criticalReboot')}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
