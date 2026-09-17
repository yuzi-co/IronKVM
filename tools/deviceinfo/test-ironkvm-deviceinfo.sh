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

sh -n "$READER"; t "the reader parses" $?
exit $fails
