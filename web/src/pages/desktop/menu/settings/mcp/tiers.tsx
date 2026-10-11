import { useEffect, useState } from 'react';
import { Alert, Checkbox } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/audit.ts';
import type { AgentTier, AgentTiers } from '@/api/audit.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { Section } from '../components/section.tsx';

type Subject = 'mcp' | 'agent';

// CapabilityTiers chooses what MCP clients and the on-board agent may do. A
// change is saved at once; the server applies it to the next call. The page
// shows it only on a server with the agent-tiers feature.
export const CapabilityTiers = () => {
  const { t } = useTranslation();

  const [tiers, setTiers] = useState<AgentTiers | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .getAgentTiers()
      .then((rsp) => {
        if (rsp.code !== 0) throw rsp;
        setTiers(rsp.data as AgentTiers);
      })
      .catch((err) => setError(describeFailure(err, t('settings.mcp.tiers.loadFailed'))));
  }, [t]);

  async function toggle(subject: Subject, tier: AgentTier, on: boolean) {
    if (!tiers || isSaving) return;
    const next = {
      mcp: tiers.mcp,
      agent: tiers.agent,
      [subject]: on ? [...tiers[subject], tier] : tiers[subject].filter((have) => have !== tier)
    };
    setError('');
    setIsSaving(true);
    try {
      const rsp = await api.setAgentTiers(next.mcp, next.agent);
      if (rsp.code !== 0) throw rsp;
      setTiers(rsp.data as AgentTiers);
    } catch (err) {
      setError(describeFailure(err, t('settings.mcp.tiers.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  const groups: Subject[] = ['mcp', 'agent'];

  return (
    <Section
      title={t('settings.mcp.tiers.title')}
      description={t('settings.mcp.tiers.description')}
    >
      {tiers && (
        <div className="grid gap-4 sm:grid-cols-2">
          {groups.map((subject) => (
            <div
              key={subject}
              className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3"
            >
              <span className="text-sm font-medium">{t(`settings.mcp.tiers.${subject}`)}</span>
              {tiers.available.map((tier) => (
                <Checkbox
                  key={tier}
                  checked={tiers[subject].includes(tier)}
                  disabled={isSaving}
                  onChange={(e) => void toggle(subject, tier, e.target.checked)}
                >
                  <span className="flex flex-col">
                    <span className="text-sm">{t(`settings.mcp.tiers.${tier}`)}</span>
                    <span className="text-xs text-neutral-500">
                      {t(`settings.mcp.tiers.${tier}Desc`)}
                    </span>
                  </span>
                </Checkbox>
              ))}
            </div>
          ))}
        </div>
      )}

      <span className="text-xs text-neutral-500">{t('settings.mcp.tiers.agentNote')}</span>
      {error && <Alert type="error" showIcon message={error} />}
    </Section>
  );
};
