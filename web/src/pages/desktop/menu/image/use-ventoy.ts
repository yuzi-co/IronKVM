import { useState } from 'react';
import { notification } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/ventoy.ts';
import type { VentoyStatus } from '@/api/ventoy.ts';

export type VentoyBusy = '' | 'install' | 'uninstall' | 'images' | 'insert' | 'eject';

type Request = () => ReturnType<typeof api.getVentoyStatus>;

// useVentoyRequests sends Ventoy changes one at a time, for the Media dialog
// and the Virtual media settings page. Each change shows the state the server
// answers with. A refusal or a failure shows the server's message as it is and
// reloads the state. onDone runs after every change, whatever its result.
export function useVentoyRequests(
  onStatusChanged: (status: VentoyStatus | null) => void,
  onDone?: (kind: VentoyBusy) => void
) {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();
  const [busy, setBusy] = useState<VentoyBusy>('');

  function reload() {
    api
      .getVentoyStatus()
      .then((rsp) => {
        if (rsp.code === 0) onStatusChanged(rsp.data);
      })
      // The failure was already reported; the next open reads the state again.
      .catch(() => {});
  }

  function run(kind: VentoyBusy, request: Request) {
    if (busy) return;
    setBusy(kind);

    request()
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({
            message: t('image.ventoy.failed'),
            description: rsp.msg,
            duration: 10
          });
          reload();
          return;
        }
        onStatusChanged(rsp.data);
      })
      .catch((err) => {
        notify.open({
          message: t('image.ventoy.failed'),
          description: err?.message,
          duration: 10
        });
        reload();
      })
      .finally(() => {
        setBusy('');
        onDone?.(kind);
      });
  }

  return { busy, run, reload, contextHolder };
}
