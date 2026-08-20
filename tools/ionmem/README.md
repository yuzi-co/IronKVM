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
clean stop releases the VI pools and the ISP buffer, since the teardown was fixed on
the same day: see the capture teardown section in `tools/README.md`.

It did not release the H.264 encoder until 2026-08-20. A restart taken while a
browser was streaming stranded 11,636,736 bytes: four `VENC_1_ReconFrameBuf` where a
single generation holds two, and a second copy of `VCODEC_H264_FW_Buffer`,
`VENC_1_BitStreamBuffer`, `VENC_1_H264_WorkBuffer`, `ISP_SHARED_BUFFER_0` and one
`VbPool`.

The cause was a reference count, not the encoder. `mmf_deinit` is
`mmf_try_deinit(false)`: it decrements `mmf_used_cnt` and tears the pipeline down
only on the call that reaches zero. `_mmf_deinit` is that teardown, and it is the
only caller of `mmf_del_venc_channel_all`. MJPEG is served through
`Image::to_jpeg`, which reaches `mmf_enc_jpg_init`, which takes a reference that
nothing gives back. So one served JPEG frame left the count permanently one too
high, every later stop decremented to one, and the encoder stayed allocated.
`kvmv_deinit` calls `mmf_try_deinit(true)` now.

Measured with `tools/vidiag/venc-leak.cpp`, which drives the same sequence with
synthetic frames and needs no HDMI signal:

| library             | start      | peak       | after the stop | stranded    |
| ------------------- | ---------- | ---------- | -------------- | ----------- |
| before (`c268de5f`) | 25,567,232 | 49,459,200 | 43,237,376     | +17,670,144 |
| after (`1f4a7955`)  | 43,237,376 | 73,646,080 | 43,237,376     |           0 |

Both runs take the same 23,891,968 bytes. The second starts higher because it runs
on the board the first one left. Its own reading of the start is 49,754,112, which
is 43,237,376 plus the 6,516,736 the library allocates while it loads: `kvm_vision`
constructs its camera in a global, so the pipeline is up before `main` runs.

A crash still strands what the process held. Nothing in the kernel returns it:
`osdrv/interdrv/v2/vcodec` builds with `-DCVI_H26X_USE_ION_FW_BUFFER`, which compiles
`vpu_free_buffers()` out of `vpu_release()`, and `soph_sys` drops its bindings when
its file descriptor closes but never walks its buffer list. The memory stays
allocated until the board reboots.

Reclaiming an orphan from the next process is not a way out, and trying it panics
the board. The user-space free path is `SYS_ION_FREE`, which reaches `_sys_ion_free`,
which dereferences `mem_info.dmabuf`. Buffers the vc driver allocates come from
`sys_ion_alloc_nofd`, which stores a null there, and no ioctl reaches
`_sys_ion_free_nofd`. `_free_leak_memory_of_ion` in `kvm_mmf.cpp` is right to
reclaim `VI_DMA_BUF` alone: that one is allocated from user space, in
`middleware/v2/modules/vpu/src/cvi_vi.c`.

## Sizes

`resize-ion.sh` refuses anything below **49,459,200** bytes, which is the peak plus one
orphaned generation. The refusal is hard rather than a warning. An undersized carveout
does not degrade: libkvm never checks an allocation result, so the server takes a
SIGSEGV on a NULL and dies before it binds its port, and the board answers nothing.

**Nothing is installed at present.** 56MB was reverted while the encoder leaked. The
leak is fixed, so the question is open again, but it needs a fresh measurement on a
board that rebooted after the fix rather than a re-reading of the numbers below.

56MB was installed on 2026-08-20 and reverted the same day. It returns 19,922,944
bytes and leaves 15.8MB over the measured peak, which is enough room for the peak
plus exactly one restart taken while streaming:

    peak                    42,942,464
    one encoder orphan      11,636,736
                          = 54,579,200   =  93% of 56MB

A second such restart exhausts it. Deploying is a normal operator action, so that is
not a corner case. On 75MB the same sequence leaves 24MB spare.

48MB was never safe: about 5MB over the peak, which one idle orphan spends.

**What has to be true before this goes back on.** A restart taken while streaming has
to cost what an idle one costs, which is nothing. That holds since 2026-08-20. What is
still missing is the peak: every figure above was measured on a board carrying orphans
from the leak, so the peak has to be measured again from a clean boot, and the size
chosen against that. Budget for one crash, not one restart.

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

## Measured, installed, and reverted

Installed 2026-08-20 at 56MB. It worked exactly as designed:

```
/proc/device-tree/reserved-memory/ion/size   0x03800000
MemTotal                                     189,244 kB   (was 169,788 kB)
carveout total                               58,720,256
idle with capture up                         19,050,496   unchanged
peak, every path re-exercised                42,942,464   unchanged, 74% of the new total
server up                                    37s after reset, no allocation errors
```

Reverted the same day, from `/data/boot.sd.pre-ion-20260820`, after a restart taken
while streaming left an orphaned encoder set and took the board to 93% with its free
space in two fragments. The reservation was not the problem and the measurement was
not wrong. The input was incomplete: every "a stop returns what it took" reading
behind the 56MB decision was taken on an idle pipeline, which holds no encoder
buffers, so none had to be released for the reading to come out clean.
