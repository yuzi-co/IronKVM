#!/bin/bash
#
# Assemble a web UI package: ironkvm-ui_<version>.tar.gz, its minisign
# signature and ui-latest.json, from a built web/dist.
#
#   UI_VERSION=<version> pnpm -C web build
#   scripts/package-ui.sh <version>
#
# The format, the trust chain and the installer are described in
# docs/specs/2026-10-11-ui-packages.md; server/service/webui is the reader. In short:
#
#   ironkvm-ui_<version>.tar.gz          manifest.json + web/ (the built UI)
#   ironkvm-ui_<version>.tar.gz.minisig  minisign signature over the tarball
#   ui-latest.json                       the feed entry, beside latest.json
#
# Environment:
#   SIGN_KEY            minisign secret key. Without it the package is not
#                       signed, and a board with a keyring refuses it.
#   SIGN_PASSWORD_FILE  a file holding the key's password on one line, read on
#                       stdin as abslots/sign-image.sh does. Without it the key
#                       must have no password.
#   VERIFY_KEYS         public keys (space separated) that boards trust, for
#                       example the release key images ship in
#                       /usr/share/ironkvm/keys. When set, one of them must
#                       verify the new signature or the signature is removed
#                       and the script fails: a package no board can verify is
#                       worse than an unsigned one, which says what it is.
#   UI_PACKAGE_URL      the absolute https URL the package will be served from,
#                       written into ui-latest.json. Without it the board looks
#                       for the package beside ui-latest.json.
#   OUT                 output directory, build/ui by default.
#   MINISIGN            the minisign binary.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"
DIST="${DIST:-$ROOT/web/dist}"
OUT="${OUT:-$ROOT/build/ui}"
MINISIGN="${MINISIGN:-minisign}"

die() { echo "[ERROR] $*" >&2; exit 1; }

[ -n "$VERSION" ] || die "usage: $0 <version>"
# The same pattern the server applies (server/service/webui/package.go).
echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' \
    || die "invalid version '$VERSION', expected semver such as 1.2.0 or 1.2.0+dev.1"

[ -f "$DIST/index.html" ] || die "no build in $DIST: run UI_VERSION=$VERSION pnpm -C web build"
[ -f "$DIST/ui-version.json" ] || die "$DIST/ui-version.json is missing: the build predates the UI version stamp"
built=$(node -e 'console.log(JSON.parse(require("fs").readFileSync(process.argv[1],"utf8")).version)' "$DIST/ui-version.json")
[ "$built" = "$VERSION" ] || die "web/dist was built as '$built', not '$VERSION': rebuild with UI_VERSION=$VERSION"

NAME="ironkvm-ui_$VERSION.tar.gz"
PKG="$OUT/$NAME"
STAGE="$OUT/stage"

rm -rf "$OUT"
mkdir -p "$STAGE"
cp -R "$DIST" "$STAGE/web"
node "$ROOT/web/scripts/ui-manifest.mjs" "$STAGE/web" > "$STAGE/manifest.json"

# Reproducible: sorted names, no owners, a fixed time, and gzip without the
# file name and time it would otherwise record.
EPOCH="${SOURCE_DATE_EPOCH:-0}"
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$EPOCH" \
    --mode='u+rwX,go+rX,go-w' -C "$STAGE" -cf - manifest.json web \
    | gzip -n -9 > "$PKG"
rm -rf "$STAGE"

if [ -n "${SIGN_KEY:-}" ]; then
    command -v "$MINISIGN" > /dev/null 2>&1 || die "SIGN_KEY is set and $MINISIGN is not on PATH"
    [ -f "$SIGN_KEY" ] || die "no secret key at $SIGN_KEY"
    if [ -n "${SIGN_PASSWORD_FILE:-}" ]; then
        "$MINISIGN" -S -s "$SIGN_KEY" -m "$PKG" -t "ironkvm-ui $VERSION" < "$SIGN_PASSWORD_FILE" > /dev/null
    else
        "$MINISIGN" -S -s "$SIGN_KEY" -m "$PKG" -t "ironkvm-ui $VERSION" < /dev/null > /dev/null
    fi
    echo "[INFO] signed $NAME.minisig"

    if [ -n "${VERIFY_KEYS:-}" ]; then
        verified=
        for k in $VERIFY_KEYS; do
            if "$MINISIGN" -V -q -p "$k" -m "$PKG" > /dev/null 2>&1; then
                verified=$k
                break
            fi
        done
        if [ -z "$verified" ]; then
            rm -f "$PKG.minisig"
            die "no key in VERIFY_KEYS verifies the signature, so no board would accept it; the signature was removed"
        fi
        echo "[INFO] $(basename "$verified") verifies it"
    fi
else
    echo "[WARN] SIGN_KEY is not set: $NAME is unsigned, and a board with a keyring refuses it"
fi

size=$(wc -c < "$PKG" | tr -d ' ')
sha512=$(openssl dgst -sha512 -binary "$PKG" | openssl base64 -A)
node -e '
const [version, name, size, sha512, url, stampFile, out] = process.argv.slice(1);
const stamp = JSON.parse(require("fs").readFileSync(stampFile, "utf8"));
const doc = { format: 1, version, name };
if (url) doc.url = url;
Object.assign(doc, { size_bytes: Number(size), sha512, api: stamp.api });
require("fs").writeFileSync(out, JSON.stringify(doc, null, 2) + "\n");
' "$VERSION" "$NAME" "$size" "$sha512" "${UI_PACKAGE_URL:-}" "$DIST/ui-version.json" "$OUT/ui-latest.json"

echo "[INFO] $PKG ($size bytes)"
echo "[INFO] $OUT/ui-latest.json"
