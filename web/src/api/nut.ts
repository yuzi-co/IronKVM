import { http } from '@/lib/http.ts';

export type NutSettings = {
  enabled: boolean;
  host: string;
  port: number;
  ups: string;
  username: string;
  // The password never comes back from the server: this says whether one is
  // stored.
  passwordSet: boolean;
  onBatteryMinutes: number;
  forceOffMinutes: number;
  powerOnRestore: boolean;
  powerOnMinCharge: number;
  powerOnDelaySeconds: number;
};

// What the page sends: the settings, and a password only when it changes.
export type NutSettingsRequest = Omit<NutSettings, 'passwordSet'> & {
  password?: string;
  clearPassword?: boolean;
};

export type NutStatus = 'off' | 'unconfigured' | 'connecting' | 'ok' | 'unreachable';
export type NutUpsState = 'online' | 'onBattery' | 'lowBattery' | 'shutdown' | 'unknown';
export type NutHostState = 'on' | 'off' | 'unknown';
export type NutPhase = 'idle' | 'shuttingDown' | 'waitingForMains' | 'needsLed';

export type NutEventKind =
  | 'unreachable'
  | 'reachable'
  | 'onBattery'
  | 'mainsBack'
  | 'shutdown'
  | 'forceOff'
  | 'hostOff'
  | 'powerOn'
  | 'pressFailed';

export type NutEvent = {
  time: string;
  kind: NutEventKind;
  // The reason of a shutdown (fsd, lowBattery, onBattery), or an error.
  detail?: string;
};

export type NutState = {
  enabled: boolean;
  status: NutStatus;
  reachable: boolean;
  ups: NutUpsState;
  flags?: string[];
  charge?: number;
  runtimeSeconds?: number;
  onBatterySeconds?: number;
  loggedIn: boolean;
  loginError?: string;
  host: NutHostState;
  phase: NutPhase;
  powerOnInSeconds?: number;
  lastPoll?: string;
  events: NutEvent[];
};

export function getNutSettings() {
  return http.get('/api/nut/settings');
}

export function setNutSettings(settings: NutSettingsRequest) {
  return http.post('/api/nut/settings', settings);
}

export function getNutState() {
  return http.get('/api/nut/state');
}
