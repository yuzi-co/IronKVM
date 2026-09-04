#!/bin/sh
# Check who releases the VI frame on each exit from frame_to_h264.
#
#   test-h264-frame-ownership.sh [kvm_vision.cpp [kvm_mmf.cpp]]
#
# Not destructive: both sources are only read.
#
# frame_to_h264 owns the VI frame from the moment it is called, the same way
# frame_to_jpeg does. The frame is one block from a pool of two, so an exit
# that keeps it stalls the pipeline within a frame or two, and an exit that
# releases it twice corrupts the pool.
#
# The two failure exits look inconsistent and are not, which is why this
# suite exists. Read on its own, the push failure releases the frame by hand
# and the pop failure appears to leak it. It does not: the pop failure runs
# after a push that succeeded, so the encoder channel is running, and
# mmf_venc_free releases every held VI frame whenever it finds the channel in
# that state. The push failure has to release by hand for the opposite
# reason, because it returns while the channel never started and
# mmf_venc_free returns early without releasing anything.
#
# That reasoning depends on a fact in a different file, so this suite checks
# that fact too. If mmf_venc_free ever stops releasing frames, the pop path
# starts leaking and nothing in kvm_vision.cpp will look wrong.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VIS=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}
MMF=${2:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}

[ -f "$VIS" ] || { echo "missing: $VIS"; exit 2; }
[ -f "$MMF" ] || { echo "missing: $MMF"; exit 2; }

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

# Comment lines are dropped, so the prose above each exit is not mistaken for
# the code it describes. This whole suite would otherwise pass on comments.
body() {
    sed -n "/^int8_t $2(/,/^}/p" "$1" | grep -v '^[[:space:]]*//'
}

h264=$(body "$VIS" frame_to_h264)
[ -n "$h264" ] || { echo "could not read frame_to_h264 out of $VIS"; exit 2; }

echo "===== the push failure releases by hand ====="

push_exit=$(printf '%s\n' "$h264" | sed -n '/if (push_ret) {/,/return -1;/p')

printf '%s\n' "$push_exit" | grep -q 'mmf_vi_frame_release(vi_ch)' \
    && note "the push failure releases the frame" OK \
    || note "the push failure releases the frame" FAIL

# Only when the frame came from the VI channel. The data pointer path has no
# frame to give back, and releasing on vi_ch of -1 is a call on a channel
# index that does not exist.
printf '%s\n' "$push_exit" | grep -B2 'mmf_vi_frame_release(vi_ch)' \
    | grep -q 'if (vi_ch >= 0)' \
    && note "it releases only the frame it was handed" OK \
    || note "it releases only the frame it was handed" FAIL

echo "===== the pop failure lets mmf_venc_free release ====="

pop_exit=$(printf '%s\n' "$h264" | sed -n '/if (mmf_venc_pop(/,/return -1;/p')
[ -n "$pop_exit" ] || { echo "could not read the pop failure exit"; exit 2; }

printf '%s\n' "$pop_exit" | grep -q 'mmf_venc_free(' \
    && note "the pop failure calls mmf_venc_free" OK \
    || note "the pop failure calls mmf_venc_free" FAIL

# The case that matters. A release added here is a second release of the same
# block, because mmf_venc_free has already given it back.
printf '%s\n' "$pop_exit" | grep -q 'mmf_vi_frame_release' \
    && note "the pop failure does not release the frame itself" FAIL \
    || note "the pop failure does not release the frame itself" OK

# Ordering, because mmf_del_venc_channel is what makes the channel stop
# reporting itself as running.
free_at=$(printf '%s\n' "$pop_exit" | grep -n 'mmf_venc_free(' | head -1 | cut -d: -f1)
del_at=$(printf '%s\n' "$pop_exit" | grep -n 'mmf_del_venc_channel(' | head -1 | cut -d: -f1)
if [ -n "$free_at" ] && [ -n "$del_at" ] && [ "$free_at" -lt "$del_at" ]; then
    note "mmf_venc_free runs before the channel is deleted" OK
else
    note "mmf_venc_free runs before the channel is deleted" FAIL
fi

# It said "push failed" on the pop path, which sent a reader looking at the
# wrong call.
printf '%s\n' "$pop_exit" | grep -q 'venc pop failed' \
    && note "the pop failure says pop and not push" OK \
    || note "the pop failure says pop and not push" FAIL

echo "===== the success exit ====="

# The success path releases through the same mmf_venc_free, so it must not
# release by hand either.
success=$(printf '%s\n' "$h264" | sed -n '/h264_stream_dump(ret_stream/,$p')
printf '%s\n' "$success" | grep -q 'mmf_venc_free(' \
    && note "the success exit calls mmf_venc_free" OK \
    || note "the success exit calls mmf_venc_free" FAIL

printf '%s\n' "$success" | grep -q 'mmf_vi_frame_release' \
    && note "the success exit does not release the frame itself" FAIL \
    || note "the success exit does not release the frame itself" OK

echo "===== every exit is accounted for ====="

# Three returns: the push failure, the pop failure and the success. A fourth
# added later has to be looked at by hand, so fail rather than pass silently.
returns=$(printf '%s\n' "$h264" | grep -c 'return ')
[ "$returns" = 3 ] \
    && note "frame_to_h264 has the 3 exits this suite covers" OK \
    || note "frame_to_h264 has $returns exits, this suite covers 3" FAIL

echo "===== what the pop path depends on, in kvm_mmf.cpp ====="

venc_free=$(sed -n '/^int mmf_venc_free(int ch) {/,/^}/p' "$MMF")
[ -n "$venc_free" ] || { echo "could not read mmf_venc_free out of $MMF"; exit 2; }

printf '%s\n' "$venc_free" | grep -q '_mmf_release_all_vi_frames()' \
    && note "mmf_venc_free releases the held VI frames" OK \
    || note "mmf_venc_free releases the held VI frames" FAIL

# The early return is the reason the push path cannot rely on this one.
printf '%s\n' "$venc_free" | grep -q '!info->is_running' \
    && note "it returns early while the channel is not running" OK \
    || note "it returns early while the channel is not running" FAIL

# And the release has to be after that test, not before it, or the push
# failure path would release twice.
run_at=$(printf '%s\n' "$venc_free" | grep -n '!info->is_running' | head -1 | cut -d: -f1)
rel_at=$(printf '%s\n' "$venc_free" | grep -n '_mmf_release_all_vi_frames()' | head -1 | cut -d: -f1)
if [ -n "$run_at" ] && [ -n "$rel_at" ] && [ "$run_at" -lt "$rel_at" ]; then
    note "the release is behind the running test" OK
else
    note "the release is behind the running test" FAIL
fi

# Both push entry points set is_running on every success, which is what makes
# "the push succeeded" and "the channel is running" the same statement.
native=$(sed -n '/^int mmf_venc_push_vi(int ch, int vi_ch) {/,/^}/p' "$MMF")
printf '%s\n' "$native" | grep -q 'info->is_running = 1' \
    && note "a native push marks the channel running" OK \
    || note "a native push marks the channel running" FAIL

copy=$(sed -n '/^static int _mmf_venc_push_copy(/,/^}/p' "$MMF")
printf '%s\n' "$copy" | grep -q 'info->is_running = 1' \
    && note "the copy fallback marks the channel running" OK \
    || note "the copy fallback marks the channel running" FAIL

# The fallback is reached from mmf_venc_push_vi when CVI_VENC_SendFrame
# declines the native frame, so it is on the pop path too.
printf '%s\n' "$native" | grep -q '_mmf_venc_push_copy(' \
    && note "the native push falls back to the copy path" OK \
    || note "the native push falls back to the copy path" FAIL

echo "===== the jpeg counterpart, which owns it the other way ====="

# frame_to_jpeg has no encoder channel to hand the frame to, so every one of
# its exits releases by hand. Keeping both shapes in one suite is what stops
# somebody making them consistent with each other.
jpeg=$(sed -n '/^static int8_t frame_to_jpeg(/,/^}/p' "$VIS" | grep -v '^[[:space:]]*//')
jreturns=$(printf '%s\n' "$jpeg" | grep -c 'return ')
jreleases=$(printf '%s\n' "$jpeg" | grep -c 'mmf_vi_frame_release(vi_ch)')
[ -n "$jpeg" ] || { echo "could not read frame_to_jpeg out of $VIS"; exit 2; }
[ "$jreturns" = "$jreleases" ] \
    && note "frame_to_jpeg releases on all $jreturns of its exits" OK \
    || note "frame_to_jpeg has $jreturns exits and $jreleases releases" FAIL

echo
if [ "$fails" = 0 ]; then
    echo "test-h264-frame-ownership.sh: all cases passed"
    exit 0
fi
echo "test-h264-frame-ownership.sh: $fails case(s) failed"
exit 1
