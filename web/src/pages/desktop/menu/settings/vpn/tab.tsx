import { useState } from 'react';
import { Segmented } from 'antd';
import { ShieldCheckIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { Netbird as NetbirdIcon } from '@/components/icons/netbird';
import { Tailscale as TailscaleIcon } from '@/components/icons/tailscale';

import { browserStorage, readStored, VPN_PROVIDER_KEY, writeStored } from '../nav.ts';
import { Netbird } from '../netbird';
import { Tailscale } from '../tailscale';
import { Wireguard } from '../wireguard';

type Provider = 'tailscale' | 'netbird' | 'wireguard';

function storedProvider(): Provider {
  const stored = readStored(browserStorage(), VPN_PROVIDER_KEY);
  return stored === 'netbird' || stored === 'wireguard' ? stored : 'tailscale';
}

type VpnTabProps = {
  isLocked: boolean;
  setIsLocked: (isLocked: boolean) => void;
};

// VpnTab puts Tailscale, NetBird and the built-in WireGuard behind one sidebar
// entry. The two add-ons are the same VpnPage underneath; WireGuard has a page
// of its own, as it needs no install or login.
export const VpnTab = ({ isLocked, setIsLocked }: VpnTabProps) => {
  const { t } = useTranslation();

  const [provider, setProvider] = useState<Provider>(storedProvider);

  function change(value: Provider) {
    setProvider(value);
    writeStored(browserStorage(), VPN_PROVIDER_KEY, value);
  }

  return (
    <>
      {/* The margin sits on a wrapper: antd's reset gives Segmented margin 0,
          and its unlayered style beats a Tailwind class on the component. */}
      <div className="mb-6">
        <Segmented<Provider>
          block
          aria-label={t('settings.nav.vpnProvider')}
          disabled={isLocked}
          value={provider}
          onChange={change}
          options={[
            {
              value: 'tailscale',
              label: (
                <div className="flex items-center justify-center space-x-2">
                  <TailscaleIcon size={15} />
                  <span>Tailscale</span>
                </div>
              )
            },
            {
              value: 'netbird',
              label: (
                <div className="flex items-center justify-center space-x-2">
                  <NetbirdIcon size={15} />
                  <span>NetBird</span>
                </div>
              )
            },
            {
              value: 'wireguard',
              label: (
                <div className="flex items-center justify-center space-x-2">
                  <ShieldCheckIcon size={15} />
                  <span>WireGuard</span>
                </div>
              )
            }
          ]}
        />
      </div>

      {provider === 'tailscale' && <Tailscale setIsLocked={setIsLocked} />}
      {provider === 'netbird' && <Netbird setIsLocked={setIsLocked} />}
      {provider === 'wireguard' && <Wireguard />}
    </>
  );
};
