import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

import type { Peer } from './types.ts';

type PeersProps = {
  peers: Peer[];
};

export const Peers = ({ peers }: PeersProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col space-y-2">
      <span>{t('settings.vpn.peers')}</span>

      {peers.length === 0 ? (
        <span className="text-sm text-neutral-500">{t('settings.vpn.noPeers')}</span>
      ) : (
        <ul className="flex flex-col space-y-1">
          {peers.map((peer) => (
            <li
              key={`${peer.name}-${peer.ip}`}
              className="flex items-center justify-between text-sm"
            >
              <span className="flex items-center space-x-2">
                <span
                  role="img"
                  aria-label={peer.online ? t('settings.vpn.online') : t('settings.vpn.offline')}
                  title={peer.online ? t('settings.vpn.online') : t('settings.vpn.offline')}
                  className={clsx(
                    'inline-block h-2 w-2 rounded-full',
                    peer.online ? 'bg-green-500' : 'bg-neutral-600'
                  )}
                />
                <span>{peer.name}</span>
              </span>
              <span className="font-mono text-neutral-400">{peer.ip}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};
