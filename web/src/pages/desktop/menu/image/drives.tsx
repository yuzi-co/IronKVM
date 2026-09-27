import { useState } from 'react';
import { Button, notification, Switch, Tooltip } from 'antd';
import clsx from 'clsx';
import { DiscIcon, HardDriveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';

type DrivesProps = {
  drives: api.Drive[];
  diskRo: boolean;
  setDiskRo: (ro: boolean) => void;
  onDrivesChanged: () => void;
};

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

          return (
            <div key={drive.id} className="flex items-center space-x-2">
              <Icon size={16} />
              <span className="w-[48px]">{t(`image.${drive.id}`)}</span>
              <span
                className={clsx(
                  'flex-1 truncate',
                  drive.file ? 'text-blue-500' : 'text-neutral-500'
                )}
              >
                {drive.file ? drive.file.replace(/^.*[\\/]/, '') : t('image.driveEmpty')}
              </span>

              {drive.id === 'disk' && (
                <Tooltip title={t('image.readOnlyTip')}>
                  <div className="flex items-center space-x-1">
                    <span className="text-xs text-neutral-400">{t('image.readOnly')}</span>
                    <Switch size="small" checked={diskRo} onChange={setDiskRo} />
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
          );
        })}
      </div>

      {contextHolder}
    </>
  );
};
