import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert, Button, Input, message, Popconfirm, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import type { CertInfo, TLSCertState } from '@/api/vm.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { CopyRow } from '../components/copy-button.tsx';
import { Box, Section } from '../components/section.tsx';

// The server refuses anything larger, so the page does not read it.
const maxFileSize = 64 * 1024;

const errorKinds = [
  'tooLarge',
  'noCertificate',
  'badCertificate',
  'noKey',
  'encryptedKey',
  'badKey',
  'keyMismatch',
  'expired',
  'notYetValid',
  'badChain',
  'weakKey',
  'notServerCert'
];

function formatDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

const Row = ({ label, value }: { label: string; value: string }) => (
  <div className="flex flex-wrap items-start justify-between gap-x-3">
    <span className="text-neutral-400">{label}</span>
    <span className="min-w-0 text-right break-all text-neutral-300">{value}</span>
  </div>
);

const CertDetails = ({ info }: { info: CertInfo }) => {
  const { t } = useTranslation();
  const names = [...(info.dnsNames ?? []), ...(info.ipAddresses ?? [])];
  return (
    <div className="flex flex-col space-y-2">
      <Box>
        <Row label={t('tlsCert.subject')} value={info.subject || '-'} />
        <Row label={t('tlsCert.issuer')} value={info.issuer || '-'} />
        <Row label={t('tlsCert.names')} value={names.length ? names.join(', ') : '-'} />
        <Row label={t('tlsCert.validFrom')} value={formatDate(info.notBefore)} />
        <Row label={t('tlsCert.validUntil')} value={formatDate(info.notAfter)} />
        <Row label={t('tlsCert.keyType')} value={info.keyType} />
        <Row label={t('tlsCert.chain')} value={String(info.chainLength)} />
        <CopyRow label={t('tlsCert.fingerprint')} value={info.fingerprint} />
      </Box>
      {info.warnings?.includes('expiresSoon') && (
        <Alert type="warning" showIcon message={t('tlsCert.warnings.expiresSoon')} />
      )}
      {info.warnings?.includes('nameMismatch') && (
        <Alert
          type="warning"
          showIcon
          message={t('tlsCert.warnings.nameMismatch', { names: (info.uncovered ?? []).join(', ') })}
        />
      )}
      {info.warnings?.includes('selfSigned') && (
        <Alert type="info" showIcon message={t('tlsCert.warnings.selfSigned')} />
      )}
    </div>
  );
};

type PemFieldProps = {
  label: string;
  value: string;
  placeholder: string;
  accept: string;
  onChange: (value: string) => void;
};

// PemField is a text box for PEM, with a button that fills it from a file.
// The file is read in the browser; nothing is sent until install.
const PemField = ({ label, value, placeholder, accept, onChange }: PemFieldProps) => {
  const { t } = useTranslation();
  const input = useRef<HTMLInputElement>(null);

  function readFile(file: File | undefined) {
    if (!file) return;
    if (file.size > maxFileSize) {
      message.error(t('tlsCert.errors.tooLarge'));
      return;
    }
    file.text().then(onChange, () => message.error(t('tlsCert.readFailed')));
  }

  return (
    <div className="flex flex-col space-y-1">
      <div className="flex items-center justify-between">
        <span className="text-sm text-neutral-400">{label}</span>
        <Button size="small" onClick={() => input.current?.click()}>
          {t('tlsCert.chooseFile')}
        </Button>
        <input
          ref={input}
          type="file"
          accept={accept}
          className="hidden"
          onChange={(event) => {
            readFile(event.target.files?.[0]);
            event.target.value = '';
          }}
        />
      </div>
      <Input.TextArea
        value={value}
        rows={4}
        spellCheck={false}
        autoComplete="off"
        className="font-mono text-xs!"
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
};

// Certificate shows the HTTPS certificate in force and installs an uploaded
// one, or goes back to the generated self-signed one. The server serves a new
// certificate from the next connection on, without a restart.
export const Certificate = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();

  const [state, setState] = useState<TLSCertState | null>(null);
  const [cert, setCert] = useState('');
  const [key, setKey] = useState('');
  const [preview, setPreview] = useState<CertInfo | null>(null);
  const [isBusy, setIsBusy] = useState(false);

  const load = useCallback(
    () =>
      api
        .getTLSCert()
        .then((rsp) => {
          if (rsp.code !== 0) throw rsp;
          setState(rsp.data as TLSCertState);
        })
        .catch((err) => {
          messageApi.error(describeFailure(err, t('tlsCert.loadFailed')));
        }),
    [messageApi, t]
  );

  useEffect(() => {
    load();
  }, [load]);

  // failure names a refused certificate in the operator's language when the
  // server says which check it failed, and in the server's words otherwise.
  function failure(err: unknown, fallback: string) {
    const rsp = err as { code?: number; msg?: string; data?: { kind?: string } };
    const kind = rsp?.data?.kind;
    if (kind && errorKinds.includes(kind)) {
      return t(`tlsCert.errors.${kind}`);
    }
    return describeFailure(err, fallback);
  }

  function edit(setter: (value: string) => void) {
    return (value: string) => {
      setter(value);
      setPreview(null);
    };
  }

  async function check() {
    setIsBusy(true);
    try {
      const rsp = await api.checkTLSCert(cert, key);
      if (rsp.code !== 0) throw rsp;
      setPreview(rsp.data as CertInfo);
    } catch (err) {
      setPreview(null);
      messageApi.error(failure(err, t('tlsCert.checkFailed')));
    } finally {
      setIsBusy(false);
    }
  }

  async function install() {
    setIsBusy(true);
    try {
      const rsp = await api.setTLSCert(cert, key);
      if (rsp.code !== 0) throw rsp;
      // The key leaves the page's memory once it is installed.
      setCert('');
      setKey('');
      setPreview(null);
      messageApi.success(t('tlsCert.installed'));
      await load();
    } catch (err) {
      messageApi.error(failure(err, t('tlsCert.installFailed')));
    } finally {
      setIsBusy(false);
    }
  }

  async function restore() {
    setIsBusy(true);
    try {
      const rsp = await api.restoreTLSCert();
      if (rsp.code !== 0) throw rsp;
      messageApi.success(t('tlsCert.restored'));
      await load();
    } catch (err) {
      messageApi.error(describeFailure(err, t('tlsCert.restoreFailed')));
    } finally {
      setIsBusy(false);
    }
  }

  const canSubmit = cert.trim() !== '' && key.trim() !== '' && !isBusy;

  return (
    <div className="flex flex-col space-y-6">
      {contextHolder}

      <Section
        title={t('tlsCert.title')}
        description={t('tlsCert.description')}
        action={
          state && (
            <Tag color={state.uploaded ? 'blue' : undefined}>
              {state.uploaded ? t('tlsCert.uploaded') : t('tlsCert.selfSigned')}
            </Tag>
          )
        }
      >
        {state && !state.https && <Alert type="info" showIcon message={t('tlsCert.httpsOff')} />}
        {state?.info ? (
          <CertDetails info={state.info} />
        ) : (
          state && <span className="text-sm text-neutral-500">{t('tlsCert.none')}</span>
        )}
        {state?.uploaded && (
          <div>
            <Popconfirm
              title={t('tlsCert.restoreConfirm')}
              description={<div className="max-w-[320px]">{t('tlsCert.restoreDesc')}</div>}
              okText={t('tlsCert.restore')}
              cancelText={t('common.cancel')}
              onConfirm={restore}
            >
              <Button danger disabled={isBusy}>
                {t('tlsCert.restore')}
              </Button>
            </Popconfirm>
          </div>
        )}
      </Section>

      <Section title={t('tlsCert.uploadTitle')} description={t('tlsCert.uploadDescription')}>
        <PemField
          label={t('tlsCert.certField')}
          value={cert}
          accept=".pem,.crt,.cer,.chain"
          placeholder="-----BEGIN CERTIFICATE-----"
          onChange={edit(setCert)}
        />
        <PemField
          label={t('tlsCert.keyField')}
          value={key}
          accept=".pem,.key"
          placeholder="-----BEGIN PRIVATE KEY-----"
          onChange={edit(setKey)}
        />

        {preview && <CertDetails info={preview} />}

        <div className="flex flex-wrap justify-end gap-2">
          <Button disabled={!canSubmit} loading={isBusy && !preview} onClick={check}>
            {t('tlsCert.check')}
          </Button>
          <Popconfirm
            title={t('tlsCert.installConfirm')}
            description={<div className="max-w-[320px]">{t('tlsCert.installDesc')}</div>}
            okText={t('tlsCert.install')}
            cancelText={t('common.cancel')}
            disabled={!canSubmit}
            onConfirm={install}
          >
            <Button type="primary" disabled={!canSubmit}>
              {t('tlsCert.install')}
            </Button>
          </Popconfirm>
        </div>
      </Section>
    </div>
  );
};
