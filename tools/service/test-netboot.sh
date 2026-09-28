#!/bin/sh
# Drive kvmapp/system/init.d/S85netboot against a scratch tree and a stub
# dnsmasq.
#
#   test-netboot.sh [path-to-S85netboot]
#
# The script runs two dnsmasq instances, one on the USB link for S03usbdev and
# one for proxy DHCP on the LAN, each by its own pid file. It decides nothing:
# the configuration files the server writes say what is on.
S85=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S85netboot}
[ -f "$S85" ] || { echo "usage: test-netboot.sh <S85netboot>"; exit 1; }

for tool in pgrep mktemp; do
    command -v "$tool" >/dev/null 2>&1 || { echo "needs $tool"; exit 2; }
done
[ -d /proc/self ] || { echo "needs /proc"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
cleanup() {
    for p in $(pgrep -f "$work/bin/dnsmasq"); do kill -9 "$p" 2>/dev/null; done
    rm -rf "$work"
}
trap cleanup EXIT

# restart runs "$0" itself, which needs an executable file. A checkout on a
# Windows host mounted into a container does not promise the bit, so the test
# runs a copy.
cp "$S85" "$work/S85netboot" && chmod 755 "$work/S85netboot"
S85=$work/S85netboot

mkdir -p "$work/bin"
# The stub records its arguments and then stays up the way dnsmasq does with
# --keep-in-foreground. Its command line names dnsmasq, which is what the
# script checks a pid against. STUB_EXIT makes it die at once, as dnsmasq does
# on a configuration it refuses.
cat > "$work/bin/dnsmasq" <<'STUB'
#!/bin/sh
echo "$*" >> "$STUB_LOG"
[ -n "$STUB_EXIT" ] && exit 1
trap 'exit 0' TERM
while :; do sleep 1; done
STUB
chmod 755 "$work/bin/dnsmasq"

reset_tree() {
    for p in $(pgrep -f "$work/bin/dnsmasq"); do kill -9 "$p" 2>/dev/null; done
    rm -rf "$work/root" "$work/stub.log"
    mkdir -p "$work/root/conf" "$work/root/cgroup" "$work/root/data/tftp"
    # As on /data: exFAT with fmask 0077 makes every file root's alone.
    echo boot > "$work/root/data/tftp/boot.ipxe"
    chmod 600 "$work/root/data/tftp/boot.ipxe"
    chmod 700 "$work/root/data/tftp"
    : > "$work/stub.log"
    for c in "$@"; do
        echo "# $c" > "$work/root/conf/$c.conf"
    done
}

run() {
    DNSMASQ="${DNSMASQ_BIN:-$work/bin/dnsmasq}" \
    CONFDIR="$work/root/conf" RUN="$work/root/run"     TFTP_SRC="$work/root/data/tftp" TFTP="$work/root/var/netboot/tftp" \
    CGROUP_PROCS="$work/root/cgroup/cgroup.procs" \
    START_WAIT=1 STOP_WAIT=5 STUB_LOG="$work/stub.log" \
        sh "$S85" "$@" > "$work/out" 2>&1
}

instances() { pgrep -f "$work/bin/dnsmasq --keep-in-foreground --conf-file=$work/root/conf/$1.conf" | wc -l | tr -d ' '; }
started_with() { grep -- "--conf-file=$work/root/conf/$1.conf" "$work/stub.log" | tail -1; }
has() {
    case " $(started_with "$1") " in
        *" $2 "*) note "$3" OK ;;
        *) note "$3" FAIL; echo "    $(started_with "$1")" ;;
    esac
}

echo "===== the USB link ====="
reset_tree usb
: > "$work/root/cgroup/cgroup.procs"
run start-usb usb1 172.31.255.2 255.255.255.252
st=$?
[ "$st" = 0 ] && grep -q '^Starting dnsmasq on usb1: OK$' "$work/out" \
    && note "start-usb reports OK and returns 0" OK \
    || { note "start-usb reports OK and returns 0 (status $st)" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(instances usb)" = 1 ] \
    && note "one link dnsmasq is running" OK \
    || note "one link dnsmasq is running (got $(instances usb))" FAIL
has usb "--interface=usb1" "it serves the interface S03usbdev names"
has usb "--dhcp-range=172.31.255.2,172.31.255.2,255.255.255.252,10d" \
    "it leases the host address alone, with the link's mask"
has usb "--dhcp-leasefile=$work/root/run/usb.leases" "its leases go to the run directory"
has usb "--log-facility=$work/root/run/usb.log" "and so does its log"
has usb "--pid-file=$work/root/run/usb.pid" "and its pid file"
pid=$(cat "$work/root/run/usb.pid" 2>/dev/null)
[ -n "$pid" ] && grep -q dnsmasq "/proc/$pid/cmdline" 2>/dev/null \
    && note "the pid file names the daemon" OK \
    || note "the pid file names the daemon" FAIL
grep -qx "[0-9]*" "$work/root/cgroup/cgroup.procs" \
    && note "the script joins the addons group before it starts dnsmasq" OK \
    || note "the script joins the addons group before it starts dnsmasq" FAIL

old=$pid
run start-usb usb1 172.31.255.2 255.255.255.252
new=$(cat "$work/root/run/usb.pid" 2>/dev/null)
[ "$(instances usb)" = 1 ] && [ -n "$new" ] && [ "$new" != "$old" ] \
    && note "a second start-usb replaces the instance, never adds one" OK \
    || note "a second start-usb replaces the instance (got $(instances usb), $old -> $new)" FAIL

run stop-usb
grep -q '^Stopping dnsmasq on the USB link: OK$' "$work/out" \
    && note "stop-usb reports OK" OK \
    || { note "stop-usb reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(instances usb)" = 0 ] && [ ! -e "$work/root/run/usb.pid" ] \
    && note "no link dnsmasq and no pid file are left" OK \
    || note "no link dnsmasq and no pid file are left (got $(instances usb))" FAIL

run stop-usb
[ $? = 0 ] && note "stop-usb with nothing running succeeds" OK \
    || note "stop-usb with nothing running succeeds" FAIL

echo
echo "===== the link falls back to udhcpd ====="
# S03usbdev runs udhcpd whenever start-usb fails, so every case in which
# dnsmasq is not serving the link must fail.
reset_tree
run start-usb usb0 172.31.255.2 255.255.255.252
[ $? != 0 ] && [ "$(instances usb)" = 0 ] \
    && note "without usb.conf network boot is off and start-usb fails" OK \
    || note "without usb.conf network boot is off and start-usb fails" FAIL

reset_tree usb
DNSMASQ_BIN="$work/bin/absent" run start-usb usb0 172.31.255.2 255.255.255.252
[ $? != 0 ] && grep -q "not found" "$work/out" \
    && note "without the add-on start-usb says so and fails" OK \
    || note "without the add-on start-usb says so and fails" FAIL

reset_tree usb
STUB_EXIT=1 run start-usb usb0 172.31.255.2 255.255.255.252
[ $? != 0 ] && [ ! -e "$work/root/run/usb.pid" ] \
    && note "a dnsmasq that dies at once fails and leaves no pid file" OK \
    || note "a dnsmasq that dies at once fails and leaves no pid file" FAIL

for args in "usb0 172.31.255.2" "usb0;reboot 172.31.255.2 255.255.255.252" \
    "usb0 172.31.255.2,1 255.255.255.252" "usb0 172.31.255 255.255.255.252" \
    "usb0 172.31.255.2 --conf-file=/x"; do
    # shellcheck disable=SC2086
    run start-usb $args
    [ $? != 0 ] && [ "$(instances usb)" = 0 ] \
        && note "refused: start-usb $args" OK \
        || note "refused: start-usb $args" FAIL
done

echo
echo "===== proxy DHCP on the LAN ====="
reset_tree
run start
[ $? = 0 ] && grep -q "off" "$work/out" && [ "$(instances lan)" = 0 ] \
    && note "without lan.conf start is a quiet no-op, as at boot" OK \
    || note "without lan.conf start is a quiet no-op, as at boot" FAIL

reset_tree lan
run start
grep -q 'on the LAN: OK$' "$work/out" && [ "$(instances lan)" = 1 ] \
    && note "with lan.conf start runs one LAN dnsmasq" OK \
    || { note "with lan.conf start runs one LAN dnsmasq" FAIL; sed 's/^/    /' "$work/out"; }
has lan "--leasefile-ro" "proxy DHCP keeps no lease file"
case "$(started_with lan)" in
    *--dhcp-range* | *--interface*) note "the LAN's range and interface come from the file alone" FAIL ;;
    *) note "the LAN's range and interface come from the file alone" OK ;;
esac

run start
[ "$(instances lan)" = 1 ] \
    && note "a second start does not start a second one" OK \
    || note "a second start does not start a second one (got $(instances lan))" FAIL

reset_tree lan usb
run start
run start-usb usb0 172.31.255.2 255.255.255.252
[ "$(instances lan)" = 1 ] && [ "$(instances usb)" = 1 ] \
    && note "the two instances run side by side" OK \
    || note "the two instances run side by side (lan $(instances lan), usb $(instances usb))" FAIL
run stop-usb
[ "$(instances lan)" = 1 ] \
    && note "stop-usb leaves the LAN instance alone" OK \
    || note "stop-usb leaves the LAN instance alone" FAIL

old=$(cat "$work/root/run/lan.pid" 2>/dev/null)
run restart
new=$(cat "$work/root/run/lan.pid" 2>/dev/null)
[ "$(instances lan)" = 1 ] && [ -n "$new" ] && [ "$new" != "$old" ] \
    && note "restart leaves one new LAN dnsmasq" OK \
    || note "restart leaves one new LAN dnsmasq ($old -> $new)" FAIL

run stop
[ "$(instances lan)" = 0 ] && [ ! -e "$work/root/run/lan.pid" ] \
    && note "stop leaves no LAN dnsmasq and no pid file" OK \
    || note "stop leaves no LAN dnsmasq and no pid file" FAIL

echo
echo "===== a pid file it cannot trust ====="
reset_tree lan
sleep 60 &
bystander=$!
mkdir -p "$work/root/run"
echo "$bystander" > "$work/root/run/lan.pid"
run stop
kill -0 "$bystander" 2>/dev/null \
    && note "stop does not signal a process that is not dnsmasq" OK \
    || note "stop does not signal a process that is not dnsmasq" FAIL
echo "$bystander" > "$work/root/run/usb.pid"
run stop-usb
kill -0 "$bystander" 2>/dev/null \
    && note "neither does stop-usb" OK \
    || note "neither does stop-usb" FAIL
kill "$bystander" 2>/dev/null

echo
echo "===== the log stays small ====="
reset_tree usb
mkdir -p "$work/root/run"
head -c 300000 /dev/zero | tr '\0' x > "$work/root/run/usb.log"
LOG_MAX=262144 run start-usb usb0 172.31.255.2 255.255.255.252
[ "$(wc -c < "$work/root/run/usb.log")" -lt 262144 ] \
    && note "a log over the limit is started afresh" OK \
    || note "a log over the limit is started afresh" FAIL
run stop-usb

echo
echo "===== the boot files dnsmasq serves ====="
# dnsmasq drops to nobody and will not start when it cannot read its TFTP
# directory, so it serves a copy that everyone can read.
reset_tree usb
# The board's umask.
umask 077
run start-usb usb0 172.31.255.2 255.255.255.252
umask 022
[ "$(stat -c %a "$work/root/var/netboot/tftp/boot.ipxe" 2>/dev/null)" = 644 ]     && note "the link's start copies the boot files, readable by all" OK     || note "the link's start copies the boot files, readable by all" FAIL
[ "$(stat -c %a "$work/root/var/netboot/tftp" 2>/dev/null)" = 755 ]     && [ "$(stat -c %a "$work/root/var/netboot" 2>/dev/null)" = 755 ]     && note "and their directory and its parent are open to all" OK     || note "and their directory and its parent are open to all" FAIL
echo new > "$work/root/data/tftp/boot.ipxe"
run start-usb usb0 172.31.255.2 255.255.255.252
[ "$(cat "$work/root/var/netboot/tftp/boot.ipxe")" = new ]     && note "a restart serves what the add-on holds now" OK     || note "a restart serves what the add-on holds now" FAIL
run stop-usb

reset_tree lan
run start
[ -f "$work/root/var/netboot/tftp/boot.ipxe" ]     && note "the LAN's start copies them too" OK     || note "the LAN's start copies them too" FAIL
run stop

reset_tree usb
rm -rf "$work/root/data/tftp"
run start-usb usb0 172.31.255.2 255.255.255.252
st=$?
[ "$st" -ne 0 ] && [ "$(instances usb)" = 0 ]     && note "without the boot files the link does not start dnsmasq" OK     || note "without the boot files the link does not start dnsmasq" FAIL

echo
echo "===== the script still parses ====="
sh -n "$S85" 2>/dev/null \
    && note "sh -n accepts the script" OK \
    || note "sh -n rejects the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
