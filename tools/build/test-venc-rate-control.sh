#!/bin/sh
# Check which rate control kvm_mmf.cpp gives the H.264 and H.265 encoder.
#
#   test-venc-rate-control.sh [path/to/kvm_mmf.cpp]
#
# Not destructive: the source is only read. The functions call CVI_VENC and
# there is no stand-in for it here, so the shape of the code is what is held.
#
# The encoder ran in constant bit rate. It spent the whole bit rate setting
# whether the screen moved or not: on 2026-09-23 two 40 second windows of an
# unchanged screen each came to exactly 13,158,312 bytes, 329 KB/s at the
# 2000 kbit/s default. Every byte of that is encrypted and sent by the server,
# and on the WebRTC path each 1200 bytes is a packet with its own SRTP seal and
# its own UDP send, about 12% of the core between them.
#
# Variable bit rate with the setting as its ceiling spends the same when the
# picture moves and far less when it does not.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SRC=${1:-$ROOT/support/sg2002/additional/kvm_mmf/src/kvm_mmf.cpp}
[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# Code only. The comments here name the modes these cases look for.
code=$(grep -v '^[[:space:]]*//' "$SRC" | grep -v '^[[:space:]]*/\?\*')

echo "===== the H.26x encoder runs in variable bit rate ====="
for mode in H264VBR H265VBR; do
    echo "$code" | grep -q "enRcMode = VENC_RC_MODE_$mode;" \
        && note "the $mode mode is selected" OK \
        || note "the $mode mode is selected" FAIL
done
for mode in H264CBR H265CBR; do
    echo "$code" | grep -q "VENC_RC_MODE_$mode" \
        && note "no $mode remains" FAIL \
        || note "no $mode remains" OK
done

echo "===== the bit rate setting is the ceiling ====="
for codec in H264 H265; do
    echo "$code" | grep -q "st${codec}Vbr.u32MaxBitRate = cfg->bitrate;" \
        && note "$codec caps the bit rate at the setting" OK \
        || note "$codec caps the bit rate at the setting" FAIL
done

echo "===== the quality limits apply to the mode that runs ====="
# CVI_VENC_SetRcParam reads the parameter block of the mode in use. Limits
# written to the CBR block while VBR runs are ignored without an error.
for codec in H264 H265; do
    echo "$code" | grep -q "stParam${codec}Vbr" \
        && note "$codec limits go to the VBR parameter block" OK \
        || note "$codec limits go to the VBR parameter block" FAIL
    echo "$code" | grep -q "stParam${codec}Cbr" \
        && note "no $codec limits go to the CBR block" FAIL \
        || note "no $codec limits go to the CBR block" OK
done

[ "$fails" -eq 0 ] || { echo "$fails case(s) FAILED"; exit 1; }
exit 0
