import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import * as netbird from '@/api/extensions/netbird.ts';
import * as tailscale from '@/api/extensions/tailscale.ts';
import { getHostname } from '@/lib/service.ts';

import { CopyBlock, CopyRow } from '../components/copy-button.tsx';
import { sshCommand } from './command.ts';
import { Box, Section } from './section.tsx';
import type { SshState } from './types.ts';

type VpnAddress = { name: string; ip: string };

const vpns = [
  { name: 'Tailscale', getStatus: tailscale.getStatus },
  { name: 'NetBird', getStatus: netbird.getStatus }
];

// vpnAddresses asks each VPN add-on for its address and keeps the ones that
// are up. An add-on that is not installed or not reachable is left out.
async function vpnAddresses(): Promise<VpnAddress[]> {
  const results = await Promise.allSettled(
    vpns.map(async ({ name, getStatus }) => {
      const rsp = await getStatus();
      const ip = rsp.code === 0 && rsp.data?.state === 'running' ? (rsp.data.ip as string) : '';
      return { name, ip };
    })
  );
  return results.flatMap((result) =>
    result.status === 'fulfilled' && result.value.ip ? [result.value] : []
  );
}

// Connection says how to reach sshd: the command for the address this page
// was opened on, the port, and the VPN addresses when a VPN is up.
export const Connection = ({ state }: { state: SshState }) => {
  const { t } = useTranslation();
  const [addresses, setAddresses] = useState<VpnAddress[]>([]);

  useEffect(() => {
    let alive = true;
    vpnAddresses().then((found) => {
      if (alive) setAddresses(found);
    });
    return () => {
      alive = false;
    };
  }, []);

  const host = getHostname();

  return (
    <Section title={t('settings.ssh.connection')}>
      {state.running ? (
        <>
          <CopyBlock title={t('settings.ssh.command')} text={sshCommand(host, state.port)} />
          <Box>
            <CopyRow label={t('settings.ssh.port')} value={String(state.port)} />
            {addresses.map(({ name, ip }) => (
              <CopyRow
                key={name}
                label={t('settings.ssh.viaVpn', { name })}
                value={sshCommand(ip, state.port)}
                copyLabel={name}
              />
            ))}
          </Box>
        </>
      ) : (
        <span className="text-xs text-neutral-500">{t('settings.ssh.notRunning')}</span>
      )}
    </Section>
  );
};
