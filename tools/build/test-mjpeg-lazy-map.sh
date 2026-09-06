#!/bin/sh
# Check that MJPEG maps a capture frame only when it is about to read one.
#
#   test-mjpeg-lazy-map.sh [kvm_vision.cpp] [kvm_mmf.cpp] [kvm_mmf.hpp]
#
# Not destructive: the sources are only read, and the compiled part is built in
# a temporary directory.
#
# Mapping a VI frame invalidates the cache over all of it, which is 3.1MB at
# 1080p in NV12. The encoder does not need that, because it reads the frame by
# its physical address. Only frame_changed() needs it, and frame_changed() runs
# on one frame in frame_detact.
#
# MJPEG used to give up the unmapped path for every frame whenever the detector
# was on, so fifty-nine frames in sixty paid for a map, a full-frame copy into
# an image::Image and a second copy into the encoder, to decide nothing.
# Measured on hardware at 1080p on 2026-09-04: 95.2% of core with the detector
# on against 91.9% with it off, while delivering 12.7% fewer frames.
#
# The first half of this suite runs frame_changed for real. It reads pixels and
# nothing else now, so it compiles and runs here, and these are the only cases
# in the file that execute the code they are about. The second half reads the
# sources, which is weaker, and is not a substitute for a run on hardware.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VIS=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}
MMF=${2:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}
HPP=${3:-$ROOT/support/sg2002/additional/kvm_mmf/include/kvm_mmf.hpp}

for f in "$VIS" "$MMF" "$HPP"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done

CXX=${CXX:-g++}
command -v "$CXX" >/dev/null 2>&1 || {
    echo "test-mjpeg-lazy-map.sh: needs $CXX, which is not on PATH." >&2
    exit 2
}

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# A function body with comment lines dropped, so prose about a call is not
# mistaken for the call.
body() { sed -n "/^[a-z0-9_ ]*[ *]$2(/,/^}/p" "$1" | grep -v '^[[:space:]]*//'; }

. "$(dirname "$0")/asan-env.sh"

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

echo "===== frame_changed, run for real ====="

# Lift the sample size and the function out of the shipped file, so what runs
# below is what ships rather than a copy that can drift away from it.
grep '^#define Farame_sample_size' "$VIS" > "$work/lifted.h" 2>/dev/null
sed -n '/^uint8_t frame_changed(/,/^}/p' "$VIS" >> "$work/lifted.h"

grep -q 'frame_changed(const uint8_t \*data, int size)' "$work/lifted.h" \
    && note "frame_changed takes pixels and a size" OK \
    || note "frame_changed takes pixels and a size" FAIL

cat > "$work/main.cpp" <<'EOF'
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "lifted.h"

static int fails = 0;
static void check(const char *what, int got, int want)
{
    if (got != want) {
        printf("  %-58s FAIL (got %d, wanted %d)\n", what, got, want);
        fails++;
    } else {
        printf("  %-58s OK\n", what);
    }
}

int main(void)
{
    const int size = 1920 * 1080 * 3 / 2;
    uint8_t *frame = (uint8_t *)malloc(size);
    if (frame == NULL) return 2;
    for (int i = 0; i < size; i++) frame[i] = (uint8_t)(i * 7);

    // The first frame of any size is a change: there is nothing to compare it
    // against.
    check("the first frame counts as changed", frame_changed(frame, size), 1);
    check("the same frame again counts as unchanged", frame_changed(frame, size), 0);
    check("and still unchanged the third time", frame_changed(frame, size), 0);

    // The sample is strided, so a byte it looks at has to move the answer and
    // one it skips has to not. Interval is size/(Farame_sample_size*1.5).
    int interval = (int)(size / (Farame_sample_size * 1.5));
    frame[interval * 3] ^= 0xFF;
    check("a byte the sample reads counts as changed", frame_changed(frame, size), 1);
    check("the changed frame settles back to unchanged", frame_changed(frame, size), 0);

    frame[interval * 3 + 1] ^= 0xFF;
    check("a byte between samples counts as unchanged", frame_changed(frame, size), 0);

    // A frame it cannot read has to count as changed. Reporting "unchanged"
    // for a frame nobody looked at stops the stream on a moving screen.
    check("a null frame counts as changed", frame_changed(NULL, size), 1);
    check("an empty frame counts as changed", frame_changed(frame, 0), 1);
    check("a negative size counts as changed", frame_changed(frame, -1), 1);

    // A size change is a change, and the next comparison is like for like.
    // The native frame carries stride padding that data_size() did not, so
    // this is the path a resolution change takes now.
    int smaller = 1280 * 720 * 3 / 2;
    check("a new frame size counts as changed", frame_changed(frame, smaller), 1);
    check("the new size settles to unchanged", frame_changed(frame, smaller), 0);

    // The two cases above pass whether or not the size is compared at all: a
    // new size moves the stride, the samples land on different bytes, and the
    // pixel comparison reports the change on its own. Only a frame whose
    // sampled bytes are identical at both sizes puts the size test on the
    // hook, and a frame of one repeated byte is that frame.
    memset(frame, 0x42, size);
    frame_changed(frame, size);
    check("a flat frame settles to unchanged", frame_changed(frame, size), 0);
    check("the same flat frame at a new size counts as changed",
        frame_changed(frame, smaller), 1);

    // Every read has to land inside the buffer it was given. Built with the
    // address sanitizer, so an overrun ends the run rather than passing.
    uint8_t *tight = (uint8_t *)malloc(smaller);
    if (tight == NULL) return 2;
    memcpy(tight, frame, smaller);
    frame_changed(tight, smaller);
    check("it reads only inside the frame it was given", 1, 1);
    free(tight);

    free(frame);
    return fails == 0 ? 0 : 1;
}
EOF

# The sanitizer is only used when it can be trusted here. asan-env.sh explains
# what makes it untrustworthy and why the answer is not simply to drop it.
if [ "$ASAN_USABLE" = 1 ]; then
    sanitize="-fsanitize=address,undefined"
else
    sanitize=""
    note "the sanitizer runs" "SKIP ($ASAN_NOTE)"
fi

if "$CXX" -std=c++11 -O1 -g $sanitize -I"$work" \
        -o "$work/t" "$work/main.cpp" 2>"$work/cc.log"; then
    asan_run "$work/t" || fails=$((fails + 1))
else
    note "the lifted frame_changed compiles" FAIL
    sed -n '1,15p' "$work/cc.log"
fi

echo
echo "===== the map happens where the read happens ====="

read_img=$(awk '/^int kvmv_read_img/,/^}/' "$VIS" | grep -v '^[[:space:]]*//')

maps=$(printf '%s\n' "$read_img" | grep -c 'mmf_vi_frame_map')
[ "$maps" = 1 ] \
    && note "kvmv_read_img maps a frame in exactly one place ($maps)" OK \
    || note "kvmv_read_img maps a frame in $maps places, wanted 1" FAIL

# The one map has to sit inside the detector branch. Anywhere above it and
# every frame pays for it again, which is the whole cost this removes.
detect=$(printf '%s\n' "$read_img" \
    | awk '/frame_undetact_count == kvmv_cfg.frame_detact/,/^            }$/')
printf '%s\n' "$detect" | grep -q 'mmf_vi_frame_map(native_vi_ch)' \
    && note "the map is inside the branch that samples" OK \
    || note "the map is inside the branch that samples" FAIL
printf '%s\n' "$detect" | grep -q 'frame_changed(pixels, native_len)' \
    && note "what it maps is what it samples" OK \
    || note "what it maps is what it samples" FAIL
printf '%s\n' "$detect" | grep -q 'mmf_vi_frame_release(native_vi_ch);' \
    && note "an unchanged frame is given back before returning" OK \
    || note "an unchanged frame is given back before returning" FAIL

# The copy path is gone entirely, so nothing can quietly reach it again.
printf '%s\n' "$read_img" | grep -q 'cam->read()' \
    && note "nothing takes a frame through cam->read() any more" FAIL \
    || note "nothing takes a frame through cam->read() any more" OK
# Anchored on the call, because frame_to_jpeg is this file's own native
# encoder and its name carries the same three words.
printf '%s\n' "$read_img" | grep -q -- '->to_jpeg(' \
    && note "nothing encodes through Image::to_jpeg any more" FAIL \
    || note "nothing encodes through Image::to_jpeg any more" OK
printf '%s\n' "$read_img" | grep -qE 'image::Image|delete img' \
    && note "the read path holds no image::Image" FAIL \
    || note "the read path holds no image::Image" OK
grep -q '^bool jpg_dump' "$VIS" \
    && note "jpg_dump went with its only caller" FAIL \
    || note "jpg_dump went with its only caller" OK

# Falling through the encode types while holding a frame takes a block out of a
# pool of two for good, which stalls capture rather than leaking heap.
printf '%s\n' "$read_img" | grep -q 'unknown encode type' \
    && note "an unknown encode type gives the frame back" OK \
    || note "an unknown encode type gives the frame back" FAIL

echo
echo "===== the map is exported deliberately ====="

map=$(body "$MMF" mmf_vi_frame_map)
printf '%s\n' "$map" | grep -q '_mmf_map_vi_frame(ch)' \
    && note "the exported map delegates to the internal one" OK \
    || note "the exported map delegates to the internal one" FAIL
grep -q 'void \*mmf_vi_frame_map(int ch);' "$HPP" \
    && note "callers outside kvm_mmf.cpp can see it" OK \
    || note "callers outside kvm_mmf.cpp can see it" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
