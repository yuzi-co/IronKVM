import assert from 'node:assert/strict';
import { test } from 'node:test';

import { createAgentEventHandler, parseAgentEvent, type AgentEventSink } from './agent-events.ts';

type Call = [string, unknown];

function recorder() {
  const calls: Call[] = [];
  const sink: AgentEventSink = {
    runState: (state) => calls.push(['runState', state]),
    transportError: () => calls.push(['transportError', undefined]),
    assistantMessage: (message) => calls.push(['assistantMessage', message]),
    assistantMessageDelete: (id) => calls.push(['assistantMessageDelete', id]),
    toolAction: (action) => calls.push(['toolAction', action]),
    observation: (observation) => calls.push(['observation', observation]),
    error: (error) => calls.push(['error', error]),
    turnDone: (done) => calls.push(['turnDone', done]),
    controlModeChanged: (control) => calls.push(['controlModeChanged', control])
  };
  let next = 0;
  const handle = createAgentEventHandler(sink, () => `gen-${++next}`);
  return {
    calls,
    send: (event: unknown) => handle(JSON.stringify(event)),
    sendRaw: handle,
    take: () => calls.splice(0, calls.length)
  };
}

test('a turn of an agent that sends turn_done', () => {
  const { send, take } = recorder();
  send({ type: 'turn_started' });
  assert.deepEqual(take(), [['runState', 'busy']]);

  send({ type: 'agent_message', messageId: 'm1', kind: 'hidden', text: 'reasoning' });
  assert.deepEqual(take(), [], 'reasoning and tool calls are not shown');

  send({ type: 'agent_message', messageId: 'm2', kind: 'placeholder', text: 'Thinking' });
  assert.deepEqual(take(), [['assistantMessage', { id: 'm2', text: 'Thinking' }]]);

  send({
    type: 'turn_done',
    requestIds: ['r1', 'r2'],
    stopReason: 'end_turn',
    usage: { inputTokens: 4380, outputTokens: 328, totalTokens: 4708, llmCalls: 2 }
  });
  assert.deepEqual(take(), [
    ['runState', 'idle'],
    [
      'turnDone',
      {
        type: 'turn_done',
        requestIds: ['r1', 'r2'],
        stopReason: 'end_turn',
        usage: { inputTokens: 4380, outputTokens: 328, totalTokens: 4708, llmCalls: 2 }
      }
    ]
  ]);

  // From now on a reply does not end the turn; turn_done does.
  send({ type: 'agent_message', messageId: 'm3', kind: 'reply', text: 'Done' });
  assert.deepEqual(take(), [['assistantMessage', { id: 'm3', text: 'Done' }]]);
});

test('a reply ends the turn of an agent that never sends turn_done', () => {
  const { send, take } = recorder();
  send({ type: 'agent_message', messageId: 'm1', kind: 'placeholder', text: '...' });
  assert.deepEqual(take(), [['assistantMessage', { id: 'm1', text: '...' }]]);
  send({ type: 'agent_message', messageId: 'm1', kind: 'reply', text: 'Done' });
  assert.deepEqual(take(), [
    ['runState', 'idle'],
    ['assistantMessage', { id: 'm1', text: 'Done' }]
  ]);
  send({ type: 'agent_message_removed', messageId: 'm1' });
  assert.deepEqual(take(), [['assistantMessageDelete', 'm1']]);
});

test('an error about one request leaves the connection fine', () => {
  const { send, take } = recorder();
  send({ type: 'error', code: 'command_disabled', message: 'no', requestId: 'r1' });
  const calls = take();
  assert.deepEqual(calls.slice(0, 1), [['runState', 'idle']]);
  assert.equal((calls[1][1] as { code: string }).code, 'command_disabled');

  send({ type: 'error' });
  const [transport, idle, error] = take();
  assert.deepEqual(transport, ['transportError', undefined]);
  assert.deepEqual(idle, ['runState', 'idle']);
  assert.deepEqual(
    { ...(error[1] as Record<string, unknown>), raw: undefined },
    { code: 'ERROR', message: 'Gateway error', raw: undefined }
  );
});

test('observations, tool calls and control changes', () => {
  const { send, take } = recorder();
  send({ type: 'observation', messageId: 'o1', text: 'shot', imageBase64: 'abc' });
  send({ type: 'observation', imageBase64: 'def' });
  send({ type: 'observation', messageId: 'o3' });
  send({ type: 'tool_call', toolCallId: 't1', title: 'click', x: 0.5, y: 0.25 });
  send({
    type: 'control_mode_changed',
    control: { mode: 'mcp', transitioning: true, canControl: false, source: 'mcp_config' }
  });
  assert.deepEqual(take(), [
    ['observation', { id: 'o1', text: 'shot', imageBase64: 'abc' }],
    ['observation', { id: 'gen-1', text: undefined, imageBase64: 'def' }],
    ['toolAction', { id: 't1', action: 'click', x: 0.5, y: 0.25 }],
    [
      'controlModeChanged',
      {
        mode: 'mcp',
        transitioning: true,
        canControl: false,
        lastError: undefined,
        changedAt: undefined,
        source: 'mcp_config'
      }
    ]
  ]);
});

test('unknown and unreadable messages', () => {
  const { send, sendRaw, take } = recorder();
  send({ type: 'message.create', payload: { content: 'raw pico' } });
  send({ type: 'agent_message', kind: 'reply', text: 'no id' });
  sendRaw(new ArrayBuffer(1));
  assert.deepEqual(take(), []);

  sendRaw('not json');
  assert.deepEqual(take(), [
    [
      'error',
      { code: 'INVALID_MESSAGE', message: 'Failed to parse gateway message', raw: 'not json' }
    ]
  ]);
});

test('turn_done stop reasons', () => {
  assert.equal(parseAgentEvent({ type: 'turn_done', stopReason: 'cancelled' })?.type, 'turn_done');
  const reason = (stopReason: unknown) => {
    const event = parseAgentEvent({ type: 'turn_done', stopReason });
    return event?.type === 'turn_done' ? event.stopReason : undefined;
  };
  assert.equal(reason('cancelled'), 'cancelled');
  assert.equal(reason('error'), 'error');
  assert.equal(reason('odd'), 'end_turn');
});
