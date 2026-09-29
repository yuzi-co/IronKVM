import type { ReactNode } from 'react';
import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { HidMode } from '../../mouse/hid-mode.tsx';
import { ResetHid } from '../../mouse/reset-hid.tsx';
import { PowerLedSetting } from '../../power/power-led-setting.tsx';
import { MenuAction } from '../components/menu-action.tsx';
import { Hdmi } from './hdmi.tsx';
import { Oled } from './oled.tsx';
import { Reboot } from './reboot.tsx';
import { VirtualDevices } from './virtual-devices.tsx';

type SectionProps = {
  title: string;
  children: ReactNode;
};

const Section = ({ title, children }: SectionProps) => (
  <section className="flex flex-col space-y-6">
    <div className="text-sm text-neutral-400">{title}</div>
    {children}
  </section>
);

// The board's hardware, by what it faces: the video it takes in, the USB
// devices it presents to the host, and its own front panel.
export const Device = () => {
  const { t } = useTranslation();

  return (
    <>
      <div className="text-base">{t('settings.device.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <Section title={t('settings.device.sections.video')}>
          <Hdmi />
        </Section>
        <Divider className="opacity-50" style={{ margin: 0 }} />

        <Section title={t('settings.device.sections.usb')}>
          <VirtualDevices />
          <MenuAction description={t('settings.device.hidModeDesc')}>
            <HidMode />
          </MenuAction>
          <MenuAction description={t('settings.device.resetHidDesc')}>
            <ResetHid />
          </MenuAction>
        </Section>
        <Divider className="opacity-50" style={{ margin: 0 }} />

        <Section title={t('settings.device.sections.frontPanel')}>
          <Oled />
          <PowerLedSetting />
        </Section>
      </div>

      <Divider className="opacity-50" />

      <Reboot />
    </>
  );
};
