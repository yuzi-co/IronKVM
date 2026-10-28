#!/bin/sh
# Each start script joins its memory group before it starts anything.
#
#   test-cgroup-join.sh
#
# ironkvm-dist's S01cgroups makes two groups, kvm and addons, and under addons
# the parent of one leaf per host add-on, addons/daemons. S95nanokvm joins kvm;
# S98tailscaled, S98netbird, S96picoclaw and S85netboot each join their own
# leaf through kvmapp/system/addon-cgroup.sh. The join is the block between
# "# --- cgroup ---" and "# --- end cgroup ---" in each script.
# Only that block is run here, never the script, because running an init
# script starts or stops real services. test-addon-cgroup.sh tests the helper
# itself and its fallbacks.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
INITD=$HERE/../../kvmapp/system/init.d
HELPER=$HERE/../../kvmapp/system/addon-cgroup.sh
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

# extract <script>: the block, into $WORK/block.sh, and a check that it comes
# before the script's case statement, so start, restart and every path through
# them run inside the group.
extract() {
    sed -n '/^# --- cgroup ---$/,/^# --- end cgroup ---$/p' "$INITD/$1" > "$WORK/block.sh"
    check "the block can be extracted" "$([ -s "$WORK/block.sh" ] && echo yes)" "yes"
    b=$(grep -n '^# --- cgroup ---$' "$INITD/$1" | cut -d: -f1)
    c=$(grep -n '^case "\$1" in' "$INITD/$1" | head -1 | cut -d: -f1)
    check "it comes before the case statement" "$([ -n "$b" ] && [ -n "$c" ] && [ "$b" -lt "$c" ] && echo yes)" "yes"
}

echo "S95nanokvm"
extract S95nanokvm
check "it names the kvm group" "$(grep -c "/sys/fs/cgroup/kvm/cgroup.procs" "$WORK/block.sh")" "1"
: > "$WORK/procs"
CGROUP_PROCS=$WORK/procs sh -c '. "$1"; echo "$$" > "$2"' _ "$WORK/block.sh" "$WORK/pid"
check "with the group, it writes its own pid" "$(cat "$WORK/procs")" "$(cat "$WORK/pid")"
CGROUP_PROCS=$WORK/absent/cgroup.procs sh -c '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
check "without the group, it goes on silently" "$(cat "$WORK/out")" "reached"
mkdir -p "$WORK/dir-as-file"
CGROUP_PROCS=$WORK/dir-as-file sh -ec '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
check "a failed write does not end the script" "$(cat "$WORK/out")" "reached"

# The tree S01cgroups makes, with the files the kernel would put in each group.
# A directory stands in for cgroup2, so a leaf the helper makes has no files;
# the leaf is made here instead, as the kernel would have filled it.
tree() {
    rm -rf "$WORK/cg"
    mkdir -p "$WORK/cg/addons/daemons/$1"
    : > "$WORK/cg/cgroup.procs"
    : > "$WORK/cg/addons/cgroup.procs"
    : > "$WORK/cg/addons/daemons/$1/cgroup.procs"
    echo max > "$WORK/cg/addons/daemons/$1/memory.max"
}

for pair in S98tailscaled:tailscale S98netbird:netbird S96picoclaw:picoclaw S85netboot:netboot; do
    s=${pair%%:*}; leaf=${pair#*:}
    echo "$s"
    extract "$s"
    check "it joins addons/daemons/$leaf with no limit of its own" \
        "$(grep -c "addon_cgroup_join $leaf max\$" "$WORK/block.sh")" "1"
    check "it sources the shared helper by default" \
        "$(grep -c 'ADDON_CGROUP_SH:-/kvmapp/system/addon-cgroup.sh}' "$WORK/block.sh")" "1"
    check "it no longer writes a group's cgroup.procs itself" "$(grep -c 'cgroup.procs' "$WORK/block.sh")" "0"

    tree "$leaf"
    ADDON_CGROUP_SH=$HELPER ADDONS_CGROUP=$WORK/cg/addons \
        sh -c '. "$1"; echo "$$" > "$2"' _ "$WORK/block.sh" "$WORK/pid"
    check "with the tree, it writes its own pid to its leaf" \
        "$(cat "$WORK/cg/addons/daemons/$leaf/cgroup.procs")" "$(cat "$WORK/pid")"
    check "and to no other group" "$(cat "$WORK/cg/addons/cgroup.procs" "$WORK/cg/cgroup.procs")" ""
    check "the leaf's memory.max is max" "$(cat "$WORK/cg/addons/daemons/$leaf/memory.max")" "max"

    ADDON_CGROUP_SH=$HELPER ADDONS_CGROUP=$WORK/absent \
        sh -c '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
    check "without the group, it goes on silently" "$(cat "$WORK/out")" "reached"

    ADDON_CGROUP_SH=$WORK/no-helper.sh ADDONS_CGROUP=$WORK/cg/addons \
        sh -euc '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
    check "without the helper, it goes on silently" "$(cat "$WORK/out")" "reached"

    # set -e and set -u in a caller must not end the script when a write fails.
    tree "$leaf"
    rm "$WORK/cg/addons/daemons/$leaf/cgroup.procs" "$WORK/cg/addons/cgroup.procs" "$WORK/cg/cgroup.procs"
    mkdir "$WORK/cg/addons/daemons/$leaf/cgroup.procs" "$WORK/cg/addons/cgroup.procs" "$WORK/cg/cgroup.procs"
    ADDON_CGROUP_SH=$HELPER ADDONS_CGROUP=$WORK/cg/addons \
        sh -euc '. "$1"; echo reached' _ "$WORK/block.sh" > "$WORK/out" 2>&1
    check "a failed write does not end the script" "$(cat "$WORK/out")" "reached"
done

[ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
[ "$fails" -eq 0 ]
