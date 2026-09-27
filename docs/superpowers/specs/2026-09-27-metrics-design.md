# Prometheus metrics for the KVM itself

Issue: yuzi-co/ironkvm-dist #15.

## Goal

A Prometheus scrape of the board shows, on a dashboard, what the board's failures have been
about: memory pressure, ION, the streams, the USB link and HID, and the server's own restarts.
A pressure wedge or a HID stall is then visible before a person notices it. The numbers also
give later work (#1, #2, #3, #4) the memory and CPU figures each of them has to state before it
ships.

## Decisions

- **The server serves it.** No exporter daemon: the board has about 110 MB available after
  boot, and a second Go process would cost more than everything it measures.
- **No `prometheus/client_golang`.** The text exposition format is a few lines of code. The
  library adds its own collectors, a registry and several MB of binary for features this
  endpoint does not use. A small writer in `service/metrics` renders the format by hand.
- **Read at scrape time.** Every gauge is read from `/proc`, `/sys` or the server's own state
  when Prometheus asks, so nothing runs between scrapes. The only new state is a handful of
  counters, each an `atomic.Uint64` beside the code that counts.
- **Path `/api/metrics`.** The static file handler skips only `/api/`, so a path outside it would
  go through the web middleware first. Prometheus takes any `metrics_path`.
- **Auth: the existing token check.** `middleware.CheckToken()` already accepts an API key as
  `Authorization: Bearer <key>`, which is what Prometheus's `authorization` block sends. Any
  role may read it: it holds no secret. No new setting, and no unauthenticated mode.

## Metrics

All names start with `ironkvm_`. Types follow the Prometheus conventions: `_total` counters,
`_bytes` and `_seconds` units.

Memory, from `/proc/meminfo`, `/proc/vmstat`, `/proc/pressure/memory` and `/sys/block/zram0`:

| Metric | Type | Source |
| --- | --- | --- |
| `ironkvm_memory_bytes{kind}` | gauge | meminfo `MemTotal`, `MemAvailable`, `MemFree`, `SwapTotal`, `SwapFree`, `CmaTotal`, `CmaFree` |
| `ironkvm_swap_pages_total{direction}` | counter | vmstat `pswpin`, `pswpout` |
| `ironkvm_zram_bytes{kind}` | gauge | mm_stat `orig_data_size`, `compr_data_size`, `mem_used_total` (reuses `parseMmStat`) |
| `ironkvm_pressure_stall_seconds_total{resource="memory",kind}` | counter | PSI `some` and `full` `total=` (microseconds / 1e6). Absent when PSI is off. |
| `ironkvm_cgroup_memory_bytes{group}` | gauge | `/sys/fs/cgroup/<group>/memory.current` for `kvm` and `addons` |
| `ironkvm_cgroup_memory_events_total{group,event}` | counter | `memory.events` `high`, `max`, `oom`, `oom_kill` |

ION, from `ion.Read()`:

| Metric | Type |
| --- | --- |
| `ironkvm_ion_bytes{kind}` | gauge: `total`, `used` |

The issue asks for the ION peak. The heap dump has no peak, and `ion.Init` resets a watermark
the server keeps for another purpose. The peak is left to Prometheus (`max_over_time`), at the
scrape interval's resolution. Recorded as a known limit.

Streams:

| Metric | Type | Source |
| --- | --- | --- |
| `ironkvm_stream_viewers{path}` | gauge | the HDMI viewer snapshot: `mjpeg`, `direct`, `webrtc` |
| `ironkvm_stream_fps` | gauge | `stream.GetFrameRateCounter().GetFPS()` |
| `ironkvm_stream_sent_bytes_total{path}` | counter | new: counted where `mjpeg` and `direct` write a frame |
| `ironkvm_stream_suppressed_frames_total` | counter | `mjpeg` `Suppressed()` |

WebRTC bytes are not counted: pion sends them, and its stats API would need a collector per
peer. Left out, and recorded as a known limit.

USB and HID:

| Metric | Type | Source |
| --- | --- | --- |
| `ironkvm_usb_udc_state{state}` | gauge, 1 for the current state | `readUSBLink()` |
| `ironkvm_hid_endpoint_state{endpoint,state}` | gauge, 1 for the current state | `(*Hid).Status()` |
| `ironkvm_hid_write_errors_total{endpoint,kind}` | counter | new: `classifyWriteResult`, `kind` is `stalled` or `detached` |
| `ironkvm_usb_recoveries_total{action}` | counter | new: the USB watchdog's rebind and rebuild |

Server:

| Metric | Type | Source |
| --- | --- | --- |
| `ironkvm_build_info{image,kernel,app}` | gauge, always 1 | the functions `GetInfo` uses |
| `ironkvm_process_start_time_seconds` | gauge | set once at start. `changes()` over it counts restarts. |
| `ironkvm_process_resident_bytes` | gauge | `/proc/self/status` `VmRSS` |
| `ironkvm_go_heap_bytes`, `ironkvm_go_goroutines` | gauge | `runtime.ReadMemStats`, `runtime.NumGoroutine` |

Supervisor actions are not in the first version. `S98supervise` is a shell script that only
appends to `/data/supervise.log`. Counting them would mean parsing that log on every scrape, or
giving the script a counter file. Restarts it causes already show in
`ironkvm_process_start_time_seconds`. Slot state is also left out: no Go code reads it today.

## Units

- `service/metrics/writer.go`: a writer that emits `# HELP`, `# TYPE` and sample lines, escapes
  label values, and writes each family once.
- `service/metrics/collect.go`: one function per section above, each taking an `io.Writer`-backed
  writer. A section whose source is missing (no PSI, no zram, no cgroup) writes nothing and
  never fails the scrape.
- `service/metrics/handler.go`: `GET /api/metrics`, content type
  `text/plain; version=0.0.4`.
- Counters: `atomic.Uint64` fields or package-level variables in `service/hid` and the
  streamers, read through small exported getters.

Readers that already exist (`parseMmStat`, `readUSBLink`, `ion.Read`) are reused through exported
wrappers, not copied. File paths are package variables so tests point them at fixtures.

## Testing

- Unit tests per section, against fixture files for `/proc` and `/sys`, compare the exact
  exposition text.
- A writer test: label escaping, and each family's `HELP` and `TYPE` written once.
- A test that a missing source leaves its section out and the rest is still served.
- On the board: `curl -H "Authorization: Bearer <key>" http://<kvm>/api/metrics` passes
  `promtool check metrics`. The scrape takes under 50 ms, and the RSS before and after a
  1-hour scrape at 15 s is recorded.

## Out of scope

Grafana dashboards, alert rules, push gateways, and an unauthenticated mode.
