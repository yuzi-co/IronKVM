#!/bin/sh
# Check that S95nanokvm puts the kernel's own libkvm.so in the server's dl_lib,
# and that an application update cannot leave the wrong one there.
#
#   test-libkvm-select.sh [path-to-S95nanokvm]
#
# Environment:
#   DEVICEINFO_BASE   a device description to run against, with its KERNEL
#                     line replaced for each case. ironkvm-dist passes its
#                     devices/fake-board/deviceinfo. Default: a minimal one.
#
# The real reader is used behind a stub, as in test-server-gogc.sh. pidof is a
# stub on PATH, so no case depends on what runs on this machine.
set -u

HERE=$(cd "$(dirname "$0")" && pwd)
S95=${1:-$HERE/../../kvmapp/system/init.d/S95nanokvm}
SUPERVISE=${SUPERVISE:-$HERE/S98supervise}
INSTALL=${INSTALL:-$HERE/../../kvmapp/system/install.sh}
READER=${READER:-$(cd "$HERE/../.." && pwd)/kvmapp/system/ironkvm-deviceinfo}
[ -f "$S95" ] || { echo "usage: test-libkvm-select.sh <S95nanokvm>"; exit 1; }
[ -f "$READER" ] || { echo "needs kvmapp/system/ironkvm-deviceinfo"; exit 2; }

fails=0
note() { printf '  %-66s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

sed -n '/^# --- Go tuning ---/,/^# --- end Go tuning ---/p' "$S95" > "$work/blocks.sh"
sed -n '/^# --- libkvm ---/,/^# --- end libkvm ---/p' "$S95" >> "$work/blocks.sh"
grep -q '^place_libkvm() {' "$work/blocks.sh" \
    && note "S95nanokvm carries the libkvm block" OK \
    || { note "S95nanokvm carries the libkvm block" FAIL; echo "$fails case(s) FAILED"; exit 1; }

mkdir -p "$work/bin"
cat > "$work/bin/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$work/deviceinfo" exec sh "$READER" "\$@"
STUB
cat > "$work/bin/pidof" <<'STUB'
#!/bin/sh
[ -f "$PIDOF_RUNNING" ] && echo 1234 && exit 0
exit 1
STUB
chmod 755 "$work/bin/ironkvm-deviceinfo" "$work/bin/pidof"
PATH=$work/bin:$PATH
export PATH
PIDOF_RUNNING=$work/running
export PIDOF_RUNNING

if [ -n "${DEVICEINFO_BASE:-}" ]; then
    [ -f "$DEVICEINFO_BASE" ] || { echo "DEVICEINFO_BASE names $DEVICEINFO_BASE, which is not a file"; exit 2; }
    BASE=$DEVICEINFO_BASE
    echo "  (device description: $BASE)"
else
    printf 'DEVICE=test-board\nKERNEL=vendor-prebuilt\n' > "$work/base"
    BASE=$work/base
fi
# kernel <value>: the description for the next case, with KERNEL replaced.
kernel() {
    grep -v '^KERNEL=' "$BASE" > "$work/deviceinfo"
    echo "KERNEL=$1" >> "$work/deviceinfo"
}

printf 'sipeed mpi library\n' > "$work/vendor.so"
printf 'v4l2 library\n'       > "$work/v4l2.so"

# slot <libraries...>: a fresh slot root with these in /usr/lib/ironkvm and
# Sipeed's library in /kvmapp/server/dl_lib, which is what every release
# tarball carries.
slot() {
    rm -rf "$work/root" "$work/running"
    mkdir -p "$work/root/usr/lib/ironkvm" "$work/root/kvmapp/server/dl_lib"
    cp "$work/vendor.so" "$work/root/kvmapp/server/dl_lib/libkvm.so"
    for l in "$@"; do
        case "$l" in
            v4l2)   cp "$work/v4l2.so"   "$work/root/usr/lib/ironkvm/libkvm-v4l2.so" ;;
            vendor) cp "$work/vendor.so" "$work/root/usr/lib/ironkvm/libkvm-vendor.so" ;;
        esac
    done
}
place() {
    (
        DEVINFO=$work/bin/ironkvm-deviceinfo
        LIBKVM_DIR=$work/root/usr/lib/ironkvm
        LIBKVM_DST=$work/root/kvmapp/server/dl_lib/libkvm.so
        export DEVINFO LIBKVM_DIR LIBKVM_DST
        . "$work/blocks.sh"
        place_libkvm
    ) > "$work/out" 2>&1
}
dl() { cat "$work/root/kvmapp/server/dl_lib/libkvm.so"; }

echo "===== the library follows KERNEL ====="
for k in mainline vendor-prebuilt source; do
    kernel "$k"
    slot v4l2
    place; st=$?
    case "$k" in
        mainline)
            [ "$st" = 0 ] && [ "$(dl)" = "v4l2 library" ] \
                && note "KERNEL=$k: libkvm-v4l2.so goes into dl_lib" OK \
                || note "KERNEL=$k: libkvm-v4l2.so goes into dl_lib (got $st, '$(dl)')" FAIL ;;
        *)
            [ "$st" = 0 ] && [ "$(dl)" = "sipeed mpi library" ] \
                && note "KERNEL=$k: a slot with only libkvm-v4l2.so keeps the application's" OK \
                || note "KERNEL=$k: a slot with only libkvm-v4l2.so keeps the application's (got $st, '$(dl)')" FAIL ;;
    esac
done

kernel vendor-prebuilt; slot v4l2 vendor
printf 'a newer sipeed library\n' > "$work/root/kvmapp/server/dl_lib/libkvm.so"
place
[ "$(dl)" = "sipeed mpi library" ] \
    && note "KERNEL=vendor-prebuilt: a slot that installs libkvm-vendor.so gets it" OK \
    || note "KERNEL=vendor-prebuilt: a slot that installs libkvm-vendor.so gets it ('$(dl)')" FAIL

kernel mainline; slot
place; st=$?
[ "$st" = 0 ] && [ "$(dl)" = "sipeed mpi library" ] \
    && note "a slot that installs no library keeps whatever /kvmapp carries" OK \
    || note "a slot that installs no library keeps whatever /kvmapp carries (got $st)" FAIL

kernel mainline; slot v4l2
place
[ -x "$work/root/kvmapp/server/dl_lib/libkvm.so" ] || [ "$(uname -o 2>/dev/null)" = Msys ] \
    && note "the placed library is executable, as the rest of dl_lib" OK \
    || note "the placed library is executable, as the rest of dl_lib" FAIL
[ ! -e "$work/root/kvmapp/server/dl_lib/libkvm.so.new" ] \
    && note "no temporary file is left behind" OK \
    || note "no temporary file is left behind" FAIL
place
[ ! -s "$work/out" ] \
    && note "a library already in place is not copied again, silently" OK \
    || note "a library already in place is not copied again ($(cat "$work/out"))" FAIL

kernel mainline; slot v4l2
: > "$work/running"
place; st=$?
[ "$st" = 1 ] && [ "$(dl)" = "sipeed mpi library" ] && grep -q 'next start' "$work/out" \
    && note "a running server's library is not replaced under it" OK \
    || note "a running server's library is not replaced under it (got $st, '$(dl)')" FAIL
rm -f "$work/running"

echo
echo "===== an application update dry run ====="
# What installPreparedPackage does to a slot, without the server: /kvmapp is
# replaced whole from the release tarball, which carries Sipeed's library and
# an init.d.install that names the vendor-only boot scripts; install.sh runs;
# then the restart runs S95nanokvm, which places the library. The real script
# runs, through its __libkvm verb, the one S98supervise uses.
update_dry_run() {   # <KERNEL>
    kernel "$1"
    slot v4l2
    printf 'S03usbdev\n' > "$work/root/init.d.refuse"
    rm -rf "$work/pkg" "$work/initd" "$work/backup"
    mkdir -p "$work/pkg/server/dl_lib" "$work/pkg/system/init.d" "$work/initd"
    cp "$work/vendor.so" "$work/pkg/server/dl_lib/libkvm.so"
    printf '#!/bin/sh\necho usb\n' > "$work/pkg/system/init.d/S03usbdev"
    cp "$S95" "$work/pkg/system/init.d/S95nanokvm"
    printf 'S03usbdev\nS95nanokvm\n' > "$work/pkg/system/init.d.install"
    # The update: /kvmapp becomes the package.
    rm -rf "$work/root/kvmapp"
    cp -R "$work/pkg" "$work/root/kvmapp"
    INSTALL_SRC=$work/root/kvmapp/system/init.d INSTALL_DEST=$work/initd \
    INSTALL_BACKUP=$work/backup INSTALL_LIST=$work/root/kvmapp/system/init.d.install \
    INSTALL_REFUSE=$work/root/init.d.refuse \
        sh "$INSTALL" > "$work/install.out" 2>&1 || return 1
    DEVINFO=$work/bin/ironkvm-deviceinfo LIBKVM_DIR=$work/root/usr/lib/ironkvm \
    LIBKVM_DST=$work/root/kvmapp/server/dl_lib/libkvm.so CGROUP_PROCS=$work/none \
        sh "$work/initd/S95nanokvm" __libkvm > "$work/out" 2>&1
}
if update_dry_run mainline; then
    [ "$(dl)" = "v4l2 library" ] \
        && note "KERNEL=mainline: after an update the server loads libkvm-v4l2.so" OK \
        || note "KERNEL=mainline: after an update the server loads libkvm-v4l2.so ('$(dl)')" FAIL
    [ ! -e "$work/initd/S03usbdev" ] && [ -f "$work/initd/S95nanokvm" ] \
        && note "and the slot's refused boot script was not put back" OK \
        || note "and the slot's refused boot script was not put back" FAIL
else
    note "KERNEL=mainline: the update dry run ran ($(tail -1 "$work/install.out"))" FAIL
fi
if update_dry_run vendor-prebuilt; then
    [ "$(dl)" = "sipeed mpi library" ] \
        && note "KERNEL=vendor-prebuilt: after an update the server keeps Sipeed's" OK \
        || note "KERNEL=vendor-prebuilt: after an update the server keeps Sipeed's ('$(dl)')" FAIL
else
    note "KERNEL=vendor-prebuilt: the update dry run ran ($(tail -1 "$work/install.out"))" FAIL
fi

echo
echo "===== S98supervise asks for the library before it restarts the server ====="
sed -n '/^# --- libkvm ---/,/^# --- end libkvm ---/p' "$SUPERVISE" > "$work/sv.sh"
grep -q '^refresh_libkvm() {' "$work/sv.sh" \
    && note "S98supervise carries the libkvm block" OK \
    || note "S98supervise carries the libkvm block" FAIL
# The restart that starts the server itself calls it on the line before.
awk '/refresh_libkvm$/ { r = NR } /"\$SERVER_BIN" < \/dev\/null >> "\$SERVER_LOG"/ { if (r && NR == r + 1) ok = 1 } END { exit !ok }' "$SUPERVISE" \
    && note "the restart path calls refresh_libkvm just before starting the server" OK \
    || note "the restart path calls refresh_libkvm just before starting the server" FAIL
cat > "$work/s95stub" <<'STUB'
#!/bin/sh
echo "$1" >> "$S95_CALLS"
echo "S95nanokvm: /kvmapp/server/dl_lib/libkvm.so is now libkvm-v4l2.so"
STUB
chmod 755 "$work/s95stub"
: > "$work/calls"
(
    S95_INIT=$work/s95stub S95_CALLS=$work/calls
    export S95_CALLS
    log() { echo "log: $*" >> "$work/svlog"; }
    warn() { echo "warn: $*" >> "$work/svlog"; }
    . "$work/sv.sh"
    S95_INIT=$work/s95stub
    refresh_libkvm
)
[ "$(cat "$work/calls")" = __libkvm ] \
    && note "refresh_libkvm runs S95nanokvm __libkvm" OK \
    || note "refresh_libkvm runs S95nanokvm __libkvm (got '$(cat "$work/calls")')" FAIL
grep -q '^log: libkvm: .*is now libkvm-v4l2.so' "$work/svlog" 2>/dev/null \
    && note "and logs a library it changed" OK \
    || note "and logs a library it changed" FAIL
( S95_INIT=$work/no-such-script; . "$work/sv.sh"; S95_INIT=$work/no-such-script; refresh_libkvm ); st=$?
[ "$st" = 0 ] \
    && note "a board with no S95nanokvm is not an error" OK \
    || note "a board with no S95nanokvm is not an error" FAIL

echo
if [ "$fails" -eq 0 ]; then echo "all cases passed"; else echo "$fails case(s) FAILED"; fi
[ "$fails" -eq 0 ]
