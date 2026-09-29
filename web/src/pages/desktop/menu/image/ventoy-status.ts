// What the Media dialog and the Virtual media settings page say about Ventoy,
// worked out from the status the server reports. Pure, so it runs under node's
// test runner.

import type { VentoyStatus } from '@/api/ventoy.ts';

export function basename(path: string) {
  return path.replace(/^.*[\\/]/, '');
}

export function formatSize(bytes: number) {
  const mib = bytes / (1024 * 1024);
  return mib >= 1024 ? `${(mib / 1024).toFixed(1)} GiB` : `${mib.toFixed(1)} MiB`;
}

// ventoyUsable reports whether the Media dialog has Ventoy to offer: the
// firmware supports it and it is installed. Setting it up is done in Settings.
export function ventoyUsable(status: VentoyStatus | null): status is VentoyStatus {
  return !!status && status.kernel && status.installed;
}

// ventoyStatusText returns the translation key, and its values, for the one
// line that sums up the state of Ventoy.
export function ventoyStatusText(status: VentoyStatus): {
  key: string;
  values?: Record<string, string | number>;
} {
  const selected = status.images ?? [];
  if (!status.kernel) return { key: 'image.ventoy.statusNoKernel' };
  if (!status.installed) return { key: 'image.ventoy.statusNotInstalled' };
  if (status.inDrive) {
    return { key: 'image.ventoy.statusInDrive', values: { size: formatSize(status.size) } };
  }
  if (selected.length > 0) {
    return { key: 'image.ventoy.statusSelected', values: { count: selected.length } };
  }
  return { key: 'image.ventoy.statusReady' };
}

// ventoyPresent is the selected images that still exist: the disk can be
// built only when there is at least one.
export function ventoyPresent(status: VentoyStatus): string[] {
  const missing = status.missing ?? [];
  return (status.images ?? []).filter((image) => !missing.includes(image));
}
