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

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
