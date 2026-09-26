#!/bin/sh
# The Go memory limit S98tailscaled gives tailscaled.
#
#   test-tailscale-memlimit.sh
#
# The limit is the owner's /etc/kvm/GOMEMLIMIT, or 512 MiB without it, capped
# at seven eighths of the addons group's memory.high. Only the block between
# "# --- memlimit ---" and "# --- end memlimit ---" runs here, never the script.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
S=$HERE/../../kvmapp/system/init.d/S98tailscaled
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

sed -n '/^# --- memlimit ---$/,/^# --- end memlimit ---$/p' "$S" > "$WORK/block.sh"
check "the block can be extracted" "$([ -s "$WORK/block.sh" ] && echo yes)" "yes"
check "start uses it" "$(grep -c 'GOMEMLIMIT="$(tailscale_memlimit_mib)MiB"' "$S")" "1"

# limit <owner setting or -> <memory.high or ->
limit() {
    rm -rf "$WORK/kvm" "$WORK/cg"; mkdir -p "$WORK/kvm" "$WORK/cg"
    [ "$1" = - ] || printf '%s\n' "$1" > "$WORK/kvm/GOMEMLIMIT"
    [ "$2" = - ] || printf '%s\n' "$2" > "$WORK/cg/memory.high"
    KVMDIR=$WORK/kvm CGROUP_PROCS=$WORK/cg/cgroup.procs \
        sh -c '. "$1"; tailscale_memlimit_mib' _ "$WORK/block.sh"
}

check "no setting, no group: 512" "$(limit - -)" "512"
check "an owner's 75, no group: 75" "$(limit 75 -)" "75"
check "no setting, memory.high 64M: 56" "$(limit - 67108864)" "56"
check "an owner's 75, memory.high 64M: 56" "$(limit 75 67108864)" "56"
check "an owner's 40, memory.high 64M: 40" "$(limit 40 67108864)" "40"
check "memory.high max (no limit): the setting stands" "$(limit 75 max)" "75"
check "a setting that is not a number: 512" "$(limit 75MiB -)" "512"
check "a setting of 0: 512" "$(limit 0 -)" "512"
check "a memory.high under 2 MiB caps nothing" "$(limit 75 1048576)" "75"

[ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
[ "$fails" -eq 0 ]
