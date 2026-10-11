import { http } from '@/lib/http.ts';

// The server contract: version, API version and features. Answered without
// authentication, so the login page can read it too.
export function getServerVersion() {
  return http.get('/api/version');
}
