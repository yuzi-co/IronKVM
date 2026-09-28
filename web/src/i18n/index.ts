import i18n from 'i18next';
import type { Resource } from 'i18next';
import { initReactI18next } from 'react-i18next';

import { getLanguage } from '@/lib/localstorage.ts';

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

function getResources(): Resource {
  const resources: Resource = {};

  const modules: Record<string, Resource> = import.meta.glob('./locales/*.ts', { eager: true });

  for (const path in modules) {
    const moduleName = path.split('/').pop()?.replace('.ts', '');
    if (moduleName) {
      resources[localeTag(moduleName)] = modules[path].default;
    }
  }

  return resources;
}

function getCurrentLanguage(): string {
  const languages = Object.keys(resources);

  const cookieLng = getLanguage();
  if (cookieLng && languages.includes(localeTag(cookieLng))) {
    return localeTag(cookieLng);
  }

  // A regional tag such as pt-BR or zh-TW first, then its language alone.
  const navigatorLng = languages.find(
    (language) => language.toLowerCase() === navigator.language.toLowerCase()
  );
  if (navigatorLng) {
    return navigatorLng;
  }

  const baseLng = navigator.language.split('-')[0];
  if (languages.includes(baseLng)) {
    return baseLng;
  }

  return 'en';
}

const resources = getResources();
const lng = getCurrentLanguage();

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false
    }
  })
  .then();

export default i18n;
