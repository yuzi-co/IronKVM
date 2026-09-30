import { useEffect, useState } from 'react';
import { Select, Tooltip } from 'antd';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';

import { formatBytes } from './format.ts';

type SwapState = { size: number; active: boolean; total: number; used: number };

export const Swap = () => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(true);
  const [size, setSize] = useState('0');
  const [state, setState] = useState<SwapState | null>(null);

  const options = [
    { value: '0', label: t('settings.device.swap.disable') },
    { value: '64', label: '64 MB' },
    { value: '128', label: '128 MB' },
    { value: '256', label: '256 MB' },
    { value: '512', label: '512 MB' }
  ];

  useEffect(() => {
    getSwap();
  }, []);

  function getSwap() {
    setIsLoading(true);
    api
      .getSwap()
      .then((rsp) => {
        if (rsp.code !== 0) return;
        setState(rsp.data);
        setSize((rsp.data?.size ?? 0).toString());
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsLoading(false);
      });
  }

  function update(value: string) {
    if (isLoading) return;
    setIsLoading(true);

    api
      .setSwap(parseInt(value))
      .then((rsp) => {
        if (!showResult(rsp, { success: t('feedback.saved') })) return;

        setSize(value);
      })
      .finally(() => {
        // What the kernel now swaps to is read back, as zram does.
        getSwap();
      });
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center space-x-2">
          <span>{t('settings.device.swap.title')}</span>

          <Tooltip
            title={t('settings.device.swap.tip')}
            className="cursor-pointer"
            placement="right"
            styles={{ root: { maxWidth: '400px' } }}
          >
            <CircleHelpIcon className="text-neutral-500" size={14} />
          </Tooltip>
        </div>
        <span className="text-xs text-neutral-500">{t('settings.device.swap.description')}</span>

        {/* Shown only when a swap file is set, like the zram line beside it. */}
        {state && state.size > 0 && (
          <span className={state.active ? 'text-xs text-neutral-400' : 'text-xs text-amber-500'}>
            {state.active
              ? t('settings.device.swap.active', {
                  used: formatBytes(state.used),
                  total: formatBytes(state.total)
                })
              : t('settings.device.swap.inactive')}
          </span>
        )}
      </div>

      <Select
        style={{ width: 150 }}
        value={size}
        options={options}
        loading={isLoading}
        onChange={update}
      />
    </div>
  );
};
