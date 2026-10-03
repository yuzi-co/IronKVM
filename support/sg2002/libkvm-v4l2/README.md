# libkvm-v4l2: libkvm.so for the SG2002 on the mainline kernel

`server/dl_lib/libkvm.so` is Sipeed's capture library. It drives the vendor kernel's MPI (VI,
VPSS, VENC) through `libkvm_mmf.so`, so it does not run on a mainline kernel. This directory
builds a replacement that implements the same `kvm_vision.h` ABI over V4L2, so NanoKVM-Server
runs unchanged on either kernel and only the library differs. Tracking issue:
yuzi-co/ironkvm-dist#37.

It is for the mainline image only. On the vendor kernel keep Sipeed's library.

## Pipeline

```
LT6911 subdev ── dv timings, source-change events
     │
sg2002-capture ──dma-buf──> sg2002-vpss ──dma-buf──> coda (Coda980) ──> H.264 Annex-B
 UYVY 1920x1080             UYVY -> NV12,            NV12 in               copied into the
                            optional downscale                             buffer the server gets
                     │
                     └─dma-buf──> sg2002-vpss ──> NV12, CPU ──> libjpeg-turbo ──> JPEG
                                  (2nd context)    (cached)      (kvmv_read_img MJPEG)
```

The order of operations follows nixos-nanokvm's `sg2002-h264-bridge` (`--scaler vpss --format
nv12`), which ran this chain on the board at about 51 fps at 1080p. The code here is written
for this repository. Unlike the bridge, the pipeline does not free run: each `kvmv_read_*` call
takes the newest captured frame, scales it, encodes it and returns its access unit, as the
vendor library does with its VI channel. Frames nobody reads go back to the capture driver
untouched.

Nodes are found by driver name and capability, never by number:

| Role | Match |
|------|-------|
| capture | `VIDIOC_QUERYCAP` driver `sg2002-capture`, video capture + streaming |
| scaler | driver `sg2002-vpss`, mem2mem |
| encoder | any mem2mem node whose CAPTURE queue offers H.264 and OUTPUT queue NV12 (the Coda's encoder node, not its decoder or JPEG nodes) |
| receiver | `/sys/class/video4linux/v4l-subdev*/name` containing `lt6911` |

Kernel requirements: the nixos-nanokvm 7.2 media patches (LT6911UXE DT probe and UXC polling,
SG2002 CSI capture with source-change events, VPSS with patch `0900`, Coda980 with NV12 and
dma-buf input), `coda-vpu` and `sg2002-vpss` loaded, `coda980.bin` installed, and a CMA
dma-heap. Without a `v4l-subdev` node the library still runs, but it cannot tell "no signal"
from "nothing captured" before a pipeline is up.

## The ABI, function by function

| Function | Here |
|----------|------|
| `kvmv_init` | Starts the monitor thread. Devices open lazily on the first read. `KVMV_DEBUG=1` or a non-zero argument enables debug logging. |
| `kvmv_read_video` | H.264: one access unit per call. `3` for a keyframe, `4` for a delta frame. Codec 2 (H.265) answers `-2`: the Coda980 encodes H.264 only. |
| `kvmv_read_img` | Type 1 as above. Type 0 (MJPEG): one software JPEG per call at the size asked, quality from the `_qlty` argument; answers `0`. See below. |
| `free_kvmv_data`, `free_all_kvmv_data` | As the vendor library: four reusable slots, a pointer stays valid until freed. |
| `set_h264_gop` | `V4L2_CID_MPEG_VIDEO_GOP_SIZE` at runtime, plus a forced keyframe, because the vendor library rebuilds its encoder here and the server relies on the next frame being a keyframe. |
| `set_h264_fps` | `VIDIOC_S_PARM` on the encoder's OUTPUT queue at runtime; returns early when unchanged. The rate given is the lower of this and `set_capture_fps`, or the measured delivery rate when frames arrive more than 10% slower (see Known issues). |
| bitrate argument | `V4L2_CID_MPEG_VIDEO_BITRATE` (kbit/s x 1000) at runtime. |
| `set_capture_fps` | The capture node has no frame interval control; unread frames cost their DMA and nothing else. Caps the rate the encoder is told, since frames cannot reach it faster. |
| `set_frame_detact` | Recorded only: every MJPEG read encodes a picture, and `5` ("not changed") is never answered. |
| `kvmv_codec_supported` | Not in `kvm_vision.h` or Sipeed's library (`abi-extensions.txt`). `1` for codecs 0 (MJPEG) and 1 (H.264), `0` for 2 (H.265). The server finds it with `dlsym` and, without it, assumes all three. |
| `set_venc_auto_recyc` | Recorded only. There is one encoder. |
| `kvmv_hdmi_control` | Stops and starts capture in software on every board: `0` tears the pipeline down and reads answer `-1`; `1` allows capture again. Answers `0`. The receiver is not powered down (the vendor library does that through a GPIO on the PCIe board only, and answers `-1` elsewhere). |
| `kvmv_hdmi_signal_active` | `1` when capture is enabled and the receiver has a source. Like the vendor library it reads `0` until `kvmv_hdmi_control(1)`. |
| `kvmv_deinit` | Joins the monitor thread, tears the pipeline down, frees the slots. `kvmv_init` may follow. |

Every runtime change the encoder refuses (an ioctl error) rebuilds the pipeline instead, which is
what the vendor library does for all of them. The encoder's actual bitrate, GOP and frame rate
are read back after every build and change and logged next to what was asked for.

### Return codes

| Code | When |
|------|------|
| `0` | An MJPEG picture. |
| `3`, `4` | A keyframe, a delta frame. |
| `-1` | No signal (receiver says no link or lock), no frame within 1 s, or capture switched off. |
| `-2` | H.265, video nodes missing, a pipeline that will not build, an encode error (H.264 or JPEG). |
| `-3` | All four slots are held by the server. |
| `-4` | The source changed mode under a running pipeline; the next read rebuilds. |
| `-5` | Another read held the lock for over a second. |
| `-6` | The source runs a mode the capture node cannot take (it is fixed at 1920x1080 progressive), or the capture link refused it at `STREAMON` (`EPIPE`). |
| `-7` | The receiver reports timings out of range (`ERANGE`). |

### Keyframes

The first picture after every pipeline build is a keyframe: the encoder is asked for one, and
delta frames before it are dropped (up to four attempts per read). Every keyframe returned
carries SPS and PPS: if the encoder emits an IDR without them, the last ones it produced are
put in front. The type is decided by scanning the access unit for NAL type 5, as the vendor
library does.

### MJPEG

There is no mainline driver for the SG2002's JPEG unit (yuzi-co/ironkvm-dist#36), so MJPEG is
encoded on the CPU with libjpeg-turbo 3.1.2's TurboJPEG API, linked statically: planar 4:2:0
from the NV12 picture (chroma deinterleaved, no RGB), fast integer DCT, quality from the
caller (1 to 100, 0 for 80). The choice and the numbers come from ironkvm-dist
`socs/sophgo-sg2002/mainline/jpeg-bench`: about 210 ms per picture at 1080p, 94 ms at 720p and
53 ms at 960x540, all of it on the board's only core.

So JPEG is encoded per `kvmv_read_img` call, never at the stream rate, and it is kept off the
H.264 path:

- A second context on the scaler node imports the same capture buffers and scales the newest
  frame to the size asked for, into one NV12 buffer of its own (cached, from the dma-heap). Its
  size does not touch the H.264 stream's, and the scaler's mem2mem queue orders its jobs with
  the H.264 ones. It is set up on the first MJPEG read, rebuilt when the size changes, and torn
  down with the pipeline.
- The pipeline lock is held only to take the picture: the capture wait, one scale, and the copy
  into the encoder's own planes (a few ms at 1080p). The encode runs after it is released, under
  a lock of its own, so an H.264 read waits at most for one scale and copy, not for a JPEG.
- With no H.264 stream running, an MJPEG read builds the pipeline at the size and bitrate the
  stream last used (capture runs inside it), and the idle timer tears it down 10 s after the
  last read of either kind. An H.264 read at a new size rebuilds it, and the next MJPEG read sets
  its scaler context up again.

The CPU cost is the limit: at 1080p a reader asking continuously takes the whole core. The
server's MJPEG consumers should ask at a low rate or a small size (screenshots, VNC at 960x540).

### Audio

`kvm_vision.h` has no audio functions; the server captures audio itself. Nothing to do here.

### Not ported

- The ION carveout check that reboots the board when the heap is full. Mainline has no ION.
- `/proc/lt_int`, the HDMI mode files under `/etc/kvm` and the resolution trial loop. The
  receiver driver reports timings directly.
- The VPSS wedge rebuild. A capture node that delivers nothing for three reads in a row is torn
  down and rebuilt from the receiver's answer instead.

## Side effects

The same files the vendor library maintains:

- `/tmp/kvm/state`: `1` or `0`, the signal state, written by temporary file and rename.
- `/kvmapp/kvm/width`, `/kvmapp/kvm/height`: the HDMI input size, when `/kvmapp/kvm` exists.
- `/tmp/nanokvm_wd`: touched every 500 ms while `/etc/kvm/watchdog` or `/tmp/watchdog` exists.

The monitor thread also asks the receiver for its timings once a second while no pipeline is
running, so the signal state stays current without a viewer, and tears down a pipeline nobody
has read for 10 s.

## Environment

| Variable | Default | |
|----------|---------|---|
| `KVMV_DEBUG` | off | Debug logging to stderr, including the measured output rate, the rate the encoder was told and the per-stage time of a read every 10 s. |
| `KVMV_CAPTURE_DEV`, `KVMV_SCALER_DEV`, `KVMV_ENCODER_DEV`, `KVMV_SUBDEV` | discovered | Override a node. |
| `KVMV_CAPTURE_BUFFERS` | 2 | Capture queue depth (4 MB each at 1080p UYVY). |
| `KVMV_MID_BUFFERS` | 2 | NV12 buffers between scaler and encoder. |
| `KVMV_BITSTREAM_BUFFERS` | 3 | Encoder output buffers. |
| `KVMV_IDLE_MS` | 10000 | Tear down an unread pipeline after this long; 0 never does. |

## Known issues

- Bitrate (yuzi-co/ironkvm-dist#35): the bridge produced about nine times its target. The bridge
  never set the encoder's frame rate; this library does, which may be part of the answer. The
  library measures its own output and logs once per pipeline when it runs above twice the
  target.
- Only 1920x1080 sources: the capture driver's DMA geometry is fixed. Downscaling for the stream
  works (the VPSS does it); other HDMI modes answer `-6`.
- Each read waits for the capture, the scaler and the encoder in turn; nothing overlaps, unlike
  the free-running bridge. A read costs the capture wait plus one scale plus one encode plus the
  copy out. With `KVMV_DEBUG=1` the library logs the average of each every 10 s ("per frame:
  capture wait, scale, encode, copy"), which is the measurement to take before pipelining.
- The encoder's bitstream buffers come from the Coda's no-map `shared-dma-pool`, so user space
  maps them uncached. The library copies each access unit out once, in 8-byte loads, and parses
  the copy. Before this it scanned the mapping byte by byte twice and then copied it, one DRAM
  round trip per byte each time, which the first board run (15.2 fps for 30 asked) paid on every
  frame.
- The encoder's frame rate (`VIDIOC_S_PARM`) follows what is delivered. With rate control on
  the Coda budgets bitrate / rate per frame, so telling it 30 while 15 arrive doubles every
  frame's share. The library tells it the lower of `set_h264_fps` and `set_capture_fps`, and
  the measured delivery rate (3 s window) when that falls more than 10% short.

## Build and test

From the repository root, in the app builder image (`tools/build/Dockerfile`):

```
docker run --rm -v "$PWD":/w -w /w/support/sg2002/libkvm-v4l2 nanokvm-app-builder \
    make all test check-symbols
```

The first build fetches libjpeg-turbo 3.1.2 (checked against a pinned SHA-256) and builds it
twice with `tools/build-libjpeg-turbo.sh` into `deps/`: for the board (static, PIC, hidden
symbols) and for the host test. That needs cmake, which the Dockerfile now installs; in an
image built before that, run `apt-get update && apt-get install -y cmake` first in the same
container. `make clean` keeps `deps/`; `make distclean` removes it.

- `build/libkvm.so`: soname `libkvm.so`, `NEEDED` `libgcc_s.so.1` and `libc.so`, no `RUNPATH`
  needed. It exports the 13 functions in `abi-symbols.txt`, all strong (`nm` type `T`), and
  `kvmv_codec_supported` (`abi-extensions.txt`), nothing else; libjpeg-turbo is inside it.
  Nothing in the library calls libgcc_s. The server needs it: its crtbegin calls
  `__register_frame_info` through the PLT, and the server NEEDs only `libkvm.so` and `libc.so`,
  relying on Sipeed's library (`NEEDED` OpenCV, `libkvm_mmf.so`, `libstdc++.so.6`,
  `libgcc_s.so.1`, `libc.so`, `libatomic.so.1`) to load it. Without it the server jumps to
  address 0 at start. Of everything the server leaves undefined, only `__register_frame_info`
  and `__deregister_frame_info` come from outside libc and libkvm.
- `build/kvmv-probe`: the on-board test below.
- `make test`: host unit tests with ASan and UBSan: Annex-B splitting and classification,
  parameter-set prefixing, SPS parsing (with emulation prevention), the policy decisions, the
  frame slots, format negotiation against mock capture, VPSS and Coda drivers, the encoder
  frame-rate policy, the uncached copy, and the JPEG path (NV12 split, encode, decode back).
- `make check-symbols`: the exports against `abi-symbols.txt`, `abi-extensions.txt` and
  Sipeed's library (any other export fails), and the `NEEDED` entries for `libgcc_s.so.1` and
  `libc.so`.

## Testing on the board

Stop the server first (`S95nanokvm stop`): it holds the same nodes. Copy
`build/libkvm.so` and `build/kvmv-probe` to one directory on the mainline image, with a 1080p60
source connected, then:

```
./kvmv-probe -n 300 -b 4000 -g 30 -f 30 -q 80 -o /tmp/probe.h264 -j /tmp
```

It loads the library with `dlopen`, checks all 13 symbols resolve, and runs:

1. `kvmv_hdmi_control(1)`, then waits up to 5 s for a signal.
2. H.265 reads answer `-2`; `kvmv_codec_supported` says 1, 1, 0 for MJPEG, H.264, H.265.
3. MJPEG: 10 reads at 1920x1080 and 20 at 960x540, each a complete JPEG of that size, with the
   first read's time and the median, minimum and maximum of the rest in ms per frame. `-j`
   saves the last picture of each size as `probe-WxH.jpg`.
4. 300 H.264 frames paced at 30 fps: the first is a keyframe, every keyframe carries SPS, PPS and
   IDR, keyframes come every 30 frames, the SPS size, and the output bitrate against the target
   (fails only above three times the target, until #35 is settled).
5. Half the bitrate and twice the GOP at runtime: a keyframe next, then the new interval. Then
   5 s of H.264 with a thread reading 960x540 MJPEG continuously: every H.264 frame must arrive,
   and the H.264 rate and MJPEG reads per second are printed.
6. A 1280x720 request: the SPS says 1280x720 and the stream starts on a keyframe (`-s` skips).
7. Capture off answers `-1` with no signal; capture on again starts on a keyframe.
8. `kvmv_deinit`, `kvmv_init` again, first frame a keyframe.

The exit status is the number of failed checks. `-v` adds per-frame output and the library's
debug log. Then play `/tmp/probe.h264` on a desktop (`ffplay`, or `ffprobe -show_frames`) to see
the picture, and with the server started on this library, watch the web UI in H.264 mode, change
the GOP, frame rate, bitrate and resolution from the menu, unplug and replug HDMI (expect `-1`
while out, then a keyframe), and switch the host to 1280x720 (expect "Unsupported resolution").
The same probe pointed at Sipeed's library on the vendor kernel gives a baseline.
