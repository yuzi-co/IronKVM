#!/bin/sh
# Check that S95nanokvm frees the server's log before it needs the space.
#
#   test-restart-space.sh [path-to-S95nanokvm]
#
# The server writes its standard output to /tmp/nanokvm-server.log, and libkvm
# prints on some error paths once per frame. A pipeline that fails in a loop
# therefore writes into tmpfs without a limit. S99vidiag empties the file while
# it reads, but that trim stops when the reader stops, so no other script can
# depend on it.
#
# tmpfs holds 80892K on the device and /kvmapp/server is 36236K. A flood fills
# the free space. The restart case then removes /tmp/server and /tmp/kvm_system,
# which returns 36540K, copies kvm_system back, which takes 328K, and copies the
# server, which needs 36236K. That leaves 24K less than the copy needs. The
# margin is not engineered: it is the difference between two unrelated sizes.
#
# A copy that runs out of space leaves a truncated NanoKVM-Server, and the case
# starts it anyway. The KVM is then down until somebody logs in.
#
# The fix is an order, not a size: empty the log first, and the flood's space is
# back before the first copy asks for it. This script checks that order.
#
# tools/vidiag/spacetest.sh replays the whole sequence on a real tmpfs of the
# device's size. Run that by hand when the numbers change.
S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-restart-space.sh <S95nanokvm>"; exit 1; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# Report the line numbers of the two events that matter: emptying the log, and
# the first copy that consumes tmpfs. Both live in start_services(), so the
# order is a property of that function.
#
# This used to track case labels instead. When the copies moved into
# start_services() the labels stopped matching, and the check reported both
# cases as "copies nothing into tmpfs" and passed. A guard that skips itself
# reads exactly like a guard that holds, so the scope is named here rather than
# inferred: an awk that finds no event at all is a failure below, not a skip.
events=$(awk '
    /^start_services\(\)/                 { fn = 1 }
    fn && /^\}/                           { fn = 0 }
    fn && /^[ \t]*: > "\$SERVER_LOG"/     { print "empty", NR }
    fn && /^[ \t]*cp -r \/kvmapp\//       { print "copy", NR }
    fn && /^[ \t]*copy_server[ \t]*$/     { print "copy", NR }
' "$S95")

echo "===== the log is emptied before the first copy ====="
# start_services copies 36MB into tmpfs, and the server's log shares that
# space. Emptying after the copy returns the room too late.
empty=$(echo "$events" | awk '$1 == "empty" { print $2; exit }')
copy=$(echo "$events"  | awk '$1 == "copy"  { print $2; exit }')

if [ -z "$copy" ]; then
    note "start_services copies into tmpfs" FAIL
elif [ -z "$empty" ]; then
    note "start_services empties the log" FAIL
elif [ "$empty" -lt "$copy" ]; then
    note "start_services empties at line $empty, copies at line $copy" OK
else
    note "start_services empties at line $empty, but copies at line $copy" FAIL
fi

echo
echo "===== both cases reach that order ====="
# One ordering covers start and restart only for as long as both go through
# start_services. A case that grew its own copy would not be seen above.
for blk in start restart; do
    calls=$(awk -v b="$blk" '
        $0 ~ "^  " b "\\)" { inblk = 1; next }
        inblk && /^  [a-z*_]+\)/ { inblk = 0 }
        inblk && /^[ \t]*start_services[ \t]*$/ { n++ }
        END { print n + 0 }
    ' "$S95")
    [ "$calls" -ge 1 ] \
        && note "$blk calls start_services" OK \
        || note "$blk calls start_services" FAIL
done

echo
echo "===== the log is named once ====="
# A path written out twice drifts. The reader in S99vidiag follows one path, so
# a second spelling here means the collector reads a file nobody writes.
defs=$(grep -c '^SERVER_LOG=' "$S95")
[ "$defs" = 1 ] && note "SERVER_LOG is defined once" OK \
                || note "SERVER_LOG is defined $defs times" FAIL

lit=$(grep -c '/tmp/nanokvm-server\.log' "$S95")
[ "$lit" = 1 ] && note "the path appears only in that definition" OK \
               || note "the path is written out $lit times" FAIL

empties=$(echo "$events" | grep -c '^empty ')
[ "$empties" = 1 ] && note "start_services empties it once" OK \
                   || note "start_services empties it $empties time(s), want 1" FAIL

echo
echo "===== the script still parses ====="
if sh -n "$S95" 2>/dev/null; then
    note "sh -n accepts the script" OK
else
    note "sh -n rejects the script" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
