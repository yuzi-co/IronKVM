import { message } from 'antd';

import i18n from '@/i18n/index.ts';

import { failureText } from './feedback-text.ts';

// The one way a settings control reports the outcome of a request: a short
// toast on success when asked for, and on failure the server's reason, a
// translated timeout or network message, or the fallback the caller names.

type Reply = { code: number; msg?: string };

type ResultOptions = {
  success?: string;
  fallback?: string;
};

export function describeFailure(cause: unknown, fallback?: string): string {
  return failureText(cause, fallback ?? i18n.t('feedback.failed'), (key) => i18n.t(key));
}

export function showFailure(cause: unknown, fallback?: string) {
  message.error(describeFailure(cause, fallback));
}

// showResult reports an API reply and returns whether it succeeded.
export function showResult(rsp: Reply, options: ResultOptions = {}): boolean {
  if (rsp.code === 0) {
    if (options.success) message.success(options.success);
    return true;
  }

  showFailure(rsp, options.fallback);
  return false;
}
