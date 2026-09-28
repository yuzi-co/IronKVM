#!/bin/sh
# Exercise the USB network link: which function the markers pick, the subnet,
# the DHCP server's configuration, and what the board does to the interface.
#
#   test-usb-network.sh [path-to-S03usbdev]
#
# Not destructive. The block is taken straight out of the script that ships,
# its /proc/ paths are pointed at a directory under mktemp, and ip, udhcpd,
# iptables, ip6tables and kill are shell functions that record their arguments
# and touch nothing.
SV=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S03usbdev}
[ -f "$SV" ] || { echo "usage: test-usb-network.sh <S03usbdev>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

{
    sed -n '/^usb_marker() {$/,/^}$/p' "$SV"
    sed -n '/^# --- usb network ---$/,/^# --- end usb network ---$/p' "$SV" \
        | sed "s|/proc/|$WORK/proc/|g"
} > "$WORK/net.sh"
grep -q '^usb_net_start() {$' "$WORK/net.sh" \
    || { echo "could not extract the usb network block"; exit 1; }

fails=0
note() { printf '  %-66s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# The stubs every case that reaches the outside world runs with. Each records
# one line in $CALLS. IPT_HAS answers iptables -C: 1 means the rule is absent.
cat > "$WORK/stubs.sh" <<'STUBS'
ip()        { echo "ip $*" >> "$CALLS"; return "${IP_FAILS:-0}"; }
udhcpd()    { echo "udhcpd $*" >> "$CALLS"; return 0; }
kill()      { echo "kill $*" >> "$CALLS"; return 0; }
iptables()  { echo "iptables $*" >> "$CALLS"; [ "$1" = -C ] && return "${IPT_HAS:-1}"; return 0; }
ip6tables() { echo "ip6tables $*" >> "$CALLS"; [ "$1" = -C ] && return "${IPT_HAS:-1}"; return 0; }
STUBS

echo "===== which network function the markers pick ====="
function_case() {
    desc="$1"; markers="$2"; want="$3"
    boot="$WORK/boot-$$-$(echo "$markers" | tr ' .' '__')"
    mkdir -p "$boot"
    for m in $markers; do : > "$boot/$m"; done
    got=$(WORK="$WORK" BOOT="$boot" sh -c '
        . "$WORK/net.sh"
        usb_marker() { [ -e "$BOOT/$1" ]; }
        usb_net_function
    ')
    [ "$got" = "$want" ] && note "$desc -> [$got]" OK || note "$desc -> [$got], want [$want]" FAIL
}
function_case "no marker picks nothing"          ""                          ""
function_case "usb.ncm picks ncm"                "usb.ncm"                   ncm
function_case "usb.ecm picks ecm"                "usb.ecm"                   ecm
function_case "usb.rndis0 picks rndis"           "usb.rndis0"                rndis
function_case "ncm outranks ecm"                 "usb.ecm usb.ncm"           ncm
function_case "ecm outranks rndis"               "usb.rndis0 usb.ecm"        ecm
function_case "all three pick ncm"               "usb.rndis0 usb.ecm usb.ncm" ncm

echo
echo "===== reading a subnet ====="
parse_case() {
    subnet="$1"; want="$2"
    got=$(WORK="$WORK" S="$subnet" sh -c '. "$WORK/net.sh"; usb_net_parse "$S"')
    status=$?
    if [ -z "$want" ]
    then
        [ "$status" -ne 0 ] && [ -z "$got" ] && note "[$subnet] is refused" OK \
            || note "[$subnet] gave [$got], status $status, want a refusal" FAIL
    else
        [ "$status" -eq 0 ] && [ "$got" = "$want" ] && note "[$subnet] -> [$got]" OK \
            || note "[$subnet] gave [$got], status $status, want [$want]" FAIL
    fi
}
#          subnet              board        host         prefix mask
parse_case 172.31.255.0/30    "172.31.255.1 172.31.255.2 30 255.255.255.252"
parse_case 10.0.0.0/24        "10.0.0.1 10.0.0.2 24 255.255.255.0"
parse_case 192.168.7.8/29     "192.168.7.9 192.168.7.10 29 255.255.255.248"
parse_case 172.16.0.128/25    "172.16.0.129 172.16.0.130 25 255.255.255.128"
parse_case 10.1.2.64/26       "10.1.2.65 10.1.2.66 26 255.255.255.192"
parse_case 10.1.2.240/28      "10.1.2.241 10.1.2.242 28 255.255.255.240"
parse_case 10.1.2.224/27      "10.1.2.225 10.1.2.226 27 255.255.255.224"
# Refused: host bits set, a public network, the edges of 172.16/12, a prefix
# outside /24 to /30, and every malformed shape a hand edit can produce.
parse_case 172.31.255.1/30    ""
parse_case 172.31.255.4/29    ""
parse_case 8.8.8.0/30         ""
parse_case 100.64.0.0/30      ""
parse_case 172.15.0.0/30      ""
parse_case 172.32.0.0/30      ""
parse_case 192.169.0.0/30     ""
parse_case 10.0.0.0/31        ""
parse_case 10.0.0.0/23        ""
parse_case 10.0.0.0/030       ""
parse_case 10.0.0.0           ""
parse_case 10.0.0/30          ""
parse_case 10.0.0.0.0/30      ""
parse_case 10.0.0.256/30      ""
parse_case 10.0.0.1000/30     ""
parse_case 10.0.0.08/30       ""
parse_case 10.0.0.-4/30       ""
parse_case a.b.c.d/30         ""
parse_case ""                 ""
parse_case "10.0.0.0/30/1"    ""

echo
echo "===== where the subnet comes from ====="
subnet_case() {
    desc="$1"; content="$2"; want="$3"; want_err="$4"
    file="$WORK/subnet-$$-$(echo "$desc" | tr -c 'a-z0-9' '_')"
    rm -f "$file"
    [ "$content" != "<absent>" ] && printf '%b' "$content" > "$file"
    got=$(WORK="$WORK" USB_NET_SUBNET_FILE="$file" sh -c '. "$WORK/net.sh"; usb_net_subnet' 2> "$WORK/err")
    [ "$got" = "$want" ] && note "$desc -> [$got]" OK || note "$desc -> [$got], want [$want]" FAIL
    if [ "$want_err" = yes ]
    then
        grep -q "using 172.31.255.0/30" "$WORK/err" && note "$desc says it fell back" OK \
            || note "$desc fell back without a word" FAIL
    else
        [ -s "$WORK/err" ] && note "$desc printed [$(cat "$WORK/err")]" FAIL || note "$desc is quiet" OK
    fi
}
subnet_case "no file uses the default"          "<absent>"               172.31.255.0/30 no
subnet_case "a valid file is used"              "10.9.8.0/29\n"          10.9.8.0/29     no
subnet_case "a file with CRLF is used"          "10.9.8.0/29\r\n"        10.9.8.0/29     no
subnet_case "only the first line counts"        "10.9.8.0/29\n8.8.8.0/30\n" 10.9.8.0/29  no
subnet_case "a public network falls back"       "8.8.8.0/30\n"           172.31.255.0/30 yes
subnet_case "an empty file falls back"          ""                       172.31.255.0/30 yes

echo
echo "===== the DHCP server offers one address and nothing else ====="
conf=$(WORK="$WORK" sh -c '. "$WORK/net.sh"; usb_net_udhcpd_conf usb1 172.31.255.2 255.255.255.252 /run/x')
conf_has() {
    printf '%s\n' "$conf" | grep -qx "$1" && note "the configuration has [$1]" OK \
        || note "the configuration lacks [$1]" FAIL
}
conf_has "interface usb1"
conf_has "start 172.31.255.2"
conf_has "end 172.31.255.2"
conf_has "max_leases 1"
conf_has "option subnet 255.255.255.252"
conf_has "pidfile /run/x/udhcpd.pid"
conf_has "lease_file /run/x/udhcpd.leases"
# The two lines that would turn the link into a route to the LAN, or take
# over the host's name resolution.
if printf '%s\n' "$conf" | grep -qiE '^(opt|option) +(router|dns|domain|wins)'
then
    note "the configuration offers a router or a name server" FAIL
else
    note "the configuration offers no router and no name server" OK
fi
if printf '%s\n' "$conf" | grep -qE '^interface (eth|wlan)'
then
    note "the configuration serves a LAN interface" FAIL
else
    note "the configuration serves the USB interface alone" OK
fi

echo
echo "===== the interface name comes from the function ====="
G="$WORK/g0"
mkdir -p "$G/functions/ncm.usb0" "$G/functions/ecm.usb0" "$G/functions/rndis.usb0"
ifname_case() {
    desc="$1"; dir="$2"; content="$3"; want="$4"
    if [ "$content" = "<absent>" ]
    then
        rm -f "$G/functions/$dir/ifname"
    else
        echo "$content" > "$G/functions/$dir/ifname"
    fi
    got=$(WORK="$WORK" USB_GADGET_DIR="$G" D="$dir" sh -c '. "$WORK/net.sh"; usb_net_ifname "$D"')
    [ "$got" = "$want" ] && note "$desc -> [$got]" OK || note "$desc -> [$got], want [$want]" FAIL
}
ifname_case "a bound ncm reads usb0"                  ncm.usb0   usb0      usb0
ifname_case "ecm after a switch reads usb1"           ecm.usb0   usb1      usb1
ifname_case "the unbound template names nothing"      rndis.usb0 'usb%d'   ""
ifname_case "a function never created names nothing"  rndis.usb0 "<absent>" ""

echo
echo "===== starting the link ====="
# ncm.usb0 holds usb0 from an earlier start, and the board now runs ECM on
# usb1. Only usb1 may end up with the address.
start_run() {
    rm -rf "$WORK/run" "$WORK/proc"
    mkdir -p "$WORK/proc/sys/net/ipv4/conf/usb1" "$WORK/proc/sys/net/ipv6/conf/usb1"
    echo 1 > "$WORK/proc/sys/net/ipv4/conf/usb1/forwarding"
    echo 1 > "$WORK/proc/sys/net/ipv6/conf/usb1/forwarding"
    : > "$WORK/calls"
    WORK="$WORK" CALLS="$WORK/calls" USB_GADGET_DIR="$G" USB_NET_RUN="$WORK/run" \
        USB_NET_SUBNET_FILE="$WORK/no-such-file" IPT_HAS="${IPT_HAS:-1}" IP_FAILS="${IP_FAILS:-0}" \
        sh -c '. "$WORK/net.sh"; . "$WORK/stubs.sh"; usb_net_start "$1"; echo "status=$?"' _ "$1" > "$WORK/out" 2>&1
}
called() {
    grep -qxF "$1" "$WORK/calls" && note "$2" OK || note "$2: no [$1] in the calls" FAIL
}
not_called() {
    grep -qxF "$1" "$WORK/calls" && note "$2: [$1] was called" FAIL || note "$2" OK
}

echo usb0 > "$G/functions/ncm.usb0/ifname"
echo usb1 > "$G/functions/ecm.usb0/ifname"
echo 'usb%d' > "$G/functions/rndis.usb0/ifname"
start_run ecm

called     "ip addr flush dev usb0"                   "the address comes off the NCM interface left from before"
not_called "ip addr add 172.31.255.1/30 dev usb0"     "the NCM interface gets no address"
called     "ip addr add 172.31.255.1/30 dev usb1"     "the ECM interface gets the board address"
called     "ip link set usb1 up"                      "the ECM interface is brought up"
called     "udhcpd -S $WORK/run/udhcpd.conf"          "udhcpd starts on the link's own configuration"
called     "iptables -I FORWARD -i usb+ -j DROP"      "forwarding in from the link is dropped"
called     "iptables -I FORWARD -o usb+ -j DROP"      "forwarding out to the link is dropped"
called     "ip6tables -I FORWARD -i usb+ -j DROP"     "IPv6 forwarding in from the link is dropped"
grep -q "status=0" "$WORK/out" && note "a good start returns 0" OK || note "a good start returned [$(cat "$WORK/out")]" FAIL
grep -qx "interface usb1" "$WORK/run/udhcpd.conf" 2>/dev/null \
    && note "udhcpd serves usb1" OK || note "udhcpd does not serve usb1" FAIL
[ -f "$WORK/run/udhcpd.leases" ] && note "the lease file exists before udhcpd reads it" OK \
    || note "no lease file for udhcpd" FAIL
[ "$(cat "$WORK/proc/sys/net/ipv4/conf/usb1/forwarding")" = 0 ] \
    && note "IPv4 forwarding is off on usb1" OK || note "IPv4 forwarding is still on for usb1" FAIL
[ "$(cat "$WORK/proc/sys/net/ipv6/conf/usb1/forwarding")" = 0 ] \
    && note "IPv6 forwarding is off on usb1" OK || note "IPv6 forwarding is still on for usb1" FAIL
if grep -qE 'nat|MASQUERADE|SNAT|ACCEPT' "$WORK/calls"
then
    note "the start set up NAT or accepted forwarded traffic" FAIL
else
    note "no NAT and no accept rule" OK
fi
# The address goes on the link before udhcpd starts: udhcpd
# reads its server address from the interface and exits if there is none.
add_line=$(grep -nxF "ip addr add 172.31.255.1/30 dev usb1" "$WORK/calls" | cut -d: -f1)
dhcp_line=$(grep -n "^udhcpd " "$WORK/calls" | cut -d: -f1)
[ -n "$add_line" ] && [ -n "$dhcp_line" ] && [ "$add_line" -lt "$dhcp_line" ] \
    && note "the address is in place before udhcpd starts" OK \
    || note "udhcpd started at call [$dhcp_line], the address at [$add_line]" FAIL

# The rules are inserted once. A second start with the rules present checks
# and inserts nothing, so a board restarted a hundred times has two rules.
IPT_HAS=0 start_run ecm
not_called "iptables -I FORWARD -i usb+ -j DROP" "a rule that is present is not inserted again"
IPT_HAS=

# A subnet from the file reaches the interface and the lease.
echo 10.9.8.0/29 > "$WORK/subnet"
rm -rf "$WORK/run"; : > "$WORK/calls"
WORK="$WORK" CALLS="$WORK/calls" USB_GADGET_DIR="$G" USB_NET_RUN="$WORK/run" USB_NET_SUBNET_FILE="$WORK/subnet" \
    sh -c '. "$WORK/net.sh"; . "$WORK/stubs.sh"; usb_net_start ncm' > "$WORK/out" 2>&1
called "ip addr add 10.9.8.1/29 dev usb0" "the configured subnet puts 10.9.8.1/29 on the link"
grep -qx "start 10.9.8.2" "$WORK/run/udhcpd.conf" 2>/dev/null \
    && note "the host is offered 10.9.8.2" OK || note "the host is not offered 10.9.8.2" FAIL
grep -qx "option subnet 255.255.255.248" "$WORK/run/udhcpd.conf" 2>/dev/null \
    && note "the host is told the /29 mask" OK || note "the host is not told the /29 mask" FAIL
called "ip addr flush dev usb1" "starting NCM clears the ECM interface"

# A function that never bound has no interface. The link is lost, and nothing
# else is touched.
start_run rndis
grep -q "status=1" "$WORK/out" && note "a function with no interface returns 1" OK \
    || note "a function with no interface returned [$(cat "$WORK/out")]" FAIL
grep -q "rndis.usb0 has no interface" "$WORK/out" && note "and says which function" OK \
    || note "and says nothing useful" FAIL
if grep -q "^ip addr add\|^udhcpd" "$WORK/calls"
then
    note "an address or udhcpd was set up with no interface" FAIL
else
    note "no address and no udhcpd without an interface" OK
fi

# An address that will not go on stops before udhcpd, which would exit anyway.
IP_FAILS=1 start_run ecm
grep -q "status=1" "$WORK/out" && note "a failed address returns 1" OK \
    || note "a failed address returned [$(cat "$WORK/out")]" FAIL
not_called "udhcpd -S $WORK/run/udhcpd.conf" "udhcpd is not started without an address"
IP_FAILS=

echo
echo "===== stopping the link ====="
mkdir -p "$WORK/run"
echo 4242 > "$WORK/run/udhcpd.pid"
: > "$WORK/calls"
WORK="$WORK" CALLS="$WORK/calls" USB_GADGET_DIR="$G" USB_NET_RUN="$WORK/run" \
    sh -c '. "$WORK/net.sh"; . "$WORK/stubs.sh"; usb_net_stop ""' > "$WORK/out" 2>&1
called "kill 4242" "stop kills the link's udhcpd by its pid file"
[ -e "$WORK/run/udhcpd.pid" ] && note "the pid file is left behind" FAIL || note "the pid file is removed" OK
called "ip addr flush dev usb0" "stop clears usb0"
called "ip addr flush dev usb1" "stop clears usb1"
if grep -q "usb%d" "$WORK/calls"
then
    note "stop flushed the unbound template name" FAIL
else
    note "stop skips a function that never bound" OK
fi
if grep -q "wlan\|eth" "$WORK/calls"
then
    note "stop touched a LAN interface" FAIL
else
    note "stop touches USB interfaces alone" OK
fi

# The Wi-Fi access point runs udhcpd too. The link's stop kills by its own pid
# file and never by name.
if sed -n '/^# --- usb network ---$/,/^# --- end usb network ---$/p' "$SV" | grep -qE 'killall|pkill|grep udhcpd'
then
    note "the link stops udhcpd by name and would take the access point's with it" FAIL
else
    note "the link stops only its own udhcpd" OK
fi

echo
echo "===== wiring in start_usb_dev ====="
sh -n "$SV" && note "S03usbdev is valid shell" OK || note "S03usbdev does not parse" FAIL

# The block could be right and never called.
body=$(sed -n '/^start_usb_dev(){$/,/^}$/p' "$SV")
bind_line=$(printf '%s\n' "$body" | grep -nE '^    usb_bind$' | cut -d: -f1)
start_line=$(printf '%s\n' "$body" | grep -nE '^        usb_net_start "\$usb_net"$' | cut -d: -f1)
stop_line=$(printf '%s\n' "$body" | grep -nE '^        usb_net_stop ""$' | cut -d: -f1)
[ -n "$start_line" ] && note "start_usb_dev starts the link" OK || note "start_usb_dev never starts the link" FAIL
[ -n "$stop_line" ] && note "start_usb_dev stops the link when the network is dropped" OK \
    || note "a start without the network leaves the old link up" FAIL
[ -n "$bind_line" ] && [ -n "$start_line" ] && [ "$bind_line" -lt "$start_line" ] \
    && note "the link starts after the bind, when the interface exists" OK \
    || note "the link starts at [$start_line], the bind is at [$bind_line]" FAIL

sed -n '/^start_usb_host(){$/,/^}$/p' "$SV" | grep -qE '^    usb_net_stop ""$' \
    && note "stop takes the link down with the gadget" OK \
    || note "stop leaves udhcpd serving an unbound gadget" FAIL

# Nothing in the script may turn forwarding on.
if grep -qE 'ip_forward|forwarding' "$SV" && grep -E 'ip_forward|forwarding' "$SV" | grep -qE 'echo 1'
then
    note "the script turns forwarding on somewhere" FAIL
else
    note "the script never turns forwarding on" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== all usb network cases pass ====="
else
    echo "===== $fails case(s) failed ====="
    exit 1
fi
