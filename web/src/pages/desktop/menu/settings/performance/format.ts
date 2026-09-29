// formatBytes keeps one decimal place, which is enough to tell 3.2 MB from
// 3.9 MB on a board where the whole swap device is 96 MB.
export function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 MB';

  const units = ['B', 'KB', 'MB', 'GB'];
  let value = bytes;
  let unit = 0;

  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }

  return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}
