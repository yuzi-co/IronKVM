#!/bin/sh
# Check that stopping the services never deletes a tree something is still
# running from.
#
#   test-server-stop.sh [path-to-S95nanokvm]
#
# stop_services removes the runtime copies in tmpfs, and start_services only
# rebuilds the server's copy when no server is running:
#
#     if ! pidof NanoKVM-Server >/dev/null 2>&1; then
#         copy_server
#         "$SERVER_DST/NanoKVM-Server" &
#     fi
#
# Those two agree only while the stop actually stops things. On this board it
# does not always: a server that outlives the stop keeps answering, the start
# then sees a live pidof and skips the copy, and what is left is a process
# serving from a directory that no longer exists.
#
# The failure is quiet and it does not look like a filesystem problem. The API
# keeps answering, because routing is in memory. Every static file returns 404,
# because the web root is derived from the executable's own path. Observed on a
# device on 2026-08-19 after a protocol change restarted the server: /api
# answered 401 and / answered 404, and /proc/<pid>/exe read
# "/tmp/server/NanoKVM-Server (deleted)".
set -u

S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-stop.sh <S95nanokvm>"; exit 1; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Lift every top-level function, so a helper the code under test calls is never
# missing and never a local reimplementation of it.
lift() {
    cat > "$work/f.sh" <<'STUB'
# Which processes are still alive after the stop is what these cases vary.
pidof() {
    case "$1" in
        NanoKVM-Server) [ "${SERVER_ALIVE:-no}" = yes ] ;;
        kvm_system)     [ "${SYSTEM_ALIVE:-no}" = yes ] ;;
        *)              false ;;
    esac
}
killall() { :; }
sleep()   { :; }
STUB
    sed -n '/^[a-z_][a-z_]*() *{$/,/^}$/p' "$S95" \
        | sed "s|/tmp/|$work/tmp/|g" >> "$work/f.sh"
}

lift
if ! grep -q '^stop_services() {' "$work/f.sh"; then
    note "S95nanokvm defines stop_services" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "S95nanokvm defines stop_services" OK

# run <server-alive> <system-alive>; leaves both trees' fates in $work/tmp.
run() {
    rm -rf "$work/tmp"
    mkdir -p "$work/tmp/server/web/assets" "$work/tmp/kvm_system"
    : > "$work/tmp/server/NanoKVM-Server"
    : > "$work/tmp/kvm_system/kvm_system"

    (
        SERVER_ALIVE=$1
        SYSTEM_ALIVE=$2
        SERVER_DST=$work/tmp/server
        SYSTEM_DST=$work/tmp/kvm_system
        export SERVER_ALIVE SYSTEM_ALIVE SERVER_DST SYSTEM_DST
        . "$work/f.sh"
        stop_services
    ) > "$work/out" 2>&1
}

present() {
    if [ -e "$1" ]; then note "$2" OK; else note "$2" FAIL; fi
}
absent() {
    if [ -e "$1" ]; then note "$2" FAIL; else note "$2" OK; fi
}

echo "===== a tree whose process survived the stop is kept ====="

run yes no
present "$work/tmp/server/NanoKVM-Server" \
        "a surviving server keeps its binary"
present "$work/tmp/server/web/assets" \
        "a surviving server keeps its web root"

run no yes
present "$work/tmp/kvm_system/kvm_system" \
        "a surviving kvm_system keeps its binary"

echo
echo "===== a tree whose process is gone is removed ====="

# The guard has to be a guard, not the deletion removed. tmpfs is RAM on this
# board, and a runtime copy left behind is pinned memory for as long as it is up.
run no no
absent "$work/tmp/server" \
       "a stopped server's tree is removed"
absent "$work/tmp/kvm_system" \
       "a stopped kvm_system's tree is removed"

echo
echo "===== each tree is judged by its own process ====="

# One rm for both trees means a server that survives also saves kvm_system's
# copy, and the reverse. They are separate processes.
run yes no
absent "$work/tmp/kvm_system" \
       "kvm_system's tree goes when only the server survived"

run no yes
absent "$work/tmp/server" \
       "the server's tree goes when only kvm_system survived"

echo
if [ "$fails" -eq 0 ]; then
    echo "===== stop_services keeps what is still running ====="
else
    echo "$fails case(s) FAILED"
    exit 1
fi
