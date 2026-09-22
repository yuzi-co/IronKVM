#!/bin/sh
# S50avahi-daemon keeps mDNS off when the owner turned it off.
#
#   test-avahi-init.sh [path-to-S50avahi-daemon]
#
# The web UI's mDNS switch deleted /etc/init.d/S50avahi-daemon to turn mDNS
# off. That file lives on the root slot, so the next image put it back and mDNS
# came back on with no word to the owner. The server now also writes
# /etc/kvm/mdns_disabled, which is on /data and survives an image, and the
# script must not start the daemon while that file exists.
#
# The daemon is a stub that records each call, so no case starts anything.
S50=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S50avahi-daemon}
[ -f "$S50" ] || { echo "no S50avahi-daemon at $S50"; exit 2; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
t() { if [ "$2" = 0 ]; then echo "ok   - $1"; else echo "FAIL - $1"; fails=1; fi; }

cat > "$WORK/avahi-daemon" <<STUB
#!/bin/sh
echo "\$*" >> "$WORK/calls"
# -c asks whether a daemon is running. Say no, so start goes on to -D.
[ "\$1" = -c ] && exit 1
exit 0
STUB
chmod +x "$WORK/avahi-daemon"

run() {
    rm -f "$WORK/calls"
    AVAHI_DAEMON="$WORK/avahi-daemon" MDNS_DISABLED_FILE="$WORK/mdns_disabled" \
        sh "$S50" "$@" > "$WORK/out" 2>&1
}

run start
t "start exits 0 with no marker" $?
grep -qx -- "-D" "$WORK/calls" 2>/dev/null
t "start runs the daemon with no marker" $?

touch "$WORK/mdns_disabled"
run start
t "start exits 0 with the marker, so rcS reports no failure" $?
[ ! -e "$WORK/calls" ]
t "start does not touch the daemon with the marker" $?
grep -q "mdns_disabled" "$WORK/out"
t "start says which file kept mDNS off" $?

run stop
grep -qx -- "-k" "$WORK/calls" 2>/dev/null || grep -qx -- "-c" "$WORK/calls" 2>/dev/null
t "stop still reaches the daemon with the marker" $?

exit $fails
