// Version comparison for the few places that order two version strings.
//
// This replaces the semver package, which was bundled for three calls and
// cost 25 KB. It follows semver's strict parser and its ordering exactly, so
// the results, errors included, are the ones semver.gt, gte and valid gave:
//
//   - an optional leading "v", and surrounding whitespace, are accepted;
//   - a prerelease sorts below its release (1.2.3-rc1 < 1.2.3), and its
//     identifiers compare numerically when both are numbers, a number sorts
//     below a word, and words compare as strings;
//   - build metadata is ignored, which is where the server puts its build
//     stamp (2.4.3+dev.20260930.1023.7ff5405f equals 2.4.3);
//   - anything else is not a version: compareVersions throws a TypeError for
//     it, as semver did, and isValidVersion answers false.
//
// version.test.ts checks every case against the semver package itself.

const MAX_LENGTH = 256;

const NUMERIC = '0|[1-9]\\d*';
const PRERELEASE_ID = `(?:${NUMERIC}|\\d*[a-zA-Z-][a-zA-Z0-9-]*)`;
const BUILD_ID = '[a-zA-Z0-9-]+';
const FULL = new RegExp(
  `^v?(${NUMERIC})\\.(${NUMERIC})\\.(${NUMERIC})` +
    `(?:-(${PRERELEASE_ID}(?:\\.${PRERELEASE_ID})*))?` +
    `(?:\\+(${BUILD_ID}(?:\\.${BUILD_ID})*))?$`
);

type Version = {
  main: [number, number, number];
  prerelease: string[];
};

function parse(version: string): Version {
  if (typeof version !== 'string') {
    throw new TypeError(`Invalid version. Must be a string. Got type "${typeof version}".`);
  }
  if (version.length > MAX_LENGTH) {
    throw new TypeError(`version is longer than ${MAX_LENGTH} characters`);
  }

  const m = version.trim().match(FULL);
  if (!m) {
    throw new TypeError(`Invalid Version: ${version}`);
  }

  const main = [+m[1], +m[2], +m[3]] as [number, number, number];
  if (main.some((part) => part > Number.MAX_SAFE_INTEGER)) {
    throw new TypeError('Invalid version number');
  }

  return { main, prerelease: m[4] ? m[4].split('.') : [] };
}

const numeric = /^[0-9]+$/;

function compareIdentifiers(a: string, b: string) {
  const aNum = numeric.test(a);
  const bNum = numeric.test(b);
  if (aNum && bNum) {
    const x = +a;
    const y = +b;
    return x === y ? 0 : x < y ? -1 : 1;
  }
  if (a === b) return 0;
  if (aNum) return -1;
  if (bNum) return 1;
  return a < b ? -1 : 1;
}

function comparePrerelease(a: string[], b: string[]) {
  if (a.length && !b.length) return -1;
  if (!a.length && b.length) return 1;

  for (let i = 0; ; i++) {
    if (i >= a.length && i >= b.length) return 0;
    if (i >= b.length) return 1;
    if (i >= a.length) return -1;
    const order = compareIdentifiers(a[i], b[i]);
    if (order !== 0) return order;
  }
}

// compareVersions answers -1, 0 or 1 as a sorts before, with or after b. It
// throws a TypeError when either is not a version.
export function compareVersions(a: string, b: string): -1 | 0 | 1 {
  const x = parse(a);
  const y = parse(b);

  for (let i = 0; i < 3; i++) {
    if (x.main[i] !== y.main[i]) return x.main[i] < y.main[i] ? -1 : 1;
  }

  return comparePrerelease(x.prerelease, y.prerelease) as -1 | 0 | 1;
}

export function isValidVersion(version: string): boolean {
  try {
    parse(version);
    return true;
  } catch {
    return false;
  }
}

export const versionGt = (a: string, b: string) => compareVersions(a, b) > 0;
export const versionGte = (a: string, b: string) => compareVersions(a, b) >= 0;
