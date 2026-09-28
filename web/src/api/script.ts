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
// than the usual minute before the request gives up.
export const FOREGROUND_TIMEOUT_MINUTES = 10;

export function runScript(name: string, type: string) {
  return http.request({
    method: 'post',
    url: '/api/vm/script/run',
    data: { name, type },
    timeout: type === 'foreground' ? FOREGROUND_TIMEOUT_MINUTES * 60 * 1000 : undefined
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
