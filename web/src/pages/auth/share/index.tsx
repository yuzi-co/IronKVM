import { ReactElement, useEffect, useState } from 'react';
import { Button, Card, Spin } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router';

import * as api from '@/api/auth.ts';
import { isShareToken } from '@/lib/share-link.ts';
import { Head } from '@/components/head.tsx';

// Share opens a view-only session from a share link: the token in the address
// goes to the device, which answers with a session that can watch the screen
// and nothing else. The address is replaced afterwards, so the token does not
// stay in the history.
export const Share = (): ReactElement => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { token } = useParams();
  const [failed, setFailed] = useState(!isShareToken(token));

  useEffect(() => {
    if (!isShareToken(token)) return;

    let active = true;
    api
      .redeemShareLink(token)
      .then((rsp) => {
        if (!active) return;
        if (rsp.code !== 0) {
          setFailed(true);
          return;
        }
        navigate('/', { replace: true });
      })
      .catch(() => {
        if (active) setFailed(true);
      });

    return () => {
      active = false;
    };
  }, [navigate, token]);

  return (
    <div className="flex h-screen w-screen items-center justify-center bg-neutral-950">
      <Head title={t('head.desktop')} />
      <Card className="w-80 text-center">
        {failed ? (
          <div className="flex flex-col items-center space-y-4">
            <span>{t('auth.share.failed')}</span>
            <Button type="primary" onClick={() => navigate('/auth/login', { replace: true })}>
              {t('auth.share.toLogin')}
            </Button>
          </div>
        ) : (
          <div className="flex flex-col items-center space-y-4">
            <Spin />
            <span className="text-neutral-400">{t('auth.share.opening')}</span>
          </div>
        )}
      </Card>
    </div>
  );
};
