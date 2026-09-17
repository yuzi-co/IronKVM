#!/bin/sh
#
# Build base/ from Sipeed's published downloads.
#
#   fetch-base.sh [target-dir]      default: ./base
#
# Needs curl, tar and sha256sum. Downloads about 16 MB.
#
# WHY THIS EXISTS RATHER THAN A HOSTED COPY
#
# The application tarball is Sipeed's build. Its GPL parts could be
# redistributed, but it also carries vendor binaries for the SG2002 whose terms
# are not stated, so this fork does not republish it. Fetching from the
# publisher and checking the bytes against tools/release/APP.sha256 gets the
# same reproducibility with no licensing question to answer on somebody else's
# behalf.
#
# The tarball is verified before it is used. It has a checksum published beside
# it in the same release, and APP.sha256 records that checksum.
#
# A card image needs more than this: the base root filesystem and /boot. Neither
# is fetched here any more. ironkvm-1.1.0 was the last Buildroot card image, and
# card images are built by the ironkvm-dist repository from now on.
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../.." && pwd)
DEST=${1:-$ROOT/base}
PINS="$HERE/APP.sha256"

APP_NAME=nanokvm_2.5.0.tar.gz
APP_URL=${APP_URL:-https://github.com/sipeed/NanoKVM/releases/download/2.5.0/$APP_NAME}

[ -f "$PINS" ] || { echo "no pin file at $PINS" >&2; exit 1; }

for t in curl tar sha256sum; do
    command -v "$t" > /dev/null 2>&1 || { echo "$t is required" >&2; exit 1; }
done

pin_for() {
    grep "  $1\( \|$\)" "$PINS" | grep -o '^[0-9a-f]\{64\}' | head -1
}

# check verifies a file against the pin recorded for its NAME, and refuses to
# continue without one. A download that is merely present proves nothing: the
# reason this script exists is that two builds must be comparable.
check() {
    want=$(pin_for "$2")
    [ -n "$want" ] || { echo "no pin recorded for $2" >&2; return 1; }
    have=$(sha256sum "$1" | cut -d' ' -f1)
    [ "$have" = "$want" ] || {
        echo "$2 is $have, pinned as $want" >&2
        return 1; }
    echo "    $2 matches its pin"
}

mkdir -p "$DEST"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

echo "==> fetching the official application"
curl -fSL --retry 3 -o "$WORK/$APP_NAME" "$APP_URL"
check "$WORK/$APP_NAME" "$APP_NAME"
cp "$WORK/$APP_NAME" "$DEST/$APP_NAME"

# The application version, which travels beside the fork's own version because
# semver cannot carry it.
echo "$APP_NAME" | sed 's/^nanokvm_//; s/\.tar\.gz$//' > "$DEST/version"

echo
echo "base/ is ready in $DEST"
ls -la "$DEST"
