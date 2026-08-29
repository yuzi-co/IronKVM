#!/bin/sh
# Hold both builder images to their pinned, verified downloads.
#
#   test-pinned-inputs.sh [path-to-tools/build/Dockerfile] [path-to-docker/Dockerfile]
#
# Not destructive: it reads the two Dockerfiles and runs no build.
#
# Why this test exists. Before 2026-08-29 neither image pinned anything it
# fetched. MaixCDK was cloned from its default branch with no tag and no
# commit, and the two archives were unpacked with no checksum. Two builds of
# the same commit of this repository could therefore link libkvm.so against
# different SDK sources, and nothing reported it. The failure is silent by
# nature, so a test is the only thing that keeps the pins in place.
HERE=$(cd "$(dirname "$0")" && pwd)
APP=${1:-$HERE/Dockerfile}
FULL=${2:-$HERE/../../docker/Dockerfile}

[ -f "$APP" ]  || { echo "no app builder Dockerfile at $APP"; exit 2; }
[ -f "$FULL" ] || { echo "no full builder Dockerfile at $FULL"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() {
    if [ "$2" = "$3" ]; then note "$1" OK; else note "$1 (got '$2', want '$3')" FAIL; fi
}

# arg_of <file> <name> prints the default of `ARG <name>=<value>`.
arg_of() {
    sed -n "s/^ARG  *$2=\\(.*\\)\$/\\1/p" "$1" | head -1
}

echo "===== the two images agree on what they download ====="
# They fetch the same two archives. An image that verifies while the other does
# not is the same hole with one more place to look for it.
for name in HOST_TOOLS_SHA256 GO_SHA256_AMD64 GO_VERSION HOST_TOOLS_URL
do
    a=$(arg_of "$APP" "$name")
    f=$(arg_of "$FULL" "$name")
    if [ -z "$a" ]; then note "$name is set in tools/build/Dockerfile" FAIL; continue; fi
    if [ -z "$f" ]; then note "$name is set in docker/Dockerfile" FAIL; continue; fi
    check "$name matches in both Dockerfiles" "$a" "$f"
done

echo
echo "===== every download is checked before it is unpacked ====="
# The shape that matters is "wget, then sha256sum -c, then tar". A tar that
# runs before the check has already trusted the bytes.
for f in "$APP" "$FULL"
do
    label=$(basename "$(dirname "$f")")/$(basename "$f")
    downloads=$(grep -c 'wget .*-O /tmp/' "$f")
    checks=$(grep -c 'sha256sum -c -' "$f")
    check "$label: one sha256sum -c per download" "$checks" "$downloads"

    # A tar that appears before any sha256sum -c in the same RUN would unpack
    # unverified bytes. Compare first-occurrence line numbers per download.
    bad=0
    for line in $(grep -n 'wget .*-O /tmp/' "$f" | cut -d: -f1)
    do
        rest=$(tail -n +"$line" "$f")
        sumline=$(printf '%s\n' "$rest" | grep -n 'sha256sum -c -' | head -1 | cut -d: -f1)
        tarline=$(printf '%s\n' "$rest" | grep -n 'tar -C /usr/local' | head -1 | cut -d: -f1)
        [ -n "$sumline" ] || { bad=$((bad + 1)); continue; }
        [ -n "$tarline" ] || continue
        [ "$sumline" -lt "$tarline" ] || bad=$((bad + 1))
    done
    check "$label: no archive is unpacked before its checksum" "$bad" "0"
done

echo
echo "===== the go tarball has a sum for every architecture it will fetch ====="
# The old expression passed `uname -m` straight into the URL, so an arm64 host
# asked for go1.25.0.linux-aarch64.tar.gz, which has never existed. The case
# statement that replaced it must not grow a default arm that skips the check.
if grep -q 'no pinned go checksum for' "$FULL"; then
    note "an unlisted architecture stops the build" OK
else
    note "an unlisted architecture stops the build" FAIL
fi
check "docker/Dockerfile carries the arm64 sum" \
    "$(arg_of "$FULL" GO_SHA256_ARM64 | wc -c | tr -d ' ')" "65"

echo
echo "===== MaixCDK is pinned to a commit, and the clone uses it ====="
commit=$(arg_of "$FULL" MAIXCDK_COMMIT)
case "$commit" in
    ????????????????????????????????????????)
        case "$commit" in
            *[!0-9a-f]*) note "MAIXCDK_COMMIT is a full hex sha" FAIL ;;
            *)           note "MAIXCDK_COMMIT is a full hex sha" OK ;;
        esac ;;
    *) note "MAIXCDK_COMMIT is a full 40-character sha (got '$commit')" FAIL ;;
esac

# A pin nothing checks out is decoration. The clone and the checkout have to be
# in the same RUN, because a later stage would start from the default branch.
if grep -q 'git checkout --detach "\$MAIXCDK_COMMIT"' "$FULL"; then
    note "the clone checks the pinned commit out" OK
else
    note "the clone checks the pinned commit out" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== every input is pinned and verified ====="
else
    echo "===== $fails check(s) failed ====="
    exit 1
fi
