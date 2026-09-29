import { useEffect } from 'react';
import { Button } from 'antd';
import { DiscIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { UPDATE_PATHS } from '@/api/updates.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';
import { UpstreamUpdate } from '@/components/upstream-update.tsx';

import { TransferStatus } from '../../media/transfer.tsx';
import { imageUpdatedEvent, useImageTransfer } from '../../media/use-image-transfer.ts';

// BootMenu downloads netboot.xyz's boot menu ISO into the image library, for
// the virtual CD, and keeps it up to date. It works without the network boot
// add-on: the host boots the ISO from the CD drive.
export const BootMenu = () => {
  const { t } = useTranslation();
  const transfer = useImageTransfer();

  // Pick up a download already running, from this page or another tab.
  const watch = useStableCallback(() => transfer.handleOpenChange(true));
  useEffect(() => {
    watch();
  }, [watch]);

  return (
    <div className="flex flex-col space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1 pr-4">
          <span className="text-sm font-medium">{t('download.bootMenu')}</span>
          <span className="text-xs text-neutral-500">{t('download.bootMenuDesc')}</span>
        </div>
        <Button
          icon={<DiscIcon size={16} />}
          onClick={transfer.downloadBootMenuImage}
          disabled={
            !transfer.diskEnabled || transfer.isCancelling || transfer.status === 'in_progress'
          }
        >
          {t('settings.netboot.isoDownload')}
        </Button>
      </div>
      {!transfer.diskEnabled && (
        <span className="text-xs text-amber-500">{t('download.disabled')}</span>
      )}
      <UpstreamUpdate
        path={UPDATE_PATHS.bootMenu}
        name="netboot.xyz.iso"
        onUpdated={() => window.dispatchEvent(new Event(imageUpdatedEvent))}
      />
      <TransferStatus transfer={transfer} origin="boot" />
    </div>
  );
};
