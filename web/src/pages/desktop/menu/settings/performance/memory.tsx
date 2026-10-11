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
import {
  groupUse,
  percentOf,
  ramUse,
  type MemoryCma,
  type MemoryPool,
  type MemoryStatus,
  type MemorySwap
} from './memory.ts';

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
// memory (the ION carveout on the vendor kernel, the reserved pool and the CMA
// areas on mainline), and what uses the most of it. It is read again every few seconds
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
  // The mainline kernel has no carveout. It reports its video pool and CMA
  // areas with the rest of the memory instead.
  const pools = memory?.videoPools ?? [];
  const cmaAreas = memory?.cma ?? [];

  // detail is the line under a pool or CMA row: the longest free run, which
  // decides whether the next large buffer fits, and the pool's peak.
  function detail(largestFree: number | undefined, peak: number) {
    const parts: string[] = [];
    if (largestFree !== undefined) {
      parts.push(t('settings.performance.memory.largestFree', { size: formatBytes(largestFree) }));
    }
    if (peak > 0) {
      parts.push(t('settings.performance.memory.peak', { size: formatBytes(peak) }));
    }
    return parts.length > 0 ? (
      <span className="text-xs text-neutral-500">{parts.join(' · ')}</span>
    ) : null;
  }

  function poolValue(pool: MemoryPool) {
    if (!pool.measured) {
      return t('settings.performance.memory.videoUnmeasured', { size: formatBytes(pool.size) });
    }
    return `${of(pool.used, pool.size)} (${percentOf(pool.used, pool.size)}%)`;
  }

  function cmaValue(area: MemoryCma) {
    return `${of(area.used, area.total)} (${percentOf(area.used, area.total)}%)`;
  }

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

          {pools.map((pool) => (
            <div key={pool.name} className="flex flex-col space-y-1">
              <Row
                label={
                  pools.length > 1
                    ? `${t('settings.performance.memory.video')} (${pool.name})`
                    : t('settings.performance.memory.video')
                }
                tip={t('settings.performance.memory.videoTip')}
                value={poolValue(pool)}
              />
              {pool.measured && detail(pool.largestFree, pool.peak)}
            </div>
          ))}

          {cmaAreas.map((area) => (
            <div key={area.name} className="flex flex-col space-y-1">
              <Row
                label={
                  cmaAreas.length > 1
                    ? `${t('settings.performance.memory.cma')} (${area.name})`
                    : t('settings.performance.memory.cma')
                }
                tip={t('settings.performance.memory.cmaTip')}
                value={cmaValue(area)}
              />
              {detail(area.largestFree, 0)}
            </div>
          ))}
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
