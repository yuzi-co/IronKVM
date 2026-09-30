import { useEffect, useState } from 'react';
import { Switch, Tooltip } from 'antd';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/application.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';

interface PreviewProps {
  checkForUpdates: () => void;
  disabled?: boolean;
}

export const Preview = ({ checkForUpdates, disabled = false }: PreviewProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(true);
  const [isEnabled, setIsEnabled] = useState(false);

  useEffect(() => {
    api
      .getPreviewUpdates()
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }

        setIsEnabled(rsp.data.enabled);
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  function setPreviewUpdates() {
    if (isLoading) return;
    setIsLoading(true);

    const enable = !isEnabled;

    api
      .setPreviewUpdates(enable)
      .then((rsp) => {
        if (!showResult(rsp)) return;

        setIsEnabled(enable);
        checkForUpdates();
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center space-x-2">
          <span>{t('settings.update.preview')}</span>

          <Tooltip
            title={t('settings.update.previewTip')}
            className="cursor-pointer"
            placement="top"
            styles={{ root: { maxWidth: '400px' } }}
          >
            <CircleHelpIcon className="text-neutral-500" size={14} />
          </Tooltip>
        </div>

        <span className="text-xs text-neutral-500">
          {disabled
            ? t('settings.update.customServer.previewDisabled')
            : t('settings.update.previewDesc')}
        </span>
      </div>

      <Switch
        checked={isEnabled}
        loading={isLoading}
        disabled={disabled}
        onChange={setPreviewUpdates}
      />
    </div>
  );
};
