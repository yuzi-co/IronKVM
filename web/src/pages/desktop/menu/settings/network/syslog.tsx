import { useEffect, useState } from 'react';
import { Button, Input } from 'antd';
import { useSetAtom } from 'jotai';
import { Trans, useTranslation } from 'react-i18next';

import * as api from '@/api/network.ts';
import type { SyslogState } from '@/api/network.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { settingsOpenRequestAtom } from '@/jotai/settings.ts';

import { CopyRow } from '../components/copy-button.tsx';
import { Box, Section } from '../components/section.tsx';
import { syslogTargetError } from './syslog-target.ts';

const value = <span className="font-mono text-xs text-neutral-300" />;

// StateLine says what the logger does now, and where that differs from what
// is saved: a restart that failed, or a setting written by hand.
const StateLine = ({ state }: { state: SyslogState }) => {
  const { t } = useTranslation();

  if (!state.supported) {
    return (
      <span className="text-xs text-yellow-400/80">{t('settings.network.syslog.unsupported')}</span>
    );
  }

  const { target, active } = state;
  let key: string;
  if (target === active) {
    key = target ? 'settings.network.syslog.forwarding' : 'settings.network.syslog.local';
  } else if (!active) {
    key = 'settings.network.syslog.savedButLocal';
  } else if (!target) {
    key = 'settings.network.syslog.offButForwarding';
  } else {
    key = 'settings.network.syslog.savedButForwarding';
  }

  return (
    <span className={`text-xs ${target === active ? 'text-neutral-500' : 'text-yellow-400/80'}`}>
      <Trans i18nKey={key} values={{ target, active }} components={{ v: value }} />
    </span>
  );
};

export const Syslog = () => {
  const { t } = useTranslation();
  const requestSettings = useSetAtom(settingsOpenRequestAtom);

  const [state, setState] = useState<SyslogState | null>(null);
  const [input, setInput] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [error, setError] = useState('');
  const [sent, setSent] = useState('');

  useEffect(() => {
    api
      .getSyslog()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(rsp.msg || t('settings.network.syslog.loadFailed'));
          return;
        }
        const data = rsp.data as SyslogState;
        setState(data);
        setInput(data.target);
      })
      .catch((err) => setError(describeFailure(err, t('settings.network.syslog.loadFailed'))));
  }, [t]);

  const trimmed = input.trim();
  const reason = trimmed ? syslogTargetError(trimmed) : null;
  const isChanged = !!state && trimmed !== state.target;

  async function save(target: string) {
    if (isSaving) return;
    setError('');
    setSent('');
    setIsSaving(true);
    try {
      const rsp = await api.setSyslog(target);
      if (rsp.code !== 0) {
        setError(rsp.msg || t('settings.network.syslog.saveFailed'));
        return;
      }
      const data = rsp.data as SyslogState;
      setState(data);
      setInput(data.target);
    } catch (err) {
      setError(describeFailure(err, t('settings.network.syslog.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  async function sendTest() {
    if (isTesting) return;
    setError('');
    setSent('');
    setIsTesting(true);
    try {
      const rsp = await api.testSyslog();
      if (rsp.code !== 0) {
        setError(rsp.msg || t('settings.network.syslog.testFailed'));
        return;
      }
      setSent((rsp.data as { message: string }).message);
    } catch (err) {
      setError(describeFailure(err, t('settings.network.syslog.testFailed')));
    } finally {
      setIsTesting(false);
    }
  }

  const metricsUrl = `${getBaseUrl('http')}/api/metrics`;

  return (
    <Section
      title={t('settings.network.syslog.title')}
      description={t('settings.network.syslog.description')}
    >
      {state && <StateLine state={state} />}

      <div className="flex flex-wrap items-start gap-2">
        <div className="flex min-w-[220px] flex-1 flex-col space-y-1">
          <Input
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              setError('');
            }}
            onPressEnter={() => isChanged && !reason && save(trimmed)}
            placeholder={t('settings.network.syslog.placeholder')}
            status={reason ? 'error' : undefined}
            disabled={!state || isSaving}
            className="font-mono"
            spellCheck={false}
            autoComplete="off"
          />
          {reason && (
            <span className="text-xs text-red-400">
              {t(`settings.network.syslog.errors.${reason}`)}
            </span>
          )}
        </div>

        <Button
          type={isChanged ? 'primary' : 'default'}
          loading={isSaving}
          disabled={!state || !isChanged || !!reason}
          onClick={() => save(trimmed)}
        >
          {t('settings.network.syslog.save')}
        </Button>
        {state?.target && (
          <Button disabled={isSaving} onClick={() => save('')}>
            {t('settings.network.syslog.turnOff')}
          </Button>
        )}
      </div>

      {error && <span className="text-xs text-red-400">{error}</span>}

      <div className="flex flex-col space-y-1">
        <div>
          <Button size="small" loading={isTesting} disabled={!state} onClick={sendTest}>
            {t('settings.network.syslog.test')}
          </Button>
        </div>
        {sent && (
          <span className="text-xs text-neutral-500">
            <Trans
              i18nKey="settings.network.syslog.sent"
              values={{ message: sent }}
              components={{ v: value }}
            />
          </span>
        )}
      </div>

      <Box>
        <span className="text-sm font-medium">{t('settings.network.syslog.metrics')}</span>
        <span className="text-xs text-neutral-500">
          <Trans
            i18nKey="settings.network.syslog.metricsDesc"
            components={{
              link: (
                <a
                  className="text-blue-400 hover:underline"
                  href="#"
                  onClick={(e) => {
                    e.preventDefault();
                    requestSettings('apiKeys');
                  }}
                />
              )
            }}
          />
        </span>
        <CopyRow label={t('settings.network.syslog.metricsUrl')} value={metricsUrl} />
      </Box>
    </Section>
  );
};
