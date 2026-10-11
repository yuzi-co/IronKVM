// Macros and the key auto-tap: the steps the server plays as a timed input
// sequence (server/service/sequence), and the recorder that turns what this
// tab sends to the host into steps.
//
// This module runs under node --test as well as in the browser, so it imports
// nothing but the key map.
import { KeycodeMap, ModifierBits } from './keymap.ts';

// The server's SequenceStep (server/proto/sequence.go).
export type KeyStep = { type: 'key'; modifiers?: number; keys?: number[] };
export type MouseStep = {
  type: 'mouse';
  absolute?: boolean;
  buttons?: number;
  x?: number;
  y?: number;
  wheel?: number;
  hwheel?: number;
};
export type DelayStep = { type: 'delay'; ms: number };
export type AtxButton = 'power' | 'reset';
export type AtxStep = { type: 'atx'; button: AtxButton; ms?: number };
export type MacroStep = KeyStep | MouseStep | DelayStep | AtxStep;

export type Macro = { id: string; name: string; steps: MacroStep[] };

// The server's limits, server/service/sequence/steps.go.
export const MAX_STEPS = 2000;
export const MAX_DELAY_MS = 60000;
export const MAX_NAME_LENGTH = 64;
export const DEFAULT_ATX_MS = 800;

// The first byte of a websocket input message, lib/websocket.ts. Repeated
// here because that module cannot load outside a browser.
const KEYBOARD_MESSAGE = 1;
const MOUSE_MESSAGE = 2;

// Consecutive pointer moves closer together than this are kept as one step.
// A pointer sends a report per animation frame; a macro needs where it went,
// not every frame on the way.
const MERGE_MOVES_MS = 50;

function int8(byte: number): number {
  return byte > 127 ? byte - 256 : byte;
}

// stepFromMessage reads one keyboard or pointer message, as the websocket
// sends it, into a step. Anything else (heartbeat, keepalive, touch) is null.
export function stepFromMessage(data: ArrayLike<number>): KeyStep | MouseStep | null {
  if (data[0] === KEYBOARD_MESSAGE && data.length === 9) {
    const keys: number[] = [];
    for (let i = 3; i < 9; i++) {
      if (data[i] !== 0) keys.push(data[i]);
    }
    const step: KeyStep = { type: 'key' };
    if (data[1] !== 0) step.modifiers = data[1];
    if (keys.length > 0) step.keys = keys;
    return step;
  }

  if (data[0] === MOUSE_MESSAGE && data.length === 6) {
    return compactMouse({
      type: 'mouse',
      buttons: data[1],
      x: int8(data[2]),
      y: int8(data[3]),
      wheel: int8(data[4]),
      hwheel: int8(data[5])
    });
  }

  if (data[0] === MOUSE_MESSAGE && data.length === 7) {
    return compactMouse({
      type: 'mouse',
      absolute: true,
      buttons: data[1],
      x: data[2] | (data[3] << 8),
      y: data[4] | (data[5] << 8),
      wheel: int8(data[6])
    });
  }

  return null;
}

// compactMouse drops the zero fields, which the server reads as zero, so a
// saved macro stays small.
function compactMouse(step: MouseStep): MouseStep {
  const out: MouseStep = { type: 'mouse' };
  if (step.absolute) out.absolute = true;
  if (step.buttons) out.buttons = step.buttons;
  if (step.x || (step.absolute && step.x === 0)) out.x = step.x;
  if (step.y || (step.absolute && step.y === 0)) out.y = step.y;
  if (step.wheel) out.wheel = step.wheel;
  if (step.hwheel) out.hwheel = step.hwheel;
  return out;
}

function isMove(step: MacroStep | undefined): step is MouseStep {
  return step?.type === 'mouse' && !step.wheel && !step.hwheel;
}

// MacroRecorder turns timed input into steps: a pause for each gap, pointer
// moves merged, and at the end a release for whatever is still held, so a
// macro never leaves a key down on the host.
export class MacroRecorder {
  private readonly steps: MacroStep[] = [];
  private last = -1;
  private keyboardHeld = false;
  private relativeButtons = 0;
  private absoluteButtons = 0;
  private absolutePosition = { x: 0, y: 0 };

  private readonly maxSteps: number;

  constructor(maxSteps = MAX_STEPS) {
    this.maxSteps = maxSteps;
  }

  get count(): number {
    return this.steps.length;
  }

  // full says the recording has reached the step limit and records no more.
  // Three steps are kept back for the releases finish adds.
  get full(): boolean {
    return this.steps.length >= this.maxSteps - 3;
  }

  // message records a websocket input message sent at now, in ms.
  message(data: ArrayLike<number>, now: number): boolean {
    const step = stepFromMessage(data);
    if (!step) return false;
    return this.add(step, now);
  }

  // atx records a front-panel press made at now. Playing the step holds the
  // button for ms before the next step, so the next gap is counted from the
  // release.
  atx(button: AtxButton, ms: number, now: number): boolean {
    if (!this.add({ type: 'atx', button, ms }, now)) return false;
    this.last = now + ms;
    return true;
  }

  add(step: MacroStep, now: number): boolean {
    if (this.full) return false;

    const gap = this.last < 0 ? 0 : Math.max(0, Math.round(now - this.last));
    this.last = now;

    if (this.mergeMove(step, gap)) return true;

    this.pause(gap);
    this.steps.push(step);
    this.track(step);
    return true;
  }

  private mergeMove(step: MacroStep, gap: number): boolean {
    if (!isMove(step) || gap >= MERGE_MOVES_MS) return false;

    const n = this.steps.length;
    const trailing = this.steps[n - 1]?.type === 'delay' ? 1 : 0;
    const previous = this.steps[n - 1 - trailing];
    if (!isMove(previous)) return false;
    if (!!previous.absolute !== !!step.absolute) return false;
    if ((previous.buttons ?? 0) !== (step.buttons ?? 0)) return false;

    if (step.absolute) {
      previous.x = step.x;
      previous.y = step.y;
    } else {
      const x = (previous.x ?? 0) + (step.x ?? 0);
      const y = (previous.y ?? 0) + (step.y ?? 0);
      if (Math.abs(x) > 127 || Math.abs(y) > 127) return false;
      if (x) previous.x = x;
      else delete previous.x;
      if (y) previous.y = y;
      else delete previous.y;
    }
    this.pause(gap);
    this.track(step);
    return true;
  }

  // pause adds ms to the pause at the end, or starts one. A pause longer
  // than the server takes in one step becomes several.
  private pause(ms: number) {
    let rest = ms;
    while (rest > 0) {
      const tail = this.steps[this.steps.length - 1];
      if (tail?.type === 'delay' && tail.ms < MAX_DELAY_MS) {
        const add = Math.min(rest, MAX_DELAY_MS - tail.ms);
        tail.ms += add;
        rest -= add;
        continue;
      }
      if (this.steps.length === 0) return; // no pause before the first step
      const add = Math.min(rest, MAX_DELAY_MS);
      this.steps.push({ type: 'delay', ms: add });
      rest -= add;
    }
  }

  private track(step: MacroStep) {
    if (step.type === 'key') {
      this.keyboardHeld = !!step.modifiers || (step.keys?.length ?? 0) > 0;
    } else if (step.type === 'mouse' && step.absolute) {
      this.absoluteButtons = step.buttons ?? 0;
      this.absolutePosition = { x: step.x ?? 0, y: step.y ?? 0 };
    } else if (step.type === 'mouse') {
      this.relativeButtons = step.buttons ?? 0;
    }
  }

  // finish returns the steps with no pause at the end and nothing held.
  finish(): MacroStep[] {
    const steps = this.steps.slice();
    while (steps[steps.length - 1]?.type === 'delay') steps.pop();
    if (this.keyboardHeld) steps.push({ type: 'key' });
    if (this.relativeButtons) steps.push({ type: 'mouse' });
    if (this.absoluteButtons) {
      steps.push({ type: 'mouse', absolute: true, ...this.absolutePosition });
    }
    return steps;
  }
}

// durationMs is how long a macro takes to play: its pauses and its ATX
// presses. Reports take next to nothing.
export function durationMs(steps: MacroStep[]): number {
  let total = 0;
  for (const step of steps) {
    if (step.type === 'delay') total += step.ms;
    if (step.type === 'atx') total += step.ms || DEFAULT_ATX_MS;
  }
  return total;
}

// The modifier byte's bits, in the order a label lists them.
const modifierLabels: [number, string][] = [
  [ModifierBits.LeftCtrl, 'Ctrl'],
  [ModifierBits.RightCtrl, 'RCtrl'],
  [ModifierBits.LeftShift, 'Shift'],
  [ModifierBits.RightShift, 'RShift'],
  [ModifierBits.LeftAlt, 'Alt'],
  [ModifierBits.RightAlt, 'AltGr'],
  [ModifierBits.LeftMeta, 'Meta'],
  [ModifierBits.RightMeta, 'RMeta']
];

const shortNames: Record<string, string> = {
  Escape: 'Esc',
  Delete: 'Del',
  Insert: 'Ins',
  Backspace: '⌫',
  Enter: '↵',
  ArrowUp: '↑',
  ArrowDown: '↓',
  ArrowLeft: '←',
  ArrowRight: '→',
  PageUp: 'PgUp',
  PageDown: 'PgDn'
};

// The first code the key map lists for each usage. The map has aliases, and
// the first is the plain name.
const codeByUsage = new Map<number, string>();
for (const [code, usage] of Object.entries(KeycodeMap)) {
  if (!codeByUsage.has(usage)) codeByUsage.set(usage, code);
}

// keyLabel names a usage the way a key cap does.
export function keyLabel(usage: number): string {
  const code = codeByUsage.get(usage);
  if (!code) return `0x${usage.toString(16).padStart(2, '0')}`;
  if (shortNames[code]) return shortNames[code];
  if (/^Key[A-Z]$/.test(code)) return code.slice(3);
  if (/^Digit[0-9]$/.test(code)) return code.slice(5);
  return code;
}

// comboLabel names a keyboard state: "Ctrl+Alt+Del", or "" when nothing is
// held.
export function comboLabel(modifiers = 0, keys: number[] = []): string {
  const parts = modifierLabels.filter(([bit]) => modifiers & bit).map(([, label]) => label);
  return [...parts, ...keys.map(keyLabel)].join('+');
}

// usageOf returns the usage of a key code (KeyboardEvent.code), or undefined.
export function usageOf(code: string): number | undefined {
  return KeycodeMap[code];
}

// The auto-tap's settings, as the keyboard menu keeps them between visits.
export type AutoTapSettings = {
  code: string; // KeyboardEvent.code of the key
  interval: number; // ms
  timeout: number; // s
  firstPress: '' | AtxButton;
  stopOnInput: boolean;
  stopOnVideo: boolean;
};

// The keys firmware menus use: setup on F2 or Del, the boot menu on F12, F11,
// F10 or Esc.
export const AUTO_TAP_PRESETS = ['F2', 'Delete', 'F12', 'Escape', 'F11', 'F10'];

// The server's limits, server/service/sequence/autotap.go.
export const AUTO_TAP_INTERVAL = { min: 50, max: 5000 };
export const AUTO_TAP_TIMEOUT = { min: 1, max: 600 };

export const DEFAULT_AUTO_TAP: AutoTapSettings = {
  code: 'F2',
  interval: 150,
  timeout: 60,
  firstPress: '',
  stopOnInput: true,
  stopOnVideo: false
};

// The keys the "other key" list offers.
export const AUTO_TAP_KEYS: string[] = [
  ...Array.from({ length: 12 }, (_, i) => `F${i + 1}`),
  'Escape',
  'Delete',
  'Enter',
  'Space',
  'Tab',
  'Backspace',
  'Insert',
  'Home',
  'End',
  'PageUp',
  'PageDown',
  'ArrowUp',
  'ArrowDown',
  'ArrowLeft',
  'ArrowRight',
  ...Array.from({ length: 26 }, (_, i) => `Key${String.fromCharCode(65 + i)}`),
  ...Array.from({ length: 10 }, (_, i) => `Digit${i}`)
].filter((code) => usageOf(code) !== undefined);

function clampInt(value: unknown, min: number, max: number, fallback: number): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) return fallback;
  return Math.min(max, Math.max(min, Math.round(value)));
}

// autoTapSettingsFrom reads stored settings, keeping each field only when it
// is one the server takes.
export function autoTapSettingsFrom(raw: unknown): AutoTapSettings {
  const d = DEFAULT_AUTO_TAP;
  if (raw === null || typeof raw !== 'object') return { ...d };
  const r = raw as Record<string, unknown>;
  return {
    code: typeof r.code === 'string' && AUTO_TAP_KEYS.includes(r.code) ? r.code : d.code,
    interval: clampInt(r.interval, AUTO_TAP_INTERVAL.min, AUTO_TAP_INTERVAL.max, d.interval),
    timeout: clampInt(r.timeout, AUTO_TAP_TIMEOUT.min, AUTO_TAP_TIMEOUT.max, d.timeout),
    firstPress: r.firstPress === 'power' || r.firstPress === 'reset' ? r.firstPress : '',
    stopOnInput: typeof r.stopOnInput === 'boolean' ? r.stopOnInput : d.stopOnInput,
    stopOnVideo: typeof r.stopOnVideo === 'boolean' ? r.stopOnVideo : d.stopOnVideo
  };
}
