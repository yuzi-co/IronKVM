#!/bin/sh
# Break each guard in the data provisioning on purpose, and fail if
# test-s01fs-provision.sh does not notice.
#
#   test-s01fs-mutation.sh
#
# A guard that stops guarding reads exactly like a guard that holds. Two suites
# in this repository had already rotted that way and kept reporting success, so
# the suites here are themselves tested.
#
# Every mutation below is a thing somebody would plausibly write. Four of them
# destroy the board's identity, and one of them makes a filesystem that nothing
# on the device can repair.
HERE=$(cd "$(dirname "$0")" && pwd)
S01="$HERE/../../../kvmapp/system/init.d/S01fs"
SUITE="$HERE/test-s01fs-provision.sh"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# caught <label> <mutated S01fs>
#
# A mutation that does not change the file is a sed that did not match, and it
# would show up as a passing case. Check the file changed before judging it.
caught() {
    if cmp -s "$S01" "$2"; then
        note "$1 (THE MUTATION DID NOT APPLY)" FAIL
        return
    fi
    if sh "$SUITE" "$2" > /dev/null 2>&1; then
        note "$1" FAIL
    else
        note "$1" OK
    fi
}

echo "===== every mutation is caught ====="

# 1. The filesystem test greps for TYPE, which busybox blkid never prints. This
#    reports a healthy data partition as empty and formats it, which destroys
#    the root password, the authorized_keys and the ssh host key.
sed 's|\[ -n "$("$BLKID" "$1" 2>/dev/null)" \]|"$BLKID" "$1" 2>/dev/null \| grep -q TYPE=|' \
    "$S01" > "$WORK/typegrep"
caught "a filesystem test that greps for TYPE" "$WORK/typegrep"

# 2. The filesystem test reads blkid's exit status, which is 0 either way.
sed 's|\[ -n "$("$BLKID" "$1" 2>/dev/null)" \]|"$BLKID" "$1" >/dev/null 2>\&1|' \
    "$S01" > "$WORK/blkidrc"
caught "a filesystem test that reads blkid's exit status" "$WORK/blkidrc"

# 3. The guard that stops a live data partition being formatted is removed.
sed '/^    if \[ -e "$dev" \] && has_filesystem "$dev"; then$/,+2d' \
    "$S01" > "$WORK/nofsguard"
caught "a provisioner that formats a partition it can already read" "$WORK/nofsguard"

# 4. The size agreement check is dropped, so a stale node is formatted at the
#    wrong size. There is no resize.exfat on this device to put it right.
sed '/if ! data_partition_ready "$dev"; then/,+3d' "$S01" > "$WORK/noready"
caught "a provisioner that formats a node the kernel has not caught up with" "$WORK/noready"

# 5. parted's exit status is trusted. It returns 1 on a busy disk after it has
#    succeeded, so this skips everything after it on every card that worked.
sed 's|"$PARTED" -s "$DISK" resizepart 4 100% > /dev/null 2>&1|"$PARTED" -s "$DISK" resizepart 4 100% > /dev/null 2>\&1 \|\| return 1|' \
    "$S01" > "$WORK/partedrc"
caught "a provisioner that trusts parted's exit status" "$WORK/partedrc"

# 6. The ntfs argument is dropped, so the partition gets type 83 and a card
#    pulled from the board is not recognised by Windows or macOS.
sed 's|mkpart logical ntfs "${start}s" 100%|mkpart logical "${start}s" 100%|' \
    "$S01" > "$WORK/nontfs"
caught "a mkpart with no filesystem argument" "$WORK/nontfs"

# 7. The explicit start is dropped and parted picks its own, 8192 sectors
#    further along, so this card's layout matches no other card's.
sed 's|mkpart logical ntfs "${start}s" 100%|mkpart logical ntfs 0% 100%|' \
    "$S01" > "$WORK/nostart"
caught "a mkpart with no explicit start sector" "$WORK/nostart"

# 8. The container gate is dropped, so sector 0 is rewritten on every boot for
#    no gain. That is the exact wear the resize guard below it was written for.
sed '/if ! container_reaches_end; then/,+1d' "$S01" > "$WORK/nogate"
caught "a resize with no gate, which rewrites sector 0 every boot" "$WORK/nogate"

echo
echo "===== the unmutated script still passes its own suite ====="
if sh "$SUITE" "$S01" > /dev/null 2>&1; then
    note "the shipped S01fs passes its own suite" OK
else
    note "the shipped S01fs passes its own suite" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
