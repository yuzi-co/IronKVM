import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Checkbox, Divider, InputNumber, message, Popconfirm, Switch } from 'antd';
import { DownloadIcon, Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/dashcam.ts';
import type { DashcamClip, DashcamSettings, DashcamState } from '@/api/dashcam.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { StaleNote, StatusTag } from '../components/status-tag.tsx';
import { formatTime } from '../streaming/format.ts';
import { SettingRow } from '../streaming/rows.tsx';
import { clipSize, LIMITS, settingsProblem, triggerKey } from './format.ts';

// The ring and the clips change on their own: a clip is written after its
// trigger, and the ring fills once the dashcam is on.
const pollMs = 5 * 1000;

// The dashcam (ironkvm-dist#119): its switch, how much it keeps, what saves a
// clip, and the clips on the board with download and delete.
export const Dashcam = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<DashcamSettings | null>(null);
  const [draft, setDraft] = useState<DashcamSettings | null>(null);
  const [state, setState] = useState<DashcamState | null>(null);
  const [clips, setClips] = useState<DashcamClip[] | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isStale, setIsStale] = useState(false);

  const loadSettings = useCallback(() => {
    api
      .getDashcamSettings()
      .then((rsp) => {
        if (rsp.code !== 0) return;
        setSettings(rsp.data);
        setDraft(rsp.data);
      })
      .catch(() => setIsStale(true));
  }, []);

  const loadState = useCallback(() => {
    Promise.all([api.getDashcamState(), api.getDashcamClips()])
      .then(([stateRsp, clipsRsp]) => {
        if (stateRsp.code === 0) setState(stateRsp.data);
        if (clipsRsp.code === 0) setClips(clipsRsp.data ?? []);
        setIsStale(stateRsp.code !== 0 || clipsRsp.code !== 0);
      })
      .catch(() => setIsStale(true));
  }, []);

  useEffect(() => {
    loadSettings();
    loadState();
    return pollWhileVisible(loadState, pollMs);
  }, [loadSettings, loadState]);

  function save(next: DashcamSettings) {
    setIsSaving(true);
    api
      .setDashcamSettings(next)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.dashcam.failed')));
          return;
        }
        setSettings(rsp.data);
        setDraft(rsp.data);
        message.success(t('settings.dashcam.saved'));
        loadState();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.dashcam.failed'))))
      .finally(() => setIsSaving(false));
  }

  function update<K extends keyof DashcamSettings>(key: K, value: DashcamSettings[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  }

  function setTrigger(key: keyof DashcamSettings['triggers'], value: boolean) {
    setDraft((current) =>
      current ? { ...current, triggers: { ...current.triggers, [key]: value } } : current
    );
  }

  function remove(name: string) {
    api
      .deleteDashcamClip(name)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.dashcam.deleteFailed')));
          return;
        }
        message.success(t('settings.dashcam.deleted'));
        loadState();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.dashcam.deleteFailed'))));
  }

  const isDirty = !!draft && !!settings && JSON.stringify(draft) !== JSON.stringify(settings);
  const problem = draft ? settingsProblem(draft) : null;

  function triggerLabel(trigger: string) {
    const key = triggerKey(trigger);
    return key ? t(`settings.dashcam.triggerNames.${key}`) : trigger;
  }

  return (
    <>
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.dashcam.title')}</span>
        {state && <StatusTag running={state.running} stale={isStale} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.dashcam.enabled')}</span>
            <span className="text-xs text-neutral-500">{t('settings.dashcam.enabledDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={(value) => settings && save({ ...settings, enabled: value })}
          />
        </div>

        <Alert type="info" showIcon message={t('settings.dashcam.cost')} />

        {state?.enabled && !state.running && (
          <Alert type="warning" showIcon message={t('settings.dashcam.noVideo')} />
        )}
        {state?.lastError && (
          <Alert
            type="error"
            showIcon
            message={t('settings.dashcam.lastError', { error: state.lastError })}
          />
        )}

        {draft && (
          <div className="flex flex-col space-y-4">
            <SettingRow
              label={t('settings.dashcam.seconds')}
              desc={t('settings.dashcam.secondsDesc')}
            >
              <InputNumber
                style={{ width: 150 }}
                min={LIMITS.minSeconds}
                max={LIMITS.maxSeconds}
                precision={0}
                addonAfter="s"
                value={draft.seconds}
                onChange={(value) => update('seconds', value ?? 60)}
              />
            </SettingRow>

            <SettingRow
              label={t('settings.dashcam.postSeconds')}
              desc={t('settings.dashcam.postSecondsDesc')}
            >
              <InputNumber
                style={{ width: 150 }}
                min={0}
                max={LIMITS.maxPostSeconds}
                precision={0}
                addonAfter="s"
                value={draft.postSeconds}
                onChange={(value) => update('postSeconds', value ?? 0)}
              />
            </SettingRow>

            <SettingRow
              label={t('settings.dashcam.bitrate')}
              desc={t('settings.dashcam.bitrateDesc')}
            >
              <div className="flex flex-col items-end space-y-2">
                <Checkbox
                  checked={draft.maxKbps === 0}
                  onChange={(e) => update('maxKbps', e.target.checked ? 0 : 2000)}
                >
                  {t('settings.dashcam.bitrateStream')}
                </Checkbox>
                {draft.maxKbps !== 0 && (
                  <InputNumber
                    style={{ width: 150 }}
                    min={LIMITS.minKbps}
                    max={LIMITS.maxKbps}
                    precision={0}
                    addonAfter="kbit/s"
                    value={draft.maxKbps}
                    onChange={(value) => update('maxKbps', value ?? LIMITS.minKbps)}
                  />
                )}
              </div>
            </SettingRow>

            <SettingRow
              label={t('settings.dashcam.storage')}
              desc={t('settings.dashcam.storageDesc')}
            >
              <InputNumber
                style={{ width: 150 }}
                min={LIMITS.minStorageMB}
                max={LIMITS.maxStorageMB}
                precision={0}
                addonAfter="MB"
                value={draft.storageMB}
                onChange={(value) => update('storageMB', value ?? 200)}
              />
            </SettingRow>

            <div className="flex flex-col space-y-2">
              <span className="text-sm">{t('settings.dashcam.triggers')}</span>
              <Checkbox
                checked={draft.triggers.atx}
                onChange={(e) => setTrigger('atx', e.target.checked)}
              >
                {t('settings.dashcam.triggerAtx')}
              </Checkbox>
              <Checkbox
                checked={draft.triggers.signal}
                onChange={(e) => setTrigger('signal', e.target.checked)}
              >
                {t('settings.dashcam.triggerSignal')}
              </Checkbox>
              <Checkbox
                checked={draft.triggers.failsafe}
                onChange={(e) => setTrigger('failsafe', e.target.checked)}
              >
                {t('settings.dashcam.triggerFailsafe')}
              </Checkbox>
              <span className="text-xs text-neutral-500">{t('settings.dashcam.triggerNote')}</span>
            </div>

            <div className="flex justify-end">
              <Button
                type="primary"
                disabled={!isDirty || problem !== null}
                loading={isSaving}
                onClick={() => draft && save(draft)}
              >
                {t('settings.dashcam.save')}
              </Button>
            </div>
          </div>
        )}

        {state?.running && (
          <SettingRow label={t('settings.dashcam.buffered')}>
            <span className="text-sm text-neutral-300">
              {t('settings.dashcam.bufferedValue', {
                seconds: Math.round(state.buffered),
                size: clipSize(state.bufferedBytes),
                budget: clipSize(state.budgetBytes)
              })}
            </span>
          </SettingRow>
        )}

        {clips && (
          <div className={`flex flex-col space-y-2 ${isStale ? 'opacity-60' : ''}`}>
            <span className="text-sm font-medium">{t('settings.dashcam.clips')}</span>
            {isStale && <StaleNote />}
            {clips.length === 0 ? (
              <span className="text-xs text-neutral-500">{t('settings.dashcam.noClips')}</span>
            ) : (
              <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                {clips.map((clip, index) => (
                  <div
                    key={clip.name}
                    className={`flex items-center justify-between px-4 py-2 ${index > 0 ? 'border-t border-neutral-800' : ''}`}
                  >
                    <div className="flex min-w-0 flex-col">
                      <span className="truncate text-sm text-neutral-300">
                        {clip.triggers.map(triggerLabel).join(', ')}
                      </span>
                      <span className="truncate text-xs text-neutral-500">
                        {formatTime(clip.time)} · {clipSize(clip.size)} · {clip.name}
                      </span>
                    </div>
                    <div className="flex shrink-0 items-center space-x-1 pl-2">
                      <Button
                        type="text"
                        size="small"
                        aria-label={t('settings.dashcam.download')}
                        icon={<DownloadIcon size={16} />}
                        href={api.dashcamClipUrl(clip.name)}
                        download={clip.name}
                      />
                      <Popconfirm
                        title={t('settings.dashcam.deleteConfirm', { name: clip.name })}
                        okText={t('settings.dashcam.delete')}
                        okButtonProps={{ danger: true }}
                        onConfirm={() => remove(clip.name)}
                      >
                        <Button
                          type="text"
                          size="small"
                          danger
                          aria-label={t('settings.dashcam.delete')}
                          icon={<Trash2Icon size={16} />}
                        />
                      </Popconfirm>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </>
  );
};
