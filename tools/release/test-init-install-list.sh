#!/bin/sh
# Every init script the application tarball carries is either installed into
# /etc/init.d by install.sh or declared package-only with a reason.
#
# This is the half of check-init-install.sh that belongs to this repository.
# The other half, that the image manifest installs the same set, moved to
# ironkvm-dist with the manifest.
set -u
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
PKG=${PKG:-$ROOT/kvmapp/system/init.d}
INSTALL=${INSTALL:-$ROOT/kvmapp/system/init.d.install}
ONLY=${ONLY:-$ROOT/kvmapp/system/init.d.package-only}
fails=0
t() { if [ "$2" = 0 ]; then echo "ok   - $1"; else echo "FAIL - $1"; fails=1; fi; }

installed=$(sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$INSTALL")
declared=$(sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$ONLY" | awk '{print $1}')

for path in "$PKG"/*; do
    [ -f "$path" ] || continue
    name=${path##*/}
    in_install=$(echo "$installed" | grep -cx "$name")
    in_only=$(echo "$declared" | grep -cx "$name")
    [ $((in_install + in_only)) -eq 1 ]
    t "$name is installed or package-only, exactly once (install=$in_install only=$in_only)" $?
done

# init.d.install also names scripts that the package takes from tools/ rather
# than from kvmapp/system/init.d. release.sh copies them into the tarball's
# system/init.d, so the tarball carries them even though this checkout holds
# them somewhere else.
for name in $installed $declared; do
    [ -f "$PKG/$name" ] || [ -f "$ROOT/tools/service/$name" ] || [ -f "$ROOT/tools/oled/$name" ]
    t "$name, named in a list, is in the package or in tools/" $?
done

exit $fails
