import { useTranslation } from 'react-i18next';

import { CopyRow } from '../components/copy-button.tsx';
import { keyLabel } from './command.ts';
import { Box, Section } from './section.tsx';
import type { SshKey } from './types.ts';

// HostKeys lists the board's host key fingerprints, to check against what ssh
// prints on the first connection.
export const HostKeys = ({ keys }: { keys: SshKey[] }) => {
  const { t } = useTranslation();

  return (
    <Section title={t('settings.ssh.hostKeys')} description={t('settings.ssh.hostKeysDesc')}>
      {keys.length > 0 ? (
        <Box>
          {keys.map((key) => (
            <CopyRow
              key={key.fingerprint}
              label={keyLabel(key.type)}
              value={key.fingerprint}
              copyLabel={keyLabel(key.type)}
            />
          ))}
        </Box>
      ) : (
        <span className="text-xs text-neutral-500">{t('settings.ssh.noHostKeys')}</span>
      )}
    </Section>
  );
};
