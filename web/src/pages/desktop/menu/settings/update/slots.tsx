import { useEffect, useRef, useState } from 'react';
import { Button, Checkbox, Input, Popconfirm, Progress, Radio, Select, Tag } from 'antd';
import { CheckIcon, DownloadIcon, RotateCwIcon, Undo2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/slot.ts';
import type { InstallProgress, SlotImage, SlotInfo, SlotName, SlotStatus } from '@/api/slot.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import { formatBytes } from '@/lib/health.ts';
import { rebootAndReload } from '@/lib/reboot.ts';
import {
  armedTrial,
  installTargets,
  isHttpsUrl,
  onTrial,
  progressPercent,
  revertTarget,
  tryTargets
} from '@/lib/slot-install.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

// The tool rewrites its progress file once a second, so asking more often
// shows nothing new.
const POLL_MS = 1000;

type Source = 'url' | 'file';
type Busy = '' | 'install' | 'try' | 'confirm' | 'revert' | 'reboot';

// Slots is the A/B system update of the distribution image: an image is
// written to the slot that is not running, tried once on the next boot, and
// kept only once confirmed. Shown only by a server whose board has the slot
// tool.
export const Slots = ({ disabled }: { disabled: boolean }) => {
  const supported = useServerFeature('slot-install');
  if (!supported) return null;
  return <SlotsSection disabled={disabled} />;
};

const SlotsSection = ({ disabled }: { disabled: boolean }) => {
  const { t } = useTranslation();
  const [status, setStatus] = useState<SlotStatus | null>(null);
  const [loadError, setLoadError] = useState('');
  const [images, setImages] = useState<SlotImage[]>([]);
  const [progress, setProgress] = useState<InstallProgress | null>(null);
  const [busy, setBusy] = useState<Busy>('');

  const [source, setSource] = useState<Source>('url');
  const [url, setUrl] = useState('');
  const [file, setFile] = useState<string | undefined>();
  const [chosen, setChosen] = useState<SlotName | undefined>();
  const [force, setForce] = useState(false);

  const label = (name: string) =>
    name === 'recovery'
      ? t('settings.update.slots.recovery')
      : t('settings.update.slots.slotName', { name: name.toUpperCase() });

  const loadStatus = useStableCallback(() => {
    api
      .getSlotStatus()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setLoadError(rsp.msg || t('settings.update.slots.loadFailed'));
          return;
        }
        setLoadError('');
        setStatus(rsp.data as SlotStatus);
      })
      .catch(() => setLoadError(t('settings.update.slots.loadFailed')));
  });

  const loadImages = useStableCallback(() => {
    api
      .getSlotImages()
      .then((rsp) => {
        if (rsp.code === 0) setImages((rsp.data as SlotImage[]) ?? []);
      })
      .catch(() => {});
  });

  // The status is read again once an install ends, because the slot's record
  // changed with it.
  const wasRunning = useRef(false);
  const loadProgress = useStableCallback(() => {
    api
      .getInstallProgress()
      .then((rsp) => {
        if (rsp.code !== 0) return;
        const p = rsp.data as InstallProgress;
        if (wasRunning.current && !p.running) loadStatus();
        wasRunning.current = p.running;
        setProgress(p);
      })
      // The board can be slow to answer while it writes; the next poll asks again.
      .catch(() => {});
  });

  useEffect(() => {
    loadStatus();
    loadImages();
    loadProgress();
  }, [loadStatus, loadImages, loadProgress]);

  const installing = progress?.running === true;
  useEffect(() => {
    if (!installing) return;
    const id = window.setInterval(loadProgress, POLL_MS);
    return () => window.clearInterval(id);
  }, [installing, loadProgress]);

  // The slot chosen, or the first one an image can go to: B or A before
  // recovery, which is written far less often.
  const targets = status ? installTargets(status) : [];
  const target =
    chosen && targets.includes(chosen)
      ? chosen
      : (targets.find((s) => s !== 'recovery') ?? targets[0]);

  const blocked = disabled || busy !== '' || installing;
  const sourceOk = source === 'url' ? isHttpsUrl(url.trim()) : !!file;

  function install() {
    if (!target || !sourceOk) return;
    setBusy('install');
    const req: api.InstallRequest = { slot: target, force };
    if (source === 'url') req.url = url.trim();
    else req.file = file;
    api
      .installSlot(req)
      .then((rsp) => {
        if (!showResult(rsp)) return;
        // An install can end before the first answer about it, so the end
        // is looked for from here and not from the first answer that runs.
        wasRunning.current = true;
        loadProgress();
      })
      .catch((err) => showFailure(err))
      .finally(() => setBusy(''));
  }

  // Each decision answers with the status after it.
  function decide(kind: 'try' | 'confirm' | 'revert', run: () => Promise<any>, success: string) {
    setBusy(kind);
    run()
      .then((rsp) => {
        if (!showResult(rsp, { success })) return;
        if (rsp.data) setStatus(rsp.data as SlotStatus);
        else loadStatus();
      })
      .catch((err) => showFailure(err))
      .finally(() => setBusy(''));
  }

  function reboot() {
    setBusy('reboot');
    rebootAndReload((msg) => {
      showFailure({ code: -1, msg: msg || t('settings.device.rebootFailed') });
      setBusy('');
    });
  }

  if (!status) {
    return (
      <div className="mt-8 flex flex-col space-y-1">
        <span>{t('settings.update.slots.title')}</span>
        {loadError && <span className="text-xs text-amber-400">{loadError}</span>}
      </div>
    );
  }

  const armed = armedTrial(status);
  const trial = onTrial(status);
  const revertTo = revertTarget(status);
  const canTry = tryTargets(status);
  const trusted = label(status.trusted);
  const running = label(status.running);
  const finishedSlot =
    progress?.phase === 'done' && !progress.running && progress.slot ? progress.slot : '';
  const offerTry = finishedSlot && (canTry as string[]).includes(finishedSlot) ? finishedSlot : '';

  return (
    <div className="mt-8 flex flex-col gap-3">
      <div className="flex min-w-0 flex-col space-y-1">
        <span>{t('settings.update.slots.title')}</span>
        <span className="text-xs text-neutral-500">{t('settings.update.slots.desc')}</span>
        <span className="text-xs text-neutral-400">
          {status.running === status.trusted
            ? t('settings.update.slots.runningTrusted', { slot: running })
            : t('settings.update.slots.runningOther', { running, trusted })}
        </span>
      </div>

      <div className="flex flex-col gap-2">
        {status.slots.map((s) => (
          <SlotRow key={s.name} info={s} status={status} armed={armed} label={label} />
        ))}
      </div>

      {trial && (
        <div className="flex flex-col gap-2 rounded border border-amber-500/40 p-3">
          <span className="text-xs text-amber-400">
            {t('settings.update.slots.onTrial', { running, trusted })}
          </span>
          <div className="flex flex-wrap gap-2">
            <Popconfirm
              title={t('settings.update.slots.confirmTitle', { slot: running })}
              description={
                <div className="max-w-[320px]">
                  {t('settings.update.slots.confirmDesc', { slot: running, previous: trusted })}
                </div>
              }
              okText={t('settings.update.confirm')}
              cancelText={t('settings.update.cancel')}
              disabled={blocked}
              onConfirm={() =>
                decide(
                  'confirm',
                  api.confirmSlot,
                  t('settings.update.slots.confirmed', { slot: running })
                )
              }
            >
              <Button
                type="primary"
                icon={<CheckIcon size={15} />}
                loading={busy === 'confirm'}
                disabled={blocked}
              >
                {t('settings.update.slots.confirm', { slot: running })}
              </Button>
            </Popconfirm>
          </div>
        </div>
      )}

      {armed && (
        <div className="flex items-center justify-between gap-3">
          <span className="min-w-0 text-xs text-neutral-400">
            {t('settings.update.slots.armed', { slot: label(armed), trusted })}
          </span>
          <Popconfirm
            title={t('settings.update.slots.rebootTitle', { slot: label(armed) })}
            okText={t('settings.update.confirm')}
            cancelText={t('settings.update.cancel')}
            disabled={blocked}
            onConfirm={reboot}
          >
            <Button
              icon={<RotateCwIcon size={15} />}
              loading={busy === 'reboot'}
              disabled={blocked}
            >
              {t('settings.update.slots.reboot')}
            </Button>
          </Popconfirm>
        </div>
      )}

      {revertTo && (
        <div className="flex items-center justify-between gap-3">
          <span className="min-w-0 text-xs text-neutral-500">
            {t('settings.update.slots.revertDesc', { slot: label(revertTo), trusted })}
          </span>
          <Popconfirm
            title={t('settings.update.slots.revertTitle', { slot: label(revertTo) })}
            okText={t('settings.update.confirm')}
            cancelText={t('settings.update.cancel')}
            disabled={blocked}
            onConfirm={() =>
              decide(
                'revert',
                api.revertSlot,
                t('settings.update.slots.reverted', { slot: label(revertTo) })
              )
            }
          >
            <Button icon={<Undo2Icon size={15} />} loading={busy === 'revert'} disabled={blocked}>
              {t('settings.update.slots.revert', { slot: label(revertTo) })}
            </Button>
          </Popconfirm>
        </div>
      )}

      <div className="flex flex-col gap-2">
        <span className="text-sm">{t('settings.update.slots.installTitle')}</span>
        <span className="text-xs text-neutral-500">{t('settings.update.slots.installDesc')}</span>
        <Radio.Group
          value={source}
          disabled={blocked}
          onChange={(e) => setSource(e.target.value as Source)}
          optionType="button"
          size="small"
          options={[
            { value: 'url', label: t('settings.update.slots.fromUrl') },
            { value: 'file', label: t('settings.update.slots.fromFile') }
          ]}
        />
        {source === 'url' ? (
          <Input
            value={url}
            disabled={blocked}
            placeholder="https://.../slot-b.img.zst"
            status={url.trim() && !isHttpsUrl(url.trim()) ? 'warning' : undefined}
            onChange={(e) => setUrl(e.target.value)}
          />
        ) : images.length === 0 ? (
          <span className="text-xs text-neutral-500">{t('settings.update.slots.noImages')}</span>
        ) : (
          <Select
            value={file}
            disabled={blocked}
            placeholder={t('settings.update.slots.chooseImage')}
            onChange={setFile}
            options={images.map((img) => ({
              value: img.path,
              label: [
                img.path,
                img.version,
                formatBytes(img.size),
                img.signed ? '' : t('settings.update.slots.unsigned')
              ]
                .filter(Boolean)
                .join(' · ')
            }))}
          />
        )}
        {source === 'url' && url.trim() && !isHttpsUrl(url.trim()) && (
          <span className="text-xs text-amber-400">{t('settings.update.slots.invalidUrl')}</span>
        )}

        <div className="flex flex-wrap items-center gap-3">
          <span className="text-xs text-neutral-500">{t('settings.update.slots.target')}</span>
          <Select
            className="min-w-[140px]"
            value={target}
            disabled={blocked || targets.length === 0}
            onChange={setChosen}
            options={targets.map((name) => ({ value: name, label: label(name) }))}
          />
          <Checkbox checked={force} disabled={blocked} onChange={(e) => setForce(e.target.checked)}>
            <span className="text-xs">{t('settings.update.slots.force')}</span>
          </Checkbox>
          <div className="flex-1" />
          <Popconfirm
            title={t('settings.update.slots.installConfirm', { slot: target ? label(target) : '' })}
            description={
              <div className="max-w-[320px]">
                {t('settings.update.slots.installConfirmDesc', {
                  slot: target ? label(target) : '',
                  running
                })}
              </div>
            }
            okText={t('settings.update.confirm')}
            cancelText={t('settings.update.cancel')}
            disabled={blocked || !target || !sourceOk}
            onConfirm={install}
          >
            <Button
              type="primary"
              icon={<DownloadIcon size={15} />}
              loading={busy === 'install'}
              disabled={blocked || !target || !sourceOk}
            >
              {t('settings.update.slots.install')}
            </Button>
          </Popconfirm>
        </div>
        {targets.length === 0 && !installing && (
          <span className="text-xs text-neutral-500">{t('settings.update.slots.noTarget')}</span>
        )}
      </div>

      {progress && progress.phase !== '' && <InstallState progress={progress} label={label} />}

      {offerTry && (
        <div className="flex items-center justify-between gap-3">
          <span className="min-w-0 text-xs text-neutral-500">
            {t('settings.update.slots.tryDesc', { slot: label(offerTry), trusted })}
          </span>
          <Popconfirm
            title={t('settings.update.slots.tryTitle', { slot: label(offerTry) })}
            okText={t('settings.update.confirm')}
            cancelText={t('settings.update.cancel')}
            disabled={blocked}
            onConfirm={() =>
              decide(
                'try',
                () => api.trySlot(offerTry as 'a' | 'b'),
                t('settings.update.slots.tried', { slot: label(offerTry) })
              )
            }
          >
            <Button type="primary" loading={busy === 'try'} disabled={blocked}>
              {t('settings.update.slots.try', { slot: label(offerTry) })}
            </Button>
          </Popconfirm>
        </div>
      )}
    </div>
  );
};

type RowProps = {
  info: SlotInfo;
  status: SlotStatus;
  armed: string;
  label: (name: string) => string;
};

const SlotRow = ({ info, status, armed, label }: RowProps) => {
  const { t } = useTranslation();
  const stateKnown = ['installed', 'installing', 'interrupted', 'bad', 'unknown'].includes(
    info.state
  );
  const trialKnown = ['armed', 'booted', 'passed', 'failed'].includes(info.trial ?? '');
  const warn = info.state === 'bad' || info.state === 'interrupted';

  return (
    <div className="flex flex-col gap-1 rounded border border-neutral-500/30 px-3 py-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-sm">{label(info.name)}</span>
        {info.name === status.running && (
          <Tag color="blue">{t('settings.update.slots.running')}</Tag>
        )}
        {info.name === status.trusted && (
          <Tag color="green">{t('settings.update.slots.trusted')}</Tag>
        )}
        {info.name === armed && <Tag color="gold">{t('settings.update.slots.nextBoot')}</Tag>}
      </div>
      <span className={`text-xs ${warn ? 'text-amber-400' : 'text-neutral-400'}`}>
        {[
          stateKnown ? t(`settings.update.slots.state.${info.state}`) : info.state,
          info.version && t('settings.update.slots.version', { version: info.version }),
          info.installed &&
            t('settings.update.slots.installedAt', {
              time: new Date(info.installed).toLocaleString()
            }),
          trialKnown && t(`settings.update.slots.trial.${info.trial}`)
        ]
          .filter(Boolean)
          .join(' · ')}
      </span>
    </div>
  );
};

const InstallState = ({
  progress,
  label
}: {
  progress: InstallProgress;
  label: (name: string) => string;
}) => {
  const { t } = useTranslation();
  const p = progress;
  const failed = p.phase === 'failed';
  const done = p.phase === 'done';
  const known = ['verify', 'write', 'readback', 'done', 'failed'].includes(p.phase);

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between gap-3 text-xs">
        <span className={failed ? 'text-red-400' : 'text-neutral-400'}>
          {known ? t(`settings.update.slots.phase.${p.phase}`) : p.phase}
        </span>
        {(p.phase === 'write' || p.phase === 'readback') && p.total > 0 && (
          <span className="text-neutral-500">
            {t('settings.update.slots.bytes', {
              done: formatBytes(p.done),
              total: formatBytes(p.total)
            })}
          </span>
        )}
      </div>
      <Progress
        percent={progressPercent(p)}
        size="small"
        status={failed ? 'exception' : done ? 'success' : 'active'}
        showInfo={!failed}
      />
      {p.slot && p.source && (
        <span className="truncate font-mono text-xs text-neutral-500">
          {t('settings.update.slots.source', { slot: label(p.slot), source: p.source })}
        </span>
      )}
      {failed && p.message && <span className="text-xs text-red-400">{p.message}</span>}
    </div>
  );
};
