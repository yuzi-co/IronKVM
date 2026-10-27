// Pure helpers of the audit page, kept apart so node --test can run them.

type WhoFields = { actor?: string; via?: string; key?: string };

// who names the actor of an event the way a reviewer reads it: the account,
// with the API key it used when there was one.
export function who(event: WhoFields): string {
  const actor = event.actor || '-';
  return event.key ? `${actor} (${event.key})` : actor;
}

// via is how the actor got in, with the start of the chat session an agent
// acted for: enough to tell two conversations apart in the list. The full id
// is in the row's tooltip and the export.
export function via(event: { via?: string; session?: string }): string {
  const session = event.session ? event.session.slice(0, 8) : '';
  return [event.via, session].filter(Boolean).join(' · ');
}

// subject joins what an event acted on with its remark, either of which may
// be missing.
export function subject(event: { target?: string; detail?: string; repeat?: number }): string {
  const parts = [event.target, event.detail].filter(Boolean) as string[];
  if (event.repeat && event.repeat > 1) parts.push(`x${event.repeat}`);
  return parts.join(' - ');
}

// auditFileName is the name an export is saved under.
export function auditFileName(host: string, now: Date): string {
  const day = now.toISOString().slice(0, 10);
  return `ironkvm-audit-${host}-${day}.jsonl`;
}
