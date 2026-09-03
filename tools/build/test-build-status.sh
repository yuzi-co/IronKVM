#!/bin/sh
# Check that support/sg2002/build reports a failed library build as a failure.
#
#   test-build-status.sh
#
# Not destructive: the shipped script is only read, and every case runs against
# a copy of its two status functions inside mktemp.
#
# On 2026-09-04 a rebuild of libkvm.so failed with a PermissionError from
# maixcdk and the script printed "Build completed!". The stale dist directory
# from an earlier build could not be removed, chack_build tested only that the
# directory existed, and the directory existed. The library that a following
# `add_to_kvmapp` would have shipped was three weeks old and built from
# different sources.
#
# That is the worst shape a build failure can take. A build that stops is an
# inconvenience; a build that reports success while the artefact on disk is
# somebody else's puts the wrong library in a release. These cases hold the
# script to the two questions it must ask: did the builder exit zero, and is
# the artefact there.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SRC=${1:-$ROOT/support/sg2002/build}

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# Take the functions out of the shipped script rather than copying them here.
# A copy drifts, and a test that checks a copy checks nothing.
lift() { sed -n "/^$1()[[:space:]]*{/,/^}/p" "$SRC"; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

for fn in maixcdk_build chack_build; do
    lift "$fn" > "$work/$fn.sh"
    [ -s "$work/$fn.sh" ] || { echo "cannot lift $fn from $SRC"; exit 2; }
done

# Run one case in its own directory with a stubbed builder. $1 is the status
# maixcdk exits with, $2 is "writes" when the builder produces the artefact,
# and $3 is "rmfails" when the dist directory cannot be removed.
#
# The artefact is seeded before the run only in the rmfails case, because that
# is the only way a file from an earlier build survives into this one: a real
# build starts by deleting dist.
run_case() {
    status=$1
    writes=$2
    rmfails=$3

    d=$work/case
    rm -rf "$d"
    mkdir -p "$d/dist/kvm_vision_test_release/dl_lib"
    lib=$d/dist/kvm_vision_test_release/dl_lib/libkvm.so
    [ "$rmfails" = rmfails ] && echo stale > "$lib"

    (
        cd "$d" || exit 2
        Build_Artifact=$lib
        # The builder. The real one is a Python entry point that exits non-zero
        # when it raises, and it links nothing when it fails.
        maixcdk() {
            cat > /dev/null 2>&1
            if [ "$writes" = writes ]; then
                mkdir -p "$(dirname "$lib")" && echo fresh > "$lib"
            fi
            return "$status"
        }
        # A dist directory owned by another user is what actually happened, and
        # root can remove one anyway, so model the refusal rather than trying to
        # provoke it with permissions.
        if [ "$rmfails" = rmfails ]; then
            rm() { echo "rm: cannot remove: Permission denied" >&2; return 1; }
        fi

        . "$work/maixcdk_build.sh"
        . "$work/chack_build.sh"

        maixcdk_build "from_zero"
        chack_build "$?"
    )
}

echo "===== the build status is the builder's status ====="

res=$(run_case 0 writes)
[ "$res" = SUCCESS ] \
    && note "a builder that exits zero and links reports SUCCESS" OK \
    || note "a builder that exits zero and links reports SUCCESS (got '$res')" FAIL

res=$(run_case 1 none)
[ "$res" = FAILURE ] \
    && note "a builder that fails reports FAILURE" OK \
    || note "a builder that fails reports FAILURE (got '$res')" FAIL

# A link step that produced nothing while the driver still exited zero.
res=$(run_case 0 none)
[ "$res" = FAILURE ] \
    && note "a missing artefact reports FAILURE" OK \
    || note "a missing artefact reports FAILURE (got '$res')" FAIL

# The other half: an artefact alone is not evidence, because dist is deleted
# and rebuilt, so a file present after a failure came from somewhere else.
res=$(run_case 1 writes)
[ "$res" = FAILURE ] \
    && note "an artefact does not excuse a non-zero exit" OK \
    || note "an artefact does not excuse a non-zero exit (got '$res')" FAIL

echo
echo "===== output that cannot be replaced stops the build ====="

# The 2026-09-04 failure, exactly: dist could not be removed, the three week
# old library survived in it, and the script printed "Build completed!". The
# builder must not run at all here.
out=$(run_case 0 writes rmfails 2>&1)
res=$(printf '%s' "$out" | tail -1)
[ "$res" = FAILURE ] \
    && note "an unremovable dist directory reports FAILURE" OK \
    || note "an unremovable dist directory reports FAILURE (got '$res')" FAIL

printf '%s' "$out" | grep -q 'Check who owns it' \
    && note "the message names the cause" OK \
    || note "the message names the cause" FAIL

echo
echo "===== every project declares its artefact ====="

# chack_build cannot test an artefact it was never given, and "None" is the
# unset value. A project added without one would report SUCCESS on the strength
# of the exit status alone, which is half the check.
projects=$(grep -c '^[[:space:]]*Project_PATH="\$NanoKVM_PATH' "$SRC")
artifacts=$(grep -c '^[[:space:]]*Build_Artifact="\$Project_PATH' "$SRC")
[ "$projects" -gt 0 ] && [ "$artifacts" = "$projects" ] \
    && note "each Project_PATH is followed by a Build_Artifact ($artifacts/$projects)" OK \
    || note "each Project_PATH is followed by a Build_Artifact ($artifacts/$projects)" FAIL

# The two artefacts must be the paths add_to_kvmapp reads, or the build proves
# the presence of a file that nothing ships.
for path in dist/kvm_system_release/kvm_system dist/kvm_vision_test_release/dl_lib; do
    grep -q "Build_Artifact=\"\$Project_PATH/$path" "$SRC" \
        && note "an artefact is declared under $path" OK \
        || note "an artefact is declared under $path" FAIL
done

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
