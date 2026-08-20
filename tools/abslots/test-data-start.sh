#!/bin/sh
# Tests for data-start.sh.
#
# The start sector of the data partition is derived rather than declared, so the
# derivation is the thing that has to be right. A wrong number here puts the
# data partition somewhere no other card has it.
HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/data-start.sh"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== the shipped table gives the sector every existing card uses ====="

# 10543104 is where the data partition sat in the fixed table this replaced. A
# card made the new way and a card made the old way must have one layout, so
# this number is a compatibility assertion and not an arithmetic one.
got=$(sh "$SCRIPT" "$HERE/partition.sfdisk" 2>&1)
[ "$got" = 10543104 ] \
    && note "the shipped table gives 10543104" OK \
    || note "the shipped table gives '$got', want 10543104" FAIL

echo
echo "===== the rule is end of the recovery slot plus 8192 ====="

cat > "$WORK/small.sfdisk" <<'EOF'
label: dos
unit: sectors

1 : start=1,     size=32768, type=c, bootable
2 : start=40960, size=16384, type=83
3 : start=57344, size=16384, type=83
4 : start=73728, size=18432, type=5
5 : start=75776, size=16384, type=83
EOF
got=$(sh "$SCRIPT" "$WORK/small.sfdisk" 2>&1)
[ "$got" = 100352 ] \
    && note "75776 + 16384 + 8192 gives 100352" OK \
    || note "a scaled table gives '$got', want 100352" FAIL

echo
echo "===== a table that still declares a data partition is refused ====="

# A table with a p6 is a table from before this change. Printing a number that
# contradicts it would put the filesystem somewhere the table does not describe.
cat "$WORK/small.sfdisk" > "$WORK/withp6.sfdisk"
printf '6 : start=100352, size=34816, type=7\n' >> "$WORK/withp6.sfdisk"
if sh "$SCRIPT" "$WORK/withp6.sfdisk" > /dev/null 2>&1; then
    note "a table declaring p6 is refused" FAIL
else
    note "a table declaring p6 is refused" OK
fi

echo
echo "===== a table with no recovery slot is refused ====="

cat > "$WORK/nop5.sfdisk" <<'EOF'
label: dos
unit: sectors

1 : start=1,     size=32768, type=c, bootable
2 : start=40960, size=16384, type=83
EOF
if sh "$SCRIPT" "$WORK/nop5.sfdisk" > /dev/null 2>&1; then
    note "a table with no p5 is refused" FAIL
else
    note "a table with no p5 is refused" OK
fi

echo
echo "===== a missing table is refused ====="
if sh "$SCRIPT" "$WORK/absent.sfdisk" > /dev/null 2>&1; then
    note "a missing table is refused" FAIL
else
    note "a missing table is refused" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
