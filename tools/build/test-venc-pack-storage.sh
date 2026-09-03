#!/bin/sh
# Check how kvm_mmf.cpp holds the pack descriptors of an encoded frame.
#
#   test-venc-pack-storage.sh [path/to/kvm_mmf.cpp]
#
# Not destructive: the source is only read.
#
# These functions cannot be lifted out and run the way the frame buffer pool
# can. Every one of them calls into CVI_VENC, and there is no stand-in for that
# here. What can still be held in place is the shape of the code, and the shape
# is where the two defects were.
#
# The serious one is an ordering. CVI_VENC_GetStream writes one descriptor for
# each pack the channel holds, into the array it is given. The array holds
# MMF_VENC_MAX_PACKS of them. The count was read from the channel and then
# tested after the call had already written, so a frame that arrived in more
# packs than that overran the buffer and the test reported it afterwards.
#
# The cheap one is a per frame malloc and free of that array, on the video path,
# with four error returns that forgot the free.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SRC=${1:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# The body of one function, with comment lines dropped. The comments here
# name the very calls these cases look for, so a match against them would
# make the checks below pass on prose.
fn() { sed -n "/^int $1(/,/^}/p" "$SRC" | grep -v '^[[:space:]]*//'; }

echo "===== the descriptors are not allocated per frame ====="

grep -q 'malloc(sizeof(VENC_PACK_S)' "$SRC" \
    && note "no pack array is malloc-ed" FAIL \
    || note "no pack array is malloc-ed" OK

grep -qE 'free\((venc_stream->pstPack|priv\.enc_jpeg_frame\.pstPack|\*_pp)' "$SRC" \
    && note "no pack array is freed" FAIL \
    || note "no pack array is freed" OK

grep -q 'static VENC_PACK_S venc_pack_storage\[MMF_VENC_MAX_CHN\]\[MMF_VENC_MAX_PACKS\];' "$SRC" \
    && note "the H.264 storage is one array per channel" OK \
    || note "the H.264 storage is one array per channel" FAIL

grep -q 'static VENC_PACK_S jpeg_pack_storage;' "$SRC" \
    && note "the JPEG storage is its own single descriptor" OK \
    || note "the JPEG storage is its own single descriptor" FAIL

echo
echo "===== the pack count is tested before the write ====="

# Read mmf_venc_pop and record the line of the bound test and of the call that
# writes the array. The first has to come first.
body=$(fn mmf_venc_pop)
bound=$(printf '%s\n' "$body" | grep -n 'u32CurPacks > MMF_VENC_MAX_PACKS' | head -1 | cut -d: -f1)
write=$(printf '%s\n' "$body" | grep -n 'CVI_VENC_GetStream' | head -1 | cut -d: -f1)

if [ -n "$bound" ] && [ -n "$write" ] && [ "$bound" -lt "$write" ]; then
    note "mmf_venc_pop bounds the count before GetStream" OK
else
    note "mmf_venc_pop bounds the count before GetStream (bound=${bound:-none} write=${write:-none})" FAIL
fi

printf '%s\n' "$body" | grep -q 'venc_stream->pstPack = venc_pack_storage\[ch\];' \
    && note "mmf_venc_pop points at the storage for its channel" OK \
    || note "mmf_venc_pop points at the storage for its channel" FAIL

# The literal 8 was the old bound, written in three places and matching the
# array size only by coincidence.
printf '%s\n' "$body" | grep -qE 'count > 8|VENC_PACK_S\) \* 8' \
    && note "no literal pack bound is left in mmf_venc_pop" FAIL \
    || note "no literal pack bound is left in mmf_venc_pop" OK

echo
echo "===== the JPEG pop reads the pack it was given ====="

body=$(fn mmf_enc_jpg_pop)

printf '%s\n' "$body" | grep -q 'u32PackCount != 1' \
    && note "it refuses a frame that is not one pack" OK \
    || note "it refuses a frame that is not one pack" FAIL

printf '%s\n' "$body" | grep -q 'pack->pu8Addr + pack->u32Offset' \
    && note "the data pointer skips the pack offset" OK \
    || note "the data pointer skips the pack offset" FAIL

printf '%s\n' "$body" | grep -q 'pack->u32Len - pack->u32Offset' \
    && note "the size excludes the pack offset" OK \
    || note "the size excludes the pack offset" FAIL

# QueryStatus answers zero while the encoder is still working on the frame that
# was submitted a moment ago, which turned latency into a dropped frame.
printf '%s\n' "$body" | grep -q 'CVI_VENC_QueryStatus' \
    && note "it does not ask QueryStatus before waiting" FAIL \
    || note "it does not ask QueryStatus before waiting" OK

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
