// failureText turns whatever a failed request left behind into one line for the
// operator. It is kept apart from the toast code so it can be tested without a
// browser.
//
// The cause is either an API reply whose code is not 0, or an exception thrown
// by the request. A reason the server gave is shown as is, since it is usually
// the only thing that says what to change. A request that never got an answer
// is told apart from one the server refused, because the remedy differs: the
// first is the network or a busy board, the second is the input.

export type Translate = (key: string) => string;

type Reply = { code?: unknown; msg?: unknown };

type RequestError = {
  code?: unknown;
  message?: unknown;
  request?: unknown;
  response?: { data?: Reply } | undefined;
};

function reason(reply: Reply | undefined): string {
  const msg = reply?.msg;
  return typeof msg === 'string' ? msg.trim() : '';
}

export function failureKind(cause: unknown): 'timeout' | 'network' | 'other' {
  if (!cause || typeof cause !== 'object') return 'other';
  const err = cause as RequestError;
  if (err.code === 'ECONNABORTED' || err.code === 'ETIMEDOUT') return 'timeout';
  if (typeof err.message === 'string' && /timeout/i.test(err.message) && !err.response) {
    return 'timeout';
  }
  if (err.code === 'ERR_NETWORK' || (err.request !== undefined && !err.response)) {
    return 'network';
  }
  return 'other';
}

export function failureText(cause: unknown, fallback: string, t: Translate): string {
  const kind = failureKind(cause);
  if (kind === 'timeout') return t('feedback.timeout');
  if (kind === 'network') return t('feedback.network');

  if (cause && typeof cause === 'object') {
    const err = cause as RequestError & Reply;
    // An API reply: { code, msg, data }.
    if (typeof err.code === 'number') {
      return reason(err) || fallback;
    }
    // An HTTP error whose body is still an API reply.
    const fromBody = reason(err.response?.data);
    if (fromBody) return fromBody;
  }

  return fallback;
}
