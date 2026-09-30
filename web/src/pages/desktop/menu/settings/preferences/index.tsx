import { useAuth } from '@/contexts/auth.ts';
import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { LeaderKey } from '../../keyboard/leader-key.tsx';
import { MenuAction } from '../components/menu-action.tsx';
import { SectionHeader } from '../components/section.tsx';
import { KeyboardLedStatusSetting } from './keyboard-led-status.tsx';
import { Language } from './language.tsx';
import { MenuIcons } from './menu-icons.tsx';
import { MenuMode } from './menu-mode.tsx';
import { WebTitle } from './web-title.tsx';

// Most of this page is stored in the browser, and only the web title and the
// leader key on the device. The two sections say which is which, so nobody
// expects a language change to follow them to another machine, or a title
// change to stay local.
export const Preferences = () => {
  const { t } = useTranslation();
  const { account } = useAuth();

  return (
    <>
      <div className="text-base">{t('settings.preferences.title')}</div>
      <Divider className="opacity-50" />

      <SectionHeader
        title={t('settings.appearance.thisBrowser')}
        description={t('settings.appearance.thisBrowserDesc')}
      />
      <Language />

      <SectionHeader title={t('settings.appearance.menuBar.title')} className="mt-8" />
      <MenuMode />
      <KeyboardLedStatusSetting />
      <MenuIcons />

      {account.role === 'admin' && (
        <>
          <Divider className="opacity-50" style={{ margin: '32px 0' }} />

          <SectionHeader
            title={t('settings.appearance.deviceWide')}
            description={t('settings.appearance.deviceWideDesc')}
          />
          <WebTitle />
          <div className="mt-8">
            <MenuAction description={t('keyboard.leaderKey.desc')}>
              <LeaderKey />
            </MenuAction>
          </div>
        </>
      )}
    </>
  );
};
