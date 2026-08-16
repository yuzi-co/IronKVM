#!/bin/sh
# Tests for kvmapp/system/install.sh.
#
# Everything is driven against a scratch tree through the script's environment
# variables, so no test touches a real /etc/init.d.

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/../../kvmapp/system/install.sh"
pass=0
fail=0

skipped=0

ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
skip()  { skipped=$((skipped + 1)); echo "  skip  $1 ($2)"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

# MSYS on Windows has no executable bit. It derives one from the file's shebang,
# so a mode check there passes whatever the script under test does, including
# nothing. Detect that and skip loudly rather than report a pass that proves
# nothing: a guard nobody can fail is a guard nobody is testing. Run this script
# in a Linux container to exercise those cases.
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
    mkdir -p "$WORK/src" "$WORK/dest" "$WORK/backup"
}

teardown() { rm -rf "$WORK"; }

run() {
    INSTALL_SRC="$WORK/src" INSTALL_DEST="$WORK/dest" INSTALL_BACKUP="$WORK/backup" \
        sh "$SCRIPT" > "$WORK/out" 2>&1
    echo $?
}

echo "install.sh"

# A new script has to arrive, executable, or the boot behaviour the package
# carries never runs.
setup
printf '#!/bin/sh\necho new\n' > "$WORK/src/S40new"
status=$(run)
check "installs a new script" "$status" "0"
check "the new script is present" "$([ -f "$WORK/dest/S40new" ] && echo yes || echo no)" "yes"
if [ "$MODES" = real ]; then
    check "the new script is executable" "$(stat -c %a "$WORK/dest/S40new")" "755"
else
    skip "the new script is executable" "this platform has no executable bit"
fi
check "a new script is recorded as new" "$(cat "$WORK/backup/manifest")" "S40new no"
teardown

# The original must be kept, or there is nothing for the rollback to restore.
setup
printf '#!/bin/sh\necho old\n' > "$WORK/dest/S40thing"
printf '#!/bin/sh\necho new\n' > "$WORK/src/S40thing"
run > /dev/null
check "the replaced original is kept" \
    "$(cat "$WORK/backup/S40thing")" "$(printf '#!/bin/sh\necho old\n')"
check "a replaced script is recorded as existing" "$(cat "$WORK/backup/manifest")" "S40thing yes"
teardown

# A script that fails a syntax check would brick the next boot. Nothing may be
# installed, not even the scripts that are fine.
setup
printf '#!/bin/sh\necho fine\n' > "$WORK/src/S40fine"
printf '#!/bin/sh\nif [ broken\n'  > "$WORK/src/S41broken"
status=$(run)
check "a syntax error fails the run" "$status" "1"
check "nothing is installed when one script is broken" \
    "$(ls "$WORK/dest" | wc -l | tr -d ' ')" "0"
teardown

# The watchdog performs the repair, so it is written last. If the run dies
# partway, the copy still in place is the one that already works.
setup
printf '#!/bin/sh\necho w\n' > "$WORK/src/S00awatchdog"
printf '#!/bin/sh\necho f\n' > "$WORK/src/S40fs"
run > /dev/null
check "the watchdog is installed last" \
    "$(tail -1 "$WORK/backup/manifest" | cut -d' ' -f1)" "S00awatchdog"
teardown

# Running twice must not erase the rollback information the first run recorded.
# The updater may retry, and a second run that finds nothing to do would
# otherwise leave an empty manifest and no way back.
setup
printf '#!/bin/sh\necho old\n' > "$WORK/dest/S40thing"
printf '#!/bin/sh\necho new\n' > "$WORK/src/S40thing"
run > /dev/null
run > /dev/null
check "a second run keeps the manifest" "$(cat "$WORK/backup/manifest")" "S40thing yes"
check "a second run keeps the original" \
    "$(cat "$WORK/backup/S40thing")" "$(printf '#!/bin/sh\necho old\n')"
teardown

# An unchanged script must not be copied, or every update would record a
# rollback to the file it just replaced with an identical one.
setup
printf '#!/bin/sh\necho same\n' > "$WORK/dest/S40same"
printf '#!/bin/sh\necho same\n' > "$WORK/src/S40same"
run > /dev/null
check "an identical script is skipped" \
    "$([ -s "$WORK/backup/manifest" ] && echo recorded || echo skipped)" "skipped"
teardown

# A package that carries no init.d directory is not a fault. The hook runs on
# every update, including one that changes nothing outside /kvmapp.
setup
rm -rf "$WORK/src"
status=$(run)
check "a missing source directory is not a failure" "$status" "0"
teardown

echo
echo "passed $pass, failed $fail, skipped $skipped"
[ "$skipped" -gt 0 ] && echo "run this in a Linux container to cover the skipped cases"
[ "$fail" -eq 0 ]
