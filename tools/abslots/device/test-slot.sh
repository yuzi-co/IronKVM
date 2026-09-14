#!/bin/sh
# Check the slot tool's guards against loop files rather than a card.
#
#   test-slot.sh [path-to-slot]
#
# install writes a whole partition with dd. These guards are the only thing
# between a typo and a root filesystem overwritten while it is running, so they
# are checked here where a mistake costs nothing.
SLOT=${1:-$(dirname "$0")/slot}
[ -f "$SLOT" ] || { echo "usage: test-slot.sh <slot>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

mkdir -p "$WORK/boot"
cat > "$WORK/conf" <<CONF
SLOT_A=$WORK/a.img
SLOT_B=$WORK/b.img
RECOVERY=$WORK/r.img
DATA_DEV=$WORK/d.img
CONF

reset() {
    rm -f "$WORK"/*.img
    truncate -s 8M "$WORK/a.img" "$WORK/b.img" "$WORK/r.img"
    printf 'candidate\n' > "$WORK/cand.img"
    truncate -s 4M "$WORK/cand.img"
    rm -f "$WORK/boot"/*
    echo a > "$WORK/boot/slot"
}

# run [RUNNING=x] <args...>
run() {
    running=a
    case "$1" in
        RUNNING=*) running=${1#RUNNING=}; shift ;;
    esac
    SLOT_CONF="$WORK/conf" BOOT="$WORK/boot" RUNNING_SLOT="$running" \
        ZSTD="${ZSTD:-zstd}" sh "$SLOT" "$@" 2>&1
}

echo "===== install refuses what it must ====="
reset

run install a "$WORK/cand.img" >/dev/null 2>&1 \
    && note "installing onto the running slot is refused" FAIL \
    || note "installing onto the running slot is refused" OK

truncate -s 16M "$WORK/toobig.img"
run install b "$WORK/toobig.img" >/dev/null 2>&1 \
    && note "an image larger than the partition is refused" FAIL \
    || note "an image larger than the partition is refused" OK

run install b "$WORK/nosuch.img" >/dev/null 2>&1 \
    && note "a missing image is refused" FAIL \
    || note "a missing image is refused" OK

run install zzz "$WORK/cand.img" >/dev/null 2>&1 \
    && note "an unknown slot name is refused" FAIL \
    || note "an unknown slot name is refused" OK

echo
echo "===== install writes and verifies ====="
reset

if run install b "$WORK/cand.img" >"$WORK/inst.log" 2>&1; then
    note "installing to a free slot succeeds" OK
else
    note "installing to a free slot succeeds" FAIL
    sed 's/^/    /' "$WORK/inst.log" | head -5
fi

# The bytes of the image must actually be at the front of the target.
isize=$(wc -c < "$WORK/cand.img")
head -c "$isize" "$WORK/b.img" | sha256sum | cut -d' ' -f1 > "$WORK/got"
sha256sum "$WORK/cand.img" | cut -d' ' -f1 > "$WORK/want"
cmp -s "$WORK/got" "$WORK/want" \
    && note "the installed bytes match the image" OK \
    || note "the installed bytes match the image" FAIL

echo
echo "===== install accepts a compressed image ====="
# A slot image is mostly zeroes, so the compressed form is a fifteenth of the
# size. /data keeps every image the board has been given, and it is also what
# has to be copied to the board in the first place.
#
# The zstd here is a stub, because the workstation that runs this suite need not
# have zstd and the device certainly does. It is exactly reversible and it fails
# a file that says CORRUPT, which is all these cases need: what is being checked
# is that slot decompresses to the device rather than writing the compressed
# bytes, and that it refuses an image whose own checksum does not hold. The last
# case below repeats the first with the real thing wherever it is installed.
reset
cat > "$WORK/zstd-stub" <<'STUB'
#!/bin/sh
# -t <file>   succeed unless the file says CORRUPT
# -dc <file>  print the file without its first line
case "$1" in
    -t)  grep -q CORRUPT "$2" && exit 1; exit 0 ;;
    -dc) tail -n +2 "$2" ;;
    *)   exit 1 ;;
esac
STUB
chmod +x "$WORK/zstd-stub"

{ echo "ZSTUB"; cat "$WORK/cand.img"; } > "$WORK/cand.img.zst"

if ZSTD="$WORK/zstd-stub" run install b "$WORK/cand.img.zst" >"$WORK/zinst.log" 2>&1; then
    note "installing a .zst image succeeds" OK
else
    note "installing a .zst image succeeds" FAIL
    sed 's/^/    /' "$WORK/zinst.log" | head -5
fi

# The decompressed bytes are what must be on the device, not the file's.
isize=$(wc -c < "$WORK/cand.img")
head -c "$isize" "$WORK/b.img" | sha256sum | cut -d' ' -f1 > "$WORK/got"
sha256sum "$WORK/cand.img" | cut -d' ' -f1 > "$WORK/want"
cmp -s "$WORK/got" "$WORK/want" \
    && note "the slot holds the decompressed image, not the file" OK \
    || note "the slot holds the decompressed image, not the file" FAIL

# The size that is checked against the partition has to be the decompressed
# size. A 15 MiB file that decompresses to 4 GiB must not be accepted for a
# 2 GiB slot because the file fits.
reset
{ echo "ZSTUB"; dd if=/dev/zero bs=1M count=16 2>/dev/null; } > "$WORK/big.img.zst"
ZSTD="$WORK/zstd-stub" run install b "$WORK/big.img.zst" >/dev/null 2>&1 \
    && note "a .zst whose contents overflow the slot is refused" FAIL \
    || note "a .zst whose contents overflow the slot is refused" OK

# A truncated or damaged stream. Without the integrity test every later pass
# would stop in the same place and agree with itself, so the install would
# report success over half an image.
reset
{ echo "ZSTUB"; echo "CORRUPT"; cat "$WORK/cand.img"; } > "$WORK/bad.img.zst"
ZSTD="$WORK/zstd-stub" run install b "$WORK/bad.img.zst" >"$WORK/bad.log" 2>&1 \
    && note "a .zst that fails its own checksum is refused" FAIL \
    || note "a .zst that fails its own checksum is refused" OK
grep -q 'integrity' "$WORK/bad.log" \
    && note "the refusal names the integrity check" OK \
    || note "the refusal said: $(head -1 "$WORK/bad.log")" FAIL

# Nothing may have been written before the integrity test ran.
[ "$(head -c 16 "$WORK/b.img" | tr -d '\0')" = "" ] \
    && note "a bad .zst leaves the target untouched" OK \
    || note "a bad .zst wrote to the target anyway" FAIL

# The integrity test covers a file that was already damaged when it arrived. It
# cannot cover a decompression that succeeds twice and then fails while the
# write is running, which is what a dying card or a truncated read looks like.
# The read back is the only thing standing there: the size and the hash were
# taken from a complete decompression, so a short write leaves the tail of the
# target holding whatever was there before.
reset
cat > "$WORK/zstd-flaky" <<'STUB'
#!/bin/sh
# Truthful about -t, and truthful for the first two -dc calls, which are the
# size and the hash. The third is the one that writes.
case "$1" in
    -t)  exit 0 ;;
    -dc) n=$(cat "$COUNT" 2>/dev/null || echo 0); n=$((n + 1)); echo "$n" > "$COUNT"
         if [ "$n" -ge 3 ]; then tail -n +2 "$2" | head -c 1024; else tail -n +2 "$2"; fi ;;
    *)   exit 1 ;;
esac
STUB
chmod +x "$WORK/zstd-flaky"
echo 0 > "$WORK/zcount"

COUNT="$WORK/zcount" ZSTD="$WORK/zstd-flaky" run install b "$WORK/cand.img.zst" \
    >"$WORK/flaky.log" 2>&1 \
    && note "a decompression that fails mid-write is caught" FAIL \
    || note "a decompression that fails mid-write is caught" OK
grep -q 'read-back mismatch' "$WORK/flaky.log" \
    && note "it is the read back that catches it" OK \
    || note "it failed for another reason: $(head -1 "$WORK/flaky.log")" FAIL

# A board whose base does not carry zstd has to say so rather than write
# something. The Sipeed rootfs has no zstd at all.
reset
ZSTD="$WORK/absent-zstd" run install b "$WORK/cand.img.zst" >"$WORK/noz.log" 2>&1 \
    && note "a .zst with no zstd on the board is refused" FAIL \
    || note "a .zst with no zstd on the board is refused" OK
grep -q 'no zstd' "$WORK/noz.log" \
    && note "the refusal says the board has no zstd" OK \
    || note "the refusal said: $(head -1 "$WORK/noz.log")" FAIL

# An uncompressed image must still install with no zstd anywhere.
reset
ZSTD="$WORK/absent-zstd" run install b "$WORK/cand.img" >/dev/null 2>&1 \
    && note "a plain image still installs with no zstd" OK \
    || note "a plain image still installs with no zstd" FAIL

# And the whole path again with real zstd, wherever it is installed. The stub
# above proves the branching; this proves the command line.
reset
if command -v zstd > /dev/null 2>&1; then
    zstd -q -f -o "$WORK/real.img.zst" "$WORK/cand.img" 2>/dev/null
    if run install b "$WORK/real.img.zst" >"$WORK/real.log" 2>&1; then
        isize=$(wc -c < "$WORK/cand.img")
        head -c "$isize" "$WORK/b.img" | sha256sum | cut -d' ' -f1 > "$WORK/got"
        sha256sum "$WORK/cand.img" | cut -d' ' -f1 > "$WORK/want"
        cmp -s "$WORK/got" "$WORK/want" \
            && note "a real zstd image installs byte for byte" OK \
            || note "a real zstd image installed the wrong bytes" FAIL
    else
        note "a real zstd image installs byte for byte" FAIL
        sed 's/^/    /' "$WORK/real.log" | head -5
    fi
else
    note "a real zstd image installs byte for byte (no zstd here)" SKIP
fi

echo
echo "===== try and confirm move the right markers ====="
reset

run try b >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot.try" 2>/dev/null)" = b ] \
    && note "try writes slot.try" OK || note "try writes slot.try" FAIL
[ "$(cat "$WORK/boot/slot")" = a ] \
    && note "try leaves the trusted marker alone" OK || note "try leaves the trusted marker alone" FAIL

run RUNNING=b confirm >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot")" = b ] \
    && note "confirm promotes the running slot" OK || note "confirm promotes the running slot" FAIL
[ "$(cat "$WORK/boot/slot.prev")" = a ] \
    && note "confirm records the previous slot" OK || note "confirm records the previous slot" FAIL

# Confirming twice must not lose the real previous slot by recording the
# current one over it.
run RUNNING=b confirm >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot.prev")" = a ] \
    && note "confirming twice keeps the real previous slot" OK \
    || note "confirming twice overwrote prev with $(cat "$WORK/boot/slot.prev")" FAIL

run revert >/dev/null 2>&1
[ "$(cat "$WORK/boot/slot")" = a ] \
    && note "revert restores the previous slot" OK || note "revert restores the previous slot" FAIL

echo
echo "===== recovery does not promote itself ====="
reset

run RUNNING=recovery confirm >/dev/null 2>&1 \
    && note "confirm from recovery is refused" FAIL \
    || note "confirm from recovery is refused" OK

run recovery >/dev/null 2>&1
[ -e "$WORK/boot/recovery" ] \
    && note "recovery writes its marker" OK || note "recovery writes its marker" FAIL

echo
echo "===== status reports the markers ====="
reset
echo b > "$WORK/boot/slot"
echo a > "$WORK/boot/slot.try"

out=$(run RUNNING=b status)
echo "$out" | grep -q "running *b"  && note "status reports the running slot" OK  || note "status reports the running slot" FAIL
echo "$out" | grep -q "trusted *b"  && note "status reports the trusted slot" OK  || note "status reports the trusted slot" FAIL
echo "$out" | grep -q "trial *a"    && note "status reports a pending trial" OK   || note "status reports a pending trial" FAIL

echo
echo "===== the script parses ====="
sh -n "$SLOT" 2>/dev/null && note "sh -n accepts the tool" OK || note "sh -n accepts the tool" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
