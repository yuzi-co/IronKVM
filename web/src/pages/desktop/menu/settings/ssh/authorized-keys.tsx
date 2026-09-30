import { useState } from 'react';
import { Button, Input, Popconfirm, Tooltip } from 'antd';
import { Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure } from '@/lib/feedback.ts';

import { Section } from '../components/section.tsx';
import { keyLabel } from './command.ts';
import { report } from './report.ts';
import type { SshKey } from './types.ts';

type AuthorizedKeysProps = {
  keys: SshKey[] | null;
  keysOnly: boolean;
  onChange: () => Promise<unknown>;
};

// AuthorizedKeys lists the public keys that can log in as root, with a box to
// add one and a button to remove each.
export const AuthorizedKeys = ({ keys, keysOnly, onChange }: AuthorizedKeysProps) => {
  const { t } = useTranslation();
  const [draft, setDraft] = useState('');
  const [isAdding, setIsAdding] = useState(false);
  const [removing, setRemoving] = useState('');

  // The server refuses the same; this says why before the click.
  const isLastUnderKeysOnly = keysOnly && keys?.length === 1;

  async function add() {
    const key = draft.trim();
    if (!key) return;
    setIsAdding(true);
    try {
      const rsp = await api.addSSHKey(key);
      const ok = report(
        rsp,
        {
          [-2]: t('settings.ssh.invalidKey'),
          [-3]: t('settings.ssh.keyOptions'),
          [-4]: t('settings.ssh.duplicateKey')
        },
        t('settings.ssh.added')
      );
      if (ok) setDraft('');
    } catch (err) {
      showFailure(err);
    } finally {
      await onChange();
      setIsAdding(false);
    }
  }

  async function remove(fingerprint: string) {
    setRemoving(fingerprint);
    try {
      const rsp = await api.deleteSSHKey(fingerprint);
      report(rsp, { [-3]: t('settings.ssh.lastKey') }, t('settings.ssh.removed'));
    } catch (err) {
      showFailure(err);
    } finally {
      await onChange();
      setRemoving('');
    }
  }

  return (
    <Section title={t('settings.ssh.keys')} description={t('settings.ssh.keysDesc')}>
      {keys && keys.length > 0 ? (
        <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
          {keys.map((key, index) => (
            <div
              key={key.fingerprint}
              className={
                index > 0
                  ? 'flex items-start justify-between gap-2 border-t border-neutral-800 px-4 py-2.5'
                  : 'flex items-start justify-between gap-2 px-4 py-2.5'
              }
            >
              <div className="flex min-w-0 flex-col space-y-0.5">
                <div className="flex min-w-0 items-center gap-2 text-sm">
                  <span className="shrink-0 text-neutral-400">{keyLabel(key.type)}</span>
                  <span className="min-w-0 break-all text-neutral-200">
                    {key.comment || (
                      <span className="text-neutral-500 italic">{t('settings.ssh.noComment')}</span>
                    )}
                  </span>
                </div>
                <span className="font-mono text-xs break-all text-neutral-400 select-all">
                  {key.fingerprint}
                </span>
              </div>

              <Popconfirm
                title={t('settings.ssh.deleteConfirm')}
                description={t('settings.ssh.deleteConfirmDesc')}
                okText={t('common.remove')}
                cancelText={t('common.cancel')}
                okButtonProps={{ danger: true }}
                disabled={isLastUnderKeysOnly}
                onConfirm={() => remove(key.fingerprint)}
              >
                <Tooltip
                  title={isLastUnderKeysOnly ? t('settings.ssh.lastKey') : t('common.remove')}
                >
                  <Button
                    type="text"
                    size="small"
                    danger
                    className="shrink-0"
                    aria-label={t('common.remove')}
                    disabled={isLastUnderKeysOnly}
                    loading={removing === key.fingerprint}
                    icon={<Trash2Icon size={15} />}
                  />
                </Tooltip>
              </Popconfirm>
            </div>
          ))}
        </div>
      ) : (
        keys && <span className="text-xs text-neutral-500">{t('settings.ssh.noKeys')}</span>
      )}

      <Input.TextArea
        value={draft}
        autoSize={{ minRows: 2, maxRows: 6 }}
        spellCheck={false}
        className="font-mono text-xs break-all"
        placeholder={t('settings.ssh.addPlaceholder')}
        onChange={(e) => setDraft(e.target.value)}
      />
      <div className="flex justify-end">
        <Button type="primary" disabled={!draft.trim()} loading={isAdding} onClick={add}>
          {t('settings.ssh.add')}
        </Button>
      </div>
    </Section>
  );
};
