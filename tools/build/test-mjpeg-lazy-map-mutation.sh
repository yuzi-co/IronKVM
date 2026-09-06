#!/bin/sh
# Check that test-mjpeg-lazy-map.sh fails when the thing it describes is broken.
#
#   test-mjpeg-lazy-map-mutation.sh
#
# Not destructive: every mutation is applied to a copy in a temporary directory
# and the shipped sources are never written.
#
# A suite that reads sources passes on an empty file just as happily as on a
# correct one. Each case below breaks one property on purpose and requires the
# suite to notice. A mutation that survives means the case that was supposed to
# cover it does not.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SUITE=$ROOT/tools/build/test-mjpeg-lazy-map.sh
VIS=$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp
MMF=$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp
HPP=$ROOT/support/sg2002/additional/kvm_mmf/include/kvm_mmf.hpp

for f in "$SUITE" "$VIS" "$MMF" "$HPP"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done

# A mutant that never finishes is not a mutant the suite caught, and a suite
# that runs one unbounded takes its parent down with it: on 2026-09-06 a mutated
# source sent AddressSanitizer into a signal-handler loop, and this suite sat on
# it until tools/run-tests.sh killed the pair at 1800 seconds. A bound turns that
# into a line of output.
#
# The result is deliberately not counted as "the suite caught it". The suite
# reported nothing, so nothing is known, and pretending otherwise is how a
# mutation suite starts passing for the wrong reason.
MUTANT_TIMEOUT=${MUTANT_TIMEOUT:-120}
run_bounded() {
    if command -v timeout > /dev/null 2>&1; then
        timeout "$MUTANT_TIMEOUT" "$@"
        status=$?
        [ "$status" = 124 ] && return 124
        return "$status"
    fi
    "$@"
}


CXX=${CXX:-g++}
command -v "$CXX" >/dev/null 2>&1 || {
    echo "test-mjpeg-lazy-map-mutation.sh: needs $CXX, which is not on PATH." >&2
    exit 2
}

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

fails=0
# A verdict here carries its reason, so the match is on the prefix. Testing
# for a bare FAIL counted none of them and reported a clean run.
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

# Runs the suite against one mutated copy. The mutation has to change the file,
# because a sed that matched nothing would otherwise read as a caught mutation
# while proving nothing at all.
mutate() {
    what=$1; target=$2; script=$3
    cp "$VIS" "$work/v.cpp"; cp "$MMF" "$work/m.cpp"; cp "$HPP" "$work/m.hpp"
    sed "$script" "$work/$target" > "$work/mutated" && mv "$work/mutated" "$work/$target"
    case $target in
        v.cpp) src=$VIS ;;
        m.cpp) src=$MMF ;;
        *)     src=$HPP ;;
    esac
    if cmp -s "$work/$target" "$src"; then
        note "$what" "FAIL (the mutation changed nothing)"
        return 0
    fi
    run_bounded sh "$SUITE" "$work/v.cpp" "$work/m.cpp" "$work/m.hpp" >/dev/null 2>&1
    status=$?
    if [ "$status" = 0 ]; then
        note "$what" "FAIL (survived)"
    elif [ "$status" = 124 ]; then
        # The suite never answered. A timeout exits non-zero, which would read
        # as "caught", and those are not the same thing.
        note "$what" "FAIL (no answer in ${MUTANT_TIMEOUT}s)"
    else
        note "$what" OK
    fi
}

echo "===== the suite notices when the map moves ====="

mutate "a map on every frame, not just the sampled one" v.cpp \
    's|int native_vi_ch = cam->get_channel();|&\n        (void)mmf_vi_frame_map(0);|'

mutate "the map taken out of the sampling branch" v.cpp \
    's|(const uint8_t \*)mmf_vi_frame_map(native_vi_ch);|(const uint8_t *)NULL;|'

mutate "sampling something other than what was mapped" v.cpp \
    's|frame_changed(pixels, native_len)|frame_changed(NULL, native_len)|'

mutate "an unchanged frame kept rather than given back" v.cpp \
    '/debug("\[kvmv\]frame not changed/,/return 5;/ s|mmf_vi_frame_release(native_vi_ch);||'

echo
echo "===== the suite notices when the copy path comes back ====="

mutate "a frame taken through cam->read() again" v.cpp \
    's|int native_vi_ch = cam->get_channel();|&\n        cam->read();|'

mutate "an encode through Image::to_jpeg again" v.cpp \
    's|int jpeg_ret = frame_to_jpeg(native_vi_ch, p_kvmv_data,|image::Image *j = img->to_jpeg(90);\n            int jpeg_ret = frame_to_jpeg(native_vi_ch, p_kvmv_data,|'

mutate "jpg_dump brought back" v.cpp \
    's|^static int8_t frame_to_jpeg|bool jpg_dump(kvmv_data_t* d, image::Image *r){return false;}\nstatic int8_t frame_to_jpeg|'

mutate "the unknown encode type falling through again" v.cpp \
    's|debug("\[kvmv\]unknown encode type %d\\n", (int)_type);||'

echo
echo "===== the suite notices when frame_changed misbehaves ====="

mutate "an unreadable frame reported as unchanged" v.cpp \
    's|    if(data == NULL \|\| size <= 0){\n        return 1;|XXX|; /if(data == NULL/,/}/ s|return 1;|return 0;|'

mutate "a size change no longer counted as a change" v.cpp \
    '/if(size != raw_size){/,/}/ s|ret = 1;|ret = 0;|'

mutate "the sample stride collapsed to a single byte" v.cpp \
    's|sample_byte = \*(data+(i\*Detection_Pixel_Interval));|sample_byte = *(data+0);|'

mutate "frame_changed taking an Image again" v.cpp \
    's|uint8_t frame_changed(const uint8_t \*data, int size)|uint8_t frame_changed(image::Image *raw)|'

echo
echo "===== the suite notices when the export goes ====="

mutate "the exported map no longer delegating" m.cpp \
    's|^\treturn _mmf_map_vi_frame(ch);|\treturn NULL;|'

mutate "the declaration dropped from the header" m.hpp \
    's|^void \*mmf_vi_frame_map(int ch);||'

echo
if [ "$fails" -eq 0 ]; then
    echo "all mutations caught"
    exit 0
fi
echo "$fails mutation(s) NOT caught"
exit 1
