#!/bin/sh
# Tests for scripts/package-ui.sh, the web UI package build.
#
# Runs the script on a small fake web/dist with a key made for the run, then
# reads what it wrote: the archive layout, the manifest's hashes, the feed entry
# and the signature. Nothing here touches a board; the server side has its own
# tests in server/service/webui.
#
# Needs bash, node, GNU tar, gzip, openssl and minisign. An Alpine container has
# them all:
#
#   docker run --rm -v "$PWD:/repo" -w /repo alpine:3.20 sh -c \
#     'apk add -q bash nodejs tar openssl minisign && sh tools/release/test-package-ui.sh'
#
# The key is made with `minisign -G -W` in a temporary directory and removed with
# it. No key is committed for this test.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../.." && pwd)
SCRIPT="$ROOT/scripts/package-ui.sh"
pass=0
fail=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

for t in bash node tar gzip openssl minisign; do
    # Exit 2: a suite that ran no case has not passed.
    command -v "$t" > /dev/null 2>&1 || { echo "$(basename "$0"): needs $t" >&2; exit 2; }
done
tar --version 2> /dev/null | grep -q 'GNU tar' || { echo "$(basename "$0"): needs GNU tar" >&2; exit 2; }

W=$(mktemp -d)
trap 'rm -rf "$W"' EXIT

minisign -G -W -p "$W/test.pub" -s "$W/test.key" > /dev/null 2>&1
minisign -G -W -p "$W/other.pub" -s "$W/other.key" > /dev/null 2>&1

dist="$W/dist"
mkdir -p "$dist/assets"
printf '<!doctype html><script type="module" src="/assets/index-abc.js"></script>' > "$dist/index.html"
printf 'console.log(1)' > "$dist/assets/index-abc.js"
printf '{"version":"1.2.0","api":{"min":1,"max":1}}\n' > "$dist/ui-version.json"

echo "package-ui.sh"

run() { DIST="$dist" OUT="$W/out" bash "$SCRIPT" "$@" > "$W/log" 2>&1; }

run 1.2 && bad "a version that is not semver is refused" || ok "a version that is not semver is refused"
run 1.3.0 && bad "a build stamped with another version is refused" || ok "a build stamped with another version is refused"
grep -q "UI_VERSION=1.3.0" "$W/log" && ok "the refusal says how to rebuild" || bad "the refusal says how to rebuild"

SIGN_KEY="$W/test.key" VERIFY_KEYS="$W/other.pub $W/test.pub" UI_PACKAGE_URL=https://example.org/ironkvm-ui_1.2.0.tar.gz run 1.2.0
check "a signed package builds" "$?" 0
pkg="$W/out/ironkvm-ui_1.2.0.tar.gz"

[ -f "$pkg" ] && ok "the package is named ironkvm-ui_<version>.tar.gz" || bad "the package is named ironkvm-ui_<version>.tar.gz"
check "the archive holds manifest.json and web/ only" \
    "$(tar -tzf "$pkg" | sed 's#/.*##' | sort -u | tr '\n' ' ')" "manifest.json web "
check "no entry is owned by anyone" "$(tar -tvzf "$pkg" | awk '{print $2}' | sort -u)" "0/0"

mkdir "$W/x" && tar -xzf "$pkg" -C "$W/x"
manifest="$W/x/manifest.json"
check "the manifest names the package" \
    "$(node -e 'const m=require(process.argv[1]);console.log(m.format,m.name,m.version,m.api.min,m.api.max)' "$manifest")" \
    "1 ironkvm-ui 1.2.0 1 1"
hashes_ok=$(node -e '
const fs=require("fs"),c=require("crypto"),m=require(process.argv[1]);
let ok=Object.keys(m.files).length===3;
for (const [f,h] of Object.entries(m.files)) {
  ok = ok && c.createHash("sha256").update(fs.readFileSync(process.argv[2]+"/web/"+f)).digest("hex")===h;
}
console.log(ok);' "$manifest" "$W/x")
check "the manifest lists every file with its sha256" "$hashes_ok" true

minisign -V -q -p "$W/test.pub" -m "$pkg" > /dev/null 2>&1
check "the signature verifies with the signing key" "$?" 0
minisign -V -q -p "$W/other.pub" -m "$pkg" > /dev/null 2>&1 && bad "another key does not verify it" || ok "another key does not verify it"

feed="$W/out/ui-latest.json"
check "the feed entry" \
    "$(node -e 'const l=require(process.argv[1]);console.log(l.format,l.version,l.name,l.url,l.api.min,l.api.max)' "$feed")" \
    "1 1.2.0 ironkvm-ui_1.2.0.tar.gz https://example.org/ironkvm-ui_1.2.0.tar.gz 1 1"
check "the feed's size is the package's" "$(node -e 'console.log(require(process.argv[1]).size_bytes)' "$feed")" "$(wc -c < "$pkg" | tr -d ' ')"
check "the feed's sha512 is the package's, in base64" \
    "$(node -e 'console.log(require(process.argv[1]).sha512)' "$feed")" \
    "$(openssl dgst -sha512 -binary "$pkg" | openssl base64 -A)"

cp "$pkg" "$W/first.tar.gz"
SIGN_KEY="$W/test.key" run 1.2.0
cmp -s "$pkg" "$W/first.tar.gz" && ok "the same build makes the same package" || bad "the same build makes the same package"

SIGN_KEY="$W/test.key" VERIFY_KEYS="$W/other.pub" run 1.2.0 \
    && bad "a signature no trusted key verifies fails the build" || ok "a signature no trusted key verifies fails the build"
[ -f "$pkg.minisig" ] && bad "and the signature is removed" || ok "and the signature is removed"

run 1.2.0
check "an unsigned build still succeeds" "$?" 0
grep -q "unsigned" "$W/log" && ok "and says it is unsigned" || bad "and says it is unsigned"
[ -f "$pkg.minisig" ] && bad "and leaves no signature" || ok "and leaves no signature"

echo "$pass passed, $fail failed"
[ "$fail" -eq 0 ]
