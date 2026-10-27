import assert from 'node:assert/strict';
import { test } from 'node:test';

import { AGENT_ROUTES, agentRoutes, LEGACY_AGENT_ROUTES } from './agent-routes.ts';
import { contractFromResponse, LEGACY_CONTRACT } from './server-contract.ts';

test('a server that declares agent-routes gets the new paths', () => {
  const c = contractFromResponse({ api: 2, features: ['agent-routes'] });
  assert.equal(agentRoutes(c), AGENT_ROUTES);
});

test('a server without it, an older server and an unknown one get the old paths', () => {
  assert.equal(agentRoutes(contractFromResponse({ api: 2, features: [] })), LEGACY_AGENT_ROUTES);
  assert.equal(agentRoutes(LEGACY_CONTRACT), LEGACY_AGENT_ROUTES);
  assert.equal(agentRoutes(null), LEGACY_AGENT_ROUTES);
});

test('both sets keep the chat socket and the session routes under one base', () => {
  for (const routes of [AGENT_ROUTES, LEGACY_AGENT_ROUTES]) {
    assert.ok(routes.chatSocket.startsWith(routes.base + '/'), routes.chatSocket);
  }
});
