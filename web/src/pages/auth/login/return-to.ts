// The login page is reached with router state saying where to go afterwards
// and, after a password change, a notice to show. Both arrive as untyped
// history state, so they are read defensively.

export type LoginState = {
  from?: { pathname?: unknown; search?: unknown; hash?: unknown };
  notice?: 'passwordChanged';
};

// returnTo is the path to open after signing in: the page the operator was
// sent away from, or the desktop. The auth pages are never a target, since
// returning to the password form or to login itself would make no sense.
export function returnTo(state: unknown): string {
  const from = (state as LoginState | null)?.from;
  const pathname = typeof from?.pathname === 'string' ? from.pathname : '';
  if (!pathname.startsWith('/') || pathname.startsWith('//') || pathname.startsWith('/auth/')) {
    return '/';
  }
  const search = typeof from?.search === 'string' ? from.search : '';
  const hash = typeof from?.hash === 'string' ? from.hash : '';
  return `${pathname}${search}${hash}`;
}

export function loginNotice(state: unknown): LoginState['notice'] {
  const notice = (state as LoginState | null)?.notice;
  return notice === 'passwordChanged' ? notice : undefined;
}
