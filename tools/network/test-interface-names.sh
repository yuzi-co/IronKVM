#!/bin/sh
# Check that the network scripts take their interface names from the device
# description.
#
#   test-interface-names.sh [path-to-init.d]
#
# eth0 and wlan0 are what this board calls its interfaces. They are not what a
# second board calls them, so the names are device facts and they come from
# deviceinfo now, with ETH_IF and WLAN_IF. Each keeps today's name as its
# default, so a board whose description says nothing gets what it always got.
#
# Two things make this worth a suite of its own rather than a grep.
#
# The name reaches seven places in S30eth: the flush, the address add, the
# route add, the arping, two udhcpc calls and the pidfile. One of them keeping
# a literal is a board that flushes one interface and addresses another, and
# that is not visible in a diff of a long script.
#
# And S30wifi has a case S30eth does not. A board with no radio declares no
# WLAN_IF, and the script has to say so and stop. Configuring an interface that
# is not there is not harmless: it writes a supplicant configuration and starts
# a client for a device the kernel does not have.
#
# HOW THESE CASES RUN THE SCRIPTS
#
# Both scripts write to absolute paths, and this suite is also run on a live
# board, where /etc/wpa_supplicant.conf and /etc/resolv.conf are the real ones
# and /boot is the real boot partition. So each case runs a copy of the shipped
# script with those path prefixes pointed at a sandbox, and nothing else about
# it changed. A copy made that way still carries every edit and every mutation,
# which is what a hand-written copy of the decisions would not.
#
# Everything the scripts then call is a stub on PATH that records its argument
# list: ip, arping, udhcpc, wpa_supplicant, wpa_passphrase, hostapd, ifconfig.
# No real interface is touched on any host.
set -u

HERE=$(cd "$(dirname "$0")" && pwd)
DIR=${1:-$HERE/../../kvmapp/system/init.d}
[ -d "$DIR" ] || { echo "usage: test-interface-names.sh <init.d dir>"; exit 1; }

ETH=$DIR/S30eth
WIFI=$DIR/S30wifi
for f in "$ETH" "$WIFI"; do
    [ -f "$f" ] || { echo "no $f"; exit 2; }
done

# The real reader, not a stub, so a key either script spells differently is a
# key this suite fails on rather than one a stub agrees with.
READER=${READER:-$HERE/../../kvmapp/system/ironkvm-deviceinfo}
[ -f "$READER" ] || { echo "no ironkvm-deviceinfo at $READER; set READER"; exit 2; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

mkdir -p "$WORK/path" "$WORK/etc" "$WORK/boot" "$WORK/kvmapp/system" \
    "$WORK/var/run" "$WORK/tmp"

# S30wifi sources /etc/profile, and the sandbox needs one of its own: the
# host's would put its own PATH in front of the stubs.
: > "$WORK/etc/profile"

cp "$READER" "$WORK/bin-deviceinfo"
chmod 755 "$WORK/bin-deviceinfo"

# The stubs. Each one appends its whole argument list to the trace, and does
# nothing else. A script that reaches a command this suite did not stub gets
# the host's, which is why the trace is read rather than the exit status.
for c in ip arping udhcpc wpa_supplicant wpa_passphrase hostapd udhcpd \
         ifconfig airmon-ng killall ps; do
    cat > "$WORK/path/$c" <<STUB
#!/bin/sh
printf '%s %s\n' "$c" "\$*" >> "\$TRACE"
exit 0
STUB
    chmod 755 "$WORK/path/$c"
done

# sandbox <script> <name>: the shipped script with the paths it writes pointed
# into the sandbox. The interface names are untouched, which is the whole point.
# /tmp comes first. $WORK is itself under /tmp, and sed applies these in order
# on the same line, so a later /tmp rule would rewrite what an earlier one
# produced.
sandbox() {
    sed -e "s|/tmp/|$WORK/tmp/|g" \
        -e "s|/etc/|$WORK/etc/|g" \
        -e "s|/boot/|$WORK/boot/|g" \
        -e "s|/kvmapp/|$WORK/kvmapp/|g" \
        -e "s|/var/run/|$WORK/var/run/|g" \
        "$1" > "$WORK/$2"
}

sandbox "$ETH" S30eth
sandbox "$WIFI" S30wifi

# describe KEY=value ...: the description the case gives the script. With no
# arguments the file is empty, which is a description that names no interface.
describe() {
    : > "$WORK/deviceinfo"
    for kv in "$@"; do
        printf '%s\n' "$kv" >> "$WORK/deviceinfo"
    done
}

# run <sandboxed script> <action>: run it with the stubs and the description,
# and leave the trace in $WORK/trace.
run() {
    : > "$WORK/trace"
    (
        PATH="$WORK/path:$PATH"
        TRACE="$WORK/trace"
        DEVINFO="$WORK/bin-deviceinfo"
        DEVICEINFO_PATHS="$WORK/deviceinfo"
        export PATH TRACE DEVINFO DEVICEINFO_PATHS
        sh "$WORK/$1" "$2"
    ) 2>&1
}

traced() {  # traced <extended regexp>: is that line in the trace?
    grep -qE -- "$1" "$WORK/trace" 2>/dev/null
}

echo "===== S30eth takes the wired interface from the description ====="

# The DHCP branch, which is what a board without /boot/eth.nodhcp runs. The
# flush, the client and the pidfile are all in it.
rm -f "$WORK/boot/eth.nodhcp"
describe ETH_IF=eth1
run S30eth start > /dev/null

traced '^ip -4 addr flush dev eth1$' \
    && note "the flush names the interface the description gives" OK \
    || note "the flush was '$(grep '^ip ' "$WORK/trace" | head -1)'" FAIL

traced '^udhcpc -i eth1 ' \
    && note "the DHCP client runs on that interface" OK \
    || note "the DHCP client ran '$(grep '^udhcpc ' "$WORK/trace" | head -1)'" FAIL

traced '^udhcpc .* -p /run/udhcpc\.eth1\.pid' \
    && note "the pidfile follows the interface name" OK \
    || note "the pidfile was not /run/udhcpc.eth1.pid" FAIL

# The static branch. arping, the address add and the route add are only here,
# and the route add is the one a careless edit leaves behind.
printf '10.9.8.7/24 10.9.8.1\n' > "$WORK/boot/eth.nodhcp"
describe ETH_IF=eth1
run S30eth start > /dev/null

traced '^arping .*-Ieth1 ' \
    && note "the duplicate-address probe names that interface" OK \
    || note "the probe was '$(grep '^arping ' "$WORK/trace" | head -1)'" FAIL

traced '^ip a add 10\.9\.8\.7/24 brd . dev eth1$' \
    && note "the address is added to that interface" OK \
    || note "the address add was '$(grep '^ip a add' "$WORK/trace" | head -1)'" FAIL

traced '^ip r add default via 10\.9\.8\.1 dev eth1$' \
    && note "the default route is added on that interface" OK \
    || note "the route add was '$(grep '^ip r add' "$WORK/trace" | head -1)'" FAIL

# The stubbed address show answers nothing, so this branch falls through to the
# reserve address as well, which is the seventh and last place the name appears.
traced '^ip a add 192\.168\.90\.1/22 brd . dev eth1$' \
    && note "the reserve address goes on that interface too" OK \
    || note "the reserve address did not go on eth1" FAIL

# A description that says nothing about the interface is every board deployed
# before this key existed, and it has to behave exactly as it did.
rm -f "$WORK/boot/eth.nodhcp"
describe
run S30eth start > /dev/null

traced '^ip -4 addr flush dev eth0$' && traced '^udhcpc -i eth0 ' \
    && note "a description with no ETH_IF still uses eth0" OK \
    || note "a description with no ETH_IF did not use eth0" FAIL

# The tarball path. A board on Sipeed's firmware has no reader on PATH, and
# the copy the application tarball carries is the only one it can reach.
: > "$WORK/trace"
cp "$READER" "$WORK/kvmapp/system/ironkvm-deviceinfo"
chmod 755 "$WORK/kvmapp/system/ironkvm-deviceinfo"
# No DEVINFO here, so the script resolves the reader itself. No host this suite
# runs on carries ironkvm-deviceinfo on PATH, which is the whole reason the
# second step exists.
describe ETH_IF=eth2
(
    PATH="$WORK/path:$PATH"
    TRACE="$WORK/trace"
    DEVICEINFO_PATHS="$WORK/deviceinfo"
    export PATH TRACE DEVICEINFO_PATHS
    sh "$WORK/S30eth" start
) > /dev/null 2>&1

traced '^ip -4 addr flush dev eth2$' \
    && note "with no reader on PATH it reads the tarball's copy" OK \
    || note "with no reader on PATH it did not read the tarball's copy" FAIL

echo
echo "===== S30wifi takes the wireless interface from the description ====="

describe ETH_IF=eth0 WLAN_IF=wlan9
run S30wifi start > /dev/null

traced '^wpa_supplicant .*-i wlan9 ' \
    && note "the supplicant runs on the interface the description gives" OK \
    || note "the supplicant ran '$(grep '^wpa_supplicant ' "$WORK/trace" | head -1)'" FAIL

traced '^udhcpc -i wlan9 .* -p /run/udhcpc\.wlan9\.pid' \
    && note "its DHCP client and pidfile follow that name" OK \
    || note "its DHCP client ran '$(grep '^udhcpc ' "$WORK/trace" | head -1)'" FAIL

describe ETH_IF=eth0 WLAN_IF=wlan9
run S30wifi stop > /dev/null

traced '^airmon-ng stop wlan9mon$' \
    && note "the monitor interface is the wireless one with mon after it" OK \
    || note "airmon-ng was stopped on '$(grep '^airmon-ng ' "$WORK/trace" | head -1)'" FAIL

echo
echo "===== a board with no wireless configures nothing ====="

# The reference board is a beta, and /sys/class/net there holds eth0 and
# nothing else. Running the whole of start on such a board writes a supplicant
# configuration and starts a client for a device the kernel does not have.
#
# The description is readable and names no WLAN_IF, which is what says the
# board has no radio. A description that cannot be read at all is the case
# after this one, and it is deliberately the opposite answer.
rm -f "$WORK/etc/wpa_supplicant.conf"
describe ETH_IF=eth0
out=$(run S30wifi start)
rc=$?

[ "$rc" -eq 0 ] \
    && note "S30wifi with no WLAN_IF exits 0" OK \
    || note "S30wifi with no WLAN_IF exited $rc" FAIL

[ ! -s "$WORK/trace" ] \
    && note "S30wifi with no WLAN_IF runs nothing" OK \
    || note "S30wifi with no WLAN_IF ran '$(head -1 "$WORK/trace")'" FAIL

[ ! -e "$WORK/etc/wpa_supplicant.conf" ] \
    && note "S30wifi with no WLAN_IF writes no supplicant configuration" OK \
    || note "S30wifi with no WLAN_IF wrote a supplicant configuration" FAIL

case "$out" in
    *"no wireless interface"*) note "S30wifi says why it did nothing" OK ;;
    *) note "S30wifi said '$out'" FAIL ;;
esac

# The opposite answer, and the reason the two cases are not one. A board whose
# description cannot be read is not a board with no radio. Reading it that way
# would take the wireless off every board that has one.
rm -f "$WORK/deviceinfo" "$WORK/etc/wpa_supplicant.conf"
run S30wifi start > /dev/null

traced '^wpa_supplicant .*-i wlan0 ' \
    && note "no description at all still uses wlan0" OK \
    || note "no description at all did not use wlan0" FAIL

echo
echo "===== no interface name is left as a literal ====="

# Every place in S30eth is executed above. This case covers both scripts at
# once, and it is the cheapest guard against the next edit putting a literal
# back somewhere no case reaches.
#
# Comments are not read. They name eth0 and wlan0 on purpose, to say what the
# default is and why.
for f in "$ETH" "$WIFI"; do
    left=$(grep -vn '^[[:space:]]*#' "$f" | grep -c 'eth0\|wlan0' || true)
    # The defaults themselves are the exception, and each script has exactly
    # one of them.
    case "$(basename "$f")" in
        S30eth)  allowed=1 ;;
        S30wifi) allowed=1 ;;
    esac
    [ "$left" -le "$allowed" ] \
        && note "$(basename "$f") names no interface outside its default" OK \
        || note "$(basename "$f") still has $left literal interface names" FAIL
done

echo
echo "===== both scripts parse ====="
for f in "$ETH" "$WIFI"; do
    sh -n "$f" 2>/dev/null \
        && note "sh -n accepts $(basename "$f")" OK \
        || note "sh -n accepts $(basename "$f")" FAIL
done

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
