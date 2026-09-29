import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { Hostname } from '../about/hostname.tsx';
import { Mdns } from '../device/mdns.tsx';
import { DNS } from './dns.tsx';
import { Ethernet } from './ethernet.tsx';
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

      <DNS />
    </>
  );
};
