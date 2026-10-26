// The pure half of the OIDC login: what the login page reads from the address
// the server sent it back to, and how Settings shows the scopes and the
// redirect URL. Runs under node --test, so it imports nothing.

// The reasons the server's callback refuses a login with, in ?oidc=
// (server/service/auth/oidc.go). Each has a message under auth.oidc.errors.
export const OIDC_REASONS = [
  'disabled',
  'locked',
  'provider',
  'denied',
  'state',
  'token',
  'account',
  'role',
  'server'
] as const;

export type OidcReason = (typeof OIDC_REASONS)[number];

// oidcReason reads the refusal from a query string. Anything that is not one
// of the known reasons is none: the page never shows text from the address.
export function oidcReason(search: string): OidcReason | null {
  const value = new URLSearchParams(search).get('oidc');
  return (OIDC_REASONS as readonly string[]).includes(value ?? '') ? (value as OidcReason) : null;
}

// scopesFromText reads the scopes field. openid is left out: the server always
// asks for it first.
export function scopesFromText(text: string): string[] {
  const scopes: string[] = [];
  for (const scope of text.split(/[\s,]+/)) {
    if (scope && scope !== 'openid' && !scopes.includes(scope)) scopes.push(scope);
  }
  return scopes;
}

export function scopesToText(scopes: string[] | undefined): string {
  return (scopes ?? []).filter((s) => s !== 'openid').join(' ');
}

// httpRedirectWarning reports whether a provider is likely to refuse the
// redirect URL: many accept plain http only for localhost.
export function httpRedirectWarning(redirectUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(redirectUrl);
  } catch {
    return false;
  }
  if (url.protocol !== 'http:') return false;
  return !['localhost', '127.0.0.1', '[::1]'].includes(url.hostname);
}
