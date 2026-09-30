import i18n from 'i18next';
import type { BackendModule, ResourceLanguage } from 'i18next';
import { initReactI18next } from 'react-i18next';

import { getLanguage } from '@/lib/localstorage.ts';

import languages from './languages.ts';
import en from './locales/en.ts';

// Four locale files are named with codes that are not language tags: cz and se
// are the country codes of Czechia and Sweden, and pt_br and zh_tw use an
// underscore. i18next picks plural rules by the language it is given, so under
// those names Czech got the browser's rules and Swedish got Northern Sami's.
// The resources are registered under the real tags, and a choice saved under an
// old name still loads.
const localeTags: Record<string, string> = {
  cz: 'cs',
  se: 'sv',
  pt_br: 'pt-BR',
  zh_tw: 'zh-TW'
};

export function localeTag(name: string): string {
  return localeTags[name] ?? name;
}

// English is bundled as the fallback. Every other locale is its own chunk, fetched
// only when that language is chosen, so a page load carries one language, not 24.
function getLoaders(): Record<string, () => Promise<ResourceLanguage>> {
  const loaders: Record<string, () => Promise<ResourceLanguage>> = {};

  const modules = import.meta.glob<ResourceLanguage>(['./locales/*.ts', '!./locales/en.ts'], {
    import: 'default'
  });

  for (const path in modules) {
    const moduleName = path.split('/').pop()?.replace('.ts', '');
    if (moduleName) {
      loaders[localeTag(moduleName)] = modules[path];
    }
  }

  return loaders;
}

const loaders = getLoaders();

// A backend in the shape of i18next-resources-to-backend: i18next asks it for a
// language it has no bundle for, and changeLanguage resolves once the bundle is in.
// Tags with no locale file, such as the bare "pt" i18next also tries for pt-BR,
// resolve empty so they fall through to English quietly.
const lazyLocales: BackendModule = {
  type: 'backend',
  init() {},
  read(language, namespace, callback) {
    const load = loaders[language];
    if (!load) {
      callback(null, {});
      return;
    }
    load()
      .then((resources) => callback(null, resources[namespace] ?? {}))
      .catch((err: Error) => callback(err, false));
  }
};

function getCurrentLanguage(): string {
  const tags = languages.map((language) => language.key);

  const cookieLng = getLanguage();
  if (cookieLng && tags.includes(localeTag(cookieLng))) {
    return localeTag(cookieLng);
  }

  // A regional tag such as pt-BR or zh-TW first, then its language alone.
  const navigatorLng = tags.find((tag) => tag.toLowerCase() === navigator.language.toLowerCase());
  if (navigatorLng) {
    return navigatorLng;
  }

  const baseLng = navigator.language.split('-')[0];
  if (tags.includes(baseLng)) {
    return baseLng;
  }

  return 'en';
}

// Resolves once the chosen language is loaded; main.tsx waits on it before the
// first render so no raw keys flash on screen.
export const i18nReady = i18n
  .use(lazyLocales)
  .use(initReactI18next)
  .init({
    resources: { en },
    partialBundledLanguages: true,
    lng: getCurrentLanguage(),
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false
    }
  });

export default i18n;
