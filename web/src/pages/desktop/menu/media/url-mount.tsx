import { useState } from 'react';
import { Button, Input, Segmented } from 'antd';
import { LinkIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/url-media.ts';
import { formatBytes } from '@/lib/health.ts';

import { cacheHitRate, isHttpUrl, urlMountLabel, type UrlMediaStatus } from './url-media.ts';

type UrlMountProps = {
  status: UrlMediaStatus | null;
  onChanged: () => void;
};

// UrlMount serves an image straight from a web server, without storing it on
// the board (ironkvm-dist#106). The board reads it on demand over HTTP range
// requests, so the server must support them. The drive list above shows and
// ejects it like any other image; this shows how the reading goes.
export const UrlMount = ({ status, onChanged }: UrlMountProps) => {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const [url, setUrl] = useState('');
  const [drive, setDrive] = useState<'cdrom' | 'disk'>('cdrom');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const unavailable = status !== null && !status.available;
  const mounts = status?.mounts ?? [];

  function mount() {
    if (busy || !isHttpUrl(url)) return;
    setBusy(true);
    setError('');
    api
      .mountUrl(drive, url.trim())
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(`${t('urlMedia.mountFailed')}: ${rsp.msg}`);
          return;
        }
        // The URL is not kept in the page either once it is mounted.
        setUrl('');
        setIsOpen(false);
      })
      .catch((err) => setError(`${t('urlMedia.mountFailed')}: ${String(err?.message ?? err)}`))
      .finally(() => {
        setBusy(false);
        onChanged();
      });
  }

  return (
    <div className="flex flex-col space-y-2">
      <Button
        className="self-start"
        icon={<LinkIcon size={16} />}
        aria-expanded={isOpen}
        disabled={unavailable}
        onClick={() => setIsOpen(!isOpen)}
      >
        {t('urlMedia.open')}
      </Button>
      {unavailable && <div className="text-xs text-neutral-500">{t('urlMedia.unavailable')}</div>}

      {isOpen && !unavailable && (
        <div className="space-y-2">
          <div className="text-xs text-neutral-500">{t('urlMedia.desc')}</div>
          <Input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onPressEnter={mount}
            disabled={busy}
            placeholder="https://example.com/image.iso"
            autoComplete="off"
          />
          <div className="flex items-center gap-2">
            <Segmented
              size="small"
              value={drive}
              onChange={(value) => setDrive(value as 'cdrom' | 'disk')}
              disabled={busy}
              options={[
                { label: t('urlMedia.asCd'), value: 'cdrom' },
                { label: t('urlMedia.asDisk'), value: 'disk' }
              ]}
            />
            <Button
              type="primary"
              className="ml-auto"
              loading={busy}
              disabled={!isHttpUrl(url)}
              onClick={mount}
            >
              {t('urlMedia.mount')}
            </Button>
          </div>
          <div className="text-xs text-neutral-500">{t('urlMedia.credentials')}</div>
        </div>
      )}
      {error && <div className="text-sm wrap-break-word text-red-500">{error}</div>}

      {mounts.map((m) => {
        const rate = cacheHitRate(m.cache);
        return (
          <div key={m.drive} className="flex flex-col text-xs text-neutral-400">
            <span className="truncate text-blue-500">
              {t(`image.${m.drive}`)}: {urlMountLabel(m.device, mounts)}
            </span>
            <span>
              {t('urlMedia.read', { read: formatBytes(m.bytesRead), size: formatBytes(m.size) })}
              {rate !== null && ` · ${t('urlMedia.hitRate', { rate })}`}
            </span>
            {m.error && (
              <span className="text-amber-500">{t('urlMedia.lastError', { error: m.error })}</span>
            )}
          </div>
        );
      })}
    </div>
  );
};
