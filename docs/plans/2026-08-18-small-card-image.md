# Small Card Image and First-Boot Data Partition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a card image that installs on an 8 GB microSD card, and make the data partition on the first boot at whatever size the card allows.

**Architecture:** The shipped partition table drops its data partition and ends after the recovery slot, so the image needs 10543104 sectors instead of 60506112. `S01fs` grows the extended container to the end of the card, makes the data partition at a derived start sector, and makes an exfat filesystem on it. Every step is gated on a condition read from the partition table, so no marker file is used and no step repeats.

**Tech Stack:** POSIX shell (busybox ash on the device), GNU parted 3.6, exfatprogs 1.2.2, sfdisk on the build host, Docker for the test containers.

**Spec:** `docs/specs/2026-08-18-small-card-image-design.md`

## Global Constraints

- The device runs **busybox 1.36.1**. No `local`, no bashisms. Every device script must pass `sh -n` under busybox.
- **`parted` exits 1 on a busy disk after it has succeeded.** Never read its exit status. Read the result back.
- **busybox `blkid` never prints `TYPE=`**, for any filesystem, and it exits 0 for a partition with no filesystem. The only valid test is whether it printed anything.
- **`resize.exfat` does not exist on the device.** An exfat filesystem is made once, at the size it keeps.
- `mkpart` requires `resizepart 4 100%` to run first, and requires the `ntfs` filesystem argument to produce partition type `7`.
- The data partition starts at `p5_start + p5_size + 8192`, which is sector **10543104**. This is the value the fixed table used, so cards made either way have one layout.
- **No em dashes** in any file: code, comment, commit message or document. Use a colon, a comma or a full stop.
- **No assistant attribution in commits.** No `Co-Authored-By` and no `Claude-Session` trailer.
- Commit on the current branch, `fork/integration`. Do not rebase and do not force push.
- Target release **1.0.1**. The version is an argument to `release.sh`; no file in the tree carries it.

## File Structure

| Path | Responsibility | Change |
| ---- | -------------- | ------ |
| `tools/abslots/data-start.sh` | Derive the data partition start from the table. The single source for three consumers. | create |
| `tools/abslots/test-data-start.sh` | Tests for the above. | create |
| `tools/abslots/partition.sfdisk` | The shipped table. Loses p6, p4 shrinks to hold only p5. | modify |
| `tools/abslots/test-partition.sh` | Table tests. Applies to a small disk, no p6, p4 ends at p5. | modify |
| `tools/abslots/build-card.sh` | Assembles the card image. No truncation step any more. | modify |
| `tools/abslots/test-build-card.sh` | Build tests. Test table loses p6. | modify |
| `tools/abslots/build-image.sh` | Writes `DATA_START` into `/etc/nanokvm-slots.conf`. | modify |
| `tools/abslots/test-build-image.sh` | Asserts the new conf line. | modify |
| `kvmapp/system/init.d/S01fs` | Reads the table, makes the partition and the filesystem, reports a failed mount. | modify |
| `tools/abslots/device/test-s01fs-provision.sh` | Unit tests for the two new blocks. | create |
| `tools/abslots/device/test-s01fs-datadev.sh` | Existing tests. One call-site regex changes. | modify |
| `tools/abslots/device/test-s01fs-mutation.sh` | Breaks each new guard and fails if the suite does not notice. | create |
| `tools/abslots/test-provision-integration.sh` | Real parted and real mkfs.exfat against a real built image. | create |
| `tools/release/Dockerfile` | Gains `parted` and `exfatprogs`. | modify |
| `README.md`, `docs/CHANGES-FROM-OFFICIAL.md`, `tools/release/release.sh`, `docs/plans/2026-08-16-ironkvm-1.0.md` | Correct the claims about the data partition and the card requirement. | modify |

### How to run the shell tests

Nothing here has a central runner. Each suite is invoked directly. The ones that
need Linux tools run in a container:

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-partition.sh
```

The device-script suites need nothing but a POSIX shell and run anywhere:

```shell
sh tools/abslots/device/test-s01fs-datadev.sh
```

---

### Task 1: The derived start sector

The number 10543104 must exist in exactly one place. Three consumers need it and
none may keep a copy.

**Files:**
- Create: `tools/abslots/data-start.sh`
- Create: `tools/abslots/test-data-start.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: `data-start.sh [table]` prints one integer on stdout and exits 0, or
  prints a message on stderr and exits 1. Task 3 and Task 4 call it.

- [ ] **Step 1: Write the failing test**

Create `tools/abslots/test-data-start.sh`:

```sh
#!/bin/sh
# Tests for data-start.sh.
#
# The start sector of the data partition is derived rather than declared, so the
# derivation is the thing that has to be right. A wrong number here puts the
# data partition somewhere no other card has it.
HERE=$(cd "$(dirname "$0")" && pwd)
SCRIPT="$HERE/data-start.sh"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== the shipped table gives the sector every existing card uses ====="

# 10543104 is where the data partition sat in the fixed table this replaced. A
# card made the new way and a card made the old way must have one layout, so
# this number is a compatibility assertion and not an arithmetic one.
got=$(sh "$SCRIPT" "$HERE/partition.sfdisk" 2>&1)
[ "$got" = 10543104 ] \
    && note "the shipped table gives 10543104" OK \
    || note "the shipped table gives '$got', want 10543104" FAIL

echo
echo "===== the rule is end of the recovery slot plus 8192 ====="

cat > "$WORK/small.sfdisk" <<'EOF'
label: dos
unit: sectors

1 : start=1,     size=32768, type=c, bootable
2 : start=40960, size=16384, type=83
3 : start=57344, size=16384, type=83
4 : start=73728, size=18432, type=5
5 : start=75776, size=16384, type=83
EOF
got=$(sh "$SCRIPT" "$WORK/small.sfdisk" 2>&1)
[ "$got" = 100352 ] \
    && note "75776 + 16384 + 8192 gives 100352" OK \
    || note "a scaled table gives '$got', want 100352" FAIL

echo
echo "===== a table that still declares a data partition is refused ====="

# A table with a p6 is a table from before this change. Printing a number that
# contradicts it would put the filesystem somewhere the table does not describe.
cat "$WORK/small.sfdisk" > "$WORK/withp6.sfdisk"
printf '6 : start=100352, size=34816, type=7\n' >> "$WORK/withp6.sfdisk"
if sh "$SCRIPT" "$WORK/withp6.sfdisk" > /dev/null 2>&1; then
    note "a table declaring p6 is refused" FAIL
else
    note "a table declaring p6 is refused" OK
fi

echo
echo "===== a table with no recovery slot is refused ====="

cat > "$WORK/nop5.sfdisk" <<'EOF'
label: dos
unit: sectors

1 : start=1,     size=32768, type=c, bootable
2 : start=40960, size=16384, type=83
EOF
if sh "$SCRIPT" "$WORK/nop5.sfdisk" > /dev/null 2>&1; then
    note "a table with no p5 is refused" FAIL
else
    note "a table with no p5 is refused" OK
fi

echo
echo "===== a missing table is refused ====="
if sh "$SCRIPT" "$WORK/absent.sfdisk" > /dev/null 2>&1; then
    note "a missing table is refused" FAIL
else
    note "a missing table is refused" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
```

- [ ] **Step 2: Run it and watch every case fail**

```shell
sh tools/abslots/test-data-start.sh
```

Expected: every case FAIL, because `data-start.sh` does not exist. The first case
reports the shell's "No such file or directory" as the value it got.

- [ ] **Step 3: Write data-start.sh**

Create `tools/abslots/data-start.sh`:

```sh
#!/bin/sh
# Print the sector the data partition starts at, derived from the table.
#
#   data-start.sh [path-to-partition.sfdisk]
#
# The shipped table declares p1 to p5 and stops. The data partition is made on
# the first boot of a card, at whatever size that card allows, so its start is a
# derived number and not a table entry. Three places need it and none of them may
# keep its own copy: build-card.sh sizes the image with it, build-image.sh writes
# it into /etc/nanokvm-slots.conf, and S01fs reads it from there.
#
#   start = the end of the recovery slot + 8192
#
# 8192 sectors is 4 MiB. It is the gap the table already leaves ahead of a
# logical partition for its EBR, and 4 MiB is the erase block size that matters
# on this card. The rule gives 10543104, which is where the data partition sat in
# the fixed table this replaced. A card made either way has one layout.
set -eu

TABLE=${1:-$(dirname "$0")/partition.sfdisk}
[ -f "$TABLE" ] || { echo "no such table: $TABLE" >&2; exit 1; }

field() {
    sed -n "s/^$1 *: *.*$2=\([0-9]*\).*/\1/p" "$TABLE" | head -1
}

# A table that still declares a data partition is a table from before this
# change. Refuse it rather than print a number that contradicts it.
if [ -n "$(field 6 start)" ]; then
    echo "$TABLE declares partition 6, and the data partition is made on the device" >&2
    exit 1
fi

START=$(field 5 start)
SIZE=$(field 5 size)
if [ -z "$START" ] || [ -z "$SIZE" ]; then
    echo "$TABLE does not describe partition 5" >&2
    exit 1
fi

echo $((START + SIZE + 8192))
```

- [ ] **Step 4: Run the test again**

```shell
sh tools/abslots/test-data-start.sh
```

Expected: the four scaled and refusal cases pass. **The first case still fails**,
because `partition.sfdisk` still declares p6 and the script refuses it. That is
correct: Task 2 removes p6 and the case turns green there.

- [ ] **Step 5: Commit**

```shell
chmod +x tools/abslots/data-start.sh
git add tools/abslots/data-start.sh tools/abslots/test-data-start.sh
git commit -F - <<'MSG'
Derive the data partition start from the table

The data partition is made on the device now, so its start sector is a
derived number rather than a table entry. Three places need it: the card
build sizes the image with it, the image build writes it into the slot
configuration, and S01fs reads it from there. Two copies of a geometry
drift, and the copy that drifts is the one nobody runs.

The rule is the end of the recovery slot plus the 4 MiB the table already
leaves ahead of a logical partition for its EBR. It gives 10543104, which
is where the data partition sat in the fixed table this replaces, so a
card made either way has one layout.

A table that still declares a data partition is refused rather than
answered, because the answer would contradict it.
MSG
```

---

### Task 2: The shipped table drops the data partition

**Files:**
- Modify: `tools/abslots/partition.sfdisk`
- Modify: `tools/abslots/test-partition.sh`

**Interfaces:**
- Consumes: `data-start.sh` from Task 1.
- Produces: a table whose last entry is partition 5, ending at sector 10534911,
  with the extended container ending at the same sector. Task 3 reads it.

- [ ] **Step 1: Write the failing tests**

In `tools/abslots/test-partition.sh`, replace the disk size constant:

```sh
# The smallest card the image is meant for, measured: an 8 GB card that reports
# 15523840 sectors, 7.40 GiB. The table has to fit inside it with room left for
# the data partition that the first boot makes.
DISK_SECTORS=15523840

# The last sector the table itself needs. A card of exactly this many sectors
# takes the image and leaves no room for data, which is the floor and not a
# target.
TABLE_SECTORS=10543104
```

Replace the alignment loop so it no longer expects a partition 6:

```sh
for p in 2 3 5; do
```

Replace the end-of-table check with these three cases:

```sh
echo
echo "===== the table stops after the recovery slot ====="

# The data partition is made on the first boot, at whatever size the card
# allows. A table that declares it needs a card of at least the sector it ends
# at, and that is what limited the 1.0.0 image to a 32 GB card.
[ -z "$(field 6 start)" ] \
    && note "no data partition is declared" OK \
    || note "a data partition is declared at $(field 6 start)" FAIL

# The container must end exactly where the recovery slot ends. Any further and
# the image carries sectors nothing uses. Any less and sfdisk drops the
# recovery slot.
p5_end=$(( $(field 5 start) + $(field 5 size) - 1 ))
p4_end=$(( $(field 4 start) + $(field 4 size) - 1 ))
[ "$p4_end" = "$p5_end" ] \
    && note "the container ends where the recovery slot ends" OK \
    || note "the container ends at $p4_end, the recovery slot at $p5_end" FAIL

[ "$(( p5_end + 1 ))" -le "$TABLE_SECTORS" ] \
    && note "the table fits in $TABLE_SECTORS sectors" OK \
    || note "the table needs $(( p5_end + 1 )) sectors, want $TABLE_SECTORS or fewer" FAIL

echo
echo "===== the derived data start agrees with the table ====="

got=$(sh "$(dirname "$0")/data-start.sh" "$TABLE" 2>&1)
[ "$got" = "$TABLE_SECTORS" ] \
    && note "data-start.sh gives $TABLE_SECTORS" OK \
    || note "data-start.sh gives '$got', want $TABLE_SECTORS" FAIL
```

- [ ] **Step 2: Run it and watch the new cases fail**

```shell
MSG=; MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-partition.sh
```

Expected: `sfdisk accepts the table` FAILs first and the run stops, because the
current table needs 60506112 sectors and the test now offers 15523840.

- [ ] **Step 3: Rewrite the table**

Replace `tools/abslots/partition.sfdisk` in full:

```
# NanoKVM A/B/recovery layout. The table ends after the recovery slot, and the
# data partition is made on the first boot of the card.
#
# p1 is copied from the card as it is: start sector 1, 32768 sectors, type c,
# bootable. The boot ROM finds fip.bin by reading the first FAT partition, so
# this entry is the one thing here that must not change, and it is not
# reformatted. The official v1.4.3 release image has the same p1, so this is
# what a NanoKVM card looks like rather than a local convention.
#
# Everything from p2 is 4 MiB aligned, which is the erase block size that
# matters on this card. p1 cannot be aligned because it cannot move.
#
# Sizes come from measurement, not from a rule applied to a guess. The root
# filesystem in use holds 911 MB, of which 132 MB is accumulated operator state
# that a built image does not carry, so a built root is about 780 MB. 2 GiB is
# 2.5x that. Bigger would cost real time on every install: slot install writes a
# whole partition with dd, so install time and SD write endurance scale with the
# partition, not with the bytes in it.
#
# p4 is the extended container. Each logical partition costs one EBR sector, so
# p5 starts 8192 sectors after the container begins rather than immediately,
# which keeps it aligned too.
#
# THE TABLE DECLARES NO DATA PARTITION. A declared one needs a card of at least
# the sector it ends at, and that is what limited the 1.0.0 image to a 32 GB
# card. S01fs makes it on the first boot, at the end of the card, so the image
# fits any card of 8 GB or more and the data partition takes whatever is left.
# Its start sector comes from data-start.sh and is 10543104, which is where it
# sat in the fixed table this replaces.
#
#   p1  FAT16        1 ..    32768     16.00 MiB  /boot
#   p2  ext4     40960 ..  4235263      2.00 GiB  root A
#   p3  ext4   4235264 ..  8429567      2.00 GiB  root B
#   p5  ext4   8437760 .. 10534911      1.00 GiB  recovery
#   --                                            made on first boot: /data
label: dos
label-id: 0x70781617
unit: sectors

1 : start=1,        size=32768,    type=c, bootable
2 : start=40960,    size=4194304,  type=83
3 : start=4235264,  size=4194304,  type=83
4 : start=8429568,  size=2105344,  type=5
5 : start=8437760,  size=2097152,  type=83
```

- [ ] **Step 4: Run both suites**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh -c 'sh tools/abslots/test-partition.sh && sh tools/abslots/test-data-start.sh'
```

Expected: both report `all cases passed`. `test-data-start.sh` now passes its
first case as well, because the shipped table no longer declares p6.

- [ ] **Step 5: Commit**

```shell
git add tools/abslots/partition.sfdisk tools/abslots/test-partition.sh
git commit -F - <<'MSG'
Stop declaring the data partition in the shipped table

A declared data partition needs a card of at least the sector it ends at.
The table ended at 60506111, so the 1.0.0 image needed a card of 30.98 GB
and refused every 8 GB and 16 GB card, while the image itself is 5 GiB.

The table now stops after the recovery slot and the container stops with
it. S01fs makes the data partition on the first boot, at the end of
whatever card the image landed on.

The tests move with it: the table is applied to a 7.40 GiB disk, which is
a real 8 GB card, and the alignment loop no longer looks for a partition
the table does not describe.
MSG
```

---

### Task 3: The card image stops carrying a hole

**Files:**
- Modify: `tools/abslots/build-card.sh`
- Modify: `tools/abslots/test-build-card.sh`

**Interfaces:**
- Consumes: `data-start.sh` from Task 1, the table from Task 2.
- Produces: a card image of exactly `data-start.sh` sectors, with five
  partitions in its table. `CARD_DROP_FROM` no longer exists.

- [ ] **Step 1: Write the failing tests**

In `tools/abslots/test-build-card.sh`, change the offsets block:

```sh
# The offsets the assertions use, kept beside the table they come from.
P1=1
P2=40960
P3=57344
P5=75776
SLOT=16384
# The end of the image, which is also where the data partition starts. Same rule
# as the real table: the end of the recovery slot plus the 4 MiB EBR gap.
END=100352
```

In `setup()`, replace the test table so it declares no p6 and the container ends
with p5:

```sh
    cat > "$WORK/table.sfdisk" <<'EOF'
label: dos
label-id: 0x70781617
unit: sectors

1 : start=1,      size=32768,  type=c, bootable
2 : start=40960,  size=16384,  type=83
3 : start=57344,  size=16384,  type=83
4 : start=73728,  size=18432,  type=5
5 : start=75776,  size=16384,  type=83
EOF
```

In `run()`, drop the removed variable:

```sh
run() {
    CARD_TABLE="$WORK/table.sfdisk" \
        sh "$SCRIPT" "$WORK/boot" "$WORK/root.img" "$WORK/recovery.img" \
           "$WORK/card.img" > "$WORK/out" 2>&1
    echo $?
}
```

Replace the two partition-count assertions with these three:

```sh
# Five, not six. A sixth would mean the table still declares a data partition,
# which is what limited the image to a 32 GB card.
check "the image carries five partitions" \
    "$(sfdisk -l "$WORK/card.img" 2>/dev/null | grep -c '^/.*card\.img[0-9]')" "5"

check "no data partition is declared" \
    "$(sfdisk -l "$WORK/card.img" 2>/dev/null | grep -c 'card\.img6')" "0"

# The image ends where the data partition will start. The 4 MiB gap ahead of it
# is carried and zeroed, so a card that previously held a different layout has
# no stale EBR sector left anywhere the new chain could reach.
check "the image ends where the data partition will start" \
    "$(sectors "$WORK/card.img")" "$END"
```

Delete the case named `the data partition reappears on a full-sized card`, along
with the `cp` and `truncate` lines above it and the now-unused `FULL` constant.

- [ ] **Step 2: Run it and watch it fail**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-build-card.sh
```

Expected: `a card is built` FAILs. `build-card.sh` reads `field 6 start` and
`field 6 size`, finds neither, and stops with
`table.sfdisk does not describe the expected layout`.

- [ ] **Step 3: Change build-card.sh**

Replace the `Environment, for tests` comment block and the variable that follows:

```sh
# Environment, for tests:
#   CARD_TABLE            the sfdisk layout        (default partition.sfdisk)
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
BOOT=${1:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}
ROOT=${2:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}
RECOVERY=${3:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}
OUT=${4:?usage: build-card.sh <boot-dir> <root.img> <recovery.img> <out.img>}

TABLE=${CARD_TABLE:-$HERE/partition.sfdisk}
```

Replace the field reads and their validation:

```sh
P1_START=$(field 1 start); P1_SIZE=$(field 1 size)
P2_START=$(field 2 start); P2_SIZE=$(field 2 size)
P5_START=$(field 5 start); P5_SIZE=$(field 5 size)

for v in "$P1_START" "$P1_SIZE" "$P2_START" "$P2_SIZE" "$P5_START" "$P5_SIZE"; do
    [ -n "$v" ] || { echo "$TABLE does not describe the expected layout" >&2; exit 1; }
done
```

Replace the `FULL` computation and delete the `KEEP` block that followed it:

```sh
# The image ends where the data partition will start, and that sector comes from
# the same rule every other consumer uses.
#
# The table declares no data partition, so there is no hole to carry and no
# truncation afterwards. The 1.0.0 image was created at 28.85 GiB and cut back to
# 5 GiB, which cost 25 GB of writes on any filesystem without sparse files.
#
# The 4 MiB gap ahead of the data partition IS carried, and it is zeroed. parted
# writes the EBR for the data partition somewhere in that gap on the first boot,
# and a card that previously held a different layout must not leave a stale one
# there.
FULL=$("$HERE/data-start.sh" "$TABLE")
```

Delete the final `truncate -s $((KEEP * 512)) "$WORK"` line and its comment, and
change the closing message:

```sh
mv "$WORK" "$OUT"
trap - EXIT
echo "built $OUT: $((FULL / 2048)) MiB, slot B empty, data partition made on first boot"
```

Update the two paragraphs in the header comment that describe truncation:

```sh
# The image ENDS where the data partition starts. The table declares no data
# partition at all: S01fs makes it on the first boot, at the end of whatever card
# the image was written to. So the image is 5.02 GiB, it fits any card of 8 GB or
# more, and there is no 23 GiB hole to create and cut back.
#
# Slot B ships EMPTY. It is what the first update writes. Populating it would
# add 2 GiB to every download to carry a second copy of slot A.
```

- [ ] **Step 4: Run the suite**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-build-card.sh
```

Expected: `passed 18, failed 0`. The count is unchanged: the full-sized-card case
is gone, and two cases replaced the one that counted partitions.

- [ ] **Step 5: Commit**

```shell
git add tools/abslots/build-card.sh tools/abslots/test-build-card.sh
git commit -F - <<'MSG'
Build the card image at its final size

The image was created at 28.85 GiB and truncated back to 5 GiB, because
the table had to describe a data partition the image did not carry. On a
filesystem without sparse files that cost 25 GB of real writes.

The table declares no data partition now, so there is no hole and no
truncation. The image is created at the sector the data partition will
start at, which is the same derived number every other consumer uses.

The 4 MiB gap ahead of that sector is still carried, and it is zeroed.
parted writes the EBR for the data partition into that gap on the first
boot, and a card that previously held another layout must not leave a
stale one where the new chain could reach it.
MSG
```

---

### Task 4: The image tells the device where the data partition goes

**Files:**
- Modify: `tools/abslots/build-image.sh`
- Modify: `tools/abslots/test-build-image.sh`

**Interfaces:**
- Consumes: `data-start.sh` from Task 1.
- Produces: `/etc/nanokvm-slots.conf` in every image carries
  `DATA_START=10543104` beside `DATA_DEV=/dev/mmcblk0p6`. Task 5 reads it.

- [ ] **Step 1: Write the failing test**

In `tools/abslots/test-build-image.sh`, beside the existing
`/etc/nanokvm-slots.conf is written` case, add:

```sh
# S01fs makes the data partition on the first boot and needs to be told where it
# goes. Without this line it makes nothing, and the board comes up with no /data
# and the factory root password.
if grep -q '^DATA_START=10543104$' "$WORK/tree/etc/nanokvm-slots.conf" 2>/dev/null; then
    note "the slot conf carries DATA_START" OK
else
    note "the slot conf carries DATA_START" FAIL
fi
```

Match the surrounding cases for how `$WORK/tree` is named in that file. If the
suite unpacks the built image to a different path, use that path instead.

- [ ] **Step 2: Run it and watch it fail**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-build-image.sh
```

Expected: `the slot conf carries DATA_START` FAIL. Every other case passes.

- [ ] **Step 3: Write the line into the conf**

In `tools/abslots/build-image.sh`, add a `HERE` next to the other top-level
variable assignments, after the `set -e` line:

```sh
HERE=$(cd "$(dirname "$0")" && pwd)
```

Then replace the `nanokvm-slots.conf` heredoc. Note the delimiter loses its
quotes, so that the shell expands the one variable:

```sh
# S01fs makes the data partition on the first boot and reads its start sector
# from here. The number is derived from partition.sfdisk rather than written
# down, because the table and this file must never disagree about where the
# partition goes.
DATA_START=$("$HERE/data-start.sh")
cat > "$STAGE/tree/etc/nanokvm-slots.conf" <<CONF
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
DATA_DEV=/dev/mmcblk0p6
DATA_START=$DATA_START
CONF
```

- [ ] **Step 4: Run the suite**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-build-image.sh
```

Expected: every case passes, including the new one.

- [ ] **Step 5: Commit**

```shell
git add tools/abslots/build-image.sh tools/abslots/test-build-image.sh
git commit -F - <<'MSG'
Tell the image where the data partition goes

S01fs makes the data partition on the first boot, so it has to be told
which sector to start it at. The number is derived from the partition
table at build time and written into the slot configuration beside the
device name that is already there.

Deriving it rather than writing it down is what keeps the table and the
configuration from disagreeing. A board told the wrong sector makes a
partition the table does not describe.
MSG
```

---

### Task 5: S01fs can read the card's own geometry

Pure readers. Nothing in this task writes anything.

**Files:**
- Modify: `kvmapp/system/init.d/S01fs`
- Create: `tools/abslots/device/test-s01fs-provision.sh`

**Interfaces:**
- Consumes: `DATA_START` in `/etc/nanokvm-slots.conf` from Task 4.
- Produces, inside a block delimited by `# --- data geometry ---` and
  `# --- end data geometry ---`:
  - `disk_sectors` prints the size of `$DISK` in sectors
  - `part_end_sector <n>` prints the last sector of partition `n`
  - `part_size_sectors <n>` prints the size of partition `n` in sectors
  - `has_filesystem <dev>` returns 0 when `blkid` printed anything
  - `data_partition_ready <dev>` returns 0 when the kernel and the table agree
  - `data_start` prints `DATA_START` from the slot conf, or nothing
  Overridable for tests: `DISK`, `PARTED`, `BLKID`, `PARTITIONS`.

- [ ] **Step 1: Write the failing test**

Create `tools/abslots/device/test-s01fs-provision.sh`:

```sh
#!/bin/sh
# Check that S01fs reads the card's geometry correctly, and that it makes the
# data partition and its filesystem only when it should.
#
#   test-s01fs-provision.sh [path-to-S01fs]
#
# The 1.0.0 card image declared a data partition and nothing ever made a
# filesystem on it. The board came up with no /data, the factory root password
# and a new ssh host key, and S01fs printed OK. This file is the reason that
# cannot happen quietly again.
#
# Two device behaviours are stubbed exactly as measured, because the obvious
# implementation is wrong for both:
#
#   parted exits 1 on a busy disk AFTER it has succeeded. It cannot make the
#   kernel re-read the whole table while the root filesystem is mounted from the
#   same disk, so it warns and returns 1, having already written the table.
#
#   busybox blkid never prints a TYPE field, for any filesystem, and it exits 0
#   for a partition that holds none. Only an empty stdout means no filesystem.
S01=${1:-$(dirname "$0")/../../../kvmapp/system/init.d/S01fs}
[ -f "$S01" ] || { echo "usage: test-s01fs-provision.sh <S01fs>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

sed -n '/^# --- data geometry ---/,/^# --- end data geometry ---/p' "$S01" > "$WORK/geo.sh"
if [ ! -s "$WORK/geo.sh" ]; then
    note "the data geometry block can be extracted" FAIL
    echo; echo "$fails case(s) FAILED"; exit 1
fi
note "the data geometry block can be extracted" OK

# A parted stub that prints the machine-readable table of the card in use, and
# returns 1 the way the real one does on a busy disk.
cat > "$WORK/parted" <<'STUB'
#!/bin/sh
case "$*" in
    *print*) cat "$PARTED_TABLE" ;;
esac
echo "$*" >> "$PARTED_LOG"
exit "${PARTED_RC:-1}"
STUB
chmod +x "$WORK/parted"

# The table of the board in use: a 32 GB card whose data partition is present.
cat > "$WORK/full.table" <<'T'
BYT;
/dev/mmcblk0:60506112s:sd/mmc:512:512:msdos:SD SA32G:;
1:1s:32768s:32768s:fat16::boot, lba;
2:40960s:4235263s:4194304s:ext4::;
3:4235264s:8429567s:4194304s:::;
4:8429568s:60506111s:52076544s:::;
5:8437760s:10534911s:2097152s:ext4::;
6:10543104s:60506111s:49963008s:::;
T

# A freshly flashed card: no data partition, container stops with the recovery
# slot, and the card is bigger than the image.
cat > "$WORK/fresh.table" <<'T'
BYT;
/dev/mmcblk0:15523840s:sd/mmc:512:512:msdos:SD SA08G:;
1:1s:32768s:32768s:fat16::boot, lba;
2:40960s:4235263s:4194304s:ext4::;
3:4235264s:8429567s:4194304s:::;
4:8429568s:10534911s:2105344s:::;
5:8437760s:10534911s:2097152s:ext4::;
T

geo() {
    PARTED_TABLE=$1; PARTED_LOG=$WORK/plog; shift
    export PARTED_TABLE PARTED_LOG PARTITIONS BLKID_OUT
    sh -c "PARTED=$WORK/parted; . $WORK/geo.sh; $*" 2>/dev/null
}

echo "===== the table is read, not guessed ====="

got=$(geo "$WORK/full.table" 'disk_sectors')
[ "$got" = 60506112 ] && note "disk_sectors reads the disk line" OK \
                      || note "disk_sectors gave '$got', want 60506112" FAIL

got=$(geo "$WORK/full.table" 'part_end_sector 4')
[ "$got" = 60506111 ] && note "part_end_sector reads the container" OK \
                      || note "part_end_sector 4 gave '$got', want 60506111" FAIL

got=$(geo "$WORK/full.table" 'part_size_sectors 6')
[ "$got" = 49963008 ] && note "part_size_sectors reads the data partition" OK \
                      || note "part_size_sectors 6 gave '$got', want 49963008" FAIL

got=$(geo "$WORK/fresh.table" 'part_end_sector 6')
[ -z "$got" ] && note "a partition that is not there reads as nothing" OK \
              || note "part_end_sector 6 gave '$got' on a fresh card" FAIL

echo
echo "===== parted's exit status is never trusted ====="

# The stub always exits 1, exactly as the real one does on a busy disk. Every
# reader above ran against it. If any of them tested the exit status, they
# would have returned nothing.
got=$(geo "$WORK/full.table" 'disk_sectors')
[ "$got" = 60506112 ] \
    && note "a parted that exits 1 is still read" OK \
    || note "a parted that exits 1 was treated as a failure" FAIL

echo
echo "===== a filesystem is recognised by output, not by TYPE or by status ====="

cat > "$WORK/blkid" <<'STUB'
#!/bin/sh
[ -n "$BLKID_OUT" ] && echo "$BLKID_OUT"
exit 0
STUB
chmod +x "$WORK/blkid"

hasfs() {
    BLKID_OUT=$1
    export BLKID_OUT
    sh -c "BLKID=$WORK/blkid; . $WORK/geo.sh; has_filesystem /dev/mmcblk0p6 && echo yes || echo no" 2>/dev/null
}

# This is the exact output busybox blkid gives for the data partition in use.
# There is no TYPE field. An implementation that greps for one formats a live
# /data and destroys the board's identity.
got=$(hasfs '/dev/mmcblk0p6: LABEL="data" UUID="EE8B-6CB5"')
[ "$got" = yes ] \
    && note "busybox output with no TYPE counts as a filesystem" OK \
    || note "busybox output with no TYPE was read as an empty partition" FAIL

got=$(hasfs '/dev/mmcblk0p6: LABEL="data" UUID="6ACE-EE79" TYPE="exfat"')
[ "$got" = yes ] \
    && note "util-linux output also counts as a filesystem" OK \
    || note "util-linux output was read as an empty partition" FAIL

# An empty partition. blkid prints nothing and still exits 0, so the exit status
# says nothing at all.
got=$(hasfs '')
[ "$got" = no ] \
    && note "no output means no filesystem" OK \
    || note "an empty partition was read as holding a filesystem" FAIL

echo
echo "===== the kernel and the table must agree before anything is formatted ====="

ready() {
    cat > "$WORK/partitions" <<PART
major minor  #blocks  name

 179        0   30253056 mmcblk0
 179        6   $1 mmcblk0p6
PART
    PARTED_TABLE=$WORK/full.table; PARTED_LOG=$WORK/plog
    export PARTED_TABLE PARTED_LOG
    sh -c "PARTED=$WORK/parted; PARTITIONS=$WORK/partitions; . $WORK/geo.sh; \
           data_partition_ready /dev/mmcblk0p6 && echo yes || echo no" 2>/dev/null
}

# /proc/partitions counts 1024 byte blocks, so 24981504 blocks is 49963008
# sectors, which is what the table says.
got=$(ready 24981504)
[ "$got" = yes ] && note "a node that matches the table is ready" OK \
                 || note "a node that matches the table was rejected" FAIL

# A stale node. Formatting at this size makes a filesystem that cannot be
# repaired, because resize.exfat does not exist on this device.
got=$(ready 1024)
[ "$got" = no ] && note "a node that disagrees with the table is refused" OK \
                || note "a stale node was accepted, which would format the wrong size" FAIL

echo
echo "===== the start sector comes from the slot configuration ====="

cat > "$WORK/slots.conf" <<'CONF'
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
DATA_DEV=/dev/mmcblk0p6
DATA_START=10543104
CONF

got=$( SLOT_CONF="$WORK/slots.conf" sh -c ". $WORK/geo.sh; data_start" )
[ "$got" = 10543104 ] \
    && note "a conf with DATA_START is honoured" OK \
    || note "a conf with DATA_START gave '$got'" FAIL

# An image from before this change, or a board that was never migrated. It must
# print nothing, so that the caller makes no partition at all. Partitioning a
# board that did not declare where its data goes is worse than not partitioning.
printf 'DATA_DEV=/dev/mmcblk0p6\n' > "$WORK/old.conf"
got=$( SLOT_CONF="$WORK/old.conf" sh -c ". $WORK/geo.sh; data_start" )
[ -z "$got" ] \
    && note "a conf without DATA_START gives nothing" OK \
    || note "a conf without DATA_START gave '$got'" FAIL

got=$( SLOT_CONF="$WORK/absent.conf" sh -c ". $WORK/geo.sh; data_start" )
[ -z "$got" ] \
    && note "no conf at all gives nothing" OK \
    || note "no conf gave '$got'" FAIL

echo
echo "===== the script still parses ====="
sh -n "$S01" 2>/dev/null && note "sh -n accepts S01fs" OK || note "sh -n accepts S01fs" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
```

- [ ] **Step 2: Run it and watch it fail**

```shell
sh tools/abslots/device/test-s01fs-provision.sh
```

Expected: `the data geometry block can be extracted` FAILs and the run stops,
because no such block exists in `S01fs`.

- [ ] **Step 3: Add the geometry block**

In `kvmapp/system/init.d/S01fs`, insert this block immediately after the
`# --- end data device ---` line. Add `data_start` to the existing
`# --- data device ---` block instead if you prefer it beside `data_device`; the
test extracts `data_start` from the geometry block, so keep it here.

```sh
# --- data geometry ---
# What the card actually is, read from the table on the card itself.
#
# parted -sm prints one colon separated line per partition, after a BYT; header
# and a line describing the disk:
#
#   BYT;
#   /dev/mmcblk0:60506112s:sd/mmc:512:512:msdos:SD SA32G:;
#   6:10543104s:60506111s:49963008s:::;
#
# Field 2 of the disk line is the size of the disk in sectors. Fields 2, 3 and 4
# of a partition line are its start, its end and its size.
#
# NOTHING HERE READS PARTED'S EXIT STATUS. On a disk with a mounted partition
# parted cannot make the kernel re-read the whole table, so it warns and returns
# 1, having already done the work. Measured on this board: both the resize and
# the mkpart return 1 and both succeed.
DISK=${DISK:-/dev/mmcblk0}
PARTED=${PARTED:-parted}
BLKID=${BLKID:-blkid}
PARTITIONS=${PARTITIONS:-/proc/partitions}

table() {
    "$PARTED" -sm "$DISK" unit s print 2>/dev/null
}

disk_sectors() {
    table | awk -F: 'NR == 2 { gsub(/[^0-9]/, "", $2); print $2 }'
}

part_end_sector() {
    table | awk -F: -v p="$1" '$1 == p { gsub(/[^0-9]/, "", $3); print $3 }'
}

part_size_sectors() {
    table | awk -F: -v p="$1" '$1 == p { gsub(/[^0-9]/, "", $4); print $4 }'
}

# Whether a partition holds a filesystem this board can name.
#
# busybox blkid prints LABEL and UUID and NEVER prints a TYPE field, for vfat,
# for ext4 and for exfat alike. For a partition with no filesystem it prints
# nothing and still exits 0. So neither a TYPE test nor the exit status carries
# any information, and the only valid test is whether it printed anything.
#
# This is the guard that stands between a re-flash and the board's identity. A
# partition blkid can name is never touched.
has_filesystem() {
    [ -n "$("$BLKID" "$1" 2>/dev/null)" ]
}

# Whether the kernel and the table agree about the data partition.
#
# /proc/partitions counts 1024 byte blocks, so a sector count is twice it. A
# filesystem made on a node whose size is stale would be the wrong size and
# could not be repaired, because resize.exfat does not exist on this device.
data_partition_ready() {
    [ -e "$1" ] || return 1
    blocks=$(awk -v n="$(basename "$1")" '$4 == n { print $3 }' "$PARTITIONS" 2>/dev/null)
    [ -n "$blocks" ] || return 1
    [ "$((blocks * 2))" = "$(part_size_sectors "${1##*p}")" ]
}

# Which sector the data partition starts at, from the slot configuration.
#
# Prints nothing when the configuration does not say. A board that did not
# declare where its data goes must never be partitioned by this script: that is
# an unmigrated board or an image from before this change, and guessing at a
# start sector on a card whose layout is unknown is how a root slot gets
# overwritten.
data_start() {
    if [ -r "$SLOT_CONF" ]; then
        DATA_START=
        . "$SLOT_CONF"
        if [ -n "$DATA_START" ]; then
            echo "$DATA_START"
            return 0
        fi
    fi
    echo ""
}
# --- end data geometry ---
```

- [ ] **Step 4: Run the suite**

```shell
sh tools/abslots/device/test-s01fs-provision.sh
```

Expected: `all cases passed`. Then confirm the existing suite is unharmed:

```shell
sh tools/abslots/device/test-s01fs-datadev.sh
```

Expected: `all cases passed`.

- [ ] **Step 5: Commit**

```shell
git add kvmapp/system/init.d/S01fs tools/abslots/device/test-s01fs-provision.sh
git commit -F - <<'MSG'
Let S01fs read the card's own geometry

Readers only. Nothing here writes anything, and every value comes from the
partition table on the card rather than from a constant.

Two of them exist because the obvious implementation is wrong. parted
returns 1 on a disk with a mounted partition even when it has already done
the work, so no reader looks at its exit status. busybox blkid never
prints a TYPE field for any filesystem and exits 0 for a partition that
holds none, so the only test that means anything is whether it printed
something at all. A TYPE test would report a healthy data partition as
empty, and the next step would format it.

data_partition_ready compares the size the kernel reports against the size
the table declares. A filesystem made on a stale node would be the wrong
size and could not be repaired, because the device has no resize.exfat.
MSG
```

---

### Task 6: S01fs makes the partition and the filesystem

**Files:**
- Modify: `kvmapp/system/init.d/S01fs`
- Modify: `tools/abslots/device/test-s01fs-provision.sh`

**Interfaces:**
- Consumes: every function from Task 5.
- Produces, inside `# --- data provisioning ---` and
  `# --- end data provisioning ---`:
  - `container_reaches_end` returns 0 when p4 ends at the last sector of the disk
  - `provision_data <dev> <start>` returns 0 when the device holds a filesystem
    afterwards. Overridable for tests: `MKFS_EXFAT`.

- [ ] **Step 1: Write the failing tests**

Append to `tools/abslots/device/test-s01fs-provision.sh`, before the
`the script still parses` section:

```sh
echo
echo "===== the partition and the filesystem are made once, and only when needed ====="

sed -n '/^# --- data provisioning ---/,/^# --- end data provisioning ---/p' "$S01" > "$WORK/prov.sh"
if [ ! -s "$WORK/prov.sh" ]; then
    note "the data provisioning block can be extracted" FAIL
else
    note "the data provisioning block can be extracted" OK

    cat > "$WORK/mkfs.exfat" <<'STUB'
#!/bin/sh
echo "mkfs $*" >> "$MKFS_LOG"
exit "${MKFS_RC:-0}"
STUB
    chmod +x "$WORK/mkfs.exfat"

    # provision <table> <blkid-output> <device-exists yes|no> <kernel-blocks>
    provision() {
        PARTED_TABLE=$1
        BLKID_OUT=$2
        PARTED_LOG="$WORK/plog"; : > "$PARTED_LOG"
        MKFS_LOG="$WORK/mlog";   : > "$MKFS_LOG"

        rm -f "$WORK/dev"
        [ "$3" = yes ] && : > "$WORK/dev"

        cat > "$WORK/partitions" <<PART
major minor  #blocks  name

 179        6   $4 dev
PART
        export PARTED_TABLE PARTED_LOG BLKID_OUT MKFS_LOG
        sh -c "PARTED=$WORK/parted; BLKID=$WORK/blkid; \
               MKFS_EXFAT=$WORK/mkfs.exfat; PARTITIONS=$WORK/partitions; \
               . $WORK/geo.sh; . $WORK/prov.sh; \
               provision_data $WORK/dev 10543104" > /dev/null 2>&1
        echo $?
    }

    # The board in use. Everything is already there, so nothing may be written.
    rc=$(provision "$WORK/full.table" '/dev/mmcblk0p6: LABEL="data" UUID="EE8B-6CB5"' yes 24981504)
    [ "$rc" = 0 ] && note "a card that is already provisioned reports success" OK \
                  || note "a provisioned card returned $rc" FAIL
    [ ! -s "$WORK/plog" ] \
        && note "and it does not run parted at all" OK \
        || note "it ran parted on a provisioned card: $(tr '\n' ';' < "$WORK/plog")" FAIL
    [ ! -s "$WORK/mlog" ] \
        && note "and it does not format anything" OK \
        || note "IT FORMATTED A LIVE DATA PARTITION" FAIL

    # A freshly flashed card. Both parted calls must run, in order, and the
    # filesystem must be made.
    rc=$(provision "$WORK/fresh.table" '' no 0)
    grep -q '^-s /dev/mmcblk0 resizepart 4 100%$' "$WORK/plog" \
        && note "a fresh card grows the container first" OK \
        || note "the container was not grown, log: $(tr '\n' ';' < "$WORK/plog")" FAIL

    # The explicit start keeps every card on one layout. Without it parted picks
    # its own aligned offset, 8192 sectors further along.
    grep -q '^-s /dev/mmcblk0 mkpart logical ntfs 10543104s 100%$' "$WORK/plog" \
        && note "it makes the partition at the declared sector, as type ntfs" OK \
        || note "the mkpart call is wrong: $(tr '\n' ';' < "$WORK/plog")" FAIL

    [ "$(head -1 "$WORK/plog" | grep -c resizepart)" = 1 ] \
        && note "and the resize runs before the mkpart" OK \
        || note "the mkpart ran first, which parted refuses" FAIL

    # A card whose container already reaches the end, from a run that got that
    # far and stopped. The resize must not run again: it would rewrite sector 0
    # for nothing, on every boot.
    cat > "$WORK/grown.table" <<'T'
BYT;
/dev/mmcblk0:15523840s:sd/mmc:512:512:msdos:SD SA08G:;
4:8429568s:15523839s:7094272s:::;
5:8437760s:10534911s:2097152s:ext4::;
T
    rc=$(provision "$WORK/grown.table" '' no 0)
    grep -q resizepart "$WORK/plog" \
        && note "a container that already reaches the end is resized again" FAIL \
        || note "a container that already reaches the end is left alone" OK

    # A table that cannot be read at all. Writing a partition table blind is
    # worse than leaving the card as it is.
    : > "$WORK/empty.table"
    rc=$(provision "$WORK/empty.table" '' no 0)
    [ "$rc" != 0 ] && note "an unreadable table reports failure" OK \
                   || note "an unreadable table reported success" FAIL
    [ ! -s "$WORK/plog" ] \
        && note "and nothing is written to it" OK \
        || note "it wrote to a disk whose table it could not read" FAIL

    # A node the kernel has not caught up with. Formatting it makes a filesystem
    # of the wrong size that nothing on this device can repair.
    rc=$(provision "$WORK/full.table" '' yes 1024)
    [ ! -s "$WORK/mlog" ] \
        && note "a stale node is not formatted" OK \
        || note "a stale node was formatted at the wrong size" FAIL
    [ "$rc" != 0 ] && note "and it reports failure" OK \
                   || note "a stale node reported success" FAIL

    # A partition that exists and is empty. This is the repair path: a card
    # whose data partition survived but whose filesystem did not.
    rc=$(provision "$WORK/full.table" '' yes 24981504)
    grep -q 'mkfs .*-L data' "$WORK/mlog" \
        && note "an empty data partition is formatted, labelled data" OK \
        || note "an empty data partition was not formatted: $(cat "$WORK/mlog")" FAIL
    grep -q resizepart "$WORK/plog" \
        && note "and the table is not rewritten to do it" FAIL \
        || note "and the table is not rewritten to do it" OK

    # mkfs failing must be reported, not swallowed. A caller that believes the
    # filesystem is there goes on to print that /data is ready.
    MKFS_RC=1 rc=$(MKFS_RC=1 provision "$WORK/full.table" '' yes 24981504)
    [ "$rc" != 0 ] && note "a failed mkfs reports failure" OK \
                   || note "a failed mkfs reported success" FAIL
fi
```

- [ ] **Step 2: Run it and watch the new cases fail**

```shell
sh tools/abslots/device/test-s01fs-provision.sh
```

Expected: `the data provisioning block can be extracted` FAILs and the cases
inside it are skipped. Every case from Task 5 still passes.

- [ ] **Step 3: Add the provisioning block**

In `kvmapp/system/init.d/S01fs`, insert immediately after
`# --- end data geometry ---`:

```sh
# --- data provisioning ---
# Make the data partition and its filesystem on the first boot of a card.
#
# The shipped table declares p1 to p5 and stops, so the image fits any card of
# 8 GB or more. This gives the data partition everything the card has left.
#
# Every step is gated on a condition read from the partition table, and each gate
# is satisfied by the step it guards. So nothing repeats, and no marker file is
# needed. The resize guard above already argues for this: a marker in the root
# filesystem does not survive a slot image built from a fresh rootfs, and the
# table is the same for every slot on the card.
#
# The gates are also what bounds the retries. A mkpart that fails writes nothing
# at all, which is measured, so a card that cannot be partitioned does not wear
# its own sector 0 on every boot.
#
# parted prints "udevadm: not found" on every call, because the board runs mdev.
# That is noise and not a fault.
MKFS_EXFAT=${MKFS_EXFAT:-mkfs.exfat}

# The container has to reach the end of the card before a logical partition can
# be made out there. Growing it rewrites sector 0. Measured on four card sizes:
# the entries for /boot, root A and root B come out bit for bit identical, and
# only the container's own entry changes.
container_reaches_end() {
    end=$(part_end_sector 4)
    disk=$(disk_sectors)
    [ -n "$end" ] && [ -n "$disk" ] && [ "$end" = "$((disk - 1))" ]
}

provision_data() {
    dev=$1
    start=$2

    # The common case, on every boot after the first. One blkid and nothing else.
    if [ -e "$dev" ] && has_filesystem "$dev"; then
        return 0
    fi

    # A table that cannot be read is a card this script must not write to.
    if [ -z "$(disk_sectors)" ] || [ -z "$(part_end_sector 4)" ]; then
        echo "S01fs: cannot read the partition table of $DISK, leaving the card alone"
        return 1
    fi

    if ! container_reaches_end; then
        echo "S01fs: growing the partition container to the end of the card"
        "$PARTED" -s "$DISK" resizepart 4 100% > /dev/null 2>&1
        if ! container_reaches_end; then
            echo "S01fs: the container did not grow, leaving the card alone"
            return 1
        fi
    fi

    if [ ! -e "$dev" ]; then
        echo "S01fs: making the data partition at sector $start"
        "$PARTED" -s "$DISK" mkpart logical ntfs "${start}s" 100% > /dev/null 2>&1
    fi

    # The size the kernel reports has to match the size the table declares before
    # anything is formatted. resize.exfat does not exist on this device, so a
    # filesystem made at the wrong size is a filesystem that stays wrong.
    if ! data_partition_ready "$dev"; then
        echo "S01fs: the data partition did not appear at $dev, leaving the card alone"
        return 1
    fi

    echo "S01fs: making an exfat filesystem on $dev, $(( $(part_size_sectors "${dev##*p}") / 2048 )) MiB"
    if ! "$MKFS_EXFAT" -L data "$dev" > /dev/null 2>&1; then
        echo "S01fs: mkfs.exfat failed on $dev"
        return 1
    fi

    return 0
}
# --- end data provisioning ---
```

- [ ] **Step 4: Run the suite**

```shell
sh tools/abslots/device/test-s01fs-provision.sh && sh tools/abslots/device/test-s01fs-datadev.sh
```

Expected: both report `all cases passed`.

- [ ] **Step 5: Commit**

```shell
git add kvmapp/system/init.d/S01fs tools/abslots/device/test-s01fs-provision.sh
git commit -F - <<'MSG'
Make the data partition on the first boot of a card

The shipped table stops after the recovery slot, so the card image fits an
8 GB card. This gives the data partition everything the card has left.

Three gates, each satisfied by the step it guards, so nothing repeats and
no marker file is needed. S01fs already argues for reading the table
rather than a marker: a marker in the root filesystem does not survive a
slot image built from a fresh rootfs.

The order is not a preference. mkpart on its own fails with "Can't have
overlapping partitions", because the container does not yet reach the new
partition, so the resize has to run first. The ntfs argument is what makes
the partition type 7 rather than 83, and a card pulled from the board is
otherwise not recognised by Windows or macOS. The explicit start sector
keeps every card on one layout: without it parted picks its own offset,
8192 sectors further along.

Nothing is formatted until the size the kernel reports matches the size
the table declares, and nothing is formatted at all if blkid can name what
is already there. That second guard is what stands between a re-flash and
the board's identity.
MSG
```

---

### Task 7: S01fs runs it, and stops reporting a mount it did not get

**Files:**
- Modify: `kvmapp/system/init.d/S01fs`
- Modify: `tools/abslots/device/test-s01fs-datadev.sh`

**Interfaces:**
- Consumes: `data_start`, `data_device`, `provision_data`, `mount_data`.
- Produces: nothing new. This is the wiring.

- [ ] **Step 1: Write the failing tests**

In `tools/abslots/device/test-s01fs-datadev.sh`, replace the call-site case,
because the line is about to gain an `if`:

```sh
# The call site must go through the function, or none of the above runs on the
# device.
if grep -qE 'mount_data[[:space:]]+"\$DATADEV"[[:space:]]+/data' "$S01"; then
    note "the call site uses mount_data" OK
else
    note "the call site still calls mount directly" FAIL
fi

# A mount that failed must not be reported as a mounted /data. S01fs discarded
# the return value and printed OK, so a board with no /data, the factory root
# password and a new ssh host key reported a clean boot. That is how the 1.0.0
# card image shipped without anybody noticing it made no data partition.
if grep -qE '^[[:space:]]*if[[:space:]]+mount_data[[:space:]]' "$S01"; then
    note "the call site tests whether the mount worked" OK
else
    note "the call site discards the result of the mount" FAIL
fi

if grep -q 'FAILED to mount' "$S01"; then
    note "and it says so when the mount fails" OK
else
    note "a failed mount is still silent" FAIL
fi

# The provisioning has to be wired in, or it never runs on the device.
if grep -qE '^[[:space:]]*provision_data[[:space:]]+"\$DATADEV"' "$S01"; then
    note "the boot path calls provision_data" OK
else
    note "provision_data is defined and never called" FAIL
fi
```

- [ ] **Step 2: Run it and watch the three new cases fail**

```shell
sh tools/abslots/device/test-s01fs-datadev.sh
```

Expected: `the call site tests whether the mount worked` FAIL,
`a failed mount is still silent` FAIL, `provision_data is defined and never
called` FAIL. The rest pass.

- [ ] **Step 3: Wire it into the boot path**

In `kvmapp/system/init.d/S01fs`, replace the block that starts `DATADEV=$(data_device)`:

```sh
        DATADEV=$(data_device)
        DATASTART=$(data_start)

        # Only a layout that declares where its data goes is ever partitioned.
        # A board without that line is an unmigrated board or an image from
        # before this existed, and guessing at a start sector on a card whose
        # layout is unknown is how a root slot gets overwritten.
        if [ -n "$DATASTART" ]
        then
                provision_data "$DATADEV" "$DATASTART" || true
        fi

        if [ -e "$DATADEV" ]
        then
                mkdir -p /data
                # The result is tested. This script used to discard it and print
                # OK, so a board with no /data reported a clean boot and only
                # said so weeks later, through the factory root password.
                if mount_data "$DATADEV" /data
                then
                        printf "(data on %s) " "$DATADEV"
                else
                        printf "(FAILED to mount %s) " "$DATADEV"
                fi
        else
                printf "(no data device at %s) " "$DATADEV"
        fi
```

The `|| true` is deliberate. `S01fs` runs under `rcS` and a nonzero return from
this branch must not stop the rest of the boot. `provision_data` has already said
what went wrong on the console.

- [ ] **Step 4: Run every device suite**

```shell
sh tools/abslots/device/test-s01fs-datadev.sh && \
sh tools/abslots/device/test-s01fs-provision.sh && \
sh -n kvmapp/system/init.d/S01fs && echo "sh -n clean"
```

Expected: both suites report `all cases passed`, and `sh -n clean`.

Then check it under the device's own shell, because busybox ash is what runs it:

```shell
timeout 60 ssh root@10.0.0.222 'sh -n /dev/stdin' < kvmapp/system/init.d/S01fs \
    && echo "busybox sh -n clean"
```

Expected: `busybox sh -n clean`. **Do not install it on the board.** The board
already has its data partition and this changes nothing for it, so a deploy buys
no evidence and risks the one working card.

- [ ] **Step 5: Commit**

```shell
git add kvmapp/system/init.d/S01fs tools/abslots/device/test-s01fs-datadev.sh
git commit -F - <<'MSG'
Run the provisioning, and stop reporting a mount that did not happen

S01fs discarded the return value of mount_data and printed OK either way.
So a board that flashed the 1.0.0 card image came up with no /data, the
factory root password and a new ssh host key, and reported a clean boot.
The fault was found by reading the code, not by the board, which is the
whole problem with a script that reports success it did not have.

The provisioning runs only when the slot configuration says where the data
partition goes. A board without that line is unmigrated or carries an
image from before this change, and guessing at a start sector on a card
whose layout is unknown is how a root slot gets overwritten.

A failure returns nonzero and the boot continues, because rcS must not
stop here. provision_data has already said what went wrong.
MSG
```

---

### Task 8: Break every new guard on purpose

**Files:**
- Create: `tools/abslots/device/test-s01fs-mutation.sh`

**Interfaces:**
- Consumes: `S01fs` and `test-s01fs-provision.sh`.
- Produces: nothing. It is a test of the tests.

- [ ] **Step 1: Write the mutation suite**

Create `tools/abslots/device/test-s01fs-mutation.sh`:

```sh
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
#    succeeded, so this skips the mkfs on every card that actually worked.
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
#    no gain. That is the exact wear the resize guard above was written for.
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
```

- [ ] **Step 2: Run it**

```shell
sh tools/abslots/device/test-s01fs-mutation.sh
```

Expected: `all cases passed`.

If a case reports `THE MUTATION DID NOT APPLY`, the `sed` did not match the file.
Fix the `sed` to match what `S01fs` actually contains. **Do not weaken the case
and do not delete it.** A mutation that never applied proves nothing, and a
suite that reports OK for it is the exact rot this file exists to catch.

If a case reports FAIL without that marker, the suite genuinely does not catch
the mutation. Add the missing case to `test-s01fs-provision.sh` first, watch it
fail against the mutant, then re-run this.

- [ ] **Step 3: Commit**

```shell
git add tools/abslots/device/test-s01fs-mutation.sh
git commit -F - <<'MSG'
Break every data provisioning guard on purpose

A guard that stops guarding reads exactly like a guard that holds, and two
suites in this repository had already rotted that way while reporting
success.

Eight mutations, and every one of them is something somebody would
plausibly write. Four destroy the board's identity by formatting a data
partition that already holds one. One makes a filesystem at a size nothing
on the device can repair. One skips the whole job on every card where it
actually worked, by trusting an exit status that means nothing here.

The runner refuses a mutation that did not change the file. A sed that
stopped matching would otherwise report a caught mutation that was never
made.
MSG
```

---

### Task 9: Prove it against real parted and real mkfs.exfat

**Files:**
- Modify: `tools/release/Dockerfile`
- Create: `tools/abslots/test-provision-integration.sh`

**Interfaces:**
- Consumes: `build-card.sh`, `data-start.sh`, the two `S01fs` blocks.
- Produces: nothing. It is the end-to-end proof that a built image reaches a
  mounted data filesystem.

- [ ] **Step 1: Add the tools to the release host**

In `tools/release/Dockerfile`, extend the apt line:

```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
        fdisk util-linux dosfstools mtools e2fsprogs zstd xz-utils \
        u-boot-tools cpio device-tree-compiler openssl ca-certificates \
        coreutils tar gzip file patchelf parted exfatprogs \
    && rm -rf /var/lib/apt/lists/*
```

`parted` and `exfatprogs` are what the integration test drives. The release
itself does not use either: the card image no longer carries a data filesystem,
because the device makes it.

- [ ] **Step 2: Rebuild the image**

```shell
MSYS_NO_PATHCONV=1 docker build -t ironkvm-release-host:latest tools/release
MSYS_NO_PATHCONV=1 docker run --rm ironkvm-release-host:latest \
    sh -c 'parted --version | head -1; mkfs.exfat -V | head -1'
```

Expected: a parted version and an exfatprogs version, both printed.

- [ ] **Step 3: Write the integration test**

Create `tools/abslots/test-provision-integration.sh`:

```sh
#!/bin/sh
# Drive a real card image through the real parted and the real mkfs.exfat.
#
#   test-provision-integration.sh
#
# This is the test that would have caught the fault in 1.0.0. That image
# declared a data partition, nothing ever made a filesystem on it, and every
# other suite in this repository passed.
#
# Needs parted, exfatprogs, sfdisk, mtools and e2fsprogs. Run it in the release
# host image, which carries all of them.
#
# What this does NOT cover: whether the kernel exposes the new partition node
# while the root filesystem is mounted from the same disk. parted is driven
# against a FILE here, so there are no partition nodes at all. That step is
# measured separately on the board through a loop device, and it is closed for
# good only by a flashed card. See docs/specs/2026-08-18-small-card-image-design.md.
set -u

HERE=$(cd "$(dirname "$0")" && pwd)
pass=0
fail=0
ok()    { pass=$((pass + 1)); echo "  ok    $1"; }
bad()   { fail=$((fail + 1)); echo "  FAIL  $1"; }
check() { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (got '$2', want '$3')"; fi; }

for t in parted mkfs.exfat blkid sfdisk mkfs.vfat mke2fs mcopy; do
    command -v "$t" > /dev/null 2>&1 || { echo "SKIP: $t is missing"; exit 0; }
done

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

# A scaled copy of the shipped table. The real one is 5 GiB, which is too slow
# to build in a test loop. p1 keeps its real size because FAT16 refuses to make
# a filesystem much smaller.
cat > "$WORK/table.sfdisk" <<'EOF'
label: dos
label-id: 0x70781617
unit: sectors

1 : start=1,      size=32768,  type=c, bootable
2 : start=40960,  size=16384,  type=83
3 : start=57344,  size=16384,  type=83
4 : start=73728,  size=18432,  type=5
5 : start=75776,  size=16384,  type=83
EOF

START=$(sh "$HERE/data-start.sh" "$WORK/table.sfdisk")
check "the derived start is the end of p5 plus 4 MiB" "$START" "100352"

mkdir -p "$WORK/boot"
printf 'fip\n' > "$WORK/boot/fip.bin"
for img in root recovery; do
    dd if=/dev/zero of="$WORK/$img.img" bs=512 count=16384 2>/dev/null
    mke2fs -q -t ext4 -F "$WORK/$img.img" > /dev/null 2>&1
done

CARD_TABLE="$WORK/table.sfdisk" sh "$HERE/build-card.sh" \
    "$WORK/boot" "$WORK/root.img" "$WORK/recovery.img" "$WORK/card.img" \
    > "$WORK/build.log" 2>&1
check "a card image is built" "$?" "0"
check "the image ends where the data partition will start" \
    "$(( $(wc -c < "$WORK/card.img") / 512 ))" "$START"

# Write it to a "card". A card is bigger than the image, which is the whole
# point of the change: 393216 sectors is 192 MiB.
CARD=393216
cp "$WORK/card.img" "$WORK/flashed.img"
truncate -s $((CARD * 512)) "$WORK/flashed.img"

# The two calls provision_data makes, in the order it makes them, against the
# real parted. Its exit status is ignored here for the same reason the device
# ignores it.
parted -s "$WORK/flashed.img" resizepart 4 100% > /dev/null 2>&1 || true
parted -s "$WORK/flashed.img" mkpart logical ntfs "${START}s" 100% > /dev/null 2>&1 || true

field() {
    sfdisk -d "$WORK/flashed.img" 2>/dev/null \
        | awk -v p="$1" -v k="$2" '
            $0 ~ (p " *:") {
                n = split($0, t, "[ ,]+")
                for (i = 1; i <= n; i++) if (t[i] == k "=") { print t[i+1]; exit }
                for (i = 1; i <= n; i++) if (t[i] ~ "^" k "=") { sub("^" k "=", "", t[i]); print t[i]; exit }
            }'
}

check "the data partition is made" "$(field "$WORK/flashed.img6" start)" "$START"
check "it is type 7, so the card reads on Windows and macOS" \
    "$(field "$WORK/flashed.img6" type)" "7"
check "it reaches the end of the card" \
    "$(( $(field "$WORK/flashed.img6" start) + $(field "$WORK/flashed.img6" size) ))" "$CARD"

# The entries for /boot, root A and root B must be untouched by the write to
# sector 0. This is the one that would be expensive to get wrong.
check "the boot partition still starts at sector 1" "$(field "$WORK/flashed.img1" start)" "1"
check "the boot partition keeps type c" "$(field "$WORK/flashed.img1" type)" "c"
check "root A is unmoved" "$(field "$WORK/flashed.img2" start)" "40960"
check "root B is unmoved" "$(field "$WORK/flashed.img3" start)" "57344"
check "the recovery slot is unmoved" "$(field "$WORK/flashed.img5" start)" "75776"

# The slots must still pass a check after sector 0 was rewritten.
dd if="$WORK/flashed.img" of="$WORK/roota.out" bs=512 skip=40960 count=16384 status=none
check "root A still passes e2fsck" \
    "$(e2fsck -fn "$WORK/roota.out" > /dev/null 2>&1 && echo clean || echo dirty)" "clean"

# And the filesystem. Carve the data partition out and format it exactly as the
# device does, then ask blkid whether it is there. Before this change nothing
# ever ran this step, on any card.
DSIZE=$(field "$WORK/flashed.img6" size)
dd if="$WORK/flashed.img" of="$WORK/data.out" bs=512 skip="$START" count="$DSIZE" status=none
mkfs.exfat -L data "$WORK/data.out" > /dev/null 2>&1
check "mkfs.exfat makes a filesystem on it" "$?" "0"
check "and blkid can name it" \
    "$([ -n "$(blkid "$WORK/data.out" 2>/dev/null)" ] && echo named || echo silent)" "named"

# The guard that decides whether to format has to say no to this one now.
S01="$HERE/../../kvmapp/system/init.d/S01fs"
sed -n '/^# --- data geometry ---/,/^# --- end data geometry ---/p' "$S01" > "$WORK/geo.sh"
got=$(sh -c ". $WORK/geo.sh; has_filesystem $WORK/data.out && echo yes || echo no")
check "has_filesystem agrees, so a re-flash would not destroy it" "$got" "yes"

echo
echo "passed $pass, failed $fail"
[ "$fail" -eq 0 ]
```

- [ ] **Step 4: Run it**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh tools/abslots/test-provision-integration.sh
```

Expected: `passed 15, failed 0`.

If `the data partition is made` fails, print the table and compare it against the
probe in the spec:

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh -c 'sfdisk -d /tmp/flashed.img'
```

- [ ] **Step 5: Commit**

```shell
git add tools/release/Dockerfile tools/abslots/test-provision-integration.sh
git commit -F - <<'MSG'
Drive a built card image through real parted and real mkfs.exfat

This is the test that would have caught the fault in 1.0.0. That image
declared a data partition, nothing ever made a filesystem on it, and every
other suite here passed.

It builds a scaled card image, writes it to a file bigger than itself the
way a flash writes to a card bigger than the image, runs the two parted
calls the device runs, and then checks the things that would be expensive
to get wrong: the data partition is at the derived sector and is type 7,
and the entries for /boot, root A and root B are unmoved by the write to
sector 0, and root A still passes e2fsck afterwards.

Then it formats the data partition and asks the shipped has_filesystem
whether it can see it, which is the guard that decides whether a re-flash
keeps the board's identity or destroys it.

The release host gains parted and exfatprogs to run it. The release itself
uses neither: the image carries no data filesystem, because the device
makes it.

What this does not cover is whether the kernel exposes the new partition
node while the root filesystem is mounted from the same disk. parted is
driven against a file here, so there are no nodes at all.
MSG
```

---

### Task 10: Correct every document that describes the old behaviour

Three files state that the first boot makes the data partition, and it never
did. One states a card requirement that is about to change.

**Files:**
- Modify: `README.md`
- Modify: `docs/CHANGES-FROM-OFFICIAL.md`
- Modify: `tools/release/release.sh`
- Modify: `docs/plans/2026-08-16-ironkvm-1.0.md`

**Interfaces:**
- Consumes: everything above.
- Produces: nothing executable.

- [ ] **Step 1: Correct the card requirement in README.md**

Replace the paragraph added in commit `51f6e60f`, the one beginning
`The card image needs a microSD card of`:

```markdown
The card image needs a microSD card of 8 GB or more. The image itself is 5.02
GiB, and the first boot gives the data partition everything the card has left, so
a bigger card is used rather than wasted.

The download does not change with the card. It is one file of about 372 MB for
every size, because the image is mostly the two slot filesystems and the empty
one compresses away.

Buy for endurance rather than for speed. The board runs the card at 25 MHz in
high-speed mode and never switches to 1.8 V signalling, so UHS-I speed grades
cannot be reached: a card rated 100 MB/s reads at about 10 MB/s here, the same as
a Class 10 card. An A1 card suits this host and an A2 card can be slower, because
A2 assumes a controller that queues commands and this one does not.
```

Keep the `## Install` step 4 as it is. It says the first boot makes the data
partition and takes longer, and after this change that is finally true.

- [ ] **Step 2: Replace the deferred item in the fork diff**

In `docs/CHANGES-FROM-OFFICIAL.md`, remove the whole
`**Growing the data partition.**` entry from the deferred list. It describes work
that is now done, and a deferred list that carries finished work stops being read.

Add it to whatever section of that file records differences that are implemented.
Match the surrounding entries for tone and length:

```markdown
- **The data partition is made on the first boot, not in the image.** The card
  image carries a partition table that stops after the recovery slot, so it
  installs on a card of 8 GB rather than 31 GB, and the data partition takes
  whatever the card has left. Upstream does the same thing for its own data
  partition. This fork had to disable that code, because upstream makes the
  partition at p3 and p3 is the second root slot in the A/B layout.
```

- [ ] **Step 3: Correct the comment in release.sh**

In `tools/release/release.sh`, replace the paragraph under `==> assembling the card`
that begins `build-card.sh creates the card at its full 28.85 GiB`:

```sh
# build-card.sh creates the card at exactly its final size. The table declares no
# data partition, so there is no hole to make and cut back: the device makes that
# partition on the first boot, at the end of whatever card the image reached.
#
# $STAGE rather than $OUT is still deliberate. The two slot filesystems are read
# back out of the image to be checked, so the same 5 GiB crosses the filesystem
# three times, and $OUT on a Windows workstation is a bind mount.
```

- [ ] **Step 4: Correct the false claim in the 1.0 plan**

`docs/plans/2026-08-16-ironkvm-1.0.md` is a record of what was planned, so do not
rewrite its tasks. Add a correction note directly beneath the sentence at line
1480 that begins `The image is **truncated after p5**`:

```markdown
> **Correction, 2026-08-18.** The sentence above is wrong, and it was wrong when
> it was written. `S01fs` never formatted the data device: its only `mkfs.exfat`
> names p3, and two separate gates stop that branch on an A/B card. So the 1.0.0
> card image made no data partition, and a board flashed with it came up with no
> `/data`, the factory root password and a new ssh host key. Nobody saw it
> because no card was ever flashed. Fixed for 1.0.1 by
> `docs/specs/2026-08-18-small-card-image-design.md`.
```

Add a matching note beneath the `**The size limit is stated, not solved.**`
paragraph at line 1484:

```markdown
> **Superseded, 2026-08-18.** The table no longer declares the data partition, so
> the size limit is gone. The image installs on a card of 8 GB or more and the
> first boot uses the whole card.
```

- [ ] **Step 5: Check the whole tree still agrees with itself**

```shell
grep -rn "60506112\|60,506,112\|28.85 GiB\|30.98 GB\|at least 32 GB" \
    README.md docs/CHANGES-FROM-OFFICIAL.md tools/ --include='*.md' --include='*.sh'
```

Expected: only matches inside `docs/specs/2026-08-18-small-card-image-design.md`
and `docs/plans/`, which are records of what was measured and what was planned.
Anything in `README.md`, `tools/` or `docs/CHANGES-FROM-OFFICIAL.md` is a
statement about the current firmware and must be corrected.

Then confirm no em dash reached any file:

```shell
git diff --cached -U0 | grep -n '^+.*—' && echo "EM DASH FOUND" || echo "clean"
```

- [ ] **Step 6: Commit**

```shell
git add README.md docs/CHANGES-FROM-OFFICIAL.md tools/release/release.sh \
        docs/plans/2026-08-16-ironkvm-1.0.md
git commit -F - <<'MSG'
Correct every document that said the first boot made the data partition

Three files stated it and none of it was true. The 1.0 plan asserted that
S01fs already formatted the data device, build-card.sh repeated it, and
README.md promised it to whoever flashed the card. The only file that
described the real behaviour was root.manifest, which explains that the
marker exists to STOP the format.

The plan keeps its original text and gains a correction beneath it,
because it is a record of what was planned rather than a description of
the firmware.

README.md now asks for a card of 8 GB, says that the download does not
change with the size of the card, and keeps the endurance advice. The
deferred item about growing the data partition moves out of the deferred
list, because a deferred list that carries finished work stops being read.
MSG
```

---

## Release gate

**1.0.1 is not published until a flashed card boots.** Everything above is
measured on this board's parted, busybox and kernel, but through a loop device
and through files. The write to sector 0 of the real boot device is untested.

- [x] **Step 1: Build the release**

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/mnt/d/projects/NanoKVM" \
    -v //var/run/docker.sock:/var/run/docker.sock \
    -w /mnt/d/projects/NanoKVM ironkvm-release-host:latest \
    sh tools/release/release.sh --dry-run 1.0.1
```

Check `release-out/ironkvm-1.0.1-sdcard.img.xz` exists, and that the uncompressed
image is 10543104 sectors:

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/w" -w /w ironkvm-release-host:latest \
    sh -c 'xz -dc release-out/ironkvm-1.0.1-sdcard.img.xz | wc -c'
```

Expected: `5398069248`, which is 10543104 sectors.

**Done, 2026-08-18.** The dry run completed and the artifacts verify:

```
uncompressed bytes : 5398069248   (10543104 sectors, as expected)
build-card.sh said : 5148 MiB, slot B empty, data partition made on first boot
table in the image : 5 partitions, no p6
                     p1 bootable, start 1, type c
                     p4 extended, 8429568 .. 10534911
                     p5 ext4,     8437760 .. 10534911
ironkvm-1.0.1-sdcard.img.xz  372289444 bytes
ironkvm_1.0.1.tar.gz          12492266 bytes
```

The compressed image is 372.3 MB against 1.0.0's 372.3 MB, which confirms what
`README.md` now says: the download does not change with the card.

**Run this from WSL, not from Git Bash.** `release.sh` starts a second container
and passes `$PWD` as the bind source, so the daemon has to resolve that path
itself. Docker Desktop's daemon cannot see `/mnt/d/...` when the outer container
is launched from Git Bash: the nested mount silently comes up empty and the build
stops at `cd: /home/build/NanoKVM/server: No such file or directory`. From WSL the
path is real and the whole thing works.

```shell
wsl.exe -e sh -c 'cd /mnt/d/projects/NanoKVM && docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD:$PWD" -w "$PWD" \
  -e BUILD_UID=<uid> -e BUILD_GID=<gid> \
  ironkvm-release-host tools/release/release.sh --dry-run 1.0.1'
```

- [x] **Step 2: Flash the A-Data card**

Use the A-Data card, not the card in the board. Flashing the working card removes
the only way back if the board does not come up, and there is no remote power
cycle for this board.

- [x] **Step 3: Boot it and take the evidence**

```shell
ssh root@10.0.0.222 '
    echo "=== mounts ==="; grep mmcblk /proc/mounts
    echo "=== table ==="; parted -sm /dev/mmcblk0 unit s print
    echo "=== data ==="; df -h /data | tail -1; blkid /dev/mmcblk0p6
    echo "=== boot log ==="; grep -i "S01fs\|data partition" /bootlog 2>/dev/null | tail -20
'
```

All four must hold:

- `/proc/mounts` carries `/dev/mmcblk0p6` on `/data`
- the table reports partition 6 at start `10543104s`, reaching the last sector of
  the card
- `blkid` names `/dev/mmcblk0p6`
- `/boot`, root A and root B are at sectors 1, 40960 and 4235264

**Done, 2026-08-22.** The A-Data card booted and all four hold:

```
/dev/mmcblk0p2 /     ext4     /dev/mmcblk0p1 /boot vfat
/dev/mmcblk0p6 /data exfat    also bound over /etc/kvm and /root/.ssh
1:1s        2:40960s     3:4235264s     boot, root A, root B
4:8429568s:61071359s     extended, grown to the end of the card
6:10543104s:61071359s    data, 50528256 sectors, 24.1G free
blkid: LABEL="data" UUID="9DA7-0008"
/etc/nanokvm-slots.conf: DATA_START=10543104
```

The card is 61071360 sectors, so p6 reaches its last sector. `S01fs` made the
partition and the filesystem on the first boot, with no marker file.

Two faults came out of this step, and both are fixed:

- `/kvmapp/version` read `2.5.0`. The image never carried the release version,
  so a flashed board reported the official application and `semver.gte` told the
  update page it was newer than every IronKVM release. See
  `Merge fix/image-version-stamp`.
- `/data/identity-system/ssh` did not exist. The host keys were never stored,
  because `S02identity` runs 48 scripts before `sshd` makes them. See
  `Merge fix/host-key-persistence`.

- [x] **Step 4: Prove the identity survives**

Set a password in the web UI, reboot, and log in with it. This is what proves
`/data` is genuinely mounted rather than a directory on the root filesystem that
happens to exist. A board that reports a mounted `/data` and loses the password
across a reboot has `S02identity` binding nothing.

**Done, 2026-08-22.** The operator set a password in the web UI. The server's
write-back ran on its own: `/etc/kvm/pwd`, `/etc/shadow` and
`/data/identity-system/shadow` were all stamped at the same second, with nothing
run by hand. A reboot then returned every part of the identity byte for byte:

```
                        before      after
/etc/shadow             9a6467bc    9a6467bc
/etc/kvm/pwd            94391245    94391245
authorized_keys         ef5c4c46    ef5c4c46
ssh_host_ed25519_key    6e56d1f3    6e56d1f3
ssh_host_rsa_key        7e57347c    7e57347c
```

The web UI takes the new password, and `S02identity` reported `rc=0` in 3
seconds. A key appended to `/root/.ssh/authorized_keys` appeared under
`/data/identity-system/root-ssh` with no further command, which is the bind
this design turns on.

The host keys are in that list only because they were stored by hand first. On
the flashed image they were not stored at all: see Step 3.

- [ ] **Step 5: Publish**

The 1.0.1 artifacts built on 2026-08-18 must not be published. Both faults found
in Step 3 are in that image. Rebuild it from a tree that carries both fixes, and
check `/kvmapp/version` in the built root filesystem before publishing anything.

Only after all of the above. Note in the release that the 1.0.0 card image is
superseded and why: a board flashed with it has no data partition and keeps the
factory root password.

---

## Self-review

**Spec coverage.**

| Spec section | Task |
| ------------ | ---- |
| The shipped table | 2 |
| Where the data partition starts | 1, 4 |
| The first boot, four steps | 5, 6, 7 |
| Reading the result back | 5 |
| Deciding whether to format | 5 |
| S01fs stops reporting success it did not have | 7 |
| Failure review, every row | 5, 6, 8 |
| Testing, `test-partition.sh` | 2 |
| Testing, `test-build-card.sh` | 3 |
| Testing, new unit tests | 5, 6 |
| Testing, integration test and the release host | 9 |
| Testing, mutation | 8 |
| The release gate | Release gate |
| Out of scope, no change to the running board | 7 Step 4 says not to install it |

The image size claim in the spec says 5.02 GiB and the table ends at 10534912,
while the image is built to 10543104 because it carries the 4 MiB EBR gap. Both
round to 5.02 GiB. Task 3 states the reason the gap is carried, which the spec
does not: a card that previously held another layout must not leave a stale EBR
where the new chain could reach it.

**Placeholder scan.** No `TBD`, no "add error handling", no "similar to Task N".
Every code step carries the code. Task 4 Step 1 says to match the surrounding
cases for the unpack path in `test-build-image.sh`, which is a real instruction to
read one line of the file rather than a placeholder.

**Name consistency.** `data_start`, `disk_sectors`, `part_end_sector`,
`part_size_sectors`, `has_filesystem`, `data_partition_ready`,
`container_reaches_end`, `provision_data` are spelled identically in Tasks 5, 6,
7, 8 and 9. The overrides `DISK`, `PARTED`, `BLKID`, `PARTITIONS`, `MKFS_EXFAT`
are declared in Task 5 and Task 6 and used with those names by every test.
`data-start.sh` takes an optional table path and prints one integer, and Tasks 2,
3, 4 and 9 all call it that way.
