#!/bin/sh
# Check that check-init-install.sh notices a boot script with no decided fate.
#
#   test-check-init-install.sh [check-init-install.sh]
#
# Not destructive: every case builds a package directory, a base directory, a
# manifest and a declaration list in a temporary tree. Nothing reads the real
# base rootfs, so this suite runs on a workstation.
#
# The fault this covers: the fork rewrote S00kmod on 2026-09-04 so the modules
# in the install package are loaded before the stock ones. The image manifest
# did not install S00kmod, because until that commit the fork did not change
# it, so every image built afterwards kept the stock loader and the fix reached
# no device. Nothing failed and nothing said anything. Case 3 below is that
# exact shape.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
CHECK=${1:-$ROOT/tools/abslots/check-init-install.sh}

[ -x "$CHECK" ] || { echo "missing or not executable: $CHECK"; exit 2; }

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

# scene builds one complete set of inputs. Each case edits it afterwards, so
# every case starts from a tree that passes.
scene() {
    rm -rf "$work/base" "$work/pkg"
    mkdir -p "$work/base" "$work/pkg"

    # Installed by the manifest, and different from the base.
    echo 'fork version' > "$work/pkg/S10changed"
    echo 'base version' > "$work/base/S10changed"

    # Carried for reference and identical to the base.
    echo 'shared' > "$work/pkg/S20same"
    echo 'shared' > "$work/base/S20same"

    # Deliberately never installed, and absent from the base.
    echo 'policy' > "$work/pkg/S30policy"

    cat > "$work/manifest" <<'EOF'
add     kvmapp/system/init.d/S10changed           /etc/init.d/S10changed       0755
add     tools/abslots/device/rcS                  /etc/init.d/rcS              0755
EOF

    cat > "$work/list" <<'EOF'
# a comment, and a blank line follow
S20same          same-as-base
S30policy        not-installed-by-policy  started on demand
EOF
}

run() {
    "$CHECK" "$work/base" "$work/pkg" "$work/manifest" "$work/list" > "$work/out" 2>&1
    echo $?
}

# ---------------------------------------------------------------- the cases

scene
status=$(run)
[ "$status" = 0 ] && note "a complete set of inputs passes" OK \
                  || note "a complete set of inputs passes" "FAIL (exit $status: $(cat "$work/out"))"

scene
echo 'new' > "$work/pkg/S40new"
status=$(run)
[ "$status" = 1 ] && note "a script in neither the manifest nor the list fails" OK \
                  || note "a script in neither the manifest nor the list fails" "FAIL (exit $status)"
grep -q 'S40new' "$work/out" \
    && note "  and the message names it" OK \
    || note "  and the message names it" "FAIL ($(cat "$work/out"))"

# The S00kmod shape: declared same-as-base, and the fork has since edited it.
scene
echo 'the fork changed this' > "$work/pkg/S20same"
status=$(run)
[ "$status" = 1 ] && note "same-as-base that no longer matches the base fails" OK \
                  || note "same-as-base that no longer matches the base fails" "FAIL (exit $status)"
grep -q 'S20same' "$work/out" \
    && note "  and the message names it" OK \
    || note "  and the message names it" "FAIL ($(cat "$work/out"))"

scene
rm -f "$work/base/S20same"
status=$(run)
[ "$status" = 1 ] && note "same-as-base with no base copy fails" OK \
                  || note "same-as-base with no base copy fails" "FAIL (exit $status)"

# Policy entries are not compared, so one that differs from a base copy of the
# same name is still fine: the point is that it does not run at boot.
scene
echo 'base has one too' > "$work/base/S30policy"
status=$(run)
[ "$status" = 0 ] && note "not-installed-by-policy is never compared" OK \
                  || note "not-installed-by-policy is never compared" "FAIL (exit $status: $(cat "$work/out"))"

scene
echo 'S10changed       same-as-base' >> "$work/list"
status=$(run)
[ "$status" = 1 ] && note "a script both installed and declared fails" OK \
                  || note "a script both installed and declared fails" "FAIL (exit $status)"

scene
echo 'S99gone          same-as-base' >> "$work/list"
status=$(run)
[ "$status" = 1 ] && note "a declaration for a script that is gone fails" OK \
                  || note "a declaration for a script that is gone fails" "FAIL (exit $status)"

scene
echo 'S30policy        because-i-said-so' > "$work/list"
echo 'S20same          same-as-base' >> "$work/list"
status=$(run)
[ "$status" = 1 ] && note "an unknown reason fails" OK \
                  || note "an unknown reason fails" "FAIL (exit $status)"

# A manifest entry that installs from tools/ rather than from the package is
# matched on its destination, so rcS must not be reported as unaccounted for.
scene
grep -q 'rcS' "$work/out" 2>/dev/null
status=$(run)
grep -q 'rcS' "$work/out" \
    && note "a manifest entry from outside the package is not reported" "FAIL ($(cat "$work/out"))" \
    || note "a manifest entry from outside the package is not reported" OK

scene
"$CHECK" "$work/no-such-dir" "$work/pkg" "$work/manifest" "$work/list" > "$work/out" 2>&1
status=$?
[ "$status" = 2 ] && note "no base directory reports cannot-run, not failure" OK \
                  || note "no base directory reports cannot-run, not failure" "FAIL (exit $status)"

echo
if [ "$fails" -eq 0 ]; then
    echo "check-init-install: all cases passed"
    exit 0
fi
echo "check-init-install: $fails case(s) failed"
exit 1
