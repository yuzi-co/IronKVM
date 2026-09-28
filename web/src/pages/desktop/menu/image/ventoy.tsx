import { useState } from 'react';
import { Button, notification, Popconfirm, Switch } from 'antd';
import clsx from 'clsx';
import { ChevronDownIcon, ChevronRightIcon, TriangleAlertIcon, XIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/ventoy.ts';
import type { VentoyStatus } from '@/api/ventoy.ts';

type VentoyProps = {
  status: VentoyStatus | null;
  images: string[];
  onStatusChanged: (status: VentoyStatus | null) => void;
  onDrivesChanged: () => void;
};

type Busy = '' | 'install' | 'uninstall' | 'images' | 'insert' | 'eject';

function basename(path: string) {
  return path.replace(/^.*[\\/]/, '');
}

function formatSize(bytes: number) {
  const mib = bytes / (1024 * 1024);
  return mib >= 1024 ? `${(mib / 1024).toFixed(1)} GiB` : `${mib.toFixed(1)} MiB`;
}

// Ventoy builds one disk from the chosen images without copying them and puts it
// into the disk drive. The section shows the state of the release and the disk, a
// switch per image for the set on the disk, and the button that inserts or ejects it.
export const Ventoy = ({ status, images, onStatusChanged, onDrivesChanged }: VentoyProps) => {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();

  const [isExpanded, setIsExpanded] = useState(false);
  const [busy, setBusy] = useState<Busy>('');

  // The status has not loaded, or the account may not use Ventoy.
  if (!status) return null;

  const selected = status.images ?? [];
  const missing = status.missing ?? [];
  const present = selected.filter((image) => !missing.includes(image));

  // run sends one change and shows the state the server answers with. A refusal
  // or a failure shows the server's message as it is and reloads the state.
  function run(kind: Busy, request: () => ReturnType<typeof api.getVentoyStatus>) {
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
        if (kind === 'insert' || kind === 'eject') onDrivesChanged();
      });
  }

  function reload() {
    api.getVentoyStatus().then((rsp) => {
      if (rsp.code === 0) onStatusChanged(rsp.data);
    });
  }

  // toggle puts an image on the disk or takes it off. Selected paths that no
  // longer exist stay in the set until they are removed on their own.
  function toggle(image: string, on: boolean) {
    const next = on ? [...selected, image] : selected.filter((path) => path !== image);
    run('images', () => api.setVentoyImages(next));
  }

  function statusText() {
    if (!status) return '';
    if (!status.kernel) return t('image.ventoy.statusNoKernel');
    if (!status.installed) return t('image.ventoy.statusNotInstalled');
    if (status.inDrive) return t('image.ventoy.statusInDrive', { size: formatSize(status.size) });
    if (selected.length > 0) return t('image.ventoy.statusSelected', { count: selected.length });
    return t('image.ventoy.statusReady');
  }

  const locked = status.inDrive || !!busy;

  return (
    <>
      <div className="flex flex-col space-y-3">
        <div
          className="flex cursor-pointer items-center space-x-1 select-none"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          {isExpanded ? <ChevronDownIcon size={16} /> : <ChevronRightIcon size={16} />}
          <span>Ventoy</span>
          <span
            className={clsx(
              'flex-1 truncate pl-2 text-right text-xs',
              status.inDrive ? 'text-blue-500' : 'text-neutral-400'
            )}
          >
            {statusText()}
          </span>
        </div>

        {isExpanded && (
          <div className="flex flex-col space-y-3 pl-5 text-sm">
            {!status.kernel ? (
              <span className="text-xs text-amber-500">{t('image.ventoy.noKernel')}</span>
            ) : !status.installed ? (
              <div className="flex flex-col space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-neutral-400">{t('image.ventoy.installDesc')}</span>
                  <Button
                    type="primary"
                    size="small"
                    loading={busy === 'install'}
                    disabled={!status.onData || !!busy}
                    onClick={() => run('install', api.installVentoy)}
                  >
                    {t('image.ventoy.install')}
                  </Button>
                </div>
                {!status.onData && (
                  <span className="text-xs text-amber-500">{t('image.ventoy.needsData')}</span>
                )}
                {busy === 'install' && (
                  <span className="text-xs text-neutral-400">{t('image.ventoy.installing')}</span>
                )}
              </div>
            ) : (
              <>
                {images.length === 0 ? (
                  <span className="text-xs text-neutral-500">{t('image.ventoy.noImages')}</span>
                ) : (
                  <div className="flex max-h-[200px] flex-col space-y-1 overflow-y-auto">
                    {images.map((image) => (
                      <div key={image} className="flex items-center space-x-2">
                        <span className="flex-1 truncate">{basename(image)}</span>
                        <Switch
                          size="small"
                          title={t('image.ventoy.onDisk')}
                          checked={selected.includes(image)}
                          disabled={locked}
                          loading={busy === 'images'}
                          onChange={(on) => toggle(image, on)}
                        />
                      </div>
                    ))}
                  </div>
                )}

                {missing.map((image) => (
                  <div key={image} className="flex items-center space-x-2 text-amber-500">
                    <TriangleAlertIcon size={14} />
                    <span className="flex-1 truncate text-xs">
                      {t('image.ventoy.missing', { file: basename(image) })}
                    </span>
                    {!locked && (
                      <div
                        className="flex h-[20px] w-[20px] cursor-pointer items-center justify-center rounded text-neutral-400 hover:bg-neutral-500/50 hover:text-white"
                        title={t('image.ventoy.remove')}
                        onClick={() => toggle(image, false)}
                      >
                        <XIcon size={14} />
                      </div>
                    )}
                  </div>
                ))}

                {status.inDrive && (
                  <span className="text-xs text-neutral-500">{t('image.ventoy.setHint')}</span>
                )}

                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1">
                    <span className="font-mono text-xs text-neutral-500">{status.version}</span>
                    <Popconfirm
                      title={t('image.ventoy.uninstallConfirm')}
                      okText={t('image.okBtn')}
                      cancelText={t('image.cancelBtn')}
                      okButtonProps={{ danger: true }}
                      disabled={locked}
                      onConfirm={() => run('uninstall', api.uninstallVentoy)}
                    >
                      <Button
                        type="text"
                        size="small"
                        className="text-xs text-neutral-500"
                        loading={busy === 'uninstall'}
                        disabled={locked}
                      >
                        {t('image.ventoy.uninstall')}
                      </Button>
                    </Popconfirm>
                  </div>

                  {status.inDrive ? (
                    <Button
                      size="small"
                      loading={busy === 'eject'}
                      disabled={!!busy}
                      onClick={() => run('eject', api.ejectVentoy)}
                    >
                      {t('image.eject')}
                    </Button>
                  ) : (
                    <Button
                      type="primary"
                      size="small"
                      loading={busy === 'insert'}
                      disabled={present.length === 0 || !!busy}
                      onClick={() => run('insert', api.insertVentoy)}
                    >
                      {t('image.ventoy.useAsDisk')}
                    </Button>
                  )}
                </div>
              </>
            )}

            <span className="text-xs text-neutral-500">{t('image.ventoy.secureBoot')}</span>
            <span className="text-xs text-neutral-500">{t('image.ventoy.readOnly')}</span>
          </div>
        )}
      </div>

      {contextHolder}
    </>
  );
};
