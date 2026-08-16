#!/bin/sh
#
# Install the boot scripts an update package carries.
#
# A package update replaces /kvmapp and nothing else. The fork's boot behaviour
# lives in /etc/init.d: the watchdog, the filesystem mounts, the identity
# carry-over, the USB gadget. Without this script a release ships that behaviour
# in the SD image and omits it from the tarball, and the two drift apart with
# nothing to show it.
#
# The updater runs this after unpacking and before restarting the service.
#
# Every replaced script is kept, with a manifest recording whether it existed at
# all, because a script the update ADDED has to be deleted to undo this, not
# restored. S00awatchdog reads that manifest and undoes the whole set when the
# board fails to come up three times running.
#
# The backup lives in the root filesystem rather than on /boot for two reasons.
# S00awatchdog runs before S01fs mounts anything, and the root filesystem is
# already writable there: the watchdog writes its own log to it on every boot.
# And the root filesystem is per-slot, while /boot is shared, so a backup taken
# on /boot from slot A could be restored into slot B.
#
# Environment, for tests:
#   INSTALL_SRC     scripts to install    (default /kvmapp/system/init.d)
#   INSTALL_DEST    where they go         (default /etc/init.d)
#   INSTALL_BACKUP  where originals go    (default /root/.ironkvm/initd-backup)

SRC=${INSTALL_SRC:-/kvmapp/system/init.d}
DEST=${INSTALL_DEST:-/etc/init.d}
BACKUP=${INSTALL_BACKUP:-/root/.ironkvm/initd-backup}

# A package that carries no init.d directory is not a fault. The hook runs on
# every update, including one that changes nothing outside /kvmapp.
[ -d "$SRC" ] || { echo "install.sh: $SRC is missing, nothing to install"; exit 0; }

# Check every script before installing any of them. A partial install is the
# worst outcome available: some scripts new, some old, and a board that may not
# come up to be repaired. `sh -n` catches the fault that actually happens, which
# is a truncated or mis-edited file.
for f in "$SRC"/*; do
    [ -f "$f" ] || continue
    if ! sh -n "$f" 2>/dev/null; then
        echo "install.sh: ${f##*/} fails a syntax check, installing nothing"
        exit 1
    fi
done

mkdir -p "$BACKUP" "$DEST" || exit 1
: > "$BACKUP/manifest.new"

install_one() {
    name=${1##*/}

    # An identical script is not a change. Copying it anyway would record a
    # rollback to the file it just replaced with a copy of itself.
    if [ -f "$DEST/$name" ] && cmp -s "$1" "$DEST/$name"; then
        return 0
    fi

    if [ -e "$DEST/$name" ]; then
        cp "$DEST/$name" "$BACKUP/$name" || return 1
        echo "$name yes" >> "$BACKUP/manifest.new"
    else
        echo "$name no" >> "$BACKUP/manifest.new"
    fi

    cp "$1" "$DEST/$name" || return 1
    chmod 755 "$DEST/$name" || return 1
    echo "install.sh: installed $name"
}

# The watchdog is what repairs the board, so it goes last. If this run dies
# partway, the copy still in place is the one that already works.
for f in "$SRC"/*; do
    [ -f "$f" ] || continue
    [ "${f##*/}" = S00awatchdog ] && continue
    install_one "$f" || exit 1
done

if [ -f "$SRC/S00awatchdog" ]; then
    install_one "$SRC/S00awatchdog" || exit 1
fi

# Only replace the manifest when this run changed something. The updater may
# retry, and a second run finds every script identical, so an unconditional move
# would leave an empty manifest and no way back from the first run.
if [ -s "$BACKUP/manifest.new" ]; then
    mv "$BACKUP/manifest.new" "$BACKUP/manifest"
else
    rm -f "$BACKUP/manifest.new"
fi

sync
exit 0
