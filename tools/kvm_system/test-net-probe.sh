#!/bin/sh
# Check the gateway probe kvm_system uses for the OLED network icons.
#
#   test-net-probe.sh [path/to/kvm_system/main]
#
# Not destructive. The ICMP cases ping 127.0.0.1 on lo, and skip themselves
# when this process may not open a raw socket.
#
# kvm_system used to run "ping -I eth0 -w 1 <gateway>" through system() on
# every pass of its one-second state loop, and a shell pipeline of ip, grep
# and awk to find the gateway. net_probe.cpp replaces both without a fork, and
# system_state.cpp probes at most every GATEWAY_PROBE_INTERVAL_MS. The file
# has no MaixCDK dependency, so it builds here with the host compiler; the
# rest of kvm_system does not, and the last section only reads it.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
MAIN=${1:-$ROOT/support/sg2002/kvm_system/main}
SRC=$MAIN/lib/system_state/net_probe.cpp
STATE=$MAIN/lib/system_state/system_state.cpp

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }
[ -f "$STATE" ] || { echo "missing: $STATE"; exit 2; }
CXX=${CXX:-g++}
command -v "$CXX" > /dev/null 2>&1 || { echo "no C++ compiler ($CXX); cannot run here"; exit 2; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

cat > "$work/t.cpp" <<'EOF'
#include "net_probe.h"
#include <arpa/inet.h>
#include <stdio.h>
#include <string.h>

static int fails = 0;
static void check(int ok, const char *what)
{
	printf("  %-58s %s\n", what, ok ? "OK" : "FAIL");
	if (!ok) fails++;
}

static int gw_is(const char *s, size_t len, const char *want)
{
	struct in_addr a;
	if (!parse_gateway(s, len, &a)) return want == NULL;
	if (want == NULL) return 0;
	char got[INET_ADDRSTRLEN];
	inet_ntop(AF_INET, &a, got, sizeof(got));
	return strcmp(got, want) == 0;
}

// What the board at hand printed, on 2026-09-30: a default route through eth0
// to 10.0.0.1 and the link route.
static const char *board_table =
	"Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"
	"eth0\t00000000\t0100000A\t0003\t0\t0\t202\t00000000\t0\t0\t0\n"
	"eth0\t0000000A\t00000000\t0001\t0\t0\t0\t00FFFFFF\t0\t0\t0\n";

int main(int argc, char **argv)
{
	int mode = argc > 1 ? argv[1][0] : 'u';
	if (mode == 'u') {
		printf("===== the route buffer reads as an address =====\n");
		check(gw_is("10.0.0.1 ", 16, "10.0.0.1"), "the space the route lookup leaves");
		check(gw_is("10.0.0.1\n", 16, "10.0.0.1"), "a newline from /etc/kvm/gateway");
		check(gw_is("10.0.0.1\0\0\0\0\0\0\0", 16, "10.0.0.1"), "a terminated buffer");
		check(gw_is("192.168.100.100X", 15, "192.168.100.100"), "15 characters and no terminator");
		check(gw_is("gateway.lan", 16, NULL), "a host name falls back to ping");
		check(gw_is("10.0.0.1x", 16, NULL), "an address with trailing junk is refused");
		check(gw_is("300.0.0.1", 16, NULL), "an octet over 255 is refused");
		check(gw_is("", 16, NULL), "an empty buffer is refused");

		printf("===== the gateway comes from /proc/net/route =====\n");
		char out[16];
		check(route_gateway_from_table(board_table, "eth0", out, sizeof(out)) == 1
			&& strcmp(out, "10.0.0.1 ") == 0, "the board's default route, with the old space");
		check(route_gateway_from_table(board_table, "wlan0", out, sizeof(out)) == 0
			&& out[0] == 0, "no default route on wlan0: empty");
		check(route_gateway_from_table(board_table, "eth", out, sizeof(out)) == 0,
			"an interface name is matched whole");
		const char *two =
			"Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"
			"wlan0\t00000000\t0101A8C0\t0003\t0\t0\t300\t00000000\t0\t0\t0\n"
			"eth0\t00000000\t0100000A\t0003\t0\t0\t202\t00000000\t0\t0\t0\n"
			"eth0\t00000000\t0200000A\t0003\t0\t0\t400\t00000000\t0\t0\t0\n";
		check(route_gateway_from_table(two, "wlan0", out, sizeof(out)) == 1
			&& strcmp(out, "192.168.1.1 ") == 0, "wlan0 gets its own gateway");
		check(route_gateway_from_table(two, "eth0", out, sizeof(out)) == 1
			&& strcmp(out, "10.0.0.1 ") == 0, "the first eth0 default route wins");
		const char *nogw =
			"Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"
			"eth0\t00000000\t00000000\t0001\t0\t0\t0\t00000000\t0\t0\t0\n";
		check(route_gateway_from_table(nogw, "eth0", out, sizeof(out)) == 0,
			"a default route with no gateway is not one");
		const char *longgw =
			"Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"
			"eth0\t00000000\t6464A8C0\t0003\t0\t0\t0\t00000000\t0\t0\t0\n";
		check(route_gateway_from_table(longgw, "eth0", out, sizeof(out)) == 1
			&& strcmp(out, "192.168.100.100") == 0, "15 characters fill the buffer, as fgets did");
		check(route_gateway_from_table("", "eth0", out, sizeof(out)) == 0, "an empty table");

		printf("===== the probe runs on the interval =====\n");
		probe_timer_t t = {};
		check(probe_due(&t, 5000, 10000), "never probed: due at once");
		probe_mark(&t, 5000);
		check(!probe_due(&t, 6000, 10000), "one second later: not due");
		check(!probe_due(&t, 14999, 10000), "just under the interval: not due");
		check(probe_due(&t, 15000, 10000), "at the interval: due");
		probe_invalidate(&t);
		check(probe_due(&t, 6000, 10000), "invalidated: due at once");
		probe_mark(&t, 0xFFFFF000U);
		check(!probe_due(&t, 0x00000800U, 10000), "across the clock wrap: not due early");
		check(probe_due(&t, 0x00002000U, 10000), "across the clock wrap: due on time");
		return fails ? 1 : 0;
	}

	// mode 'i': ICMP on loopback. Exit 3 when raw sockets are not allowed.
	struct in_addr lo;
	inet_pton(AF_INET, "127.0.0.1", &lo);
	int r = icmp_probe("lo", lo, 1000);
	if (r < 0) return 3;
	printf("===== the probe pings without a fork =====\n");
	check(r == 1, "127.0.0.1 on lo answers");
	check(icmp_probe("nosuchif0", lo, 200) == -1, "an interface that does not exist: fall back");
	struct in_addr none;
	inet_pton(AF_INET, "192.0.2.1", &none);
	uint32_t start = probe_now_ms();
	int silent = icmp_probe("lo", none, 300);
	uint32_t took = probe_now_ms() - start;
	check(silent == 0, "a gateway that never answers: 0");
	check(took < 1000, "and the wait is bounded by the timeout");
	return fails ? 1 : 0;
}
EOF

if ! "$CXX" -std=gnu++11 -Wall -Wextra -Werror -I"$MAIN/lib/system_state" \
        -o "$work/t" "$work/t.cpp" "$SRC" 2> "$work/cc.log"; then
    sed 's/^/    /' "$work/cc.log"
    note "net_probe.cpp builds alone, warnings as errors" FAIL
    exit 1
fi

"$work/t" u || fails=$((fails + 1))

"$work/t" i
rc=$?
case $rc in
    0) ;;
    3) echo "  (no raw socket here: the ICMP cases did not run)" ;;
    *) fails=$((fails + 1)) ;;
esac

echo "===== the state loop does not fork for the network ====="
# The sections that run every pass. A shell here is a fork a second on a
# single-core board, which is what this change removed.
body() { awk -v f="$1" '$0 ~ "^[a-z].*[ *]" f "\\(" { p = 1 } p { print } p && /^}/ { exit }' "$STATE"; }
for fn in kvm_update_eth_state kvm_update_wifi_state; do
    b=$(body "$fn")
    [ -n "$b" ] || { note "$fn found" FAIL; continue; }
    if printf '%s\n' "$b" | grep -q 'popen\|system("echo'; then
        note "$fn runs no shell each pass" FAIL
    else
        note "$fn runs no shell each pass" OK
    fi
    if printf '%s\n' "$b" | grep -q 'probe_due'; then
        note "$fn probes on the interval" OK
    else
        note "$fn probes on the interval" FAIL
    fi
done
# get_ip_addr looks the gateway up on every pass until it finds one.
if grep -q 'popen' "$STATE"; then
    note "system_state.cpp opens no pipe to a shell" FAIL
else
    note "system_state.cpp opens no pipe to a shell" OK
fi

echo
if [ "$fails" -gt 0 ]; then
    echo "FAILED: $fails"
    exit 1
fi
echo "all passed"
