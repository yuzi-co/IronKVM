#!/bin/sh
# Check that S95nanokvm gives NanoKVM-Server a Go heap limit before it starts.
#
#   test-server-memlimit.sh [path-to-S95nanokvm]
#
# The board has 166MB that Linux can see, and about 96MB of that is available
# once the video carveout, the kernel and the running services are counted.
# NanoKVM-Server peaked at 27.8MB over 26 hours on the reference device, read
# from /proc/<pid>/status VmHWM.
#
# Go grows the heap until the collector decides to run, and the collector's
# only input by default is how much the heap has grown. It does not know what
# the board has left. Below about 30MB free this board stops being able to
# start a process at all: fork fails, the kernel does not choose a victim, and
# nothing recovers it except a power cycle. So the failure this guards against
# is not a killed server, it is a board that answers nothing and cannot be
# reached to fix it.
#
# GOMEMLIMIT gives the collector the missing input. The default here is 64MiB,
# a little over twice the measured peak, which leaves the collector room to
# work before it becomes the thing consuming the board.
#
# tailscaled already gets this treatment at S98tailscaled:46-53. The server,
# the process that actually has to answer, did not.
#
# The limit comes from three places, most specific first: /etc/kvm/GOMEMLIMIT.server,
# then SERVER_GOMEMLIMIT_MIB in the device description, then the compiled-in 64.
# The file stays on top because it lets one board differ from a description that
# every board of its device shares.
#
# The real reader is used behind a stub, not a hand-written fake. A fake
# would answer whatever this suite expects, including for a key that S95nanokvm
# asks for under the wrong name.
set -u

S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-server-memlimit.sh <S95nanokvm>"; exit 1; }

READER=${READER:-$(cd "$(dirname "$0")/../.." && pwd)/kvmapp/system/ironkvm-deviceinfo}
[ -f "$READER" ] || { echo "needs kvmapp/system/ironkvm-deviceinfo"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "===== the server starts under a Go heap limit ====="

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

grep -q '^server_memlimit() {' "$work/f.sh" \
    && note "the block defines server_memlimit" OK \
    || note "the block defines server_memlimit" FAIL

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
    # as a CHILD process sees it, because that is the only thing that matters:
    # the server is started as a child, so a value that is assigned but never
    # exported reaches nothing. Reading it back in the same shell would pass
    # either way.
    #
    # With no $2 the reader names a path that does not exist, so the case reads
    # no description at all and the workstation's own files cannot reach it.
    (
        MEMLIMIT_DIR=$1
        DEVINFO=${2:-$work/no-such-reader}
        export MEMLIMIT_DIR DEVINFO
        . "$work/f.sh"
        server_memlimit
        sh -c 'echo "${GOMEMLIMIT:-unset}"'
    ) 2>/dev/null
}

# The same call, keeping what it said rather than what it set.
run_said() {
    (
        MEMLIMIT_DIR=$1
        DEVINFO=${2:-$work/no-such-reader}
        export MEMLIMIT_DIR DEVINFO
        . "$work/f.sh"
        server_memlimit
    ) 2>&1 >/dev/null
}

# With no override file, the compiled-in default applies.
mkdir -p "$work/empty"
got=$(run "$work/empty")
[ "$got" = "64MiB" ] && note "with no override the limit is 64MiB" OK \
                     || note "with no override the limit is 64MiB (got '$got')" FAIL

# An operator can raise or lower it with a file holding a number of MiB.
mkdir -p "$work/set"
echo 96 > "$work/set/GOMEMLIMIT.server"
got=$(run "$work/set")
[ "$got" = "96MiB" ] && note "an override file sets the limit" OK \
                     || note "an override file sets the limit (got '$got')" FAIL

# A file that does not hold a plain number must not produce a malformed value.
# Go refuses to start when GOMEMLIMIT does not parse, so a typo would take the
# server down rather than merely mis-size it.
for bad in "" "  " "512MiB" "abc" "-5" "12 34"; do
    mkdir -p "$work/bad"
    printf '%s\n' "$bad" > "$work/bad/GOMEMLIMIT.server"
    got=$(run "$work/bad")
    [ "$got" = "64MiB" ] && note "a junk override ('$bad') falls back to the default" OK \
                         || note "a junk override ('$bad') falls back to the default (got '$got')" FAIL
    rm -rf "$work/bad"
done

echo "===== the device description supplies the limit ====="

# A device that allows more heap says so once, in its description, and no edit
# here describes a second device.
mkdir -p "$work/empty"
deviceinfo DEVICE=other RAM_MIB=512 SERVER_GOMEMLIMIT_MIB=96
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "96MiB" ] && note "a description saying 96 gives 96MiB" OK \
                     || note "a description saying 96 gives 96MiB (got '$got')" FAIL

# The per-board file is on top of the description, and this is the case that
# says which way round the two are read. The file exists so that one board can
# differ from a description its whole device shares.
mkdir -p "$work/over"
echo 128 > "$work/over/GOMEMLIMIT.server"
deviceinfo DEVICE=other RAM_MIB=512 SERVER_GOMEMLIMIT_MIB=96
got=$(run "$work/over" "$STUB_READER")
[ "$got" = "128MiB" ] && note "/etc/kvm/GOMEMLIMIT.server beats the description" OK \
                      || note "/etc/kvm/GOMEMLIMIT.server beats the description (got '$got')" FAIL

# This board. The description in devices/sipeed-nanokvm and the copy in
# kvmapp/system/deviceinfo both say 64, which is what was compiled in before
# the key had a reader, so no board changes behaviour.
deviceinfo DEVICE=sipeed-nanokvm RAM_MIB=256 SERVER_GOMEMLIMIT_MIB=64
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "64MiB" ] && note "this board's description keeps 64MiB" OK \
                     || note "this board's description keeps 64MiB (got '$got')" FAIL

# A device that tunes nothing is the normal case, and it is silent.
deviceinfo DEVICE=other RAM_MIB=256
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "64MiB" ] && note "a description with no key leaves today's 64" OK \
                     || note "a description with no key leaves today's 64 (got '$got')" FAIL
said=$(run_said "$work/empty" "$STUB_READER")
[ -z "$said" ] && note "a description with no key says nothing" OK \
               || note "a description with no key said '$said'" FAIL

# Go refuses to start on a GOMEMLIMIT it cannot parse, and a description is read
# by every board of its device, so a typo there would stop all of them. The
# value is ignored, the default stands, and a line names the key.
#
# The value and the line are one case on purpose. The block checks the number
# twice, once where it reads the description and once before it builds the
# string, so a reader that accepted anything would still produce 64MiB here.
# The line is what tells the two apart, and it is read from stderr: the value
# is read through a command substitution, so a report on stdout would be
# swallowed into the number and nobody would see it.
for bad in abc 64MiB -5 0; do
    deviceinfo DEVICE=other RAM_MIB=256 "SERVER_GOMEMLIMIT_MIB=$bad"
    got=$(run "$work/empty" "$STUB_READER")
    said=$(run_said "$work/empty" "$STUB_READER")
    if [ "$got" = "64MiB" ]; then
        case "$said" in
            *SERVER_GOMEMLIMIT_MIB*"$bad"*)
                note "a non-numeric value ('$bad') leaves today's 64 and says so" OK ;;
            *)
                note "a non-numeric value ('$bad') left 64 but said '$said'" FAIL ;;
        esac
    else
        note "a non-numeric value ('$bad') leaves today's 64 (got '$got')" FAIL
    fi
done

# A key the description carries with no value at all. The default stands, and
# there is nothing worth naming in a message.
deviceinfo DEVICE=other RAM_MIB=256 SERVER_GOMEMLIMIT_MIB=
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "64MiB" ] && note "an empty value leaves today's 64" OK \
                     || note "an empty value leaves today's 64 (got '$got')" FAIL

# A board file that holds junk is treated as absent rather than as a reason to
# discard the description. The two are unrelated, and a typo in one board's
# file must not also throw away what the device says.
mkdir -p "$work/junk"
printf '%s\n' abc > "$work/junk/GOMEMLIMIT.server"
deviceinfo DEVICE=other RAM_MIB=512 SERVER_GOMEMLIMIT_MIB=96
got=$(run "$work/junk" "$STUB_READER")
[ "$got" = "96MiB" ] && note "a junk board file falls through to the description" OK \
                     || note "a junk board file falls through to the description (got '$got')" FAIL

# A reader that is not installed is not a fault. A board running Sipeed's
# firmware may carry neither copy, and it has to start the server anyway.
rm -f "$work/deviceinfo"
got=$(run "$work/empty" "$STUB_READER")
[ "$got" = "64MiB" ] && note "no description at all leaves today's 64" OK \
                     || note "no description at all leaves today's 64 (got '$got')" FAIL

got=$(run "$work/empty" "$work/no-such-reader")
[ "$got" = "64MiB" ] && note "no reader at all leaves today's 64" OK \
                     || note "no reader at all leaves today's 64 (got '$got')" FAIL

echo "===== it reaches the server ====="

# The limit has to be in the environment before the server is executed, not
# after. A value exported below the exec line does nothing at all.
limit_line=$(grep -n 'server_memlimit' "$S95" | grep -v '^\s*#' | grep -v 'server_memlimit() {' | head -1 | cut -d: -f1)
exec_line=$(grep -n '"\$SERVER_DST/NanoKVM-Server"' "$S95" | head -1 | cut -d: -f1)
if [ -n "$limit_line" ] && [ -n "$exec_line" ] && [ "$limit_line" -lt "$exec_line" ]; then
    note "the limit is set before the server is started" OK
else
    note "the limit is set before the server is started (call=$limit_line exec=$exec_line)" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
