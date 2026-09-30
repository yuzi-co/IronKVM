import { useState } from 'react';
import { Button, notification, Switch, Tooltip } from 'antd';
import clsx from 'clsx';
import { DiscIcon, HardDriveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { VENTOY_DEVICE } from '@/api/ventoy.ts';
import { formatBytes } from '@/lib/health.ts';

import { MediaWarnings } from './media-warnings.tsx';
import { CD_MAX_BYTES, driveWarnings } from './warnings.ts';

type DrivesProps = {
  drives: api.Drive[];
  diskRo: boolean;
  setDiskRo: (ro: boolean) => void;
  onDrivesChanged: () => void;
};

// driveLabel names what a drive holds: the image's file name, or Ventoy for the
// disk built from the Ventoy set.
function driveLabel(file: string) {
  return file === VENTOY_DEVICE ? 'Ventoy' : file.replace(/^.*[\\/]/, '');
}

// Drives shows what each virtual drive holds, with an eject button per drive
// and the read-only switch that the next disk insert uses.
export const Drives = ({ drives, diskRo, setDiskRo, onDrivesChanged }: DrivesProps) => {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();
  const [ejecting, setEjecting] = useState('');

  function eject(id: api.DriveId) {
    if (ejecting) return;
    setEjecting(id);

    api
      .ejectDrive(id)
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({ message: t('image.ejectFailed'), description: rsp.msg, duration: 10 });
        }
      })
      .finally(() => {
        setEjecting('');
        onDrivesChanged();
      });
  }

  if (drives.length === 0) {
    return <div className="text-sm text-neutral-500">{t('image.noDrives')}</div>;
  }

  return (
    <>
      <div className="flex flex-col space-y-3">
        {drives.map((drive) => {
          const Icon = drive.id === 'cdrom' ? DiscIcon : HardDriveIcon;
          const warnings = driveWarnings(drive);

          return (
            <div key={drive.id} className="flex flex-col space-y-1">
              <div className="flex items-center space-x-2">
                <Icon size={18} />
                <span className="w-[48px]">{t(`image.${drive.id}`)}</span>
                <span
                  className={clsx(
                    'flex-1 truncate',
                    drive.file ? 'text-blue-500' : 'text-neutral-500'
                  )}
                >
                  {drive.file ? driveLabel(drive.file) : t('image.driveEmpty')}
                </span>

                {/* The flag is fixed while a disk is in: the gadget reads it at
                    insert time, so the switch waits for the next insert. */}
                {drive.id === 'disk' && (
                  <Tooltip title={t(drive.file ? 'image.readOnlyLocked' : 'image.readOnlyTip')}>
                    <div className="flex items-center space-x-1">
                      <span className="text-xs text-neutral-400">{t('image.readOnly')}</span>
                      <Switch
                        size="small"
                        checked={diskRo}
                        disabled={!!drive.file}
                        onChange={setDiskRo}
                      />
                    </div>
                  </Tooltip>
                )}

                <Button
                  size="small"
                  disabled={!drive.file}
                  loading={ejecting === drive.id}
                  onClick={() => eject(drive.id)}
                >
                  {t('image.eject')}
                </Button>
              </div>
              <MediaWarnings
                warnings={warnings}
                values={{ max: formatBytes(CD_MAX_BYTES), size: formatBytes(drive.size ?? 0) }}
              />
            </div>
          );
        })}
      </div>

      {contextHolder}
    </>
  );
};
