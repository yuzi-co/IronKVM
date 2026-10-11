import { useState } from 'react';
import { AlertTriangle, XIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { apiSupported, formatRange } from '@/lib/server-contract.ts';
import { useServerContract } from '@/hooks/useServerFeature.ts';

// ServerCompatNotice says so when the server speaks an API version this UI
// does not support. The UI still runs: most of the API survives a version
// bump, and the update page that fixes the mismatch is part of what has to
// keep working. It can be dismissed for the rest of the page's life.
export const ServerCompatNotice = () => {
  const { t } = useTranslation();
  const contract = useServerContract();
  const [dismissed, setDismissed] = useState(false);

  if (!contract || dismissed || apiSupported(contract)) return null;

  return (
    <div
      role="alert"
      className="fixed top-3 left-1/2 z-[2000] flex w-[min(640px,calc(100vw-2rem))] -translate-x-1/2 items-start gap-2.5 rounded-lg border border-amber-500/50 bg-neutral-900/95 px-4 py-3 text-sm text-amber-300 shadow-xl shadow-amber-900/20 backdrop-blur-md"
    >
      <AlertTriangle size={20} className="mt-0.5 shrink-0" />
      <div className="min-w-0 flex-1">
        <div className="font-medium">{t('serverContract.incompatibleTitle')}</div>
        <div className="mt-1 text-neutral-300">
          {t('serverContract.incompatible', { ui: formatRange(), server: contract.api })}
        </div>
      </div>
      <button
        type="button"
        aria-label={t('serverContract.dismiss')}
        className="shrink-0 cursor-pointer text-neutral-400 hover:text-neutral-200"
        onClick={() => setDismissed(true)}
      >
        <XIcon size={16} />
      </button>
    </div>
  );
};
