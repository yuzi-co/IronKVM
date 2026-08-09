# Full-rate Opus Audio Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Send the managed host's audio at the rate the USB gadget already captures it — 48 kHz stereo Opus — instead of decimating it to 8 kHz mono G.711.

**Architecture:** libopus ships as a committed riscv64 static archive. A thin cgo wrapper implements a small `Encoder` interface, so the capture pipeline is testable on a workstation with a fake while the real codec is exercised by a cross-compiled test binary run on the board. `Stream.consume` calls the encoder instead of a FIR and a mu-law table; the WebRTC layer swaps payloader, clock rate and MIME type. The 129-tap decimator and the G.711 encoder are deleted.

**Tech Stack:** Go 1.24 with cgo, libopus 1.5.2, pion/webrtc v4, pion/rtp v1.8.18, the Xuantie riscv64-musl toolchain inside the repository's Docker builder image.

**Design:** `docs/superpowers/specs/2026-08-09-opus-audio-design.md`

## Global Constraints

- **Branch:** `feat/opus-audio`, already created, cut from `fork/integration` at `124e41a6`. Do not commit to `fork/integration`.
- **libopus version:** 1.5.2. Source `https://downloads.xiph.org/releases/opus/opus-1.5.2.tar.gz`, sha256 `65c1d2f78b9f2fb20082c38cbe47c951ad5839345876e46941612ee87f9a7ce1`.
- **libopus build:** floating point (NOT `--enable-fixed-point`; fixed point measured 24% slower on this core).
- **Compiler flags for every C artefact:** `-O3 -mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d`
- **Encoder settings, fixed at compile time, not configurable:** 48000 Hz, 2 channels, 960 samples per frame, 96000 bit/s, complexity 3, `OPUS_APPLICATION_AUDIO`.
- **Builder image name:** `nanokvm-builder-local-197609-197121`. The Makefile targets need a POSIX shell and there is no `make` in Git Bash, so invoke `docker run` directly. Prefix docker commands with `MSYS_NO_PATHCONV=1`.
- **Device:** `root@10.0.0.222`. It is a live KVM. Availability outranks this feature.
- **Deploy:** `DEPLOY_TIMEOUT=200 sh /data/deploy-server /data/NanoKVM-Server.new`. A restart takes about 137 seconds; the script's 45 second default would roll back a good build.
- **After linking:** `patchelf --print-needed` on `NanoKVM-Server` must report exactly `libkvm.so` and `libc.so`. A static archive adds no entry; a new entry means something went wrong.
- **Commit messages are normal English prose**, never caveman, and end with the repository's `Co-Authored-By` and `Claude-Session` trailers.
- **Never write runtime state under `/kvmapp`.** It is the boot SD card.

### The two test commands

**Off-device (portable code, runs on the workstation):**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision ./...
```

Mount the repository root, not `server/`: `service/vm` tests read `../../../kvmapp/system/init.d/S03usbdev` and fail otherwise.

**On-device (the cgo encoder). Build the test binary in the builder image, then run it on the board:**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/home/build/NanoKVM" \
  nanokvm-builder-local-197609-197121 bash -c \
  'cd /home/build/NanoKVM/server && CGO_ENABLED=1 GOOS=linux GOARCH=riscv64 \
   CC=riscv64-unknown-linux-musl-gcc \
   CGO_CFLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d" \
   go test -c -buildvcs=false -o /home/build/NanoKVM/audio.test ./service/stream/audio'

scp audio.test root@10.0.0.222:/tmp/
ssh root@10.0.0.222 'chmod +x /tmp/audio.test && /tmp/audio.test -test.v; rm -f /tmp/audio.test'
rm -f audio.test
```

`audio.test` must never be committed. Delete it after every run.

---

## File Structure

| path | responsibility |
| --- | --- |
| `server/dl_lib/libopus.a` | The codec. Committed binary, riscv64, float. |
| `server/include/opus/opus.h` | Public header. |
| `server/include/opus/opus_defines.h` | Public header, included by `opus.h`. |
| `server/include/opus/opus_types.h` | Public header, included by both. |
| `server/service/stream/audio/encoder.go` | The `Encoder` interface, the format constants, the `newEncoder` seam. Portable — no cgo, no build tag. |
| `server/service/stream/audio/encoder_opus.go` | `//go:build !novision`. The cgo wrapper. The only file that knows libopus exists. |
| `server/service/stream/audio/encoder_stub.go` | `//go:build novision`. Reports that this build has no encoder. |
| `server/service/stream/audio/encoder_opus_test.go` | `//go:build !novision`. Runs on the device only. |
| `server/service/stream/audio/encoder_test.go` | The fake encoder and the portable tests that use it. |
| `server/service/stream/audio/audio.go` | `Stream`: owns the encoder, converts chunks to packets. |
| `server/service/stream/audio/source.go` | Unchanged except `ChunkBytes`, which loses its magic number. |
| `tools/opusbench/build.sh` | Rebuilds `libopus.a`, the headers, and the benchmark. |
| `tools/opusbench/opusbench.c` | The measurement harness. Encodes and decodes, so it is also the round-trip check. |
| `tools/opusbench/README.md` | Version, flags, how to run, what the numbers mean. |

**Deleted:** `resample.go`, `resample_test.go`, `g711.go`, `g711_test.go`.

---

### Task 1: Ship libopus and the measurement harness

Nothing in Go changes here. The deliverable is a reproducible archive and a benchmark that runs on the board.

**Files:**
- Create: `tools/opusbench/build.sh`
- Create: `tools/opusbench/opusbench.c`
- Create: `tools/opusbench/README.md`
- Create: `server/dl_lib/libopus.a` (build output, committed)
- Create: `server/include/opus/opus.h`, `opus_defines.h`, `opus_types.h` (build output, committed)
- Modify: `tools/README.md` (add one table row)

**Interfaces:**
- Consumes: nothing.
- Produces: `server/dl_lib/libopus.a` and `server/include/opus/*.h`, which Task 2's cgo directives reference by path.

- [ ] **Step 1: Copy the benchmark source into the repository**

The harness already exists and has been run on the device. Copy it verbatim:

```bash
mkdir -p tools/opusbench
cp "/c/Users/vadim/AppData/Local/Temp/claude/D--projects-NanoKVM/b5dd3500-66be-479e-a2de-c0cc34935cfd/scratchpad/opusbench/opusbench.c" \
   tools/opusbench/opusbench.c
```

Verify it contains all four of these, because they are what make its numbers trustworthy:
- `clock_gettime(CLOCK_PROCESS_CPUTIME_ID, ...)` in `cpu_seconds` — wall time would count the KVM's own work.
- `static volatile unsigned long g711_checksum` and the `sum += out[n]` that feeds it — without a consumer the compiler deletes the 129-tap loop and reports an impossible 0.10%.
- A `bench_g711` that transcribes the 129-tap Hamming-windowed sinc at 3.4 kHz with decimation by 6.
- A `bench_opus` whose `with_decode` branch creates an `OpusDecoder` and measures `opus_decode`.

- [ ] **Step 2: Write the build script**

Create `tools/opusbench/build.sh`:

```bash
#!/bin/bash
# Rebuild libopus for the NanoKVM's C906 core, install it into the server tree,
# and build the measurement harness.
#
# Run this from the repository root, inside the builder image:
#
#   MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/home/build/NanoKVM" \
#     nanokvm-builder-local-$(id -u)-$(id -g) \
#     bash /home/build/NanoKVM/tools/opusbench/build.sh
#
# The archive and the headers it installs are committed. The benchmark binary
# is not: it is built when somebody wants to measure.
set -e

VERSION=1.5.2
SHA256=65c1d2f78b9f2fb20082c38cbe47c951ad5839345876e46941612ee87f9a7ce1

HOST=riscv64-unknown-linux-musl
CC=$HOST-gcc

# The same flags the Go cgo build uses in the repository Makefile. The board
# has a hardware FPU, so libopus is built in floating point. A fixed-point
# build was measured at 9.78% of the core against 7.88% for float, at
# complexity 5, stereo, 96 kbit/s.
ARCH_FLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d"

root=$(cd "$(dirname "$0")/../.." && pwd)
work=$root/tools/opusbench/.build

mkdir -p "$work"
cd "$work"

if [ ! -f "opus-$VERSION.tar.gz" ]; then
    curl -sSL -o "opus-$VERSION.tar.gz" \
        "https://downloads.xiph.org/releases/opus/opus-$VERSION.tar.gz"
fi

echo "$SHA256  opus-$VERSION.tar.gz" | sha256sum -c -

rm -rf "opus-$VERSION" build inst
tar xzf "opus-$VERSION.tar.gz"

mkdir build
cd build

../opus-$VERSION/configure \
    --host="$HOST" \
    --prefix="$work/inst" \
    --disable-shared --enable-static \
    --disable-doc --disable-extra-programs \
    CFLAGS="-O3 $ARCH_FLAGS" >configure.log 2>&1

make -j"$(nproc)" >make.log 2>&1
make install >>make.log 2>&1

cd "$work"

install -D -m 644 inst/lib/libopus.a "$root/server/dl_lib/libopus.a"
for header in opus.h opus_defines.h opus_types.h; do
    install -D -m 644 "inst/include/opus/$header" "$root/server/include/opus/$header"
done

$CC -O3 $ARCH_FLAGS -static \
    -I "$root/server/include/opus" \
    "$root/tools/opusbench/opusbench.c" \
    "$root/server/dl_lib/libopus.a" \
    -lm -o "$root/tools/opusbench/opusbench"

echo "installed server/dl_lib/libopus.a ($(stat -c %s "$root/server/dl_lib/libopus.a") bytes)"
echo "built tools/opusbench/opusbench ($(stat -c %s "$root/tools/opusbench/opusbench") bytes)"
```

Make it executable:

```bash
git update-index --add --chmod=+x tools/opusbench/build.sh 2>/dev/null || chmod +x tools/opusbench/build.sh
```

- [ ] **Step 3: Run the build**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/home/build/NanoKVM" \
  nanokvm-builder-local-197609-197121 \
  bash /home/build/NanoKVM/tools/opusbench/build.sh
```

Expected: `installed server/dl_lib/libopus.a (1395710 bytes)` — a few bytes either way is fine, a wildly different size is not.

Confirm the archive is for the right machine:

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/w" -w /w \
  nanokvm-builder-local-197609-197121 \
  bash -c 'riscv64-unknown-linux-musl-objdump -f server/dl_lib/libopus.a | head -4'
```

Expected: `file format elf64-littleriscv`.

- [ ] **Step 4: Run the benchmark on the device**

```bash
scp tools/opusbench/opusbench root@10.0.0.222:/tmp/
ssh root@10.0.0.222 'chmod +x /tmp/opusbench && nice -n 15 /tmp/opusbench; rm -f /tmp/opusbench'
```

Expected, within about 10% (the board is doing other work):

```
opus/audio/tone        stereo      96k     3    0.560     5.60      96.4
```

If complexity 3 stereo costs much more than 5.60% of the core, stop and find out why before continuing — the whole design rests on that number.

- [ ] **Step 5: Write the tools README**

Create `tools/opusbench/README.md`:

```markdown
# opusbench

Measure what an Opus encoder costs on the NanoKVM's single C906 core, and
rebuild the `libopus.a` that the server links against.

## Rebuild the archive

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/home/build/NanoKVM" \
  nanokvm-builder-local-$(id -u)-$(id -g) \
  bash /home/build/NanoKVM/tools/opusbench/build.sh
```

The script downloads libopus 1.5.2, checks its sha256, cross-compiles it in
floating point with the repository's own flags, and installs
`server/dl_lib/libopus.a` and `server/include/opus/*.h`. Commit both after a
rebuild, and run the benchmark again.

## Measure

```shell
scp tools/opusbench/opusbench root@<device>:/tmp/
ssh root@<device> 'nice -n 15 /tmp/opusbench'
```

Each row encodes 10 seconds of 48 kHz audio in 20 ms frames. `core_%` is the
share of one core the row needs to keep up with real time. CPU time comes from
`CLOCK_PROCESS_CPUTIME_ID`, so other work on the board does not inflate it.

## What the numbers said on 2026-08-09

The server ships stereo, 96 kbit/s, complexity 3, which measured 5.60%. The
129-tap FIR and G.711 path it replaced measured 4.66%, so full-rate audio
costs about one percentage point more than telephone audio did.

Do not build libopus in fixed point. The C906B has a hardware FPU, and the
fixed-point build measured 9.78% against 7.88% for float at complexity 5.

The `g711-fir` row carries a checksum. It is load-bearing: without a consumer
for the filtered samples the compiler deletes the 129-tap loop and the row
reports an impossible 0.10% of the core.
```

- [ ] **Step 6: Add the row to the tools index**

In `tools/README.md`, after the `audiodiag/` row, add:

```
| `opusbench/`  | Rebuild `libopus.a` for the board, and measure what it costs.          |
```

- [ ] **Step 7: Ignore the build directory**

Add to `.gitignore` (create the entry if the file already exists, do not replace the file):

```
tools/opusbench/.build/
tools/opusbench/opusbench
```

- [ ] **Step 8: Commit**

```bash
git add tools/opusbench server/dl_lib/libopus.a server/include/opus tools/README.md .gitignore
git status --porcelain   # confirm .build/ and the opusbench binary are NOT staged
git commit
```

Message body should say: libopus 1.5.2 built float with the repository's own toolchain flags; fixed point was measured slower on this core; the archive is committed rather than built at release time because the deploy is one file.

---

### Task 2: The Encoder interface and its libopus implementation

**Files:**
- Create: `server/service/stream/audio/encoder.go`
- Create: `server/service/stream/audio/encoder_opus.go`
- Create: `server/service/stream/audio/encoder_stub.go`
- Create: `server/service/stream/audio/encoder_opus_test.go`
- Modify: `server/service/stream/audio/source.go:19` (`ChunkBytes` loses its magic number)

**Interfaces:**
- Consumes: `server/dl_lib/libopus.a` and `server/include/opus/*.h` from Task 1.
- Produces:
  - `type Encoder interface { Encode(pcm []byte, dst []byte) ([]byte, error); Close() }`
  - `func newOpusEncoder() (Encoder, error)` — defined twice, once per build tag.
  - Constants `SampleRate`, `Channels`, `SamplesPerFrame`, `ChunkBytes`, `Bitrate`, `Complexity`, `maxPacketBytes`.

- [ ] **Step 1: Write the failing device test**

Create `server/service/stream/audio/encoder_opus_test.go`:

```go
//go:build !novision

package audio

import (
	"math"
	"testing"
)

// These run on the device. The archive is riscv64, so the encoder does not
// link anywhere else. Build with `go test -c` in the builder image and run the
// binary on the board; the command is in the plan and in CLAUDE.md.

// tone fills one chunk with a 440 Hz sine at 48 kHz, stereo, S16_LE. Silence
// would encode to a handful of bytes and would not show that the encoder is
// carrying a signal.
func tone() []byte {
	chunk := make([]byte, ChunkBytes)

	for i := range SamplesPerFrame {
		value := int16(20000 * math.Sin(2*math.Pi*440*float64(i)/SampleRate))

		for c := range Channels {
			offset := (i*Channels + c) * 2
			chunk[offset] = byte(value)
			chunk[offset+1] = byte(value >> 8)
		}
	}

	return chunk
}

func TestOpusEncoderEncodesAChunk(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	packet, err := encoder.Encode(tone(), nil)
	if err != nil {
		t.Fatalf("failed to encode: %s", err)
	}

	// 20 ms at 96 kbit/s is about 240 bytes. A packet of two or three bytes is
	// what Opus emits for silence, and it would mean the samples never arrived.
	if len(packet) < 50 {
		t.Errorf("the packet is %d bytes, which is too small to carry a tone", len(packet))
	}

	if len(packet) > maxPacketBytes {
		t.Errorf("the packet is %d bytes, over the %d byte buffer", len(packet), maxPacketBytes)
	}
}

func TestOpusEncoderAppendsToDestination(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	dst := []byte{0xAA, 0xBB}

	packet, err := encoder.Encode(tone(), dst)
	if err != nil {
		t.Fatalf("failed to encode: %s", err)
	}

	if len(packet) <= 2 || packet[0] != 0xAA || packet[1] != 0xBB {
		t.Errorf("Encode did not append to the destination it was given")
	}
}

// The chunk length is checked on the Go side because the C call reads
// SamplesPerFrame*Channels samples whatever the slice actually holds. A short
// chunk would read past the end, and this codec runs in the server process.
func TestOpusEncoderRejectsAShortChunk(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	if _, err := encoder.Encode(make([]byte, ChunkBytes-2), nil); err == nil {
		t.Error("Encode accepted a short chunk")
	}
}

func TestOpusEncoderRejectsAnEmptyChunk(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	if _, err := encoder.Encode(nil, nil); err == nil {
		t.Error("Encode accepted an empty chunk")
	}
}
```

- [ ] **Step 2: Run the device test to verify it fails**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/home/build/NanoKVM" \
  nanokvm-builder-local-197609-197121 bash -c \
  'cd /home/build/NanoKVM/server && CGO_ENABLED=1 GOOS=linux GOARCH=riscv64 \
   CC=riscv64-unknown-linux-musl-gcc \
   CGO_CFLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d" \
   go test -c -buildvcs=false -o /home/build/NanoKVM/audio.test ./service/stream/audio'
```

Expected: FAIL to compile, `undefined: newOpusEncoder`, `undefined: SamplesPerFrame`, `undefined: maxPacketBytes`.

- [ ] **Step 3: Write the portable half**

Create `server/service/stream/audio/encoder.go`:

```go
package audio

// The capture format. arecord is started with these and the encoder is created
// with these, so they are one set of constants rather than two sets that have
// to agree.
const (
	SampleRate = 48000

	// Channels is 2 because the gadget captures stereo and this is what the
	// managed host sends. Downmixing would save about 2% of the core and throw
	// away half the information.
	Channels = 2

	// SamplesPerFrame is 20 ms per channel, which is the period arecord is
	// asked for and the frame size Opus is given.
	SamplesPerFrame = 960

	// Bitrate and Complexity were chosen by measurement on the device. At
	// complexity 3 the encoder needs 5.60% of the core; the 129-tap FIR and
	// G.711 path it replaced needed 4.66%. Complexity 5 costs 7.88% and buys
	// little at this bitrate. Bitrate barely moves the CPU cost at all: four
	// times the rate costs one fifth more. See tools/opusbench.
	Bitrate    = 96000
	Complexity = 3

	// maxPacketBytes bounds one encoded frame. 20 ms of stereo at 96 kbit/s is
	// about 240 bytes, and Opus never exceeds 1275 bytes per channel per
	// frame, so this has room to spare.
	maxPacketBytes = 4000
)

// Encoder turns one 20 ms chunk of 48 kHz stereo S16_LE into one packet.
//
// It is an interface because the implementation is a static riscv64 archive:
// it links on the device and nowhere else. Every test of the pipeline supplies
// its own encoder, and only encoder_opus_test.go exercises the real one.
type Encoder interface {
	// Encode appends the packet to dst and returns the extended slice, which
	// is the convention the rest of this package uses. Passing dst[:0] reuses
	// the caller's buffer.
	Encode(pcm []byte, dst []byte) ([]byte, error)

	// Close releases the encoder. It is called once, from the goroutine that
	// owns the stream.
	Close()
}
```

Create `server/service/stream/audio/encoder_stub.go`:

```go
//go:build novision

package audio

import "errors"

// errNoEncoder is what a build without the device libraries reports.
//
// The pipeline still builds and is still tested here: every test supplies its
// own Encoder. Only the codec is missing, and a workstation has no UAC1 gadget
// to capture from either, so Available reports false long before this matters.
var errNoEncoder = errors.New("audio: this build has no Opus encoder")

func newOpusEncoder() (Encoder, error) {
	return nil, errNoEncoder
}
```

- [ ] **Step 4: Write the cgo wrapper**

Create `server/service/stream/audio/encoder_opus.go`:

```go
//go:build !novision

package audio

/*
	#cgo CFLAGS: -I${SRCDIR}/../../../include
	#cgo LDFLAGS: ${SRCDIR}/../../../dl_lib/libopus.a -lm

	#include <opus/opus.h>

	// opus_encoder_ctl is variadic, and cgo cannot call a variadic C function.
	// This wrapper is not, so cgo can. Every setting this package needs takes
	// one opus_int32, so one wrapper covers all of them.
	static int kvm_opus_set(OpusEncoder *encoder, int request, opus_int32 value) {
		return opus_encoder_ctl(encoder, request, value);
	}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type opusEncoder struct {
	state *C.OpusEncoder

	// scratch receives the packet from C. One encoder belongs to one stream
	// and Encode is called from one goroutine, so a single buffer is enough
	// and it keeps 50 allocations a second off the heap.
	scratch []byte
}

func newOpusEncoder() (Encoder, error) {
	var status C.int

	state := C.opus_encoder_create(
		C.opus_int32(SampleRate),
		C.int(Channels),
		C.OPUS_APPLICATION_AUDIO,
		&status,
	)
	if state == nil || status != C.OPUS_OK {
		return nil, fmt.Errorf("opus_encoder_create: %s", opusError(status))
	}

	encoder := &opusEncoder{state: state, scratch: make([]byte, maxPacketBytes)}

	// OPUS_APPLICATION_AUDIO rather than VOIP. This carries desktop audio, not
	// a phone call, and VOIP was measured more expensive as well as wrong:
	// mono at 32 kbit/s in VOIP mode costs more of the core than stereo at
	// 96 kbit/s in AUDIO mode, because SILK is not cheap.
	settings := []struct {
		name    string
		request C.int
		value   C.opus_int32
	}{
		{"bitrate", C.OPUS_SET_BITRATE_REQUEST, C.opus_int32(Bitrate)},
		{"complexity", C.OPUS_SET_COMPLEXITY_REQUEST, C.opus_int32(Complexity)},
	}

	for _, setting := range settings {
		if status := C.kvm_opus_set(state, setting.request, setting.value); status != C.OPUS_OK {
			C.opus_encoder_destroy(state)
			return nil, fmt.Errorf("opus_encoder_ctl(%s): %s", setting.name, opusError(status))
		}
	}

	return encoder, nil
}

// Encode hands one chunk to libopus.
//
// The length check is not defensive noise. This codec runs inside the server
// process, unlike arecord, which is a child that the source restarts when it
// dies. opus_encode reads SamplesPerFrame*Channels samples from the pointer it
// is given whatever the slice behind it holds, so a short chunk reads past the
// end and takes the whole server with it. Source.runOnce fills chunks with
// io.ReadFull and cannot produce a short one, but that guarantee lives in
// another file and this is where it has to hold.
func (e *opusEncoder) Encode(pcm []byte, dst []byte) ([]byte, error) {
	if len(pcm) != ChunkBytes {
		return dst, fmt.Errorf("audio: chunk is %d bytes, want %d", len(pcm), ChunkBytes)
	}

	// The samples are S16_LE and this is a little-endian machine, so the byte
	// slice is already an array of opus_int16. A slice from make is aligned
	// well enough for the cast.
	written := C.opus_encode(
		e.state,
		(*C.opus_int16)(unsafe.Pointer(&pcm[0])),
		C.int(SamplesPerFrame),
		(*C.uchar)(unsafe.Pointer(&e.scratch[0])),
		C.opus_int32(len(e.scratch)),
	)
	if written < 0 {
		return dst, fmt.Errorf("opus_encode: %s", opusError(C.int(written)))
	}

	return append(dst, e.scratch[:written]...), nil
}

func (e *opusEncoder) Close() {
	if e.state == nil {
		return
	}

	C.opus_encoder_destroy(e.state)
	e.state = nil
}

func opusError(status C.int) string {
	return C.GoString(C.opus_strerror(status))
}
```

- [ ] **Step 5: Remove the magic number from ChunkBytes**

In `server/service/stream/audio/source.go`, replace:

```go
	// ChunkBytes is 20 ms of 48 kHz stereo S16_LE: 960 frames of 4 bytes.
	ChunkBytes = 960 * 4
```

with:

```go
	// ChunkBytes is 20 ms of 48 kHz stereo S16_LE. It is derived rather than
	// written out, because arecord and the encoder have to agree on it.
	ChunkBytes = SamplesPerFrame * Channels * 2
```

- [ ] **Step 6: Run the device test to verify it passes**

Build, copy, run, and delete, using the on-device command from Global Constraints.

Expected: `PASS`, with `TestOpusEncoderEncodesAChunk`, `TestOpusEncoderAppendsToDestination`, `TestOpusEncoderRejectsAShortChunk` and `TestOpusEncoderRejectsAnEmptyChunk` all passing.

- [ ] **Step 7: Verify the off-device build still works**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision ./...
```

Expected: every package ok. The `novision` build takes `encoder_stub.go`, so nothing tries to link the archive.

- [ ] **Step 8: Commit**

```bash
rm -f audio.test
git add server/service/stream/audio/encoder.go server/service/stream/audio/encoder_opus.go \
       server/service/stream/audio/encoder_stub.go server/service/stream/audio/encoder_opus_test.go \
       server/service/stream/audio/source.go
git commit
```

Message body should explain: why the interface exists (the archive is riscv64, so the real encoder links only on the device), why `opus_encoder_ctl` needs a C shim (cgo cannot call variadic functions), and why `Encode` checks the chunk length (the codec now runs in the server process, where a bad read costs a reboot rather than a child restart).

---

### Task 3: Switch the pipeline to Opus

This task changes the audio package and the WebRTC layer together. It has to: `FrameSamples` means both "RTP sample count" and "frame length in bytes" today, and removing it breaks both packages at once. Splitting the task would leave a commit that compiles but packetizes 960 samples of mu-law.

**Files:**
- Modify: `server/service/stream/audio/audio.go`
- Create: `server/service/stream/audio/encoder_test.go`
- Modify: `server/service/stream/audio/audio_test.go:82-101`
- Modify: `server/service/stream/audio/stream_linux_test.go:20-29,41-53,107-142,145-166`
- Modify: `server/service/stream/webrtc/track.go:12-13`
- Modify: `server/service/stream/webrtc/manager.go:41-48`
- Modify: `server/service/stream/webrtc/client.go:220-225`
- Modify: `server/service/stream/webrtc/audio.go:184`
- Modify: `server/service/stream/webrtc/audio_test.go:402`
- Delete: `server/service/stream/audio/resample.go`, `resample_test.go`, `g711.go`, `g711_test.go`

**Interfaces:**
- Consumes: `Encoder`, `newOpusEncoder`, `SamplesPerFrame`, `ChunkBytes`, `maxPacketBytes` from Task 2.
- Produces: `Stream` with a `newEncoder func() (Encoder, error)` field that tests assign before `Start`. `audio.FrameSamples`, `audio.InputRate`, `audio.OutputRate`, `audio.Decimation`, `audio.EncodeULaw` and `audio.NewDecimator` no longer exist.

- [ ] **Step 1: Write the fake encoder and the first failing test**

Create `server/service/stream/audio/encoder_test.go`:

```go
package audio

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeEncoder stands in for libopus. The real encoder is a riscv64 archive and
// does not link on a workstation, so every test of the pipeline uses this.
type fakeEncoder struct {
	mutex  sync.Mutex
	chunks [][]byte
	packet []byte
	err    error
	closed bool
}

func (f *fakeEncoder) Encode(pcm []byte, dst []byte) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.chunks = append(f.chunks, append([]byte(nil), pcm...))

	if f.err != nil {
		return dst, f.err
	}

	return append(dst, f.packet...), nil
}

func (f *fakeEncoder) Close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.closed = true
}

func (f *fakeEncoder) count() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return len(f.chunks)
}

func (f *fakeEncoder) isClosed() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.closed
}

func (f *fakeEncoder) setError(err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.err = err
}

// fakeStream returns a stream ready to consume chunks, with no capture child
// behind it.
//
// It sets the encoder directly rather than calling Start. Start would launch
// the source, which runs arecord: absent on a workstation, so the retry loop
// would spin and log for the length of every test that only wants to exercise
// consume. The two tests that are about the lifecycle call Start themselves
// and supply an inert child.
func fakeStream(encoder *fakeEncoder) *Stream {
	stream := NewStream()
	stream.newEncoder = func() (Encoder, error) { return encoder, nil }
	stream.encoder = encoder

	return stream
}

func TestStreamDeliversWhatTheEncoderReturns(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("opus-packet")}
	stream := fakeStream(encoder)

	go stream.consume(make([]byte, ChunkBytes))

	select {
	case frame, ok := <-stream.Frames():
		if !ok {
			t.Fatal("the frame channel closed before delivering a frame")
		}

		if string(frame) != "opus-packet" {
			t.Errorf("the frame carried %q, want the encoder's packet", frame)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no frame arrived within 5s")
	}

	stream.Stop()
}
```

- [ ] **Step 2: Run it and verify it fails**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 golang:1.25 \
  go test -tags novision -run TestStreamDeliversWhatTheEncoderReturns -v ./service/stream/audio/
```

Expected: FAIL to compile, `stream.newEncoder undefined`.

- [ ] **Step 3: Rewire Stream onto the encoder**

In `server/service/stream/audio/audio.go`:

Delete the `FrameSamples` constant and its comment at lines 15-17.

Replace the `Stream` struct and `NewStream` with:

```go
// Stream turns the capture device into Opus packets.
type Stream struct {
	source *Source
	frames chan []byte

	// newEncoder is a field so a test can supply an encoder that needs no
	// libopus. The production value is the cgo wrapper, or the stub on a build
	// without the device libraries.
	newEncoder func() (Encoder, error)

	// encoder and packet are touched only by the source goroutine, between
	// Start launching it and Run returning.
	encoder Encoder
	packet  []byte

	// encodeFailures counts consecutive failed frames, so the log can fall
	// quiet when failure is the steady state.
	encodeFailures int

	mutex     sync.Mutex
	started   bool
	closeOnce sync.Once
	done      chan struct{}
}

func NewStream() *Stream {
	return &Stream{
		source:     NewSource(),
		newEncoder: newOpusEncoder,
		// Four frames of slack. A consumer further behind than 80 ms is not
		// going to catch up, and buffering only adds delay.
		frames: make(chan []byte, 4),
		packet: make([]byte, 0, maxPacketBytes),
		done:   make(chan struct{}),
	}
}
```

Replace `Start` with:

```go
// Start begins capture. Frames arrive on Frames until Stop.
func (s *Stream) Start() {
	s.mutex.Lock()
	if s.started {
		s.mutex.Unlock()
		return
	}
	s.started = true
	s.mutex.Unlock()

	encoder, err := s.newEncoder()
	if err != nil {
		// This is permanent, and it is deliberately unlike the source's retry
		// loop. A host that plays nothing is idle and may start at any moment,
		// so capture retries it forever. An encoder that cannot be built will
		// not build itself later, so the stream ends and the manager learns
		// that audio stopped.
		//
		// Nothing was launched, so this path closes done itself. The goroutine
		// below is the only other closer and it never runs.
		log.Errorf("audio is off: %s", err)
		close(s.done)
		s.closeFrames()

		return
	}

	s.encoder = encoder

	go func() {
		defer close(s.done)
		defer s.encoder.Close()

		s.source.Run(s.consume)

		// Run returns when Stop killed the child. Closing here is what lets a
		// consumer that is draining Frames finish.
		s.closeFrames()
	}()
}
```

Replace `consume` with:

```go
// consume encodes one capture chunk and offers the packet. It never blocks: a
// consumer that is behind loses 20 ms rather than stalling capture.
//
// A drop here is not the same as the per-client drop in Client.enqueueAudio,
// and it is the worse of the two. The client drop happens after packetization,
// so the receiver sees a sequence gap and a timestamp jump and treats it as
// loss, which its jitter buffer is built for. A drop here never reaches the
// packetizer, so the RTP timestamp does not advance across the gap: the
// receiver hears a stream with the silence cut out of it, time-compressed and
// drifting further from the host with every drop, and nothing in the protocol
// reports that it happened. This channel therefore has to keep up, and the
// only consumer is the send loop, which does not block.
func (s *Stream) consume(chunk []byte) {
	packet, err := s.encoder.Encode(chunk, s.packet[:0])
	if err != nil {
		s.reportEncodeFailure(err)
		return
	}

	// Keep the grown buffer for the next frame.
	s.packet = packet

	if s.encodeFailures >= quietAfterFailures {
		log.Infof("audio encode recovered after %d failed frames", s.encodeFailures)
	}
	s.encodeFailures = 0

	// The buffer is reused, so the channel gets a copy.
	frame := make([]byte, len(packet))
	copy(frame, packet)

	select {
	case s.frames <- frame:
	default:
	}
}

// reportEncodeFailure logs the first few failures and then goes quiet, the way
// Source does and for the same reason: the log is /tmp/nanokvm-server.log,
// which S99vidiag and the supervisor both read, and a frame arrives fifty
// times a second.
//
// A failed frame is dropped rather than fatal. One bad 20 ms is not a reason
// to end a stream that may encode the next one.
func (s *Stream) reportEncodeFailure(err error) {
	s.encodeFailures++

	switch {
	case s.encodeFailures < quietAfterFailures:
		log.Warnf("audio encode failed: %s (frame %d)", err, s.encodeFailures)
	case s.encodeFailures == quietAfterFailures:
		log.Warnf("audio encode has failed %d times, and the last reason was %s; "+
			"it stays quiet until a frame encodes again", s.encodeFailures, err)
	}
}
```

Delete the `samples` and `frame` fields, the `decimator` field, and their initialisation. Remove `NewDecimator` from `NewStream`.

- [ ] **Step 4: Run the test to verify it passes**

Same command as Step 2. Expected: PASS.

- [ ] **Step 5: Write the failing test for a rejected frame**

Append to `encoder_test.go`:

```go
// A frame the encoder rejects is dropped. The stream stays open, because the
// next frame may encode: an encoder that fails once is not an encoder that is
// gone, and the listener has nowhere else to get audio from.
func TestStreamDropsARejectedFrameAndKeepsGoing(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("opus-packet")}
	encoder.setError(errors.New("synthetic encode failure"))

	stream := fakeStream(encoder)

	stream.consume(make([]byte, ChunkBytes))

	select {
	case frame, ok := <-stream.Frames():
		if !ok {
			t.Fatal("the frame channel closed because one frame failed to encode")
		}
		t.Fatalf("a rejected frame reached the channel as %q", frame)
	case <-time.After(100 * time.Millisecond):
	}

	// The next frame encodes, and it must arrive.
	encoder.setError(nil)
	go stream.consume(make([]byte, ChunkBytes))

	select {
	case frame, ok := <-stream.Frames():
		if !ok {
			t.Fatal("the frame channel closed instead of delivering the next frame")
		}
		if string(frame) != "opus-packet" {
			t.Errorf("the frame carried %q, want the encoder's packet", frame)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no frame arrived after the encoder recovered")
	}

	stream.Stop()
}

// An encoder that cannot be built is permanent. The stream has to end, or the
// consumer blocks on a channel nothing will close and the manager goes on
// believing audio is being sent.
func TestStreamClosesFramesWhenTheEncoderCannotBeBuilt(t *testing.T) {
	stream := NewStream()
	stream.newEncoder = func() (Encoder, error) {
		return nil, errors.New("synthetic construction failure")
	}

	stream.Start()

	done := make(chan struct{})
	go func() {
		for range stream.Frames() { //nolint:revive // draining is the point
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the frame channel stayed open after the encoder failed to build")
	}

	// Stop must still be safe, and must not block on a goroutine that never ran.
	stream.Stop()
}

func TestStreamGivesEveryChunkToTheEncoder(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("p")}
	stream := fakeStream(encoder)

	// The channel holds four, so three fit without a reader.
	for range 3 {
		stream.consume(make([]byte, ChunkBytes))
	}

	if got := encoder.count(); got != 3 {
		t.Errorf("the encoder saw %d chunks, want 3", got)
	}

	stream.Stop()
}

```

`TestStreamClosesTheEncoderWhenCaptureEnds` belongs with the other lifecycle
tests, because it needs a real child process to stop. It is added to
`stream_linux_test.go` in Step 8, which already carries the `linux` build tag
and already imports `os/exec`. Keeping it out of `encoder_test.go` is what lets
that file stay portable: nothing in it spawns a process, so it compiles and
runs anywhere.

- [ ] **Step 6: Run and verify**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision -v ./service/stream/audio/
```

Expected: the four tests in `encoder_test.go` PASS. `TestStreamEncodesSourceChunksIntoFrames` and the tests in `stream_linux_test.go` FAIL — they are fixed in the next steps.

- [ ] **Step 7: Replace the mu-law test in audio_test.go**

In `server/service/stream/audio/audio_test.go`, delete `TestStreamEncodesSourceChunksIntoFrames` in full — it asserts a 160 byte frame and `0xFF` for mu-law silence, and both facts are gone. `TestStreamDeliversWhatTheEncoderReturns` in `encoder_test.go` covers what it was for.

Leave `TestStreamClosesFramesOnStop` and the four `Available` tests untouched.

- [ ] **Step 8: Give the linux stream tests an encoder**

In `server/service/stream/audio/stream_linux_test.go`:

Replace `deliveringStream` with:

```go
func deliveringStream() *Stream {
	stream := NewStream()
	stream.newEncoder = func() (Encoder, error) {
		return &fakeEncoder{packet: []byte("opus-packet")}, nil
	}
	stream.source.minBackoff = time.Millisecond
	stream.source.maxBackoff = time.Millisecond
	stream.source.newCmd = func() *exec.Cmd {
		return exec.Command("yes")
	}

	return stream
}
```

In `TestStopEndsAStreamThatIsDelivering`, replace the frame length assertion:

```go
			if len(frame) != FrameSamples {
				t.Fatalf("frame is %d bytes, want %d", len(frame), FrameSamples)
			}
```

with:

```go
			if string(frame) != "opus-packet" {
				t.Fatalf("frame carried %q, want the encoder's packet", frame)
			}
```

In `TestStreamKeepsFramesOpenWhileCaptureRetries` and `TestStreamStopIsSafeWhileCaptureIsFailing`, both build a stream inline with `NewStream()`. Add the encoder to each, immediately after `NewStream()`:

```go
	stream.newEncoder = func() (Encoder, error) {
		return &fakeEncoder{packet: []byte("opus-packet")}, nil
	}
```

Without it those two streams take `newOpusEncoder`, which is the stub under `novision` and would close their frames at once — turning a test about capture retries into a test about a missing codec.

Finally, append the encoder lifecycle test to the same file. It lives here rather than in `encoder_test.go` because it needs a real child process, and `encoder_test.go` stays portable:

```go
// The encoder is a C object. Whatever ends the stream has to release it, or a
// gadget that is switched off and on again leaks one encoder per cycle.
func TestStreamClosesTheEncoderWhenCaptureEnds(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("opus-packet")}

	stream := NewStream()
	stream.newEncoder = func() (Encoder, error) { return encoder, nil }
	// An inert child: it blocks the way arecord does while the host plays
	// nothing, and Stop is what ends it.
	stream.source.newCmd = func() *exec.Cmd {
		return exec.Command("sh", "-c", "sleep 60")
	}

	stream.Start()
	stream.Stop()

	// Stop waits for the source goroutine, and that goroutine closes the
	// encoder on its way out.
	if !encoder.isClosed() {
		t.Error("the encoder was not closed when the stream stopped")
	}
}
```

- [ ] **Step 9: Delete the decimator and the mu-law encoder**

```bash
git rm server/service/stream/audio/resample.go server/service/stream/audio/resample_test.go \
       server/service/stream/audio/g711.go server/service/stream/audio/g711_test.go
```

- [ ] **Step 10: Run the audio package and verify it is green**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision -v ./service/stream/audio/
```

Expected: PASS, with no reference left to `FrameSamples`, `NewDecimator` or `EncodeULaw`.

- [ ] **Step 11: Switch the WebRTC layer**

`server/service/stream/webrtc/track.go` lines 12-13, replace:

```go
// audioClockRate is the RTP clock for G.711.
const audioClockRate = 8000
```

with:

```go
// audioClockRate is the RTP clock for Opus. Opus always uses 48000 in RTP,
// whatever rate the encoder was created with.
const audioClockRate = 48000
```

`server/service/stream/webrtc/manager.go`, in the `audioPacketizer` literal, replace `&codecs.G711Payloader{}` with `&codecs.OpusPayloader{}`.

`server/service/stream/webrtc/client.go` lines 220-225, replace:

```go
			webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypePCMU,
				ClockRate: audioClockRate,
				Channels:  1,
			},
```

with:

```go
			webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: audioClockRate,
				Channels:  audio.Channels,
			},
```

`server/service/stream/webrtc/audio.go` line 184, replace:

```go
	packets := m.audioPacketizer.Packetize(frame, audio.FrameSamples)
```

with:

```go
	// The sample count is per channel and fixed at 20 ms, whatever the encoded
	// packet happens to be long.
	packets := m.audioPacketizer.Packetize(frame, audio.SamplesPerFrame)
```

`server/service/stream/webrtc/audio_test.go` line 402, replace:

```go
	manager.deliverAudioFrame(make([]byte, audio.FrameSamples))
```

with:

```go
	// A plausible Opus packet. Nothing here decodes it; only its presence and
	// its length matter.
	manager.deliverAudioFrame(make([]byte, 240))
```

Also update the payload literal in `audioFrame()` at line 18 from `[]byte("mulaw")` to `[]byte("opus")`, and the comment above `audioPayloadType` in `webrtc/audio.go` line 10-12, replacing:

```go
	// audioPayloadType 0 is PCMU's static assignment. pion rewrites it per
```

with:

```go
	// audioPayloadType is a placeholder. Opus has no static assignment, and
	// pion rewrites it per
```

Leave the constant's value at 0: pion overwrites it during negotiation, which is what the rest of that comment already says.

- [ ] **Step 12: Run the whole suite off-device**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision ./...
```

Expected: every package ok.

Note: `service/stream/webrtc` has a heartbeat test that flakes under full-parallel load and passes alone. If a websocket heartbeat test fails, re-run that package by itself before treating it as a regression.

- [ ] **Step 13: Cross-check the target architecture**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/src" -v nanokvm-gomod:/go/pkg/mod \
  -w /src/server -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=riscv64 golang:1.25 \
  go build -tags novision ./...
```

Expected: no output.

- [ ] **Step 14: Run the device tests again**

The encoder tests from Task 2 must still pass, and the whole package now compiles with cgo. Use the on-device command from Global Constraints.

Expected: PASS.

- [ ] **Step 15: Commit**

```bash
rm -f audio.test
git add -A server/service/stream/audio server/service/stream/webrtc
git commit
```

Message body should explain why the audio package and the WebRTC layer change in one commit: `FrameSamples` carries two meanings that Opus separates, so removing it touches both, and a commit that changed only one would packetize 960 samples of mu-law.

---

### Task 4: Documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `server/README.md`, `server/README_ZH.md`, `server/README_JA.md` (only where they state the audio codec)

**Interfaces:**
- Consumes: everything from Tasks 1-3.
- Produces: nothing that code depends on.

- [ ] **Step 1: Find every place that states the old codec**

```bash
grep -rn "G.711\|G711\|mu-law\|mulaw\|PCMU\|8 kHz\|8kHz" --include="*.md" . | grep -v docs/superpowers
```

Each hit is either a statement about the audio path, which changes, or something unrelated. Read before editing.

- [ ] **Step 2: Update the cgo boundary section in CLAUDE.md**

In the "Backend architecture" section, the paragraph that begins "**The cgo boundary** is `common/kvm_vision.go`" now describes only one of two cgo boundaries. Extend it to say that `service/stream/audio/encoder_opus.go` is the second, that it links `server/dl_lib/libopus.a` — a static archive, so it adds no `NEEDED` entry — and that `tools/opusbench/build.sh` rebuilds it.

State plainly what `novision` now means: not "no vision" but "no device-native libraries", covering libkvm and libopus both.

- [ ] **Step 3: Record the device test command in CLAUDE.md**

The backend commands section documents `go test -tags novision ./...` for off-device work. Add the counterpart: the cgo encoder cannot be built on a workstation, so its tests are cross-compiled with `go test -c` in the builder image and run on the board. Include the full command from this plan's Global Constraints, and the warning that `audio.test` must not be committed.

- [ ] **Step 4: Update the audio description in the server README**

Find where `server/README.md` describes the audio path and change it to state 48 kHz stereo Opus at 96 kbit/s. Write the English source to ASD-STE100 where it fits: one instruction per sentence, active voice, present tense, keep the articles.

Then make the same change in `server/README_ZH.md` and `server/README_JA.md`. The repository rule is that the three stay in step.

- [ ] **Step 5: Verify no stale references remain**

```bash
grep -rn "G.711\|G711\|mu-law\|PCMU" --include="*.md" . | grep -v docs/superpowers
grep -rn "FrameSamples\|EncodeULaw\|NewDecimator" server/ --include=*.go
```

Expected: the first returns only historical notes in the design specs, which are dated records and stay as they are. The second returns nothing.

- [ ] **Step 6: Commit**

```bash
git add CLAUDE.md server/README.md server/README_ZH.md server/README_JA.md
git commit
```

---

### Task 5: Build, deploy, and prove it on hardware

**Files:** none changed. This task produces evidence.

**Interfaces:**
- Consumes: the whole of Tasks 1-4.
- Produces: a running server, and a recorded measurement.

- [ ] **Step 1: Build the server**

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM:/home/build/NanoKVM" \
  nanokvm-builder-local-197609-197121 bash -c \
  'cd /home/build/NanoKVM/server && go mod tidy && CGO_ENABLED=1 GOOS=linux GOARCH=riscv64 \
   CC=riscv64-unknown-linux-musl-gcc \
   CGO_CFLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d" \
   go build -buildvcs=false -ldflags "-X NanoKVM-Server/common/version.Build=dev.$(date +%Y%m%d.%H%M).$(git rev-parse --short HEAD)"'
```

- [ ] **Step 2: Patch the RPATH and check the dependencies**

`make app` does not run patchelf, so do it by hand:

```bash
MSYS_NO_PATHCONV=1 docker run --rm -v "D:/projects/NanoKVM/server:/src" -w /src ubuntu:24.04 \
  sh -c 'apt-get update -qq && apt-get install -y -qq patchelf \
         && patchelf --add-rpath "\$ORIGIN/dl_lib" NanoKVM-Server \
         && patchelf --print-needed NanoKVM-Server'
```

Expected: exactly `libkvm.so` and `libc.so`. A `libopus.so` in that list means the archive was not linked statically — stop and fix it, because the device has no such library and the server would not start.

- [ ] **Step 3: Deploy with the guarded script**

```bash
scp server/NanoKVM-Server root@10.0.0.222:/data/NanoKVM-Server.new
ssh root@10.0.0.222 'DEPLOY_TIMEOUT=200 sh /data/deploy-server /data/NanoKVM-Server.new'
```

Expected: `deploy: OK, serving within 200s and running what was installed`. The script snapshots the old binary and restores it if the new one does not serve.

- [ ] **Step 4: Prove the running process is the new build**

The file's timestamp proves nothing — the old server survives `killall` and keeps answering. Compare the running image:

```bash
md5sum server/NanoKVM-Server
ssh root@10.0.0.222 'pid=$(pidof NanoKVM-Server); echo "pid $pid"; md5sum /proc/$pid/exe /kvmapp/server/NanoKVM-Server'
```

Expected: three identical digests.

- [ ] **Step 5: Confirm the gadget and the codec**

```bash
ssh root@10.0.0.222 'sh /data/audiodiag.sh; echo "exit=$?"'
```

Expected: exit 0 if the managed host is streaming, exit 2 if it is not. Exit 1 means the KVM is misconfigured and the deploy is not the cause.

- [ ] **Step 6: Listen**

Open the KVM in a browser, unmute the speaker, and play a 440 Hz tone on the managed host. A previous session proved this chain at 8 kHz; at 48 kHz the difference should be obvious on anything with high-frequency content — play music rather than a sine to hear it.

While it plays, confirm capture is actually running:

```bash
ssh root@10.0.0.222 'ps w | grep "[a]record"; cat /proc/asound/UAC1Gadget/pcm0c/sub0/status'
```

Expected: one `arecord` process, and `state: RUNNING` with `hw_ptr` advancing between two reads. A frozen `hw_ptr` means the host is not driving the gadget, which is a host problem and not this change.

- [ ] **Step 7: Check the log is quiet**

```bash
ssh root@10.0.0.222 'grep -c "audio" /tmp/nanokvm-server.log; tail -20 /tmp/nanokvm-server.log'
```

Expected: no repeating `audio encode failed` lines. A handful at startup followed by silence is the throttle working; a line every 20 ms is a defect.

- [ ] **Step 8: Record the result and merge**

Append a short section to `docs/superpowers/specs/2026-08-09-opus-audio-design.md` under a heading `## Verified on hardware`, stating the date, the build stamp, what was heard, and the `arecord`/`hw_ptr` evidence. Commit it.

Then merge into the branch the device runs:

```bash
git checkout fork/integration
git merge --no-ff feat/opus-audio
```

Ask the user before pushing.

---

## Self-Review

**Spec coverage.** Every section of the design maps to a task: the archive and the measurement harness to Task 1; the `Encoder` interface, the constants, the cgo wrapper, the stub and the chunk-length check to Task 2; the pipeline switch, the deleted decimator, the WebRTC changes and every listed test to Task 3; the documentation list to Task 4; the build, deploy and hardware verification to Task 5. The design's "Risks" section needs no task — the bandwidth rise and the upstream point are statements, and the in-process codec risk is implemented as the length check in Task 2.

**Placeholders.** None. Every code step carries the code, every command carries its expected output, and no step says "add error handling" or "similar to Task N".

**Type consistency.** `Encoder.Encode(pcm []byte, dst []byte) ([]byte, error)` and `Encoder.Close()` are declared in Task 2 Step 3 and used with those exact signatures by `opusEncoder` in Task 2 Step 4, by `fakeEncoder` in Task 3 Step 1, and by `Stream.consume` in Task 3 Step 3. `newOpusEncoder() (Encoder, error)` matches the `newEncoder` field type in Task 3 Step 3. `SamplesPerFrame`, `ChunkBytes`, `Channels` and `maxPacketBytes` are defined once, in Task 2, and every later use matches. `quietAfterFailures` is reused from `source.go`, where it already exists.
