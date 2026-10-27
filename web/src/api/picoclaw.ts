import { currentServerContract } from '@/hooks/useServerFeature.ts';
import { agentRoutes } from '@/lib/agent-routes.ts';
import { http } from '@/lib/http.ts';
import {
  picoclawGateway,
  type GatewayRunState,
  type GatewayTransportState
} from '@/lib/picoclaw-gateway.ts';

// routes are the agent routes of this server. Model and profile settings are
// PicoClaw's own and stay under /api/picoclaw.
function routes() {
  return agentRoutes(currentServerContract());
}

type ModelConfigRequest = {
  model: string;
  api_base: string;
  api_key: string;
};

type AgentProfileRequest = {
  profile: string;
};

export type PicoclawSessionListItem = {
  id: string;
  title: string;
  preview: string;
  message_count: number;
  created: string;
  updated: string;
};

export type PicoclawSessionDetailMessage = {
  role: 'user' | 'assistant';
  content: string;
};

export type PicoclawSessionDetail = {
  id: string;
  messages: PicoclawSessionDetailMessage[];
  summary?: string;
  created: string;
  updated: string;
};

export function setPicoclawModelConfig(data: ModelConfigRequest) {
  return http.post('/api/picoclaw/model/config', data);
}

export function setPicoclawAgentProfile(data: AgentProfileRequest) {
  return http.post('/api/picoclaw/agent/profile', data);
}

export function listPicoclawSessions(params?: { offset?: number; limit?: number }) {
  return http.get(`${routes().base}/sessions`, { params });
}

export function getPicoclawSession(id: string) {
  return http.get(`${routes().base}/sessions/${id}`);
}

export function deletePicoclawSession(id: string) {
  return http.delete(`${routes().base}/sessions/${id}`);
}

export function getRuntimeStatus() {
  return http.get(`${routes().base}/runtime/status`);
}

export function startRuntime() {
  return http.post(`${routes().base}/runtime/start`);
}

export function stopRuntime() {
  return http.post(`${routes().base}/runtime/stop`);
}

export function getAIControlStatus() {
  return http.get('/api/ai/control/status');
}

export function setAIControlMode(mode: 'off' | 'mcp' | 'picoclaw') {
  return http.request({
    method: 'put',
    url: '/api/ai/control/mode',
    data: { mode }
  });
}

export function installRuntime() {
  return http.post(`${routes().base}/runtime/install`);
}

export function uninstallRuntime() {
  return http.post(`${routes().base}/runtime/uninstall`);
}

export function connectGateway(sessionId?: string) {
  return picoclawGateway.connect({ sessionId });
}

export function releaseRuntimeSession(sessionId?: string) {
  const activeSessionId = sessionId || picoclawGateway.getSessionId();
  if (!activeSessionId) {
    return Promise.resolve(null);
  }

  const { base, sessionHeader } = routes();
  return http.request({
    method: 'delete',
    url: `${base}/runtime/session`,
    headers: {
      [sessionHeader]: activeSessionId
    }
  });
}

export function sendChatMessage(
  content: string,
  options?: { id?: string; maxSteps?: number; maxRuntimeMs?: number; trackState?: boolean }
) {
  return picoclawGateway.sendChatMessage(content, options);
}

export function sendStopMessage() {
  return picoclawGateway.sendStopMessage();
}

export async function closeGateway() {
  const activeSessionId = picoclawGateway.getSessionId();
  picoclawGateway.close();
  if (!activeSessionId) {
    return;
  }

  await releaseRuntimeSession(activeSessionId).catch(() => undefined);
}

export function onGatewayConnectionState(listener: (state: GatewayTransportState) => void) {
  return picoclawGateway.on('transport_state', listener);
}

export function onGatewayRunState(listener: (state: GatewayRunState) => void) {
  return picoclawGateway.on('run_state', listener);
}

export { picoclawGateway };
