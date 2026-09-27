import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/netbird.ts';

import { VpnPage } from '../vpn/page.tsx';
import type { VpnInfo } from '../vpn/types.ts';
import { Login } from './login.tsx';

type NetbirdProps = {
  setIsLocked: (isLocked: boolean) => void;
};

export const Netbird = ({ setIsLocked }: NetbirdProps) => {
  const { t } = useTranslation();

  const vpn: VpnInfo = {
    id: 'netbird',
    title: 'NetBird',
    api,
    logoutLabel: t('settings.netbird.logout'),
    logoutWarning: t('settings.netbird.logoutDesc'),
    renderLogin: (onSuccess) => <Login onSuccess={onSuccess} />
  };

  return <VpnPage vpn={vpn} setIsLocked={setIsLocked} />;
};
