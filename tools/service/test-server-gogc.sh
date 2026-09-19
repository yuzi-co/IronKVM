#!/bin/sh
# Check that S95nanokvm gives NanoKVM-Server a collector target before it
# starts.
#
#   test-server-gogc.sh [path-to-S95nanokvm]
#
# GOMEMLIMIT, covered by test-server-memlimit.sh, tells the collector where the
# ceiling is. GOGC tells it how hard to work below that ceiling, and left unset
# it is 100: a collection every time the heap grows by the size of the live
# heap.
#
# On the MJPEG path the capture loop hands over a freshly allocated slice per
# frame, about 150KiB thirty times a second. That allocation cannot be pooled,
# because the slice outlives the loop iteration in each client's FrameSlot and
# in the screenshot cache. So a small live heap is passed several times a
# second and the collector runs about twice a second.
#
# Measured on the reference device on 2026-09-04, with that frame size and rate
# and two client slots holding each frame, as a percentage of the one core:
#
#   live heap   GOGC=100   GOGC=150   GOGC=200   GOGC=400
#   2MiB          4.3%       3.2%       2.6%       1.7%
#   4MiB          5.3%       3.8%       3.1%         -
#   8MiB          5.4%       3.5%       2.8%       2.3%
#
# The peak heap follows the ratio, at about (1 + GOGC/100) times the live heap,
# a model that matched the measurement at every step.
#
# The server peaks at 27.8MB of RSS over 26 hours, so its live heap is at the
# small end of that table. 150 buys back about 1.1 to 1.9 points of the one
# core for half the live heap again. 200 and 400 were measured and neither was
# taken: 200 is another 0.7 points for another 2 to 3MiB, and 400 is 10 to
# 24MiB beyond that, on a board that stops being able to start a process below
# about 30MB free.
#
# The H.264 paths are unaffected. Their frames are a few KiB, and the same
# sweep records no collection at all above GOGC=100.
#
# The target comes from three places, most specific first: /etc/kvm/GOGC.server,
# then SERVER_GOGC in the device description, then the compiled-in 150. The file
# stays on top because it lets one board differ from a description that every
# board of its device shares.
#
# The real reader is used behind a stub, not a hand-written fake. A fake
# would answer whatever this suite expects, including for a key that S95nanokvm
# asks for under the wrong name.
set -u

S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-gogc.sh <S95nanokvm>"; exit 1; }

READER=${READER:-$(cd "$(dirname "$0")/../.." && pwd)/kvmapp/system/ironkvm-deviceinfo}
[ -f "$READER" ] || { echo "needs kvmapp/system/ironkvm-deviceinfo"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "===== the server starts under a collector target ====="

# Run the shipped text, not a copy of it. The block is taken whole, because the
# function now reads the device description through the reader the same block
# resolves.
sed -n '/^# --- Go tuning ---/,/^# --- end Go tuning ---/p' "$S95" > "$work/f.sh"
if [ ! -s "$work/f.sh" ]; then
    note "S95nanokvm carries the Go tuning block" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi
note "S95nanokvm carries the Go tuning block" OK

grep -q '^server_gogc() {' "$work/f.sh" \
    && note "the block defines server_gogc" OK \
    || note "the block defines server_gogc" FAIL

# The reader is resolved the way S01zram and S01fs resolve it. The fallback is
# load-bearing: nothing puts the reader on PATH on a board running Sipeed's
# firmware, so a block that only asked PATH would read no description there.
grep -q '^DEVINFO_TARBALL=${DEVINFO_TARBALL:-/kvmapp/system/ironkvm-deviceinfo}$' "$work/f.sh" \
    && note "the block falls back to the reader the tarball carries" OK \
    || note "the block falls back to the reader the tarball carries" FAIL

# A stub in front of the real reader, pointed at this suite's fixture. The cases
# name it through DEVINFO rather than putting it on PATH: how the reader is
# found is settled in the zram and S01fs suites, and what is under test here is
# what the block does with the answer.
mkdir -p "$work/bin"
cat > "$work/bin/ironkvm-deviceinfo" <<STUB
#!/bin/sh
DEVICEINFO_PATHS="$work/deviceinfo" exec sh "$READER" "\$@"
STUB
chmod 755 "$work/bin/ironkvm-deviceinfo"
STUB_READER=$work/bin/ironkvm-deviceinfo

# Write a description for the next case. Each argument is one KEY=value line.
deviceinfo() { printf '%s\n' "$@" > "$work/deviceinfo"; }

run() {
    # $1 is the config directory to read from, $2 the reader. Prints the value
    # as a CHILD process sees it. The server is started as a child, so a value
    # assigned but never exported reaches nothing, and reading it back in the
    # same shell would pass either way.
    #
    # With no $2 the reader names a path that does not exist, so the case reads
    # no description at all and the workstation's own files cannot reach it.
    (
        MEMLIMIT_DIR=$1
        DEVINFO=${2:-$work/no-such-reader}
        export MEMLIMIT_DIR DEVINFO
        . "$work/f.sh"
        server_gogc
        sh -c 'echo "${GOGC:-unset}"'
    ) 2>/dev/null
}

# The same call, keeping what it said rather than what it set.
run_said() {
    (
        MEMLIMIT_DIR=$1
        DEVINFO=${2:-$work/no-such-reader}
        export MEMLIMIT_DIR DEVINFO
        . "$work/f.sh"
        server_gogc
    ) 2>&1 >/dev/null
}

mkdir -p "$work/empty"
got=$(run "$work/empty")
[ "$got" = "150" ] && note "with no override the target is 150" OK \
                   || note "with no override the target is 150 (got '$got')" FAIL

mkdir -p "$work/set"
echo 300 > "$work/set/GOGC.server"
got=$(run "$work/set")
[ "$got" = "300" ] && note "an override file sets the target" OK \
                   || note "an override file sets the target (got '$got')" FAIL

# Go refuses to start on a GOGC it cannot parse, so a typo in that file would
# take the server down rather than merely mis-tune it.
for bad in "" "  " "150%" "abc" "-5" "12 34" "1e3"; do
    mkdir -p "$work/bad"
    printf '%s\n' "$bad" > "$work/bad/GOGC.server"
    got=$(run "$work/bad")
    [ "$got" = "150" ] && note "a junk override ('$bad') falls back to the default" OK \
                       || note "a junk override ('$bad') falls back to the default (got '$got')" FAIL
    rm -rf "$work/bad"
done

# "off" parses in Go and is the one value that must not get through. It removes
# the ratio entirely and leaves the memory limit as the only trigger, which is
# the configuration that collects hardest exactly when the board can least
# afford it.
mkdir -p "$work/off"
echo off > "$work/off/GOGC.server"
got=$(run "$work/off")
[ "$got" = "150" ] && note "'off' is refused and falls back to the default" OK \
                   || note "'off' is refused and falls back to the default (got '$got')" FAIL

echo "===== the device description supplies the target ====="

# A device whose collector should work differently says so once, in its
# description, and no edit here describes a second device.
mkdir -p "$work/empty"
deviceinfo DEVICE=other RAM_MIB=512 SERVER_GOGC=96
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "96" ] && note "a description saying 96 gives 96" OK \
                  || note "a description saying 96 gives 96 (got '$got')" FAIL

# The per-board file is on top of the description, and this is the case that
# says which way round the two are read. The file exists so that one board can
# differ from a description its whole device shares.
mkdir -p "$work/over"
echo 300 > "$work/over/GOGC.server"
deviceinfo DEVICE=other RAM_MIB=512 SERVER_GOGC=96
got=$(run "$work/over" "$STUB_READER")
[ "$got" = "300" ] && note "/etc/kvm/GOGC.server beats the description" OK \
                   || note "/etc/kvm/GOGC.server beats the description (got '$got')" FAIL

# This board. The description in devices/sipeed-nanokvm and the copy in
# kvmapp/system/deviceinfo both say 150, which is what was compiled in before
# the key had a reader, so no board changes behaviour.
deviceinfo DEVICE=sipeed-nanokvm RAM_MIB=256 SERVER_GOGC=150
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "150" ] && note "this board's description keeps 150" OK \
                   || note "this board's description keeps 150 (got '$got')" FAIL

# A device that tunes nothing is the normal case, and it is silent.
deviceinfo DEVICE=other RAM_MIB=256
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "150" ] && note "a description with no key leaves today's 150" OK \
                   || note "a description with no key leaves today's 150 (got '$got')" FAIL
said=$(run_said "$work/empty" "$STUB_READER")
[ -z "$said" ] && note "a description with no key says nothing" OK \
               || note "a description with no key said '$said'" FAIL

# Go refuses to start on a GOGC it cannot parse, and a description is read by
# every board of its device, so a typo there would stop all of them. The value
# is ignored, the default stands, and a line names the key.
#
# "off" is in this list and is the one entry Go would accept. It removes the
# ratio entirely and leaves the memory limit as the only trigger, which is the
# configuration that collects hardest exactly when the board can least afford
# it. A description must not be able to ask for it, any more than a board file
# can.
#
# The value and the line are one case on purpose. The block checks the number
# twice, once where it reads the description and once before it exports it, so
# a reader that accepted anything would still produce 150 here. The line is
# what tells the two apart, and it is read from stderr: the value is read
# through a command substitution, so a report on stdout would be swallowed into
# the number and nobody would see it.
for bad in abc 150% -5 0 off; do
    deviceinfo DEVICE=other RAM_MIB=256 "SERVER_GOGC=$bad"
    got=$(run "$work/empty" "$STUB_READER")
    said=$(run_said "$work/empty" "$STUB_READER")
    if [ "$got" = "150" ]; then
        case "$said" in
            *SERVER_GOGC*"$bad"*)
                note "a non-numeric value ('$bad') leaves today's 150 and says so" OK ;;
            *)
                note "a non-numeric value ('$bad') left 150 but said '$said'" FAIL ;;
        esac
    else
        note "a non-numeric value ('$bad') leaves today's 150 (got '$got')" FAIL
    fi
done

# A key the description carries with no value at all. The default stands, and
# there is nothing worth naming in a message.
deviceinfo DEVICE=other RAM_MIB=256 SERVER_GOGC=
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "150" ] && note "an empty value leaves today's 150" OK \
                   || note "an empty value leaves today's 150 (got '$got')" FAIL

# "off" in the board file is refused, and refusing it must not also throw away
# what the device says. The two are unrelated.
deviceinfo DEVICE=other RAM_MIB=512 SERVER_GOGC=200
got=$(run "$work/off" "$STUB_READER")
[ "$got" = "200" ] && note "'off' in the board file falls through to the description" OK \
                   || note "'off' in the board file falls through to the description (got '$got')" FAIL

# A reader that is not installed is not a fault. A board running Sipeed's
# firmware may carry neither copy, and it has to start the server anyway.
rm -f "$work/deviceinfo"
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "150" ] && note "no description at all leaves today's 150" OK \
                   || note "no description at all leaves today's 150 (got '$got')" FAIL

got=$(run "$work/empty" "$work/no-such-reader")
[ "$got" = "150" ] && note "no reader at all leaves today's 150" OK \
                   || note "no reader at all leaves today's 150 (got '$got')" FAIL

echo "===== it reaches the server ====="

# Set before the exec line, or it does nothing at all.
gogc_line=$(grep -n 'server_gogc' "$S95" | grep -v '^\s*#' | grep -v 'server_gogc() {' | head -1 | cut -d: -f1)
exec_line=$(grep -n '"\$SERVER_DST/NanoKVM-Server"' "$S95" | head -1 | cut -d: -f1)
if [ -n "$gogc_line" ] && [ -n "$exec_line" ] && [ "$gogc_line" -lt "$exec_line" ]; then
    note "the target is set before the server is started" OK
else
    note "the target is set before the server is started (call=$gogc_line exec=$exec_line)" FAIL
fi

echo "===== the ceiling is still there ====="

# This is the load-bearing pair. GOGC is a ratio with no upper bound of its
# own: a live heap that grows takes the trigger with it. GOMEMLIMIT is what
# stops that, and raising GOGC without it would turn a memory fault this board
# cannot recover from into a likelier one.
grep -q '^server_memlimit() {' "$S95" \
    && note "server_memlimit is still defined" OK \
    || note "server_memlimit is still defined" FAIL

memlimit_line=$(grep -n 'server_memlimit' "$S95" | grep -v '^\s*#' | grep -v 'server_memlimit() {' | head -1 | cut -d: -f1)
if [ -n "$memlimit_line" ] && [ -n "$exec_line" ] && [ "$memlimit_line" -lt "$exec_line" ]; then
    note "the limit is still set before the server is started" OK
else
    note "the limit is still set before the server is started" FAIL
fi

# Both in the same branch, so neither can be moved out from under the other by
# an edit that only looks at one of them.
if [ -n "$gogc_line" ] && [ -n "$memlimit_line" ]; then
    gap=$((gogc_line - memlimit_line))
    [ "$gap" -gt -6 ] && [ "$gap" -lt 6 ] \
        && note "the two are set together" OK \
        || note "the two are $gap lines apart, wanted them together" FAIL
else
    note "the two are set together" FAIL
fi

# The two override files are separate names on purpose, so an operator raising
# one cannot silently disable the other.
grep -q 'GOGC.server' "$S95" \
    && note "the override file is GOGC.server" OK \
    || note "the override file is GOGC.server" FAIL

grep -q 'GOMEMLIMIT.server' "$S95" \
    && note "the limit override file is still GOMEMLIMIT.server" OK \
    || note "the limit override file is still GOMEMLIMIT.server" FAIL

echo "===== what the operator is told ====="

# The startup line names both, because a board behaving oddly is diagnosed from
# that log and a value that is set but never printed is a value nobody checks.
grep -q 'GOMEMLIMIT set to .*GOGC set to' "$S95" \
    && note "the startup line reports both values" OK \
    || note "the startup line reports both values" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
