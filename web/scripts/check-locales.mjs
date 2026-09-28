// Checks that every locale in src/i18n/locales carries the same keys as en.ts.
//
// en.ts is the source of truth. For each other locale the script reports:
//   - missing keys: in en.ts, absent from the locale (the app falls back to English);
//   - stale keys: in the locale, absent from en.ts (nothing can show them);
//   - unknown placeholders: a {{name}} in a translation that the English string does not have.
//
// Plural keys follow i18next: a key with an _other form in en.ts is a plural base, and each
// locale must carry the forms its language uses for integer counts, as Intl.PluralRules reports
// them, plus _other. Russian, for example, needs _one, _few, _many and _other; Japanese needs
// only _other.
//
// Run with `node --experimental-strip-types scripts/check-locales.mjs`. The locale files are
// plain objects, so Node can import them once it strips the (absent) types.

import { readdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const localesDir = join(dirname(fileURLToPath(import.meta.url)), '..', 'src', 'i18n', 'locales');

// File names that are not BCP 47 tags for the language they hold.
const intlLocale = { cz: 'cs', se: 'sv', pt_br: 'pt-BR', zh_tw: 'zh-TW' };

const pluralSuffixes = ['zero', 'one', 'two', 'few', 'many', 'other'];
const pluralRe = new RegExp(`^(.*)_(${pluralSuffixes.join('|')})$`);
const placeholderRe = /\{\{\s*([^}\s,]+)[^}]*\}\}/g;

function flatten(obj, prefix = '', out = new Map()) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v !== null && typeof v === 'object') flatten(v, key, out);
    else out.set(key, String(v));
  }
  return out;
}

function placeholders(s) {
  return new Set([...s.matchAll(placeholderRe)].map((m) => m[1]));
}

// The plural categories a language uses for whole-number counts, plus 'other', which i18next
// needs as the final fallback.
function pluralCategories(file) {
  const rules = new Intl.PluralRules(intlLocale[file] ?? file);
  const cats = new Set(['other']);
  for (let n = 0; n <= 1000; n++) cats.add(rules.select(n));
  return pluralSuffixes.filter((c) => cats.has(c));
}

async function load(file) {
  const mod = await import(pathToFileURL(join(localesDir, `${file}.ts`)).href);
  return flatten(mod.default);
}

const files = readdirSync(localesDir)
  .filter((f) => f.endsWith('.ts'))
  .map((f) => f.slice(0, -3))
  .sort();

const en = await load('en');

// Split en.ts into plain keys and plural bases.
const plainKeys = [];
const pluralBases = new Map(); // base -> placeholders across the English forms
for (const [key, value] of en) {
  const m = key.match(pluralRe);
  if (m && en.has(`${m[1]}_other`)) {
    const ph = pluralBases.get(m[1]) ?? new Set();
    for (const p of placeholders(value)) ph.add(p);
    pluralBases.set(m[1], ph);
  } else {
    plainKeys.push(key);
  }
}

let failed = false;
let totalMissing = 0;
let totalStale = 0;
let totalPlaceholder = 0;

for (const file of files) {
  if (file === 'en') continue;
  const loc = await load(file);

  // Expected key -> the placeholders the English string allows.
  const expected = new Map();
  for (const key of plainKeys) expected.set(key, placeholders(en.get(key)));
  const cats = pluralCategories(file);
  for (const [base, ph] of pluralBases) {
    for (const c of cats) expected.set(`${base}_${c}`, ph);
  }

  const missing = [...expected.keys()].filter((k) => !loc.has(k));
  const stale = [...loc.keys()].filter((k) => !expected.has(k));
  const badPlaceholders = [];
  for (const [key, value] of loc) {
    const allowed = expected.get(key);
    if (!allowed) continue;
    const unknown = [...placeholders(value)].filter((p) => !allowed.has(p));
    if (unknown.length) badPlaceholders.push(`${key} ({{${unknown.join('}}, {{')}}})`);
  }

  if (missing.length || stale.length || badPlaceholders.length) {
    failed = true;
    console.log(`\n${file}.ts:`);
    if (missing.length)
      console.log(`  missing (${missing.length}):\n    ${missing.join('\n    ')}`);
    if (stale.length) console.log(`  stale (${stale.length}):\n    ${stale.join('\n    ')}`);
    if (badPlaceholders.length) {
      console.log(
        `  unknown placeholders (${badPlaceholders.length}):\n    ${badPlaceholders.join('\n    ')}`
      );
    }
  }
  totalMissing += missing.length;
  totalStale += stale.length;
  totalPlaceholder += badPlaceholders.length;
}

if (failed) {
  console.log(
    `\ncheck-locales: ${totalMissing} missing, ${totalStale} stale, ` +
      `${totalPlaceholder} with unknown placeholders, across ${files.length - 1} locales`
  );
  process.exit(1);
}
console.log(`check-locales: ${files.length - 1} locales match en.ts (${en.size} keys)`);
