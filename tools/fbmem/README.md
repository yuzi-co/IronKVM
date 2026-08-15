# fbmem

Return the framebuffer reservation to Linux. About 8 MB, on a board that has
158 MB and no display output.

## What is reserved

The device tree carves out a region for a framebuffer and then nothing uses it:

```
/proc/device-tree/reserved-memory/cvifb   size 0x7d0000 (8000 KiB) at 0x8ab30000, status "okay"
/sys/bus/platform/devices/a088000.cvifb   present, no driver bound
/dev/fb*                                  absent
/sys/class/graphics/                      empty
```

Measured on a NanoKVM, hardware revision beta, kernel 5.10.4. The region is a
*dynamic* reservation: it carries `size` and `alloc-ranges` rather than a fixed
`reg`, and the top level `cvifb` node refers to it by phandle.

It costs the full 8000 KiB whether or not anything binds to it, because the
reservation happens in early boot from the flattened device tree, long before
any driver is probed.

## Why it is free to take

The region sits directly below the ION base at `0x8b300000`. Removing it moves
no other address, and the 75 MB capture carveout keeps both its base and its
size. There is no display on any NanoKVM model, so nothing is lost.

## What the tool changes

`repack-fdt.sh` sets `status = "disabled"` on two nodes and nothing else:

| node | what it is |
| --- | --- |
| `/reserved-memory/cvifb` | the reservation |
| `/cvifb` | the `cvitek,fb` device that consumes it |

The kernel honours that. `__fdt_scan_reserved_mem()` in
`linux_5.10/drivers/of/fdt.c` returns early for a node that is not available,
before it reserves anything:

```c
if (!of_fdt_device_is_available(initial_boot_params, node))
        return 0;

err = __reserved_mem_reserve_reg(node, uname);
```

Disabling is better than deleting here. `/cvifb` refers to the reservation by
phandle, so removing the node would leave that phandle dangling. Two property
values can also be read off a diff and reversed by hand.

## Why not fix it at the source

Upstream does, in [sipeed/LicheeRV-Nano-Build#836](https://github.com/sipeed/LicheeRV-Nano-Build/pull/836):
`memmap.py` sets `FRAMEBUFFER_SIZE` to zero, so the region is never described.
That is the better fix, and it is the one to prefer for anyone who builds the
image. This fork does not build the image. It edits the one that ships.

## Build

Run it in a throwaway container. Nothing here runs on the device.

```shell
docker run --rm -v "$PWD:/w" -w /w debian:forky-slim sh -c '
  apt-get update -qq && apt-get install -y -qq u-boot-tools device-tree-compiler
  tools/fbmem/repack-fdt.sh /w/boot.sd /w/build'
```

The tool refuses to emit an image unless the kernel and the ramdisk come back
byte-identical, the two nodes read back as disabled, the ION reservation and
the memory node are untouched, exactly two lines differ in the decompiled
source, and the FIT structure matches. On any failure it deletes its own output
so that the image cannot be installed by accident.

It also refuses an image that already carries the change.

```shell
sh tools/fbmem/test-repack-fdt.sh /w/boot.sd /tmp/work
```

The test extracts all three images from the built file and compares them
against the input, rather than trusting what the tool reported about itself.

## Install

`boot.sd` is 11.5 MB and `/boot` is a 16 MB FAT partition with about 4.5 MB
free, so the image is written in place. The installer is the one the slot work
already uses:

```shell
tools/slots/device/install-boot.sh /data/boot.sd.new <sha256>
reboot
```

It hash-checks the staged file, saves a stock image to `/data/boot.sd.orig` the
first time it runs, checks the growth against the free space, writes with
`dd conv=notrunc`, and reads the bytes back.

## Verify after the reboot

```shell
head -1 /proc/meminfo                                  # MemTotal, about 8000 kB higher
cat /proc/device-tree/reserved-memory/cvifb/status     # disabled
dmesg | grep "Ion: Ion memory setup"                   # still 75 MiB at 0x8b300000
```

If `MemTotal` does not move, u-boot is re-adding the reservation when it fixes
up the device tree, and the change has to move into u-boot instead.

## Roll back

Install the image you were running before and reboot. `install-boot.sh` writes
`/data/boot.sd.orig` only once, so it always holds the image that was there
before the first install and cannot be overwritten by a later one.

Keep the image you are replacing as well when it is not the stock one. A slot
build carries a patched `/init`, and `/data/boot.sd.orig` predates it.

If the board will not boot far enough to run the installer, the recovery
section in `tools/README.md` covers exposing the SD card as USB mass storage.
