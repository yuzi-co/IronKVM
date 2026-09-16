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

# A tool that is absent here is not a defect in what this file checks. Name the
# missing one and stop with status 2, which tools/run-tests.sh counts as
# skipped. The old behaviour ran on and reported a case as FAILED, which reads
# as a broken image or a broken table and teaches whoever sees it to stop
# believing the suite.
need() {
    for _cmd in "$@"
    do
        command -v "$_cmd" >/dev/null 2>&1 && continue
        echo "$(basename "$0"): needs $_cmd, which is not on PATH." >&2
        echo "the release host image carries it:" >&2
        echo "  docker build -t ironkvm-release-host tools/release" >&2
        echo "  docker run --rm -v \"$PWD:/repo\" -w /repo ironkvm-release-host sh $0" >&2
        exit 2
    done
}
need mke2fs e2fsck debugfs tar zstd

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

# A payload directory added whole, of which one file is not wanted. This is the
# shape the vendor library directory has: 38 libraries arrive together and 21 of
# them are opened by nothing, and remove cannot take them out because remove has
# already run by the time add creates them.
mkdir -p "$WORK/payload/libs"
printf 'used\n'   > "$WORK/payload/libs/libkept.so"
printf 'unused\n' > "$WORK/payload/libs/libdead.so"

cat > "$WORK/good.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     libs                      /usr/lib/dl
drop    /usr/lib/dl/libdead.so
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST

echo "===== the builder applies a manifest ====="

if [ ! -x "$BUILD" ] && [ ! -f "$BUILD" ]; then
    note "build-image.sh exists" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "build-image.sh exists" OK

if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/out.img" \
   > "$WORK/build.log" 2>&1; then
    note "the build succeeds" OK
else
    note "the build succeeds" FAIL
    sed 's/^/    /' "$WORK/build.log" | tail -20
    echo; echo "$fails case(s) FAILED"; exit 1
fi

present() { debugfs -R "stat $1" "$WORK/out.img" 2>/dev/null | grep -q 'Inode:'; }

present /etc/init.d/S00awatchdog && note "an added file is present" OK          || note "an added file is present" FAIL
present /etc/kvm/ssh_stop        && note "a removed file is gone" FAIL          || note "a removed file is gone" OK
present /root/notes.txt          && note "a removed directory is gone" FAIL     || note "a removed directory is gone" OK
present /etc/kvm.disk0           && note "a touched file is present" OK         || note "a touched file is present" FAIL
present /etc/init.d/S50sshd      && note "the base survives" OK                 || note "the base survives" FAIL
present /etc/kvm/pwd             && note "an untouched base file survives" OK   || note "an untouched base file survives" FAIL
present /usr/lib/dl/libdead.so   && note "a dropped file is gone" FAIL          || note "a dropped file is gone" OK
present /usr/lib/dl/libkept.so   && note "its siblings survive the drop" OK     || note "its siblings survive the drop" FAIL

echo
echo "===== provenance and device names are recorded ====="
present /etc/slot-manifest      && note "/etc/slot-manifest is written" OK      || note "/etc/slot-manifest is written" FAIL
present /etc/nanokvm-slots.conf && note "/etc/nanokvm-slots.conf is written" OK || note "/etc/nanokvm-slots.conf is written" FAIL

# S01fs makes the data partition on the first boot and needs to be told where it
# goes. Without this line it makes nothing, and the board comes up with no /data
# and the factory root password.
debugfs -R "cat /etc/nanokvm-slots.conf" "$WORK/out.img" 2>/dev/null > "$WORK/slots"
grep -q '^DATA_DEV=/dev/mmcblk0p6$' "$WORK/slots" \
    && note "the slot conf names the data device" OK \
    || note "the slot conf names the data device" FAIL
grep -q '^DATA_START=10543104$' "$WORK/slots" \
    && note "the slot conf carries DATA_START" OK \
    || note "the slot conf carries DATA_START, got: $(grep DATA_START "$WORK/slots" || echo nothing)" FAIL

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

dumpe2fs -h "$WORK/out.img" 2>/dev/null > "$WORK/geom"

# The image is built smaller than the slot and grown by S01fs on the first boot
# of that slot. Without resize_inode it cannot be grown, and the slot keeps an
# eighth of the space it should have with nothing said about why.
grep -q '^Filesystem features:.*resize_inode' "$WORK/geom" \
    && note "the image can be grown to fill its slot" OK \
    || note "the image has no resize_inode, so S01fs could never grow it" FAIL

# mke2fs picks the block size and the inode ratio from the size of the
# filesystem it is making, and this one is made at a fraction of the size it
# will run at. resize2fs can change neither, so a 256 MiB image left to mke2fs
# gets 1 KiB blocks and grows into a 2 GiB filesystem with four times the block
# groups it should have and 128 MiB of inode tables for a thousand files.
bs=$(awk '$1 == "Block" && $2 == "size:" { print $3 }' "$WORK/geom")
[ "$bs" = 4096 ] \
    && note "the block size is 4096, whatever size the image is built at" OK \
    || note "the block size is $bs; resize2fs cannot change it later" FAIL

# One inode per 16 KiB is what mke2fs picks for a 2 GiB filesystem, so the grown
# result has the shape a full-size image used to have.
ratio=$(awk '$1 == "Inode" && $2 == "count:" { c = $3 } $1 == "Block" && $2 == "count:" { b = $3 } END { printf "%d\n", b * 4096 / c }' "$WORK/geom")
[ "$ratio" = 16384 ] \
    && note "one inode per 16 KiB, whatever size the image is built at" OK \
    || note "one inode per $ratio bytes; a grown slot would carry the wrong count" FAIL

# 5% of a 2 GiB slot is 102 MiB held back for a recovery login that this board
# does not have: every process on it is already root.
res=$(awk '/^Reserved block count:/ { print $4 }' "$WORK/geom")
[ "$res" = 0 ] \
    && note "no blocks are reserved for root" OK \
    || note "$res blocks are reserved for root, which nothing here needs" FAIL

# mke2fs sizes the journal from the filesystem, so a 2 GiB image used to get a
# 64 MiB journal: four times what this filesystem needs, written to the card on
# every install, and it stays that size after the grow.
jsize=$(awk '/^Total journal size:/ { print $4 }' "$WORK/geom")
case "$jsize" in
    16M|16384k) note "the journal is pinned at 16 MiB" OK ;;
    *)          note "the journal is $jsize, not the pinned 16 MiB" FAIL ;;
esac

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
echo "===== nothing is world-writable ====="
#
# The first Alpine slot to reach userland, on 2026-09-16, had every directory at
# 0777: its tree had passed through a virtiofs mount, which reports that mode for
# everything. sshd refused to start because /var/empty was world-writable, and no
# gate had looked.

mkdir -p "$WORK/payload/wide/sub"
printf 'data\n' > "$WORK/payload/wide/sub/file"
chmod 777 "$WORK/payload/wide" "$WORK/payload/wide/sub" "$WORK/payload/wide/sub/file"
cat > "$WORK/wide.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
add     wide                      /opt/wide
remove  /etc/kvm/ssh_stop
remove  /root
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/wide.manifest" "$WORK/payload" 64 "$WORK/wide.img" \
   > "$WORK/wide.log" 2>&1; then
    note "a world-writable payload still builds" OK
    for p in /opt/wide /opt/wide/sub /opt/wide/sub/file; do
        m=$(statline "$WORK/wide.img" "$p" | grep -o 'Mode: [0-7]*' | cut -d' ' -f2)
        case "$m" in
            0755|0644) note "the payload's $p loses its write bits ($m)" OK ;;
            *)         note "the payload's $p loses its write bits (got '$m')" FAIL ;;
        esac
    done
else
    note "a world-writable payload still builds" FAIL
    sed 's/^/    /' "$WORK/wide.log" | tail -20
fi

# The base is not rewritten, because it cannot be told apart from a base that
# means it. It is refused instead, and a sticky directory is not a fault.
mkdir -p "$WORK/widebase/etc/init.d" "$WORK/widebase/etc/kvm" "$WORK/widebase/root" \
    "$WORK/widebase/var/empty" "$WORK/widebase/tmp"
printf '#!/bin/sh\necho sshd\n' > "$WORK/widebase/etc/init.d/S50sshd"
printf '' > "$WORK/widebase/etc/kvm/ssh_stop"
chmod 755 "$WORK/widebase/etc/init.d/S50sshd"
chmod 1777 "$WORK/widebase/tmp"
chmod 777 "$WORK/widebase/var/empty"
( cd "$WORK/widebase" && tar --numeric-owner -cf - . | zstd -q -o "$WORK/widebase.tar.zst" )
if sh "$BUILD" "$WORK/widebase.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/widebase.img" \
   > "$WORK/widebase.log" 2>&1; then
    note "a base with a world-writable directory is refused" FAIL
else
    note "a base with a world-writable directory is refused" OK
fi
grep -q 'world-writable: /var/empty' "$WORK/widebase.log" \
    && note "the refusal names the path" OK \
    || note "the refusal names the path" FAIL
grep -q 'world-writable: /tmp' "$WORK/widebase.log" \
    && note "a sticky /tmp is not reported" FAIL \
    || note "a sticky /tmp is not reported" OK
[ -f "$WORK/widebase.img" ] \
    && note "the refused base leaves no image behind" FAIL \
    || note "the refused base leaves no image behind" OK
chmod 755 "$WORK/widebase/var/empty"
( cd "$WORK/widebase" && tar --numeric-owner -cf - . | zstd -q -f -o "$WORK/widebase.tar.zst" )
if sh "$BUILD" "$WORK/widebase.tar.zst" "$WORK/good.manifest" "$WORK/payload" 64 "$WORK/widebase.img" \
   > "$WORK/widebase.log" 2>&1; then
    note "the same base at 0755 builds" OK
else
    note "the same base at 0755 builds" FAIL
    sed 's/^/    /' "$WORK/widebase.log" | tail -20
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

# A drop that matches nothing means the list of files to leave out no longer
# describes the base. That list is 21 vendor libraries chosen because a loader
# was measured and none of them was opened; a silent no-op would let it go on
# being trusted after the directory it describes had changed underneath it.
cat > "$WORK/staledrop.manifest" <<'MANIFEST'
add     scripts/S00awatchdog      /etc/init.d/S00awatchdog
drop    /usr/lib/dl/libgone.so
remove  /etc/kvm/ssh_stop
touch   /etc/kvm.disk0
MANIFEST
if sh "$BUILD" "$WORK/base.tar.zst" "$WORK/staledrop.manifest" "$WORK/payload" 64 "$WORK/staledrop.img" \
   > "$WORK/staledrop.log" 2>&1; then
    note "a drop that matches nothing is refused" FAIL
else
    note "a drop that matches nothing is refused" OK
fi
grep -q 'no longer describes' "$WORK/staledrop.log" \
    && note "it says why the drop was refused" OK \
    || note "it says why the drop was refused" FAIL
[ -f "$WORK/staledrop.img" ] \
    && note "the refused drop build leaves no image behind" FAIL \
    || note "the refused drop build leaves no image behind" OK

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
    exit 1
fi
