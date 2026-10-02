import assert from 'node:assert/strict';
import { test } from 'node:test';

import { picoAssistantKind, picoError, picoMessageId, picoTurnDone } from './pico-message.ts';

test('the message id comes from the payload, as PicoClaw 0.3 sends it', () => {
  assert.equal(
    picoMessageId({ type: 'message.create', payload: { content: 'hi', message_id: 'm1' } }),
    'm1'
  );
  assert.equal(picoMessageId({ type: 'message.create', id: 'm2', payload: {} }), 'm2');
  assert.equal(picoMessageId({ type: 'message.create', id: '', payload: {} }), undefined);
});

test('reasoning and tool calls are not shown as replies', () => {
  assert.equal(
    picoAssistantKind({ payload: { kind: 'thought', thought: true, content: 'x' } }),
    'hidden'
  );
  assert.equal(picoAssistantKind({ payload: { thought: true, content: 'x' } }), 'hidden');
  assert.equal(picoAssistantKind({ payload: { kind: 'tool_calls', content: '' } }), 'hidden');
});

test('a placeholder does not end the turn, a reply does', () => {
  assert.equal(
    picoAssistantKind({ payload: { placeholder: true, content: 'Thinking' } }),
    'placeholder'
  );
  assert.equal(picoAssistantKind({ payload: { content: 'Done', context_usage: {} } }), 'reply');
  assert.equal(picoAssistantKind({ content: 'Done' }), 'reply');
});

test('an error from PicoClaw carries its code and text in the payload', () => {
  assert.deepEqual(
    picoError({
      type: 'error',
      payload: {
        code: 'command_disabled',
        message: 'control commands are disabled',
        request_id: 'r1'
      }
    }),
    { code: 'command_disabled', message: 'control commands are disabled', requestScoped: true }
  );
});

test('an error from the server keeps its top-level fields', () => {
  assert.deepEqual(picoError({ type: 'error', code: 'AI_LOCK_HELD', message: 'held' }), {
    code: 'AI_LOCK_HELD',
    message: 'held',
    requestScoped: false
  });
  assert.deepEqual(picoError({ type: 'error' }), {
    code: 'ERROR',
    message: 'Gateway error',
    requestScoped: false
  });
});

test('turn.done gives the requests, the status and the usage of the turn', () => {
  assert.deepEqual(
    picoTurnDone({
      type: 'turn.done',
      payload: {
        request_id: 'req-1',
        request_ids: ['req-1', 'req-2'],
        status: 'ok',
        usage: { input_tokens: 4380, output_tokens: 328, total_tokens: 4708, llm_calls: 2 }
      }
    }),
    {
      requestIds: ['req-1', 'req-2'],
      status: 'ok',
      usage: { inputTokens: 4380, outputTokens: 328, totalTokens: 4708, llmCalls: 2 }
    }
  );
});

test('turn.done without usage or ids, and with an error or cancel status', () => {
  assert.deepEqual(picoTurnDone({ type: 'turn.done', payload: { status: 'error' } }), {
    requestIds: [],
    status: 'error',
    usage: undefined
  });
  assert.equal(
    picoTurnDone({ type: 'turn.done', payload: { status: 'canceled' } })?.status,
    'canceled'
  );
  assert.equal(picoTurnDone({ type: 'turn.done', payload: { status: 'odd' } })?.status, 'ok');
  assert.deepEqual(picoTurnDone({ type: 'turn.done', payload: { request_id: 'r' } })?.requestIds, [
    'r'
  ]);
});

test('other messages are not turn.done', () => {
  assert.equal(picoTurnDone({ type: 'message.create', payload: { status: 'ok' } }), undefined);
});
