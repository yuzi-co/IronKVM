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

The server ships stereo, 96 kbit/s, complexity 3, which measured 5.60% of the
core.

Read the `g711-fir` row with care, because two different numbers describe the
path Opus replaced. That row is a C transcription of the decimator and the
mu-law encoder, compiled at `-O3`, and it measures 1.64%. The Go code that
actually ran in the server measures 4.66% for the same work, benchmarked on
the device with a cross-compiled `go test -c` binary. Go pays for bounds
checks and for the ring index, and the gap is 2.8 times.

The comparison that decides the design is therefore 5.60% against 4.66%: what
the server spends now, against what it spent before.

Do not build libopus in fixed point. The C906B has a hardware FPU, and the
fixed-point build measured 9.78% against 7.88% for float at complexity 5.

The `g711-fir` row carries a checksum. It is load-bearing: without a consumer
for the filtered samples the compiler deletes the 129-tap loop and the row
reports an impossible 0.10% of the core.
