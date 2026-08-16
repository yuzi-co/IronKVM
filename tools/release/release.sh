#!/bin/sh
#
# Build and publish an IronKVM release.
#
#   release.sh 1.0.0                build, verify and publish
#   release.sh --dry-run 1.0.0      build and verify, publish nothing
#   release.sh --verify-only 1.0.0  check an existing output directory
#
# Runs on a Linux host that has Docker, the MaixCDK builder image, a Sipeed base
# image, minisign and gh. A hosted runner has none of the first three, which is
# why this is a script and not a workflow. A workflow file that cannot run is
# worse than no workflow file.
#
# Environment:
#   RELEASE_OUT        where artifacts are written    (default ./release-out)
#   BASE_TAR           the pinned base rootfs tarball (default base/rootfs.tar.zst)
#   BASE_BOOT          a directory holding /boot      (default base/boot)
#   STOCK_BOOT_SD      the stock boot.sd to repack    (default base/boot.sd)
#   BASE_VERSION_FILE  the official version it is from (default base/version)
#   REPO               the GitHub repository          (default yuzi-co/IronKVM)

set -eu

DRY_RUN=0
VERIFY_ONLY=0
VERSION=""

while [ $# -gt 0 ]; do
    case "$1" in
        --dry-run)     DRY_RUN=1; shift ;;
        --verify-only) VERIFY_ONLY=1; shift ;;
        -*)            echo "unknown option: $1" >&2; exit 1 ;;
        *)             VERSION=$1; shift ;;
    esac
done

OUT=${RELEASE_OUT:-./release-out}
REPO=${REPO:-yuzi-co/IronKVM}
BASE_TAR=${BASE_TAR:-base/rootfs.tar.zst}
BASE_BOOT=${BASE_BOOT:-base/boot}
STOCK_BOOT_SD=${STOCK_BOOT_SD:-base/boot.sd}

[ -n "$VERSION" ] || { echo "usage: release.sh [--dry-run|--verify-only] X.Y.Z" >&2; exit 1; }

# The device refuses any package name that is not (nanokvm|ironkvm)_X.Y.Z.tar.gz,
# so an unusable version has to be caught here rather than on a board.
#
# A prerelease sorts BELOW the release it came from and build metadata compares
# EQUAL to it, because semver ignores metadata entirely. The first would leave
# the update page offering an older build for ever; the second would make the
# release invisible to every device already on that version. Neither is a
# theoretical worry: the build stamp uses metadata for exactly that property.
case "$VERSION" in
    *-*|*+*)
        echo "version must carry no prerelease and no build metadata, got '$VERSION'" >&2
        exit 1 ;;
esac
case "$VERSION" in
    *[!0-9.]*|.*|*.|*..*)
        echo "version must be X.Y.Z, got '$VERSION'" >&2; exit 1 ;;
esac
[ "$(echo "$VERSION" | tr -cd . | wc -c)" -eq 2 ] || {
    echo "version must be X.Y.Z, got '$VERSION'" >&2; exit 1; }

PKG="ironkvm_${VERSION}.tar.gz"
IMG="ironkvm-${VERSION}-sdcard.img.xz"

json_string() {
    sed -n "s/.*\"$1\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" "$OUT/latest.json" | head -1
}

json_number() {
    sed -n "s/.*\"$1\"[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p" "$OUT/latest.json" | head -1
}

# verify checks that the manifest describes the artifact it names, and that the
# artifact is shaped the way the updater expects.
#
# It exists because a feed pointing at a package it does not describe fails
# nowhere until a device tries to install, and because three guards in this
# repository have already rotted into passing while testing nothing. A release
# script that cannot check its own output is one of those waiting to happen.
verify() {
    [ -f "$OUT/$PKG" ]        || { echo "missing $OUT/$PKG" >&2; return 1; }
    [ -f "$OUT/latest.json" ] || { echo "missing $OUT/latest.json" >&2; return 1; }

    named=$(json_string name)
    [ "$named" = "$PKG" ] || {
        echo "latest.json names '$named' but the build produced '$PKG'" >&2; return 1; }

    actual=$(openssl dgst -sha512 -binary "$OUT/$PKG" | openssl base64 -A)
    stated=$(json_string sha512)
    [ "$actual" = "$stated" ] || {
        echo "the sha512 in latest.json does not match $PKG" >&2; return 1; }

    size=$(wc -c < "$OUT/$PKG" | tr -d ' ')
    [ "$size" = "$(json_number size_bytes)" ] || {
        echo "size_bytes in latest.json does not match $PKG" >&2; return 1; }

    # The updater derives the directory it expects from the package name, so a
    # tarball whose top directory disagrees unpacks into something it refuses.
    root=$(tar tzf "$OUT/$PKG" 2>/dev/null | head -1 | cut -d/ -f1)
    [ "$root" = "ironkvm_${VERSION}" ] || {
        echo "the tarball's top directory is '$root', expected ironkvm_${VERSION}" >&2
        return 1; }

    echo "verify: latest.json and $PKG agree"
}

if [ "$VERIFY_ONLY" = "1" ]; then
    verify
    exit $?
fi

git diff --quiet && git diff --cached --quiet || {
    echo "the tree is dirty; commit or stash first" >&2; exit 1; }
git rev-parse "v$VERSION" > /dev/null 2>&1 && {
    echo "tag v$VERSION already exists" >&2; exit 1; }

for f in "$BASE_TAR" "$STOCK_BOOT_SD"; do
    [ -f "$f" ] || { echo "no such base input: $f" >&2; exit 1; }
done
[ -d "$BASE_BOOT" ] || { echo "no such boot directory: $BASE_BOOT" >&2; exit 1; }

mkdir -p "$OUT"
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT

echo "==> building the web user interface"
( cd web && pnpm install --frozen-lockfile && pnpm build )

echo "==> cross-compiling the server"
make app DOCKER_TTY=
# make app does not patch the RPATH, and a binary without it does not start on
# the device.
docker run --rm -v "$PWD/server:/src" -w /src ubuntu:24.04 \
    sh -c 'apt-get update -qq && apt-get install -y -qq patchelf \
           && patchelf --add-rpath "\$ORIGIN/dl_lib" NanoKVM-Server'

echo "==> assembling the package"
PAYLOAD="$STAGE/ironkvm_${VERSION}"
mkdir -p "$PAYLOAD"
cp -a kvmapp/. "$PAYLOAD/"
cp    server/NanoKVM-Server "$PAYLOAD/server/NanoKVM-Server"
cp -a web/dist              "$PAYLOAD/server/web"
echo "$VERSION" > "$PAYLOAD/version"

# The base travels beside the version because semver cannot carry it.
BASE=$(cat "${BASE_VERSION_FILE:-base/version}" 2>/dev/null || echo "unknown")
echo "$BASE" > "$PAYLOAD/base-version"

chmod 755 "$PAYLOAD/system/install.sh" "$PAYLOAD/server/NanoKVM-Server"
tar czf "$OUT/$PKG" -C "$STAGE" "ironkvm_${VERSION}"

echo "==> building the slot filesystems"
# build-image.sh takes five positional arguments and makes ONE ext4 root. The
# sizes are the partition sizes from partition.sfdisk.
tools/abslots/build-image.sh "$BASE_TAR" tools/abslots/manifest/root.manifest \
    "$PAYLOAD" 2048 "$STAGE/root.img"
tools/abslots/build-image.sh "$BASE_TAR" tools/abslots/manifest/recovery.manifest \
    "$PAYLOAD" 1024 "$STAGE/recovery.img"

echo "==> repacking the boot image"
tools/abslots/repack-boot.sh "$STOCK_BOOT_SD" "$STAGE/bootbuild"
cp -a "$BASE_BOOT/." "$STAGE/boot/"
cp "$STAGE/bootbuild/boot.sd" "$STAGE/boot/boot.sd"

echo "==> assembling the card"
tools/abslots/build-card.sh "$STAGE/boot" "$STAGE/root.img" "$STAGE/recovery.img" \
    "$OUT/ironkvm-${VERSION}-sdcard.img"
xz -T0 -f "$OUT/ironkvm-${VERSION}-sdcard.img"

echo "==> writing latest.json"
SHA=$(openssl dgst -sha512 -binary "$OUT/$PKG" | openssl base64 -A)
SIZE=$(wc -c < "$OUT/$PKG" | tr -d ' ')
UNPACKED=$(tar tzvf "$OUT/$PKG" | awk '{ total += $3 } END { print total }')
cat > "$OUT/latest.json" <<EOF
{
  "manifest_version": 2,
  "version": "$VERSION",
  "name": "$PKG",
  "url": "https://github.com/$REPO/releases/download/v$VERSION/$PKG",
  "sha512": "$SHA",
  "size": $SIZE,
  "size_bytes": $SIZE,
  "unpacked_size_bytes": $UNPACKED
}
EOF

echo "==> checksums"
( cd "$OUT" && sha256sum "$PKG" "$IMG" > SHA256SUMS )
minisign -Sm "$OUT/SHA256SUMS"

verify

publish() {
    if [ "$DRY_RUN" = "1" ]; then
        echo "dry run: nothing published. The artifacts are in $OUT"
        return 0
    fi

    git tag -a "v$VERSION" -m "IronKVM $VERSION"
    git push origin "v$VERSION"
    gh release create "v$VERSION" --repo "$REPO" --title "IronKVM $VERSION" \
        --notes-file "$OUT/notes.md" \
        "$OUT/$PKG" "$OUT/$IMG" "$OUT/SHA256SUMS" "$OUT/SHA256SUMS.minisig"

    # The feed lives on its own branch so a 26 MB package never enters the
    # repository's history.
    git worktree add "$STAGE/pages" gh-pages
    cp "$OUT/latest.json" "$STAGE/pages/latest.json"
    ( cd "$STAGE/pages" && git add latest.json \
      && git commit -m "Publish IronKVM $VERSION" && git push origin gh-pages )
    git worktree remove "$STAGE/pages"
}

publish
echo "IronKVM $VERSION done"
