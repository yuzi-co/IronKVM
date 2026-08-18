# A card image that fits a small card, and a data partition made on first boot

Status: design, approved 2026-08-18. Target release 1.0.1.

## Goal

The card image must install on a microSD card of 8 GB, and it must give the data
partition all of the space that is left on whatever card it lands on.

A second goal comes out of the investigation and is more urgent than the first.
The 1.0.0 card image makes no data partition at all. The same code fixes both,
because a data partition can only be made once, so the size it is made at is the
size it keeps.

## What is broken today

The card image declares a data partition on p6 and carries no bytes for it.
Nothing then creates a filesystem there.

`build-card.sh` truncates the image at the first sector of p6. The only
`mkfs.exfat` in the device tree is in `S01fs`, it names `/dev/mmcblk0p3`, and two
separate gates stop it on an A/B card:

- `may_autopartition` returns false when `/etc/nanokvm-slots.conf` is readable,
  and `build-image.sh` writes that file into every image.
- `root.manifest` runs `touch /etc/kvm.disk0`, which is the stock gate for the
  same branch.

Both gates are correct. The branch formats p3, and p3 is root B in this layout.
The fault is that no replacement was ever written.

`S01fs` then hides the result. It discards the return value of `mount_data` and
prints `OK`:

```sh
mount_data "$DATADEV" /data
printf "(data on %s) " "$DATADEV"
echo "OK"
```

So a board that flashes the 1.0.0 image comes up with no `/data`, with the
factory root password, and with a new ssh host key, and it reports a clean boot.
`S02identity` depends on `/data`, and its own header records that this exact loss
happened once before.

`README.md` states that the first boot makes the data partition. It does not.
`build-card.sh` and `docs/plans/2026-08-16-ironkvm-1.0.md` state the same thing.
All three are wrong. `root.manifest` states the opposite and is right.

Nobody saw the fault because no card was ever flashed with the image. The board
in use has a data partition that was made by hand during the A/B migration.

## What the device can do

Measured on the board, not assumed.

| Tool | State |
| ---- | ----- |
| `parted` | present, GNU parted 3.6 |
| `mkfs.exfat` | present, exfatprogs 1.2.2 |
| `fsck.exfat` | present |
| `blkid` | busybox |
| `losetup` | busybox, supports `-P` |
| `resize.exfat` | **absent** |
| `sfdisk` | absent |
| `partx` | absent |
| `blockdev` | absent |
| `udevadm` | absent |

`resize.exfat` is absent, so an exfat filesystem cannot grow. The filesystem must
be made once, at the size it keeps. Therefore the partition must reach its final
size before the filesystem is made, and a design that ships a small filesystem
and grows it later is not possible on this device.

### busybox blkid prints no filesystem type

```
/dev/mmcblk0p1: LABEL="boot" UUID="8990-56F2"
/dev/mmcblk0p2: LABEL="roota" UUID="5a0a2fde-815c-4c3d-a33d-3b53da166d40"
/dev/mmcblk0p6: LABEL="data" UUID="EE8B-6CB5"
/dev/mmcblk0p3: (no output, exit status 0)
```

No line carries a `TYPE=` field, for any filesystem. An empty partition gives no
output and exit status 0. So a test on `TYPE=` reports "no filesystem" for a
healthy data partition, and a test on the exit status reports nothing at all.
The only valid test is whether `blkid` printed anything.

### parted reports failure when it succeeds

On a disk with a mounted partition, parted cannot make the kernel re-read the
whole table. It writes the table, it adds the partition through `BLKPG`, and it
exits 1. Measured against a loop device on the board:

```
resizepart 3 100%        rc=1
mkpart logical ntfs ...  rc=1
/proc/partitions         259 0 20480 loop0p6      <- the partition is there
```

So the exit status must be ignored, and the result must be read back.

### The kernel exposes the new partition without a reboot

The same measurement shows the new node present while another partition of the
same disk is mounted. The first boot needs no reboot.

### parted needs two calls, in order

`mkpart` on its own fails with "Can't have overlapping partitions", because the
extended container does not yet reach the new partition. `resizepart 4 100%`
must run first.

`mkpart` also needs a filesystem argument. Without `ntfs` the partition gets type
`83` instead of type `7`, and a card pulled from the board is then not recognised
by Windows or macOS.

### The MBR write changes one entry

parted rewrites sector 0. Compare the 64 bytes of partition entries before and
after, on four card sizes:

```
entry 1   80 000200 0c 0a0902 01000000 00800000   identical
entry 2   00 8c0b02 83 a15a07 00a00000 00400000   identical
entry 3   00 a15b07 83 b6aa0c 00a04000 00400000   identical
entry 4   00 b6ab0c 05 c3b48f 00a08000 00202000   before
          00 b6ab0c 05 50cac6 00a08000 00406c00   after
```

Only the entry for the extended container changes, and within it only the CHS end
field and the size field. The entries for `/boot`, root A and root B come out bit
for bit identical.

### The result matches the card in use

On a 60506112 sector card the two parted calls give:

```
6 : start=10543104, size=49963008, type=7
```

That is what the fixed table declares today. A card made the new way and a card
made the old way have the same layout.

## The design

### The shipped table

`tools/abslots/partition.sfdisk` loses its last line, and p4 shrinks to hold only
p5:

```
1 : start=1,        size=32768,    type=c, bootable
2 : start=40960,    size=4194304,  type=83
3 : start=4235264,  size=4194304,  type=83
4 : start=8429568,  size=2105344,  type=5
5 : start=8437760,  size=2097152,  type=83
```

The table ends at sector 10534912. `build-card.sh` truncates the image there, so
the image is 5.02 GiB and it needs a card of 10534912 sectors or more. Any card
of 8 GB or more holds it and leaves space for data.

The compressed download does not change. It is 372 MB today and it stays 372 MB,
because the 2 GiB of slot B is already zeros and already compresses away. What
changes is the card that accepts the image, not the bytes that are fetched. State
this in `README.md`, because "a smaller image" reads as "a smaller download".

### Where the data partition starts

The start sector is derived, so that no number is written down twice:

```
DATA_START = p5_start + p5_size + 8192
```

8192 sectors is the 4 MiB gap the table already leaves before a logical
partition, for its EBR. The rule gives 10543104, which is the value the fixed
table uses today. `build-image.sh` computes it from `partition.sfdisk` and writes
it into `/etc/nanokvm-slots.conf` beside `DATA_DEV`.

### The first boot

Four steps in `S01fs`, before the data partition is mounted. Each step has a gate
that it satisfies by running, so no step repeats.

```
grow the container   p4 does not end at the last sector of the disk
                     -> parted -s /dev/mmcblk0 resizepart 4 100%

make the partition   p6 is absent
                     -> parted -s /dev/mmcblk0 mkpart logical ntfs 10543104s 100%
                     -> read the result back, and stop if it is not there

make the filesystem  blkid prints nothing for p6
                     -> mkfs.exfat -L data /dev/mmcblk0p6

mount it             the existing mount_data, which now reports failure
```

No marker file is used. Every gate reads the partition table. `S01fs` already
argues for this in its resize guard: a marker in the root filesystem does not
survive a slot image built from a fresh rootfs, and the partition table is the
same for every slot on the card.

The gates are also what bounds the retries. A `mkpart` that fails writes nothing,
which is measured, so a card that cannot be partitioned does not wear its own
sector 0 on every boot. A `resizepart` that has already run is skipped, because
p4 then ends at the last sector of the disk.

The gates are correct for the board in use. Its p4 already ends at the last
sector, its p6 is present, and `blkid` prints a line for it. All three gates say
no, and nothing is written.

### Reading the result back

Every geometry value comes from the table on the card, read through the
machine-readable output of parted. Field 2 of the disk line is the size of the
disk in sectors. Fields 2, 3 and 4 of a partition line are its start, its end and
its size.

```sh
part_size_sectors() {
    parted -sm "$DISK" unit s print 2>/dev/null \
        | awk -F: -v p="$1" '$1 == p {gsub(/[^0-9]/, "", $4); print $4}'
}

data_partition_ready() {
    [ -e "$DATA_DEV" ] || return 1
    blocks=$(awk -v n="$(basename "$DATA_DEV")" '$4 == n {print $3}' /proc/partitions)
    [ -n "$blocks" ] || return 1
    [ "$((blocks * 2))" = "$(part_size_sectors 6)" ]
}
```

`/proc/partitions` counts 1024 byte blocks, so a sector count is twice it. The
filesystem is made only when the kernel and the table agree on the size. A
filesystem made on a stale node would be the wrong size, and it could not be
repaired, because `resize.exfat` does not exist.

### Deciding whether to format

```sh
has_filesystem() {
    [ -n "$(blkid "$1" 2>/dev/null)" ]
}
```

A partition that `blkid` can name is never touched. This is what protects a
re-flash: an operator who writes a new image and keeps the data partition keeps
the identity on it. A partition that `blkid` cannot name is either new, or it
holds nothing this board can read, and it is formatted.

### S01fs stops reporting success it did not have

```sh
if mount_data "$DATADEV" /data; then
        printf "(data on %s) " "$DATADEV"
else
        printf "(FAILED to mount %s) " "$DATADEV"
fi
```

## Failure review

| Failure | Handling |
| ------- | -------- |
| No data partition on a fresh flash | This design. It is the reason for the release |
| parted exit status read as truth | The result is read back from `/proc/partitions` |
| `blkid` tested for `TYPE=` | The test is whether `blkid` printed anything |
| `mkpart` without a filesystem argument | `ntfs` is passed, so the type is 7 |
| `mkpart` without an explicit start | The start is passed, so every card matches |
| The MBR write damages p1, p2 or p3 | Measured bit identical. Only p4 changes |
| Power is lost during the sector 0 write | Accepted. See below |
| The watchdog rolls back a slow first boot | The deadline is 300s. `mkfs.exfat` takes seconds |
| A card too small for a useful data partition | `mkpart` succeeds and makes a small partition. The size is logged |
| `udevadm: not found` on every parted call | Expected. The board runs mdev. It is noise, not a fault |
| The evidence comes from a loop device | Not closed. See the release gate |

### The power cut

The write to sector 0 cannot be avoided. A data partition has to be logical,
because p1, p2, p3 and p4 use all four primary entries, and a logical partition
needs the container to reach it. Three things limit the cost:

- The write happens once, on the first boot of a new card.
- The operator is present, with the card in hand, so a card that does not come up
  is written again.
- The entries for `/boot`, root A and root B are not changed by it.

Stock firmware makes the same write on every board it installs.

## Testing

- `test-partition.sh` applies the table to a 10534912 sector disk, checks that p1
  is untouched, that p4 ends where p5 ends, and that no p6 is declared.
- `test-build-card.sh` checks that the image is truncated at the last sector of
  p5, and that the table in the image declares five partitions.
- New unit tests drive each gate against a scratch tree, with stubs for `parted`,
  `blkid` and `/proc/partitions`. They cover: p4 already grown, p6 already
  present, `blkid` printing a line, `blkid` printing nothing, and a `parted` that
  exits 1 after it succeeded.
- An integration test loop mounts a built image in a container, runs the real
  block against the real `parted` and the real `mkfs.exfat`, and asserts that
  `/data` mounts. This is the test that would have found the fault in 1.0.0.
  `ironkvm-release-host` gains `parted` and `exfatprogs`, which it does not carry.
- Mutation tests over the new block, as the repository does for its other shell.

## The release gate

The mechanics are measured on this board's own parted, busybox and kernel, but
through a loop device. The write to the master boot record of the real boot
device is not tested.

1.0.1 is not published until a flashed card boots and shows:

- `/proc/mounts` carries `/dev/mmcblk0p6` on `/data`
- `parted -sm /dev/mmcblk0 unit s print` reports p6 at start 10543104, type 7
- a password set in the web UI survives a reboot

The card for this is the A-Data card, which has not arrived. The card in use is
not flashed for the test, because that removes the way back.

## Out of scope

- Growing the data partition of a card that already has one. `resize.exfat` does
  not exist on the device.
- Any change to the running board. Every gate reads false there.
- Any change to the update package. The package does not own the partition table.
