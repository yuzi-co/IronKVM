import { useEffect, useState } from 'react';
import { Button, Checkbox, InputNumber, message, Popover, Select, Tooltip } from 'antd';
import { useAtom, useSetAtom } from 'jotai';
import { RepeatIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/sequence.ts';
import { showFailure } from '@/lib/feedback.ts';
import { getAutoTapSettings, setAutoTapSettings } from '@/lib/localstorage.ts';
import {
  AUTO_TAP_INTERVAL,
  AUTO_TAP_KEYS,
  AUTO_TAP_PRESETS,
  AUTO_TAP_TIMEOUT,
  autoTapSettingsFrom,
  keyLabel,
  usageOf,
  type AutoTapSettings
} from '@/lib/macro.ts';
import { keyboardLockAtom } from '@/jotai/keyboard.ts';
import { sequenceStatusAtom } from '@/jotai/sequence.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

function label(code: string): string {
  const usage = usageOf(code);
  return usage === undefined ? code : keyLabel(usage);
}

// AutoTap presses one key again and again until the firmware menu opens: F2
// or Del for setup, F12, F11, F10 or Esc for the boot menu. The server does
// the pressing, so it goes on when this tab closes, and every press is
// released before the next. The sequence banner shows the run and stops it.
export const AutoTap = () => {
  const { t } = useTranslation();
  const supported = useServerFeature('hid-autotap');

  const [isOpen, setIsOpen] = useState(false);
  const [settings, setSettings] = useState<AutoTapSettings>(() =>
    autoTapSettingsFrom(getAutoTapSettings())
  );
  const [status, setStatus] = useAtom(sequenceStatusAtom);
  const [isBusy, setIsBusy] = useState(false);
  const setKeyboardLock = useSetAtom(keyboardLockAtom);

  // While the form is open, digits typed into it must not reach the host.
  useEffect(() => {
    setKeyboardLock({ source: 'auto-tap', locked: isOpen });
    return () => setKeyboardLock({ source: 'auto-tap', locked: false });
  }, [isOpen, setKeyboardLock]);

  if (!supported) return null;

  const running = status?.running === true;
  const isOther = !AUTO_TAP_PRESETS.includes(settings.code);

  function update(change: Partial<AutoTapSettings>) {
    const next = { ...settings, ...change };
    setSettings(next);
    setAutoTapSettings(next);
  }

  function start() {
    const usage = usageOf(settings.code);
    if (isBusy || usage === undefined) return;
    setIsBusy(true);
    api
      .startAutoTap({
        label: label(settings.code),
        modifiers: 0,
        keys: [usage],
        intervalMs: settings.interval,
        timeoutS: settings.timeout,
        stopOnInput: settings.stopOnInput,
        stopOnVideo: settings.stopOnVideo,
        atx: settings.firstPress
      })
      .then((rsp) => {
        if (rsp.code === -3) {
          message.error(t('keyboard.sequence.alreadyRunning'));
          return;
        }
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        setStatus(rsp.data);
        setIsOpen(false);
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsBusy(false));
  }

  function stop() {
    if (isBusy) return;
    setIsBusy(true);
    api
      .stopSequence()
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        setStatus(rsp.data);
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsBusy(false));
  }

  const keyButton = (code: string) => (
    <button
      type="button"
      key={code}
      className={`h-7 min-w-[44px] cursor-pointer rounded border px-2 text-xs ${
        settings.code === code
          ? 'border-blue-500 bg-blue-500/20 text-blue-200'
          : 'border-neutral-700 text-neutral-300 hover:bg-neutral-700/70'
      }`}
      onClick={() => update({ code })}
    >
      {label(code)}
    </button>
  );

  const content = (
    <div className="flex w-[300px] flex-col space-y-3 p-1 text-sm text-neutral-300">
      <div className="text-xs text-neutral-400">{t('keyboard.autoTap.tip')}</div>

      <div className="flex flex-col space-y-1.5">
        <span className="text-xs text-neutral-400">{t('keyboard.autoTap.key')}</span>
        <div className="flex flex-wrap gap-1">{AUTO_TAP_PRESETS.map(keyButton)}</div>
        <Select
          size="small"
          showSearch
          value={isOther ? settings.code : undefined}
          placeholder={t('keyboard.autoTap.otherKey')}
          options={AUTO_TAP_KEYS.map((code) => ({ value: code, label: label(code) }))}
          optionFilterProp="label"
          onChange={(code: string) => update({ code })}
        />
      </div>

      <div className="flex space-x-3">
        <label className="flex flex-1 flex-col space-y-1">
          <span className="text-xs text-neutral-400">{t('keyboard.autoTap.interval')}</span>
          <InputNumber
            size="small"
            min={AUTO_TAP_INTERVAL.min}
            max={AUTO_TAP_INTERVAL.max}
            step={50}
            value={settings.interval}
            onChange={(value) => value !== null && update({ interval: value })}
          />
        </label>
        <label className="flex flex-1 flex-col space-y-1">
          <span className="text-xs text-neutral-400">{t('keyboard.autoTap.timeout')}</span>
          <InputNumber
            size="small"
            min={AUTO_TAP_TIMEOUT.min}
            max={AUTO_TAP_TIMEOUT.max}
            value={settings.timeout}
            onChange={(value) => value !== null && update({ timeout: value })}
          />
        </label>
      </div>

      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('keyboard.autoTap.firstPress')}</span>
        <Select
          size="small"
          value={settings.firstPress}
          options={[
            { value: '', label: t('keyboard.autoTap.firstPressNone') },
            { value: 'power', label: t('keyboard.autoTap.firstPressPower') },
            { value: 'reset', label: t('keyboard.autoTap.firstPressReset') }
          ]}
          onChange={(firstPress: AutoTapSettings['firstPress']) => update({ firstPress })}
        />
      </label>

      <div className="flex flex-col space-y-1">
        <Checkbox
          checked={settings.stopOnInput}
          onChange={(e) => update({ stopOnInput: e.target.checked })}
        >
          {t('keyboard.autoTap.stopOnInput')}
        </Checkbox>
        <Tooltip title={t('keyboard.autoTap.stopOnVideoTip')} placement="right">
          <Checkbox
            checked={settings.stopOnVideo}
            onChange={(e) => update({ stopOnVideo: e.target.checked })}
          >
            {t('keyboard.autoTap.stopOnVideo')}
          </Checkbox>
        </Tooltip>
      </div>

      {running ? (
        <Button danger loading={isBusy} onClick={stop}>
          {t('keyboard.autoTap.stop')}
        </Button>
      ) : (
        <Button type="primary" loading={isBusy} onClick={start}>
          {t('keyboard.autoTap.start')}
        </Button>
      )}
    </div>
  );

  return (
    <Popover
      content={content}
      trigger="click"
      placement="rightTop"
      align={{ offset: [14, 0] }}
      open={isOpen}
      onOpenChange={setIsOpen}
      arrow={false}
    >
      <div className="flex cursor-pointer items-center space-x-2 rounded py-1 pr-5 pl-2 select-none hover:bg-neutral-700/70">
        <RepeatIcon size={18} />
        <span>{t('keyboard.autoTap.title')}</span>
        {running && status?.kind === 'autotap' && (
          <span className="ml-auto pl-3 text-xs text-blue-400">{status.name}</span>
        )}
      </div>
    </Popover>
  );
};
