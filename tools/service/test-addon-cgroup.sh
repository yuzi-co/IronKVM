#!/bin/sh
# kvmapp/system/addon-cgroup.sh, the helper the add-on start scripts source to
# join their leaf of ironkvm-dist's addons memory group.
#
#   test-addon-cgroup.sh [path-to-addon-cgroup.sh]
#
# A scratch directory stands in for /sys/fs/cgroup. Where the kernel would
# fill a new group with its files, the test makes them by hand. A directory
# named cgroup.procs stands in for a group that refuses a process: the shell
# can test it as writable, as with the real file, and the write then fails.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
HELPER=${1:-$HERE/../../kvmapp/system/addon-cgroup.sh}
[ -f "$HELPER" ] || { echo "no helper at $HELPER"; exit 1; }
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

# tree [daemons]: the root and addons/ as S01cgroups leaves them, with
# addons/daemons unless the argument is "old" (an S01cgroups before the leaves).
tree() {
    rm -rf "$WORK/cg"
    mkdir -p "$WORK/cg/addons"
    : > "$WORK/cg/cgroup.procs"
    : > "$WORK/cg/addons/cgroup.procs"
    [ "${1:-}" = old ] || mkdir "$WORK/cg/addons/daemons"
}

# leaf <name>: the leaf as the kernel would show it after mkdir.
leaf() {
    mkdir -p "$WORK/cg/addons/daemons/$1"
    : > "$WORK/cg/addons/daemons/$1/cgroup.procs"
    echo max > "$WORK/cg/addons/daemons/$1/memory.max"
}

refuse() { rm -f "$1/cgroup.procs"; mkdir "$1/cgroup.procs"; }

# join <name> <memory.max> [sh flags]: source the helper and join, then print
# the shell's pid to $WORK/pid and "reached" to $WORK/out.
join() {
    ADDONS_CGROUP=$WORK/cg/addons sh ${3:+"$3"} -c \
        '. "$1"; addon_cgroup_join "$2" "$3"; echo "$$" > "$4"; echo reached' \
        _ "$HELPER" "$1" "$2" "$WORK/pid" > "$WORK/out" 2>&1
}

procs() { cat "$WORK/cg/$1/cgroup.procs" 2>/dev/null; }

echo "a leaf that is there"
tree; leaf tailscale
join tailscale 48M
check "the shell joins the leaf" "$(procs addons/daemons/tailscale)" "$(cat "$WORK/pid")"
check "and no other group" "$(procs addons)$(procs .)" ""
check "the leaf gets the caller's memory.max" "$(cat "$WORK/cg/addons/daemons/tailscale/memory.max")" "48M"
check "it prints nothing" "$(cat "$WORK/out")" "reached"

echo "a limit of max"
tree; leaf netbird
echo 32M > "$WORK/cg/addons/daemons/netbird/memory.max"
join netbird max
check "replaces a limit left from before" "$(cat "$WORK/cg/addons/daemons/netbird/memory.max")" "max"

echo "no limit given"
tree; leaf netboot
echo 32M > "$WORK/cg/addons/daemons/netboot/memory.max"
join netboot ""
check "leaves the leaf's limit alone" "$(cat "$WORK/cg/addons/daemons/netboot/memory.max")" "32M"
check "and still joins" "$(procs addons/daemons/netboot)" "$(cat "$WORK/pid")"

echo "a leaf that is not there yet"
tree
join picoclaw max
check "the helper makes it" "$([ -d "$WORK/cg/addons/daemons/picoclaw" ] && echo yes)" "yes"
# The scratch leaf has no cgroup.procs, where the kernel would have made one,
# so the join falls through to addons/: what a leaf that cannot be made does
# on an older image.
check "and, with no cgroup.procs in it, falls back to addons/" "$(procs addons)" "$(cat "$WORK/pid")"

echo "a leaf that cannot be made"
tree
chmod 555 "$WORK/cg/addons/daemons"
refuse "$WORK/cg/addons"
join picoclaw max -eu
chmod 755 "$WORK/cg/addons/daemons"
check "the shell goes on under set -eu" "$(cat "$WORK/out")" "reached"
check "addons/ refuses it, so it joins the root group" "$(procs .)" "$(cat "$WORK/pid")"

echo "an older image: addons/ without daemons/"
tree old
join tailscale max
check "the shell joins addons/ itself, as before" "$(procs addons)" "$(cat "$WORK/pid")"
check "and makes no daemons/" "$([ -e "$WORK/cg/addons/daemons" ] && echo made || echo none)" "none"

echo "no addons/ (a kernel without cgroup2, or no S01cgroups)"
rm -rf "$WORK/cg"; mkdir -p "$WORK/cg"; : > "$WORK/cg/cgroup.procs"
join tailscale max -eu
check "the shell goes on" "$(cat "$WORK/out")" "reached"
check "and joins nothing, not even the root group" "$(procs .)" ""
check "and makes nothing" "$(find "$WORK/cg" -mindepth 1 | sed "s,^$WORK/cg/,,")" "cgroup.procs"

echo "every group refuses"
tree; leaf tailscale
refuse "$WORK/cg/addons/daemons/tailscale"; refuse "$WORK/cg/addons"; refuse "$WORK/cg"
chmod 444 "$WORK/cg/addons/daemons/tailscale/memory.max"
join tailscale max -eu
chmod 644 "$WORK/cg/addons/daemons/tailscale/memory.max"
check "the shell goes on under set -eu, silently" "$(cat "$WORK/out")" "reached"

echo "no name"
tree
join "" max -eu
check "the shell goes on" "$(cat "$WORK/out")" "reached"
check "it makes no leaf" "$(find "$WORK/cg/addons/daemons" -mindepth 1 | wc -l | tr -d ' ')" "0"
check "and joins addons/" "$(procs addons)" "$(cat "$WORK/pid")"

[ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
[ "$fails" -eq 0 ]
