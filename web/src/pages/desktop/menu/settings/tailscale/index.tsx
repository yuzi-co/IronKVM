import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/tailscale.ts';

import { VpnPage } from '../vpn/page.tsx';
import type { VpnInfo } from '../vpn/types.ts';
import { InstallHelp } from './install-help.tsx';
import { Login } from './login.tsx';

type TailscaleProps = {
  setIsLocked: (isLocked: boolean) => void;
};

export const Tailscale = ({ setIsLocked }: TailscaleProps) => {
  const { t } = useTranslation();

  const vpn: VpnInfo = {
    id: 'tailscale',
    title: 'Tailscale',
    api,
    logoutLabel: t('settings.tailscale.logout'),
    logoutWarning: t('settings.tailscale.logoutDesc'),
    renderLogin: (onSuccess) => <Login onSuccess={onSuccess} />,
    installHelp: <InstallHelp />
  };

  return <VpnPage vpn={vpn} setIsLocked={setIsLocked} />;
};
