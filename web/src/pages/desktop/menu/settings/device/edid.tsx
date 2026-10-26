import { useEffect, useRef, useState } from 'react';
import type { ChangeEvent, ReactNode } from 'react';
import { Alert, Button, Modal, Tag } from 'antd';
import { DownloadIcon, EyeIcon, UploadIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/edid.ts';
import { formatAudio, formatMode, formatTiming, hexDump, timingDetail } from '@/lib/edid.ts';
import type { CurrentEdid, DecodedEdid, EdidInfo, EdidPreset } from '@/lib/edid.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { Box, Section } from '../components/section.tsx';

const originColors: Record<string, string> = {
  generated: 'blue',
  sipeed: 'default',
  bridge: 'purple'
};

type Shown = { title: string; edid: DecodedEdid; preset?: EdidPreset };

// Edid shows EDIDs and decodes them: the board's own (which cannot be read
// back, see server/service/edid), the preset library and a file the user
// picks. Applying one is parked (ironkvm-dist#95), and the section says so.
export const Edid = () => {
  const hasEdid = useServerFeature('edid-view') === true;
  return hasEdid ? <EdidSection /> : null;
};

const EdidSection = () => {
  const { t } = useTranslation();

  const fileInput = useRef<HTMLInputElement>(null);
  const [presets, setPresets] = useState<EdidPreset[]>([]);
  const [current, setCurrent] = useState<CurrentEdid | null>(null);
  const [loadFailed, setLoadFailed] = useState(false);
  const [opening, setOpening] = useState<string | null>(null);
  const [shown, setShown] = useState<Shown | null>(null);
  const [failure, setFailure] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.listEdidPresets(), api.getCurrentEdid()])
      .then(([list, cur]) => {
        if (list.code !== 0 || cur.code !== 0) {
          setLoadFailed(true);
          return;
        }
        setPresets(list.data?.presets ?? []);
        setCurrent(cur.data);
      })
      .catch(() => setLoadFailed(true));
  }, []);

  async function openPreset(preset: EdidPreset) {
    if (opening) return;
    setOpening(preset.id);
    setFailure(null);
    try {
      const rsp = await api.getEdidPreset(preset.id);
      if (rsp.code !== 0) {
        setFailure(describeFailure(rsp, t('settings.device.edid.loadFailed')));
        return;
      }
      setShown({ title: preset.name, edid: rsp.data, preset });
    } catch (err) {
      setFailure(describeFailure(err, t('settings.device.edid.loadFailed')));
    } finally {
      setOpening(null);
    }
  }

  async function inspectFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    // The same file may be chosen again after a cancel.
    event.target.value = '';
    if (!file) return;

    setOpening('file');
    setFailure(null);
    try {
      const rsp = await api.decodeEdid(file);
      if (rsp.code !== 0) {
        setFailure(`${t('settings.device.edid.decodeFailed')}: ${rsp.msg}`);
        return;
      }
      setShown({ title: file.name, edid: rsp.data });
    } catch (err) {
      setFailure(describeFailure(err, t('settings.device.edid.decodeFailed')));
    } finally {
      setOpening(null);
    }
  }

  const builtin = presets.find((p) => p.id === current?.builtinPreset);

  return (
    <Section
      title={t('settings.device.edid.title')}
      description={t('settings.device.edid.description')}
    >
      <Alert type="info" showIcon message={t('settings.device.edid.parked')} />

      {loadFailed ? (
        <div className="text-xs text-neutral-500">{t('settings.device.edid.loadFailed')}</div>
      ) : (
        <>
          <Box>
            <span className="font-medium">{t('settings.device.edid.current')}</span>
            {current?.readable && current.hex && current.decoded ? (
              <Button
                size="small"
                className="self-start"
                onClick={() =>
                  setShown({
                    title: t('settings.device.edid.current'),
                    edid: { hex: current.hex!, decoded: current.decoded! }
                  })
                }
              >
                {t('settings.device.edid.view')}
              </Button>
            ) : (
              <>
                <span className="text-xs text-neutral-400">
                  {t('settings.device.edid.notReadable')}
                </span>
                {builtin && (
                  <Button
                    size="small"
                    className="self-start"
                    icon={<EyeIcon size={14} />}
                    loading={opening === builtin.id}
                    onClick={() => openPreset(builtin)}
                  >
                    {t('settings.device.edid.viewBuiltin')}
                  </Button>
                )}
              </>
            )}
          </Box>

          <div className="flex flex-col space-y-1">
            <span className="text-sm">{t('settings.device.edid.presets')}</span>
            {presets.map((preset) => (
              <div
                key={preset.id}
                className="flex flex-wrap items-center justify-between gap-2 border-b border-neutral-700/50 py-2 last:border-b-0"
              >
                <div className="flex min-w-0 flex-col">
                  <span className="text-sm">{preset.name}</span>
                  <span className="text-xs text-neutral-500">
                    <Tag color={originColors[preset.origin]} className="mr-2">
                      {t(`settings.device.edid.origin.${preset.origin}`)}
                    </Tag>
                    {preset.capturedOn.length > 0 &&
                      t('settings.device.edid.capturedOn', {
                        slots: preset.capturedOn.map((s) => s.toUpperCase()).join(', ')
                      })}
                  </span>
                </div>
                <div className="flex gap-2">
                  <Button
                    size="small"
                    icon={<EyeIcon size={14} />}
                    loading={opening === preset.id}
                    onClick={() => openPreset(preset)}
                  >
                    {t('settings.device.edid.view')}
                  </Button>
                  <Button
                    size="small"
                    icon={<DownloadIcon size={14} />}
                    href={api.getEdidPresetUrl(preset.id)}
                    download={`edid-${preset.id}.bin`}
                  >
                    {t('settings.device.edid.download')}
                  </Button>
                </div>
              </div>
            ))}
          </div>

          <div className="flex flex-col space-y-2">
            <span className="text-xs text-neutral-500">{t('settings.device.edid.uploadDesc')}</span>
            <input
              ref={fileInput}
              type="file"
              accept=".bin,.edid,.hex,.txt,.dat"
              className="hidden"
              onChange={inspectFile}
            />
            <Button
              className="self-start"
              icon={<UploadIcon size={14} />}
              loading={opening === 'file'}
              onClick={() => fileInput.current?.click()}
            >
              {t('settings.device.edid.upload')}
            </Button>
          </div>
        </>
      )}

      {failure && <Alert type="error" showIcon message={failure} />}

      <Modal
        open={shown !== null}
        title={shown?.title}
        width={760}
        footer={null}
        onCancel={() => setShown(null)}
        destroyOnHidden
      >
        {shown && <EdidView edid={shown.edid} preset={shown.preset} />}
      </Modal>
    </Section>
  );
};

const Row = ({ label, children }: { label: string; children: ReactNode }) => (
  <div className="grid grid-cols-[10rem_1fr] gap-2 py-0.5 text-sm max-sm:grid-cols-1 max-sm:gap-0">
    <span className="text-neutral-500">{label}</span>
    <span className="min-w-0 break-words">{children}</span>
  </div>
);

const Chips = ({ items }: { items: string[] }) => (
  <div className="flex flex-wrap gap-1">
    {items.map((item, i) => (
      <span key={`${item}-${i}`} className="rounded bg-neutral-800 px-1.5 py-0.5 font-mono text-xs">
        {item}
      </span>
    ))}
  </div>
);

// EdidView is the decoded EDID as a table, then its bytes.
const EdidView = ({ edid, preset }: { edid: DecodedEdid; preset?: EdidPreset }) => {
  const { t } = useTranslation();
  const info: EdidInfo = edid.decoded;
  const cta = info.cta;
  const otherModes = [...info.establishedTimings, ...info.standardTimings].map(formatMode);

  return (
    <div className="flex flex-col space-y-3">
      {info.valid ? (
        <Tag color="green" className="self-start">
          {t('settings.device.edid.valid')}
        </Tag>
      ) : (
        <Alert
          type="error"
          showIcon
          message={t('settings.device.edid.problems')}
          description={
            <ul className="list-disc pl-4 text-xs">
              {info.problems.map((p) => (
                <li key={p}>{p}</li>
              ))}
            </ul>
          }
        />
      )}

      <div className="flex flex-col">
        <Row label={t('settings.device.edid.manufacturer')}>
          <span className="font-mono">{info.manufacturer}</span>
        </Row>
        <Row label={t('settings.device.edid.product')}>
          <span className="font-mono">0x{info.productCode.toString(16).padStart(4, '0')}</span>
          {info.serialNumber > 0 && ` · ${t('settings.device.edid.serial')} ${info.serialNumber}`}
        </Row>
        {(info.name || info.serialText) && (
          <Row label={t('settings.device.edid.name')}>
            {[info.name, info.serialText].filter(Boolean).join(' · ')}
          </Row>
        )}
        <Row label={t('settings.device.edid.made')}>
          {info.modelYear
            ? t('settings.device.edid.modelYear', { year: info.year })
            : info.week > 0
              ? t('settings.device.edid.madeWeek', { year: info.year, week: info.week })
              : info.year}
        </Row>
        <Row label={t('settings.device.edid.version')}>
          {info.version} · {info.size} B · {info.extensions.map((e) => e.name).join(', ') || '-'}
        </Row>
        <Row label={t('settings.device.edid.preferred')}>
          {info.preferred ? formatTiming(info.preferred) : '-'}
        </Row>
        {info.rangeLimits && (
          <Row label={t('settings.device.edid.range')}>
            <span className="font-mono text-xs">
              {info.rangeLimits.minVHz}-{info.rangeLimits.maxVHz} Hz · {info.rangeLimits.minHKHz}-
              {info.rangeLimits.maxHKHz} kHz · {info.rangeLimits.maxPixelClockMHz} MHz
            </span>
          </Row>
        )}
        <Row label={t('settings.device.edid.timings')}>
          <div className="flex flex-col space-y-1">
            {info.detailedTimings.map((timing, i) => (
              <div key={i} className="flex flex-col">
                <span>{formatTiming(timing)}</span>
                <span className="font-mono text-xs text-neutral-500">{timingDetail(timing)}</span>
              </div>
            ))}
          </div>
        </Row>
        {otherModes.length > 0 && (
          <Row label={t('settings.device.edid.modes')}>
            <Chips items={otherModes} />
          </Row>
        )}
        {cta && cta.vics.length > 0 && (
          <Row label={t('settings.device.edid.vics')}>
            <Chips
              items={cta.vics.map((v) =>
                v.width ? `${v.code}: ${formatMode(v)}${v.native ? ' *' : ''}` : `${v.code}`
              )}
            />
          </Row>
        )}
        {cta && (
          <Row label={t('settings.device.edid.audio')}>
            {cta.audio.length > 0 ? (
              <div className="flex flex-col">
                {cta.audio.map((a, i) => (
                  <span key={i}>{formatAudio(a)}</span>
                ))}
                {cta.speakers.length > 0 && (
                  <span className="text-xs text-neutral-500">
                    {t('settings.device.edid.speakers')}: {cta.speakers.join(', ')}
                  </span>
                )}
              </div>
            ) : (
              t('settings.device.edid.noAudio')
            )}
          </Row>
        )}
        {cta?.hdmi && (
          <Row label="HDMI">
            <span className="font-mono text-xs">
              {cta.hdmi.physicalAddress}
              {cta.hdmi.maxTMDSMHz > 0 && ` · TMDS ${cta.hdmi.maxTMDSMHz} MHz`}
              {cta.hdmiForum && ` · HDMI Forum, ${cta.hdmiForum.maxTMDSMHz} MHz`}
            </span>
          </Row>
        )}
        {cta && cta.dataBlocks.length > 0 && (
          <Row label={t('settings.device.edid.dataBlocks')}>
            <Chips items={cta.dataBlocks.map((b) => b.name)} />
          </Row>
        )}
        <Row label={t('settings.device.edid.checksums')}>
          <span className="font-mono text-xs">
            {info.checksums
              .map(
                (c) =>
                  `${c.block}: 0x${c.stored.toString(16).padStart(2, '0')} ${c.valid ? '✓' : '✗'}`
              )
              .join(' · ')}
          </span>
        </Row>
        {preset && (
          <>
            <Row label={t('settings.device.edid.source')}>
              <span className="text-xs">{preset.source}</span>
            </Row>
            <Row label={t('settings.device.edid.licence')}>
              <span className="text-xs">{preset.licence}</span>
            </Row>
          </>
        )}
      </div>

      <details className="text-xs text-neutral-400">
        <summary className="cursor-pointer select-none hover:text-neutral-200">
          {t('settings.device.edid.hex')}
        </summary>
        <pre className="mt-2 overflow-x-auto font-mono text-xs text-neutral-300">
          {hexDump(edid.hex).join('\n')}
        </pre>
      </details>
    </div>
  );
};
