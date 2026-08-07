import { http, HttpResponse } from 'msw';
import { setupWorker } from 'msw/browser';

let isLoggedIn = false;

// Flip this to exercise each ION verdict against the desktop gate and badge:
// 'ok' | 'warn' | 'critical' | 'unavailable'.
const ION_VERDICT = 'critical'

export const handlers = [
  http.post('/api/auth/login', () => {
    isLoggedIn = true;
    return HttpResponse.json({
      code: 0,
      data: {}
    });
  }),
  http.get('/api/auth/account', () => {
    if (!isLoggedIn) {
      return HttpResponse.json('unauthorized', { status: 401 });
    }
    return HttpResponse.json({
      code: 0,
      data: { username: 'admin', role: 'admin' }
    });
  }),
  http.post('/api/auth/logout', () => {
    isLoggedIn = false;
    return HttpResponse.json({ code: 0 });
  }),
  http.get('/api/vm/ion', () => {
    return HttpResponse.json({
      code: 0,
      data: {
        total: 78643200,
        used: 70778880,
        free: 7864320,
        usageRate: 90,
        generations: 3,
        reserve: 8388608,
        measured: true,
        verdict: ION_VERDICT
      }
    });
  })
];
export const worker = setupWorker(...handlers);
