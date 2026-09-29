import { useState } from 'react';
import { Button, Switch } from 'antd';
import clsx from 'clsx';
import { ChevronDownIcon, ChevronRightIcon, TriangleAlertIcon, XIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/ventoy.ts';
import type { VentoyStatus } from '@/api/ventoy.ts';

import { useVentoyRequests } from './use-ventoy.ts';
import { basename, ventoyPresent, ventoyStatusText } from './ventoy-status.ts';

type VentoyProps = {
  status: VentoyStatus;
  images: string[];
  onStatusChanged: (status: VentoyStatus | null) => void;
  onDrivesChanged: () => void;
};

// Ventoy builds one disk from the chosen images without copying them and puts
// it into the disk drive. In the Media dialog it is one row: the state and the
// button that inserts or ejects the disk, which opens onto a switch per image
// for the set on the disk. Installing, updating and removing Ventoy are in
// Settings, on the Virtual media page.
export const Ventoy = ({ status, images, onStatusChanged, onDrivesChanged }: VentoyProps) => {
  const { t } = useTranslation();
  const [isExpanded, setIsExpanded] = useState(false);
  const { busy, run, contextHolder } = useVentoyRequests(onStatusChanged, (kind) => {
    if (kind === 'insert' || kind === 'eject') onDrivesChanged();
  });

  const selected = status.images ?? [];
  const missing = status.missing ?? [];
  const present = ventoyPresent(status);
  const summary = ventoyStatusText(status);
  const locked = status.inDrive || !!busy;

  // toggle puts an image on the disk or takes it off. Selected paths that no
  // longer exist stay in the set until they are removed on their own.
  function toggle(image: string, on: boolean) {
    const next = on ? [...selected, image] : selected.filter((path) => path !== image);
    run('images', () => api.setVentoyImages(next));
  }

  return (
    <>
      <div className="flex flex-col space-y-3">
        <div className="flex items-center space-x-2">
          <button
            type="button"
            aria-expanded={isExpanded}
            className="flex min-w-0 flex-1 cursor-pointer items-center space-x-1 p-0 text-left select-none"
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
              {t(summary.key, summary.values)}
            </span>
          </button>

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

        {isExpanded && (
          <div className="flex flex-col space-y-3 pl-5 text-sm">
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
                  <button
                    type="button"
                    className="flex h-[20px] w-[20px] cursor-pointer items-center justify-center rounded p-0 text-neutral-400 hover:bg-neutral-500/50 hover:text-white"
                    title={t('image.ventoy.remove')}
                    aria-label={t('image.ventoy.remove')}
                    onClick={() => toggle(image, false)}
                  >
                    <XIcon size={14} />
                  </button>
                )}
              </div>
            ))}

            {status.inDrive && (
              <span className="text-xs text-neutral-500">{t('image.ventoy.setHint')}</span>
            )}
          </div>
        )}
      </div>

      {contextHolder}
    </>
  );
};
