# Shared by the start scripts of the host add-ons (S85netboot, S96picoclaw,
# S98netbird, S98tailscaled), which source it. It is not run on its own.
#
# ironkvm-dist's S01cgroups makes this tree under /sys/fs/cgroup:
#
#   kvm/                  the server and kvm_system
#   addons/               memory.high, memory.max and pids.max for every
#                         add-on together; it holds no process
#     daemons/<name>/     one leaf per host add-on, made here
#     lxc/                where lxc-start runs, for the containers
#
# cgroup2 lets a group hand its controllers to child groups only while it
# holds no process itself, so addons/ and addons/daemons/ stay empty and each
# add-on joins a leaf of its own. A leaf's memory.max then lets one add-on's
# leak kill that add-on alone, while the limits on addons/ still bound them
# all.
#
# addon_cgroup_join <name> <memory.max> moves the calling shell into
# addons/daemons/<name>, making the leaf if it is not there, and writes the
# given memory.max to it ("max" for none of its own, or a size such as 48M).
# The daemon the script then starts inherits the group, however the script was
# run: at boot, or from the web UI through the server, whose kvm group it would
# otherwise inherit.
#
# Where that cannot be done it falls back, in order:
#
#   - addons/ without daemons/ (an image whose S01cgroups predates the leaves):
#     it joins addons/ itself, which then hands no controllers down.
#   - addons/ that refuses the shell (the leaf could not be made): it joins
#     the root group, so that a script started by the server does not stay in
#     kvm, whose memory.min would protect the add-on.
#   - no addons/ at all (a kernel without cgroup2, or no S01cgroups): nothing
#     changes, as before the groups existed.
#
# Nothing here prints or fails: a caller under set -e or set -u goes on.
# ADDONS_CGROUP is a variable so the tests can point it at a scratch tree.
ADDONS_CGROUP=${ADDONS_CGROUP:-/sys/fs/cgroup/addons}

# _addon_cgroup_enter <group>: write this shell's pid to the group's
# cgroup.procs, and report whether the kernel took it.
_addon_cgroup_enter() {
    [ -w "$1/cgroup.procs" ] || return 1
    { echo $$ > "$1/cgroup.procs"; } 2>/dev/null
}

addon_cgroup_join() {
    [ -d "$ADDONS_CGROUP" ] || return 0
    if [ -n "${1:-}" ] && [ -d "$ADDONS_CGROUP/daemons" ]; then
        _addon_cgroup_leaf=$ADDONS_CGROUP/daemons/$1
        [ -d "$_addon_cgroup_leaf" ] || mkdir "$_addon_cgroup_leaf" 2>/dev/null || true
        # The limit goes on before the shell joins, so the daemon never runs
        # under a stale one. Without the memory controller there is no file.
        if [ -n "${2:-}" ] && [ -w "$_addon_cgroup_leaf/memory.max" ]; then
            { echo "$2" > "$_addon_cgroup_leaf/memory.max"; } 2>/dev/null || true
        fi
        _addon_cgroup_enter "$_addon_cgroup_leaf" && return 0
    fi
    _addon_cgroup_enter "$ADDONS_CGROUP" && return 0
    _addon_cgroup_enter "${ADDONS_CGROUP%/*}" || true
    return 0
}
