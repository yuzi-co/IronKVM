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
#   INSTALL_LIST    which ones to install (default /kvmapp/system/init.d.install)

SRC=${INSTALL_SRC:-/kvmapp/system/init.d}
DEST=${INSTALL_DEST:-/etc/init.d}
BACKUP=${INSTALL_BACKUP:-/root/.ironkvm/initd-backup}
LIST=${INSTALL_LIST:-$(dirname "$SRC")/init.d.install}

# A package that carries no init.d directory is not a fault. The hook runs on
# every update, including one that changes nothing outside /kvmapp.
[ -d "$SRC" ] || { echo "install.sh: $SRC is missing, nothing to install"; exit 0; }

# The list names the scripts this package may install, and release.sh derives it
# from the image manifest so the two can never install different sets.
#
# Installing the whole directory is wrong and was the first thing this script
# did. /kvmapp/system/init.d is the application's own reference copy: it carries
# 20 scripts and the image installs 10. The image leaves S50sshd, S00kmod,
# S15kvmhwd and S80dnsmasq at the versions the base system shipped, and it never
# installs avahi, ssdpd, tailscaled, picoclaw, wifi or usbhid at all. Installing
# the directory would start six daemons at the next boot that the same release's
# SD image never starts, on a board with 166MB of RAM.
#
# A package with no list was not built by release.sh, and the safe reading is
# that it changes no boot script. Guessing is what caused the fault above.
[ -f "$LIST" ] || { echo "install.sh: $LIST is missing, no boot scripts to install"; exit 0; }
NAMES=$(grep -v '^[[:space:]]*$' "$LIST")

# Check every script before installing any of them. A partial install is the
# worst outcome available: some scripts new, some old, and a board that may not
# come up to be repaired. `sh -n` catches the fault that actually happens, which
# is a truncated or mis-edited file.
#
# Only the listed ones are checked. A script nothing installs cannot break a
# boot, so it must not be able to stop the scripts that will.
for n in $NAMES; do
    if [ ! -f "$SRC/$n" ]; then
        echo "install.sh: $LIST names $n and the package does not carry it, installing nothing"
        exit 1
    fi
    if ! sh -n "$SRC/$n" 2>/dev/null; then
        echo "install.sh: $n fails a syntax check, installing nothing"
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
for n in $NAMES; do
    [ "$n" = S00awatchdog ] && continue
    install_one "$SRC/$n" || exit 1
done

for n in $NAMES; do
    if [ "$n" = S00awatchdog ]; then
        install_one "$SRC/S00awatchdog" || exit 1
    fi
done

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
