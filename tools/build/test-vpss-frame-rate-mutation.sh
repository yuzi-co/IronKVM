#!/bin/sh
# Check that test-vpss-frame-rate.sh fails when the thing it describes is broken.
#
#   test-vpss-frame-rate-mutation.sh
#
# Not destructive: every mutation is applied to a copy in a temporary directory
# and the shipped sources are never written.
#
# A suite that reads sources passes on an empty file as happily as on a correct
# one. Each case below breaks one property on purpose and requires the suite to
# notice. The first mutation is the bug this change removed, so that one is the
# case that matters most: the suite has to reject the code as it used to be.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SUITE=$ROOT/tools/build/test-vpss-frame-rate.sh
MMF=$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp
HPP=$ROOT/support/sg2002/additional/kvm_mmf/include/kvm_mmf.hpp
VIS=$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp
VHD=$ROOT/server/include/kvm_vision.h
LHD=$ROOT/support/sg2002/additional/kvm/include/kvm_vision.h

for f in "$SUITE" "$MMF" "$HPP" "$VIS" "$VHD" "$LHD"; do
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
    what=$1; target=$2; script=$3
    cp "$MMF" "$work/m.cpp"; cp "$HPP" "$work/m.hpp"
    cp "$VIS" "$work/v.cpp"; cp "$VHD" "$work/v.h"; cp "$LHD" "$work/l.h"
    sed "$script" "$work/$target" > "$work/mutated" && mv "$work/mutated" "$work/$target"
    case $target in
        m.cpp) src=$MMF ;;
        m.hpp) src=$HPP ;;
        v.cpp) src=$VIS ;;
        l.h)   src=$LHD ;;
        *)     src=$VHD ;;
    esac
    if cmp -s "$work/$target" "$src"; then
        note "$what" "FAIL (the mutation changed nothing)"
        return 0
    fi
    if sh "$SUITE" "$work/m.cpp" "$work/m.hpp" "$work/v.cpp" "$work/v.h" "$work/l.h" >/dev/null 2>&1; then
        note "$what" "FAIL (survived)"
    else
        note "$what" OK
    fi
}

echo "===== the suite rejects the bug this removed ====="

mutate "source and destination the same value again" m.cpp \
    's|chn_attr.stFrameRate.s32DstFrameRate = dst_fps;|chn_attr.stFrameRate.s32DstFrameRate = fps;|'

mutate "one branch of three left on the old value" m.cpp \
    '0,/chn_attr.stFrameRate.s32DstFrameRate = dst_fps;/ s||chn_attr.stFrameRate.s32DstFrameRate = fps;|'

mutate "the output rate taken off the signature" m.cpp \
    's|int fps, int dst_fps, int depth|int fps, int depth|'

mutate "the fallback for an absent rate dropped" m.cpp \
    's|if (dst_fps <= 0 \|\| dst_fps > fps) {|if (0) {|'

echo
echo "===== the suite notices when a rebuild loses the rate ====="

mutate "the per-channel rate dropped" m.cpp \
    's|^\tint vi_chn_dst_fps\[MMF_VI_MAX_CHN\];||'

mutate "a rebuilt channel back at the source rate" m.cpp \
    's|priv.vi_chn_dst_fps\[ch\], depth, mirror, flip|0, depth, mirror, flip|'

# Moves the store past the early return, so a rate set while capture is down is
# forgotten and the channel that comes up next runs at the source rate.
mutate "the rate recorded only when a channel is open" m.cpp \
    '/^int mmf_vi_set_chn_fps/,/^}/ {
       s|^\tpriv.vi_chn_dst_fps\[ch\] = dst_fps;||
       s|^\tVPSS_CHN_ATTR_S chn_attr;|\tpriv.vi_chn_dst_fps[ch] = dst_fps;\n\tVPSS_CHN_ATTR_S chn_attr;| }'

mutate "the current attributes invented rather than read" m.cpp \
    's|CVI_S32 s32Ret = CVI_VPSS_GetChnAttr(0, ch, &chn_attr);|CVI_S32 s32Ret = CVI_SUCCESS;|'

mutate "no clamp against the real source rate" m.cpp \
    's|int src = chn_attr.stFrameRate.s32SrcFrameRate;|int src = 0;|'

mutate "the declaration dropped from the header" m.hpp \
    's|^int mmf_vi_set_chn_fps(int ch, int dst_fps);||'

echo
echo "===== the suite notices when libkvm applies it wrongly ====="

mutate "the setter reaching for the channel itself" v.cpp \
    '/^void set_capture_fps/,/^}/ s|^    kvmvi_fps_pending = 1;|    mmf_vi_set_chn_fps(0, (int)fps);|'

mutate "the caller rate no longer clamped" v.cpp \
    's|uint8_t fps = maxmin_data(60, 10, (int)_fps);|uint8_t fps = _fps;|'

mutate "nothing applying it on the read" v.cpp \
    's|            mmf_vi_set_chn_fps(cam->get_channel(), (int)kvmvi_dst_fps);||'

mutate "a default that caps the channel before anyone asks" v.cpp \
    's|^uint8_t kvmvi_dst_fps = 0;|uint8_t kvmvi_dst_fps = 30;|'

mutate "the weak attribute dropped from the header" v.h \
    's|void set_capture_fps(uint8_t _fps) __attribute__((weak));|void set_capture_fps(uint8_t _fps);|'

# The one that nearly shipped. Declared only on the server side, the definition
# comes out C++ mangled, and because the server links it weakly that does not
# fail: the reference resolves to zero and the call does nothing at all, which
# from the outside looks exactly like a deploy that worked.
mutate "declared outside the header extern C reaches" l.h \
    's|^void set_capture_fps(uint8_t _fps);||'

mutate "an older setter losing the same declaration" l.h \
    's|^void set_h264_fps(uint8_t _fps);||'

echo
if [ "$fails" -eq 0 ]; then
    echo "all mutations caught"
    exit 0
fi
echo "$fails mutation(s) NOT caught"
exit 1
