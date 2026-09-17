#!/bin/sh
# Check that S01zram sizes the swap device from the device description.
#
#   test-zram-disksize.sh [path-to-S01zram]
#
# 96M was chosen for a 256MB NanoKVM. A board with more or less memory carries
# its own ZRAM_MIB in deviceinfo, and this script must not hold one board's
# number.
#
# The fallback matters as much as the value. S01zram runs from rcS, long before
# the server, the log or the OLED exist, so a board whose description is missing
# has no way to report the problem. It gets 96M and swap rather than a clean
# refusal and none.
#
# The real reader is used behind a stub on PATH, not a hand-written fake. A fake
# would answer whatever this suite expects, including for a key that S01zram
# asks for under the wrong name.
S01=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S01zram}
[ -f "$S01" ] || { echo "usage: test-zram-disksize.sh <S01zram>"; exit 1; }

READER=${READER:-$(cd "$(dirname "$0")/../.." && pwd)/kvmapp/system/ironkvm-deviceinfo}
[ -f "$READER" ] || { echo "needs kvmapp/system/ironkvm-deviceinfo"; exit 2; }

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "===== the swap device is as big as the board says ====="

# Run the shipped text rather than a copy of it.
sed -n '/^# --- disk size ---/,/^# --- end disk size ---/p' "$S01" > "$work/block.sh"

if [ ! -s "$work/block.sh" ]; then
    note "the disk size block can be extracted" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi
note "the disk size block can be extracted" OK

mkdir -p "$work/bin"
cat > "$work/bin/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$work/deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$work/bin/ironkvm-deviceinfo"

# Write a description for the next case. Each argument is one KEY=value line.
deviceinfo() { printf '%s\n' "$@" > "$work/deviceinfo"; }

# The block ends in the assignment, so sourcing it is what this script does at
# boot. DEVINFO names the reader, exactly as the init script's own default does.
size() { DEVINFO=${1:-$work/bin/ironkvm-deviceinfo} sh -c ". $work/block.sh; echo \$ZRAM_DISKSIZE" 2>/dev/null; }

# This board. The description in devices/sipeed-nanokvm says 96.
deviceinfo DEVICE=sipeed-nanokvm RAM_MIB=256 ZRAM_MIB=96
got=$(size)
[ "$got" = 96M ] \
    && note "this board's 96 gives 96M" OK \
    || note "this board gave '$got', want 96M" FAIL

# Another board, another number. This is the whole point of reading a
# description: no edit here describes a second board.
deviceinfo DEVICE=other RAM_MIB=128 ZRAM_MIB=64
got=$(size)
[ "$got" = 64M ] \
    && note "ZRAM_MIB=64 gives 64M" OK \
    || note "ZRAM_MIB=64 gave '$got', want 64M" FAIL

deviceinfo DEVICE=big RAM_MIB=1024 ZRAM_MIB=384
got=$(size)
[ "$got" = 384M ] \
    && note "ZRAM_MIB=384 gives 384M" OK \
    || note "ZRAM_MIB=384 gave '$got', want 384M" FAIL

echo
echo "===== a description that cannot be used still leaves swap ====="

# A description with no ZRAM_MIB. The reader exits 1 and prints nothing.
deviceinfo DEVICE=sipeed-nanokvm RAM_MIB=256
got=$(size)
[ "$got" = 96M ] \
    && note "a description without ZRAM_MIB falls back to 96M" OK \
    || note "a description without ZRAM_MIB gave '$got'" FAIL

# No description at all. The reader exits 3.
rm -f "$work/deviceinfo"
got=$(size)
[ "$got" = 96M ] \
    && note "no description at all falls back to 96M" OK \
    || note "no description at all gave '$got'" FAIL

# No reader on PATH either. This is a board running the application tarball on
# an image that installs no reader, and it must still swap.
got=$(size "$work/bin/absent-reader")
[ "$got" = 96M ] \
    && note "no reader falls back to 96M" OK \
    || note "no reader gave '$got'" FAIL

deviceinfo DEVICE=sipeed-nanokvm ZRAM_MIB=lots
got=$(size)
[ "$got" = 96M ] \
    && note "a ZRAM_MIB that is not a number falls back to 96M" OK \
    || note "ZRAM_MIB=lots gave '$got'" FAIL

# Zero is a number and is still not a size. A zero-length device takes mkswap
# and then holds nothing, which is no swap wearing the name of swap.
deviceinfo DEVICE=sipeed-nanokvm ZRAM_MIB=0
got=$(size)
[ "$got" = 96M ] \
    && note "ZRAM_MIB=0 falls back to 96M" OK \
    || note "ZRAM_MIB=0 gave '$got'" FAIL

# The number is what is checked, not the string built from it. ZRAM_MIB=1M
# passes a check that allows the suffix and then writes 1MM to sysfs, which the
# kernel rejects: the board loses its swap to a value that looked almost right.
deviceinfo DEVICE=sipeed-nanokvm ZRAM_MIB=1M
got=$(size)
[ "$got" = 96M ] \
    && note "ZRAM_MIB with a suffix of its own falls back to 96M" OK \
    || note "ZRAM_MIB=1M gave '$got', which sysfs refuses" FAIL

echo
echo "===== the reader is found on a board that has none on PATH ====="
#
# An image built by ironkvm-dist installs the reader at
# /usr/bin/ironkvm-deviceinfo, so the bare name resolves through PATH. A board
# that runs Sipeed's firmware installs the application tarball instead, and its
# install.sh comes from the official base tarball, which the fork cannot change:
# nothing puts the reader on PATH there. The tarball carries the reader at
# /kvmapp/system/ironkvm-deviceinfo, so that is where the script looks next.
#
# This script degrades to 96M without it, so the fault is quiet. S01fs, which
# resolves the reader the same way, leaves /data unmounted instead.

# A second stub stands at the path the tarball installs, and answers from a
# description of its own, so every case below says which reader replied.
mkdir -p "$work/tarball" "$work/other"
cat > "$work/tarball/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$work/tarball-deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$work/tarball/ironkvm-deviceinfo"
TARBALL="$work/tarball/ironkvm-deviceinfo"
printf '%s\n' DEVICE=tarball ZRAM_MIB=384 > "$work/tarball-deviceinfo"

# A third reader, for the case that names one outright.
cat > "$work/other/reader" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$work/other-deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$work/other/reader"
printf '%s\n' DEVICE=other ZRAM_MIB=128 > "$work/other-deviceinfo"

# A PATH with no reader on it, which is what a stock firmware board has.
BARE=/usr/bin:/bin

# Each case drives the block with DEVINFO unset, so the script's own default is
# what resolves the reader. The prefixes go on sh rather than on a function,
# because an assignment in front of a function call outlives the call in ash.
deviceinfo DEVICE=sipeed-nanokvm RAM_MIB=256 ZRAM_MIB=64
got=$(PATH="$work/bin:$PATH" DEVINFO_TARBALL="$TARBALL" \
    sh -c ". $work/block.sh; echo \$ZRAM_DISKSIZE" 2>/dev/null)
[ "$got" = 64M ] \
    && note "a reader on PATH is the one that answers" OK \
    || note "a reader on PATH gave '$got', want 64M" FAIL

got=$(PATH="$BARE" DEVINFO_TARBALL="$TARBALL" \
    sh -c ". $work/block.sh; echo \$ZRAM_DISKSIZE" 2>/dev/null)
[ "$got" = 384M ] \
    && note "no reader on PATH falls back to the tarball's copy" OK \
    || note "no reader on PATH gave '$got', want 384M" FAIL

got=$(PATH="$work/bin:$PATH" DEVINFO="$work/other/reader" DEVINFO_TARBALL="$TARBALL" \
    sh -c ". $work/block.sh; echo \$ZRAM_DISKSIZE" 2>/dev/null)
[ "$got" = 128M ] \
    && note "a DEVINFO from the environment beats both" OK \
    || note "a DEVINFO from the environment gave '$got', want 128M" FAIL

# A DEVINFO that names a reader which is not there keeps the behaviour this
# block already has for a missing reader: 96M and swap. The suites drive that
# case this way, so the fallback must not rescue it.
got=$(PATH="$work/bin:$PATH" DEVINFO="$work/no-such-reader" DEVINFO_TARBALL="$TARBALL" \
    sh -c ". $work/block.sh; echo \$ZRAM_DISKSIZE" 2>/dev/null)
[ "$got" = 96M ] \
    && note "a DEVINFO that names nothing still falls back to 96M" OK \
    || note "a DEVINFO that names nothing gave '$got'" FAIL

# The fallback is in /kvmapp because that is the only path the application
# tarball owns. A file written into the vendor root filesystem is removed by the
# next firmware update and belongs to nobody.
if grep -q '^DEVINFO_TARBALL=${DEVINFO_TARBALL:-/kvmapp/system/ironkvm-deviceinfo}$' "$S01"; then
    note "the fallback is the tarball's own path" OK
else
    note "the fallback is not /kvmapp/system/ironkvm-deviceinfo" FAIL
fi

if grep -v '^[[:space:]]*#' "$S01" | grep -q '/usr/'; then
    note "no line outside a comment names a path in /usr" FAIL
else
    note "no line outside a comment names a path in /usr" OK
fi

echo
echo "===== the size is not written into the script ====="

if grep -qE '^ZRAM_DISKSIZE=[0-9]' "$S01"; then
    note "the script names no fixed size of its own" FAIL
else
    note "the script names no fixed size of its own" OK
fi

if grep -qE '^ZRAM_DISKSIZE=\$\{ZRAM_DISKSIZE:-\$\(zram_disksize\)\}$' "$S01"; then
    note "the size comes from zram_disksize" OK
else
    note "the size no longer comes from zram_disksize" FAIL
fi

# start() writes $ZRAM_DISKSIZE to sysfs, so a rename there would leave the
# board on whatever the kernel defaults to.
if grep -q 'echo "\$ZRAM_DISKSIZE" > "\$ZRAM_SYSFS/disksize"' "$S01"; then
    note "the size is what start writes to disksize" OK
else
    note "start no longer writes ZRAM_DISKSIZE to disksize" FAIL
fi

echo
echo "===== the script still parses ====="
sh -n "$S01" 2>/dev/null && note "sh -n accepts S01zram" OK || note "sh -n accepts S01zram" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
