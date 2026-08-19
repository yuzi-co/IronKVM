#!/bin/sh
# Prove that test-usb-link-record.sh fails when the boot record is broken.
#
#   test-usb-link-record-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped script is only ever read.
#
# This record is worth breaking on purpose because of what it is for. It exists
# to survive a failure nobody is present for, and a log that quietly records the
# wrong thing is worse than no log: the next investigation reads it, believes
# it, and rules out the true cause. A case that only asserts "a file appeared"
# passes against every mutation below.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the cases pass because there is nothing wrong with it, and the
# run reads as proof when it proved nothing.
DIR=$(dirname "$0")
S03=${1:-$DIR/../../kvmapp/system/init.d/S03usbdev}
TEST=$DIR/test-usb-link-record.sh
for f in "$S03" "$TEST"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 1; }
done

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    mkdir -p "$d/init.d"
    cp "$S03" "$d/init.d/S03usbdev"
    "$@" "$d/init.d/S03usbdev"

    if cmp -s "$d/init.d/S03usbdev" "$S03"; then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    if timeout 60 sh "$TEST" "$d/init.d" > /dev/null 2>&1; then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# The trim goes, so a board that reboots in a loop fills the partition it
# shares with the disk images.
m_notrim() {
    sed -i 's@^        tail -n .*@        :@' "$1"
}

# The ring is copied on every boot, so the good case costs the whole buffer
# rather than one line.
m_alwaysring() {
    sed -i 's@^        if \[ "$speed" != high-speed \].*@        if true@' "$1"
}

# The ring is never copied, so the one boot that needed the evidence is the one
# that recorded none of it.
m_neverring() {
    sed -i 's@^        if \[ "$speed" != high-speed \].*@        if false@' "$1"
}

# The speed itself stops being recorded. The line still appears, still carries a
# timestamp and a state, and answers none of the question it exists for.
m_nospeed() {
    sed -i 's@ speed=$speed@@' "$1"
}

# Only a boot that worked is recorded, so the failure the log exists for leaves
# nothing behind.
m_onlysuccess() {
    sed -i 's@^    usb_record_link "$rebinds"$@    [ "$watched" -eq 0 ] \&\& usb_record_link "$rebinds"@' "$1"
}

# Recording decides the watch's return code, so a gadget the host never took is
# reported as a success.
m_recordrc() {
    sed -i 's@^    return "$watched"$@    return $?@' "$1"
}

# The log moves onto the boot SD card, which is the one place runtime state must
# never go.
m_bootcard() {
    sed -i 's@USB_LINK_LOG:-/data/usb-link.log@USB_LINK_LOG:-/kvmapp/usb-link.log@' "$1"
}

# The whole ring is copied instead of the dwc2 lines, so the evidence arrives
# buried in a video message printed once a second.
m_wholering() {
    sed -i 's@dmesg | grep dwc2 | sed@dmesg | sed@' "$1"
}

echo "===== every mutation must be caught ====="
try "the trim is removed"                       m_notrim
try "the ring is copied on every boot"          m_alwaysring
try "the ring is never copied"                  m_neverring
try "the speed is left out of the record"       m_nospeed
try "only a boot that worked is recorded"       m_onlysuccess
try "recording decides the watch's exit code"   m_recordrc
try "the log moves onto the boot card"          m_bootcard
try "the whole ring is copied, not just dwc2"   m_wholering

echo
if [ "$fail" -eq 0 ]; then
    echo "===== all $pass mutations caught ====="
else
    echo "===== $fail of $((pass + fail)) mutations were not caught ====="
    exit 1
fi
