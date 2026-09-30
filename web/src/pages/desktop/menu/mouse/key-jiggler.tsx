import { useEffect, useState } from 'react';
import { Popover, Switch, Tooltip } from 'antd';
import { CheckIcon, KeyboardIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import { keyJigglerKeyFrom, keyJigglerKeys, type KeyJigglerKey } from '@/lib/key-jiggler.ts';

type KeyJigglerProps = {
  // Changes each time the mouse menu opens, so the row shows what the device
  // runs now rather than what it ran when the menu first opened.
  refreshKey: number;
};

const keyLabels: Record<KeyJigglerKey, string> = {
  f15: 'mouse.keyJigglerF15',
  shift: 'mouse.keyJigglerShift',
  ctrl: 'mouse.keyJigglerCtrl'
};

// KeyJiggler turns the key jiggler on and off and picks the key it presses. It
// is separate from the mouse jiggler above it; both can be on, and both wait
// for the same idle time. The server keeps the key while the key jiggler is
// off, so it can be picked before turning it on.
export const KeyJiggler = ({ refreshKey }: KeyJigglerProps) => {
  const { t } = useTranslation();

  const [enabled, setEnabled] = useState(false);
  const [key, setKey] = useState<KeyJigglerKey>('f15');
  // The read counts as done once it has answered for the current refresh, so
  // a new one shows the switch as loading without setting state in the effect.
  const [loadedKey, setLoadedKey] = useState<number | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const isLoading = loadedKey !== refreshKey || isSaving;

  useEffect(() => {
    let active = true;

    api
      .getKeyJiggler()
      .then((rsp) => {
        if (!active) return;
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }

        setEnabled(rsp.data.enabled);
        setKey(keyJigglerKeyFrom(rsp.data.key));
      })
      .catch((err) => active && showFailure(err))
      .finally(() => active && setLoadedKey(refreshKey));

    return () => {
      active = false;
    };
  }, [refreshKey]);

  function save(nextEnabled: boolean, nextKey: KeyJigglerKey) {
    if (isLoading) return;
    setIsSaving(true);

    api
      .setKeyJiggler(nextEnabled, nextKey)
      .then((rsp) => {
        if (!showResult(rsp)) return;
        setEnabled(nextEnabled);
        setKey(nextKey);
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsSaving(false));
  }

  const content = (
    <>
      {keyJigglerKeys.map((value) => {
        const item = (
          <button
            type="button"
            key={value}
            disabled={isLoading}
            className="flex w-full cursor-pointer items-center space-x-1 rounded p-0 py-1.5 pr-5 pl-2 text-left hover:bg-neutral-700/70 disabled:cursor-wait"
            onClick={() => value !== key && save(enabled, value)}
          >
            <div className="flex h-[16px] w-[16px] items-end text-blue-500">
              {value === key && <CheckIcon size={14} />}
            </div>
            <span>{t(keyLabels[value])}</span>
          </button>
        );
        return value === 'f15' ? (
          <Tooltip key={value} title={t('mouse.keyJigglerF15Tip')} placement="right">
            {item}
          </Tooltip>
        ) : (
          item
        );
      })}
    </>
  );

  return (
    <div className="flex h-[30px] items-center justify-between space-x-5 rounded px-3 text-neutral-300">
      <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
        <div className="flex min-w-0 flex-1 cursor-pointer items-center space-x-2">
          <KeyboardIcon size={18} />
          <span className="select-none">{t('mouse.keyJiggler')}</span>
          <span className="ml-auto pl-3 text-xs whitespace-nowrap text-neutral-500">
            {t(keyLabels[key])}
          </span>
        </div>
      </Popover>
      <Switch
        size="small"
        checked={enabled}
        loading={isLoading}
        onChange={(value) => save(value, key)}
      />
    </div>
  );
};
