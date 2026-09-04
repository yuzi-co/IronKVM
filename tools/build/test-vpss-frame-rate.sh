#!/bin/sh
# Check that the VI channel is told how many frames a second to hand out.
#
#   test-vpss-frame-rate.sh [kvm_mmf.cpp] [kvm_mmf.hpp] [kvm_vision.cpp] [kvm_vision.h]
#
# Not destructive: the sources are only read.
#
# The VPSS channel came up with s32SrcFrameRate and s32DstFrameRate set to the
# same value, which asks the hardware to drop nothing, because frame rate
# control only drops when the destination is below the source. It therefore
# handed out every frame a 60Hz source produced while the server read at its
# configured 30. Measured on 2026-09-04: VIFPS 60, VIDevFPS about 50, in every
# mode including H.264 with the core 72% idle.
#
# Each frame nobody reads is still 3.1MB written to DDR at 1080p, so the cost is
# memory bandwidth and it never appears in a CPU figure. That is the whole
# reason these cases read the source rather than a measurement.
#
# None of this can be run here: every call goes to CVI_VPSS and there is no
# stand-in. This is weaker than a run on hardware and not a substitute for one.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
MMF=${1:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}
HPP=${2:-$ROOT/support/sg2002/additional/kvm_mmf/include/kvm_mmf.hpp}
VIS=${3:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}
VHD=${4:-$ROOT/server/include/kvm_vision.h}
LHD=${5:-$ROOT/support/sg2002/additional/kvm/include/kvm_vision.h}

for f in "$MMF" "$HPP" "$VIS" "$VHD" "$LHD"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

body() { sed -n "/^[a-z0-9_ ]*[ *]$2(/,/^}/p" "$1" | grep -v '^[[:space:]]*//'; }

echo "===== the channel is built with a rate of its own ====="

init=$(awk '/^static CVI_S32 _mmf_vpss_chn_init/,/^}/' "$MMF" | grep -v '^[[:space:]]*//')

printf '%s\n' "$init" | grep -q 'int fps, int dst_fps, int depth' \
    && note "the channel init takes a source and an output rate" OK \
    || note "the channel init takes a source and an output rate" FAIL

# The bug was that both came from the same variable. Every branch that sets a
# destination has to set it from dst_fps, and none from fps.
src=$(printf '%s\n' "$init" | grep -c 's32SrcFrameRate = fps;')
dstgood=$(printf '%s\n' "$init" | grep -c 's32DstFrameRate = dst_fps;')
dstbad=$(printf '%s\n' "$init" | grep -c 's32DstFrameRate = fps;')
[ "$src" -gt 0 ] && [ "$dstgood" = "$src" ] && [ "$dstbad" = 0 ] \
    && note "every branch sets the output rate from dst_fps ($dstgood of $src)" OK \
    || note "$dstbad branch(es) still set the output rate from fps" FAIL

printf '%s\n' "$init" | grep -q 'if (dst_fps <= 0 || dst_fps > fps)' \
    && note "an absent or impossible rate falls back to the source rate" OK \
    || note "an absent or impossible rate falls back to the source rate" FAIL

echo
echo "===== the rate survives a channel rebuild ====="

# A resolution change destroys the channel and builds another. The rate has to
# live somewhere the rebuild reads, or the board quietly returns to 60.
grep -q 'int vi_chn_dst_fps\[MMF_VI_MAX_CHN\];' "$MMF" \
    && note "the wanted rate is kept per channel" OK \
    || note "the wanted rate is kept per channel" FAIL

add=$(awk '/^static int _mmf_add_vi_channel/,/^}/' "$MMF" | grep -v '^[[:space:]]*//')
printf '%s\n' "$add" | grep -q 'priv.vi_chn_dst_fps\[ch\], depth, mirror, flip' \
    && note "a rebuilt channel comes up at the rate that was asked for" OK \
    || note "a rebuilt channel comes up at the rate that was asked for" FAIL

set_fps=$(body "$MMF" mmf_vi_set_chn_fps)

# Recording the rate has to happen before the "is a channel open" test, so a
# rate set while capture is down still reaches the channel that comes up next.
printf '%s\n' "$set_fps" | awk '
    /priv.vi_chn_dst_fps\[ch\] = dst_fps;/ { stored = NR }
    /!priv.vi_chn_is_inited\[ch\]/ { checked = NR }
    END { exit !(stored && checked && stored < checked) }' \
    && note "the rate is recorded before the channel is looked at" OK \
    || note "the rate is recorded before the channel is looked at" FAIL

printf '%s\n' "$set_fps" | grep -q 'CVI_VPSS_GetChnAttr(0, ch, &chn_attr)' \
    && note "it changes the attributes the channel has" OK \
    || note "it changes the attributes the channel has" FAIL
printf '%s\n' "$set_fps" | grep -q 's32SrcFrameRate' \
    && note "it clamps against the rate the source actually runs at" OK \
    || note "it clamps against the rate the source actually runs at" FAIL
printf '%s\n' "$set_fps" | grep -q 'CVI_VPSS_SetChnAttr(0, ch, &chn_attr)' \
    && note "it writes the attributes back" OK \
    || note "it writes the attributes back" FAIL

grep -q 'int mmf_vi_set_chn_fps(int ch, int dst_fps);' "$HPP" \
    && note "callers outside kvm_mmf.cpp can see it" OK \
    || note "callers outside kvm_mmf.cpp can see it" FAIL

echo
echo "===== libkvm applies it on the read, under the lock ====="

set_cap=$(body "$VIS" set_capture_fps)

printf '%s\n' "$set_cap" | grep -q 'maxmin_data(60, 10' \
    && note "the rate a caller asks for is clamped" OK \
    || note "the rate a caller asks for is clamped" FAIL

# The detector thread rebuilds the channel. Touching CVI_VPSS from the setter
# would race with that, so the setter only records and the next read applies it.
printf '%s\n' "$set_cap" | grep -qE 'CVI_VPSS|mmf_vi_set_chn_fps' \
    && note "the setter does not touch the channel itself" FAIL \
    || note "the setter does not touch the channel itself" OK
printf '%s\n' "$set_cap" | grep -q 'kvmvi_fps_pending = 1;' \
    && note "the setter leaves the work for the next read" OK \
    || note "the setter leaves the work for the next read" FAIL

read_img=$(awk '/^int kvmv_read_img/,/^}/' "$VIS" | grep -v '^[[:space:]]*//')
printf '%s\n' "$read_img" | grep -q 'mmf_vi_set_chn_fps(cam->get_channel()' \
    && note "the read applies it, which is under vi_mutex" OK \
    || note "the read applies it, which is under vi_mutex" FAIL

# Default 0 means hand out everything, which is what the channel did before any
# of this existed, so a server too old to call the setter is unaffected.
grep -q '^uint8_t kvmvi_dst_fps = 0;' "$VIS" \
    && note "the rate defaults to handing out every frame" OK \
    || note "the rate defaults to handing out every frame" FAIL

grep -q 'void set_capture_fps(uint8_t _fps) __attribute__((weak));' "$VHD" \
    && note "the server links it weakly, as it does set_h264_fps" OK \
    || note "the server links it weakly, as it does set_h264_fps" FAIL
grep -q 'set_capture_fps_if_available' "$VHD" \
    && note "the server has a guarded way to call it" OK \
    || note "the server has a guarded way to call it" FAIL

# Every entry point the server calls has to be declared in the libkvm header as
# well, because that is the header with extern "C" round it. Declared only on
# the server side, the definition comes out C++ mangled. The server links these
# weakly, so that does not fail to link: the reference resolves to zero and the
# call quietly does nothing, which looks exactly like a deploy that worked.
# set_capture_fps was written that way first and the symbol table is what said
# so.
for fn in set_h264_gop set_h264_fps set_capture_fps set_frame_detact; do
    grep -q "^void $fn(uint8_t" "$LHD" \
        && note "$fn is declared where extern \"C\" reaches it" OK \
        || note "$fn is declared where extern \"C\" reaches it" FAIL
done

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
