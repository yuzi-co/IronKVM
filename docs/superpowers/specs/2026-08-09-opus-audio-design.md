# Full-rate Opus audio for NanoKVM — Design

**Status:** approved 2026-08-09
**Branch:** `feat/opus-audio`, cut from `fork/integration`
**Supersedes:** the codec half of `2026-08-04-usb-audio-design.md`. The capture half of that
document — the UAC1 gadget, the `arecord` child, availability, start and stop — stays in force.

## Summary

The USB audio gadget delivers 48 kHz stereo 16-bit. The server throws almost all of it away: a
129-tap FIR decimates to 8 kHz mono, and G.711 mu-law carries the result as PCMU. The audio
ceiling is 3.4 kHz, which is telephone quality.

This design replaces that path with Opus at the rate the gadget already captures. It sends
48 kHz stereo over the same WebRTC connection, at 96 kbit/s, encoder complexity 3.

Measurements on the device decide the shape of this change. The FIR is what costs CPU, and
Opus removes it. Full-rate Opus therefore costs about the same as the telephone path it
replaces.

## What we measured, 2026-08-09

Every number below comes from the device at 10.0.0.222 (SG2002, one C906B core). CPU time
comes from `CLOCK_PROCESS_CPUTIME_ID`, so other work on the board does not inflate it. The
harness encodes 10 seconds of 48 kHz audio in 20 ms frames per row, and `core_%` is the share
of one core that the row needs to keep up with real time.

`libopus` is version 1.5.2. It is cross-compiled with the repository's own toolchain and the
repository's own flags: `-O3 -mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany
-mabi=lp64d`.

### The current path, and Opus beside it

| what | channels | bitrate | complexity | core % |
| --- | --- | --- | --- | --- |
| **the Go pipeline today** (FIR + mu-law) | stereo in, mono out | 64k | — | **4.66** |
| Opus | stereo | 96k | 0 | **4.46** |
| Opus | stereo | 96k | 1 | 5.04 |
| Opus | stereo | 96k | **3** | **5.60** |
| Opus | stereo | 96k | 5 | 7.88 |
| Opus | stereo | 96k | 8 | 12.92 |
| Opus | stereo | 96k | 10 | 12.90 |
| Opus, pink noise | stereo | 96k | 3 | 6.18 |
| Opus, pink noise | stereo | 96k | 5 | 8.20 |
| Opus | mono | 64k | 3 | 3.53 |
| Opus **decode** | stereo | 96k | — | 3.79 |

The Go figure is the real one. A temporary `BenchmarkPipeline` ran the committed `Decimator`
and `EncodeULaw` on the device — cross-compiled with `go test -c`, `GOARCH=riscv64` — and
reported 931,543 ns for one 20 ms chunk. A chunk arrives every 20,000,000 ns, so the path
needs 4.66% of the core. That benchmark is not committed: this change deletes the code it
measures, so the number is recorded here instead.

Selected settings — stereo, 96 kbit/s, complexity 3 — cost 5.60%. The change is therefore
about one percentage point of one core, and it lifts the audio ceiling from 3.4 kHz to the
full band.

### Bitrate hardly matters; the mode does

At complexity 5, stereo: 32k costs 6.99%, 64k costs 7.63%, 96k costs 7.88%, 128k costs 8.35%.
Four times the bitrate costs one fifth more CPU. Bitrate is therefore not a CPU lever.

`OPUS_APPLICATION_VOIP` is worse than `OPUS_APPLICATION_AUDIO` for this workload. Mono at
32 kbit/s in VOIP mode costs 10.71%, which is more than stereo at 96 kbit/s in AUDIO mode.
SILK is expensive. The gadget carries desktop audio rather than a phone call, so AUDIO is both
the correct mode and the cheap one.

### Fixed point loses

libopus was also built with `--enable-fixed-point`. It is slower on this core: 9.78% against
7.88% at complexity 5, stereo, 96 kbit/s. The C906B has a hardware FPU and a double-float ABI,
so the fixed-point paths that help a soft-float core only add work here. Decode is marginally
faster in fixed point (3.43% against 3.79%), which does not pay for the encoder.

**The float build ships.**

### Two measurement traps, recorded so the numbers can be trusted

The first harness reported the FIR at 0.10% of the core. That is 10.3 million multiply-adds in
10 ms, which is one per cycle on a 1 GHz in-order core — the theoretical limit, and therefore
not a real result. The filter output was unused, so the compiler deleted the loop. A checksum
that reaches `printf` restores it, and the honest C figure is 1.64%.

The C figure is still not the Go figure. The same filter written in Go costs 4.66%, which is
2.8 times more. Bounds checks and the `% TAPS` ring index account for it. Only a benchmark of
the committed Go code answers "what does this cost today", so that is what the table reports.

An earlier estimate in this session put Opus at 61% of the core. That number came from
ffmpeg's native Opus encoder, which the device carries and which is marked experimental.
libopus is about five times faster than it. The estimate was wrong and the measurement
replaces it.

## Scope

In scope:

- `libopus` as a committed static archive, and the cgo wrapper around it.
- The audio package: encode Opus instead of FIR and mu-law.
- The WebRTC layer: Opus payloader, 48 kHz clock, stereo track.
- `tools/opusbench/`, the harness that produced the table above.
- Documentation that states the codec or the meaning of the `novision` tag.

Out of scope:

- The UAC1 gadget, `S03usbdev`, and the `/boot/usb.uac` marker. They do not change.
- The `arecord` child, its retry loop and its log throttle. They do not change.
- Microphone support, which needs `p_chmask`, an inbound track and HTTPS. It is a separate
  design.
- The MJPEG and direct video paths, which carry no audio.

## Architecture

### New files

| path | what |
| --- | --- |
| `server/dl_lib/libopus.a` | libopus 1.5.2, riscv64, float, built with the repository's flags |
| `server/include/opus/` | `opus.h`, `opus_defines.h`, `opus_types.h` |
| `server/service/stream/audio/encoder.go` | the `Encoder` interface and the constants. Portable. |
| `server/service/stream/audio/encoder_opus.go` | `//go:build !novision`. The cgo wrapper. |
| `server/service/stream/audio/encoder_stub.go` | `//go:build novision`. Construction returns an error. |
| `server/service/stream/audio/encoder_test.go` | the fake, and the tests that use it |
| `tools/opusbench/` | the C harness, its build script, and a README |

### Deleted files

`resample.go`, `resample_test.go`, `g711.go` and `g711_test.go`. Nothing calls them after this
change, and a dead decimator invites somebody to reconnect it.

### The interface

```go
// Encoder turns one 20 ms chunk of 48 kHz stereo S16_LE into one packet.
type Encoder interface {
    Encode(pcm []byte, dst []byte) ([]byte, error)
    Close()
}
```

The interface exists so that the pipeline is testable on a workstation. The archive is
riscv64, so the real encoder does not link anywhere else. A fake implements this interface,
and every test of `Stream` uses it.

`Encode` appends to `dst` and returns the extended slice, which is the convention that
`EncodeULaw` already uses in this package.

### Constants

```go
const (
    SampleRate      = 48000
    Channels        = 2
    SamplesPerFrame = 960          // 20 ms per channel
    ChunkBytes      = SamplesPerFrame * Channels * 2
    Bitrate         = 96000
    Complexity      = 3
)
```

`ChunkBytes` is 3840, which is what it is today. The `arecord` command line does not change:
20 ms of 48 kHz stereo is the same number of bytes whichever codec follows it.

`FrameSamples` is removed rather than renamed. Today it means two things at once — the RTP
sample count, and the length of a frame in bytes — because mu-law codes one byte per sample.
Opus separates them: the sample count stays 960 and the packet length varies. A rename would
let the old meaning survive in a call site; removal will not compile.

`Bitrate` and `Complexity` are constants and not configuration. No operator has a reason to
choose 5 over 3 on this board, and a knob in `server.yaml` is one more thing to support.

### Changes to `server/service/stream/webrtc`

Three lines, and one call:

- `track.go` — `audioClockRate` becomes 48000.
- `manager.go` — the audio packetizer takes `&codecs.OpusPayloader{}`.
- `client.go` — the track takes `webrtc.MimeTypeOpus`, clock rate 48000, `Channels: 2`.
- `audio.go` — `Packetize(packet, audio.SamplesPerFrame)`.

`RegisterDefaultCodecs()` already registers Opus, so negotiation of the codec itself needs
nothing. The static payload type constant keeps its comment: Opus is a dynamic type and pion
rewrites it per connection.

That is not the whole story for stereo. RFC 7587 §7.1 defaults the `stereo` fmtp parameter to 0,
and Chrome's offer omits it, so an unmodified negotiation configures Chrome's decoder for one
channel and it downmixes whatever the device sends. The device pays for a stereo encode either
way — 5.60% of the core against 3.53% for mono, from the table above — so an unmodified
negotiation buys nothing for that cost. The browser's own offer is what configures its receive
side, and the answerer cannot override it, so `web/src/lib/sdp-opus.ts` rewrites the offer's Opus
`a=fmtp` line to add `stereo=1;sprop-stereo=1` before it becomes the local description. See that
file for the rewrite and why it targets the offer rather than the answer.

## Data flow

```
arecord (48 kHz stereo S16_LE)
  → Source.runOnce reads exactly ChunkBytes (3840) with io.ReadFull
  → Stream.consume
      → Encoder.Encode  → one Opus packet, about 240 bytes at 96 kbit/s
      → copy onto the frames channel (capacity 4), non-blocking
  → WebRTCManager.sendAudioStream
      → Packetize(packet, 960) once
      → every client that negotiated an audio track gets the same packets
```

The channel keeps its four-frame slack and its non-blocking send. The comment in `consume`
that explains why a drop here is worse than a drop at the client stays true and stays
relevant: a frame dropped before packetization does not advance the RTP timestamp, so the
receiver hears time-compressed audio and nothing reports it.

## Error handling

Three failures, three different answers.

**The encoder cannot be created.** `Stream.Start` builds it. On failure the stream logs one
line and closes its frames channel. This is permanent and it is deliberately unlike the source
retry loop: a host that plays nothing is idle and deserves an unbounded retry, but an encoder
that cannot be constructed will not become constructible. `sendAudioStream` already ends when
the channel closes and calls `clearAudioStream`, so the manager learns that audio stopped.

In practice `opus_encoder_create` fails only on invalid arguments, and every argument here is
a compile-time constant. The path exists because the API can fail, not because we expect it
to.

**`opus_encode` returns an error.** The frame is dropped and counted. The log uses the same
"five lines then quiet until it recovers" throttle that `Source` uses, and for the same
reason: `/tmp/nanokvm-server.log` is read by `S99vidiag` and the supervisor, and an unbounded
line rate wears the card. One bad 20 ms is not a reason to end a stream.

**A short chunk reaches the C boundary.** `Encode` checks `len(pcm) == ChunkBytes` and returns
an error before it calls into C. This is a new class of risk: the encoder runs in the server
process. An `arecord` that crashes today is a child, and the retry loop restarts it. A libopus
call that reads past the end of a buffer takes the whole server down, and this board already
has a documented case where a segfault costs a reboot. `io.ReadFull` guarantees full chunks
two files away; the check makes the guarantee local.

## Testing

Test-driven, red before green, one behaviour per test.

Off-device, with the fake encoder, in `go test -tags novision ./...`:

- `Stream` gives each chunk to the encoder exactly once.
- The packet the encoder returns is the packet that arrives on `Frames`.
- A chunk the encoder rejects is dropped, and `Frames` stays open.
- An encoder that fails to construct closes `Frames`.
- `Stop` is safe while an encode is in flight.

Two existing test files change. `stream_linux_test.go` builds streams whose child is `yes` or
`sh -c "exit 1"`; those streams must take the fake encoder, because the real one does not link
on a workstation. `webrtc/audio_test.go` refers to `audio.FrameSamples` and moves to
`SamplesPerFrame`.

On the device:

- `tools/opusbench` encodes and decodes with the shipped archive. It is the round-trip check
  and the way to re-measure after a libopus change.
- Deploy, open a viewer, and listen. A 440 Hz tone on the managed host proved the chain on
  2026-08-09 at 8 kHz; the same test at 48 kHz should sound obviously different.

The Opus bitstream is not asserted in Go. libopus has its own test suite and we are not going
to reimplement it.

## Build and deploy

`encoder_opus.go` carries its own directives:

```go
#cgo CFLAGS: -I${SRCDIR}/../../../include
#cgo LDFLAGS: ${SRCDIR}/../../../dl_lib/libopus.a -lm
```

`${SRCDIR}` is used so the flags do not depend on the working directory. `common/kvm_vision.go`
uses relative paths because it sits one level below `server/`; this package sits three levels
below.

Nothing else in the build changes. `make app` still cross-compiles in the builder image, and
`patchelf --add-rpath '$ORIGIN/dl_lib'` is still required afterwards. A static archive adds no
`NEEDED` entry, so `patchelf --print-needed` must still report exactly `libkvm.so` and
`libc.so`.

The binary grows by about 400 KB. **The deploy stays one file**, which is why a static archive
was chosen over a shared library: `/root/nanokvm-deploy-libkvm.sh` snapshots and restores the
server binary, and a second artefact would give it a pair to keep consistent that it does not
know about. Deploy with `DEPLOY_TIMEOUT=200`; a restart takes about 137 seconds.

## Risks

**Audio bandwidth rises from 64 to 96 kbit/s.** This is nothing on a LAN. Over Tailscale on a
poor link it is 50% more audio, competing with video on the same connection.

**cgo moves the codec into the server process.** Covered under error handling. It is the one
genuine uptime risk in this change.

**This work cannot go upstream.** It adds a binary to `server/dl_lib/`, which CLAUDE.md
records as belonging to the fork. There is no extraction branch for it, and a pull request to
`sipeed/NanoKVM` would have to build libopus rather than commit it.

**`libopus.a` is a committed binary that nothing rebuilds automatically.** `tools/opusbench`
carries the build script, so the archive can be reproduced, and the README records the exact
version and flags.

## Alternatives considered

**G.722 at 16 kHz.** Costs about 1% of the core and doubles the audio ceiling to 7 kHz. It was
the recommendation until libopus was measured. Opus gives the full band for the same CPU, so
G.722 buys nothing.

**Raw PCM over the WebRTC data channel, decoded in an AudioWorklet.** Bit-exact and needs no
encoder, but costs 1.5 Mbit/s and a jitter buffer that we would have to write. Opus at 96
kbit/s is 16 times cheaper on the wire and the browser's own jitter buffer handles it.

**ffmpeg as a child process.** The device already carries ffmpeg and libavcodec. Its native
Opus encoder costs 61% of the core, and libopus is five times faster. Rejected on measurement.

**A shared `libopus.so` in `dl_lib`, beside `libkvm.so`.** Matches the existing pattern, but
makes the deploy two files and gives the guarded restore an inconsistent pair to recover.

**A vendored cgo module such as `hraban/opus`.** Needs no committed binary, but adds a Go
dependency, pins libopus 1.3.1 rather than the 1.5.2 that was measured, and recompiles all of
libopus on every build.

## Verified on hardware

**Date:** 2026-08-10. **Build stamp:** `dev.20260810.1121.9ffb0251`. **Device:** 10.0.0.222.

### The archive links statically

`patchelf --print-needed NanoKVM-Server` reports `libkvm.so` and `libc.so`, and nothing else.
This is the gate that matters for uptime. The device carries no `libopus.so`, so a dynamic link
would give a server that cannot start. The list agrees with the committed library.

### The deploy replaced the running process

The guarded script `/data/deploy-server` ran with `DEPLOY_TIMEOUT=200`. It snapshotted the old
binary, installed the new one, and reported `deploy: OK, serving within 200s and running what
was installed` 43 seconds after the restart. The process id changed from 1044 to 1450.

Three digests agree, which is the only proof that the new image is the one that runs:

```
6b63cc9f073b850c0f30a36326619014  (host) server/NanoKVM-Server
6b63cc9f073b850c0f30a36326619014  /proc/1450/exe
6b63cc9f073b850c0f30a36326619014  /kvmapp/server/NanoKVM-Server
```

The server answers on loopback with HTTP 401, which shows that it bound its port and read its
config. The log records no `audio`, `opus`, `panic` or `SIGSEGV` line.

### libopus 1.5.2 runs on the C906

`tools/opusbench`, built against the committed `server/dl_lib/libopus.a` and run on the device
under `nice -n 15`, reproduces the numbers this design was approved on. The shipped
configuration is the `stereo / 96k / cx 3` row:

```
what                   chans      rate    cx    cpu_s   core_%    kbit/s
g711-fir(today)        stereo        -     -    0.164     1.64      64.0
opus/audio/tone        stereo      96k     3    0.559     5.59      96.4
opus/audio/pink        stereo      96k     3    0.623     6.23      96.4
opus-DECODE/tone       stereo      96k     5    0.374     3.74         -
```

5.59% against the 5.60% this design records, which reproduces the measurement the decision was
made on. The pink noise row, 6.23%, is the more honest figure for real content.

Opus costs more than the path it replaces, as the design states: 5.59% against the 4.66% the Go
G.711 code measured. The rise is under 1% of one core, and it buys the full audio band. The
`g711-fir` row above is the C transcription at 1.64%, not the Go cost — do not read it as the
comparison.

### The capture path did not run

`/data/audiodiag.sh` exits 2: the KVM is correct and the managed host is not streaming. The
gadget reports `chmask 3 / rate 48000 / sample size 2`, which is the 48 kHz stereo this design
consumes. `arecord` and the `UAC1Gadget` card are present. No frame arrives, and `arecord`
reports `pcm_read: read error: I/O error`.

`/proc/asound/UAC1Gadget/pcm0c/sub0/status` reads `closed` on two reads four seconds apart, and
no `arecord` process runs. That is the correct idle state: capture starts only when a viewer
opens audio, and the host is not driving the gadget in any case. The managed host has no
`snd-usb-audio` module, which `unraid-needs-a-built-usb-audio-module` already records. This is
a host condition and not a fault in this change.

### What this does not verify

**Nobody listened.** The end-to-end chain — browser, WebRTC, Opus decode, speaker — is
unproven on hardware. Confirm it by hand.

**The in-process encoder never encoded a frame on the device.** `opus_encoder_create` and
`opus_encode` are proven on this CPU by `opusbench`, and the Go pipeline around them is proven
by the unit tests, but the two have not run together on the board. They cannot until the host
streams audio. This is the residual risk the design names under "cgo moves the codec into the
server process".

**The bandwidth rise is not measured on the wire.** 96 kbit/s is the encoder's own figure.

**Update, 2026-08-10: the musl thread-stack risk is now verified.** `cgocall` runs `opus_encode`
on the calling M's `g0` stack, and this binary links with `riscv64-unknown-linux-musl-gcc`, whose
default thread stack is 128 KB rather than glibc's 8 MB. libopus here is a float build with no
`--enable-alloca`, so the CELT encoder's C99 VLAs are sized from the frame on whatever stack they
get. `opusbench` and the unit tests above do not cover this: `opusbench` calls `opus_encode` from
a static binary's `main()` on the process's own large stack, and the other tests in this file
very likely stay on the runtime's first M, whose `g0` is also the large main stack — not a musl
thread stack. Production runs the encoder on a goroutine that can land on any M the runtime
creates.

`TestOpusEncoderUnderGOMAXPROCSConcurrency` in `encoder_opus_test.go` closes that gap: it raises
`GOMAXPROCS` and runs 8 goroutines, each with its own encoder, each encoding 500 chunks (4,000
encodes total), which is what forces the runtime to create new Ms with their own musl-sized `g0`
stacks. Cross-compiled and run on the device on 2026-08-10, it passed in 5.31s alongside the rest
of the package's tests, with no crash. The thread-stack risk is verified rather than reasoned
about.
