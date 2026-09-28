import { useAuth } from '@/contexts/auth.ts';
import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { KeyboardLedStatusSetting } from './keyboard-led-status.tsx';
import { Language } from './language.tsx';
import { MenuIcons } from './menu-icons.tsx';
import { MenuMode } from './menu-mode.tsx';
import { WebTitle } from './web-title.tsx';

// Most of this page is stored in the browser, and only the web title on the
// device. The two sections say which is which, so nobody expects a language
// change to follow them to another machine, or a title change to stay local.
export const Appearance = () => {
  const { t } = useTranslation();
  const { account } = useAuth();

  return (
    <>
      <div className="text-base">{t('settings.appearance.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-1">
        <span className="text-neutral-400">{t('settings.appearance.thisBrowser')}</span>
        <span className="text-xs text-neutral-500">{t('settings.appearance.thisBrowserDesc')}</span>
      </div>
      <Language />

      <div className="mt-8 text-sm text-neutral-400">{t('settings.appearance.menuBar.title')}</div>
      <MenuMode />
      <KeyboardLedStatusSetting />
      <MenuIcons />

      {account.role === 'admin' && (
        <>
          <Divider className="opacity-50" style={{ margin: '32px 0' }} />

          <div className="flex flex-col space-y-1">
            <span className="text-neutral-400">{t('settings.appearance.deviceWide')}</span>
            <span className="text-xs text-neutral-500">
              {t('settings.appearance.deviceWideDesc')}
            </span>
          </div>
          <WebTitle />
        </>
      )}
    </>
  );
};
