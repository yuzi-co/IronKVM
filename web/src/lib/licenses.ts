// The image's license manifest, as GET /api/licenses lists it: the files that
// ironkvm-dist's licenses/make-manifest.sh writes into
// /usr/share/licenses/ironkvm.

export type LicenseFile = {
  // Relative to the manifest directory, with forward slashes.
  path: string;
  size: number;
};

export type LicenseText = LicenseFile & {
  // The license id, the file name under texts/.
  name: string;
};

export type LicenseGroups = {
  // NOTICE, COMPONENTS and MANIFEST in that order, then any other top-level
  // file the manifest gained since.
  summary: LicenseFile[];
  // One file per license id, sorted.
  texts: LicenseText[];
};

const summaryOrder = ['NOTICE', 'COMPONENTS', 'MANIFEST'];
const textsDir = 'texts/';

function rank(path: string): number {
  const index = summaryOrder.indexOf(path);
  return index === -1 ? summaryOrder.length : index;
}

export function groupLicenseFiles(files: LicenseFile[] | undefined): LicenseGroups {
  const summary: LicenseFile[] = [];
  const texts: LicenseText[] = [];

  for (const file of files ?? []) {
    if (file.path.startsWith(textsDir)) {
      texts.push({ ...file, name: file.path.slice(textsDir.length) });
    } else if (!file.path.includes('/')) {
      summary.push(file);
    }
  }

  summary.sort((a, b) => rank(a.path) - rank(b.path) || a.path.localeCompare(b.path));
  texts.sort((a, b) => a.name.localeCompare(b.name));
  return { summary, texts };
}

// licenseFileUrl is the link that opens one file as plain text in a new tab.
// The session cookie goes with it, as with any request to the device.
export function licenseFileUrl(baseUrl: string, path: string): string {
  return `${baseUrl}/api/licenses/file?path=${encodeURIComponent(path)}`;
}
