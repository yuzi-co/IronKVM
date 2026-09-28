import { useState } from 'react';
import { ReloadOutlined } from '@ant-design/icons';
import { Button, message, Popconfirm } from 'antd';
import { useTranslation } from 'react-i18next';

import { rebootAndReload } from '@/lib/reboot.ts';

export const Reboot = () => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);

  function reboot() {
    if (isLoading) return;
    setIsLoading(true);

    rebootAndReload((msg) => {
      message.error(msg || t('settings.device.rebootFailed'));
      setIsLoading(false);
    });
  }

  return (
    <div className="flex justify-center pt-3">
      <Popconfirm
        placement="bottom"
        title={t('settings.device.rebootDesc')}
        okText={t('settings.device.okBtn')}
        cancelText={t('settings.device.cancelBtn')}
        onConfirm={reboot}
      >
        <Button
          danger
          type="primary"
          size="large"
          shape="round"
          loading={isLoading}
          icon={<ReloadOutlined />}
        >
          {t('settings.device.reboot')}
        </Button>
      </Popconfirm>
    </div>
  );
};
