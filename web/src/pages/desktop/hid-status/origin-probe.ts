// What POST /api/origin-check said, reduced to what the notice needs.
//
// status is the HTTP status, or 0 when the request failed before any answer
// (a network error). data is the envelope's data of a 403: the Origin the
// device saw and the Host it saw (ironkvm-dist#178).
export type OriginProbe = {
  status: number;
  data?: { origin?: unknown; host?: unknown } | null;
};

export type OriginMismatch = { origin: string; host: string };

// originMismatch returns the two names when the device refused the page's
// origin, and null for any other answer, so the caller keeps its own choice.
export function originMismatch(probe: OriginProbe | null): OriginMismatch | null {
  if (!probe || probe.status !== 403) return null;
  const { origin, host } = probe.data ?? {};
  if (typeof origin !== 'string' || typeof host !== 'string' || !origin || !host) return null;
  return { origin, host };
}

// Reads the probe out of a failed axios request. The envelope of the 403 is
// error.response.data, and its data field carries the two names.
export function probeFromError(error: unknown): OriginProbe {
  const response = (error as { response?: { status?: number; data?: { data?: unknown } } } | null)
    ?.response;
  if (!response) return { status: 0 };
  const data = response.data?.data;
  return {
    status: response.status ?? 0,
    data: data && typeof data === 'object' ? (data as OriginProbe['data']) : null
  };
}
