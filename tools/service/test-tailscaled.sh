#!/bin/sh
# Drive kvmapp/system/init.d/S98tailscaled against a scratch tree and a stub
# daemon.
#
#   test-tailscaled.sh [path-to-S98tailscaled]
#
# Two things are held here. The script starts and stops the daemon without
# start-stop-daemon, which Alpine's busybox does not build. And the node's
# state file lives on /data, so both slots are the same machine on the tailnet.
S98=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S98tailscaled}
[ -f "$S98" ] || { echo "usage: test-tailscaled.sh <S98tailscaled>"; exit 1; }

for tool in pgrep mktemp; do
    command -v "$tool" >/dev/null 2>&1 || { echo "needs $tool"; exit 2; }
done
[ -d /proc/self ] || { echo "needs /proc"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
cleanup() {
    for p in $(pgrep -f "$work/bin/tailscaled"); do kill -9 "$p" 2>/dev/null; done
    rm -rf "$work"
}
trap cleanup EXIT

mkdir -p "$work/bin"
# The stub records how it was called and then stays up the way the daemon does.
# Its command line names tailscaled, which is what the script checks a pid
# against. STUB_EXIT makes it die at once, as a daemon with a bad flag would.
cat > "$work/bin/tailscaled" <<'STUB'
#!/bin/sh
case "$1" in
    --version) echo "1.2.3"; exit 0 ;;
    --cleanup) echo cleanup >> "$STUB_LOG"; exit 0 ;;
esac
echo "$*" >> "$STUB_LOG"
[ -n "$STUB_EXIT" ] && exit 1
trap 'exit 0' TERM
while :; do sleep 1; done
STUB
cat > "$work/bin/tailscale" <<'STUB'
#!/bin/sh
echo "tailscale $*" >> "$STUB_LOG"
STUB
chmod 755 "$work/bin/tailscaled" "$work/bin/tailscale"

reset_tree() {
    rm -rf "$work/root" "$work/stub.log"
    mkdir -p "$work/root/data" "$work/root/var/lib/tailscale" "$work/root/etc/kvm"
    : > "$work/stub.log"
    if [ "$1" = mounted ]; then
        printf '/dev/mmcblk0p6 %s exfat rw 0 0\n' "$work/root/data" > "$work/root/mounts"
    else
        : > "$work/root/mounts"
    fi
}

run() {
    TAILSCALED="$work/bin/tailscaled" TAILSCALE="$work/bin/tailscale" \
    PIDFILE="$work/root/var/run/tailscaled.pid" \
    STATEDIR="$work/root/var/lib/tailscale" \
    SOCKETDIR="$work/root/var/run/tailscale" \
    KVMDIR="$work/root/etc/kvm" MOUNTS="$work/root/mounts" DATA="$work/root/data" \
    START_WAIT=1 STOP_WAIT=5 STUB_LOG="$work/stub.log" \
        sh "$S98" "$@" > "$work/out" 2>&1
}

daemons() { pgrep -f "$work/bin/tailscaled --state" | wc -l | tr -d ' '; }
started_with() { grep -- "--state" "$work/stub.log" | tail -1; }

shared="$work/root/data/identity-system/tailscale/tailscaled.state"
slot="$work/root/var/lib/tailscale/tailscaled.state"

echo "===== start and stop, with no start-stop-daemon ====="
reset_tree mounted
PATH="$work/bin:/usr/bin:/bin" run start
grep -q '^Starting.*OK$' "$work/out" \
    && note "start reports OK" OK \
    || { note "start reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(daemons)" = 1 ] \
    && note "one daemon is running" OK \
    || note "one daemon is running (got $(daemons))" FAIL
pid=$(cat "$work/root/var/run/tailscaled.pid" 2>/dev/null)
[ -n "$pid" ] && grep -q tailscaled "/proc/$pid/cmdline" 2>/dev/null \
    && note "the pid file names the daemon" OK \
    || note "the pid file names the daemon" FAIL
grep -q '^tailscale set --accept-dns=false$' "$work/stub.log" \
    && note "it turns off tailscale DNS, as before" OK \
    || note "it turns off tailscale DNS, as before" FAIL

run start
[ "$(daemons)" = 1 ] \
    && note "a second start does not start a second daemon" OK \
    || note "a second start does not start a second daemon (got $(daemons))" FAIL

run stop
grep -q '^Stopping tailscaled: OK$' "$work/out" \
    && note "stop reports OK" OK \
    || { note "stop reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(daemons)" = 0 ] \
    && note "no daemon is left running" OK \
    || note "no daemon is left running (got $(daemons))" FAIL
[ ! -e "$work/root/var/run/tailscaled.pid" ] \
    && note "the pid file is removed" OK \
    || note "the pid file is removed" FAIL
grep -q '^cleanup$' "$work/stub.log" \
    && note "stop still runs tailscaled --cleanup" OK \
    || note "stop still runs tailscaled --cleanup" FAIL

echo
echo "===== the node state lives on /data ====="
reset_tree mounted
run start; run stop
case "$(started_with)" in
    *"--state=$shared "*) note "a new node keeps its state on /data" OK ;;
    *) note "a new node keeps its state on /data" FAIL; echo "    $(started_with)" ;;
esac
case "$(started_with)" in
    *"--statedir=$work/root/var/lib/tailscale "*) note "logs and caches stay on the slot" OK ;;
    *) note "logs and caches stay on the slot" FAIL ;;
esac

# A slot that was already logged in. The first start copies its state, so the
# node does not need a new login.
reset_tree mounted
echo slot-node > "$slot"
run start; run stop
[ "$(cat "$shared" 2>/dev/null)" = slot-node ] \
    && note "an existing login is copied to /data" OK \
    || note "an existing login is copied to /data" FAIL
case "$(started_with)" in
    *"--state=$shared "*) note "and the copy is what the daemon uses" OK ;;
    *) note "and the copy is what the daemon uses" FAIL ;;
esac

# The other slot has an older state of its own. The board's copy wins.
reset_tree mounted
mkdir -p "$(dirname "$shared")"
echo board-node > "$shared"
echo other-slot-node > "$slot"
run start; run stop
[ "$(cat "$shared")" = board-node ] \
    && note "the state on /data is not overwritten by the slot's" OK \
    || note "the state on /data is not overwritten by the slot's" FAIL

# A copy that fails must not cost the login the slot has.
reset_tree mounted
echo slot-node > "$slot"
cat > "$work/bin/cp" <<'STUB'
#!/bin/sh
exit 1
STUB
chmod 755 "$work/bin/cp"
PATH="$work/bin:$PATH" run start; run stop
rm -f "$work/bin/cp"
case "$(started_with)" in
    *"--state=$slot "*) note "a failed copy keeps the slot's own login" OK ;;
    *) note "a failed copy keeps the slot's own login" FAIL; echo "    $(started_with)" ;;
esac

reset_tree unmounted
run start; run stop
case "$(started_with)" in
    *"--state=$slot "*) note "without /data the state stays on the slot" OK ;;
    *) note "without /data the state stays on the slot" FAIL ;;
esac
[ ! -d "$work/root/data/identity-system" ] \
    && note "and nothing is written under an unmounted /data" OK \
    || note "and nothing is written under an unmounted /data" FAIL

echo
echo "===== a pid file it cannot trust ====="
# After a power cut the file can name a pid that another process now holds.
reset_tree mounted
sleep 60 &
bystander=$!
mkdir -p "$work/root/var/run"
echo "$bystander" > "$work/root/var/run/tailscaled.pid"
run stop
kill -0 "$bystander" 2>/dev/null \
    && note "stop does not signal a process that is not tailscaled" OK \
    || note "stop does not signal a process that is not tailscaled" FAIL
echo "$bystander" > "$work/root/var/run/tailscaled.pid"
run start
[ "$(daemons)" = 1 ] \
    && note "start is not fooled into thinking it is running" OK \
    || note "start is not fooled into thinking it is running" FAIL
run stop
kill "$bystander" 2>/dev/null

echo
echo "===== a daemon that dies at once ====="
reset_tree mounted
STUB_EXIT=1 run start
grep -q '^Starting.*FAIL$' "$work/out" \
    && note "start reports FAIL" OK \
    || { note "start reports FAIL" FAIL; sed 's/^/    /' "$work/out"; }
[ ! -e "$work/root/var/run/tailscaled.pid" ] \
    && note "and leaves no pid file behind" OK \
    || note "and leaves no pid file behind" FAIL
grep -q '^tailscale set' "$work/stub.log" \
    && note "and does not configure a daemon that is not there" FAIL \
    || note "and does not configure a daemon that is not there" OK

echo
echo "===== the script still parses ====="
sh -n "$S98" 2>/dev/null \
    && note "sh -n accepts the script" OK \
    || note "sh -n rejects the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
