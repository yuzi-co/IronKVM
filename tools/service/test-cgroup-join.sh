#!/bin/sh
# Each start script joins its memory group before it starts anything.
#
#   test-cgroup-join.sh
#
# ironkvm-dist's S01cgroups makes two groups, kvm and addons. S95nanokvm joins
# kvm; S98tailscaled and S96picoclaw join addons. The join is the block between
# "# --- cgroup ---" and "# --- end cgroup ---" in each script. Only that block
# is run here, never the script, because running an init script starts or stops
# real services.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
INITD=$HERE/../../kvmapp/system/init.d
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

for pair in S95nanokvm:kvm S98tailscaled:addons S96picoclaw:addons; do
    s=${pair%%:*}; g=${pair#*:}
    echo "$s"
    sed -n '/^# --- cgroup ---$/,/^# --- end cgroup ---$/p' "$INITD/$s" > "$WORK/block.sh"
    check "the block can be extracted" "$([ -s "$WORK/block.sh" ] && echo yes)" "yes"
    check "it names the $g group" "$(grep -c "/sys/fs/cgroup/$g/cgroup.procs" "$WORK/block.sh")" "1"

    # The block runs before the script's case statement, so start, restart and
    # every path through them run inside the group.
    b=$(grep -n '^# --- cgroup ---$' "$INITD/$s" | cut -d: -f1)
    c=$(grep -n '^case "\$1" in' "$INITD/$s" | head -1 | cut -d: -f1)
    check "it comes before the case statement" "$([ -n "$b" ] && [ -n "$c" ] && [ "$b" -lt "$c" ] && echo yes)" "yes"

    : > "$WORK/procs"
    CGROUP_PROCS=$WORK/procs sh -c '. "$1"; echo "$$" > "$2"' _ "$WORK/block.sh" "$WORK/pid"
    check "with the group, it writes its own pid" "$(cat "$WORK/procs")" "$(cat "$WORK/pid")"

    CGROUP_PROCS=$WORK/absent/cgroup.procs sh -c '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
    check "without the group, it goes on silently" "$(cat "$WORK/out")" "reached"

    # set -e in a caller must not end the script when the write fails.
    mkdir -p "$WORK/dir-as-file"
    CGROUP_PROCS=$WORK/dir-as-file sh -ec '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
    check "a failed write does not end the script" "$(cat "$WORK/out")" "reached"
done

[ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
[ "$fails" -eq 0 ]
