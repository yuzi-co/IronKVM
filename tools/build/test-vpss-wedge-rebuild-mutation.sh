#!/bin/sh
# Break the wedge detection on purpose and check that the suite next to this
# one notices.
#
#   test-vpss-wedge-rebuild-mutation.sh [kvm_vision.cpp]
#
# Not destructive: the source is copied first and only the copy is edited.
#
# Each case edits one thing, confirms the edit landed, and then expects
# test-vpss-wedge-rebuild.sh to fail against the copy. A mutation that survives
# means the suite passes for a reason other than the one it claims, and the
# usual cause is an edit that never applied rather than a missing case, so the
# applied check comes first and reports separately.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VIS=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}
SUITE=$ROOT/tools/build/test-vpss-wedge-rebuild.sh

[ -f "$VIS" ] || { echo "missing: $VIS"; exit 2; }
[ -f "$SUITE" ] || { echo "missing: $SUITE"; exit 2; }

CXX=${CXX:-g++}
command -v "$CXX" >/dev/null 2>&1 || {
    echo "test-vpss-wedge-rebuild-mutation.sh: needs $CXX, which is not on PATH." >&2
    exit 2
}

# The suite has to pass on the unedited source, or every case below is
# meaningless.
if ! sh "$SUITE" "$VIS" >/dev/null 2>&1; then
    echo "test-vpss-wedge-rebuild-mutation.sh: the suite fails on the unedited" >&2
    echo "source, so no mutation here proves anything. Fix that first." >&2
    exit 1
fi

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

try() {
    desc=$1
    expr=$2
    cp "$VIS" "$work/m.cpp"
    sed -i "$expr" "$work/m.cpp"
    if cmp -s "$VIS" "$work/m.cpp"; then
        note "$desc" "FAIL (the edit never applied)"
        return 0
    fi
    if sh "$SUITE" "$work/m.cpp" >/dev/null 2>&1; then
        note "$desc" "FAIL (survived)"
    else
        note "$desc" OK
    fi
}

echo "===== the policy ====="

try "a dead VI device no longer stops a rebuild" \
    's/if(vi_live == 0){/if(0){/'

try "the threshold drops to a single failure" \
    's/w->fail_count < wedge_fail_threshold/w->fail_count < 1U/'

try "the minimum run is no longer required" \
    's/now_ms - w->first_fail_ms < wedge_min_run_ms/0/'

try "the cooldown is no longer required" \
    's/w->rebuilt_before != 0 && now_ms - w->last_rebuild_ms < cooldown/0/'

try "the backoff no longer caps" \
    's|cooldown > wedge_max_cooldown_ms / 2|0|'

try "the backoff no longer doubles" \
    's|: cooldown \* 2;|: cooldown;|'

try "a frame no longer clears the run" \
    's/w->fail_count = 0;//'

try "a frame no longer resets the cooldown" \
    '/^void wedge_note_frame(/,/^}/ s/w->cooldown_ms = wedge_first_cooldown_ms;//'

try "the failure counter wraps instead of saturating" \
    's/if(w->fail_count < 0xFFFFFFFFU){/if(1){/'

try "a null state is dereferenced" \
    '/^void wedge_note_frame(/,/^}/ s/if(w == NULL){/if(0){/'

echo "===== the call sites ====="

try "the reader stops noting that a frame arrived" \
    's/            wedge_note_frame(&vi_wedge);//'

try "the reader stops noting an empty call" \
    's/    if (wedge_note_no_frame(&vi_wedge, vi_state_shared::monotonic_ms(),/    if (wedge_note_nothing(\&vi_wedge, vi_state_shared::monotonic_ms(),/'

try "the reader never raises the request" \
    's/        __atomic_store_n(&vi_wedge_rebuild_request, 1, __ATOMIC_RELEASE);//'

try "the reader stops passing the VI device state" \
    's/kvmv_hdmi_signal_active()) != 0) {/1) != 0) {/'

echo "===== the rebuild ====="

try "a transition no longer stands the rebuild down" \
    's/kvmv_cfg.reopen_cam_flag == 0 &&//'

try "the request is no longer cleared" \
    's/__atomic_store_n(&vi_wedge_rebuild_request, 0, __ATOMIC_RELEASE);//'

try "the reader restarts the camera itself" \
    's/        __atomic_store_n(&vi_wedge_rebuild_request, 1, __ATOMIC_RELEASE);/        cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);/'

try "the ask goes back to being invisible" \
    's/printf("\[kvmv\]channel handed out nothing/debug("[kvmv]channel handed out nothing/'

try "the rebuild goes back to being invisible" \
    's/printf("\[kvmv\]rebuilding the VI channel/debug("[kvmv]rebuilding the VI channel/'

try "a rebuild that stands down says nothing" \
    's/printf("\[kvmv\]a rebuild was asked for while/debug("[kvmv]a rebuild was asked for while/'

try "a second lock site appears on vi_mutex" \
    's/    int mutex_res = pthread_mutex_timedlock(&vi_mutex, &ts);/    pthread_mutex_lock(\&vi_mutex);\n    int mutex_res = pthread_mutex_timedlock(\&vi_mutex, \&ts);/'

echo
if [ "$fails" = 0 ]; then
    echo "test-vpss-wedge-rebuild-mutation.sh: all mutations were caught"
    exit 0
fi
echo "test-vpss-wedge-rebuild-mutation.sh: $fails mutation(s) not caught"
exit 1
