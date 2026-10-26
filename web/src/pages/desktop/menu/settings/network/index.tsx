import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { Hostname } from '../about/hostname.tsx';
import { Mdns } from '../device/mdns.tsx';
import { AllowedNetworks } from './allowlist.tsx';
import { DNS } from './dns.tsx';
import { Ethernet } from './ethernet.tsx';
import { IPv6 } from './ipv6.tsx';
import { StunServer } from './stun.tsx';
import { Syslog } from './syslog.tsx';
import { WebRTCForward } from './webrtc.tsx';
import { Wifi } from './wifi.tsx';

export const Network = () => {
  const { t } = useTranslation();

  return (
    <>
      <div className="text-base">{t('settings.network.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <Hostname editable={true} />
        <Mdns />
      </div>

      <Divider className="opacity-50" style={{ margin: '32px 0' }} />

      {/* Wi-Fi renders nothing on boards without it, so it shares a block
          with Ethernet rather than getting dividers of its own. */}
      <div className="flex flex-col space-y-8">
        <Wifi />
        <Ethernet />
      </div>

      <Divider className="opacity-50" style={{ margin: '32px 0' }} />

      <IPv6Block />

      <DNS />

      <Divider className="opacity-50" style={{ margin: '32px 0' }} />

      <Syslog />

      <AllowedNetworksBlock />

      <WebRTCForwardBlock />

      <StunServerBlock />
    </>
  );
};

// The allowlist renders nothing on a server without it, so its divider goes
// with it.
const AllowedNetworksBlock = () => {
  const supported = useServerFeature('ip-allowlist');
  if (!supported) return null;

  return (
    <>
      <Divider className="opacity-50" style={{ margin: '32px 0' }} />
      <AllowedNetworks />
    </>
  );
};

// IPv6 renders nothing on a server without it, so its divider goes with it.
const IPv6Block = () => {
  const supported = useServerFeature('ipv6-settings');
  if (!supported) return null;

  return (
    <>
      <IPv6 />
      <Divider className="opacity-50" style={{ margin: '32px 0' }} />
    </>
  );
};

// The port forward renders nothing on a server without it, so its divider goes
// with it.
const WebRTCForwardBlock = () => {
  const supported = useServerFeature('webrtc-port-forward');
  if (!supported) return null;

  return (
    <>
      <Divider className="opacity-50" style={{ margin: '32px 0' }} />
      <WebRTCForward />
    </>
  );
};

// The STUN server renders nothing on a server without the setting, so its
// divider goes with it.
const StunServerBlock = () => {
  const supported = useServerFeature('stun-server');
  if (!supported) return null;

  return (
    <>
      <Divider className="opacity-50" style={{ margin: '32px 0' }} />
      <StunServer />
    </>
  );
};
