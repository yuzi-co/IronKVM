// formatTime shows a time the server sent in the browser's locale, or the
// value as it came when it is not a time.
export function formatTime(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}
