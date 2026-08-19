import { IMessageEvent, w3cwebsocket as W3cWebSocket } from 'websocket';

import { getBaseUrl } from '@/lib/service.ts';

type MessageHandler = (message: IMessageEvent) => void;
type SendData = number[] | ArrayBuffer | Uint8Array;

export enum MessageEvent {
  Heartbeat = 0,
  Keyboard = 1,
  Mouse = 2
}

// Keyboard and mouse travel over this socket and nothing else carries them, so
// a socket that will not open is an input outage even though the page, the REST
// calls and the video all keep working.
//
// A browser will not say so. It asks the operator about an untrusted
// certificate when it loads a page and refuses a websocket to the same origin
// without asking, reporting nothing the page can catch. Enabling HTTPS was
// therefore enough to remove the input with no message anywhere, which is what
// this state exists to make visible.
export type WsStatus = {
  connected: boolean;
  // Retries since the last time the socket was open. This client retries for
  // as long as the tab is open, so the count is the only measure of how wrong
  // things are.
  attempts: number;
  // False until the socket has opened at least once. A socket that has never
  // opened is a different fault from one that dropped: the first points at the
  // certificate or a blocked port, the second usually at a server restart.
  everConnected: boolean;
};

type StatusHandler = (status: WsStatus) => void;

interface WsClientOptions {
  url?: string;
  heartbeatInterval?: number;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

const DEFAULT_OPTIONS: Required<WsClientOptions> = {
  url: `${getBaseUrl('ws')}/api/ws`,
  heartbeatInterval: 10 * 1000,
  reconnectInterval: 3 * 1000,
  maxReconnectAttempts: Number.POSITIVE_INFINITY
};

export class WsClient {
  private readonly options: Required<WsClientOptions>;
  private instance: W3cWebSocket | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private shouldReconnect = true;
  private everConnected = false;
  private readonly statusHandlers = new Set<StatusHandler>();

  private readonly eventHandlers = new Map<string, Set<MessageHandler>>();

  constructor(options: WsClientOptions = {}) {
    this.options = { ...DEFAULT_OPTIONS, ...options };
  }

  public connect(): void {
    this.shouldReconnect = true;
    this.reconnectAttempts = 0;
    this.createConnection();
  }

  public close(): void {
    this.shouldReconnect = false;
    this.cleanup();

    if (this.instance && this.instance.readyState === W3cWebSocket.OPEN) {
      this.instance.close();
    }

    this.instance = null;
  }

  public on(type: string, handler: MessageHandler): () => void {
    if (!this.eventHandlers.has(type)) {
      this.eventHandlers.set(type, new Set());
    }

    this.eventHandlers.get(type)!.add(handler);

    return () => {
      const handlers = this.eventHandlers.get(type);
      if (handlers) {
        handlers.delete(handler);
        if (handlers.size === 0) {
          this.eventHandlers.delete(type);
        }
      }
    };
  }

  public off(type: string, handler?: MessageHandler): void {
    if (handler) {
      const handlers = this.eventHandlers.get(type);
      if (handlers) {
        handlers.delete(handler);
        if (handlers.size === 0) {
          this.eventHandlers.delete(type);
        }
      }
    } else {
      this.eventHandlers.delete(type);
    }
  }

  public send(data: SendData): boolean {
    if (!this.instance || !this.isConnected) {
      return false;
    }

    if (data instanceof ArrayBuffer || (data as unknown) instanceof Uint8Array) {
      this.instance.send(data);
    } else {
      this.instance.send(JSON.stringify(data));
    }

    return true;
  }

  public get isConnected(): boolean {
    return this.instance?.readyState === W3cWebSocket.OPEN;
  }

  // onStatus reports the connection state and returns an unsubscribe function.
  // It fires immediately with the current state, so a component that mounts
  // after the socket has already failed still learns about it.
  public onStatus(handler: StatusHandler): () => void {
    this.statusHandlers.add(handler);
    handler(this.status);

    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  public get status(): WsStatus {
    return {
      connected: this.isConnected,
      attempts: this.reconnectAttempts,
      everConnected: this.everConnected
    };
  }

  private emitStatus(): void {
    const status = this.status;
    this.statusHandlers.forEach((handler) => handler(status));
  }

  private createConnection(): void {
    this.cleanup();

    this.instance = new W3cWebSocket(this.options.url);
    this.instance.binaryType = 'arraybuffer';

    this.instance.onopen = this.handleOpen.bind(this);
    this.instance.onclose = this.handleClose.bind(this);
    this.instance.onerror = this.handleError.bind(this);
    this.instance.onmessage = this.handleMessage.bind(this);
  }

  private handleOpen(): void {
    this.reconnectAttempts = 0;
    this.everConnected = true;
    this.startHeartbeat();
    this.emitStatus();
  }

  private handleClose(): void {
    this.stopHeartbeat();
    this.emitStatus();
    this.scheduleReconnect();
  }

  private handleError(error: Error): void {
    console.error('[WebSocket] Error:', error);
  }

  private handleMessage(message: IMessageEvent): void {
    try {
      const data = JSON.parse(message.data as string);
      const handlers = this.eventHandlers.get(data.type);

      if (handlers) {
        handlers.forEach((handler) => handler(message));
      }
    } catch (err) {
      console.log(err);
    }
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      this.send(new Uint8Array([MessageEvent.Heartbeat]));
    }, this.options.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect) {
      return;
    }

    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.error('[WebSocket] Max reconnect attempts reached');
      return;
    }

    this.reconnectAttempts++;
    console.log(`[WebSocket] Reconnecting... (attempt ${this.reconnectAttempts})`);
    this.emitStatus();

    this.reconnectTimer = setTimeout(() => {
      this.createConnection();
    }, this.options.reconnectInterval);
  }

  private cleanup(): void {
    this.stopHeartbeat();

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}

export const client = new WsClient();
