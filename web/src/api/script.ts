import { http } from '@/lib/http.ts';

export function uploadScript(formData: FormData) {
  return http.request({
    url: '/api/vm/script/upload',
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data: formData
  });
}

// A foreground run answers only when the script ends, so it gets far longer
// than the usual minute before the request gives up. The board kills a run at
// the same limit and answers with what it printed; the grace lets that answer
// arrive before the browser stops listening.
export const FOREGROUND_TIMEOUT_MINUTES = 10;
const FOREGROUND_GRACE_MS = 30 * 1000;

export function runScript(name: string, type: string) {
  return http.request({
    method: 'post',
    url: '/api/vm/script/run',
    data: { name, type },
    timeout:
      type === 'foreground'
        ? FOREGROUND_TIMEOUT_MINUTES * 60 * 1000 + FOREGROUND_GRACE_MS
        : undefined
  });
}

export function getScripts() {
  return http.get('/api/vm/script');
}

export function deleteScript(name: string) {
  const data = {
    name
  };
  return http.delete('/api/vm/script', data);
}
