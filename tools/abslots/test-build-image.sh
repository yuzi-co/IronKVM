#!/bin/sh
# Check that a built image contains what the manifest says and nothing else.
#
#   test-build-image.sh [path-to-build-image.sh]
#
# Needs mke2fs, e2fsck, debugfs, tar and zstd. Run it in the container.
#
# Every gate here names a fault this board has actually had:
#
#   a CRLF init script            exits 127 at boot with nothing logged
#   a missing /etc/kvm.disk0      makes the first boot run mkfs.exfat on /data
#   a surviving /etc/kvm/ssh_stop makes S50sshd exit 0 having started nothing
#   an absent S00awatchdog        is half of why the 2026-08-15 card came out
#
# The ssh_stop case is not hypothetical. That file is present in the official
# v1.4.3 rootfs this fork builds on, so the manifest's remove line is what
# stands between a build and a board with no ssh.
BUILD=${1:-$(dirname "$0")/build-image.sh}
[ -f "$BUILD" ] || { echo "usage: test-build-image.sh <build-image.sh>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# A base shaped like the real one: it carries both known traps and a directory
# of operator state that must not survive into an image.
mkdir -p "$WORK/base/etc/init.d" "$WORK/base/etc/kvm" "$WORK/base/root" "$WORK/base/usr/bin"
printf '#!/bin/sh\necho sshd\n'   > "$WORK/base/etc/init.d/S50sshd"
printf ''                        > "$WORK/base/etc/kvm/ssh_stop"
printf 'hashed\n'                > "$WORK/base/etc/kvm/pwd"
printf 'operator notes\n'        > "$WORK/base/root/notes.txt"
printf 'binary\n'                > "$WORK/base/usr/bin/true"
chmod 755 "$WORK/base/etc/init.d/S50sshd"
( cd "$WORK/base" && tar --numeric-owner -cf - . | zstd -q -o "$WORK/base.tar.zst" )

mkdir -p "$WORK/payload/scripts"
printf '#!/bin/sh\necho watchdog\n' > "$WORK/payload/scripts/S00awatchdog"
chmod 755 "$WORK/payload/scripts/S00awatchdog"

cat > "$WORK/good.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST

echo "===== the builder applies a manifest ====="

if [ ! -x "$BUILD" ] && [ ! -f "$BUILD" ]; then
    note "build-image.sh exists" FAIL
    echo; echo "$fails case(s) FAILED"; exit "$fails"
fi
note "build-image.sh exists" OK

if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/out.img" \
   > "$WORK/build.log" 2>&1; then
    note "the build succeeds" OK
else
    note "the build succeeds" FAIL
    sed 's/^/    /' "$WORK/build.log" | tail -20
    echo; echo "$fails case(s) FAILED"; exit "$fails"
fi

present() { debugfs -R "stat $1" "$WORK/out.img" 2>/dev/null | grep -q 'Inode:'; }

present /etc/init.d/S00awatchdog && note "an added file is present" OK          || note "an added file is present" FAIL
present /etc/kvm/ssh_stop        && note "a removed file is gone" FAIL          || note "a removed file is gone" OK
present /root/notes.txt          && note "a removed directory is gone" FAIL     || note "a removed directory is gone" OK
present /etc/kvm.disk0           && note "a touched file is present" OK         || note "a touched file is present" FAIL
present /etc/init.d/S50sshd      && note "the base survives" OK                 || note "the base survives" FAIL
present /etc/kvm/pwd             && note "an untouched base file survives" OK   || note "an untouched base file survives" FAIL

echo
echo "===== provenance and device names are recorded ====="
present /etc/slot-manifest      && note "/etc/slot-manifest is written" OK      || note "/etc/slot-manifest is written" FAIL
present /etc/nanokvm-slots.conf && note "/etc/nanokvm-slots.conf is written" OK || note "/etc/nanokvm-slots.conf is written" FAIL

debugfs -R "cat /etc/slot-manifest" "$WORK/out.img" 2>/dev/null > "$WORK/prov"
grep -q 'basesha ' "$WORK/prov"     && note "provenance records the base hash" OK     || note "provenance records the base hash" FAIL
grep -q 'manifestsha ' "$WORK/prov" && note "provenance records the manifest hash" OK || note "provenance records the manifest hash" FAIL

echo
echo "===== the filesystem is sound ====="
e2fsck -fn "$WORK/out.img" > "$WORK/fsck.log" 2>&1 \
    && note "e2fsck finds no errors" OK \
    || note "e2fsck finds no errors" FAIL

dumpe2fs -h "$WORK/out.img" 2>/dev/null | grep -q 'has_journal' \
    && note "the image keeps its journal" OK \
    || note "the image has no journal, which a power cut would punish" FAIL

echo
echo "===== the gates refuse a bad build, and write nothing ====="

# A CRLF init script must stop the build. It exits 127 on the device with
# nothing logged that names the cause.
printf '#!/bin/sh\r\necho watchdog\r\n' > "$WORK/payload/scripts/S00awatchdog"
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/crlf.img" \
   > "$WORK/crlf.log" 2>&1; then
    note "a CRLF init script is refused" FAIL
else
    note "a CRLF init script is refused" OK
fi
[ -f "$WORK/crlf.img" ] \
    && note "the refused build leaves no image behind" FAIL \
    || note "the refused build leaves no image behind" OK
printf '#!/bin/sh\necho watchdog\n' > "$WORK/payload/scripts/S00awatchdog"
chmod 755 "$WORK/payload/scripts/S00awatchdog"

# A manifest that forgets to remove ssh_stop must stop the build, because the
# official base ships it.
cat > "$WORK/nostop.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/nostop.manifest" "$WORK/payload" 64 "$WORK/nostop.img" \
   > "$WORK/nostop.log" 2>&1; then
    note "a surviving ssh_stop is refused" FAIL
else
    note "a surviving ssh_stop is refused" OK
fi

# A manifest that forgets /etc/kvm.disk0 must stop the build, because the first
# boot would run mkfs.exfat on the data partition.
cat > "$WORK/nodisk0.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
remove  /etc/kvm/ssh_stop
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/nodisk0.manifest" "$WORK/payload" 64 "$WORK/nodisk0.img" \
   > "$WORK/nodisk0.log" 2>&1; then
    note "a missing /etc/kvm.disk0 is refused" FAIL
else
    note "a missing /etc/kvm.disk0 is refused" OK
fi

# An add whose source is not in the payload must stop the build rather than
# silently produce an image without it.
cat > "$WORK/missing.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     scripts/nothere           /etc/init.d/S99nothere
remove  /etc/kvm/ssh_stop
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/missing.manifest" "$WORK/payload" 64 "$WORK/missing.img" \
   > "$WORK/missing.log" 2>&1; then
    note "an add with no source is refused" FAIL
else
    note "an add with no source is refused" OK
fi

echo
echo "===== two builds agree on content ====="

sh "$BUILD" "$WORK/base.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/again.img" \
   > "$WORK/again.log" 2>&1
inventory() {
    debugfs -R "ls -l -R /" "$1" 2>/dev/null | awk '{ $1=""; $6=""; $7=""; $8=""; print }' | sort
}
inventory "$WORK/out.img"   > "$WORK/inv1"
inventory "$WORK/again.img" > "$WORK/inv2"
cmp -s "$WORK/inv1" "$WORK/inv2" \
    && note "two builds produce the same inventory" OK \
    || note "two builds produce different inventories" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
