#!/bin/sh
# Tests for the init.d rollback in S00awatchdog.
#
# The block under test is extracted from the script by its markers and sourced,
# so these drive the real code rather than a copy of it that could drift.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
WATCHDOG="$HERE/S00awatchdog"
pass=0
fail=0

skipped=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
skip()  { skipped=$((skipped + 1)); echo "  skip  $1 ($2)"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

# MSYS on Windows has no executable bit and derives one from the shebang, so a
# mode check there passes whatever the code does. Skip loudly rather than report
# a pass that proves nothing. Run this in a Linux container to cover it.
modes_are_real() {
    probe=$(mktemp)
    printf '#!/bin/sh\n' > "$probe"
    chmod 644 "$probe"
    if [ -x "$probe" ]; then rm -f "$probe"; return 1; fi
    rm -f "$probe"
    return 0
}

if modes_are_real; then MODES=real; else MODES=fake; fi

setup() {
    WORK=$(mktemp -d)
    mkdir -p "$WORK/state/initd-backup" "$WORK/initd"
    sed -n '/^# --- initd rollback ---$/,/^# --- end initd rollback ---$/p' \
        "$WATCHDOG" > "$WORK/block.sh"
    if [ ! -s "$WORK/block.sh" ]; then
        bad "the initd rollback block is missing from S00awatchdog"
        return 1
    fi
    IRONDIR="$WORK/state"
    ATTEMPTS="$WORK/state/boot-attempts"
    INITD_BACKUP="$WORK/state/initd-backup"
    INITD="$WORK/initd"
    log() { :; }
    . "$WORK/block.sh"
}

teardown() { rm -rf "$WORK"; }

echo "S00awatchdog init.d rollback"

# The counter has to survive a boot that never finishes, which is the only kind
# this mechanism ever sees.
setup || exit 1
check "the first boot counts one"  "$(bump_attempts)" "1"
check "the second boot counts two" "$(bump_attempts)" "2"
teardown

# A board that comes up must forget the attempt, or three unrelated slow boots
# would undo a good update on a board that is fine.
setup || exit 1
bump_attempts > /dev/null
clear_attempts
check "a reachable boot clears the counter" \
    "$([ -f "$ATTEMPTS" ] && echo present || echo gone)" "gone"
teardown

# Garbage in the counter must not stop the mechanism. The file is written during
# a boot that a power cut can end at any moment.
setup || exit 1
echo "not a number" > "$ATTEMPTS"
check "a corrupt counter restarts at one" "$(bump_attempts)" "1"
teardown

# Restoring a replaced script is the whole point.
setup || exit 1
printf 'old\n' > "$INITD_BACKUP/S40thing"
printf 'new\n' > "$INITD/S40thing"
echo "S40thing yes" > "$INITD_BACKUP/manifest"
chmod 644 "$INITD_BACKUP/S40thing"
restore_initd
check "a replaced script is restored" "$(cat "$INITD/S40thing")" "old"
# A restored script that is not executable does not run, and the script most
# likely to be restored is the one that mounts the filesystems.
if [ "$MODES" = real ]; then
    check "the restored script is executable" "$(stat -c %a "$INITD/S40thing")" "755"
else
    skip "the restored script is executable" "this platform has no executable bit"
fi
teardown

# A script the update ADDED has no original. Undoing it means deleting it, and
# restoring nothing would leave the fault in place on every later boot.
setup || exit 1
printf 'new\n' > "$INITD/S41added"
echo "S41added no" > "$INITD_BACKUP/manifest"
restore_initd
check "an added script is deleted" \
    "$([ -e "$INITD/S41added" ] && echo present || echo gone)" "gone"
teardown

# Restoring twice would undo a later, good update, and the board could never
# move forward. Once the manifest is spent it must not be read again.
setup || exit 1
printf 'old\n' > "$INITD_BACKUP/S40thing"
printf 'new\n' > "$INITD/S40thing"
echo "S40thing yes" > "$INITD_BACKUP/manifest"
restore_initd
printf 'newer\n' > "$INITD/S40thing"
if restore_initd; then bad "a spent manifest must not restore again"
else ok "a spent manifest must not restore again"; fi
check "the later script survives" "$(cat "$INITD/S40thing")" "newer"
teardown

# With no manifest there is nothing to undo. Saying so is what stops the caller
# reporting a repair that did not happen, which is the fault this whole design
# was written after.
setup || exit 1
if restore_initd 2> "$WORK/err"; then bad "no manifest means no restore"
else ok "no manifest means no restore"; fi
# Every normal boot takes this path. The early return is what keeps it quiet:
# without it the read loop opens a file that is not there and complains on the
# console at every single boot. Silence is the only observable difference
# between that guard and the count check below it, so it is what proves it runs.
check "a boot with no manifest is silent" "$(wc -c < "$WORK/err" | tr -d ' ')" "0"
teardown

# A manifest naming a script that is no longer backed up must not count as a
# restore either.
setup || exit 1
echo "S40missing yes" > "$INITD_BACKUP/manifest"
if restore_initd 2> "$WORK/err"; then bad "a manifest with no backup file restores nothing"
else ok "a manifest with no backup file restores nothing"; fi
check "a missing backup file is not reported to the console" \
    "$(wc -c < "$WORK/err" | tr -d ' ')" "0"
teardown

# The manifest is written by this system, but it lives on a filesystem a power
# cut can scramble, and this runs as root before anything else in the boot.
#
# The dangerous branch is the one that DELETES. A name recorded as new is
# removed outright, so a name carrying a path would delete a file outside
# /etc/init.d, running as root before anything else in the boot.
setup || exit 1
mkdir -p "$WORK/outside"
printf 'keep\n' > "$WORK/outside/target"
echo "../outside/target no" > "$INITD_BACKUP/manifest"
restore_initd
check "a path in the manifest cannot delete outside the directory" \
    "$([ -f "$WORK/outside/target" ] && echo kept || echo DELETED)" "kept"

# And the copying branch, for the same reason in the other direction.
printf 'evil\n' > "$INITD_BACKUP/escape"
printf 'keep\n' > "$WORK/outside/write-target"
echo "../outside/write-target yes" > "$INITD_BACKUP/manifest"
restore_initd
check "a path in the manifest cannot overwrite outside the directory" \
    "$(cat "$WORK/outside/write-target")" "keep"
teardown

# The restore has to happen in start(), before the rc sequence reaches S01. The
# watcher below it waits up to five minutes, by which time every other script
# has already run or hung.
echo "S00awatchdog wiring"

# The escalate path has to try the undo BEFORE the recovery marker. Without
# that, the counter in start() can never reach its limit on a board that has
# this watchdog: either the board is reachable and the count clears, or it is
# not and recovery takes it on the first boot. The rollback would be dead code.
check "escalate tries the rollback before recovery" \
    "$(sed -n '/^escalate()/,/^}/p' "$WATCHDOG" | grep 'restore_initd\|: > "\$BOOT/recovery"' \
       | head -1 | grep -c restore_initd)" "1"
check "escalate still falls back to recovery" \
    "$(sed -n '/^escalate()/,/^}/p' "$WATCHDOG" | grep -c ': > "\$BOOT/recovery"')" "1"
check "the rollback reboot is guarded by the dry run" \
    "$(sed -n '/^escalate()/,/^}/p' "$WATCHDOG" | grep -c 'would have restored')" "1"

# install.sh can now put this watchdog on a board with stock partitioning, where
# there is no recovery slot and the marker is an inert file. Rebooting there
# every DEADLINE seconds for ever is worse than doing nothing.
check "escalate checks for a recovery slot before using one" \
    "$(sed -n '/^escalate()/,/^}/p' "$WATCHDOG" | grep 'command -v slot\|: > "\$BOOT/recovery"' \
       | head -1 | grep -c 'command -v slot')" "1"
check "start bumps the counter before the subshell" \
    "$(sed -n '/^start()/,/^    (/p' "$WATCHDOG" | grep -c 'bump_attempts')" "1"
check "start restores when the limit is reached" \
    "$(sed -n '/^start()/,/^    (/p' "$WATCHDOG" | grep -c 'restore_initd')" "1"
# The loop lives in watch() now, because the defect it was refactored for was in
# what the loop did with the answer rather than in the answer. This stays a
# shape check, and the behaviour it stands for is executed in test-watchdog.sh:
# "a reachable boot clears the update counter", beside the case that proves an
# unreachable one does not.
check "the success path clears the counter, and only it" \
    "$(sed -n '/^watch()/,/^}/p' "$WATCHDOG" | grep -c 'clear_attempts')" "1"

echo
echo "passed $pass, failed $fail, skipped $skipped"
[ "$skipped" -gt 0 ] && echo "run this in a Linux container to cover the skipped cases"
[ "$fail" -eq 0 ]
