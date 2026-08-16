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
echo "===== modes and ownership are declared, not inherited ====="
#
# The payload is a copy of a Windows checkout on a Linux build host, so its
# modes and its ownership describe the copy and not the image. On 2026-08-16 a
# root A image shipped /kvmapp/server/NanoKVM-Server at 0644 owned by the
# builder's uid, which is a slot that boots and then cannot start its server.
#
# So the image declares both: the manifest names the mode, the builder writes
# the owner, and a gate catches an add that forgot.

mkdir -p "$WORK/payload/scripts"
printf '#!/bin/sh\necho declared\n' > "$WORK/payload/scripts/S01declared"
chmod 644 "$WORK/payload/scripts/S01declared"

cat > "$WORK/mode.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     scripts/S01declared       /etc/init.d/S01declared      0755
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST

if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/mode.manifest" "$WORK/payload" 64 "$WORK/mode.img" \
   > "$WORK/mode.log" 2>&1; then
    note "a manifest mode column builds" OK
else
    note "a manifest mode column builds" FAIL
    sed 's/^/    /' "$WORK/mode.log" | tail -20
fi

statline() { debugfs -R "stat $2" "$1" 2>/dev/null | tr -s ' '; }

if [ -f "$WORK/mode.img" ]; then
    statline "$WORK/mode.img" /etc/init.d/S01declared | grep -q 'Mode: 0755' \
        && note "the declared mode reaches the image" OK \
        || note "the declared mode reaches the image, got: $(statline "$WORK/mode.img" /etc/init.d/S01declared | grep -o 'Mode: [0-7]*')" FAIL
else
    note "the declared mode reaches the image" FAIL
fi

# Ownership. Only meaningful where the test itself can chown, which means
# running as root on a filesystem that stores uids. Say SKIP otherwise rather
# than pass quietly.
: > "$WORK/ownprobe"
if chown 1000:1000 "$WORK/ownprobe" 2>/dev/null \
   && [ "$(find "$WORK/ownprobe" -user 1000 | wc -l)" -eq 1 ]; then
    chown 1000:1000 "$WORK/payload/scripts/S00awatchdog"
    sh "$BUILD" "$WORK/base.tar.zst" "$WORK/mode.manifest" "$WORK/payload" 64 "$WORK/own.img" \
       > "$WORK/own.log" 2>&1
    if [ -f "$WORK/own.img" ]; then
        statline "$WORK/own.img" /etc/init.d/S00awatchdog | grep -q 'User: 0 Group: 0' \
            && note "an added file is owned by root, not by the builder" OK \
            || note "an added file is owned by $(statline "$WORK/own.img" /etc/init.d/S00awatchdog | grep -o 'User: [0-9]* Group: [0-9]*')" FAIL
    else
        note "an added file is owned by root, not by the builder" FAIL
        sed 's/^/    /' "$WORK/own.log" | tail -20
    fi
    chown 0:0 "$WORK/payload/scripts/S00awatchdog"
else
    note "an added file is owned by root (this host cannot chown)" SKIP
fi

echo
echo "===== the web UI in the image is the fork's, not the base's ====="
#
# The first root A shipped the fork's server with Sipeed's official 2.5.0 web
# UI, because root.manifest never added a built web and the official one came
# in with the application tarball. The two disagree about the API.
#
# /api/vm/device/virtual returns {enabled, active, cost} per device in this
# fork and a plain boolean upstream. An object is truthy in JavaScript, so the
# official UI drew the virtual disk and the virtual network as permanently ON,
# and every click to turn them off was read by the server as a request to turn
# them on. The switches could not be moved.
#
# A mismatch like that cannot be caught by looking at either half alone, so the
# gate looks for a string only the fork's UI has.

# Built from a copy. An earlier revision added the web fixture to $WORK/base
# itself, and the ownership case further down reuses that base: it then carried
# a /kvmapp with the official web, the new gate refused its build, and a case
# about chown failed for a reason that had nothing to do with chown.
rm -rf "$WORK/base3"
cp -a "$WORK/base" "$WORK/base3"
mkdir -p "$WORK/base3/kvmapp/server/web/assets"
printf 'var a=e.data.disk,b=e.data.network;\n' > "$WORK/base3/kvmapp/server/web/assets/official.js"
( cd "$WORK/base3" && tar --numeric-owner -cf - . | zstd -q -o "$WORK/base3.tar.zst" )

cat > "$WORK/noweb.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base3.tar.zst" "$WORK/noweb.manifest" "$WORK/payload" 64 "$WORK/noweb.img" \
   > "$WORK/noweb.log" 2>&1; then
    note "an image carrying the base's own web UI is refused" FAIL
else
    note "an image carrying the base's own web UI is refused" OK
fi
[ -f "$WORK/noweb.img" ] \
    && note "the refused web build leaves no image behind" FAIL \
    || note "the refused web build leaves no image behind" OK

mkdir -p "$WORK/payload/webdist/assets"
printf 'var t="settings.device.endpoints.cost";\n' > "$WORK/payload/webdist/assets/desktop.js"
cat > "$WORK/web.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     webdist                   /kvmapp/server/web
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base3.tar.zst" "$WORK/web.manifest" "$WORK/payload" 64 "$WORK/web.img" \
   > "$WORK/web.log" 2>&1; then
    note "replacing it with the fork's build passes" OK
else
    note "replacing it with the fork's build passes" FAIL
    sed 's/^/    /' "$WORK/web.log" | tail -20
fi

# A non-merge add must REPLACE the directory. A merge would leave the base's
# hashed asset filenames alongside the fork's, because they never collide.
present2() { debugfs -R "stat $1" "$WORK/web.img" 2>/dev/null | grep -q 'Inode:'; }
present2 /kvmapp/server/web/assets/official.js \
    && note "the base's own assets survived alongside the fork's" FAIL \
    || note "the base's own assets are gone, not merged with" OK

echo
echo "===== the base's own boot path is owned by root ====="
#
# Sipeed's v1.4.3 rootfs ships /etc/init.d, its 24 scripts and /usr/sbin/tailscaled
# owned by uid 1000, which is their build host's user and exists in no passwd
# file on the board. Root ignores the mode when it executes them, so the board
# boots and nothing looks wrong.
#
# It is still a writable-by-a-stranger boot path. The manifest already replaces
# a third of the files in that directory, so this repository owns it either
# way, and an owner that means nothing here should be root.
#
# Only /etc/init.d is corrected. The rest of the base keeps what Sipeed
# shipped, because /etc/bind and /var/www are legitimately owned by named and
# www-data, and a blanket chown would break exactly the cases that are right.

if chown 1000:1000 "$WORK/ownprobe" 2>/dev/null \
   && [ "$(find "$WORK/ownprobe" -user 1000 | wc -l)" -eq 1 ]; then
    rm -rf "$WORK/base2"
    cp -a "$WORK/base" "$WORK/base2"
    chown -R 1000:1000 "$WORK/base2/etc/init.d"
    ( cd "$WORK/base2" && tar --numeric-owner -cf - . | zstd -q -o "$WORK/base2.tar.zst" )

    sh "$BUILD" "$WORK/base2.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/base2.img" \
       > "$WORK/base2.log" 2>&1
    if [ -f "$WORK/base2.img" ]; then
        statline "$WORK/base2.img" /etc/init.d/S50sshd | grep -q 'User: 0 Group: 0' \
            && note "a base script owned by uid 1000 becomes root-owned" OK \
            || note "a base script stayed $(statline "$WORK/base2.img" /etc/init.d/S50sshd | grep -o 'User: [0-9]* Group: [0-9]*')" FAIL
        statline "$WORK/base2.img" /etc/init.d | grep -q 'User: 0 Group: 0' \
            && note "and the directory itself is root-owned" OK \
            || note "the directory itself stayed $(statline "$WORK/base2.img" /etc/init.d | grep -o 'User: [0-9]* Group: [0-9]*')" FAIL
    else
        note "a base script owned by uid 1000 becomes root-owned" FAIL
        sed 's/^/    /' "$WORK/base2.log" | tail -20
    fi
else
    note "the base's boot path is owned by root (this host cannot chown)" SKIP
fi

echo
echo "===== the gates refuse a bad build, and write nothing ====="

# An ELF that is not executable is the 2026-08-16 fault itself. The init.d gate
# does not see it, because the server binary is not an init script.
printf '\177ELF\002\001\001\0\0\0\0\0\0\0\0\0' > "$WORK/payload/scripts/fakebin"
chmod 644 "$WORK/payload/scripts/fakebin"
cat > "$WORK/elf.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     scripts/fakebin           /usr/bin/fakebin
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/elf.manifest" "$WORK/payload" 64 "$WORK/elf.img" \
   > "$WORK/elf.log" 2>&1; then
    note "a non-executable ELF is refused" FAIL
else
    note "a non-executable ELF is refused" OK
fi
[ -f "$WORK/elf.img" ] \
    && note "the refused ELF build leaves no image behind" FAIL \
    || note "the refused ELF build leaves no image behind" OK

# And the same ELF with a declared mode must build, because that is the fix an
# operator reaches for when the gate fires.
cat > "$WORK/elfok.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     scripts/fakebin           /usr/bin/fakebin             0755
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/elfok.manifest" "$WORK/payload" 64 "$WORK/elfok.img" \
   > "$WORK/elfok.log" 2>&1; then
    note "the same ELF with a declared mode builds" OK
else
    note "the same ELF with a declared mode builds" FAIL
    sed 's/^/    /' "$WORK/elfok.log" | tail -20
fi
rm -f "$WORK/payload/scripts/fakebin"

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

# `ls -l -R` is not debugfs syntax. debugfs rejects the -R and prints nothing,
# so the first version of this case compared two empty files and passed by
# doing nothing, for every build, including builds that differed. That is the
# third guard in this repository to rot exactly that way, so the emptiness
# check below is not decoration and neither is the negative control.
inventory() {
    d=$(mktemp -d)
    debugfs -R "rdump / $d" "$1" >/dev/null 2>&1
    ( cd "$d" && find . -printf '%y %m %U %G %s %p\n' 2>/dev/null | sort )
    rm -rf "$d"
}
inventory "$WORK/out.img"   > "$WORK/inv1"
inventory "$WORK/again.img" > "$WORK/inv2"

[ "$(wc -l < "$WORK/inv1")" -gt 5 ] \
    && note "the inventory reads more than nothing" OK \
    || note "the inventory read $(wc -l < "$WORK/inv1") lines, so it compares nothing" FAIL

cmp -s "$WORK/inv1" "$WORK/inv2" \
    && note "two builds produce the same inventory" OK \
    || note "two builds produce different inventories" FAIL

# The negative control. A comparison that cannot tell two different images
# apart is not a comparison, and this is the only thing that proves it can.
cat > "$WORK/extra.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     scripts/S01declared       /etc/init.d/S01declared      0755
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST
sh "$BUILD" "$WORK/base.tar.zst" "$WORK/extra.manifest" "$WORK/payload" 64 "$WORK/extra.img" \
   > "$WORK/extra.log" 2>&1
inventory "$WORK/extra.img" > "$WORK/inv3"
cmp -s "$WORK/inv1" "$WORK/inv3" \
    && note "an image with an extra file reads as identical, so the check is blind" FAIL \
    || note "an image with an extra file reads as different" OK

echo
echo "===== a backup file in /etc/init.d is refused ====="

# rcS runs every file in /etc/init.d whose name begins with S, so a backup left
# beside a script is a second script that runs at every boot. On 2026-08-16 a
# board carried /etc/init.d/S02identity.rollback, an older identity script that
# ran after the current one and bound /etc/kvm a second time. Nothing reported
# it: the board booted, served and looked healthy.
cp "$WORK/payload/scripts/S00awatchdog" "$WORK/payload/scripts/S02identity.rollback"
cat > "$WORK/stray.manifest" <<'MANIFEST'
add     scripts/S00awatchdog          /etc/init.d/S00awatchdog          0755
add     scripts/S02identity.rollback  /etc/init.d/S02identity.rollback  0755
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/stray.manifest" "$WORK/payload" 64 \
      "$WORK/stray.img" > "$WORK/stray.log" 2>&1; then
    note "a build carrying a .rollback in init.d is refused" FAIL
else
    note "a build carrying a .rollback in init.d is refused" OK
fi
grep -q 'stray file in init.d: S02identity.rollback' "$WORK/stray.log" \
    && note "it names the file it refused" OK \
    || note "it names the file it refused" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
