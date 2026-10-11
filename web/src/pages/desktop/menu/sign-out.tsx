import { Popconfirm, Tooltip } from 'antd';
import { LogOutIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router';

import * as api from '@/api/auth.ts';
import { notifyAuthExpired } from '@/lib/auth-events.ts';
import { showFailure } from '@/lib/feedback.ts';

// SignOut is the one account control a view-only session has: it has no
// settings page to find the logout button in.
export const SignOut = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  function logout() {
    api
      .logout()
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }

        notifyAuthExpired();
        navigate('/auth/login');
      })
      .catch((err) => showFailure(err));
  }

  return (
    <Popconfirm
      placement="bottom"
      title={t('settings.account.logoutDesc')}
      okText={t('settings.account.okBtn')}
      cancelText={t('settings.account.cancelBtn')}
      onConfirm={logout}
    >
      <Tooltip title={t('settings.account.logoutBtn')} placement="bottom" mouseEnterDelay={0.6}>
        <button
          type="button"
          aria-label={t('settings.account.logoutBtn')}
          className="flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 text-neutral-300 hover:bg-neutral-700/80 hover:text-white"
        >
          <LogOutIcon size={18} />
        </button>
      </Tooltip>
    </Popconfirm>
  );
};
