#!/bin/sh
# Drive kvmapp/system/init.d/S98netbird against a scratch tree and a stub
# daemon.
#
#   test-netbird.sh [path-to-S98netbird]
#
# The script starts and stops the daemon by pid file, without
# start-stop-daemon, as S98tailscaled does. And the peer's identity lives on
# /data, so both slots are the same peer on the NetBird network.
S98=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S98netbird}
[ -f "$S98" ] || { echo "usage: test-netbird.sh <S98netbird>"; exit 1; }

for tool in pgrep mktemp; do
    command -v "$tool" >/dev/null 2>&1 || { echo "needs $tool"; exit 2; }
done
[ -d /proc/self ] || { echo "needs /proc"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
cleanup() {
    for p in $(pgrep -f "$work/bin/netbird"); do kill -9 "$p" 2>/dev/null; done
    rm -rf "$work"
}
trap cleanup EXIT

# restart runs "$0" itself, which needs an executable file. A checkout on a
# Windows host mounted into a container does not promise the bit, so the test
# runs a copy.
cp "$S98" "$work/S98netbird" && chmod 755 "$work/S98netbird"
S98=$work/S98netbird

mkdir -p "$work/bin"
# The stub records its environment and arguments and then stays up the way
# `netbird service run` does. Its command line names netbird, which is what the
# script checks a pid against. STUB_EXIT makes it die at once.
cat > "$work/bin/netbird" <<'STUB'
#!/bin/sh
echo "NB_STATE_DIR=$NB_STATE_DIR GOMEMLIMIT=$GOMEMLIMIT $*" >> "$STUB_LOG"
[ -n "$STUB_EXIT" ] && exit 1
trap 'exit 0' TERM
while :; do sleep 1; done
STUB
chmod 755 "$work/bin/netbird"

reset_tree() {
    rm -rf "$work/root" "$work/stub.log"
    mkdir -p "$work/root/data" "$work/root/etc/kvm"
    : > "$work/stub.log"
    if [ "$1" = mounted ]; then
        printf '/dev/mmcblk0p6 %s exfat rw 0 0\n' "$work/root/data" > "$work/root/mounts"
    else
        : > "$work/root/mounts"
    fi
}

run() {
    NETBIRD="$work/bin/netbird" \
    PIDFILE="$work/root/var/run/netbird.pid" \
    SOCKET="$work/root/var/run/netbird.sock" \
    LOGFILE="$work/root/var/log/netbird.log" \
    STATEDIR="$work/root/var/lib/netbird" \
    KVMDIR="$work/root/etc/kvm" MOUNTS="$work/root/mounts" DATA="$work/root/data" \
    CGROUP_PROCS="$work/root/no-cgroup/cgroup.procs" \
    START_WAIT=1 STOP_WAIT=5 STUB_LOG="$work/stub.log" \
        sh "$S98" "$@" > "$work/out" 2>&1
}

daemons() { pgrep -f "$work/bin/netbird service run" | wc -l | tr -d ' '; }
started_with() { grep -- "service run" "$work/stub.log" | tail -1; }

shared="$work/root/data/identity-system/netbird"
slot="$work/root/var/lib/netbird"

echo "===== start and stop ====="
reset_tree mounted
run start
grep -q '^Starting netbird: OK$' "$work/out" \
    && note "start reports OK" OK \
    || { note "start reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(daemons)" = 1 ] \
    && note "one daemon is running" OK \
    || note "one daemon is running (got $(daemons))" FAIL
pid=$(cat "$work/root/var/run/netbird.pid" 2>/dev/null)
[ -n "$pid" ] && grep -q netbird "/proc/$pid/cmdline" 2>/dev/null \
    && note "the pid file names the daemon" OK \
    || note "the pid file names the daemon" FAIL
case "$(started_with)" in
    *"--daemon-addr unix://$work/root/var/run/netbird.sock "*) note "it serves the CLI on the socket" OK ;;
    *) note "it serves the CLI on the socket" FAIL; echo "    $(started_with)" ;;
esac
case "$(started_with)" in
    *"--log-file $work/root/var/log/netbird.log"*) note "it logs to the log file" OK ;;
    *) note "it logs to the log file" FAIL ;;
esac
case "$(started_with)" in
    *"GOMEMLIMIT=512MiB "*) note "with no group and no setting the limit is 512MiB" OK ;;
    *) note "with no group and no setting the limit is 512MiB" FAIL; echo "    $(started_with)" ;;
esac

run start
[ "$(daemons)" = 1 ] \
    && note "a second start does not start a second daemon" OK \
    || note "a second start does not start a second daemon (got $(daemons))" FAIL

old=$(cat "$work/root/var/run/netbird.pid" 2>/dev/null)
run restart
new=$(cat "$work/root/var/run/netbird.pid" 2>/dev/null)
[ "$(daemons)" = 1 ] && [ -n "$new" ] && [ "$new" != "$old" ] \
    && note "restart leaves one new daemon" OK \
    || note "restart leaves one new daemon (got $(daemons), $old -> $new)" FAIL

run stop
grep -q '^Stopping netbird: OK$' "$work/out" \
    && note "stop reports OK" OK \
    || { note "stop reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(daemons)" = 0 ] \
    && note "no daemon is left running" OK \
    || note "no daemon is left running (got $(daemons))" FAIL
[ ! -e "$work/root/var/run/netbird.pid" ] \
    && note "the pid file is removed" OK \
    || note "the pid file is removed" FAIL

echo
echo "===== the identity lives on /data ====="
reset_tree mounted
run start; run stop
case "$(started_with)" in
    *"NB_STATE_DIR=$shared "*) note "the profile directory is on /data" OK ;;
    *) note "the profile directory is on /data" FAIL; echo "    $(started_with)" ;;
esac
case "$(started_with)" in
    *"--config $shared/config.json "*) note "and so is the config" OK ;;
    *) note "and so is the config" FAIL ;;
esac

reset_tree unmounted
run start; run stop
case "$(started_with)" in
    *"NB_STATE_DIR=$slot "*"--config $slot/config.json "*) note "without /data it stays on the slot" OK ;;
    *) note "without /data it stays on the slot" FAIL; echo "    $(started_with)" ;;
esac
[ ! -d "$work/root/data/identity-system" ] \
    && note "and nothing is written under an unmounted /data" OK \
    || note "and nothing is written under an unmounted /data" FAIL

echo
echo "===== a pid file it cannot trust ====="
reset_tree mounted
sleep 60 &
bystander=$!
mkdir -p "$work/root/var/run"
echo "$bystander" > "$work/root/var/run/netbird.pid"
run stop
kill -0 "$bystander" 2>/dev/null \
    && note "stop does not signal a process that is not netbird" OK \
    || note "stop does not signal a process that is not netbird" FAIL
echo "$bystander" > "$work/root/var/run/netbird.pid"
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
grep -q '^Starting netbird: FAIL$' "$work/out" \
    && note "start reports FAIL" OK \
    || { note "start reports FAIL" FAIL; sed 's/^/    /' "$work/out"; }
[ ! -e "$work/root/var/run/netbird.pid" ] \
    && note "and leaves no pid file behind" OK \
    || note "and leaves no pid file behind" FAIL

echo
echo "===== no binary ====="
reset_tree mounted
NETBIRD="$work/bin/absent" sh "$S98" start > "$work/out" 2>&1
st=$?
[ "$st" -ne 0 ] && grep -q "not found" "$work/out" \
    && note "a missing binary is reported and fails" OK \
    || note "a missing binary is reported and fails (status $st)" FAIL

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
