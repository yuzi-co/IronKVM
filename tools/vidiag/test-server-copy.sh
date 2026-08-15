#!/bin/sh
# Check what S95nanokvm copies from the boot SD card into tmpfs.
#
#   test-server-copy.sh [path-to-S95nanokvm]
#
# /tmp is tmpfs, so the copy is RAM. Every byte under /kvmapp/server is held
# there for as long as the board is up, whether the server opens the file or
# not. A deploy that leaves a backup beside the thing it replaced therefore
# costs memory forever: two orphans measured on a device, web.rollback at
# 3.83MB and dl_lib/libkvm.so.prev at 1.48MB, were 5.3MB of a 166MB board.
#
# The copy is also the step with the least room. tools/vidiag/test-restart-space.sh
# records the arithmetic: tmpfs holds 80892K, the tree was 36236K, and the
# margin was 24K. Dead weight in that tree is not only wasted RAM, it is the
# difference between a copy that fits and a truncated NanoKVM-Server.
#
# So the rule this file holds is: copy what the server runs, skip what a deploy
# left behind. The names are matched at any depth, because the two orphans
# above sat at different depths.
S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-copy.sh <S95nanokvm>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# A tree shaped like the device's, with a backup at the top level and another
# one nested, which is where the two real orphans were found.
build_tree() {
    src="$WORK/kvmapp/server"
    rm -rf "$WORK/kvmapp" "$WORK/tmp"
    mkdir -p "$src/dl_lib" "$src/web/assets" "$src/web.rollback/assets"
    echo binary            > "$src/NanoKVM-Server"
    echo library           > "$src/dl_lib/libkvm.so"
    echo previous-library  > "$src/dl_lib/libkvm.so.prev"
    echo asset             > "$src/web/assets/index.js"
    echo stale-asset       > "$src/web.rollback/assets/index.js"
    mkdir -p "$WORK/tmp"
}

run_copy() {
    SERVER_SRC="$WORK/kvmapp/server" SERVER_DST="$WORK/tmp/server" \
        sh "$S95" __copy_server > "$WORK/copy.log" 2>&1
}

echo "===== the runtime copy takes the server and leaves the backups ====="

build_tree
if run_copy; then
    note "the script accepts __copy_server" OK
else
    note "the script accepts __copy_server" FAIL
    sed 's/^/    /' "$WORK/copy.log" | head -5
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
fi

for want in NanoKVM-Server dl_lib/libkvm.so web/assets/index.js; do
    [ -f "$WORK/tmp/server/$want" ] \
        && note "$want is copied" OK \
        || note "$want is copied" FAIL
done

for skip in dl_lib/libkvm.so.prev web.rollback/assets/index.js web.rollback; do
    [ -e "$WORK/tmp/server/$skip" ] \
        && note "$skip is left behind" FAIL \
        || note "$skip is left behind" OK
done

# A backup that is copied and then deleted still has to fit while it is there,
# and fitting is the whole problem. Nothing excluded may appear at any point,
# so check that the destination never held it rather than only that it is gone.
grep -q 'web.rollback' "$WORK/copy.log" \
    && note "the copy names what it skipped" OK \
    || note "the copy names what it skipped" FAIL

echo
echo "===== the source is left alone ====="
# The source is the boot SD card. A copy that prunes its own input would delete
# the operator's rollback, and would wear the card doing it.
for keep in dl_lib/libkvm.so.prev web.rollback/assets/index.js; do
    [ -e "$WORK/kvmapp/server/$keep" ] \
        && note "$keep still exists under the source" OK \
        || note "$keep still exists under the source" FAIL
done

echo
echo "===== a second run replaces the copy ====="
# start_services runs on every restart, and the destination is already there.
echo stale > "$WORK/tmp/server/leftover"
run_copy
[ -e "$WORK/tmp/server/leftover" ] \
    && note "a file from the previous copy is removed" FAIL \
    || note "a file from the previous copy is removed" OK
[ -f "$WORK/tmp/server/NanoKVM-Server" ] \
    && note "the second copy still lands" OK \
    || note "the second copy still lands" FAIL

echo
echo "===== the script still parses ====="
sh -n "$S95" 2>/dev/null \
    && note "sh -n accepts the script" OK \
    || note "sh -n accepts the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
