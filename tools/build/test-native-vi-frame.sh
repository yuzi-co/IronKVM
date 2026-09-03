#!/bin/sh
# Check the unmapped VI frame path through kvm_mmf.cpp and kvm_vision.cpp.
#
#   test-native-vi-frame.sh [kvm_mmf.cpp] [kvm_vision.cpp]
#
# Not destructive: both sources are only read.
#
# A VI frame is a buffer the capture hardware filled. Taking one and not giving
# it back removes it from the VPSS pool for good, and a few of those end the
# stream: CVI_VPSS_GetChnFrame has nothing left to answer with. That failure
# does not look like a leak from the outside. It looks like the capture stopped.
#
# The encoder reads a frame by its physical address, so a frame that only ever
# goes to the encoder never needs to be mapped into this process, and the cache
# invalidate that comes with mapping covers the whole frame: 3.1MB at 1080p in
# NV12, at whatever rate the stream runs. That is what the unmapped path is for.
# The cost of it is that the release is now the caller's job.
#
# None of this can be run here. Every function involved calls into CVI_VPSS or
# CVI_VENC and there is no stand-in for either. What these cases hold in place
# is that each exit which can be reached holding a frame still hands it back,
# and that the count of those exits has not changed behind the reader's back.
#
# This is a weaker statement than a run on hardware, and it is not a substitute
# for one.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
MMF=${1:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}
VIS=${2:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}

for f in "$MMF" "$VIS"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# A function body with comment lines dropped, so prose about a call is not
# mistaken for the call.
body() { sed -n "/^[a-z0-9_ ]*[ *]$2(/,/^}/p" "$1" | grep -v '^[[:space:]]*//'; }

echo "===== acquiring and mapping are separate ====="

acquire=$(body "$MMF" _mmf_acquire_vi_frame)
printf '%s\n' "$acquire" | grep -q '_mmf_release_vi_frame(ch);' \
    && note "acquiring gives back the frame the channel held" OK \
    || note "acquiring gives back the frame the channel held" FAIL
printf '%s\n' "$acquire" | grep -q 'priv.vi_frame_mapped\[ch\] = false;' \
    && note "an acquired frame is not marked mapped" OK \
    || note "an acquired frame is not marked mapped" FAIL
printf '%s\n' "$acquire" | grep -qE 'CVI_SYS_MmapCache|IonInvalidateCache' \
    && note "acquiring does not map" FAIL \
    || note "acquiring does not map" OK

map=$(body "$MMF" _mmf_map_vi_frame)
printf '%s\n' "$map" | grep -q 'if (priv.vi_frame_mapped\[ch\]) {' \
    && note "mapping an already mapped frame maps nothing" OK \
    || note "mapping an already mapped frame maps nothing" FAIL

native=$(body "$MMF" mmf_vi_frame_pop_native)
printf '%s\n' "$native" | grep -q 'priv.vi_frame_deferred\[ch\] = true;' \
    && note "the native pop marks the frame deferred" OK \
    || note "the native pop marks the frame deferred" FAIL
printf '%s\n' "$native" | grep -qE 'CVI_SYS_MmapCache|_mmf_map_vi_frame' \
    && note "the native pop does not map" FAIL \
    || note "the native pop does not map" OK

pop=$(body "$MMF" mmf_vi_frame_pop)
printf '%s\n' "$pop" | grep -q '_mmf_release_vi_frame(ch);' \
    && note "the mapping pop releases when the mapping fails" OK \
    || note "the mapping pop releases when the mapping fails" FAIL

push=$(body "$MMF" mmf_venc_push_vi)
printf '%s\n' "$push" | grep -q 'priv.vi_frame_valid\[vi_ch\] || !priv.vi_frame_deferred\[vi_ch\]' \
    && note "the native push refuses a frame it does not hold" OK \
    || note "the native push refuses a frame it does not hold" FAIL
printf '%s\n' "$push" | grep -q '_mmf_map_vi_frame(vi_ch)' \
    && note "the native push maps only to fall back to a copy" OK \
    || note "the native push maps only to fall back to a copy" FAIL

echo
echo "===== every exit holding a frame gives it back ====="

# frame_to_h264 reaches mmf_venc_free on every path but one: a push that fails
# leaves the channel not running, and mmf_venc_free returns early in that state
# without releasing anything.
enc=$(body "$VIS" frame_to_h264)
printf '%s\n' "$enc" | grep -q 'mmf_vi_frame_release(vi_ch);' \
    && note "a failed push releases the frame it was given" OK \
    || note "a failed push releases the frame it was given" FAIL

releases=$(awk '/^int kvmv_read_img/,/^}/' "$VIS" | grep -c 'mmf_vi_frame_release(native_vi_ch)')
[ "$releases" = 2 ] \
    && note "kvmv_read_img releases at both of its early exits ($releases)" OK \
    || note "kvmv_read_img releases at both of its early exits ($releases, wanted 2)" FAIL

# The count above only means something while the number of ways out of that
# region is the number it was written against. Between taking the frame and
# handing it to the encoder there are seven:
#
#   two release the frame, and are the two counted above;
#   one leaves because there was no frame at all;
#   one belongs to the frame detector, which only runs for MJPEG;
#   three belong to the MJPEG encode.
#
# The last four cannot be reached holding a native frame while only H.264 takes
# one. A new way out, or MJPEG learning to take one, changes that and this case
# is what says so.
exits=$(awk '/int native_vi_ch = -1;/,/frame_to_h264\(NULL/' "$VIS" \
    | grep -cE '^[[:space:]]*(return |continue;)')
[ "$exits" = 7 ] \
    && note "the region still has seven ways out ($exits)" OK \
    || note "the region has $exits ways out, not the seven this was written against" FAIL

printf '%s\n' "$(awk '/^int kvmv_read_img/,/^}/' "$VIS")" | grep -q 'if (_type == VENC_H264) {' \
    && note "only H.264 takes the unmapped frame" OK \
    || note "only H.264 takes the unmapped frame" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
