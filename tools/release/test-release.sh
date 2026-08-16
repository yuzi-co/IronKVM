#!/bin/sh
# Tests for release.sh.
#
# Only the parts that can run without a builder image and without a Sipeed base
# image: the argument handling, the manifest it writes, and the self-check that
# stops a mismatched release from being published.
#
# Needs openssl and tar.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/release.sh"
pass=0
fail=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

for t in openssl tar; do
    command -v "$t" > /dev/null 2>&1 || { echo "SKIP: $t is missing"; exit 0; }
done

echo "release.sh"

# A version that is not X.Y.Z produces a package name the device refuses, and
# the refusal would happen on the device rather than here.
check "a two-component version is refused" \
    "$(sh "$SCRIPT" --dry-run 1.0 > /dev/null 2>&1; echo $?)" "1"

# A prerelease sorts BELOW the release it came from, so the update page would
# offer the older build as an upgrade for ever.
check "a prerelease version is refused" \
    "$(sh "$SCRIPT" --dry-run 1.0.0-rc1 > /dev/null 2>&1; echo $?)" "1"
# The shape check below would also reject it, for having a letter in it. What
# the dedicated check adds is a message that names the actual reason, so the
# message is what proves it ran.
check "it says why a prerelease is refused" \
    "$(sh "$SCRIPT" --dry-run 1.0.0-rc1 2>&1 | grep -c 'prerelease')" "1"

# Build metadata compares EQUAL under semver, so a release carrying it would be
# invisible to every device already on that version.
check "a version with build metadata is refused" \
    "$(sh "$SCRIPT" --dry-run '1.0.0+iron' > /dev/null 2>&1; echo $?)" "1"

check "no version at all is refused" \
    "$(sh "$SCRIPT" --dry-run > /dev/null 2>&1; echo $?)" "1"

# The manifest and the artifact have to agree. This repository has already lost
# three guards to rot, and a feed pointing at a package it does not describe is
# exactly the failure nobody notices until a device tries to install it.
build_fixture() {
    WORK=$(mktemp -d)
    mkdir -p "$WORK/out" "$WORK/stage/ironkvm_1.0.0"
    printf 'payload\n' > "$WORK/stage/ironkvm_1.0.0/marker"
    tar czf "$WORK/out/ironkvm_1.0.0.tar.gz" -C "$WORK/stage" ironkvm_1.0.0

    sum=$(openssl dgst -sha512 -binary "$WORK/out/ironkvm_1.0.0.tar.gz" | openssl base64 -A)
    size=$(wc -c < "$WORK/out/ironkvm_1.0.0.tar.gz" | tr -d ' ')
    cat > "$WORK/out/latest.json" <<EOF
{
  "manifest_version": 2,
  "version": "1.0.0",
  "name": "ironkvm_1.0.0.tar.gz",
  "url": "https://example.com/ironkvm_1.0.0.tar.gz",
  "sha512": "$sum",
  "size": $size,
  "size_bytes": $size,
  "unpacked_size_bytes": $((size * 2))
}
EOF
}

verify() {
    RELEASE_OUT="$WORK/out" sh "$SCRIPT" --verify-only 1.0.0 > "$WORK/verify.out" 2>&1
    echo $?
}

build_fixture
check "a matching manifest passes the self-check" "$(verify)" "0"

# Only the recorded checksum is wrong. The package keeps its size, its name and
# its top directory, so every other check still passes and this one is the only
# thing that can catch it.
#
# Two earlier versions of this case did not isolate it. Appending to the tarball
# changed the size, and corrupting it in place broke gzip so the top-directory
# check fired first. Both passed with the checksum comparison deleted.
sed -i 's|"sha512": ".*"|"sha512": "'"$(printf 'wrong' | openssl base64 -A)"'"|' \
    "$WORK/out/latest.json"
check "a mismatched sha512 fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# The updater derives the directory it expects from the package name, so a
# tarball whose top directory disagrees unpacks into something it refuses.
build_fixture
rm -f "$WORK/out/ironkvm_1.0.0.tar.gz"
mkdir -p "$WORK/stage/wrong_name"
printf 'payload\n' > "$WORK/stage/wrong_name/marker"
tar czf "$WORK/out/ironkvm_1.0.0.tar.gz" -C "$WORK/stage" wrong_name
sum=$(openssl dgst -sha512 -binary "$WORK/out/ironkvm_1.0.0.tar.gz" | openssl base64 -A)
size=$(wc -c < "$WORK/out/ironkvm_1.0.0.tar.gz" | tr -d ' ')
sed -i "s|\"sha512\": \".*\"|\"sha512\": \"$sum\"|; s|\"size\": [0-9]*|\"size\": $size|; s|\"size_bytes\": [0-9]*|\"size_bytes\": $size|" "$WORK/out/latest.json"
check "a wrong tarball top directory fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# A manifest naming a different package than the one built would send every
# device to a file that is not there.
build_fixture
sed -i 's|"name": "ironkvm_1.0.0.tar.gz"|"name": "ironkvm_9.9.9.tar.gz"|' "$WORK/out/latest.json"
check "a manifest naming another package fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# A size that disagrees with the artifact makes the device refuse the download
# after spending the bandwidth.
build_fixture
sed -i 's|"size_bytes": [0-9]*|"size_bytes": 12345|' "$WORK/out/latest.json"
check "a mismatched size_bytes fails the self-check" "$(verify)" "1"
rm -rf "$WORK"

# Dry run is what makes the rest of this script testable at all, so it is the
# guard worth asserting hardest.
check "publishing is guarded by the dry-run flag" \
    "$(sed -n '/^publish()/,/^}/p' "$SCRIPT" | grep -c 'DRY_RUN')" "1"
check "nothing outside publish() calls gh" \
    "$(grep -c '^[^#]*gh release' "$SCRIPT")" "1"

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
