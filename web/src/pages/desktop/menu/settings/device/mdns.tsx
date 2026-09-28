import { useEffect, useState } from 'react';
import { Switch, Tooltip } from 'antd';
import { CircleAlertIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';

export const Mdns = () => {
  const { t } = useTranslation();

  const [isEnabled, setIsEnabled] = useState(false);
  // The request starts on mount, so the control begins in its loading state.
  // Setting the flag inside the effect left one paint where the control was
  // interactive and the value behind it was not yet known.
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api
      .getMdnsState()
      .then((rsp) => {
        if (rsp.data?.enabled) {
          setIsEnabled(true);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  async function update() {
    if (isLoading) return;
    setIsLoading(true);

    const enable = !isEnabled;
    try {
      const rsp = enable ? await api.enableMdns() : await api.disableMdns();
      const success = t(enable ? 'feedback.enabled' : 'feedback.disabled', { name: 'mDNS' });
      if (showResult(rsp, { success })) setIsEnabled(enable);
    } catch (err) {
      showFailure(err);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center space-x-2">
          <span>mDNS</span>

          <Tooltip
            title={t('settings.device.mdns.tip')}
            className="cursor-pointer"
            placement="right"
            styles={{ root: { maxWidth: '400px' } }}
          >
            <CircleAlertIcon className="text-neutral-500" size={14} />
          </Tooltip>
        </div>

        <span className="text-xs text-neutral-500">{t('settings.device.mdns.description')}</span>
      </div>

      <Switch checked={isEnabled} loading={isLoading} onChange={update} />
    </div>
  );
};
