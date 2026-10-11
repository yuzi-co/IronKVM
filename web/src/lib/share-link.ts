// Share links: an expiring, revocable URL that opens the screen with no
// control (server/service/auth/share.go).

// The lifetimes the page offers, shortest first. The server accepts a minute
// to thirty days; these are the ones worth a button.
export const SHARE_DURATIONS = [
  { key: 'h1', seconds: 60 * 60 },
  { key: 'h8', seconds: 8 * 60 * 60 },
  { key: 'd1', seconds: 24 * 60 * 60 },
  { key: 'd7', seconds: 7 * 24 * 60 * 60 },
  { key: 'd30', seconds: 30 * 24 * 60 * 60 }
] as const;

export type ShareDurationKey = (typeof SHARE_DURATIONS)[number]['key'];

export const DEFAULT_SHARE_DURATION: ShareDurationKey = 'd1';

export function shareDurationSeconds(key: ShareDurationKey): number {
  return SHARE_DURATIONS.find((duration) => duration.key === key)?.seconds ?? 24 * 60 * 60;
}

// shareLinkUrl is the address to hand out. The token goes after the #, so it is
// never sent to the server in a request line, a log or a Referer header. The
// page opens it by posting the token to /api/auth/share.
export function shareLinkUrl(
  location: { origin: string; pathname: string },
  token: string
): string {
  return `${location.origin}${location.pathname}#/share/${encodeURIComponent(token)}`;
}

// A token as the server makes them: URL-safe base64, 43 characters.
const TOKEN_PATTERN = /^[A-Za-z0-9_-]{16,128}$/;

export function isShareToken(value: string | undefined): value is string {
  return value !== undefined && TOKEN_PATTERN.test(value);
}
