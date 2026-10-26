import { useAuth } from '@/contexts/auth.ts';
import { Button, Divider } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router';

import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { Logout } from './logout.tsx';
import { Oidc } from './oidc.tsx';
import { Sessions } from './sessions.tsx';
import { ShareLinks } from './share-links.tsx';
import { TwoFactor } from './two-factor.tsx';
import { Users } from './users.tsx';

export const Account = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { account } = useAuth();
  const hasShareLinks = useServerFeature('share-links') === true;
  const hasOidc = useServerFeature('oidc-login') === true;
  // A share session has no account, so no sessions of its own to list.
  const hasSessions =
    useServerFeature('login-sessions') === true && !account.username.startsWith('share:');

  function changePassword() {
    navigate('/auth/password');
  }

  return (
    <>
      <div className="text-base">{t('settings.account.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <div className="flex items-center justify-between">
          <span>{t('settings.account.webAccount')}</span>
          <span>{account.username}</span>
        </div>

        <div className="flex items-center justify-between">
          <span>{t('settings.account.role')}</span>
          <span>{t(`settings.account.roles.${account.role}`)}</span>
        </div>

        <div className="flex items-center justify-between">
          <span>{t('settings.account.password')}</span>
          <Button type="primary" onClick={changePassword}>
            {t('settings.account.updateBtn')}
          </Button>
        </div>
      </div>

      <Divider className="opacity-50" />

      <TwoFactor />

      <Divider className="opacity-50" />

      {hasSessions && (
        <>
          <Sessions scope="own" />
          <Divider className="opacity-50" />
        </>
      )}

      {account.role === 'admin' && (
        <>
          <Users />
          <Divider className="opacity-50" />
        </>
      )}

      {account.role === 'admin' && hasSessions && (
        <>
          <Sessions scope="all" />
          <Divider className="opacity-50" />
        </>
      )}

      {account.role === 'admin' && hasOidc && (
        <>
          <Oidc />
          <Divider className="opacity-50" />
        </>
      )}

      {account.role === 'admin' && hasShareLinks && (
        <>
          <ShareLinks />
          <Divider className="opacity-50" />
        </>
      )}

      <Logout />
    </>
  );
};
