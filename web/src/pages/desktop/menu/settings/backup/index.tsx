import { useRef, useState } from 'react';
import type { ChangeEvent } from 'react';
import { Alert, Button, Checkbox, Divider, message, Modal, Tag } from 'antd';
import { DownloadIcon, FileArchiveIcon, UploadIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/backup.ts';
import type { Download, Preview, SectionPreview } from '@/api/backup.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { rebootAndReload } from '@/lib/reboot.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { Section } from '../components/section.tsx';
import {
  canImport,
  defaultSelection,
  downloadName,
  formatValue,
  isSelectable,
  parseSettingsFile,
  toggle
} from './flow.ts';

const statusColors: Record<string, string> = {
  change: 'blue',
  same: 'default',
  invalid: 'red',
  unknown: 'orange'
};

function save(download: Download, fallback: string) {
  const url = URL.createObjectURL(download.blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = downloadName(download.fileName, fallback);
  link.click();
  URL.revokeObjectURL(url);
}

// Backup holds the two support features: the settings file, exported and
// imported without secrets, and the diagnostics bundle.
export const Backup = () => {
  const { t } = useTranslation();
  const [messageApi, messageHolder] = message.useMessage();
  const [modal, modalHolder] = Modal.useModal();
  const hasSettings = useServerFeature('settings-backup') === true;
  const hasDiagnostics = useServerFeature('diagnostics-bundle') === true;

  const fileInput = useRef<HTMLInputElement>(null);
  const [isExporting, setIsExporting] = useState(false);
  const [isReading, setIsReading] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [settingsDoc, setSettingsDoc] = useState<Record<string, unknown> | null>(null);
  const [preview, setPreview] = useState<Preview | null>(null);
  const [selected, setSelected] = useState<string[]>([]);

  const host = window.location.hostname || 'board';

  async function exportSettings() {
    if (isExporting) return;
    setIsExporting(true);
    try {
      save(await api.exportSettings(), `ironkvm-settings-${host}.json`);
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.backup.exportFailed')));
    } finally {
      setIsExporting(false);
    }
  }

  async function downloadDiagnostics() {
    if (isDownloading) return;
    setIsDownloading(true);
    try {
      save(await api.downloadDiagnostics(), `ironkvm-diagnostics-${host}.tar.gz`);
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.backup.downloadFailed')));
    } finally {
      setIsDownloading(false);
    }
  }

  async function readFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    // The same file may be chosen again after a cancel.
    event.target.value = '';
    if (!file) return;

    setIsReading(true);
    try {
      const parsed = parseSettingsFile(await file.text());
      if (!parsed.ok) {
        messageApi.error(t(`settings.backup.${parsed.reason}`));
        return;
      }
      const rsp = await api.previewSettings(parsed.document);
      if (rsp.code !== 0) {
        messageApi.error(describeFailure(rsp, t('settings.backup.previewFailed')));
        return;
      }
      setSettingsDoc(parsed.document);
      setPreview(rsp.data);
      setSelected(defaultSelection(rsp.data));
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.backup.previewFailed')));
    } finally {
      setIsReading(false);
    }
  }

  function closePreview() {
    setSettingsDoc(null);
    setPreview(null);
    setSelected([]);
  }

  async function importSelected() {
    if (!settingsDoc || !preview || isImporting) return;
    setIsImporting(true);
    try {
      const rsp = await api.importSettings(settingsDoc, selected);
      if (rsp.code !== 0) {
        messageApi.error(describeFailure(rsp, t('settings.backup.importFailed')));
        return;
      }
      closePreview();
      messageApi.success(t('settings.backup.imported'));
      modal.confirm({
        title: t('settings.backup.rebootTitle'),
        content: (
          <span className="text-sm text-neutral-400">{t('settings.backup.rebootDesc')}</span>
        ),
        okText: t('settings.backup.rebootNow'),
        cancelText: t('settings.backup.rebootLater'),
        onOk: () =>
          rebootAndReload((msg) => messageApi.error(msg || t('settings.backup.rebootFailed')))
      });
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.backup.importFailed')));
    } finally {
      setIsImporting(false);
    }
  }

  function renderValue(value: unknown) {
    const shown = formatValue(value);
    return typeof shown === 'string' ? shown : t('settings.backup.items', { n: shown.items });
  }

  function renderSection(section: SectionPreview) {
    const selectable = isSelectable(section);
    return (
      <div
        key={section.section}
        className="flex flex-col space-y-1 border-b border-neutral-700/50 py-2 last:border-b-0"
      >
        <div className="flex items-center justify-between gap-2">
          <Checkbox
            disabled={!selectable}
            checked={selectable && selected.includes(section.section)}
            onChange={(e) => setSelected((cur) => toggle(cur, section.section, e.target.checked))}
          >
            {t(`settings.backup.sections.${section.section}`, { defaultValue: section.section })}
          </Checkbox>
          <Tag color={statusColors[section.status]}>
            {t(`settings.backup.status.${section.status}`)}
          </Tag>
        </div>
        {section.changes?.map((change) => (
          <div key={change.field} className="pl-6 text-xs break-all text-neutral-400">
            <span className="font-mono">{change.field}</span>: {renderValue(change.from)} →{' '}
            <span className="text-neutral-200">{renderValue(change.to)}</span>
          </div>
        ))}
        {section.error && <div className="pl-6 text-xs text-red-400">{section.error}</div>}
        {section.warnings?.map((warning) => (
          <div key={warning} className="pl-6 text-xs text-orange-400">
            {warning}
          </div>
        ))}
      </div>
    );
  }

  return (
    <>
      {messageHolder}
      {modalHolder}
      <div className="text-base">{t('settings.backup.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        {hasSettings && (
          <Section
            title={t('settings.backup.settingsTitle')}
            description={t('settings.backup.settingsDesc')}
          >
            <Alert type="info" showIcon message={t('settings.backup.excluded')} />
            <div className="flex flex-wrap gap-2">
              <Button
                icon={<DownloadIcon size={14} />}
                loading={isExporting}
                onClick={() => void exportSettings()}
              >
                {t('settings.backup.export')}
              </Button>
              <Button
                icon={<UploadIcon size={14} />}
                loading={isReading}
                onClick={() => fileInput.current?.click()}
              >
                {t('settings.backup.import')}
              </Button>
              <input
                ref={fileInput}
                type="file"
                accept=".json,application/json"
                className="hidden"
                data-testid="settings-file"
                onChange={(e) => void readFile(e)}
              />
            </div>
          </Section>
        )}

        {hasDiagnostics && (
          <Section
            title={t('settings.backup.diagnosticsTitle')}
            description={t('settings.backup.diagnosticsDesc')}
          >
            <div>
              <Button
                icon={<FileArchiveIcon size={14} />}
                loading={isDownloading}
                onClick={() => void downloadDiagnostics()}
              >
                {t('settings.backup.download')}
              </Button>
            </div>
          </Section>
        )}
      </div>

      <Modal
        open={preview !== null}
        title={t('settings.backup.previewTitle')}
        width={640}
        onCancel={closePreview}
        footer={[
          <Button key="cancel" onClick={closePreview}>
            {t('settings.backup.cancel')}
          </Button>,
          <Button
            key="import"
            type="primary"
            loading={isImporting}
            disabled={!preview || !canImport(preview, selected)}
            onClick={() => void importSelected()}
          >
            {t('settings.backup.importSelected')}
          </Button>
        ]}
      >
        {preview && (
          <div className="flex flex-col space-y-3">
            <span className="text-xs text-neutral-500">
              {t('settings.backup.previewSource', {
                host: preview.source.hostname || '-',
                version: preview.source.application || '-'
              })}
            </span>
            {defaultSelection(preview).length === 0 && (
              <Alert type="info" showIcon message={t('settings.backup.noChanges')} />
            )}
            <div className="max-h-[50vh] overflow-y-auto pr-1">
              {preview.sections.map(renderSection)}
            </div>
          </div>
        )}
      </Modal>
    </>
  );
};
