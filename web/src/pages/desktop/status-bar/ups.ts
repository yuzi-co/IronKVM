import type { NutState, NutUpsState } from '@/api/nut.ts';

// The status bar's UPS item. Pure, so it runs under node --test.

export type UpsTone = 'ok' | 'warning' | 'error' | 'off';

export type UpsLine = {
  // The UPS state, for the label and the tone.
  state: NutUpsState;
  tone: UpsTone;
  // Battery charge in percent, and the runtime in whole minutes, when the UPS
  // reports them.
  charge: number | null;
  runtimeMinutes: number | null;
};

const tones: Record<NutUpsState, UpsTone> = {
  online: 'ok',
  onBattery: 'warning',
  lowBattery: 'error',
  shutdown: 'error',
  unknown: 'off'
};

// upsLine is what the bar shows for a state, or null when it shows nothing: the
// client is off or not set up. A server that cannot be reached is shown, as an
// unknown UPS, so an outage of the link is not mistaken for a quiet UPS.
export function upsLine(state: NutState | null): UpsLine | null {
  if (!state || !state.enabled || state.status === 'off' || state.status === 'unconfigured') {
    return null;
  }

  const ups: NutUpsState = state.status === 'ok' ? state.ups : 'unknown';
  return {
    state: ups,
    tone: tones[ups],
    charge: state.status === 'ok' && state.charge !== undefined ? state.charge : null,
    runtimeMinutes:
      state.status === 'ok' && state.runtimeSeconds !== undefined
        ? Math.floor(state.runtimeSeconds / 60)
        : null
  };
}
