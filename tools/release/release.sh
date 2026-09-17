#!/bin/sh
#
# Build and publish an IronKVM release.
#
#   release.sh 1.0.0                build, verify and publish
#   release.sh --dry-run 1.0.0      build and verify, publish nothing
#   release.sh --verify-only 1.0.0  check an existing output directory
#
# A release publishes the application tarball and latest.json. It does not
# publish a card image any more: ironkvm-1.1.0 was the last Buildroot card, and
# card images come from the ironkvm-dist repository from now on.
#
# Runs on a Linux host that has Docker, the MaixCDK builder image, Sipeed's
# pinned application tarball and gh. A hosted runner has none of the first
# three, which is why this is a script and not a workflow. A workflow file that
# cannot run is worse than no workflow file.
#
# Environment:
#   RELEASE_OUT        where artifacts are written    (default ./release-out)
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

# The tag carries a name and not only a number, because the two projects share
# one tag namespace and upstream is still using it. Sipeed tags each card image
# revision as vX.Y.Z, and those tags reach this repository with every fetch:
# v1.1.0 is Sipeed's, dated 2024-07-08, and v1.4.3 is Sipeed's too, dated
# 2026-06-09. The fork's first three releases took v1.0.1 to v1.0.3 and did not
# collide by luck rather than by design. The fourth wanted v1.1.0, which was
# taken.
#
# A prefixed tag cannot collide with an image revision, whatever number either
# project reaches next. The version itself does not change: latest.json, the
# About panel and the package name all read $VERSION, and only the tag and the
# asset URL carry the prefix.
#
# Releases 1.0.1 to 1.0.3 keep the tags they were published under. A retag
# would break the download URL of every asset they hold.
TAG="ironkvm-${VERSION}"

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
git rev-parse "$TAG" > /dev/null 2>&1 && {
    echo "tag $TAG already exists" >&2; exit 1; }

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

# The release notes, checked here and not at the point of use. publish() reaches
# for $OUT/notes.md after it has already pushed the tag, so a missing file there
# leaves a tag on the remote, no release under it, and a feed still pointing at
# the version before. Nothing wrote that file at all until this check existed.
#
# A dry run does not need notes, the same way it does not need gh.
NOTES_SRC="tools/release/notes/${VERSION}.md"
[ "$DRY_RUN" = "1" ] || [ -s "$NOTES_SRC" ] || {
    echo "$NOTES_SRC is missing or empty; write the release notes first" >&2
    exit 1; }

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

[ -f "$OFFICIAL_APP" ] || { echo "no such base input: $OFFICIAL_APP" >&2; exit 1; }

# Check the official application against the pin the repository records, rather
# than trusting a file name. The tarball is Sipeed's and the package layers the
# fork's own tree over it, so the pin is the only statement of which bytes a
# release was built from. A base that drifts produces a package nobody can
# reproduce and nobody would notice.
echo "==> verifying the official application against tools/release/APP.sha256"
have=$(sha256sum "$OFFICIAL_APP" | cut -d' ' -f1)
grep -q "^$have  " tools/release/APP.sha256 || {
    echo "$OFFICIAL_APP is $have, which is not pinned in tools/release/APP.sha256" >&2
    exit 1; }
echo "    $(basename "$OFFICIAL_APP") matches its pin"

mkdir -p "$OUT"
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT

# Every boot script the package carries must be either installed into
# /etc/init.d by install.sh or declared package-only. The check runs here,
# before anything is built, because its answer does not depend on the build and
# a release that would ship a fix nobody receives should cost seconds to stop.
#
# It caught nothing when it was written and would have caught S00kmod. The fork
# rewrote that script on 2026-09-04 to load the modules the package ships, the
# install list did not name it because until then the fork did not change it,
# and every board updated afterwards kept the stock loader.
echo "==> checking that every packaged boot script has a decided fate"
sh tools/release/test-init-install-list.sh > /dev/null || {
    sh tools/release/test-init-install-list.sh | grep FAIL >&2; exit 1; }

echo "==> building the web user interface"
# Built from a copy in $STAGE, without the developer's node_modules.
#
# web/node_modules belongs to whoever works in this checkout. Installing into it
# writes it for the host that builds the release, so a Windows workstation that
# also runs `pnpm dev` had its modules directory replaced with Linux binaries on
# every release. Reinstalling afterwards is a step somebody has to remember,
# which is the same as not having one.
#
# This costs nothing. The build runs in a container that is removed afterwards,
# so pnpm's store starts empty either way, and the only new work is copying the
# sources. Nothing in web/ refers to a path outside it.
#
# CI=true still: pnpm asks before it purges a modules directory another pnpm
# version left behind, and it refuses to purge without a TTY.
mkdir -p "$STAGE/web"
tar -cf - -C web --exclude=./node_modules --exclude=./dist . | tar -xf - -C "$STAGE/web"
( cd "$STAGE/web" && CI=true pnpm install --frozen-lockfile && pnpm build )

# The package takes web/dist from the repository root, so the output has to come
# back. It is plain JavaScript and carries no platform.
rm -rf web/dist
cp -a "$STAGE/web/dist" web/dist

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
# The package layers this under the fork's own kvmapp, because kvmapp/ holds
# only what the fork changes. The newer official tree has to be in place before
# the fork's files land on top of it.
rm -rf official-kvmapp
mkdir -p official-kvmapp
tar xzf "$OFFICIAL_APP" -C official-kvmapp --strip-components=1
[ -d official-kvmapp/server ] || {
    echo "$OFFICIAL_APP did not unpack into the expected shape" >&2; exit 1; }

echo "==> assembling the package"
PAYLOAD="$STAGE/ironkvm_${VERSION}"
mkdir -p "$PAYLOAD"
# The same layers, in the same order, that a slot image is built from.
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

# The two native libraries the fork builds, from the tracked path. The copy
# above takes them from kvmapp/server/dl_lib, which .gitignore excludes, so on a
# fresh clone it brings nothing and the package would ship Sipeed's libkvm.so
# with the fork's server: the H.264 encoder would strand its carveout on every
# stop and nothing would say so. The ironkvm-dist image manifest names the same
# two files, for the same reason.
cp server/dl_lib/libkvm.so server/dl_lib/libkvm_mmf.so "$PAYLOAD/server/dl_lib/"

# Read back what was assembled, not what was asked for. The whole point of the
# two lines above is that a missing source is silent, so the check has to look
# at the result.
for lib in libkvm.so libkvm_mmf.so; do
    want=$(md5sum < "server/dl_lib/$lib" | cut -d' ' -f1)
    got=$(md5sum < "$PAYLOAD/server/dl_lib/$lib" 2>/dev/null | cut -d' ' -f1)
    [ "$want" = "$got" ] || {
        echo "the package's $lib is not the one server/dl_lib holds" >&2
        echo "  want $want" >&2
        echo "  got  ${got:-nothing}" >&2
        exit 1; }
done

# Replaced, not merged. Two builds never collide on a hashed asset name, so a
# merge would leave the official bundle's files beside the fork's.
rm -rf "$PAYLOAD/server/web"
cp -a web/dist "$PAYLOAD/server/web"
echo "$VERSION" > "$PAYLOAD/version"

# Some boot scripts live in tools/ rather than in kvmapp/system/init.d, and the
# install list names them all the same. install.sh copies only what the package
# carries, so without this the tarball would ship the fork's boot behaviour
# minus its supervisor, its hardware watchdog and the OLED nudge: an update
# would install the scripts that can break a boot and leave out the ones that
# watch it.
#
# S00awatchdog and S02identity are not here any more. They belong to the
# ironkvm-dist repository, which installs them from its own image, and a
# Buildroot board already has the copies its card installed.
for s in tools/service/S98supervise tools/service/S01hwdt tools/oled/S97oled-nudge; do
    [ -f "$s" ] || { echo "missing boot script: $s" >&2; exit 1; }
    cp "$s" "$PAYLOAD/system/init.d/${s##*/}"
done

# Which scripts install.sh puts into /etc/init.d. The list is committed beside
# the package, and tools/release/test-init-install-list.sh holds it against the
# scripts the package carries.
#
# /kvmapp/system/init.d is the application's own reference copy. The list names
# the scripts the fork changes or adds, and leaves the rest at their base
# versions. install.sh installed the whole directory once, so an update started
# daemons that a board never ran before.
#
# rcS is not in the list, and it is deliberate. Every S* script an update
# installs is covered by the watchdog, which puts the previous set back when the
# board cannot be reached. rcS is what RUNS the watchdog. A syntactically valid
# rcS that does the wrong thing means
# nothing would run at all, including the watchdog: no attempt would be counted,
# no marker would be set for the recovery slot, and the board would boot into
# silence for ever. install.sh checks syntax, not behaviour.
#
# The cost of holding it back is that a board updated by package rather than by
# image keeps the stock rcS and writes no /bootlog. The stock one still runs
# every S* file, so the watchdog, the slots and the identity carry-over all
# work. A diagnostic is worth less than the path that repairs the board.
cp kvmapp/system/init.d.install "$PAYLOAD/system/init.d.install"
for want in $(cat kvmapp/system/init.d.install); do
    [ -f "$PAYLOAD/system/init.d/$want" ] || {
        echo "init.d.install names $want but the package does not carry it" >&2
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

echo "==> writing latest.json"
SHA=$(openssl dgst -sha512 -binary "$OUT/$PKG" | openssl base64 -A)
SIZE=$(wc -c < "$OUT/$PKG" | tr -d ' ')
UNPACKED=$(tar tzvf "$OUT/$PKG" | awk '{ total += $3 } END { print total }')
cat > "$OUT/latest.json" <<EOF
{
  "manifest_version": 2,
  "version": "$VERSION",
  "name": "$PKG",
  "url": "https://github.com/$REPO/releases/download/$TAG/$PKG",
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
( cd "$OUT" && sha256sum "$PKG" > SHA256SUMS )

# The notes travel with the artifacts, so what is published is a file in the
# repository and not prose typed at the prompt. A dry run copies them too: it is
# the only way to see the notes before they are the release.
if [ -s "$NOTES_SRC" ]; then
    cp "$NOTES_SRC" "$OUT/notes.md"
fi

verify

publish() {
    if [ "$DRY_RUN" = "1" ]; then
        echo "dry run: nothing published. The artifacts are in $OUT"
        return 0
    fi

    git tag -a "$TAG" -m "IronKVM $VERSION"
    git push origin "$TAG"
    gh release create "$TAG" --repo "$REPO" --title "IronKVM $VERSION" \
        --notes-file "$OUT/notes.md" \
        "$OUT/$PKG" "$OUT/SHA256SUMS"

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
