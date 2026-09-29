import { useEffect, useState } from 'react';
import { Alert, Button, Divider, Popconfirm } from 'antd';
import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

import { UPDATE_PATHS } from '@/api/updates.ts';
import * as api from '@/api/ventoy.ts';
import type { VentoyStatus } from '@/api/ventoy.ts';
import { UpstreamUpdate } from '@/components/upstream-update.tsx';

import { useVentoyRequests } from '../../image/use-ventoy.ts';
import { ventoyStatusText } from '../../image/ventoy-status.ts';

type VirtualMediaProps = {
  setIsLocked: (isLocked: boolean) => void;
};

// Virtual media holds the setup behind the toolbar's Media dialog: installing,
// updating and removing Ventoy. Mounting images, adding them and choosing the
// Ventoy set stay in the dialog.
export const VirtualMedia = ({ setIsLocked }: VirtualMediaProps) => {
  const { t } = useTranslation();
  const [status, setStatus] = useState<VentoyStatus | null>(null);
  const { busy, run, reload, contextHolder } = useVentoyRequests(setStatus, (kind) => {
    if (kind === 'install') setIsLocked(false);
  });

  useEffect(() => {
    api
      .getVentoyStatus()
      .then((rsp) => {
        if (rsp.code === 0) setStatus(rsp.data);
      })
      // The page stays empty; opening it again asks again.
      .catch(() => {});
  }, []);

  function install() {
    setIsLocked(true);
    run('install', api.installVentoy);
  }

  const summary = status && ventoyStatusText(status);

  return (
    <>
      {contextHolder}
      <span className="text-base">{t('settings.media.title')}</span>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <span className="text-xs text-neutral-500">{t('settings.media.description')}</span>

        <div className="flex flex-col space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex flex-col space-y-1 pr-4">
              <span className="text-sm font-medium">Ventoy</span>
              <span className="text-xs text-neutral-500">{t('image.ventoy.installDesc')}</span>
              {summary && (
                <span
                  className={clsx(
                    'text-xs',
                    status?.inDrive ? 'text-blue-500' : 'text-neutral-400'
                  )}
                >
                  {t(summary.key, summary.values)}
                </span>
              )}
            </div>

            {status?.kernel && status.installed && (
              <Popconfirm
                title={t('image.ventoy.uninstallConfirm')}
                okText={t('image.okBtn')}
                cancelText={t('image.cancelBtn')}
                okButtonProps={{ danger: true }}
                disabled={status.inDrive || !!busy}
                onConfirm={() => run('uninstall', api.uninstallVentoy)}
              >
                <Button danger loading={busy === 'uninstall'} disabled={status.inDrive || !!busy}>
                  {t('image.ventoy.uninstall')}
                </Button>
              </Popconfirm>
            )}
            {status?.kernel && !status.installed && (
              <Button
                type="primary"
                loading={busy === 'install'}
                disabled={!status.onData || !!busy}
                onClick={install}
              >
                {t('image.ventoy.install')}
              </Button>
            )}
          </div>

          {status && !status.kernel && (
            <Alert type="warning" showIcon message={t('image.ventoy.noKernel')} />
          )}
          {status?.kernel && !status.installed && !status.onData && (
            <Alert type="info" showIcon message={t('image.ventoy.needsData')} />
          )}
          {busy === 'install' && (
            <Alert type="info" showIcon message={t('image.ventoy.installing')} />
          )}
          {status?.kernel && status.installed && (
            <UpstreamUpdate path={UPDATE_PATHS.ventoy} name="Ventoy" onUpdated={reload} />
          )}
          {status?.installed && status.inDrive && (
            <span className="text-xs text-neutral-500">{t('settings.media.ejectFirst')}</span>
          )}

          <span className="text-xs text-neutral-500">{t('image.ventoy.secureBoot')}</span>
          <span className="text-xs text-neutral-500">{t('image.ventoy.readOnly')}</span>
        </div>
      </div>
    </>
  );
};
