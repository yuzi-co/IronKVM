// The UI's half of the server contract (GET /api/version, answered by
// server/router/version.go from server/common/apicontract).
//
// This module runs under node --test as well as in the browser, so it imports
// nothing but the range below.

// The server API versions this UI works with, both ends included, from
// ui-api-range.json. Raise max when the UI has been checked against a server
// that bumped its API version; raise min when the UI stops supporting an old
// one. The file is JSON so that vite-plugin-ui-version.ts can read it without
// importing the app's sources, and write it into ui-version.json and from
// there into a UI package's manifest.
import range from './ui-api-range.json' with { type: 'json' };

export const UI_API_RANGE: { readonly min: number; readonly max: number } = range;

// What a server that predates GET /api/version speaks.
export const LEGACY_API_VERSION = 1;

// The server features this UI gates on, one line each, by the name the server
// registers (server/common/apicontract/contract.go). Keep the list sorted.
export const SERVER_FEATURES = [
  'direct-clock', // the clock request on the direct H.264 socket (status bar latency)
  'ip-allowlist', // /api/network/allowlist (Settings > Network > Allowed networks)
  'status-bar', // GET /api/stream/info (status bar source line)
  'ui-update' // /api/ui: separate UI packages (Settings > Update)
] as const;

export type ServerFeature = (typeof SERVER_FEATURES)[number];

// The features a server could have before it answered the contract. Only
// these are asked for on a server that does not answer it; a feature added
// since then is absent from every such server by definition.
const PRE_CONTRACT_FEATURES: ReadonlySet<ServerFeature> = new Set(['direct-clock', 'status-bar']);

export type ServerContract = {
  // False for a server that does not answer GET /api/version: an older server,
  // or a request that failed. Its features are then not known.
  known: boolean;
  version: string;
  api: number;
  features: ReadonlySet<string>;
};

export const LEGACY_CONTRACT: ServerContract = {
  known: false,
  version: '',
  api: LEGACY_API_VERSION,
  features: new Set()
};

// contractFromResponse reads the data field of GET /api/version's envelope. An
// answer that does not have the expected shape is read as a legacy server,
// never as one that has no features.
export function contractFromResponse(data: unknown): ServerContract {
  if (data === null || typeof data !== 'object') return LEGACY_CONTRACT;
  const d = data as { version?: unknown; api?: unknown; features?: unknown };
  if (typeof d.api !== 'number' || !Number.isInteger(d.api) || d.api < 1) return LEGACY_CONTRACT;
  const features = Array.isArray(d.features)
    ? d.features.filter((f): f is string => typeof f === 'string')
    : [];
  return {
    known: true,
    version: typeof d.version === 'string' ? d.version : '',
    api: d.api,
    features: new Set(features)
  };
}

// hasFeature decides whether the UI may use a feature. A server that declares
// its features is taken at its word. A server that does not is asked anyway
// for a feature older than the contract, and the caller copes with a 404 as it
// did before the contract existed: the alternative would hide working
// features on every board that has not yet been updated.
export function hasFeature(contract: ServerContract, name: ServerFeature): boolean {
  return contract.known ? contract.features.has(name) : PRE_CONTRACT_FEATURES.has(name);
}

// apiSupported reports whether this UI supports the server's API version.
export function apiSupported(
  contract: ServerContract,
  range: { min: number; max: number } = UI_API_RANGE
): boolean {
  return contract.api >= range.min && contract.api <= range.max;
}

// formatRange is the range as the notice shows it.
export function formatRange(range: { min: number; max: number } = UI_API_RANGE): string {
  return range.min === range.max ? String(range.min) : `${range.min}–${range.max}`;
}
