// The chat socket's protocol. The server translates the on-device agent's own
// protocol into these normalized events, so this file and the sidebar do not
// depend on which agent runs. server/service/agent/event.go is the source of
// the shapes; the names follow the Agent Client Protocol where it fits.

export type AgentMessageKind = 'reply' | 'placeholder' | 'hidden';

export type AgentStopReason = 'end_turn' | 'cancelled' | 'error';

export type AgentUsage = {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  llmCalls?: number;
};

export type AgentControlMode = {
  mode: string;
  transitioning: boolean;
  canControl: boolean;
  lastError?: string;
  changedAt?: string;
  source?: string;
};

export type AgentEvent =
  | { type: 'turn_started' }
  | { type: 'agent_message'; messageId: string; kind: AgentMessageKind; text: string }
  | { type: 'agent_message_removed'; messageId: string }
  | { type: 'tool_call'; toolCallId: string; title: string; x?: number; y?: number }
  | {
      type: 'observation';
      messageId: string;
      text?: string;
      imageBase64: string;
      mimeType?: string;
    }
  | { type: 'error'; code: string; message: string; requestId?: string }
  | { type: 'turn_done'; requestIds: string[]; stopReason: AgentStopReason; usage?: AgentUsage }
  | { type: 'control_mode_changed'; control: AgentControlMode };

// Commands the browser sends.
export type AgentCommand =
  | {
      type: 'session/prompt';
      requestId: string;
      text: string;
      maxSteps?: number;
      maxRuntimeMs?: number;
    }
  | { type: 'session/cancel'; requestId?: string };

type JsonObject = Record<string, unknown>;

function str(value: unknown): string | undefined {
  return typeof value === 'string' && value !== '' ? value : undefined;
}

function num(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined;
}

// parseAgentEvent reads one message of the chat socket. It returns undefined
// for a message that is not a known event.
export function parseAgentEvent(data: unknown): AgentEvent | undefined {
  if (!data || typeof data !== 'object') {
    return undefined;
  }
  const event = data as JsonObject;

  switch (event.type) {
    case 'turn_started':
      return { type: 'turn_started' };
    case 'agent_message': {
      const messageId = str(event.messageId);
      if (!messageId) return undefined;
      const kind = event.kind === 'placeholder' || event.kind === 'hidden' ? event.kind : 'reply';
      return { type: 'agent_message', messageId, kind, text: str(event.text) ?? '' };
    }
    case 'agent_message_removed': {
      const messageId = str(event.messageId);
      return messageId ? { type: 'agent_message_removed', messageId } : undefined;
    }
    case 'tool_call': {
      const title = str(event.title);
      if (!title) return undefined;
      return {
        type: 'tool_call',
        toolCallId: str(event.toolCallId) ?? '',
        title,
        x: num(event.x),
        y: num(event.y)
      };
    }
    case 'observation': {
      const imageBase64 = str(event.imageBase64);
      if (!imageBase64) return undefined;
      return {
        type: 'observation',
        messageId: str(event.messageId) ?? '',
        text: str(event.text),
        imageBase64,
        mimeType: str(event.mimeType)
      };
    }
    case 'error':
      return {
        type: 'error',
        code: str(event.code) ?? 'ERROR',
        message: str(event.message) ?? 'Gateway error',
        requestId: str(event.requestId)
      };
    case 'turn_done': {
      const requestIds = Array.isArray(event.requestIds)
        ? event.requestIds.filter((id): id is string => typeof id === 'string' && id !== '')
        : [];
      const stopReason =
        event.stopReason === 'cancelled' || event.stopReason === 'error'
          ? event.stopReason
          : 'end_turn';
      return { type: 'turn_done', requestIds, stopReason, usage: usageOf(event.usage) };
    }
    case 'control_mode_changed': {
      const control = event.control;
      if (!control || typeof control !== 'object') return undefined;
      const value = control as JsonObject;
      return {
        type: 'control_mode_changed',
        control: {
          mode: str(value.mode) ?? 'off',
          transitioning: value.transitioning === true,
          canControl: value.canControl === true,
          lastError: str(value.lastError),
          changedAt: str(value.changedAt),
          source: str(value.source)
        }
      };
    }
    default:
      return undefined;
  }
}

function usageOf(value: unknown): AgentUsage | undefined {
  if (!value || typeof value !== 'object') {
    return undefined;
  }
  const usage = value as JsonObject;
  return {
    inputTokens: num(usage.inputTokens) ?? 0,
    outputTokens: num(usage.outputTokens) ?? 0,
    totalTokens: num(usage.totalTokens) ?? 0,
    llmCalls: num(usage.llmCalls)
  };
}

// What the sidebar does with an event, decided without touching the UI so a
// test can check it.
export type AgentEventSink = {
  runState: (state: 'idle' | 'busy') => void;
  transportError: () => void;
  assistantMessage: (message: { id: string; text: string }) => void;
  assistantMessageDelete: (id: string) => void;
  toolAction: (action: { id: string; action: string; x?: number; y?: number }) => void;
  observation: (observation: { id: string; text?: string; imageBase64: string }) => void;
  error: (error: { code: string; message: string; raw?: unknown }) => void;
  turnDone: (done: Extract<AgentEvent, { type: 'turn_done' }>) => void;
  controlModeChanged: (control: AgentControlMode) => void;
};

// createAgentEventHandler returns the function that handles each message of
// one chat socket. newId names a message the server sent without an id.
export function createAgentEventHandler(sink: AgentEventSink, newId: () => string) {
  // Set once the agent has sent a turn_done. From then on only turn_done ends
  // a turn; before, a reply does, for an agent that never sends it.
  let sendsTurnDone = false;

  return (rawData: unknown) => {
    if (typeof rawData !== 'string') {
      return;
    }
    let data: unknown;
    try {
      data = JSON.parse(rawData);
    } catch {
      sink.error({
        code: 'INVALID_MESSAGE',
        message: 'Failed to parse gateway message',
        raw: rawData
      });
      return;
    }
    const event = parseAgentEvent(data);
    if (!event) {
      return;
    }

    switch (event.type) {
      case 'turn_started':
        sink.runState('busy');
        return;
      case 'error':
        // An error that names a request, such as a refused command, leaves
        // the connection fine.
        if (!event.requestId) {
          sink.transportError();
        }
        sink.runState('idle');
        sink.error({ code: event.code, message: event.message, raw: event });
        return;
      case 'turn_done':
        sendsTurnDone = true;
        sink.runState('idle');
        sink.turnDone(event);
        return;
      case 'control_mode_changed':
        sink.controlModeChanged(event.control);
        return;
      case 'agent_message':
        if (event.kind === 'hidden') {
          return;
        }
        if (event.kind === 'reply' && !sendsTurnDone) {
          sink.runState('idle');
        }
        sink.assistantMessage({ id: event.messageId, text: event.text });
        return;
      case 'agent_message_removed':
        sink.assistantMessageDelete(event.messageId);
        return;
      case 'observation':
        sink.observation({
          id: event.messageId || newId(),
          text: event.text,
          imageBase64: event.imageBase64
        });
        return;
      case 'tool_call':
        sink.toolAction({
          id: event.toolCallId || newId(),
          action: event.title,
          x: event.x,
          y: event.y
        });
        return;
    }
  };
}
