#!/bin/sh
# Check that S95nanokvm frees the server's log before it needs the space.
#
#   test-restart-space.sh [path-to-S95nanokvm]
#
# The server writes its standard output to /tmp/nanokvm-server.log, and libkvm
# prints on some error paths once per frame. A pipeline that fails in a loop
# therefore writes into tmpfs without a limit. S98vidiag empties the file while
# it reads, but that trim stops when the reader stops, so no other script can
# depend on it.
#
# This was found when S95nanokvm still copied the server into tmpfs. tmpfs holds
# 80892K on the device and /kvmapp/server was 36236K. A flood filled the free
# space, the restart case removed /tmp/server and /tmp/kvm_system, which
# returned 36540K, copied kvm_system back, which took 328K, and copied the
# server, which needed 36236K. That left 24K less than the copy needed.
#
# The server is no longer copied: /tmp/server is a link to /kvmapp/server. The
# kvm_system copy remains, and a flood can leave no room at all, so a copy into
# a full tmpfs still leaves a truncated binary that the case then starts.
#
# The fix is an order, not a size: empty the log first, and the flood's space is
# back before the first copy asks for it. This script checks that order.
#
# tools/vidiag/spacetest.sh replays the old sequence on a real tmpfs of the
# device's size.
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
    fn && /^[ \t]*refresh_kvm_system[ \t]*$/ { print "copy", NR }
' "$S95")

echo "===== the log is emptied before the first copy ====="
# start_services copies kvm_system into tmpfs, and the server's log shares that
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
# A path written out twice drifts. The reader in S98vidiag follows one path, so
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
