#!/bin/sh
# Check the partition table before it is ever applied to a card.
#
#   test-partition.sh [path-to-partition.sfdisk]
#
# Needs sfdisk. Run it in the container.
#
# The table is applied once, to the only card. p1 holds fip.bin and boot.sd, and
# the boot ROM finds them by reading the first FAT partition, so an entry that
# moves p1, resizes it, changes its type or drops its boot flag makes the board
# unbootable by every recovery in this design. That is the case this file exists
# for. The rest is arithmetic, and arithmetic is still worth checking when it is
# applied once and cannot be undone.
#
# The official v1.4.3 release image has the same p1: start sector 1, 32768
# sectors, type c, bootable. So this is not a local convention, it is what a
# NanoKVM card looks like.
TABLE=${1:-$(dirname "$0")/partition.sfdisk}
[ -f "$TABLE" ] || { echo "usage: test-partition.sh <partition.sfdisk>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# The card, measured: 60506112 sectors of 512 bytes, 28.85 GiB.
DISK_SECTORS=60506112

echo "===== the table applies to a disk of the card's size ====="

truncate -s $((DISK_SECTORS * 512)) "$WORK/disk.img"
if sfdisk --no-reread --no-tell-kernel "$WORK/disk.img" < "$TABLE" > "$WORK/apply.log" 2>&1; then
    note "sfdisk accepts the table" OK
else
    note "sfdisk accepts the table" FAIL
    sed 's/^/    /' "$WORK/apply.log" | tail -12
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
fi

sfdisk -d "$WORK/disk.img" > "$WORK/back" 2>/dev/null

# sfdisk -d prints "device1 : start= 1, size= 32768, type=c, bootable".
# Read the value back rather than trusting what was written: the point of this
# file is what sfdisk actually did, not what the input said.
field() {
    awk -v p="$1" -v k="$2" '
        $0 ~ (p " *:") {
            n = split($0, t, "[ ,]+")
            for (i = 1; i <= n; i++) if (t[i] == k "=") { print t[i+1]; exit }
            for (i = 1; i <= n; i++) if (t[i] ~ "^" k "=") { sub("^" k "=", "", t[i]); print t[i]; exit }
        }' "$WORK/back"
}

echo
echo "===== p1 is untouched, which is the one that cannot be undone ====="

[ "$(field 1 start)" = 1 ]    && note "p1 starts at sector 1" OK    || note "p1 starts at sector $(field 1 start), want 1" FAIL
[ "$(field 1 size)" = 32768 ] && note "p1 is 32768 sectors" OK      || note "p1 is $(field 1 size) sectors, want 32768" FAIL
[ "$(field 1 type)" = c ]     && note "p1 keeps type c" OK          || note "p1 has type $(field 1 type), want c" FAIL
grep -q '1 *:.*bootable' "$WORK/back" && note "p1 keeps the boot flag" OK || note "p1 lost the boot flag" FAIL

echo
echo "===== the roots are equal, and everything after p1 is aligned ====="

[ "$(field 2 size)" = "$(field 3 size)" ] \
    && note "root A and root B are the same size" OK \
    || note "root A is $(field 2 size), root B is $(field 3 size)" FAIL

[ "$(field 2 size)" = 4194304 ] && note "each root is 2.00 GiB" OK || note "each root is $(field 2 size) sectors, want 4194304" FAIL
[ "$(field 5 size)" = 2097152 ] && note "recovery is 1.00 GiB" OK  || note "recovery is $(field 5 size) sectors, want 2097152" FAIL

# 4 MiB is the erase block size that matters on this card. p1 cannot be aligned
# because it cannot move.
for p in 2 3 5 6; do
    s=$(field "$p" start)
    [ -n "$s" ] && [ $((s % 8192)) -eq 0 ] \
        && note "p$p starts on a 4 MiB boundary" OK \
        || note "p$p starts at ${s:-nothing}, not a 4 MiB boundary" FAIL
done

echo
echo "===== nothing overlaps and nothing runs off the end ====="

last=$(( $(field 6 start) + $(field 6 size) ))
[ "$last" -le "$DISK_SECTORS" ] \
    && note "the last partition ends within the disk" OK \
    || note "the last partition ends at $last, disk holds $DISK_SECTORS" FAIL

# The extended container is excluded: it is meant to span its logicals, so
# including it reports an overlap on a table that is correct.
#
# sfdisk -d pads its numbers, so "start=" and its value split into two tokens
# while "type=5" stays as one. Matching "type=" and then reading the next token
# therefore never fires, which is how this check first reported an overlap on a
# table whose every range was right.
overlap=$(awk '
    /: *start=/ {
        n = split($0, t, "[ ,]+")
        s = 0; z = 0; ext = 0
        for (i = 1; i <= n; i++) {
            if (t[i] == "start=") s = t[i+1] + 0
            if (t[i] == "size=")  z = t[i+1] + 0
            if (t[i] == "type=5" || (t[i] == "type=" && t[i+1] == "5")) ext = 1
        }
        if (!ext && z > 0) print s, s + z - 1
    }' "$WORK/back" | sort -n | awk 'NR > 1 && $1 <= prev { print "overlap" } { prev = $2 }')
[ -z "$overlap" ] && note "no two data partitions overlap" OK || note "partitions overlap" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
