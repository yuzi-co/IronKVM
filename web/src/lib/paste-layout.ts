// Layouts the server can type on, in the order the menu lists them.
export const pasteLayoutIds = [
  'us',
  'uk',
  'de',
  'fr',
  'es',
  'it',
  'pt-br',
  'se',
  'ru',
  'ja',
  'ko',
  'am',
  'ar',
  'bepo',
  'cz',
  'da',
  'de-ch',
  'en-us-altgr-intl',
  'en-us-colemak',
  'et',
  'fo',
  'fr-be',
  'fr-ca',
  'fr-ch',
  'hr',
  'hu',
  'is',
  'lt',
  'lv',
  'mk',
  'nl',
  'no',
  'pl',
  'pt',
  'sl',
  'th',
  'tr'
] as const;

// The keyboard layout the owner set for the host on the device. Configured is
// false while none is set and the server types US.
export type HostLayout = { layout: string; configured: boolean };

function isPasteLayout(id: string | null | undefined): id is string {
  return !!id && (pasteLayoutIds as readonly string[]).includes(id);
}

// The translation key of a layout ID: 'pt-br' is 'ptBr'.
export function layoutLabelKey(id: string) {
  return id.replace(/-(\w)/g, (_, c: string) => c.toUpperCase());
}

// The host's layout is not the browser's, but they agree more often than not,
// so the browser's language is the guess when nothing better is known.
export function guessPasteLayout(browserLanguage: string) {
  const language = browserLanguage.toLowerCase();
  if (language === 'en-gb' || language === 'en-ie') return 'uk';
  if (language === 'pt-br') return 'pt-br';
  if (['de-ch', 'fr-be', 'fr-ca', 'fr-ch'].includes(language)) return language;
  if (language.startsWith('sv') || language.startsWith('fi')) return 'se';

  // A few layout IDs are not the language's own code.
  const primary = language.split('-')[0];
  const byLanguage: Record<string, string> = { cs: 'cz', hy: 'am', nb: 'no', nn: 'no' };
  const guess = byLanguage[primary] ?? primary;
  return isPasteLayout(guess) ? guess : 'us';
}

// initialPasteLayout is the layout the paste dialog starts on: the one this
// browser picked, else the one the owner set on the device, else the guess.
export function initialPasteLayout(stored: string | null, host: HostLayout | null, guess: string) {
  if (isPasteLayout(stored)) return stored;
  if (host?.configured && isPasteLayout(host.layout)) return host.layout;
  return guess;
}
