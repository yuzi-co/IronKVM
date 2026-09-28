const languages = [
  { key: 'ca', name: 'Català' },
  { key: 'nl', name: 'Nederlands' },
  { key: 'da', name: 'Danish' },
  { key: 'de', name: 'Deutsch' },
  { key: 'en', name: 'English' },
  { key: 'es', name: 'Español' },
  { key: 'fr', name: 'Français' },
  { key: 'id', name: 'Indonesia' },
  { key: 'it', name: 'Italian' },
  { key: 'pl', name: 'Polski' },
  { key: 'pt-BR', name: 'Português (Brasil)' },
  { key: 'ru', name: 'Русский' },
  { key: 'tr', name: 'Türkçe' },
  { key: 'ko', name: '한국어' },
  { key: 'zh', name: '简体中文' },
  { key: 'zh-TW', name: '繁體中文' },
  { key: 'hu', name: 'Magyar' },
  { key: 'vi', name: 'Tiếng Việt' },
  { key: 'ja', name: '日本語' },
  { key: 'cs', name: 'Česky' },
  { key: 'uk', name: 'Українська' },
  { key: 'nb', name: 'Norsk, bokmål' },
  { key: 'th', name: 'ภาษาไทย' },
  { key: 'sv', name: 'Svenska' }
];

languages.sort((a, b) => a.name.localeCompare(b.name, 'en', { sensitivity: 'base' }));

export default languages;
