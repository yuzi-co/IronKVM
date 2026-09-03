#!/bin/sh
# Check that every strided row copy in kvm_mmf.cpp advances by its own counter.
#
#   test-stride-copy.sh [path/to/kvm_mmf.cpp]
#
# Not destructive: the source is only read.
#
# The frame buffers VENC and VPSS hand back are padded, so a row of w pixels
# sits in a row of u32Stride[0] bytes and each of the seven copy loops in this
# file walks the destination by stride * <row>. One of them walked it by
# stride * h, the frame height, which is a constant inside the loop: all
# h * 3 / 2 rows landed on the first row of the chroma plane, the luma plane
# kept whatever the buffer held before, and the encoder read a corrupt picture.
#
# It stayed inside the buffer, so there was no crash to notice, and it only
# fires when the stride differs from the width, which for this board means a
# capture width that is not a multiple of the alignment. 1920 and 1280 are;
# 1366 and 1600 are not.
#
# A C++ unit test would be better and this tree has no harness for one. This is
# the cheap guard that would have caught it: a destination offset that names a
# variable other than the loop counter is the whole defect.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SRC=${1:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== each strided destination walks its own loop counter ====="

# Report one line per loop: "<line> <loop variable> <offset variable>". A
# destination offset more than four lines below its "for" is a shape this
# check does not understand, and the count assertion below catches that.
report=$(awk '
    match($0, /for[ ]*\([ ]*(int|CVI_U32|unsigned|size_t)[ ]+[A-Za-z_][A-Za-z0-9_]*[ ]*=/) {
        line = $0
        sub(/.*for[ ]*\([ ]*(int|CVI_U32|unsigned|size_t)[ ]+/, "", line)
        sub(/[ ]*=.*/, "", line)
        var = line
        at = NR
        next_var = ""
        next
    }
    /u32Stride\[0\][ ]*\*[ ]*[A-Za-z_]/ {
        if (var != "" && NR - at <= 4) {
            off = $0
            sub(/.*u32Stride\[0\][ ]*\*[ ]*/, "", off)
            sub(/[^A-Za-z0-9_].*/, "", off)
            printf "%d %s %s\n", NR, var, off
        }
    }
' "$SRC")

found=$(printf '%s\n' "$report" | grep -c . )

# Seven loops write a strided destination in this file. A change to that count
# means a loop was added, removed, or reshaped past what the scan above reads,
# and either way the result below stops being a statement about the whole file.
if [ "$found" = 7 ]; then
    note "seven strided destination loops found" OK
else
    note "seven strided destination loops found (got $found)" FAIL
    printf '%s\n' "$report" | sed 's/^/      /'
fi

printf '%s\n' "$report" | while read -r line var off; do
    [ -n "$line" ] || continue
    if [ "$var" = "$off" ]; then
        printf '  %-58s %s\n' "line $line walks $off, its own counter" OK
    else
        printf '  %-58s %s\n' "line $line walks $off, not its counter $var" FAIL
    fi
done

# The loop above runs in a subshell, so count the failures again here.
bad=$(printf '%s\n' "$report" | awk 'NF && $2 != $3' | grep -c .)
[ "$bad" -gt 0 ] && fails=$((fails + bad))

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
