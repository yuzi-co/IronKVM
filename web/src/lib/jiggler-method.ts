// The jiggler either moves the mouse or presses one harmless key. The server
// keeps that as a method and a key; the menu offers it as one choice.
export const jigglerChoices = ['mouse', 'f15', 'shift', 'ctrl'] as const;

export type JigglerChoice = (typeof jigglerChoices)[number];

export type JigglerMethod = { method: 'mouse' | 'key'; key: string };

// jigglerChoiceFrom reads the server's answer. A server that predates the key
// method sends neither field, and one that does not know the key is treated
// as the mouse, which is what it would do.
export function jigglerChoiceFrom(method?: string, key?: string): JigglerChoice {
  if (method !== 'key') return 'mouse';
  if (!key) return 'f15';
  return (jigglerChoices as readonly string[]).includes(key) && key !== 'mouse'
    ? (key as JigglerChoice)
    : 'mouse';
}

export function jigglerMethodOf(choice: JigglerChoice): JigglerMethod {
  return choice === 'mouse' ? { method: 'mouse', key: '' } : { method: 'key', key: choice };
}
