import { useState } from 'react';
import { message, Switch, Tooltip } from 'antd';
import { useAtom } from 'jotai';
import { CameraIcon, EyeIcon, LoaderCircleIcon, PauseIcon } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { getScreenshot } from '@/api/stream.ts';
import { pauseWhenHiddenAtom, viewOnlyAtom } from '@/jotai/screen.ts';

const rowClass =
  'flex h-[30px] w-full items-center space-x-2 rounded p-0 pr-5 pl-3 text-left text-neutral-300';

// toPng re-encodes the board's JPEG as PNG, so a screenshot edited and saved
// again loses nothing more.
async function toPng(jpeg: Blob): Promise<Blob> {
  const bitmap = await createImageBitmap(jpeg);
  try {
    const canvas = document.createElement('canvas');
    canvas.width = bitmap.width;
    canvas.height = bitmap.height;
    const context = canvas.getContext('2d');
    if (!context) throw new Error('the browser gave no 2D canvas');
    context.drawImage(bitmap, 0, 0);
    return await new Promise<Blob>((resolve, reject) => {
      canvas.toBlob(
        (blob) => (blob ? resolve(blob) : reject(new Error('the frame could not be encoded'))),
        'image/png'
      );
    });
  } finally {
    bitmap.close();
  }
}

function screenshotName(now: Date) {
  const pad = (n: number) => String(n).padStart(2, '0');
  const date = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}`;
  const time = `${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
  return `screenshot-${date}-${time}.png`;
}

function save(blob: Blob, name: string) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  document.body.appendChild(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

type ToggleRowProps = {
  icon: LucideIcon;
  label: string;
  tip: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
};

const ToggleRow = ({ icon: Icon, label, tip, checked, onChange }: ToggleRowProps) => (
  <Tooltip placement="rightTop" title={tip} color="#262626" arrow>
    <label className={`${rowClass} cursor-pointer hover:bg-neutral-700/70`}>
      <Icon size={18} className={checked ? 'text-amber-400' : undefined} />
      <span className="flex-1 text-sm select-none">{label}</span>
      <Switch size="small" className="ml-6!" checked={checked} onChange={onChange} />
    </label>
  </Tooltip>
);

// Session holds the switches that belong to this tab rather than to the
// board, as PiKVM's System menu does: view only, pausing the stream while the
// tab is hidden, and a screenshot of the host screen.
export const Session = () => {
  const { t } = useTranslation();
  const [viewOnly, setViewOnly] = useAtom(viewOnlyAtom);
  const [pauseWhenHidden, setPauseWhenHidden] = useAtom(pauseWhenHiddenAtom);
  const [isCapturing, setIsCapturing] = useState(false);

  // The frame comes from the board, as OCR takes it, rather than from the
  // video element: the H.264 direct canvas belongs to a worker and cannot be
  // read back, and the board's frame is full size whatever the view scale.
  async function screenshot() {
    if (isCapturing) return;
    setIsCapturing(true);
    try {
      const png = await toPng(await getScreenshot());
      save(png, screenshotName(new Date()));
    } catch (error) {
      const detail = error instanceof Error ? error.message : '';
      message.error(
        detail ? `${t('screen.screenshotFailed')}: ${detail}` : t('screen.screenshotFailed')
      );
    } finally {
      setIsCapturing(false);
    }
  }

  return (
    <>
      <ToggleRow
        icon={EyeIcon}
        label={t('screen.viewOnly')}
        tip={t('screen.viewOnlyTip')}
        checked={viewOnly}
        onChange={setViewOnly}
      />
      <ToggleRow
        icon={PauseIcon}
        label={t('screen.pauseHidden')}
        tip={t('screen.pauseHiddenTip')}
        checked={pauseWhenHidden}
        onChange={setPauseWhenHidden}
      />
      <Tooltip placement="rightTop" title={t('screen.screenshotTip')} color="#262626" arrow>
        <button
          type="button"
          className={`${rowClass} cursor-pointer hover:bg-neutral-700/70`}
          onClick={screenshot}
        >
          {isCapturing ? (
            <LoaderCircleIcon className="animate-spin" size={18} />
          ) : (
            <CameraIcon size={18} />
          )}
          <span className="text-sm select-none">{t('screen.screenshot')}</span>
        </button>
      </Tooltip>
    </>
  );
};
