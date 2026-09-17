#!/bin/sh
# Check how S95nanokvm stages the server at /tmp/server.
#
#   test-server-link.sh [path-to-S95nanokvm]
#
# /tmp/server used to be a copy of /kvmapp/server in tmpfs, 32.7MB of RAM the
# kernel could never reclaim. It is now a link to /kvmapp/server. The path stays
# because S98supervise reads it as the operator's intent and kvm_system starts
# the server through it.
#
# Two things can go wrong and both are expensive. A link is also a way into the
# boot SD card, so staging or removing it must never write through it. And a
# real copy left at the path, by kvm_system's update path or by an older
# S95nanokvm, is 32MB of RAM that nothing runs again, so staging has to replace
# a directory as well as a link.
S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-link.sh <S95nanokvm>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

build_tree() {
    src="$WORK/kvmapp/server"
    rm -rf "$WORK/kvmapp" "$WORK/tmp"
    mkdir -p "$src/dl_lib" "$src/web/assets" "$WORK/tmp"
    echo binary  > "$src/NanoKVM-Server"
    echo library > "$src/dl_lib/libkvm.so"
    echo asset   > "$src/web/assets/index.js"
}

run() {
    SERVER_SRC="$WORK/kvmapp/server" SERVER_DST="$WORK/tmp/server" \
        sh "$S95" "$@" > "$WORK/run.log" 2>&1
}

installed_intact() {
    [ -f "$WORK/kvmapp/server/NanoKVM-Server" ] \
        && [ -f "$WORK/kvmapp/server/dl_lib/libkvm.so" ] \
        && [ -f "$WORK/kvmapp/server/web/assets/index.js" ]
}

echo "===== staging links the installed tree ====="

build_tree
if run __stage_server; then
    note "the script accepts __stage_server" OK
else
    note "the script accepts __stage_server" FAIL
    sed 's/^/    /' "$WORK/run.log" | head -5
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi

[ -L "$WORK/tmp/server" ] \
    && note "/tmp/server is a link" OK \
    || note "/tmp/server is a link" FAIL
[ "$(readlink "$WORK/tmp/server")" = "$WORK/kvmapp/server" ] \
    && note "the link points at the installed tree" OK \
    || note "the link points at the installed tree" FAIL

# The web root is derived from the executable's path, and the supervisor tests
# the binary through the link, so every one of these must resolve.
for want in NanoKVM-Server dl_lib/libkvm.so web/assets/index.js; do
    [ -f "$WORK/tmp/server/$want" ] \
        && note "$want resolves through the link" OK \
        || note "$want resolves through the link" FAIL
done

echo
echo "===== staging again replaces the link ====="
# start_services stages on every restart, and the link is already there. ln -s
# onto an existing link to a directory creates a link inside the target, which
# here is the SD card.
run __stage_server
[ -L "$WORK/tmp/server" ] \
    && note "still a link after a second staging" OK \
    || note "still a link after a second staging" FAIL
[ -e "$WORK/kvmapp/server/server" ] \
    && note "nothing was written into the installed tree" FAIL \
    || note "nothing was written into the installed tree" OK

echo
echo "===== staging replaces a real copy ====="
# kvm_system's update path leaves one: rm -r /tmp/server, cp -r /kvmapp/server /tmp/.
rm -f "$WORK/tmp/server"
mkdir -p "$WORK/tmp/server/web"
echo stale > "$WORK/tmp/server/NanoKVM-Server"
run __stage_server
[ -L "$WORK/tmp/server" ] \
    && note "a copied directory is replaced by the link" OK \
    || note "a copied directory is replaced by the link" FAIL

echo
echo "===== a stop removes the link and not the installed tree ====="
# stop_services removes the path with rm -rf once the server is gone. The
# supervisor needs the path gone, and the SD card needs the tree left alone.
run __stage_server
SERVER_DST="$WORK/tmp/server" SYSTEM_DST="$WORK/tmp/kvm_system" sh -c ". /dev/stdin" <<STOP > "$WORK/stop.log" 2>&1
$(sed -n "/^stop_services() {$/,/^}$/p" "$S95")
pidof() { return 1; }
stop_process() { :; }
stop_audio_capture() { :; }
stop_services
STOP
if [ -e "$WORK/tmp/server" ] || [ -L "$WORK/tmp/server" ]; then
    note "the link is removed" FAIL
else
    note "the link is removed" OK
fi
if installed_intact; then
    note "the installed tree survives the stop" OK
else
    note "the installed tree survives the stop" FAIL
fi

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
    exit 1
fi
