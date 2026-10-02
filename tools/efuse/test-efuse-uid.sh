#!/bin/sh
# Hold the three copies of efuse_uid to each other and to the vendor format.
#
#   test-efuse-uid.sh [path-to-init.d]
#
# S03usbdev, S30wifi and S95nanokvm each read the chip id. On the vendor kernel
# they read /sys/class/cvi-base/base_uid, which holds "UID: %08x_%08x". The
# mainline kernel has no cvi-base driver, so each script carries efuse_uid,
# which rebuilds the same line from the eFuse nvmem device. The serial number,
# the USB gadget MACs, the access point address and /device_key are all derived
# from that line, so one copy that drifts gives the board a second identity on
# one kernel and not the other.
#
# Not destructive: every copy runs against a fake tree under mktemp.
DIR=${1:-$(dirname "$0")/../../kvmapp/system/init.d}
SCRIPTS="S03usbdev S30wifi S95nanokvm"

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

lift() {
    sed -n '/^efuse_uid() *{$/,/^}$/p' "$1"
}

first=
for s in $SCRIPTS
do
    [ -f "$DIR/$s" ] || { echo "no $DIR/$s"; exit 2; }
    lift "$DIR/$s" > "$work/$s.fn"
    if [ ! -s "$work/$s.fn" ]
    then
        note "$s defines efuse_uid" FAIL
        continue
    fi
    note "$s defines efuse_uid" OK
    if [ -z "$first" ]
    then
        first=$s
    elif cmp -s "$work/$first.fn" "$work/$s.fn"
    then
        note "$s's copy is the same as $first's" OK
    else
        note "$s's copy differs from $first's" FAIL
    fi
done

efuse=$work/sys/bus/nvmem/devices/cv1800-efuse/nvmem
mkdir -p "${efuse%/*}"

for s in $SCRIPTS
do
    [ -s "$work/$s.fn" ] || continue
    sed "s|/sys/|$work/sys/|g" "$work/$s.fn" > "$work/run.sh"
    echo 'efuse_uid' >> "$work/run.sh"

    # Words 0x0c and 0x10, little-endian, as the shadow registers hold them.
    printf '\000\000\000\000\000\000\000\000\000\000\000\000\147\105\043\001\357\315\253\211' > "$efuse"
    got=$(sh "$work/run.sh"); st=$?
    [ "$st" = 0 ] && [ "$got" = "UID: 01234567_89abcdef" ] \
        && note "$s: the vendor line from the eFuse words" OK \
        || note "$s: got '$got' (status $st), want 'UID: 01234567_89abcdef'" FAIL

    # The text and its newline are what sha512sum hashes, as of the vendor file.
    h=$(sh "$work/run.sh" | sha512sum | head -c 8)
    want=$(printf 'UID: 01234567_89abcdef\n' | sha512sum | head -c 8)
    [ "$h" = "$want" ] && note "$s: hashes like the vendor file" OK \
        || note "$s: hashes to $h, the vendor file to $want" FAIL

    printf '\000\000\000\000\000\000\000\000\000\000\000\000\147\105' > "$efuse"
    got=$(sh "$work/run.sh"); st=$?
    [ "$st" = 1 ] && [ -z "$got" ] && note "$s: a short device is no id" OK \
        || note "$s: a short device gave '$got' (status $st)" FAIL

    rm -f "$efuse"
    got=$(sh "$work/run.sh"); st=$?
    [ "$st" = 1 ] && [ -z "$got" ] && note "$s: no device is no id" OK \
        || note "$s: no device gave '$got' (status $st)" FAIL
done

if [ "$fails" -gt 0 ]
then
    echo "$fails case(s) failed"
    exit 1
fi
echo "all cases passed"
