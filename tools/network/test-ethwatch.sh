#!/bin/sh
# S31ethwatch restarts DHCP on the wired interface when the gateway stops
# answering, and only then.
#
#   test-ethwatch.sh [path-to-S31ethwatch]
#
# The watch block is lifted out of the script and its tick run by hand, with
# arping, ip and the probing udhcpc stubbed on PATH. The udhcpc that holds the
# lease is a shell that records the signals it gets. Nothing touches a real
# interface: on the reference board that is the network this suite is reached
# over.
SCRIPT=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S31ethwatch}
[ -f "$SCRIPT" ] || { echo "no S31ethwatch at $SCRIPT"; exit 1; }

WORK=$(mktemp -d)
FAKE=
trap '[ -n "$FAKE" ] && kill "$FAKE" 2>/dev/null; rm -rf "$WORK"' EXIT

fails=0
t() { if [ "$2" = 0 ]; then echo "ok   - $1"; else echo "FAIL - $1"; fails=1; fi; }

sed -n '/^# --- watch ---$/,/^# --- end watch ---$/p' "$SCRIPT" > "$WORK/watch.sh"
[ -s "$WORK/watch.sh" ] || { echo "FAIL - S31ethwatch has no watch block"; exit 1; }

mkdir -p "$WORK/bin" "$WORK/net/eth0" "$WORK/run" "$WORK/boot"

# Each stub appends its arguments to calls and answers from a file the case
# sets: arping.rc for the gateway (or arping.seq, one answer per line, while
# it lasts), udhcpc.rc for a DHCP server, route for the
# default route on the interface.
cat > "$WORK/bin/arping" <<EOF
#!/bin/sh
echo "arping \$*" >> "$WORK/calls"
if [ -s "$WORK/arping.seq" ]; then
    rc=\$(head -1 "$WORK/arping.seq")
    sed -i 1d "$WORK/arping.seq"
    exit \$rc
fi
exit \$(cat "$WORK/arping.rc")
EOF
cat > "$WORK/bin/udhcpc" <<EOF
#!/bin/sh
echo "udhcpc \$*" >> "$WORK/calls"
exit \$(cat "$WORK/udhcpc.rc")
EOF
cat > "$WORK/bin/ip" <<EOF
#!/bin/sh
echo "ip \$*" >> "$WORK/calls"
cat "$WORK/route" 2>/dev/null
EOF
cat > "$WORK/bin/logger" <<EOF
#!/bin/sh
echo "logger \$*" >> "$WORK/log"
EOF
chmod +x "$WORK/bin/"*

# The udhcpc that holds the lease. It writes down USR1 and USR2 as they come.
sh -c 'trap "echo USR1 >> \"$1\"" USR1; trap "echo USR2 >> \"$1\"" USR2
       while :; do sleep 0.1; done' sh "$WORK/signals" &
FAKE=$!
echo "$FAKE" > "$WORK/run/udhcpc.eth0.pid"

# reset puts every case back on a board that has a DHCP lease, a cable, a
# gateway that answers and a DHCP server.
reset() {
    rm -f "$WORK/calls" "$WORK/log" "$WORK/signals" "$WORK/boot/eth.nodhcp" \
        "$WORK/arping.seq"
    echo 1 > "$WORK/net/eth0/carrier"
    echo "default via 192.168.1.1" > "$WORK/route"
    echo 0 > "$WORK/arping.rc"
    echo 0 > "$WORK/udhcpc.rc"
    echo "$FAKE" > "$WORK/run/udhcpc.eth0.pid"
}

# ticks N runs the watch's check N times in one shell, so its counters carry
# from one to the next as they do in the loop. Three misses make the limit, and
# the hold-off after acting is two ticks.
ticks() {
    PATH="$WORK/bin:$PATH" IF=eth0 SYS_NET="$WORK/net" ETH_RUN_DIR="$WORK/run" \
    ETH_NODHCP="$WORK/boot/eth.nodhcp" ETHWATCH_LOGGER="$WORK/bin/logger" \
    ETHWATCH_INTERVAL=10 ETHWATCH_FAIL_LIMIT=3 ETHWATCH_HOLDOFF=20 \
    ETHWATCH_RELEASE_GAP=0 \
        sh -c '. "$1"; i=0; while [ $i -lt "$2" ]; do tick; i=$((i + 1)); done' \
        sh "$WORK/watch.sh" "$1" > /dev/null 2>&1
    # The fake handles a signal once its sleep returns.
    sleep 0.5
}
count() { cat "$WORK/calls" 2>/dev/null | grep -c "^$1"; }
signals() { sort "$WORK/signals" 2>/dev/null | tr '\n' ' '; }

reset
ticks 10
[ "$(count arping)" = 10 ]
t "a gateway that answers is checked on every tick" $?
grep -q -- '-I eth0 192.168.1.1' "$WORK/calls"
t "the check is an ARP request for the gateway on the wired interface" $?
[ "$(count udhcpc)" = 0 ] && [ -z "$(signals)" ] && [ ! -s "$WORK/log" ]
t "a gateway that answers: no probe, no signal, nothing logged" $?

reset
echo 1 > "$WORK/arping.rc"
ticks 2
[ "$(count udhcpc)" = 0 ] && [ -z "$(signals)" ]
t "a gateway silent for less than the limit: nothing happens yet" $?

reset
echo 1 > "$WORK/arping.rc"
ticks 3
[ "$(count udhcpc)" = 1 ]
t "a gateway silent to the limit: one probe for a DHCP server" $?
grep -q -- '-s /bin/true' "$WORK/calls" && grep -q -- ' -n' "$WORK/calls"
t "the probe configures nothing and gives up when no server answers" $?
[ "$(signals)" = "USR1 USR2 " ]
t "a server answers: the lease is released (USR2) and renewed (USR1)" $?
grep -q 'daemon.warning' "$WORK/log" && grep -q 'restarting DHCP' "$WORK/log"
t "the restart is logged at daemon.warning" $?

reset
echo 1 > "$WORK/arping.rc"
ticks 5
[ "$(count arping)" = 3 ]
t "no check during the hold-off after a restart" $?

reset
echo 1 > "$WORK/arping.rc"
echo 1 > "$WORK/udhcpc.rc"
ticks 3
[ "$(count udhcpc)" = 1 ] && [ -z "$(signals)" ]
t "no DHCP server answers: the lease is kept, no signal" $?
grep -q 'keeping the current lease' "$WORK/log"
t "keeping the lease is logged" $?

reset
echo 1 > "$WORK/arping.rc"
echo 1 > "$WORK/udhcpc.rc"
ticks 13
[ "$(count udhcpc)" = 3 ] && [ "$(grep -c 'keeping' "$WORK/log")" = 1 ]
t "a long outage probes again after each hold-off, and logs it once" $?

reset
printf '1\n1\n1\n' > "$WORK/arping.seq"
echo 1 > "$WORK/udhcpc.rc"
ticks 6
grep -q 'answers again' "$WORK/log"
t "the gateway coming back after an outage is logged" $?

reset
echo "192.168.1.50/24 192.168.1.1" > "$WORK/boot/eth.nodhcp"
echo 1 > "$WORK/arping.rc"
ticks 10
[ "$(count arping)" = 0 ] && [ "$(count udhcpc)" = 0 ] && [ -z "$(signals)" ]
t "a static address: the watchdog does nothing" $?

reset
rm -f "$WORK/run/udhcpc.eth0.pid"
echo 1 > "$WORK/arping.rc"
ticks 10
[ "$(count arping)" = 0 ] && [ "$(count udhcpc)" = 0 ]
t "no DHCP client running: the watchdog does nothing" $?

reset
echo 0 > "$WORK/net/eth0/carrier"
echo 1 > "$WORK/arping.rc"
ticks 10
[ "$(count arping)" = 0 ] && [ "$(count udhcpc)" = 0 ]
t "no cable: the watchdog does nothing" $?

reset
: > "$WORK/route"
echo 1 > "$WORK/arping.rc"
ticks 10
[ "$(count arping)" = 0 ] && [ "$(count udhcpc)" = 0 ]
t "no default route on the wire (Wi-Fi holds it): the watchdog does nothing" $?
grep -q 'dev eth0' "$WORK/calls"
t "the gateway is read from the wired interface's own default route" $?

exit $fails
