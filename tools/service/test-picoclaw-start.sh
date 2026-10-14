#!/bin/sh
# Which PicoClaw S96picoclaw starts, now that the image carries none.
#
#   test-picoclaw-start.sh
#
# The server installs PicoClaw on demand into /data/ironkvm/addons/picoclaw.
# The script starts that copy when it is there, /usr/bin/picoclaw when it is
# not (an install off a distribution image), PICOCLAW_BIN_PATH over both, and
# nothing, successfully, when PicoClaw is not installed. The binaries here are
# stubs that write what they were run as.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
S=$HERE/../../kvmapp/system/init.d/S96picoclaw
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

stub() {   # stub <path>: a binary that records its arguments and leaves
    mkdir -p "$(dirname "$1")"
    printf '#!/bin/sh\necho "$0 $*" > "%s/ran"\n' "$WORK" > "$1"
    chmod 0755 "$1"
}

# run <action> [VAR=value ...]: the script with the work tree's paths.
run() {
    action=$1; shift
    rm -f "$WORK/ran"
    mkdir -p "$WORK/home"
    echo '{}' > "$WORK/home/config.json"
    env PICOCLAW_HOME="$WORK/home" PICOCLAW_LOG_FILE="$WORK/log" KVMDIR="$WORK/kvm" \
        CGROUP_PROCS="$WORK/no-group/cgroup.procs" PICOCLAW_ADDON_BIN="$WORK/addons/picoclaw/picoclaw" \
        "$@" sh "$S" "$action" > "$WORK/out" 2>&1
}

# ran: what the stub recorded, waiting up to five seconds for the gateway,
# which the script starts in the background.
ran() {
    i=0
    until [ -s "$WORK/ran" ] || [ "$i" -ge 50 ]; do i=$((i + 1)); sleep 0.1; done
    cat "$WORK/ran" 2>/dev/null
}

echo "S96picoclaw start"

if [ -e /usr/bin/picoclaw ]; then
    note "not installed: skipped, this host has a /usr/bin/picoclaw" SKIP
else
    run start; st=$?
    check "not installed: start succeeds" "$st" "0"
    check "not installed: it says so" "$(grep -c 'not installed, nothing to start' "$WORK/out")" "1"
    check "not installed: nothing runs" "$(cat "$WORK/ran" 2>/dev/null)" ""
fi

stub "$WORK/addons/picoclaw/picoclaw"
run start; st=$?
check "the add-on's copy: start succeeds" "$st" "0"
check "the add-on's copy is the one started" "$(ran)" "$WORK/addons/picoclaw/picoclaw gateway"

stub "$WORK/other/picoclaw"
run start PICOCLAW_BIN_PATH="$WORK/other/picoclaw"
check "PICOCLAW_BIN_PATH wins over the add-on's copy" "$(ran)" "$WORK/other/picoclaw gateway"

rm -rf "$WORK/addons"
run start PICOCLAW_BIN_PATH="$WORK/missing/picoclaw"; st=$?
check "a PICOCLAW_BIN_PATH that is not there: nothing to start" "$st/$(cat "$WORK/ran" 2>/dev/null)" "0/"

[ "$fails" -eq 0 ] || { echo "$fails failed"; exit 1; }
echo "all passed"
