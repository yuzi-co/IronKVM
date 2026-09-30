import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { Tooltip } from 'antd';
import clsx from 'clsx';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { formatBytes } from '@/lib/health.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import type { IonStatus } from '../../../ion-status/model';
import { Box, Section } from '../components/section.tsx';
import { usePoll } from '../components/use-poll.ts';
import { groupUse, ramUse, type MemoryStatus, type MemorySwap } from './memory.ts';

const pollMs = 5 * 1000;

type RowProps = {
  label: ReactNode;
  tip?: string;
  value: ReactNode;
  className?: string;
};

// Row is one reading: its name on the left, the numbers on the right.
const Row = ({ label, tip, value, className }: RowProps) => (
  <div className="flex min-h-[24px] items-center justify-between gap-3">
    <span className="flex items-center space-x-2 text-sm">
      <span>{label}</span>
      {tip && (
        <Tooltip title={tip} placement="right" styles={{ root: { maxWidth: '400px' } }}>
          <CircleHelpIcon className="cursor-pointer text-neutral-500" size={14} />
        </Tooltip>
      )}
    </span>
    <span className={clsx('text-right font-mono text-xs', className ?? 'text-neutral-300')}>
      {value}
    </span>
  </div>
);

// Memory is the board's memory in one place: RAM, swap, the video capture's
// carveout, and what uses the most of it. It is read again every few seconds
// while the page is open and the browser tab is visible.
export const Memory = () => {
  const { t } = useTranslation();

  const [memory, setMemory] = useState<MemoryStatus>();
  const [ion, setIon] = useState<IonStatus>();
  const inFlight = useRef(false);

  const load = useStableCallback(() => {
    if (inFlight.current) return;
    inFlight.current = true;

    Promise.all([
      api
        .getMemory()
        .then((rsp) => {
          if (rsp.code === 0) setMemory(rsp.data);
        })
        .catch(() => {}),
      api
        .getIon()
        .then((rsp) => {
          if (rsp.code === 0) setIon(rsp.data);
        })
        .catch(() => {})
    ]).finally(() => {
      inFlight.current = false;
    });
  });

  useEffect(() => {
    load();
  }, [load]);
  usePoll(load, pollMs);

  const of = (used: number, total: number) =>
    t('settings.performance.memory.of', { used: formatBytes(used), total: formatBytes(total) });

  function swapLabel(swap: MemorySwap) {
    if (swap.kind === 'zram') return t('settings.performance.memory.zram');
    if (swap.kind === 'file') return t('settings.performance.memory.swapFile');
    return swap.name;
  }

  const ram = memory ? ramUse(memory) : undefined;
  const swaps = memory?.swaps ?? [];
  const addons = memory?.addons ? groupUse(memory.addons) : undefined;
  // A board that cannot report its carveout shows no row for it.
  const showIon = !!ion && ion.verdict !== 'unavailable';

  return (
    <Section
      title={t('settings.performance.memory.title')}
      description={t('settings.performance.memory.description')}
    >
      {memory && ram && memory.total > 0 && (
        <Box>
          <div className="flex flex-col space-y-1.5">
            <Row
              label={t('settings.performance.memory.ram')}
              value={of(ram.used, memory.total)}
              className={ram.low ? 'text-amber-500' : undefined}
            />
            <div
              role="progressbar"
              aria-label={t('settings.performance.memory.ram')}
              aria-valuenow={ram.percent}
              aria-valuemin={0}
              aria-valuemax={100}
              className="h-1.5 w-full overflow-hidden rounded-full bg-neutral-700/60"
            >
              <div
                className={clsx('h-full rounded-full', ram.low ? 'bg-amber-500' : 'bg-blue-500')}
                style={{ width: `${ram.percent}%` }}
              />
            </div>
            <span className={clsx('text-xs', ram.low ? 'text-amber-500' : 'text-neutral-500')}>
              {t(
                ram.low
                  ? 'settings.performance.memory.availableLow'
                  : 'settings.performance.memory.available',
                { available: formatBytes(memory.available) }
              )}
            </span>
          </div>

          {swaps.length === 0 ? (
            <Row label={t('settings.performance.memory.swap')} value={t('common.off')} />
          ) : (
            swaps.map((swap) => (
              <Row
                key={swap.name}
                label={swapLabel(swap)}
                value={
                  swap.kind === 'zram' && memory.zramMemUsed > 0
                    ? `${of(swap.used, swap.size)} · ${t('settings.performance.memory.zramRam', {
                        ram: formatBytes(memory.zramMemUsed)
                      })}`
                    : of(swap.used, swap.size)
                }
              />
            ))
          )}

          {showIon && (
            <div className="flex flex-col space-y-1">
              <Row
                label={t('settings.performance.memory.video')}
                tip={t('settings.performance.memory.videoTip')}
                value={`${of(ion.used, ion.total)} (${ion.usageRate}%)`}
                className={
                  ion.verdict === 'critical'
                    ? 'text-red-400'
                    : ion.verdict === 'warn'
                      ? 'text-amber-400'
                      : undefined
                }
              />
              {ion.generations > 1 && (
                <span className="text-xs text-neutral-500">
                  {t('settings.performance.memory.videoGenerations', {
                    count: ion.generations - 1
                  })}{' '}
                  {t('settings.performance.memory.videoReboot')}
                </span>
              )}
            </div>
          )}
        </Box>
      )}

      {memory && (memory.processes.length > 0 || memory.addons) && (
        <>
          <span className="text-xs text-neutral-500">
            {t('settings.performance.memory.consumers')}
          </span>
          <Box>
            {memory.processes.map((p) => (
              <Row key={p.name} label={p.name} value={formatBytes(p.rss)} />
            ))}
            {memory.addons && addons && (
              <Row
                label={t('settings.performance.memory.addons')}
                tip={t('settings.performance.memory.addonsTip')}
                value={
                  addons.limit > 0
                    ? of(memory.addons.current, addons.limit)
                    : formatBytes(memory.addons.current)
                }
                className={addons.pressed ? 'text-amber-500' : undefined}
              />
            )}
          </Box>
        </>
      )}
    </Section>
  );
};
