#!/bin/sh
#
# Build and publish an IronKVM release.
#
#   release.sh 1.0.0                build, verify and publish
#   release.sh --dry-run 1.0.0      build and verify, publish nothing
#   release.sh --verify-only 1.0.0  check an existing output directory
#
# Runs on a Linux host that has Docker, the MaixCDK builder image, a Sipeed base
# image and gh. A hosted runner has none of the first three, which is
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
#   BUILDER_IMAGE      the MaixCDK builder image      (default from BUILD_UID/GID)
#   BUILD_UID/GID      the identity the builder runs as (default the host's own)

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

# Every tool the run will need, checked before anything is built. A build that
# discovers a missing tool at its last step has already spent twenty minutes and
# leaves half a release behind. gh is only needed to publish, so a dry run does
# not demand it.
echo "==> checking the host"
NEED="docker pnpm tar xz openssl sha256sum git"
[ "$DRY_RUN" = "1" ] || NEED="$NEED gh"
missing=""
for t in $NEED; do
    command -v "$t" > /dev/null 2>&1 || missing="$missing $t"
done
[ -z "$missing" ] || { echo "missing tools:$missing" >&2; exit 1; }

# The builder image bakes the ownership of /home/build in at build time, and its
# entrypoint drops to whatever UID it is given. The two agree only on the machine
# that built the image. A WSL shell reports 1000, and an image built from a
# Windows checkout carries that account's id, so handing the host's own id to a
# foreign image produces a build that runs as a user which cannot write $HOME,
# and go stops at the module cache. Override both when they differ.
BUILD_UID=${BUILD_UID:-$(id -u)}
BUILD_GID=${BUILD_GID:-$(id -g)}
BUILDER=${BUILDER_IMAGE:-nanokvm-builder-local-${BUILD_UID}-${BUILD_GID}}
docker image inspect "$BUILDER" > /dev/null 2>&1 || {
    echo "no builder image '$BUILDER'; build it once with 'make shell'" >&2; exit 1; }
echo "    tools present, builder image $BUILDER"

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
# CI=true because pnpm asks before it purges a modules directory that another
# platform or another pnpm version installed, and it refuses to purge without a
# TTY. A release runs from a script and has no answer to give.
#
# The purge is not a side issue. web/node_modules is the developer's, and the
# install that follows writes it for the host that builds the release. A
# workstation that also runs `pnpm dev` gets its modules directory rebuilt.
( cd web && CI=true pnpm install --frozen-lockfile && pnpm build )

echo "==> cross-compiling the server"
# The Makefile's own recipe, run directly. `make` is not installed on every host
# that has Docker and the builder image, and this script must not need a fourth
# tool to run one command. -buildvcs=false because Docker shows the bind mount as
# root-owned, git then refuses the checkout as dubious, and Go stops.
docker run -e UID="$BUILD_UID" -e GID="$BUILD_GID" -v "$PWD:/home/build/NanoKVM" --rm \
    "$BUILDER" /bin/bash -c \
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
# The same layers, in the same order, that root.manifest builds the image from.
#
# The updater REPLACES /kvmapp rather than merging into it: it moves the whole
# tree to the backup directory and moves the new one in. Every file the package
# leaves out is therefore a file the device loses.
#
# kvmapp/ holds only what the fork changes. It carries 14 of the 37 libraries in
# server/dl_lib and neither kvm_system nor system/tool, so a package built from
# it alone installs a server that cannot load libkvm.so. The board still answers
# ssh after that, so the watchdog reports it healthy, nothing rolls back, and
# the KVM is simply gone.
cp -a official-kvmapp/. "$PAYLOAD/"
cp -a kvmapp/.          "$PAYLOAD/"
cp    server/NanoKVM-Server "$PAYLOAD/server/NanoKVM-Server"

# Replaced, not merged. Two builds never collide on a hashed asset name, so a
# merge would leave the official bundle's files beside the fork's.
rm -rf "$PAYLOAD/server/web"
cp -a web/dist "$PAYLOAD/server/web"
echo "$VERSION" > "$PAYLOAD/version"

# Some boot scripts live in tools/ rather than in kvmapp/system/init.d, and the
# image manifest adds them to /etc/init.d straight from there. install.sh copies
# only what the package carries, so without this the tarball would ship the
# fork's boot behaviour minus its watchdog, its supervisor and the OLED nudge:
# an update would install the scripts that can break a boot and leave out the
# one that undoes them.
for s in tools/abslots/device/S00awatchdog tools/service/S98supervise \
         tools/oled/S97oled-nudge tools/abslots/device/S02identity; do
    [ -f "$s" ] || { echo "missing boot script: $s" >&2; exit 1; }
    cp "$s" "$PAYLOAD/system/init.d/${s##*/}"
done

# Every script the image installs to /etc/init.d must also be in the package, or
# an image and a tarball of the same release boot differently.
for want in $(sed -n 's|^add .* /etc/init.d/\([^ ]*\).*|\1|p' \
              tools/abslots/manifest/root.manifest); do
    # rcS is the one exception, and it is deliberate. Every S* script an update
    # installs is covered by the watchdog, which puts the previous set back when
    # the board cannot be reached. rcS is what RUNS the watchdog. A syntactically
    # valid rcS that does the wrong thing means
    # nothing would run at all, including the watchdog: no attempt would be
    # counted, no marker would be set for the recovery slot, and the board would
    # boot into silence for ever. install.sh checks syntax, not behaviour.
    #
    # The cost of holding it back is that a board updated by package rather than
    # by image keeps the stock rcS and writes no /bootlog. The stock one still
    # runs every S* file, so the watchdog, the slots and the identity carry-over
    # all work. A diagnostic is worth less than the path that repairs the board.
    [ "$want" = rcS ] && continue
    [ -f "$PAYLOAD/system/init.d/$want" ] || {
        echo "the image installs /etc/init.d/$want but the package does not carry it" >&2
        exit 1; }
done

# The base travels beside the version because semver cannot carry it.
BASE=$(cat "${BASE_VERSION_FILE:-base/version}" 2>/dev/null || echo "unknown")
echo "$BASE" > "$PAYLOAD/base-version"

# Nothing the official package carries may go missing. The layering above is
# what puts those files there, and this is what says so afterwards: a guard that
# reads the built tree cannot rot into passing the way a comment can.
#
# Two exceptions. kvm/ holds runtime state the device writes for itself. The web
# directory is replaced wholesale by this fork's own build, so none of the
# official bundle's hashed asset names survive, and neither does its icon.
tar tzf "$OFFICIAL_APP" | sed 's|^[^/]*/||' | grep -v '/$' | sort > "$STAGE/official.list"
( cd "$PAYLOAD" && find . -type f | sed 's|^\./||' | sort ) > "$STAGE/payload.list"
gone=$(comm -13 "$STAGE/payload.list" "$STAGE/official.list" | grep -vE '^(kvm/|server/web/)') || true
[ -z "$gone" ] || {
    echo "the package drops files the official one carries:" >&2
    echo "$gone" >&2
    exit 1; }

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
# repack-boot.sh writes boot.sd.new, and the name is deliberate: it says the
# image has not been accepted yet. It is accepted here, after that script's own
# verification has passed.
cp "$STAGE/bootbuild/boot.sd.new" "$STAGE/boot/boot.sd"

echo "==> assembling the card"
# Assembled in $STAGE, and only the compressed image is moved out.
#
# build-card.sh creates the card at its full 28.85 GiB and truncates it
# afterwards, because the table has to describe a data partition the image does
# not carry. On a filesystem with sparse files that hole costs nothing. On one
# without it costs 25 GB of real writes, and the two slots are then read back out
# of it to be checked, so the same bytes cross the filesystem three times.
#
# $OUT is wherever the publisher keeps artifacts. On a Windows workstation that
# is a bind mount with no sparse support: measured at 8.0G allocated for an 8.0G
# hole, against 0 on ext4. $STAGE is the build host's own filesystem and always
# has them. This is not a flag to remember, which is the point.
tools/abslots/build-card.sh "$STAGE/boot" "$STAGE/root.img" "$STAGE/recovery.img" \
    "$STAGE/ironkvm-${VERSION}-sdcard.img"
xz -T0 -f "$STAGE/ironkvm-${VERSION}-sdcard.img"
mv "$STAGE/ironkvm-${VERSION}-sdcard.img.xz" "$OUT/$IMG"

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
# SHA256SUMS is not signed. It proves a download is intact, and it proves
# nothing about who produced it: anyone who can replace the artifacts can
# replace this file beside them.
#
# A signature was designed in and then dropped for 1.0, because it is only worth
# what the key's safekeeping is worth. An offline key would let somebody who
# pinned it detect a later compromise of this repository; a key sitting on the
# build machine with no backup would add ceremony and no protection, and losing
# it would force every user to re-trust from scratch. The README says the
# checksums are unsigned rather than implying otherwise.
#
# The device does not read this file at all. Its update path checks a sha512
# from latest.json over TLS, so signing here would never have protected it.
( cd "$OUT" && sha256sum "$PKG" "$IMG" > SHA256SUMS )

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
        "$OUT/$PKG" "$OUT/$IMG" "$OUT/SHA256SUMS"

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
