// EDIDs as the server decodes them (server/service/edid), and how the
// settings page writes their parts out. Pure, so it can be tested without a
// browser.

export type Timing = {
  block: number;
  pixelClockKHz: number;
  hActive: number;
  hBlank: number;
  hFrontPorch: number;
  hSync: number;
  vActive: number;
  vBlank: number;
  vFrontPorch: number;
  vSync: number;
  widthMm: number;
  heightMm: number;
  interlaced: boolean;
  hSyncPositive: boolean;
  vSyncPositive: boolean;
  refreshHz: number;
  hFreqKHz: number;
};

export type Mode = { width: number; height: number; refreshHz: number; interlaced?: boolean };

export type Vic = {
  code: number;
  native: boolean;
  width: number;
  height: number;
  refreshHz: number;
  interlaced: boolean;
  aspect: string;
};

export type AudioFormat = {
  format: string;
  channels: number;
  sampleRatesKHz: number[];
  bitDepths?: number[];
  maxBitrateKbps?: number;
};

export type Cta = {
  block: number;
  revision: number;
  underscan: boolean;
  basicAudio: boolean;
  ycbcr444: boolean;
  ycbcr422: boolean;
  nativeDTDs: number;
  vics: Vic[];
  audio: AudioFormat[];
  speakers: string[];
  hdmi: {
    physicalAddress: string;
    maxTMDSMHz: number;
    deepColor30: boolean;
    deepColor36: boolean;
    deepColor48: boolean;
  } | null;
  hdmiForum: { version: number; maxTMDSMHz: number } | null;
  dataBlocks: { tag: number; extTag?: number; name: string; length: number }[];
};

export type EdidInfo = {
  size: number;
  blocks: number;
  valid: boolean;
  problems: string[];
  checksums: { block: number; stored: number; computed: number; valid: boolean }[];
  version: string;
  manufacturer: string;
  productCode: number;
  serialNumber: number;
  serialText?: string;
  name?: string;
  text?: string;
  week: number;
  year: number;
  modelYear: boolean;
  digital: boolean;
  widthCm: number;
  heightCm: number;
  gamma?: number;
  preferred: Timing | null;
  detailedTimings: Timing[];
  establishedTimings: Mode[];
  standardTimings: Mode[];
  rangeLimits: {
    minVHz: number;
    maxVHz: number;
    minHKHz: number;
    maxHKHz: number;
    maxPixelClockMHz: number;
  } | null;
  extensions: { block: number; tag: number; name: string }[];
  cta: Cta | null;
};

export type EdidOrigin = 'generated' | 'sipeed' | 'bridge';

export type EdidPreset = {
  id: string;
  name: string;
  origin: EdidOrigin;
  source: string;
  licence: string;
  capturedOn: string[];
  size: number;
  sha256: string;
  preferred: Timing | null;
};

export type DecodedEdid = { hex: string; decoded: EdidInfo };

export type CurrentEdid = {
  readable: boolean;
  reason?: string;
  builtinPreset?: string;
} & Partial<DecodedEdid>;

// formatRate drops the decimals a rate does not need, keeping at most two.
export function formatRate(hz: number): string {
  return String(Math.round(hz * 100) / 100);
}

export function formatMode(mode: Mode): string {
  return `${mode.width}x${mode.height}${mode.interlaced ? 'i' : ''} @ ${formatRate(mode.refreshHz)} Hz`;
}

// formatTiming names a detailed timing by its mode. An interlaced timing
// counts lines per field, so the frame has twice as many.
export function formatTiming(t: Timing): string {
  return formatMode({
    width: t.hActive,
    height: t.interlaced ? t.vActive * 2 : t.vActive,
    refreshHz: t.refreshHz,
    interlaced: t.interlaced
  });
}

// timingDetail is the rest of a detailed timing: pixel clock, line rate,
// front porch/sync/back porch each way, and the sync polarities.
export function timingDetail(t: Timing): string {
  const hBack = t.hBlank - t.hFrontPorch - t.hSync;
  const vBack = t.vBlank - t.vFrontPorch - t.vSync;
  return [
    `${t.pixelClockKHz / 1000} MHz`,
    `${formatRate(t.hFreqKHz)} kHz`,
    `H ${t.hFrontPorch}/${t.hSync}/${hBack}`,
    `V ${t.vFrontPorch}/${t.vSync}/${vBack}`,
    `${t.hSyncPositive ? '+' : '−'}H ${t.vSyncPositive ? '+' : '−'}V`
  ].join(' · ');
}

export function formatAudio(a: AudioFormat): string {
  const parts = [a.format, `${a.channels} ch`, `${a.sampleRatesKHz.join('/')} kHz`];
  if (a.bitDepths?.length) parts.push(`${a.bitDepths.join('/')} bit`);
  if (a.maxBitrateKbps) parts.push(`${a.maxBitrateKbps} kbit/s`);
  return parts.join(', ');
}

// hexDump lays bytes out as offset and sixteen bytes a line.
export function hexDump(hex: string): string[] {
  const lines: string[] = [];
  for (let i = 0; i < hex.length; i += 32) {
    const bytes = hex.slice(i, i + 32).match(/../g) ?? [];
    lines.push(`${(i / 2).toString(16).padStart(4, '0')}  ${bytes.join(' ')}`);
  }
  return lines;
}

export function presetDownloadPath(id: string): string {
  return `/api/edid/presets/${encodeURIComponent(id)}/bin`;
}
