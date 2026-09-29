import { useEffect, useState } from 'react';
import { Switch } from 'antd';
import { MoveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';

type JigglerProps = {
  // Changes each time the mouse menu opens, so the switch shows what the
  // device runs now rather than what it ran when the menu first opened.
  refreshKey: number;
};

// Jiggler turns the mouse jiggler on and off during a session. The mode
// (relative or absolute) is a setting; the server keeps it while the jiggler
// is off, so turning it back on here names the mode it already has.
export const Jiggler = ({ refreshKey }: JigglerProps) => {
  const { t } = useTranslation();

  const [enabled, setEnabled] = useState(false);
  const [mode, setMode] = useState('relative');
  // The read counts as done once it has answered for the current key, so a
  // new key shows the switch as loading without setting state in the effect.
  const [loadedKey, setLoadedKey] = useState<number | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const isLoading = loadedKey !== refreshKey || isSaving;

  useEffect(() => {
    let active = true;

    api
      .getMouseJiggler()
      .then((rsp) => {
        if (!active) return;
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }

        setEnabled(rsp.data.enabled);
        if (rsp.data.mode) setMode(rsp.data.mode);
      })
      .catch((err) => active && showFailure(err))
      .finally(() => active && setLoadedKey(refreshKey));

    return () => {
      active = false;
    };
  }, [refreshKey]);

  function update(value: boolean) {
    if (isLoading) return;
    setIsSaving(true);

    api
      .setMouseJiggler(value, mode)
      .then((rsp) => {
        if (showResult(rsp)) setEnabled(value);
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsSaving(false));
  }

  return (
    <div className="flex h-[30px] items-center justify-between space-x-5 rounded px-3 text-neutral-300">
      <div className="flex items-center space-x-2">
        <MoveIcon size={18} />
        <span className="select-none">{t('mouse.jiggler')}</span>
      </div>
      <Switch size="small" checked={enabled} loading={isLoading} onChange={update} />
    </div>
  );
};
