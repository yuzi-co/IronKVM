#!/bin/sh
# S30eth records, once per boot, how long the Ethernet link took to come up.
#
#   test-eth-link-record.sh [path-to-S30eth]
#
# On the reference board the link has come up anywhere from a few seconds to
# 94s after the PHY was configured, while rcS finished at 12 to 25s every time.
# The wait decides when the board is reachable, and nothing kept it: dmesg and
# /watchdog.log are per boot or per slot, so every slot switch lost the record.
# /data/eth-link.log keeps one line per boot.
#
# Only the recorder is lifted out of the script and run. S30eth itself flushes
# addresses and starts a DHCP client, and on the reference board that is the
# network this suite is reached over.
S30=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S30eth}
[ -f "$S30" ] || { echo "no S30eth at $S30"; exit 2; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
t() { if [ "$2" = 0 ]; then echo "ok   - $1"; else echo "FAIL - $1"; fails=1; fi; }

sed -n '/^# --- link record ---$/,/^# --- end link record ---$/p' "$S30" > "$WORK/record.sh"
[ -s "$WORK/record.sh" ] || { echo "FAIL - S30eth has no link record block"; exit 1; }

mkdir -p "$WORK/net/eth0" "$WORK/run"
echo 100 > "$WORK/net/eth0/speed"

rec() {
    IF=eth0 SYS_NET="$WORK/net" ETH_RUN_DIR="$WORK/run" \
    ETH_LINK_LOG="$WORK/eth-link.log" ETH_DATA_MOUNTED="${MOUNTED:-yes}" \
    ETH_BOOT_ID="${BOOT:-boot-1}" ETH_LINK_WAIT="${WAIT:-5}" \
        sh -c '. "$1"; eth_record_link' sh "$WORK/record.sh" > /dev/null 2>&1
}
lines() { wc -l < "$WORK/eth-link.log" 2>/dev/null | tr -d ' ' || echo 0; }

echo 1 > "$WORK/net/eth0/carrier"
rec
grep -q ' up=[0-9][0-9]*s speed=100 ' "$WORK/eth-link.log" 2>/dev/null
t "a link that is up records the uptime and the speed" $?
grep -q ' boot=boot-1$' "$WORK/eth-link.log" 2>/dev/null
t "the line names the boot, so a second line from one boot is visible" $?

rec
[ "$(lines)" = 1 ]
t "a second run in the same boot writes nothing" $?

BOOT=boot-2 rec
[ "$(lines)" = 2 ]
t "the next boot writes its own line" $?

echo 0 > "$WORK/net/eth0/carrier"
BOOT=boot-3 WAIT=2 rec
tail -1 "$WORK/eth-link.log" | grep -q 'no link after 2s'
t "a link that never comes up says so at the deadline" $?

echo 1 > "$WORK/net/eth0/carrier"
BOOT=boot-4 MOUNTED=no rec
[ "$(lines)" = 3 ]
t "nothing is written when /data is not mounted" $?

exit $fails
