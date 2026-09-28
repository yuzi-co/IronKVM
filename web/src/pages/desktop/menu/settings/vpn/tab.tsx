import { useState } from 'react';
import { Segmented } from 'antd';
import { useTranslation } from 'react-i18next';

import { Netbird as NetbirdIcon } from '@/components/icons/netbird';
import { Tailscale as TailscaleIcon } from '@/components/icons/tailscale';

import { browserStorage, readStored, VPN_PROVIDER_KEY, writeStored } from '../nav.ts';
import { Netbird } from '../netbird';
import { Tailscale } from '../tailscale';

type Provider = 'tailscale' | 'netbird';

type VpnTabProps = {
  isLocked: boolean;
  setIsLocked: (isLocked: boolean) => void;
};

// VpnTab puts Tailscale and NetBird behind one sidebar entry. Both are the same
// VpnPage underneath, so the switch only picks which one is shown.
export const VpnTab = ({ isLocked, setIsLocked }: VpnTabProps) => {
  const { t } = useTranslation();

  const [provider, setProvider] = useState<Provider>(() =>
    readStored(browserStorage(), VPN_PROVIDER_KEY) === 'netbird' ? 'netbird' : 'tailscale'
  );

  function change(value: Provider) {
    setProvider(value);
    writeStored(browserStorage(), VPN_PROVIDER_KEY, value);
  }

  return (
    <>
      <Segmented<Provider>
        block
        className="mb-6"
        aria-label={t('settings.nav.vpnProvider')}
        disabled={isLocked}
        value={provider}
        onChange={change}
        options={[
          {
            value: 'tailscale',
            label: (
              <div className="flex items-center justify-center space-x-2">
                <TailscaleIcon />
                <span>Tailscale</span>
              </div>
            )
          },
          {
            value: 'netbird',
            label: (
              <div className="flex items-center justify-center space-x-2">
                <NetbirdIcon />
                <span>NetBird</span>
              </div>
            )
          }
        ]}
      />

      {provider === 'tailscale' ? (
        <Tailscale setIsLocked={setIsLocked} />
      ) : (
        <Netbird setIsLocked={setIsLocked} />
      )}
    </>
  );
};
