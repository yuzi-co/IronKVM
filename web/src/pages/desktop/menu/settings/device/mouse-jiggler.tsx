import { useEffect, useState } from 'react';
import { Select } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';

// One select holds both the switch and the mode: off, or on in one of the two
// modes. The server keeps the mode while off, so turning it back on from here
// always names the mode it should run in.
type Choice = 'off' | 'relative' | 'absolute';

export const MouseJiggler = () => {
  const { t } = useTranslation();

  const [choice, setChoice] = useState<Choice>('off');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api
      .getMouseJiggler()
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }

        setChoice(rsp.data.enabled ? (rsp.data.mode as Choice) : 'off');
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  const options = [
    { value: 'off', label: t('settings.device.mouseJiggler.disable') },
    { value: 'relative', label: t('settings.device.mouseJiggler.relative') },
    { value: 'absolute', label: t('settings.device.mouseJiggler.absolute') }
  ];

  function update(value: Choice) {
    if (isLoading) return;
    setIsLoading(true);

    const enabled = value !== 'off';
    const mode = enabled ? value : 'relative';

    api
      .setMouseJiggler(enabled, mode)
      .then((rsp) => {
        if (!showResult(rsp)) return;
        setChoice(value);
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <span>{t('settings.device.mouseJiggler.title')}</span>
        <span className="text-xs text-neutral-500">
          {t('settings.device.mouseJiggler.description')}
        </span>
      </div>

      <Select<Choice>
        style={{ width: 150 }}
        value={choice}
        options={options}
        loading={isLoading}
        disabled={isLoading}
        onChange={update}
      />
    </div>
  );
};
