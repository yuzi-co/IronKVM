#!/bin/sh
# The Go memory limit S96picoclaw gives the PicoClaw gateway.
#
#   test-picoclaw-memlimit.sh
#
# The limit is the owner's /etc/kvm/GOMEMLIMIT.picoclaw, or 32 MiB without it,
# capped at seven eighths of what the addons group's memory.high leaves after
# the VPN daemon's 56 MiB, and never under 8. Only the block between
# "# --- memlimit ---" and "# --- end memlimit ---" runs here, never the script.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
S=$HERE/../../kvmapp/system/init.d/S96picoclaw
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

# limit <owner setting or -> <memory.high or ->
limit() {
    rm -rf "$WORK/kvm" "$WORK/cg"; mkdir -p "$WORK/kvm" "$WORK/cg"
    [ "$1" = - ] || printf '%s\n' "$1" > "$WORK/kvm/GOMEMLIMIT.picoclaw"
    [ "$2" = - ] || printf '%s\n' "$2" > "$WORK/cg/memory.high"
    KVMDIR=$WORK/kvm CGROUP_PROCS=$WORK/cg/cgroup.procs \
        sh -c '. "$1"; memlimit_mib' _ "$WORK/block.sh"
}

echo "S96picoclaw"
sed -n '/^# --- memlimit ---$/,/^# --- end memlimit ---$/p' "$S" > "$WORK/block.sh"
check "the block can be extracted" "$([ -s "$WORK/block.sh" ] && echo yes)" "yes"
check "the gateway starts under it" "$(grep -c 'GOMEMLIMIT="$(memlimit_mib)MiB" "$BIN_PATH" gateway' "$S")" "1"
check "the VPN daemons' setting is not read" "$(grep -c 'KVMDIR/GOMEMLIMIT"' "$WORK/block.sh")" "0"

check "no setting, no group: 32" "$(limit - -)" "32"
check "no setting, memory.high 96M: 32" "$(limit - 100663296)" "32"
check "an owner's 48, memory.high 96M: 35" "$(limit 48 100663296)" "35"
check "an owner's 20, memory.high 96M: 20" "$(limit 20 100663296)" "20"
check "an owner's 48, no group: 48" "$(limit 48 -)" "48"
check "no setting, memory.high 80M: 21" "$(limit - 83886080)" "21"
check "no setting, memory.high 128M: 32" "$(limit - 134217728)" "32"
check "an owner's 100, memory.high 128M: 63" "$(limit 100 134217728)" "63"
check "no setting, memory.high 64M: the floor of 8" "$(limit - 67108864)" "8"
check "no setting, memory.high 32M: the floor of 8" "$(limit - 33554432)" "8"
check "memory.high max (no limit): the setting stands" "$(limit 48 max)" "48"
check "a setting that is not a number: 32" "$(limit 48MiB -)" "32"
check "a setting of 0: 32" "$(limit 0 -)" "32"

[ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
[ "$fails" -eq 0 ]
