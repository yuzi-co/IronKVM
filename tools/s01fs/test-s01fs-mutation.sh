#!/bin/sh
# Break each guard in S01fs on purpose, and fail if the suite that owns it does
# not notice.
#
#   test-s01fs-mutation.sh
#
# A guard that stops guarding reads exactly like a guard that holds. Two suites
# in this repository had already rotted that way and kept reporting success, so
# the suites here are themselves tested.
#
# Every mutation below is a thing somebody would plausibly write. Several of
# them destroy the board's identity, one makes a filesystem that nothing on the
# device can repair, and one formats the root slot the board is running from.
HERE=$(cd "$(dirname "$0")" && pwd)
S01="$HERE/../../kvmapp/system/init.d/S01fs"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# caught <label> <suite> <mutated S01fs>
#
# A mutation that does not change the file is a sed that did not match, and it
# would show up as a passing case. A mutation that does not parse is a sed that
# broke the script rather than its behaviour, and every suite fails it for the
# wrong reason. Both are checked before the suite is believed.
#
# The suite is named per mutation, because the guards live in different blocks
# and each block has the suite that owns it.
caught() {
    if cmp -s "$S01" "$3"; then
        note "$1 (THE MUTATION DID NOT APPLY)" FAIL
        return
    fi
    if ! sh -n "$3" 2>/dev/null; then
        note "$1 (THE MUTATION BROKE THE SYNTAX)" FAIL
        return
    fi
    if sh "$HERE/$2" "$3" > /dev/null 2>&1; then
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
caught "a filesystem test that greps for TYPE" test-s01fs-provision.sh "$WORK/typegrep"

# 2. The filesystem test reads blkid's exit status, which is 0 either way.
sed 's|\[ -n "$("$BLKID" "$1" 2>/dev/null)" \]|"$BLKID" "$1" >/dev/null 2>\&1|' \
    "$S01" > "$WORK/blkidrc"
caught "a filesystem test that reads blkid's exit status" test-s01fs-provision.sh "$WORK/blkidrc"

# 3. The guard that stops a live data partition being formatted is removed.
sed '/^    if \[ -e "$dev" \] && has_filesystem "$dev"; then$/,+2d' \
    "$S01" > "$WORK/nofsguard"
caught "a provisioner that formats a partition it can already read" test-s01fs-provision.sh "$WORK/nofsguard"

# 4. The size agreement check is dropped, so a stale node is formatted at the
#    wrong size. There is no resize.exfat on this device to put it right.
sed '/if ! data_partition_ready "$dev"; then/,+3d' "$S01" > "$WORK/noready"
caught "a provisioner that formats a node the kernel has not caught up with" test-s01fs-provision.sh "$WORK/noready"

# 5. parted's exit status is trusted. It returns 1 on a busy disk after it has
#    succeeded, so this skips everything after it on every card that worked.
sed 's|"$PARTED" -s "$DISK" resizepart 4 100% > /dev/null 2>&1|"$PARTED" -s "$DISK" resizepart 4 100% > /dev/null 2>\&1 \|\| return 1|' \
    "$S01" > "$WORK/partedrc"
caught "a provisioner that trusts parted's exit status" test-s01fs-provision.sh "$WORK/partedrc"

# 6. The ntfs argument is dropped, so the partition gets type 83 and a card
#    pulled from the board is not recognised by Windows or macOS.
sed 's|mkpart logical ntfs "${start}s" 100%|mkpart logical "${start}s" 100%|' \
    "$S01" > "$WORK/nontfs"
caught "a mkpart with no filesystem argument" test-s01fs-provision.sh "$WORK/nontfs"

# 7. The explicit start is dropped and parted picks its own, 8192 sectors
#    further along, so this card's layout matches no other card's.
sed 's|mkpart logical ntfs "${start}s" 100%|mkpart logical ntfs 0% 100%|' \
    "$S01" > "$WORK/nostart"
caught "a mkpart with no explicit start sector" test-s01fs-provision.sh "$WORK/nostart"

# 8. The container gate is dropped, so sector 0 is rewritten on every boot for
#    no gain. That is the exact wear the resize guard below it was written for.
#
#    The outer gate only. Its indentation is what tells it from the test that
#    follows the resize, and a sed that took both would change two behaviours
#    at once. This used to delete the gate and its echo, which left the fi
#    below unmatched: the suite then failed on the syntax rather than on the
#    missing gate, so the case passed for the wrong reason.
sed 's|^    if ! container_reaches_end; then$|    if true; then|' "$S01" > "$WORK/nogate"
caught "a resize with no gate, which rewrites sector 0 every boot" test-s01fs-provision.sh "$WORK/nogate"

# 9. The data device stops falling back to the partition table, so a board whose
#    description declares no DATA mounts no /data at all. That is every board on
#    Sipeed's firmware.
sed 's|^    data_device_from_table$|    return 0|' "$S01" > "$WORK/nofallback"
caught "a data device that never looks at the card" test-s01fs-datadev.sh "$WORK/nofallback"

# 10. The search takes the first partition that holds any filesystem, which on
#     both layouts is the root filesystem.
sed "s|^            \\*'LABEL=\"data\"'\\*)|            *)|" "$S01" > "$WORK/anyfs"
caught "a search that takes any filesystem for the data one" test-s01fs-datadev.sh "$WORK/anyfs"

# 11. The search stops skipping the partitions the description keeps, so a
#     declared slot can be returned as the data device and then mounted.
sed 's|^        is_reserved_device "$candidate" && continue$|        false \&\& continue|' \
    "$S01" > "$WORK/noskip"
caught "a search that can return a declared slot" test-s01fs-datadev.sh "$WORK/noskip"

# 12. A board with no data device says nothing on the console, which is how the
#     1.0.0 card image shipped with no data partition and nobody noticed.
sed 's|.*no data device: none declared.*|                :|' "$S01" > "$WORK/silent"
caught "a board with no data device that says nothing" test-s01fs-datadev.sh "$WORK/silent"

# 13. The stock provisioning stops asking whether the device is a slot. On an
#     A/B card that formats the root filesystem the board is running from.
sed 's|^    if is_reserved_device "$DISK0_DEV"; then$|    if false; then|' \
    "$S01" > "$WORK/noslotguard"
caught "a stock provisioner that formats a declared slot" test-s01fs-disk0.sh "$WORK/noslotguard"

# 14. It stops asking whether the device already holds a filesystem, which is
#     the only guard left where the description declares no layout.
sed 's|^    if has_filesystem "$DISK0_DEV"; then$|    if false; then|' \
    "$S01" > "$WORK/nofsguard0"
caught "a stock provisioner that formats a filesystem it can read" test-s01fs-disk0.sh "$WORK/nofsguard0"

# 15. The exception is widened: any in-flight marker excuses the format, whoever
#     wrote it and whatever device it was about.
sed 's|^    \[ "$(cat "$DISK0_PENDING" 2>/dev/null)" = "$DISK0_DEV" \]$|    return 0|' \
    "$S01" > "$WORK/anymarker"
caught "a retry that any in-flight marker excuses" test-s01fs-disk0.sh "$WORK/anymarker"

# 16. The exception is widened the other way: the marker names the device, so
#     whatever is on it is formatted. A root slot is ext4.
sed 's|^        if holds_foreign_filesystem; then$|        if false; then|' \
    "$S01" > "$WORK/anyresume"
caught "a retry that formats whatever the device holds" test-s01fs-disk0.sh "$WORK/anyresume"

# 17. The marker goes back to being an empty file, so it can no longer say which
#     device the interrupted format was on.
sed 's|^        printf .%s.*"$DISK0_PENDING"$|        touch "$DISK0_PENDING"|' \
    "$S01" > "$WORK/bareflag"
caught "an in-flight marker that records no device" test-s01fs-disk0.sh "$WORK/bareflag"

echo
echo "===== the unmutated script still passes every suite above ====="
for suite in test-s01fs-provision.sh test-s01fs-datadev.sh test-s01fs-disk0.sh; do
    if sh "$HERE/$suite" "$S01" > /dev/null 2>&1; then
        note "the shipped S01fs passes $suite" OK
    else
        note "the shipped S01fs passes $suite" FAIL
    fi
done

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
