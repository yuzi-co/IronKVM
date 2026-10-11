// The host log page's rules, apart from React so that node --test can run them.
import type { HostLogEntry, HostLogFrame } from '../api/hostlog.ts';

// syslog severities, most urgent first.
const SEVERITIES = ['emerg', 'alert', 'crit', 'err', 'warning', 'notice', 'info', 'debug'] as const;

export function severityName(severity: number): string {
  return SEVERITIES[severity] ?? '';
}

// A Tailwind text colour for a line: red for errors and worse, amber for
// warnings, grey for debug, none for the rest.
export function severityColor(severity: number): string {
  if (severity < 0) return '';
  if (severity <= 3) return 'text-red-400';
  if (severity === 4) return 'text-amber-400';
  if (severity === 7) return 'text-neutral-500';
  return '';
}

// How many lines the page keeps. The server's ring holds more; the page shows
// the newest ones so a long session does not grow without end.
export const MAX_LINES = 2000;

// mergeEntries appends the entries newer than the last one held and keeps the
// newest max. The socket can repeat a line that the backlog already carried.
export function mergeEntries(
  current: HostLogEntry[],
  incoming: HostLogEntry[],
  max: number = MAX_LINES
): HostLogEntry[] {
  let last = current.length > 0 ? current[current.length - 1].seq : 0;
  const fresh: HostLogEntry[] = [];
  for (const entry of incoming) {
    if (entry.seq > last) {
      fresh.push(entry);
      last = entry.seq;
    }
  }
  if (fresh.length === 0) return current;

  const all = current.concat(fresh);
  return all.length > max ? all.slice(all.length - max) : all;
}

// parseFrame reads one message of the live tail, or null when it is not one.
export function parseFrame(data: unknown): HostLogFrame | null {
  if (typeof data !== 'string') return null;
  try {
    const frame = JSON.parse(data) as HostLogFrame;
    if ((frame.type !== 'backlog' && frame.type !== 'entries') || !Array.isArray(frame.entries)) {
      return null;
    }
    return frame;
  } catch {
    return null;
  }
}

// parseList splits a text box into its non-empty trimmed items. Items are
// separated by new lines, and for addresses by commas and spaces too.
export function parseList(text: string, extraSeparators = false): string[] {
  const pattern = extraSeparators ? /[\n,\s]+/ : /\n/;
  const out: string[] = [];
  for (const item of text.split(pattern)) {
    const trimmed = item.trim();
    if (trimmed && !out.includes(trimmed)) out.push(trimmed);
  }
  return out;
}

// plainLine is a line as the page copies it.
export function plainLine(entry: HostLogEntry): string {
  const who = [entry.host, entry.tag ? `${entry.tag}:` : ''].filter(Boolean).join(' ');
  return [entry.time, entry.source, who, entry.msg].filter(Boolean).join(' ');
}

// retryDelay is how long the live tail waits before it opens the socket again:
// a second at first, doubling to half a minute.
export function retryDelay(attempt: number): number {
  return Math.min(30_000, 1000 * 2 ** Math.max(0, attempt));
}
