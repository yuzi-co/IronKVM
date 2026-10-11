// The roles an account can have, as the server reports them.
//
// admin controls the board and its settings. user controls the host but not
// the board. viewer sees the video and nothing else: the server refuses it
// every other route, and the page offers it nothing to press.
export type UserRole = 'admin' | 'user' | 'viewer';

// parseRole reads the role a server sent. A value it does not know counts as
// user, the middle role: an older server never sends viewer, and a role this
// page cannot name must not unlock more than the page was written for.
export function parseRole(value: unknown): UserRole {
  if (value === 'admin') return 'admin';
  if (value === 'viewer') return 'viewer';
  return 'user';
}
