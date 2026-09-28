# Ventoy Without Copies

**Date:** 2026-09-28
**Issue:** yuzi-co/ironkvm-dist#31
**Repos:** IronKVM for the code. ironkvm-dist for two kernel options, which this branch does not
change: the options are listed below and go into a combined image.

## Purpose

The host boots a Ventoy menu from the virtual disk and picks any of the selected images on
`/data`. No image is copied. The disk is a device-mapper device that the kernel reads from the
images in place, so `/data` grows only by a generated head file and Ventoy's 32 MB EFI partition.

## The disk

The virtual disk is `/dev/mapper/ventoy`, one `dm-linear` table over read-only loop devices:

| Disk sectors                        | Source                                                   |
| ----------------------------------- | -------------------------------------------------------- |
| 0                                   | head: MBR, with Ventoy's `boot.img` and our partition table |
| 1 to 2047                           | head: Ventoy's `core.img`                                 |
| 2048 to the first file's cluster    | head: the generated exFAT metadata of partition 1         |
| each file's clusters                | the file's loop device, then the tail and padding         |
| the end of partition 1              | padding                                                   |
| `part2` to `part2 + 65535`          | `vtoyefi.bin`: Ventoy's `ventoy.disk.img`, 32 MB          |

A file's clusters hold its whole 512-byte sectors, taken from its loop device. A loop device ends
at the last whole sector of its file, so the partial last sector of a file whose size is not a
multiple of 512 is copied into the head, one sector per such file. Padding up to the end of a
cluster, and at the end of partition 1, is mapped to one run of zero sectors in the head. Every
padding extent points at that same run, so the head holds it once. This avoids the `zero` target
and its kernel option.

The head file, `vtoyefi.bin` and every image are attached with the loop driver's ioctls from Go,
read-only and with autoclear set. The server holds the loop devices open until the table is
loaded, and device-mapper then holds them. Removing the device-mapper device frees them.

### Partition layout, as Ventoy 1.1.17 makes it

Read from `Ventoy2Disk.sh`, `tool/VentoyWorker.sh` and `tool/ventoy_lib.sh` of the release,
MBR style and Secure Boot support on, which are the defaults:

- Partition 1 starts at sector 2048, type `0x07`, active (`0x80`).
- Partition 2 is 65536 sectors (32 MB), type `0xEF`, not active, and starts right after
  partition 1. `format_ventoy_disk_mbr` moves its start down to a multiple of 8 sectors and ends
  partition 1 one sector before it.
- The first 446 bytes of the MBR are `boot/boot.img`. Bytes 384 to 399 are a random disk UUID and
  bytes 440 to 443 a random disk signature.
- `boot/core.img` (decompressed, 2047 sectors) is written from sector 1.
- `ventoy/ventoy.disk.img` (decompressed) is partition 2.
- Partition 1 is exFAT, labelled `Ventoy`, with 32 KB clusters up to 32 GB and 128 KB above.

We size the disk ourselves: partition 1 is exactly as large as its metadata and files need,
rounded so partition 2 starts on a multiple of 8 sectors, and the disk ends where partition 2
does. The UUID, the signature and the exFAT serial number are drawn once and kept in the state
file, so the host sees one disk across rebuilds.

The CHS fields use 255 heads and 63 sectors per track and saturate at 1023/254/63, as fdisk
writes them.

### exFAT

The generator follows Microsoft's exFAT specification. Given the files as (name, size, modified
time), it lays out:

- The main and backup boot regions, 12 sectors each: the boot sector, eight extended boot sectors,
  the OEM parameters, a reserved sector and the checksum sector. `PartitionOffset` is 2048.
- One FAT at sector 128, with a chain for every allocation. Files are contiguous and their stream
  extensions set `NoFatChain`; their chains are written as well, so the FAT agrees.
- The cluster heap, aligned to the cluster size: the allocation bitmap, the up-case table (the
  specification's recommended table in its compressed form, 5836 bytes, checksum `0xE619D30D`),
  the root directory, then the files in order.
- The root directory: the volume label, the bitmap and up-case entries, and one file entry set per
  file (file, stream extension, name entries), with the name hash and the set checksum.

`ClusterCount` is exactly what the volume needs, so the bitmap is full and `PercentInUse` is 100,
except that a volume is never smaller than the 1 MB the specification requires. Names are the
images' base names in UTF-16. `/data` is exFAT itself, so every name is already a legal exFAT
name. Two images with one base name in different directories get ` (2)`, ` (3)` before the
extension, compared case-insensitively through the up-case table.

## On the board

- **Files.** `/data/ironkvm/ventoy/`: `boot.bin`, `core.bin` and `vtoyefi.bin` from the pinned
  release (`boot.img`, `core.img` and `ventoy.disk.img` there), `head.bin`, and `state.json` (the
  selected images, the disk UUID, the signature, the serial). None has an `.img` or `.iso` suffix,
  because the image list walks all of `/data` and would show them.
- **Release.** Ventoy 1.1.17, `ventoy-1.1.17-linux.tar.gz` from the GitHub release, SHA-256
  `7fb4ed08…bc43805`. Downloaded on demand to `/data/ironkvm/ventoy/.fetch`, checked, and the three
  members taken out and checked against their own pinned sums after `xzcat`. The archive is removed
  afterwards. Ventoy is GPLv3 and is never shipped in the image.
- **Device-mapper.** The server issues the DM ioctls itself through `golang.org/x/sys/unix`: create,
  load a read-only table, resume, and remove. No `dmsetup`, so no add-on. It then makes the
  block node `/dev/mapper/ventoy` itself, as `dmsetup` does without udev, because the mass storage
  function reports the path it was given, and a symlink would read back as `/dev/dm-0`.
- **Kernel check.** Device-mapper is present when `/proc/misc` lists `device-mapper`. Without it the
  status says so and building is refused.
- **The drive.** `storage.InsertDevice` puts the device into lun.0 read-only, through the same lock
  and eject as any insert. Only a node under `/dev/mapper/` is accepted, and only the server calls
  it.
- **Guards.**
  - The set, the release and the head change only while the Ventoy disk is in no drive.
  - `storage.RegisterDevice` names the Ventoy device and a function listing its images. The delete
    guard refuses to delete an image that is on the Ventoy disk while that disk is in a drive,
    as it does for an image in a drive.
  - When the disk drive lets go of the device, by an eject or by another insert, storage calls the
    device's release hook after unlocking, and the Ventoy service removes the device-mapper device.
  - "Use as virtual disk" always rebuilds the head and the table, so an image deleted or replaced
    since the last build is never served stale.
- **Server restart.** The table lives in the kernel, so a restart does not disturb a host that is
  reading. At start the service adopts a `ventoy` device that is in a drive, and removes one that
  is not.

## API

All under `/api/ventoy`, behind the session check and the admin role, with the usual envelope.
Code -2 is a refusal because of the board's state, and its message is shown as it is; -3 is a
failure on the way.

- `GET /status`:
  `{kernel, onData, installed, version, images: [path], missing: [path], inDrive, device, size}`.
  `size` is the disk size in bytes of the device in the drive, or 0.
- `POST /install`, `POST /uninstall`: fetch or remove the release files.
- `POST /images` `{images: [path]}`: replace the set. Refused while the disk is in a drive.
- `POST /insert`: build and insert into the disk drive, read-only.
- `POST /eject`: eject from the disk drive and remove the device.

## Web

The image manager gets a Ventoy section below the image list: the status line, a switch per image
("on the Ventoy disk"), the install button when the release is missing, "Use as virtual disk" or
"Eject", the kernel note when device-mapper is missing, and the Secure Boot note: the host has to
enroll Ventoy's key in MokManager once. The drive list shows the device as "Ventoy".

## Kernel

`CONFIG_MD=y` and `CONFIG_BLK_DEV_DM=y`. `dm-linear` is part of the core module, `dm-mod`.
`CONFIG_DM_ZERO` is not needed, because padding comes from the head.

## Tests

- Unit tests for the exFAT structures, the checksums, the name hash, the MBR and the extent map.
- An image test, behind the `ventoyimage` build tag, in a privileged container: assemble the image
  by concatenation from sparse files, run `fsck.exfat`, mount it and compare every file. Names
  that need UTF-16, a zero-length file, a file over 4 GB, and many files.
- The same test through the real assembly: loop devices and the DM ioctls, on the container's
  kernel, which has device-mapper.
- A QEMU check with SeaBIOS and OVMF against the assembled image.

## Out of scope

- Ventoy's persistence and a `ventoy/ventoy.json` on the drive: the disk is read-only. A generated
  `ventoy.json` in the synthetic filesystem could come later.
- GPT style.
- Files other than those in the image list (`.iso`, `.img`).
