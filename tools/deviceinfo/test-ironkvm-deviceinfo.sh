#!/bin/sh
# Drive kvmapp/system/ironkvm-deviceinfo against scratch files.
#
# deviceinfo is data, and the reader must never run it. The cases that matter
# most are the ones where a shell would have done something: a command
# substitution, a variable, a backslash. The reader must refuse each of them,
# and the marker file must never appear.
set -u
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
READER=${1:-$ROOT/kvmapp/system/ironkvm-deviceinfo}
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
fails=0
t() { if [ "$2" = 0 ]; then echo "ok   - $1"; else echo "FAIL - $1"; fails=1; fi; }

get() { DEVICEINFO_PATHS="$WORK/a $WORK/b" sh "$READER" get "$1" 2>"$WORK/err"; }

cat > "$WORK/a" <<'EOF'
# comment
DEVICE=sipeed-nanokvm

LOADER_ALIASES="ld-one.so.1 ld-two.so.1"
SLOT_A=2
EOF

[ "$(get DEVICE)" = sipeed-nanokvm ]; t "a bare value" $?
[ "$(get LOADER_ALIASES)" = "ld-one.so.1 ld-two.so.1" ]; t "a quoted value keeps its spaces" $?
[ "$(get SLOT_A)" = 2 ]; t "a number is a string" $?

get NOPE >/dev/null; st=$?
[ "$st" = 1 ] && grep -q 'NOPE is not set in' "$WORK/err"; t "a missing key exits 1 and names the key (got $st)" $?

# The first path is used when it exists; the second only when it does not.
echo 'DEVICE=from-b' > "$WORK/b"
[ "$(get DEVICE)" = sipeed-nanokvm ]; t "the first file wins" $?
mv "$WORK/a" "$WORK/a.off"
[ "$(get DEVICE)" = from-b ]; t "the second file is the fallback" $?
rm "$WORK/b"
get DEVICE >/dev/null; st=$?
[ "$st" = 3 ] && grep -q "no deviceinfo at $WORK/a or $WORK/b" "$WORK/err"
t "no file exits 3 and names both paths (got $st)" $?
mv "$WORK/a.off" "$WORK/a"

bad() {
    desc=$1; line=$2
    printf '%s\n' 'DEVICE=ok' "$line" > "$WORK/a"
    get DEVICE >/dev/null; st=$?
    [ "$st" = 4 ] && grep -q "$WORK/a:2: not KEY=value" "$WORK/err"
    t "$desc is refused (got $st)" $?
    [ ! -e "$WORK/ran" ]; t "and nothing ran for it" $?
}
bad "a command substitution"        'X=$(touch '"$WORK"'/ran)'
bad "backticks in quotes"           'X="`touch '"$WORK"'/ran`"'
bad "a variable"                    'X=$HOME'
bad "a backslash"                   'X=a\b'
bad "a space in a bare value"       'X=a b'
bad "a lowercase key"               'x=1'
bad "a key with a dash"             'SLOT-A=2'
bad "a key that starts with a digit" '2X=1'
bad "an unterminated quote"         'X="abc'
bad "a single-quoted value"         "X='abc'"
bad "export in front"               'export X=1'
bad "a repeated key"                'DEVICE=twice'

# A value may be empty, and that is different from absent.
printf 'DEVICE=ok\nEMPTY=\nQEMPTY=""\n' > "$WORK/a"
get EMPTY >/dev/null; t "an empty bare value exits 0" $?
get QEMPTY >/dev/null; t "an empty quoted value exits 0" $?

# part turns a partition number into a device path. The separator belongs to the
# disk: mmcblk and nvme name partitions with a p, sd and vd do not.
part() { DEVICEINFO_PATHS="$WORK/a" sh "$READER" part "$1" 2>"$WORK/err"; }

printf 'DISK=/dev/mmcblk0\nSLOT_A=2\nDATA=6\nBOOT_PART=1\n' > "$WORK/a"
[ "$(part SLOT_A)" = /dev/mmcblk0p2 ]; t "a numbered disk gets a p" $?
[ "$(part DATA)" = /dev/mmcblk0p6 ];  t "another partition on the same disk" $?

printf 'DISK=/dev/sda\nSLOT_A=2\n' > "$WORK/a"
[ "$(part SLOT_A)" = /dev/sda2 ]; t "a letter-named disk gets no p" $?

printf 'DISK=/dev/nvme0n1\nSLOT_A=2\n' > "$WORK/a"
[ "$(part SLOT_A)" = /dev/nvme0n1p2 ]; t "nvme gets a p" $?

printf 'DISK=/dev/mmcblk0\nSLOT_A=2\n' > "$WORK/a"
part RECOVERY >/dev/null; st=$?
[ "$st" = 1 ]; t "a partition key that is not set exits 1 (got $st)" $?

printf 'SLOT_A=2\n' > "$WORK/a"
part SLOT_A >/dev/null; st=$?
[ "$st" = 1 ]; t "no DISK exits 1 (got $st)" $?

printf 'DISK=/dev/mmcblk0\nSLOT_A=x\n' > "$WORK/a"
part SLOT_A >/dev/null; st=$?
[ "$st" = 5 ] && grep -q 'SLOT_A=x is not a partition number' "$WORK/err"
t "a partition number that is not a number exits 5 (got $st)" $?

printf 'DISK=mmcblk0\nSLOT_A=2\n' > "$WORK/a"
part SLOT_A >/dev/null; st=$?
[ "$st" = 5 ] && grep -q 'DISK=mmcblk0 is not a path' "$WORK/err"
t "a DISK that is not a path exits 5 (got $st)" $?

printf 'DISK=/dev/mmcblk0\nSLOT_A=0\n' > "$WORK/a"
part SLOT_A >/dev/null; st=$?
[ "$st" = 5 ]; t "partition 0 exits 5 (got $st)" $?

# Usage.
DEVICEINFO_PATHS="$WORK/a" sh "$READER" >/dev/null 2>&1; st=$?
[ "$st" = 2 ]; t "no arguments exits 2 (got $st)" $?

# The description the application tarball carries, at kvmapp/system/deviceinfo.
#
# It is read only on a board that does not run an IronKVM image, which means a
# board on Sipeed's firmware, whose card has the stock layout: p1 boot, p2 root,
# p3 data, and no sixth partition. It was a byte copy of the image's own
# description, so it declared DATA=6. S01fs then named /dev/mmcblk0p6 on a card
# whose highest partition is p3, /data was not mounted, and upstream's own
# behaviour of mounting p3 was lost.
#
# So it states what the device is and nothing about how a card is laid out. The
# scripts find the data partition on the card itself.
SHIPPED=$ROOT/kvmapp/system/deviceinfo
shipped() { DEVICEINFO_PATHS="$SHIPPED" sh "$READER" "$@" 2>"$WORK/err"; }

# Whether a description states a key that is true only of a card the build laid
# out. The check below runs it on the shipped file, and the mutation at the end
# runs it on a copy that declares one, because a check that cannot fail proves
# nothing.
declares_layout() {
    for k in SLOT_A SLOT_B RECOVERY DATA KERNEL_IN_SLOT SLOT_SIZE_MIB DATA_START
    do
        DEVICEINFO_PATHS="$1" sh "$READER" get "$k" > /dev/null 2>&1 && return 0
    done
    return 1
}

[ -f "$SHIPPED" ]; t "the tarball carries a description" $?

# The reader checks the whole file before it prints anything, so one read that
# works is a file every other read works on.
[ "$(shipped get DEVICE)" = sipeed-nanokvm ]; t "the reader accepts every line of it" $?
[ "$(shipped get DISK)" = /dev/mmcblk0 ]; t "it names the disk" $?
[ "$(shipped get DATA_FS)" = exfat ]; t "it names the data filesystem" $?

# Partition 1 is the boot partition of both layouts, so this one number is a
# fact about the device.
[ "$(shipped part BOOT_PART)" = /dev/mmcblk0p1 ]; t "it names the boot partition" $?

# And every key that is true only of a card the build laid out is gone.
for key in SLOT_A SLOT_B RECOVERY DATA KERNEL_IN_SLOT SLOT_SIZE_MIB DATA_START
do
    shipped get "$key" > /dev/null; st=$?
    [ "$st" = 1 ]; t "it declares no $key (got $st)" $?
done

! declares_layout "$SHIPPED"; t "so the shipped description declares no layout" $?

# The mutation. A description that takes the layout keys back is the state this
# whole rule exists to prevent, and the check above has to notice it. Done on a
# copy: the shipped file is never written.
cp "$SHIPPED" "$WORK/mutated"
printf '%s\n' DATA=6 >> "$WORK/mutated"
if cmp -s "$SHIPPED" "$WORK/mutated"; then
    echo "FAIL - the mutation did not apply"; fails=1
else
    declares_layout "$WORK/mutated"; t "and a description that takes DATA back is caught" $?
fi

sh -n "$READER"; t "the reader parses" $?
exit $fails
