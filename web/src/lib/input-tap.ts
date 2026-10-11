// A tap on the input this tab sends to the host, for the macro recorder. The
// websocket reports every keyboard and pointer message it sends, and the
// power menu every ATX press. With nobody listening a report costs one size
// check.
export type InputTapEvent =
  | { kind: 'message'; data: ArrayLike<number> }
  | { kind: 'atx'; button: 'power' | 'reset'; ms: number };

type Listener = (event: InputTapEvent) => void;

const listeners = new Set<Listener>();

export function tapInput(event: InputTapEvent) {
  if (listeners.size === 0) return;
  listeners.forEach((listener) => listener(event));
}

// onInputTap listens and returns a function that stops listening.
export function onInputTap(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

// isTapping says whether anything listens, so a caller can skip building an
// event nobody reads.
export function isTapping(): boolean {
  return listeners.size > 0;
}
