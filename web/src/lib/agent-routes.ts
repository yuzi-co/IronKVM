// Where the server serves the on-board agent: its runtime, its chat history
// and its chat socket. They moved from /api/picoclaw to /api/agent when a
// second agent became possible. The old paths stay on the server as aliases,
// so they are what a UI uses until it knows the server has the new ones.
//
// This module runs under node --test as well as in the browser.

import { hasFeature, type ServerContract } from './server-contract.ts';

export type AgentRoutes = {
  // base is the prefix of the runtime, session and history routes.
  base: string;
  // chatSocket is the chat socket's path.
  chatSocket: string;
  // sessionHeader names the chat session a request is about.
  sessionHeader: string;
};

export const AGENT_ROUTES: AgentRoutes = {
  base: '/api/agent',
  chatSocket: '/api/agent/chat/ws',
  sessionHeader: 'X-Agent-Session-ID'
};

export const LEGACY_AGENT_ROUTES: AgentRoutes = {
  base: '/api/picoclaw',
  chatSocket: '/api/picoclaw/gateway/ws',
  sessionHeader: 'X-PicoClaw-Session-ID'
};

// agentRoutes picks the routes for a server. contract is null while the
// server's contract is not yet known.
export function agentRoutes(contract: ServerContract | null): AgentRoutes {
  return contract !== null && hasFeature(contract, 'agent-routes') ? AGENT_ROUTES : LEGACY_AGENT_ROUTES;
}
