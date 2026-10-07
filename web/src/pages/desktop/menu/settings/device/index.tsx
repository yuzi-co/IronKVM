import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { PowerLedSetting } from '../../power/power-led-setting.tsx';
import { Section } from '../components/section.tsx';
import { Hdmi } from './hdmi.tsx';
import { Oled } from './oled.tsx';
import { Reboot } from './reboot.tsx';
import { RoomMic } from './room-mic.tsx';
import { VirtualDevices } from './virtual-devices.tsx';

// The board's hardware, by what it faces: the video it takes in (PCIe boards
// only, so Hdmi draws its own section), the USB devices it presents to the
// host, and its own front panel.
export const Device = () => {
  const { t } = useTranslation();

  return (
    <>
      <div className="text-base">{t('settings.device.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <Hdmi />

        <Section loose title={t('settings.device.sections.usb')}>
          <VirtualDevices />
        </Section>
        <Divider className="opacity-50" style={{ margin: 0 }} />

        <Section loose title={t('settings.device.sections.frontPanel')}>
          <Oled />
          <PowerLedSetting framed={false} />
        </Section>
        <Divider className="opacity-50" style={{ margin: 0 }} />

        <RoomMic />
      </div>

      <Divider className="opacity-50" />

      <Reboot />
    </>
  );
};
