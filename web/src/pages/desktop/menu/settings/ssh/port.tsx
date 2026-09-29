import { useState } from 'react';
import { Button, InputNumber, Popconfirm } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure } from '@/lib/feedback.ts';

import { isPortChange } from './command.ts';
import { report } from './report.ts';

type PortRowProps = {
  port: number;
  onChange: () => Promise<unknown>;
};

// PortRow shows the port sshd listens on and changes it. The change asks
// first: open sessions survive it, but a client or a firewall still set for
// the old port does not.
export const PortRow = ({ port, onChange }: PortRowProps) => {
  const { t } = useTranslation();
  const [draft, setDraft] = useState<number | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const isEditing = draft !== null;

  async function save() {
    if (!isPortChange(draft, port)) return;
    setIsSaving(true);
    try {
      const rsp = await api.setSSHPort(draft);
      const ok = report(
        rsp,
        {
          [-2]: t('settings.ssh.portInvalid'),
          [-3]: t('settings.ssh.portReserved'),
          [-4]: t('settings.ssh.portInUse'),
          [-5]: t('settings.ssh.portNotHonoured'),
          [-6]: t('settings.ssh.reloadFailed')
        },
        t('settings.ssh.portChanged', { port: draft })
      );
      if (ok) setDraft(null);
    } catch (err) {
      showFailure(err);
    } finally {
      await onChange();
      setIsSaving(false);
    }
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-2">
      <span className="text-neutral-400">{t('settings.ssh.port')}</span>
      {isEditing ? (
        <div className="flex items-center gap-2">
          <InputNumber
            size="small"
            className="w-24"
            min={1}
            max={65535}
            precision={0}
            value={draft}
            autoFocus
            aria-label={t('settings.ssh.port')}
            onChange={(value) => setDraft(value ?? port)}
          />
          <Popconfirm
            title={t('settings.ssh.portConfirm', { port: draft })}
            description={
              <span className="block max-w-72">
                {t('settings.ssh.portConfirmDesc', { port: draft })}
              </span>
            }
            okText={t('common.save')}
            cancelText={t('common.cancel')}
            disabled={!isPortChange(draft, port)}
            onConfirm={save}
          >
            <Button
              size="small"
              type="primary"
              loading={isSaving}
              disabled={!isPortChange(draft, port)}
            >
              {t('common.save')}
            </Button>
          </Popconfirm>
          <Button size="small" disabled={isSaving} onClick={() => setDraft(null)}>
            {t('common.cancel')}
          </Button>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs text-neutral-300">{port}</span>
          <Button size="small" type="link" className="px-0" onClick={() => setDraft(port)}>
            {t('settings.ssh.changePort')}
          </Button>
        </div>
      )}
    </div>
  );
};
