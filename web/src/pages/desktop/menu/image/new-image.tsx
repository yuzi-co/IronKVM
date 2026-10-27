import { useEffect, useState } from 'react';
import { Button, Input, InputNumber, Modal, Radio, Select } from 'antd';
import clsx from 'clsx';
import { FilePlus2Icon, LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { formatBytes } from '@/lib/health.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';

import {
  checkFloppySize,
  checkName,
  checkSize,
  defaultSize,
  FLOPPIES,
  FLOPPY_DEFAULT,
  imageFileName,
  PRESETS,
  toBytes,
  type ImageSpace,
  type SizeUnit
} from './new-image.ts';

const imageUpdatedEvent = 'nanokvm:image-updated';

// NewImageButton opens the dialog that creates an exchange image.
export const NewImageButton = ({ onClick }: { onClick: () => void }) => {
  const { t } = useTranslation();
  return (
    <Button className="self-start" icon={<FilePlus2Icon size={16} />} onClick={onClick}>
      {t('image.create.button')}
    </Button>
  );
};

type NewImageModalProps = {
  open: boolean;
  onClose: () => void;
  // onCreated gets the new image's path, as the image list names it.
  onCreated: (image: string) => void;
};

// NewImageModal creates an empty disk image for exchanging files with the
// host: a name, a size the device has room for, and a filesystem. Or a blank
// FAT12 floppy, at one of the floppy sizes. The image joins the library,
// ready to fill from the file browser or to insert into the disk drive.
export const NewImageModal = ({ open, onClose, onCreated }: NewImageModalProps) => {
  const { t } = useTranslation();
  useKeyboardLock('image-create', open);

  const [space, setSpace] = useState<ImageSpace | null>(null);
  const [spaceError, setSpaceError] = useState('');
  const [name, setName] = useState('exchange');
  const [kind, setKind] = useState<'disk' | 'floppy'>('disk');
  const [fs, setFs] = useState<api.ImageFs>('exfat');
  const [floppySize, setFloppySize] = useState(FLOPPY_DEFAULT);
  const [preset, setPreset] = useState<number | 'custom'>(0);
  const [custom, setCustom] = useState<number | null>(2);
  const [unit, setUnit] = useState<SizeUnit>('GiB');
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState('');

  // The free space is read each time the dialog opens: an upload or a
  // download may have taken some since. The caller keys the dialog by each
  // opening, so it starts from fresh state.
  useEffect(() => {
    if (!open) return;
    api
      .getImageSpace()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setSpaceError(rsp.msg);
          return;
        }
        const s = rsp.data as ImageSpace;
        setSpace(s);
        setPreset(defaultSize(s) || 'custom');
      })
      .catch((err) => setSpaceError(err?.message ?? ''));
  }, [open]);

  const floppy = kind === 'floppy';
  const diskSize = preset === 'custom' ? toBytes(custom ?? 0, unit) : preset;
  const size = floppy ? floppySize : diskSize;
  const nameProblem = checkName(name);
  const check = floppy ? checkFloppySize : checkSize;
  const sizeProblem = space ? check(size, space) : '';
  const canCreate = !!space && !nameProblem && !sizeProblem && !creating;

  // The name follows the kind while it is still the suggested one.
  function changeKind(next: 'disk' | 'floppy') {
    setKind(next);
    if (name === 'exchange' && next === 'floppy') setName('floppy');
    if (name === 'floppy' && next === 'disk') setName('exchange');
  }

  function create() {
    if (!canCreate) return;
    setCreating(true);
    setError('');
    api
      .createImage(name.trim(), size, floppy ? 'fat12' : fs)
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(rsp.msg || t('image.create.failed'));
          return;
        }
        window.dispatchEvent(new Event(imageUpdatedEvent));
        onCreated(rsp.data.file);
      })
      .catch((err) => setError(err?.message || t('image.create.failed')))
      .finally(() => setCreating(false));
  }

  const label = 'mb-1 text-sm text-neutral-400';

  return (
    <Modal
      title={t('image.create.title')}
      open={open}
      width={480}
      centered
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          {t('image.cancelBtn')}
        </Button>,
        <Button key="create" type="primary" disabled={!canCreate} onClick={create}>
          {creating && <LoaderCircleIcon className="animate-spin" size={14} />}
          {t('image.create.submit')}
        </Button>
      ]}
    >
      <div className="flex flex-col space-y-4 pt-2">
        <Radio.Group
          value={kind}
          optionType="button"
          buttonStyle="solid"
          onChange={(e) => changeKind(e.target.value)}
        >
          <Radio.Button value="disk">{t('image.disk')}</Radio.Button>
          <Radio.Button value="floppy">{t('image.floppy')}</Radio.Button>
        </Radio.Group>

        <span className="text-sm text-neutral-400">
          {floppy ? t('image.create.floppyDesc') : t('image.create.desc')}
        </span>

        <div>
          <div className={label}>{t('image.create.name')}</div>
          <Input
            value={name}
            maxLength={255}
            status={name !== '' && nameProblem ? 'error' : undefined}
            onChange={(e) => setName(e.target.value)}
            onPressEnter={create}
          />
          <div className={clsx('mt-1 text-xs', nameProblem ? 'text-red-500' : 'text-neutral-500')}>
            {nameProblem ? t(`image.create.${nameProblem}`) : imageFileName(name)}
          </div>
        </div>

        <div>
          <div className={label}>{t('image.create.size')}</div>
          {!space && !spaceError && (
            <LoaderCircleIcon className="animate-spin text-neutral-400" size={16} />
          )}
          {spaceError && <div className="text-sm text-red-500">{spaceError}</div>}
          {space && floppy && (
            <>
              <Radio.Group
                value={floppySize}
                optionType="button"
                buttonStyle="solid"
                size="small"
                onChange={(e) => setFloppySize(e.target.value)}
              >
                {FLOPPIES.map((f) => (
                  <Radio.Button key={f.size} value={f.size} disabled={f.size > space.max}>
                    {f.name}
                  </Radio.Button>
                ))}
              </Radio.Group>
              {sizeProblem && (
                <div className="mt-1 text-xs text-red-500">
                  {t(`image.create.${sizeProblem}`, { max: formatBytes(space.max) })}
                </div>
              )}
            </>
          )}
          {space && !floppy && (
            <>
              <Radio.Group
                value={preset}
                optionType="button"
                buttonStyle="solid"
                size="small"
                onChange={(e) => setPreset(e.target.value)}
              >
                {PRESETS.map((p) => (
                  <Radio.Button key={p} value={p} disabled={p > space.max}>
                    {formatBytes(p).replace('.0 ', ' ')}
                  </Radio.Button>
                ))}
                <Radio.Button value="custom">{t('image.create.custom')}</Radio.Button>
              </Radio.Group>

              {preset === 'custom' && (
                <div className="mt-2 flex items-center gap-2">
                  <InputNumber
                    min={0}
                    step={1}
                    value={custom}
                    onChange={(v) => setCustom(v)}
                    className="w-32"
                  />
                  <Select
                    value={unit}
                    onChange={setUnit}
                    options={[
                      { value: 'MiB', label: 'MiB' },
                      { value: 'GiB', label: 'GiB' }
                    ]}
                    className="w-20"
                  />
                </div>
              )}

              <div
                className={clsx('mt-1 text-xs', sizeProblem ? 'text-red-500' : 'text-neutral-500')}
              >
                {sizeProblem
                  ? t(`image.create.${sizeProblem}`, {
                      min: formatBytes(space.min),
                      max: formatBytes(space.max)
                    })
                  : t('image.create.room', { max: formatBytes(space.max) })}
              </div>
            </>
          )}
        </div>

        {!floppy && (
          <div>
            <div className={label}>{t('image.create.fs')}</div>
            <Radio.Group value={fs} onChange={(e) => setFs(e.target.value)}>
              <div className="flex flex-col space-y-2">
                <Radio value="exfat">
                  <span>exFAT</span>
                  <div className="text-xs text-neutral-500">{t('image.create.exfatHint')}</div>
                </Radio>
                <Radio value="fat32">
                  <span>FAT32</span>
                  <div className="text-xs text-neutral-500">{t('image.create.fat32Hint')}</div>
                </Radio>
              </div>
            </Radio.Group>
          </div>
        )}

        {error && <div className="text-sm text-red-500">{error}</div>}
      </div>
    </Modal>
  );
};
