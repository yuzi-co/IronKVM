#!/bin/sh
# Check that test-vision-no-shell.sh fails when the thing it describes is broken.
#
#   test-vision-no-shell-mutation.sh
#
# Not destructive: every mutation is applied to a copy in a temporary directory
# and the shipped source is never written.
#
# A suite that reads a source passes on an empty file as happily as on a correct
# one. Each case below breaks one property on purpose and requires the suite to
# notice. The first mutation is the code this change removed, so that is the
# case that matters most: the suite has to reject the file as it used to be.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SUITE=$ROOT/tools/build/test-vision-no-shell.sh
VIS=$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp

for f in "$SUITE" "$VIS"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

# The mutation has to change the file. A sed that matched nothing would read as
# a caught mutation while proving nothing.
mutate() {
    what=$1; script=$2
    sed "$script" "$VIS" > "$work/v.cpp" || { note "$what" "FAIL (sed failed)"; return 0; }
    if cmp -s "$work/v.cpp" "$VIS"; then
        note "$what" "FAIL (the mutation changed nothing)"
        return 0
    fi
    if sh "$SUITE" "$work/v.cpp" >/dev/null 2>&1; then
        note "$what" "FAIL (survived)"
    else
        note "$what" OK
    fi
}

echo "===== the suite rejects the shells this removed ====="

# One resolution write put back the way it was written before this change.
mutate "a resolution write goes back through a shell" \
    's|^\( *\)write_small_file(vi_width_path, "%d", _width);|\1char Cmd[100]={0};\n\1sprintf(Cmd, "echo %d > %s", _width, vi_width_path);\n\1system(Cmd);|'

# The GPIO writes were literal commands rather than a formatted one, so they
# are a second shape and they need their own case.
mutate "a gpio write goes back through a shell" \
    's|^\( *\)write_small_file("/sys/class/gpio/gpio451/value", "%d", 1);|\1system("echo 1 > /sys/class/gpio/gpio451/value");|'

# sync(2) was reached through a shell for no reason at all.
mutate "sync goes back through a shell" \
    's|^\( *\)sync();|\1system("sync");|'

echo "===== the suite rejects a file that stopped measuring anything ====="

# Deleting the helper has to fail even though no shell comes back with it,
# because a file with neither is a file that no longer writes its state.
mutate "the helper is gone" 's|write_small_file|write_small_file_renamed|g'

# The allowlist is the other half. If the reboot disappears, the rule "no
# system() outside the allowlist" is satisfied by a file that lost a feature.
mutate "the allowed reboot is gone" 's|^\( *\)system("reboot");|\1// removed|'

mutate "the allowed version probe is gone" \
    's|^\( *\)system("/kvmapp/system/init.d/S15kvmhwd get_hdmi_version");|\1// removed|'

echo "===== a comment naming system(3) is not a call ====="

# The helper's own comment explains why system(3) is wrong, so it names it.
# A check that counted that as a call would make the fix unwritable.
cp "$VIS" "$work/v.cpp"
if sh "$SUITE" "$work/v.cpp" >/dev/null 2>&1; then
    note "the shipped source passes" OK
else
    note "the shipped source passes" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
