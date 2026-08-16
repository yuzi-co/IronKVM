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
#   STOCK_BOOT_SD      the stock boot.sd to repack    (default base/boot/boot.sd)
#   OFFICIAL_APP       the pinned official application (default base/nanokvm_2.5.0.tar.gz)
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
STOCK_BOOT_SD=${STOCK_BOOT_SD:-base/boot/boot.sd}
OFFICIAL_APP=${OFFICIAL_APP:-base/nanokvm_2.5.0.tar.gz}

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

for f in "$BASE_TAR" "$STOCK_BOOT_SD" "$OFFICIAL_APP"; do
    [ -f "$f" ] || { echo "no such base input: $f" >&2; exit 1; }
done
[ -d "$BASE_BOOT" ] || { echo "no such boot directory: $BASE_BOOT" >&2; exit 1; }

# Check the base against the pins the repository records, rather than trusting a
# file name. tools/abslots/BASE.sha256 exists because the artefacts come from
# Sipeed and one of them ships no checksum of its own, so the pin is the only
# statement of which bytes a slot was ever built from. A base that drifts
# produces an image nobody can reproduce and nobody would notice.
echo "==> verifying the base against tools/abslots/BASE.sha256"
for f in "$BASE_TAR" "$OFFICIAL_APP"; do
    have=$(sha256sum "$f" | cut -d' ' -f1)
    grep -q "^$have  " tools/abslots/BASE.sha256 || {
        echo "$f is $have, which is not pinned in tools/abslots/BASE.sha256" >&2
        exit 1; }
    echo "    $(basename "$f") matches its pin"
done

mkdir -p "$OUT"
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT

echo "==> building the web user interface"
( cd web && pnpm install --frozen-lockfile && pnpm build )

echo "==> cross-compiling the server"
# The Makefile's own recipe, run directly. `make` is not installed on every host
# that has Docker and the builder image, and this script must not need a fourth
# tool to run one command. -buildvcs=false because Docker shows the bind mount as
# root-owned, git then refuses the checkout as dubious, and Go stops.
docker run -e UID="$(id -u)" -e GID="$(id -g)" -v "$PWD:/home/build/NanoKVM" --rm \
    "${BUILDER_IMAGE:-nanokvm-builder-local-$(id -u)-$(id -g)}" /bin/bash -c \
    "cd /home/build/NanoKVM/server && go mod tidy \
     && CGO_ENABLED=1 GOOS=linux GOARCH=riscv64 CC=riscv64-unknown-linux-musl-gcc \
        CGO_CFLAGS='-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d' \
        go build -buildvcs=false -ldflags '-s -w -X NanoKVM-Server/common/version.Build='"

# The build does not patch the RPATH, and a binary without it does not start on
# the device: the loader cannot find libkvm.so.
docker run --rm -v "$PWD/server:/src" -w /src ubuntu:24.04 \
    sh -c 'apt-get update -qq && apt-get install -y -qq patchelf \
           && patchelf --add-rpath "\$ORIGIN/dl_lib" NanoKVM-Server'

# The release binary carries no build stamp, the way a release does. The stamp
# exists to identify a hand-built server; a released one is identified by the
# version the updater writes.
echo "==> unpacking the pinned official application"
# root.manifest layers this under the fork's own kvmapp, because the base rootfs
# carries 2.4.3 and the manifest's first add has to put the newer official tree
# in place before the fork's files land on top of it.
rm -rf official-kvmapp
mkdir -p official-kvmapp
tar xzf "$OFFICIAL_APP" -C official-kvmapp --strip-components=1
[ -d official-kvmapp/server ] || {
    echo "$OFFICIAL_APP did not unpack into the expected shape" >&2; exit 1; }

echo "==> assembling the package"
PAYLOAD="$STAGE/ironkvm_${VERSION}"
mkdir -p "$PAYLOAD"
cp -a kvmapp/. "$PAYLOAD/"
cp    server/NanoKVM-Server "$PAYLOAD/server/NanoKVM-Server"
cp -a web/dist              "$PAYLOAD/server/web"
echo "$VERSION" > "$PAYLOAD/version"

# Some boot scripts live in tools/ rather than in kvmapp/system/init.d, and the
# image manifest adds them to /etc/init.d straight from there. install.sh copies
# only what the package carries, so without this the tarball would ship the
# fork's boot behaviour minus its watchdog, its supervisor and the OLED nudge:
# an update would install the scripts that can break a boot and leave out the
# one that undoes them.
for s in tools/abslots/device/S00awatchdog tools/service/S98supervise \
         tools/oled/S97oled-nudge; do
    [ -f "$s" ] || { echo "missing boot script: $s" >&2; exit 1; }
    cp "$s" "$PAYLOAD/system/init.d/${s##*/}"
done

# Every script the image installs to /etc/init.d must also be in the package, or
# an image and a tarball of the same release boot differently.
for want in $(sed -n 's|^add .* /etc/init.d/\([^ ]*\).*|\1|p' \
              tools/abslots/manifest/root.manifest); do
    [ -f "$PAYLOAD/system/init.d/$want" ] || {
        echo "the image installs /etc/init.d/$want but the package does not carry it" >&2
        exit 1; }
done

# The base travels beside the version because semver cannot carry it.
BASE=$(cat "${BASE_VERSION_FILE:-base/version}" 2>/dev/null || echo "unknown")
echo "$BASE" > "$PAYLOAD/base-version"

chmod 755 "$PAYLOAD/system/install.sh" "$PAYLOAD/server/NanoKVM-Server"
tar czf "$OUT/$PKG" -C "$STAGE" "ironkvm_${VERSION}"

echo "==> building the slot filesystems"
# build-image.sh takes five positional arguments and makes ONE ext4 root. The
# sizes are the partition sizes from partition.sfdisk.
#
# The payload is the REPOSITORY ROOT, not the package staged above. root.manifest
# names paths like official-kvmapp/, kvmapp/, server/NanoKVM-Server, web/dist and
# tools/abslots/device/*, which only exist together here. Handing it the package
# directory instead produces an image missing everything the manifest adds from
# tools/, and the build reports success either way.
tools/abslots/build-image.sh "$BASE_TAR" tools/abslots/manifest/root.manifest \
    . 2048 "$STAGE/root.img"
tools/abslots/build-image.sh "$BASE_TAR" tools/abslots/manifest/recovery.manifest \
    . 1024 "$STAGE/recovery.img"

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
