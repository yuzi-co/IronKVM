import { Tooltip } from 'antd';
import { useAtom } from 'jotai';
import { EyeIcon, XIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { viewOnlyAtom } from '@/jotai/screen.ts';

// ViewOnlyBadge stays on screen while view only is on, so nobody types at a
// host that is not listening and wonders why. A click turns it off.
export const ViewOnlyBadge = () => {
  const { t } = useTranslation();
  const [viewOnly, setViewOnly] = useAtom(viewOnlyAtom);

  if (!viewOnly) return null;

  return (
    <Tooltip title={t('screen.viewOnlyOff')} placement="top">
      <button
        type="button"
        className="fixed bottom-3 left-3 z-1000 flex cursor-pointer items-center space-x-1.5 rounded-full border border-amber-500/60 bg-neutral-900/90 px-3 py-1 text-xs text-amber-400 shadow-lg hover:border-amber-400"
        onClick={() => setViewOnly(false)}
      >
        <EyeIcon size={14} />
        <span>{t('screen.viewOnly')}</span>
        <XIcon size={12} className="text-neutral-400" />
      </button>
    </Tooltip>
  );
};
