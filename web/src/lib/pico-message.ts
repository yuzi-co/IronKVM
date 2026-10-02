// Reading the Pico protocol messages PicoClaw sends through the gateway.
//
// PicoClaw 0.3 puts the message id, the error code and the error text in the
// payload, and sends its reasoning and its tool calls as messages of their own,
// marked by payload.kind. The server's own messages keep them at the top level.

type PicoMessage = Record<string, unknown>;

function payloadOf(message: PicoMessage): Record<string, unknown> {
  const payload = message.payload;
  return payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {};
}

function nonEmptyString(value: unknown): string | undefined {
  return typeof value === 'string' && value !== '' ? value : undefined;
}

// The id a later message.update or message.delete refers to.
export function picoMessageId(message: PicoMessage): string | undefined {
  return nonEmptyString(payloadOf(message).message_id) ?? nonEmptyString(message.id);
}

// How a message.create or message.update is shown:
//   reply        the assistant's answer, which also ends the turn
//   placeholder  a "thinking" text that a reply replaces; the turn goes on
//   hidden       reasoning or a tool call; the screenshots and actions show
//                what the tools did
export type PicoAssistantKind = 'reply' | 'placeholder' | 'hidden';

export function picoAssistantKind(message: PicoMessage): PicoAssistantKind {
  const payload = payloadOf(message);
  if (payload.kind === 'thought' || payload.kind === 'tool_calls' || payload.thought === true) {
    return 'hidden';
  }
  if (payload.placeholder === true) {
    return 'placeholder';
  }
  return 'reply';
}

export type PicoError = {
  code: string;
  message: string;
  // An error about one request, such as a refused command. The connection is fine.
  requestScoped: boolean;
};

export function picoError(message: PicoMessage): PicoError {
  const payload = payloadOf(message);
  return {
    code: nonEmptyString(message.code) ?? nonEmptyString(payload.code) ?? 'ERROR',
    message: nonEmptyString(message.message) ?? nonEmptyString(payload.message) ?? 'Gateway error',
    requestScoped: nonEmptyString(payload.request_id) !== undefined
  };
}
