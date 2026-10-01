import assert from 'node:assert/strict';
import { test } from 'node:test';

import { modelProviderAllowsEmptyApiKey } from './picoclaw-model.ts';

test('local inference providers need no API key', () => {
  for (const id of [
    'ollama/qwen3.5:9b',
    ' Ollama/llama3 ',
    'lmstudio/local-model',
    'vllm/meta-llama/x'
  ]) {
    assert.equal(modelProviderAllowsEmptyApiKey(id), true, id);
  }
});

test('hosted providers and bare names still need one', () => {
  for (const id of ['openai/gpt-5.4', 'openai_compatible/x', 'qwen3.5:9b', '/ollama', '']) {
    assert.equal(modelProviderAllowsEmptyApiKey(id), false, JSON.stringify(id));
  }
});
