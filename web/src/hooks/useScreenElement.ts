import { useEffect, useState } from 'react';

// useScreenElement is the element that shows the host screen (#screen), and
// follows it as it comes and goes. The mouse handlers attach to it, and they
// mount before it does: the desktop holds the stream back until the ion check
// answers, so when the screen settings arrive first, the handlers' last setup
// found no #screen and never ran again, and the mouse did nothing until the
// resolution changed. "Pause when tab is hidden" also unmounts the screen and
// mounts a new element, which left the handlers on the old one.
export function useScreenElement(): HTMLElement | null {
  const [element, setElement] = useState<HTMLElement | null>(() =>
    document.getElementById('screen')
  );

  useEffect(() => {
    // A setter called with the same element does not re-render.
    const update = () => setElement(document.getElementById('screen'));
    update();

    const observer = new MutationObserver(update);
    observer.observe(document.body, { childList: true, subtree: true });
    return () => observer.disconnect();
  }, []);

  return element;
}
