# ionmem

Shrink the ION carveout in the boot image and give the difference to Linux.

The board reserves 75MB at `/reserved-memory/ion` for the video pipeline. The region
is not CMA, so none of it comes back when capture is idle. Whatever the reservation
says is gone from Linux for the whole uptime, and on a board with 166MB that Linux can
address, it is the largest single claim on the memory this fork runs short of.

The size is a device tree property, not a build constant, so this is a boot image edit
rather than an SDK rebuild. The node is dynamic: it carries `size` and no fixed `reg`,
so the kernel places the region and only the length changes.

## What the pipeline needs

Measured on the device 2026-08-20 at 1920x1080 60fps, which is the LT6911 sensor's
ceiling, with MJPEG, H.264 direct and WebRTC each exercised from a browser:

| state                         | bytes      |
| ----------------------------- | ---------- |
| peak, every delivery path      | 42,942,464 |
| idle, capture up               | 19,050,496 |
| one orphaned generation        |  6,516,736 |

WebRTC adds nothing over H.264 direct. They share the `VENC_1_*` encoder buffers, and
the peak did not move when the mode changed.

A crash still orphans a generation, because a process that dies runs no teardown. A
clean stop does not, since the teardown was fixed on the same day: see the capture
teardown section in `tools/README.md`.

## Sizes

`resize-ion.sh` refuses anything below **49,459,200** bytes, which is the peak plus one
orphaned generation. The refusal is hard rather than a warning. An undersized carveout
does not degrade: libkvm never checks an allocation result, so the server takes a
SIGSEGV on a NULL and dies before it binds its port, and the board answers nothing.

This fork installs **56MB** (`0x03800000`). It returns 19,922,944 bytes to Linux and
leaves 15.8MB over the measured peak, or 8.8MB if a crash orphans a generation before
anyone reboots.

48MB looks tempting and is not safe: about 5MB over the peak, which one orphan spends.

## Use

```shell
docker run --rm -v "$PWD:/repo" -w /repo ironkvm-release-host \
  sh tools/ionmem/resize-ion.sh /repo/boot.sd /repo/out 58720256
```

Then on the device:

```shell
tools/slots/device/install-boot.sh /data/boot.sd.new <sha256>
reboot
```

The tool verifies before it writes anything: the kernel and the ramdisk must come back
byte-identical, the device tree must survive the FIT round trip, exactly one line may
differ in the decompiled source, and the framebuffer reservation this fork already
disables must be untouched. A failed check removes the output so it cannot be
installed.

`test-resize-ion.sh` covers the edit and the five refusals. Both need `u-boot-tools`
and `device-tree-compiler`, so they exit 2 outside the release host image.

## Going back

There is **one** boot image on this board, not one per slot, so nothing reverts this
automatically.

`install-boot.sh` keeps the first image it ever replaced at `/data/boot.sd.orig`. That
one predates the framebuffer change, so it is the wrong target for undoing this alone.
Copy the running image aside before installing, as this fork did with
`/data/boot.sd.pre-ion-20260820`, and install that to revert.

The realistic failure is a kernel that boots and a server that cannot allocate, which
leaves ssh up and the image restorable over the network. A kernel that will not boot
needs hands on the device, and there is no remote power cycle here.

## Measured after the change

Installed 2026-08-20:

```
/proc/device-tree/reserved-memory/ion/size   0x03800000
MemTotal                                     189,244 kB   (was 169,788 kB)
carveout total                               58,720,256
idle with capture up                         19,050,496   (unchanged)
server up                                    37s after reset, no allocation errors
```
