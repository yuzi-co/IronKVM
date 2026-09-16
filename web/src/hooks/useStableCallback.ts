import { useCallback, useLayoutEffect, useRef } from 'react';

// useStableCallback returns a function whose identity never changes and which
// always runs the latest version of fn. It stands in for useEffectEvent, which
// React 18 does not have.
//
// It exists for the effect that calls a component function: listing the
// function as a dependency re-runs the effect on every render, and leaving it
// out makes the effect call the copy from the first render. A stable wrapper can
// be listed, so the effect runs when its real dependencies change and nothing
// else.
//
// Call the result from effects and event handlers, never during render. The ref
// is brought up to date after each commit.
export function useStableCallback<Args extends unknown[], Result>(
  fn: (...args: Args) => Result
): (...args: Args) => Result {
  const ref = useRef(fn);

  useLayoutEffect(() => {
    ref.current = fn;
  });

  return useCallback((...args: Args) => ref.current(...args), []);
}
