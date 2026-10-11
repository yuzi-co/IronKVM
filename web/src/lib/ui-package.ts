// Decisions for the web UI section of Settings > Update. No imports, so it runs
// under node --test.

// The name server/service/webui accepts for a package.
const PACKAGE_NAME = /^ironkvm-ui_\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?\.tar\.gz$/;
const SIGNATURE_SUFFIX = '.minisig';

type NamedFile = { name: string };

export type UploadPick<F extends NamedFile> =
  | { ok: true; pkg: F; signature: F | null }
  | { ok: false; error: 'noPackage' | 'tooMany' | 'badName' | 'signatureMismatch' };

// pickUpload sorts the files chosen in one dialog into the package and its
// signature. The signature is optional here because the board may have the
// check turned off; the server decides whether it is needed.
export function pickUpload<F extends NamedFile>(files: readonly F[]): UploadPick<F> {
  if (files.length === 0) return { ok: false, error: 'noPackage' };
  if (files.length > 2) return { ok: false, error: 'tooMany' };

  const signatures = files.filter((f) => f.name.endsWith(SIGNATURE_SUFFIX));
  const packages = files.filter((f) => !f.name.endsWith(SIGNATURE_SUFFIX));
  if (packages.length !== 1 || signatures.length > 1) {
    return { ok: false, error: packages.length === 0 ? 'noPackage' : 'tooMany' };
  }

  const pkg = packages[0];
  if (!PACKAGE_NAME.test(pkg.name)) return { ok: false, error: 'badName' };

  const signature = signatures[0] ?? null;
  // A browser may add " (1)" to a second download of the same name, so the
  // signature has to name the package rather than merely end in .minisig.
  if (signature && !signature.name.startsWith(pkg.name)) {
    return { ok: false, error: 'signatureMismatch' };
  }

  return { ok: true, pkg, signature };
}

export type UiOffer =
  | { kind: 'none' } // nothing in the feed, or not asked yet
  | { kind: 'current'; version: string } // the feed's UI is the one served
  | { kind: 'incompatible'; version: string; api: { min: number; max: number } }
  | { kind: 'available'; version: string };

// uiOffer turns GET /api/ui/latest's answer into what the page offers.
export function uiOffer(
  latest: {
    version: string;
    compatible: boolean;
    active: boolean;
    api: { min: number; max: number };
  } | null
): UiOffer {
  if (!latest || !latest.version) return { kind: 'none' };
  if (latest.active) return { kind: 'current', version: latest.version };
  if (!latest.compatible) return { kind: 'incompatible', version: latest.version, api: latest.api };
  return { kind: 'available', version: latest.version };
}
