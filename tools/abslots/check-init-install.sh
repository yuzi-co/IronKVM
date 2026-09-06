#!/bin/sh
# Check that every boot script the install package carries has a decided fate.
#
#   check-init-install.sh <base-init.d-dir> [package-init.d-dir] [manifest] [list]
#
# /etc/init.d is what boots. kvmapp/system/init.d is the application's own
# reference copy. A script that lands in only one of them does not run, and
# nothing about the package or the image looks wrong when that happens.
#
# On 2026-09-04 the fork rewrote S00kmod so the modules in the install package
# are loaded before the stock ones. The image manifest did not install S00kmod,
# because until that commit the fork did not change it, so every image built
# afterwards kept the stock loader and the patched soph_vi.ko it exists to load
# was never inserted. Nothing failed. The fix simply was not there.
#
# This is the check that would have caught it. Every script in the package is
# either installed by root.manifest or declared in init.d.package-only, and a
# script declared same-as-base has to still be the base's bytes.
#
# Exit 0 when every script is accounted for, 1 when one is not, and 2 when the
# check cannot run: no base directory to compare against.

BASE_DIR=${1:-}
PKG_DIR=${2:-kvmapp/system/init.d}
MANIFEST=${3:-tools/abslots/manifest/root.manifest}
LIST=${4:-tools/abslots/manifest/init.d.package-only}

[ -n "$BASE_DIR" ] || {
    echo "usage: check-init-install.sh <base-init.d-dir> [pkg-dir] [manifest] [list]" >&2
    exit 2; }
[ -d "$BASE_DIR" ] || { echo "no base init.d directory: $BASE_DIR" >&2; exit 2; }
[ -d "$PKG_DIR" ]  || { echo "no package init.d directory: $PKG_DIR" >&2; exit 1; }
[ -f "$MANIFEST" ] || { echo "no manifest: $MANIFEST" >&2; exit 1; }
[ -f "$LIST" ]     || { echo "no package-only list: $LIST" >&2; exit 1; }

fail=0
note() { echo "check-init-install: $*" >&2; fail=1; }

# The scripts the image installs into /etc/init.d, by destination name. The
# manifest also installs scripts from tools/, and those are not in the package
# directory this checks, so matching on the destination is what makes the two
# lists comparable.
installed=$(sed -n 's|^add .* /etc/init.d/\([^ ]*\).*|\1|p' "$MANIFEST")

# The declarations, with comments and blank lines removed.
declared=$(sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$LIST")

has() {
    for entry in $2; do [ "$entry" = "$1" ] && return 0; done
    return 1
}

reason_for() {
    echo "$declared" | awk -v want="$1" '$1 == want { print $2; exit }'
}

for path in "$PKG_DIR"/*; do
    [ -f "$path" ] || continue
    name=${path##*/}

    reason=$(reason_for "$name")

    if has "$name" "$installed"; then
        [ -z "$reason" ] || note "$name is installed by the manifest and also declared package-only; one of the two is wrong"
        continue
    fi

    if [ -z "$reason" ]; then
        note "$name is in the package and no image installs it: add it to ${MANIFEST##*/}, or say why not in ${LIST##*/}"
        continue
    fi

    case "$reason" in
        not-installed-by-policy)
            ;;
        same-as-base)
            if [ ! -f "$BASE_DIR/$name" ]; then
                note "$name is declared same-as-base and the base does not carry it"
            elif ! cmp -s "$path" "$BASE_DIR/$name"; then
                note "$name is declared same-as-base and the fork now changes it: install it from ${MANIFEST##*/}, or the change does not reach /etc/init.d"
            fi
            ;;
        *)
            note "$name is declared with an unknown reason '$reason'"
            ;;
    esac
done

# A declaration for a script the package no longer carries is a stale line, and
# a stale line is what makes the next reader distrust the whole file.
echo "$declared" | while read -r name _rest; do
    [ -n "$name" ] || continue
    [ -f "$PKG_DIR/$name" ] || echo "check-init-install: ${LIST##*/} names $name and the package does not carry it" >&2
done

# The loop above runs in a subshell, so it reports without being able to set
# fail. Count the same condition again here, where the result can be used.
for name in $(echo "$declared" | awk '{ print $1 }'); do
    [ -f "$PKG_DIR/$name" ] || fail=1
done

[ "$fail" = 0 ] || exit 1

count=$(echo "$installed" | grep -c .)
echo "check-init-install: every script in $PKG_DIR is installed or declared ($count installed to /etc/init.d)"
