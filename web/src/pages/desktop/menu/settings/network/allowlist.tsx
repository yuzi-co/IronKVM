import { useEffect, useState } from 'react';
import { Button, Input } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/network.ts';
import type { AllowlistState } from '@/api/network.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { Section } from '../components/section.tsx';
import { formatNetworks, parseNetworks, sameNetworks, withClient } from './allowlist-text.ts';

// AllowedNetworks is the client address allowlist. The server refuses a list
// that leaves out the address saving it, so this page cannot lock its own
// user out; it shows that address and offers to add it. The page shows it
// only on a server with the ip-allowlist feature.
export const AllowedNetworks = () => {
  const { t } = useTranslation();

  const [state, setState] = useState<AllowlistState | null>(null);
  const [text, setText] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .getAllowlist()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(rsp.msg || t('settings.network.allowlist.loadFailed'));
          return;
        }
        const data = rsp.data as AllowlistState;
        setState(data);
        setText(formatNetworks(data.networks));
      })
      .catch((err) => setError(describeFailure(err, t('settings.network.allowlist.loadFailed'))));
  }, [t]);

  const isChanged = !!state && !sameNetworks(text, state.networks);

  async function save(networks: string[]) {
    if (isSaving) return;
    setError('');
    setIsSaving(true);
    try {
      const rsp = await api.setAllowlist(networks);
      if (rsp.code !== 0) {
        setError(rsp.msg || t('settings.network.allowlist.saveFailed'));
        return;
      }
      const data = rsp.data as AllowlistState;
      setState(data);
      setText(formatNetworks(data.networks));
    } catch (err) {
      setError(describeFailure(err, t('settings.network.allowlist.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  const restricted = !!state && state.networks.length > 0;

  return (
    <Section
      title={t('settings.network.allowlist.title')}
      description={t('settings.network.allowlist.description')}
    >
      {state && (
        <span className={`text-xs ${restricted ? 'text-yellow-400/80' : 'text-neutral-500'}`}>
          {restricted
            ? t('settings.network.allowlist.restricted')
            : t('settings.network.allowlist.everyone')}
        </span>
      )}

      <Input.TextArea
        value={text}
        onChange={(e) => {
          setText(e.target.value);
          setError('');
        }}
        placeholder={t('settings.network.allowlist.placeholder')}
        autoSize={{ minRows: 2, maxRows: 8 }}
        disabled={!state || isSaving}
        className="font-mono"
        spellCheck={false}
        autoComplete="off"
      />

      {state?.client && (
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs text-neutral-500">
            {t('settings.network.allowlist.yourAddress', { client: state.client })}
          </span>
          <Button
            size="small"
            disabled={isSaving || parseNetworks(text).includes(state.client)}
            onClick={() => setText(withClient(text, state.client))}
          >
            {t('settings.network.allowlist.addMine')}
          </Button>
        </div>
      )}

      <div className="flex flex-wrap gap-2">
        <Button
          type={isChanged ? 'primary' : 'default'}
          loading={isSaving}
          disabled={!state || !isChanged}
          onClick={() => save(parseNetworks(text))}
        >
          {t('settings.network.allowlist.save')}
        </Button>
        {restricted && (
          <Button disabled={isSaving} onClick={() => save([])}>
            {t('settings.network.allowlist.allowAll')}
          </Button>
        )}
      </div>

      {error && <span className="text-xs text-red-400">{error}</span>}
    </Section>
  );
};
