// The settings import flow without the page: reading a file, choosing the
// sections and showing the values. Kept apart so node --test can run it.

import type { Preview, SectionPreview } from '@/api/backup.ts';

// The file the server writes and the format version this UI knows. The server
// checks both again; the check here only gives a clearer message sooner.
export const SETTINGS_FORMAT = 'ironkvm-settings';
export const SETTINGS_VERSION = 1;

// The server's limit on a settings file, service/settingsio MaxDocumentBytes.
export const MAX_FILE_BYTES = 8 << 20;

export type ParseFailure = 'tooLarge' | 'notJson' | 'notSettings' | 'newer';

export type ParsedFile =
  { ok: true; document: Record<string, unknown> } | { ok: false; reason: ParseFailure };

// parseSettingsFile reads the text of a chosen file.
export function parseSettingsFile(text: string): ParsedFile {
  if (text.length > MAX_FILE_BYTES) return { ok: false, reason: 'tooLarge' };

  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    return { ok: false, reason: 'notJson' };
  }
  if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
    return { ok: false, reason: 'notSettings' };
  }
  const document = parsed as Record<string, unknown>;
  if (document.format !== SETTINGS_FORMAT) return { ok: false, reason: 'notSettings' };
  if (typeof document.version === 'number' && document.version > SETTINGS_VERSION) {
    return { ok: false, reason: 'newer' };
  }
  return { ok: true, document };
}

// isSelectable says whether a section can be chosen: only one that changes
// something. An invalid or unknown section is shown and cannot be chosen.
export function isSelectable(section: SectionPreview): boolean {
  return section.status === 'change';
}

// defaultSelection chooses every section that changes something.
export function defaultSelection(preview: Preview): string[] {
  return preview.sections.filter(isSelectable).map((s) => s.section);
}

// canImport says whether the chosen sections make an import.
export function canImport(preview: Preview, selected: string[]): boolean {
  return preview.sections.some((s) => isSelectable(s) && selected.includes(s.section));
}

export function toggle(selected: string[], section: string, on: boolean): string[] {
  const rest = selected.filter((s) => s !== section);
  return on ? [...rest, section] : rest;
}

export type ListSummary = { items: number };

function isListSummary(value: unknown): value is ListSummary {
  return (
    value !== null &&
    typeof value === 'object' &&
    !Array.isArray(value) &&
    Object.keys(value).length === 1 &&
    typeof (value as ListSummary).items === 'number'
  );
}

const MAX_VALUE_CHARS = 120;

// formatValue shows one value of a change. A long list comes from the server
// as its length, which the page words in the user's language.
export function formatValue(value: unknown): string | ListSummary {
  if (value === null || value === undefined) return '-';
  if (isListSummary(value)) return value;
  const text = typeof value === 'string' && value !== '' ? value : JSON.stringify(value);
  return text.length > MAX_VALUE_CHARS ? `${text.slice(0, MAX_VALUE_CHARS)}…` : text;
}

const safeName = /^[A-Za-z0-9._-]+$/;

// downloadName takes the file name from a Content-Disposition header, or the
// fallback when there is none or it is not a plain name.
export function downloadName(disposition: string | null, fallback: string): string {
  const match = disposition?.match(/filename="?([^";]+)"?/);
  const name = match?.[1];
  return name && safeName.test(name) && !name.startsWith('.') ? name : fallback;
}
