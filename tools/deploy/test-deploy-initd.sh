#!/bin/sh
# Check that deploy-initd installs an init script to both locations and
# registers it with the boot watchdog.
#
#   test-deploy-initd.sh [path-to-deploy-initd]
#
# Everything runs in temporary directories. The script takes all four of its
# paths from the environment for exactly this reason, so no case here goes near
# a real /etc/init.d.
#
# The load-bearing case is the last group. Registering is not the point:
# surviving restore_initd is the point, so the manifest this script writes is
# fed to the watchdog's own restore_initd, lifted out of S00awatchdog. A
# manifest that looks right and that the watchdog skips would pass every other
# case here.
set -u

DEPLOY=${1:-$(dirname "$0")/deploy-initd}
[ -f "$DEPLOY" ] || { echo "usage: test-deploy-initd.sh <deploy-initd>"; exit 1; }
DEPLOY=$(cd "$(dirname "$DEPLOY")" && pwd)/$(basename "$DEPLOY")

# tools/abslots/device/ is the copy this device runs, byte for byte.
# tools/slots/device/ holds an older one with no restore_initd at all, and
# pointing at it would skip the only group that proves anything.
WATCHDOG=${WATCHDOG:-$(dirname "$0")/../abslots/device/S00awatchdog}

for tool in md5sum awk sed tr; do
    command -v "$tool" >/dev/null 2>&1 || { echo "SKIP: $tool is not on PATH"; exit 2; }
done

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# A fresh device-shaped tree per case, so no case can pass on state another one
# left behind.
n=0
new_tree() {
    n=$((n + 1))
    T="$work/t$n"
    mkdir -p "$T/etc" "$T/pkg" "$T/iron/initd-backup"
    export INITD="$T/etc" PKG_INITD="$T/pkg" IRONDIR="$T/iron" \
           INITD_BACKUP="$T/iron/initd-backup"
}

script() { printf '#!/bin/sh\n# %s\necho %s\n' "$1" "$1"; }

run() { sh "$DEPLOY" "$@" 2>&1; }

echo "===== it refuses a candidate it should not install ====="

new_tree
out=$(run 2>&1); status=$?
[ "$status" -eq 2 ] && note "no candidate exits 2" OK || note "no candidate exits 2 (got $status)" FAIL

out=$(run "$T/missing" 2>&1); status=$?
[ "$status" -eq 2 ] && note "a candidate that does not exist exits 2" OK \
                    || note "a candidate that does not exist exits 2 (got $status)" FAIL

: > "$T/S10empty"
out=$(run "$T/S10empty" 2>&1); status=$?
[ "$status" -eq 2 ] && note "an empty candidate is refused" OK \
                    || note "an empty candidate is refused (got $status)" FAIL

# A carriage return is the failure that shows up at boot as a port that never
# answers, so it must never reach /etc/init.d.
printf '#!/bin/sh\r\necho hi\r\n' > "$T/S10crlf"
out=$(run "$T/S10crlf" 2>&1); status=$?
[ "$status" -eq 2 ] && note "a candidate with carriage returns is refused" OK \
                    || note "a candidate with carriage returns is refused (got $status)" FAIL
[ ! -f "$T/etc/S10crlf" ] && note "the refused candidate was not installed" OK \
                          || note "the refused candidate was not installed" FAIL

printf '#!/bin/sh\nif [ 1 = 1 ]\necho broken\n' > "$T/S10bad"
out=$(run "$T/S10bad" 2>&1); status=$?
[ "$status" -eq 2 ] && note "a candidate that does not parse is refused" OK \
                    || note "a candidate that does not parse is refused (got $status)" FAIL

# restore_initd skips these names, so registering one would record a protection
# that cannot fire.
new_tree
script new > "$T/cand"
out=$(run "$T/cand" "sub/S10x" 2>&1); status=$?
[ "$status" -eq 2 ] && note "a name with a slash is refused" OK \
                    || note "a name with a slash is refused (got $status)" FAIL
out=$(run "$T/cand" ".hidden" 2>&1); status=$?
[ "$status" -eq 2 ] && note "a name starting with a dot is refused" OK \
                    || note "a name starting with a dot is refused (got $status)" FAIL

echo "===== replacing a script that is already there ====="

new_tree
script old > "$T/etc/S10thing";  chmod 700 "$T/etc/S10thing"
script old > "$T/pkg/S10thing";  chmod 755 "$T/pkg/S10thing"
old_sum=$(md5sum "$T/etc/S10thing" | cut -d' ' -f1)
script new > "$T/S10thing.new"
new_sum=$(md5sum "$T/S10thing.new" | cut -d' ' -f1)

out=$(run "$T/S10thing.new"); status=$?
[ "$status" -eq 0 ] && note "a good candidate exits 0" OK || note "a good candidate exits 0 (got $status)" FAIL

# The staged name loses its .new, or the device gets a script rcS never runs.
[ "$(md5sum "$T/etc/S10thing" | cut -d' ' -f1)" = "$new_sum" ] \
    && note "it installs to /etc/init.d under the name without .new" OK \
    || note "it installs to /etc/init.d under the name without .new" FAIL
[ "$(md5sum "$T/pkg/S10thing" | cut -d' ' -f1)" = "$new_sum" ] \
    && note "it installs to the package copy as well" OK \
    || note "it installs to the package copy as well" FAIL
[ ! -f "$T/etc/S10thing.new" ] && note "no file called S10thing.new is left behind" OK \
                               || note "no file called S10thing.new is left behind" FAIL

# The two locations disagree on this device and a deploy is not the place to
# decide they should not.
etc_mode=$(ls -l "$T/etc/S10thing" | cut -c1-10)
pkg_mode=$(ls -l "$T/pkg/S10thing" | cut -c1-10)
case "$etc_mode" in
    -rwx------) note "the /etc copy keeps its 700" OK ;;
    *)          note "the /etc copy keeps its 700 (got $etc_mode)" FAIL ;;
esac
case "$pkg_mode" in
    -rwxr-xr-x) note "the package copy keeps its 755" OK ;;
    *)          note "the package copy keeps its 755 (got $pkg_mode)" FAIL ;;
esac

# The whole point: the version being replaced is what the watchdog can restore.
[ "$(md5sum "$INITD_BACKUP/S10thing" 2>/dev/null | cut -d' ' -f1)" = "$old_sum" ] \
    && note "the replaced script is saved for the watchdog" OK \
    || note "the replaced script is saved for the watchdog" FAIL

grep -q '^S10thing yes$' "$INITD_BACKUP/manifest" \
    && note "the manifest records it as replacing an existing script" OK \
    || note "the manifest records it as replacing an existing script" FAIL

# read -r needs it, or the last entry is never read.
[ "$(tail -c 1 "$INITD_BACKUP/manifest" | od -An -tx1 | tr -d ' \n')" = "0a" ] \
    && note "the manifest ends in a newline" OK \
    || note "the manifest ends in a newline" FAIL

echo "===== adding a script that was not there ====="

new_tree
script new > "$T/S20added"
out=$(run "$T/S20added"); status=$?
[ "$status" -eq 0 ] && note "adding a new script exits 0" OK \
                    || note "adding a new script exits 0 (got $status)" FAIL
grep -q '^S20added no$' "$INITD_BACKUP/manifest" \
    && note "a new script is recorded as 'no', so a rollback removes it" OK \
    || note "a new script is recorded as 'no', so a rollback removes it" FAIL
[ ! -f "$INITD_BACKUP/S20added" ] \
    && note "no backup file is written for a script that did not exist" OK \
    || note "no backup file is written for a script that did not exist" FAIL

echo "===== deploying twice ====="

# The second deploy must not overwrite the backup with the first deploy's
# script. The version worth restoring is the one that was there before any of
# this started, and losing it would leave the watchdog able to roll back only to
# the previous broken attempt.
new_tree
script original > "$T/etc/S30twice"
orig_sum=$(md5sum "$T/etc/S30twice" | cut -d' ' -f1)

script first  > "$T/first";  run "$T/first"  S30twice > /dev/null
script second > "$T/second"; out=$(run "$T/second" S30twice); status=$?

[ "$status" -eq 0 ] && note "a second deploy exits 0" OK || note "a second deploy exits 0 (got $status)" FAIL
[ "$(md5sum "$INITD_BACKUP/S30twice" | cut -d' ' -f1)" = "$orig_sum" ] \
    && note "the backup still holds the original, not the first deploy" OK \
    || note "the backup still holds the original, not the first deploy" FAIL
[ "$(grep -c '^S30twice ' "$INITD_BACKUP/manifest")" = "1" ] \
    && note "the manifest names it once, not twice" OK \
    || note "the manifest names it once, not twice" FAIL

echo "===== an existing manifest is not disturbed ====="

new_tree
printf 'S95nanokvm yes\nS98vidiag no\n' > "$INITD_BACKUP/manifest"
script old > "$T/etc/S40new"
script new > "$T/S40new.cand"
run "$T/S40new.cand" S40new > /dev/null

grep -q '^S95nanokvm yes$' "$INITD_BACKUP/manifest" \
    && note "an earlier entry survives" OK || note "an earlier entry survives" FAIL
grep -q '^S98vidiag no$' "$INITD_BACKUP/manifest" \
    && note "a second earlier entry survives" OK || note "a second earlier entry survives" FAIL
[ "$(grep -c . "$INITD_BACKUP/manifest")" = "3" ] \
    && note "the manifest has three entries" OK \
    || note "the manifest has three entries (got $(grep -c . "$INITD_BACKUP/manifest"))" FAIL

echo "===== the watchdog actually accepts what was written ====="

# Lifted from the shipped watchdog, not copied, so a change to one is a change
# to both. Registering in a shape restore_initd skips is the failure this whole
# script exists to prevent, and only running the real function proves it does not
# happen.
if [ ! -f "$WATCHDOG" ]; then
    # Not a pass. This is the group that proves the manifest is usable, so
    # without it the suite has not tested the thing it exists to test.
    echo
    echo "SKIP: no S00awatchdog at $WATCHDOG, so the rollback was not exercised"
    exit 2
else
    sed -n '/^restore_initd() {/,/^}/p' "$WATCHDOG" > "$work/restore.sh"
    if [ ! -s "$work/restore.sh" ]; then
        note "S00awatchdog defines restore_initd" FAIL
    else
        note "S00awatchdog defines restore_initd" OK

        new_tree
        script original > "$T/etc/S50roll"
        orig_sum=$(md5sum "$T/etc/S50roll" | cut -d' ' -f1)
        script broken > "$T/S50roll.new"
        run "$T/S50roll.new" > /dev/null

        # Sanity: the broken one is in place before the rollback runs.
        [ "$(md5sum "$T/etc/S50roll" | cut -d' ' -f1)" != "$orig_sum" ] \
            && note "the new script is in place before the rollback" OK \
            || note "the new script is in place before the rollback" FAIL

        . "$work/restore.sh"
        if INITD_BACKUP="$INITD_BACKUP" INITD="$T/etc" restore_initd; then
            note "restore_initd accepts the manifest and reports a repair" OK
        else
            note "restore_initd accepts the manifest and reports a repair" FAIL
        fi

        [ "$(md5sum "$T/etc/S50roll" | cut -d' ' -f1)" = "$orig_sum" ] \
            && note "the original script is back in /etc/init.d" OK \
            || note "the original script is back in /etc/init.d" FAIL

        [ -f "$INITD_BACKUP/manifest.done" ] \
            && note "the manifest is spent, so it cannot undo the next update" OK \
            || note "the manifest is spent, so it cannot undo the next update" FAIL
    fi
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
