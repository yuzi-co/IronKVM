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
                     └─dma-buf──> sg2002-vpss ──dma-buf──> sg2002-jpeg ──> JPEG
                                  (2nd context,              (JPEG unit;      (kvmv_read_img MJPEG)
                                   full range)                libjpeg-turbo
                                                              without it)
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
| H.265 encoder (optional) | a mem2mem node whose CAPTURE queue offers HEVC and OUTPUT queue NV12 (`wave420l`, ironkvm-dist#55) |
| receiver | `/sys/class/video4linux/v4l-subdev*/name` containing `lt6911` |
| JPEG unit (optional) | driver `sg2002-jpeg`, mem2mem |

Kernel requirements: the nixos-nanokvm 7.2 media patches (LT6911UXE DT probe and UXC polling,
SG2002 CSI capture with source-change events, VPSS with patch `0900`, Coda980 with NV12 and
dma-buf input), `coda-vpu` and `sg2002-vpss` loaded (and `sg2002-jpeg` for hardware MJPEG, patch 0908 for
full-range MJPEG), `coda980.bin` installed, and a CMA
dma-heap. Without a `v4l-subdev` node the library still runs, but it cannot tell "no signal"
from "nothing captured" before a pipeline is up.

## The ABI, function by function

| Function | Here |
|----------|------|
| `kvmv_init` | Starts the monitor thread. Devices open lazily on the first read. `KVMV_DEBUG=1` or a non-zero argument enables debug logging. |
| `kvmv_read_video` | Codec 1, H.264 (the Coda980), or codec 2, H.265 (the WAVE420L): one access unit per call. `3` for a keyframe, `4` for a delta frame, `0` with no data when the output-rate guard left the picture out (see "H.264 quality"). A codec switch rebuilds the pipeline with the other encoder; the two share the codec SRAM and take turns. Codec 2 answers `-2` on a kernel without the WAVE420L driver. |
| `kvmv_read_img` | Type 1 as above. Type 0 (MJPEG): one JPEG per call at the size asked, from the JPEG unit or in software, quality from the `_qlty` argument; answers `0`. See below. |
| `free_kvmv_data`, `free_all_kvmv_data` | As the vendor library: four reusable slots, a pointer stays valid until freed. |
| `set_h264_gop` | `V4L2_CID_MPEG_VIDEO_GOP_SIZE` at runtime, plus a forced keyframe, because the vendor library rebuilds its encoder here and the server relies on the next frame being a keyframe. |
| `set_h264_fps` | `VIDIOC_S_PARM` on the encoder's OUTPUT queue at runtime; returns early when unchanged. The rate given is the lower of this and `set_capture_fps`, or the measured delivery rate when frames arrive more than 10% slower (see Known issues). |
| bitrate argument | `V4L2_CID_MPEG_VIDEO_BITRATE` (kbit/s x 1000) at runtime, clamped to 500..10000. `0` gives the default for the stream's size (below). |
| `set_capture_fps` | The capture node has no frame interval control; unread frames cost their DMA and nothing else. Caps the rate the encoder is told, since frames cannot reach it faster. |
| `set_frame_detact` | Recorded only: every MJPEG read encodes a picture, and `5` ("not changed") is never answered. |
| `kvmv_codec_supported` | Not in `kvm_vision.h` or Sipeed's library (`abi-extensions.txt`). `1` for codec 0 (MJPEG); `1` for codecs 1 (H.264) and 2 (H.265) when their encoder node exists. The server finds it with `dlsym` and, without it, assumes all three. |
| `kvmv_set_keep_aspect` | Not in `kvm_vision.h` or Sipeed's library (`abi-extensions.txt`). How a source of another shape than the requested size is fitted: `1` (the default) keeps the source's shape at the requested height, `0` stretches it to the requested size. The next read rebuilds the encoder pipeline when it changed; MJPEG follows from the next picture. The server sets it from its aspect setting through `dlsym`. |
| `set_venc_auto_recyc` | Recorded only. There is one encoder. |
| `kvmv_hdmi_control` | Stops and starts capture in software on every board: `0` tears the pipeline down and reads answer `-1`; `1` allows capture again. Answers `0`. The receiver is not powered down (the vendor library does that through a GPIO on the PCIe board only, and answers `-1` elsewhere). |
| `kvmv_hdmi_signal_active` | `1` when capture is enabled and the receiver has a source. Like the vendor library it reads `0` until `kvmv_hdmi_control(1)`. |
| `kvmv_deinit` | Joins the monitor thread, tears the pipeline down, frees the slots. `kvmv_init` may follow. |

Every runtime change the encoder refuses (an ioctl error) rebuilds the pipeline instead, which is
what the vendor library does for all of them. The encoder's actual bitrate, GOP and frame rate
are read back after every build and change and logged next to what was asked for.

**H.265 without the scaler.** At the capture's own size (1080p) the WAVE420L reads the UYVY
capture buffer itself (ironkvm-dist patch 0933; the core codes 4:2:0 from it): its OUTPUT queue
imports the capture dma-bufs, one slot per capture buffer, and each read encodes the newest frame
and queues it back to the capture when the encoder is done with it. That saves the scaler's
4 MB read and 3 MB write a picture and its 8 ms: the capture age at encoded is about 13 ms at
60 fps, against 24 to 26 through the scaler with scale-ahead. Limited range goes through as it
did (H.265 carries limited range; the scaler did not convert it either). Other sizes, H.264 (the
Coda980 reads 4:2:0 only) and a kernel without 0933 (the encoder answers NV12) keep the scaler;
`KVMV_HEVC_DIRECT=0` too. The pipeline log line says which path a build took.

### H.264 quality

Every pipeline build sets, besides the bitrate, GOP and frame rate:

| Setting | Value | Why |
|---------|-------|-----|
| QP range (`V4L2_CID_MPEG_VIDEO_H264_MIN_QP`, `MAX_QP`) | 18 to 51 | Below 18 a still screen gains nothing visible. The maximum is the encoder's own: until ironkvm-dist trial 51 it was 42, so text stayed legible on a busy screen, but the stream then ran at two to fifteen times its bitrate on scrolling text and full-screen changes, and a WebRTC viewer stalled. Kernels before ironkvm-dist patch 0910 have no minimum control on the Coda980 and ignore the maximum; the stream comes up without them. |
| Initial delay (`V4L2_CID_MPEG_VIDEO_VBV_DELAY`) | 1000 ms | The Coda980's rate-control buffer. With 0 (the control's default) it codes every picture at QP 49 to 51 and pads the stream to the bitrate. Patch 0910 uses 1000 ms itself when it is 0. |
| GOP | as asked, held to the encoder's range | The Coda takes a GOP of at most 99; the ABI allows 100. |
| Output-rate guard | 140% of the bitrate over 1 s | When the pictures handed out in the last second reach this, the next read encodes nothing and answers `0` with no data; the server sends nothing for it. Even at QP 51 a 1080p screen that changes everywhere costs 40 kB a picture in H.264, 12 in H.265; the guard holds any second to the limit and one picture, and leaves keyframes that are owed alone. It acts only while delta pictures cost their share of the bitrate (bitrate / frame rate) or more, so the small pictures after a calm screen's large keyframe still go out. `KVMV_RATE_GUARD` sets the percentage, 0 turns it off. |

When no bitrate is given (an MJPEG reader with no H.264 stream before it, or `0`), the stream
gets 3000 kbit/s at 1080p, the server's default, with half of that following the pixel count:
2200 kbit/s at 720p, 1700 at 640x480.

Measured on the board with ironkvm-dist patches 0910 and 0911 (run sheet, trial 10): luma PSNR
of 1080p desktop pictures fed to the encoder from a file, at 2, 4 and 8 Mbit/s. The numbers and
the method are in that run sheet.

### Return codes

| Code | When |
|------|------|
| `0` | An MJPEG picture. |
| `3`, `4` | A keyframe, a delta frame. |
| `-1` | No signal (receiver says no link or lock), no frame within 1 s, or capture switched off. |
| `-2` | H.265, video nodes missing, a pipeline that will not build, an encode error (H.264 or JPEG). |
| `-3` | All four slots are held by the server. |
| `-4` | The source changed mode under a running pipeline, or between the receiver query and the capture's `STREAMON` (which fails with `EPIPE`, or with `EINVAL`, `ENOLINK` or `ERANGE` from a receiver between modes); the next read rebuilds. |
| `-5` | Another read held the lock for over a second. |
| `-6` | The source runs a mode the capture cannot take: interlaced, a frame larger than a 1920x1200 one (4.4 MB), or on a kernel without ironkvm-dist patch 0936 any size but 1920x1080. |
| `-7` | The receiver reports timings out of range (`ERANGE`). |

### Keyframes

The first picture after every pipeline build is a keyframe: the encoder is asked for one, and
delta frames before it are dropped (up to four attempts per read). The coda driver creates the
keyframe control only from yuzi-co/ironkvm-dist patch 0912 (#37). Without it the request fails,
the library says so once ("keyframes wait for the GOP"), drops delta frames until the GOP's next
IDR, which costs up to seven -1 reads at GOP 30, and rebuilds the pipeline on a GOP change,
because such a kernel applies the GOP only at stream start. Every keyframe returned
carries SPS and PPS: if the encoder emits an IDR without them, the last ones it produced are
put in front. The type is decided by scanning the access unit for NAL type 5, as the vendor
library does.

### MJPEG

MJPEG goes to the SG2002's JPEG unit when the kernel has its driver, `sg2002-jpeg`
(yuzi-co/ironkvm-dist patch 0907, #36), and to libjpeg-turbo on the CPU when it does not (the
vendor kernel, or a mainline kernel without 0907). Both take the same picture:

- A second context on the scaler node imports the same capture buffers and scales the newest
  frame to the size asked for, into one NV12 buffer of its own (from the dma-heap). Its size
  does not touch the H.264 stream's, and the scaler's mem2mem queue orders its jobs with the
  H.264 ones. It is set up on the first MJPEG read, rebuilt when the size changes, and torn down
  with the pipeline.
- That context asks the scaler for full range (`V4L2_QUANTIZATION_FULL_RANGE`), because a JPEG
  file is full range and the capture is limited range (16 to 235). A scaler with ironkvm-dist
  patch 0908 expands it in its input CSC; one without keeps limited range and the library says
  so once in the log ("pictures look flat"). The H.264 context never asks: H.264 carries limited
  range as standard.
- With no H.264 stream running, an MJPEG read builds the pipeline at the size and bitrate the
  stream last used (capture runs inside it), and the idle timer tears it down 10 s after the
  last read of either kind. An H.264 read at a new size rebuilds it, and the next MJPEG read sets
  its scaler context up again.

**The JPEG unit.** The node is found by driver name, `sg2002-jpeg`, or named with
`KVMV_JPEG_DEV`; `KVMV_JPEG=sw` keeps the software encoder. The snapshot buffer's dma-buf is
queued on the unit's OUTPUT queue as it is, so no pixel passes through the CPU and no cache is
maintained for it; `KVMV_JPEG_IMPORT=0` copies the picture into a buffer of the driver's
instead, for comparison. The quality argument (1 to 100, 0 for 80) is
`V4L2_CID_JPEG_COMPRESSION_QUALITY`, set when it changes. The JPEG comes back in the driver's
CAPTURE buffer, which is mapped uncached, and is copied once, into the slot the server gets. A
1080p read takes about 10 ms, the encode about 5 ms of it, and runs under the pipeline lock.
The unit encodes widths in steps of 16 only (it rounds the width and writes the rounded one
into the header), so other widths, and three hardware failures in a row, fall back to
software. The node is closed when the pipeline goes down, which turns the unit's clocks off.

**Without the scaler.** When the picture asked for is the capture's own size (1080p) and the
capture is UYVY, the unit reads the capture buffer itself: the buffer is lent by the pipeline
(`kvmv_pipe_lend`), imported on the unit's OUTPUT queue (one buffer index per capture buffer, so
each keeps its attachment) and queued back to the capture once the JPEG is done. The unit codes
4:2:0 and expands limited range to full range in its quantiser (ironkvm-dist patches 0920 and
0921). That saves the scaler's 4 MB read and 3 MB write and the unit's 3 MB read back for every
picture, on a core whose DRAM path is the bottleneck: trial 22 measured 27.6 against 26.1 fps
over HTTPS at 1080p, the core 80 to 88% busy against 91 to 96%, and MJPEG capture latency p50
61 to 71 against 77 to 88 ms. A kernel without 0920 answers NV12 to the format; the library
says so once and keeps the scaler. `KVMV_JPEG_DIRECT=0` keeps the scaler too, for comparison.
The capture buffer is the frame's 1080 lines plus the eight the unit reads past them (0921); the
library takes its size from the dma-buf.

**Software.** libjpeg-turbo 3.1.2's TurboJPEG API, linked statically: planar 4:2:0 from the NV12
picture (chroma deinterleaved, no RGB), fast integer DCT. ironkvm-dist
`socs/sophgo-sg2002/mainline/jpeg-bench` measured about 210 ms per picture at 1080p, 94 ms at
720p and 53 ms at 960x540, all of it on the board's only core. The pipeline lock is held only to
take the picture: the capture wait, one scale, and the copy into the encoder's own planes. The
encode runs after it is released, under a lock of its own, so an H.264 read waits at most for
one scale and copy. At 1080p a reader asking continuously takes the whole core.

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
| `KVMV_DEBUG` | off | Debug logging to stderr, including the measured output rate, the rate the encoder was told, the per-stage time of a read every 10 s with how old the captured frame was when it was taken and when it was encoded, how long the scaler worked beside the encoder, and the time of each step of a pipeline build and to its first picture. |
| `KVMV_CAPTURE_DEV`, `KVMV_SCALER_DEV`, `KVMV_ENCODER_DEV`, `KVMV_SUBDEV` | discovered | Override a node. |
| `KVMV_CAPTURE_BUFFERS` | 3 | Capture queue depth (4 MB each at 1080p UYVY). Kernels before ironkvm-dist patch 0916 give 2, which loses the newest frame for a 30 fps reader of a 60 fps source. |
| `KVMV_PICKUP_WAIT_MS` | 6 | Longest a read waits for a capture frame about to complete rather than take one over half a frame time old; 0 never waits. |
| `KVMV_MID_BUFFERS` | 2 | NV12 buffers between scaler and encoder. |
| `KVMV_AHEAD_DELAY_US` | 0 | Scale-ahead gives the scaler a frame no earlier than this long after the encoder took its picture, so it takes a newer frame and works beside the encoder for less of its time. From a 110 Hz source, 4000 takes about 7 ms off the capture age at encoded for 2 fps less at 1080p (ironkvm-dist run sheet, trial 33). |
| `KVMV_BITSTREAM_BUFFERS` | 3 | Encoder output buffers. |
| `KVMV_HEVC_DIRECT` | 1 | H.265 at the capture's size reads the UYVY capture buffer itself (ironkvm-dist patch 0933), with no scaler pass: same 60 fps at 1080p from a 60 Hz source, capture age at encoded 13 ms against 24 to 26 through the scaler with scale-ahead. From a 110 Hz source asked for 110 fps it gives 71 fps against scale-ahead's 76 (ironkvm-dist run sheet, trial 36). 0 keeps the scaler. A kernel without 0933 keeps it anyway. |
| `KVMV_IDLE_MS` | 10000 | Tear down an unread pipeline after this long; 0 never does. |
| `KVMV_PARK_ENCODER` | 1 | Keep the encoder node, streamed off, with its bitstream and middle buffers over a teardown, for the next start of the same codec and size. A parked H.264 encoder keeps the Coda980's clocks on (the driver enables them at open); 0 closes it. |
| `KVMV_PRIME` | 0 | Encoder start: 0 no priming picture, 1 encode a black picture and wait for it (before trial 24), 2 queue it and let the first read collect its output. |
| `KVMV_RECEIVER_CACHE_MS` | 1500 | A build uses the monitor thread's receiver answer up to this old instead of asking again; 0 always asks. |
| `KVMV_H264_QP` | `18:51` | H.264 QP range, `min:max`, 0 to 51. `0:51` leaves the encoder's own. |
| `KVMV_H264_VBV_DELAY_MS` | 1000 | Rate-control initial delay; 0 leaves the encoder's own. |
| `KVMV_H265_QP` | `12:51` | H.265 QP range, as `KVMV_H264_QP`. Lower than H.264's minimum: at 18 the WAVE420L stops short of the asked bitrate on screen content (ironkvm-dist run sheet, trial 20). |
| `KVMV_H265_VBV_DELAY_MS` | 2000 | H.265 rate-control initial delay. At 1000 the WAVE420L starves the IDR, which carries most of a 1 s GOP's bits on a screen (trial 20). |
| `KVMV_RATE_GUARD` | 140 | The output-rate guard's limit in percent of the bitrate over one second (see "H.264 quality"); 0 turns it off. |

## Known issues

- Bitrate and quality (yuzi-co/ironkvm-dist#35): the bridge produced about nine times its
  target; patch 0905 made the Coda980 follow it, but only by padding pictures coded at QP 51.
  Patches 0910 (rate control) and 0911 (motion estimation at 1080p) fix the picture; on older
  kernels the stream is legible only on a still screen. The library measures its own output and
  logs once per pipeline when it runs above twice the target. Busy screens ran far over the
  target with the maximum QP at 42 (trial 50); from trial 51 the QP goes to 51 and the
  output-rate guard leaves pictures out when even that is not enough, and logs once when it does.
- Before ironkvm-dist patch 0912 the GOP set at runtime reached the Coda only at the next
  pipeline build, and no keyframe could be forced (see "Keyframes").
- HDMI modes (yuzi-co/ironkvm-dist#37): any progressive mode the LT6911UXC locks whose UYVY
  frame is no larger than a 1920x1200 one (4.4 MB). Before ironkvm-dist patch 0940 the video
  pool rounded each capture buffer up to a power of two, and three 8 MiB blocks for 1920x1200
  left the encoders short; the kernel and the library go together on slot B. The capture node
  follows the receiver's size while it holds no buffers (ironkvm-dist patch 0936), the scaler
  reads its 16-byte padded line for 1366 wide modes (0937), and the LT6911UXC's modes under
  1024x768 pass its timing check (0938). The stream is the source's size, or the size asked for
  when smaller; a size beyond 1920x1080 is scaled into it with its shape kept. A source of
  another shape than the request keeps its shape at the requested height (`kvmv_set_keep_aspect`,
  default): 1920x1200 asked for 1920x1080 is 1728x1080, 1024x768 asked for 1280x720 is 960x720,
  1280x1024 asked for 1280x720 is 896x716 (the width in steps of 16, the height following it).
  With `kvmv_set_keep_aspect(0)` each side is held to the request on its own and the picture is
  stretched (1920x1080, 1280x720), as before. Trial 43 streamed
  1280x720, 1024x768, 1680x1050, 1366x768 (as 1360x768), 1280x1024, 1280x800, 1280x960,
  800x600 and 720x400 at 70 Hz; the receiver itself reports no link for 640x480 at 60 Hz and
  720x576 at 50 Hz, and the read answers `-1` as with no cable. Before 0936 only 1920x1080
  sources work and other modes answer `-6`.
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
  `kvmv_codec_supported` and `kvmv_set_keep_aspect` (`abi-extensions.txt`), nothing else; libjpeg-turbo is inside it.
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
  frame-rate policy, the uncached copy, the software JPEG path (NV12 split, encode, decode back), and the JPEG
  unit against a mock `sg2002-jpeg` (copy and dma-buf import, quality, reconfiguration, a bad
  picture, widths it cannot take).
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
2. `kvmv_codec_supported` says 1 for MJPEG and H.264, and whether H.265 is there. Without
   it, an H.265 read must answer `-2`.
3. MJPEG: 30 reads at 1920x1080, 30 at 1280x720 and 20 at 960x540, each a complete JPEG of that
   size, with the first read's time, the median, minimum and maximum of the rest in ms per
   frame, and the rate back to back. `-j` saves the last picture of each size as
   `probe-WxH.jpg`. Then 10 s each at 1080p and 720p paced at `-f`: the delivered rate must reach
   90% of it (it does with the JPEG unit, not in software), and the probe's CPU share is printed.
   The library logs which encoder it uses and whether the scaler gave full range.
4. 300 H.264 frames paced at 30 fps: the first is a keyframe, every keyframe carries SPS, PPS and
   IDR, keyframes come every 30 frames, no read answered an error (dropped delta frames at the
   start show as -1 reads), the SPS size, and the output bitrate against the target (fails only
   above three times the target, until #35 is settled).
5. Half the bitrate and twice the GOP at runtime: a keyframe next with no dropped reads, then
   the new interval. Then
   5 s of H.264 with a thread reading 960x540 MJPEG continuously: every H.264 frame must arrive,
   and the H.264 rate and MJPEG reads per second are printed.
5c. With H.265: 300 frames (first a keyframe, every keyframe carrying VPS, SPS, PPS and an IRAP
   picture, the GOP, the SPS size, the bitrate), then 5 s with an MJPEG reader alongside, then
   back to H.264, which must start on a keyframe with no error reads.
6. A 1280x720 request: the SPS says 1280x720 and the stream starts on a keyframe (`-s` skips).
7. Capture off answers `-1` with no signal; capture on again starts on a keyframe.
8. `kvmv_deinit`, `kvmv_init` again, first frame a keyframe.

The exit status is the number of failed checks. `-v` adds per-frame output and the library's
debug log. Then play `/tmp/probe.h264` on a desktop (`ffplay`, or `ffprobe -show_frames`) to see
the picture, and with the server started on this library, watch the web UI in H.264 mode, change
the GOP, frame rate, bitrate and resolution from the menu, unplug and replug HDMI (expect `-1`
while out, then a keyframe), and switch the host to 1280x720 (expect `-4` once, then the stream
at 1280x720; "Unsupported resolution" on a kernel without patch 0936).
The same probe pointed at Sipeed's library on the vendor kernel gives a baseline.
