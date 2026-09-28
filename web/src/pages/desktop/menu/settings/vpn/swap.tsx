import { useEffect, useState } from 'react';
import { message, Switch, Tooltip } from 'antd';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';

// The size this switch turns swap on with when the board has none, the one
// its tooltip names. A size chosen in Settings > Device is kept.
const DEFAULT_SIZE_MB = 256;

export const Swap = () => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(true);
  const [isEnabled, setIsEnabled] = useState(false);
  // The size last seen, so turning swap off and on here does not replace the
  // size chosen in Settings > Device.
  const [size, setSize] = useState(DEFAULT_SIZE_MB);

  useEffect(() => {
    api
      .getSwap()
      .then((rsp) => {
        if (rsp.data?.size > 0) {
          setIsEnabled(true);
          setSize(rsp.data.size);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  function update(enable: boolean) {
    if (isLoading) return;
    setIsLoading(true);

    api
      .setSwap(enable ? size : 0)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(rsp.msg || t('settings.vpn.swap.failed'));
          return;
        }

        setIsEnabled(enable);
      })
      .catch(() => message.error(t('settings.vpn.swap.failed')))
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="flex h-[40px] cursor-pointer items-center justify-between space-x-6 rounded px-2 text-neutral-300 hover:bg-neutral-700/70">
      <div className="flex items-center space-x-1">
        <span>{t('settings.vpn.swap.title')}</span>
        <Tooltip
          title={t('settings.vpn.swap.tip')}
          className="cursor-pointer text-neutral-500"
          placement="top"
          styles={{ root: { maxWidth: '400px' } }}
        >
          <CircleHelpIcon size={15} />
        </Tooltip>
      </div>

      <Switch value={isEnabled} loading={isLoading} size="small" onChange={update} />
    </div>
  );
};
