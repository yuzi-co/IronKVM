import { useEffect, useState } from 'react';
import { Popover, Switch, Tooltip } from 'antd';
import { CheckIcon, MoveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import {
  jigglerChoiceFrom,
  jigglerChoices,
  jigglerMethodOf,
  type JigglerChoice
} from '@/lib/jiggler-method.ts';

type JigglerProps = {
  // Changes each time the mouse menu opens, so the switch shows what the
  // device runs now rather than what it ran when the menu first opened.
  refreshKey: number;
};

const choiceLabels: Record<JigglerChoice, string> = {
  mouse: 'mouse.jigglerMouse',
  f15: 'mouse.jigglerF15',
  shift: 'mouse.jigglerShift',
  ctrl: 'mouse.jigglerCtrl'
};

// Jiggler turns the jiggler on and off during a session and picks what it
// does: move the mouse, or press a key the host ignores. The mode (relative or
// absolute) is a setting; the server keeps it while the jiggler is off, so
// turning it back on here names the mode it already has. The method is kept
// the same way, so it can be picked while the jiggler is off.
export const Jiggler = ({ refreshKey }: JigglerProps) => {
  const { t } = useTranslation();

  const [enabled, setEnabled] = useState(false);
  const [mode, setMode] = useState('relative');
  const [choice, setChoice] = useState<JigglerChoice>('mouse');
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
        setChoice(jigglerChoiceFrom(rsp.data.method, rsp.data.key));
      })
      .catch((err) => active && showFailure(err))
      .finally(() => active && setLoadedKey(refreshKey));

    return () => {
      active = false;
    };
  }, [refreshKey]);

  function save(nextEnabled: boolean, nextChoice: JigglerChoice) {
    if (isLoading) return;
    setIsSaving(true);

    const { method, key } = jigglerMethodOf(nextChoice);
    api
      .setMouseJiggler(nextEnabled, mode, method, key)
      .then((rsp) => {
        if (!showResult(rsp)) return;
        setEnabled(nextEnabled);
        setChoice(nextChoice);
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsSaving(false));
  }

  const content = (
    <>
      {jigglerChoices.map((value) => {
        const item = (
          <button
            type="button"
            key={value}
            disabled={isLoading}
            className="flex w-full cursor-pointer items-center space-x-1 rounded p-0 py-1.5 pr-5 pl-2 text-left hover:bg-neutral-700/70 disabled:cursor-wait"
            onClick={() => value !== choice && save(enabled, value)}
          >
            <div className="flex h-[16px] w-[16px] items-end text-blue-500">
              {value === choice && <CheckIcon size={14} />}
            </div>
            <span>{t(choiceLabels[value])}</span>
          </button>
        );
        return value === 'f15' ? (
          <Tooltip key={value} title={t('mouse.jigglerF15Tip')} placement="right">
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
          <MoveIcon size={18} />
          <span className="select-none">{t('mouse.jiggler')}</span>
          <span className="ml-auto pl-3 text-xs whitespace-nowrap text-neutral-500">
            {t(choiceLabels[choice])}
          </span>
        </div>
      </Popover>
      <Switch
        size="small"
        checked={enabled}
        loading={isLoading}
        onChange={(value) => save(value, choice)}
      />
    </div>
  );
};
