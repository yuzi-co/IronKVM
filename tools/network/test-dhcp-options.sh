#!/bin/sh
# Check that every udhcpc call asks the server for classless static routes.
#
#   test-dhcp-options.sh [path-to-init.d]
#
# A DHCP server can hand out routes that are not the default route, through
# option 121, "classless static routes" (RFC 3442). udhcpc does not ask for
# that option unless the command line names it, so a network that depends on
# those routes leaves the device able to reach its own subnet and the default
# gateway and nothing else. The symptom is a device that pings but cannot
# reach half the estate, which reads like a firewall fault rather than a DHCP
# one.
#
# The device already knows what to do with the answer. Its udhcpc handler,
# /usr/share/udhcpc/default.script, tests "$staticroutes" and installs each
# pair. That was read off the board at /usr/share/udhcpc/default.script:73.
# Only the request is missing.
#
# Both interfaces are checked. A route that matters on the wire matters the
# same over Wi-Fi, and a rule that covers one file invites the other to drift.
set -u

DIR=${1:-$(dirname "$0")/../../kvmapp/system/init.d}
[ -d "$DIR" ] || { echo "usage: test-dhcp-options.sh <init.d dir>"; exit 1; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== every udhcpc call requests option 121 ====="

for f in S30eth S30wifi; do
    script="$DIR/$f"
    if [ ! -f "$script" ]; then
        note "$f exists" FAIL
        continue
    fi

    # Every line that runs udhcpc, ignoring the ones that only kill it or read
    # its pid file.
    calls=$(grep -n 'udhcpc ' "$script" | grep -v 'kill\|rm \|-e "\|\[ ' || true)

    if [ -z "$calls" ]; then
        note "$f calls udhcpc" FAIL
        continue
    fi

    total=0
    missing=0
    for n in $(echo "$calls" | cut -d: -f1); do
        total=$((total + 1))
        text=$(sed -n "${n}p" "$script")
        case "$text" in
            *-O\ 121*) : ;;
            *) missing=$((missing + 1)); note "$f:$n requests option 121" FAIL ;;
        esac
    done

    [ "$missing" = 0 ] && note "$f: all $total udhcpc call(s) request option 121" OK
done

echo "===== every udhcpc call asks for a broadcast reply ====="

# A client that holds no address cannot be reached by unicast. RFC 2131 4.4.1
# lets the server answer a DISCOVER either way, and a relay agent can only use
# the broadcast one, because the unicast one asks it to ARP for a host that has
# not configured the address yet. It drops the reply instead. The symptom is a
# board that leases in a second on a directly attached network and never leases
# through a switch that relays DHCP from another VLAN.
#
# busybox sets the flag only while ciaddr is zero, in add_client_options(), so
# -B reaches the DISCOVER and the SELECT and leaves a RENEW alone. That is what
# RFC 2131 4.3.6 asks for, and it is why the option can go on every call here.
#
# S30wifi also passes -B to wpa_supplicant, where it means "run in the
# background". Only the udhcpc lines are read below.
for f in S30eth S30wifi; do
    script="$DIR/$f"
    if [ ! -f "$script" ]; then
        note "$f exists" FAIL
        continue
    fi

    calls=$(grep -n 'udhcpc ' "$script" | grep -v 'kill\|rm \|-e "\|\[ ' || true)

    if [ -z "$calls" ]; then
        note "$f calls udhcpc" FAIL
        continue
    fi

    total=0
    missing=0
    for n in $(echo "$calls" | cut -d: -f1); do
        total=$((total + 1))
        text=" $(sed -n "${n}p" "$script") "
        case "$text" in
            *\ -B\ *) : ;;
            *) missing=$((missing + 1)); note "$f:$n asks for a broadcast reply" FAIL ;;
        esac
    done

    [ "$missing" = 0 ] && note "$f: all $total udhcpc call(s) ask for a broadcast reply" OK
done

echo "===== no address flush takes the IPv6 addresses with it ====="

# `ip addr flush dev eth0` clears every family. It removes the IPv6 link-local
# as well as the global, so the interface keeps IPv4 and loses IPv6.
#
# At boot the loss does not show, because the link comes up after the script
# runs. S30eth also runs on every revert of an address trial, and there it does
# show: measured on a device on 2026-09-07, eth0 had no IPv6 at all for several
# minutes after a revert, and when the global address did come back on its own
# the link-local did not come with it.
#
# Point this at /etc/init.d on a device to check the copy that actually boots,
# which is the copy that matters and the one a package update can replace.
for f in S30eth S30wifi; do
    script="$DIR/$f"
    [ -f "$script" ] || continue

    flushes=$(grep -n '^[[:space:]]*ip .*addr .*flush' "$script" || true)
    if [ -z "$flushes" ]; then
        note "$f flushes no addresses" OK
        continue
    fi

    total=0
    bare=0
    for n in $(echo "$flushes" | cut -d: -f1); do
        total=$((total + 1))
        text=$(sed -n "${n}p" "$script")
        case "$text" in
            *ip\ -4\ *) : ;;
            *) bare=$((bare + 1)); note "$f:$n names the address family" FAIL ;;
        esac
    done

    [ "$bare" = 0 ] && note "$f: all $total flush(es) keep IPv6" OK
done

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
