import { useState } from 'react';
import { Alert, Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure } from '@/lib/feedback.ts';

import { report } from './report.ts';
import { Section } from './section.tsx';
import type { SshState } from './types.ts';

type KeysOnlyProps = {
  state: SshState;
  keyCount: number;
  onChange: () => Promise<unknown>;
};

// KeysOnly turns password login off. It cannot be turned on with no key to
// log in with, and says so instead of leaving a switch that fails.
export const KeysOnly = ({ state, keyCount, onChange }: KeysOnlyProps) => {
  const { t } = useTranslation();
  const [isSaving, setIsSaving] = useState(false);

  const noKeys = keyCount === 0 && !state.keysOnly;

  async function toggle(enabled: boolean) {
    setIsSaving(true);
    try {
      const rsp = await api.setSSHKeysOnly(enabled);
      report(
        rsp,
        {
          [-2]: t('settings.ssh.keysOnlyNeedsKey'),
          [-3]: t('settings.ssh.notHonoured'),
          [-4]: t('settings.ssh.reloadFailed')
        },
        t(enabled ? 'settings.ssh.keysOnlyOn' : 'settings.ssh.keysOnlyOff')
      );
    } catch (err) {
      showFailure(err);
    } finally {
      await onChange();
      setIsSaving(false);
    }
  }

  // The setting is on, yet sshd says it still takes passwords: an sshd that
  // was not reloaded, or one that ignores the drop-in.
  const notApplied = state.keysOnly && state.running && state.passwordAuth === 'yes';

  return (
    <Section
      title={t('settings.ssh.keysOnly')}
      description={t('settings.ssh.keysOnlyDesc')}
      action={
        <Switch checked={state.keysOnly} disabled={noKeys} loading={isSaving} onChange={toggle} />
      }
    >
      {noKeys && (
        <span className="text-xs text-amber-500">{t('settings.ssh.keysOnlyNeedsKey')}</span>
      )}
      {notApplied && <Alert type="warning" showIcon message={t('settings.ssh.notApplied')} />}
    </Section>
  );
};
