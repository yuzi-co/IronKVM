import { Alert } from 'antd';
import { useTranslation } from 'react-i18next';

import { vpnTitles } from './format.ts';

type NoticeProps = {
  blockedBy: string;
};

// Notice says, before any button is pressed, that the other VPN runs or
// starts at boot, and that only one runs at a time.
export const Notice = ({ blockedBy }: NoticeProps) => {
  const { t } = useTranslation();
  if (!blockedBy) return null;

  return (
    <Alert
      className="mt-5!"
      type="warning"
      showIcon
      message={t('settings.vpn.blocked', { other: vpnTitles[blockedBy] ?? blockedBy })}
    />
  );
};
