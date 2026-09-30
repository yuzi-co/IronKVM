import { useState } from 'react';
import clsx from 'clsx';
import { ChevronDownIcon, ChevronRightIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { Peer } from './types.ts';
import { countIdle, countOnline, sortPeers } from './view.ts';

type PeersProps = {
  peers: Peer[];
};

// Peers is folded by default under a count. Open, it lists the online peers
// and NetBird's peers on demand; the offline ones, often many old machines,
// wait behind a link.
export const Peers = ({ peers }: PeersProps) => {
  const { t } = useTranslation();

  const [isOpen, setIsOpen] = useState(false);
  const [showOffline, setShowOffline] = useState(false);

  const online = countOnline(peers);
  const idle = countIdle(peers);
  const offline = peers.length - online - idle;
  const shown = sortPeers(peers).filter((p) => p.online || p.idle || showOffline);

  return (
    <section className="flex flex-col space-y-3">
      <button
        type="button"
        aria-expanded={isOpen}
        className="flex w-full cursor-pointer items-center space-x-1 p-0 text-left text-sm font-medium select-none"
        onClick={() => setIsOpen(!isOpen)}
      >
        {isOpen ? <ChevronDownIcon size={16} /> : <ChevronRightIcon size={16} />}
        <span>
          {idle > 0
            ? t('settings.vpn.peersSummaryIdle', { online, idle, total: peers.length })
            : t('settings.vpn.peersSummary', { online, total: peers.length })}
        </span>
      </button>

      {isOpen && (
        <div className="flex flex-col space-y-2 pl-5">
          {peers.length === 0 && (
            <span className="text-xs text-neutral-500">{t('settings.vpn.noPeers')}</span>
          )}

          {shown.length > 0 && (
            <ul className="flex flex-col space-y-1">
              {shown.map((peer) => {
                const state = peer.online
                  ? t('settings.vpn.online')
                  : peer.idle
                    ? t('settings.vpn.idle')
                    : t('settings.vpn.offline');
                return (
                  <li
                    key={`${peer.name}-${peer.ip}`}
                    className="flex items-center justify-between text-sm"
                  >
                    <span className="flex items-center space-x-2">
                      <span
                        role="img"
                        aria-label={state}
                        title={state}
                        className={clsx(
                          'inline-block h-2 w-2 rounded-full',
                          peer.online
                            ? 'bg-green-500'
                            : peer.idle
                              ? 'border border-green-500'
                              : 'bg-neutral-600'
                        )}
                      />
                      <span className={clsx(!peer.online && !peer.idle && 'text-neutral-500')}>
                        {peer.name}
                      </span>
                    </span>
                    <span className="font-mono text-neutral-400">{peer.ip}</span>
                  </li>
                );
              })}
            </ul>
          )}

          {offline > 0 && (
            <button
              type="button"
              className="w-fit cursor-pointer p-0 text-left text-xs text-neutral-500 hover:text-neutral-300"
              onClick={() => setShowOffline(!showOffline)}
            >
              {showOffline
                ? t('settings.vpn.hideOffline')
                : t('settings.vpn.showOffline', { offline })}
            </button>
          )}
        </div>
      )}
    </section>
  );
};
