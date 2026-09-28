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

All names start with `ironkvm_`, except the node_exporter families at the end of this section. Types follow the Prometheus conventions: `_total` counters,
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

Host watchdog:

| Metric | Type | Source |
| --- | --- | --- |
| `ironkvm_watchdog_actions_total{action}` | counter | `watchdog.ActionCounts()`, `action` is `reset` or `power` |

### The node_exporter families

The endpoint also writes a subset of what node_exporter's default collectors write. The names,
the labels and the units are the same as node_exporter's, so a dashboard made for node_exporter
(for example "Node Exporter Full") reads this endpoint unchanged. The board then needs no
separate node_exporter process. These families are the only ones without the `ironkvm_` prefix.

| Section | Families | Source |
| --- | --- | --- |
| `node_system` | `node_uname_info`, `node_time_seconds`, `node_boot_time_seconds`, `node_cpu_seconds_total{cpu,mode}`, `node_context_switches_total`, `node_intr_total`, `node_forks_total`, `node_procs_running`, `node_procs_blocked`, `node_load1`, `node_load5`, `node_load15` | `/proc/sys/kernel`, `/proc/stat` (USER_HZ 100), `/proc/loadavg` |
| `node_memory` | `node_memory_<field>_bytes` for each kB field, `node_memory_<field>` for a count such as `HugePages_Total` | `/proc/meminfo` |
| `node_pressure` | `node_pressure_{cpu,io,memory}_waiting_seconds_total` (the `some` line), `..._stalled_seconds_total` (the `full` line) | `/proc/pressure` |
| `node_filesystem` | `node_filesystem_{size,free,avail}_bytes`, `node_filesystem_files`, `node_filesystem_files_free`, `node_filesystem_readonly`, labels `device`, `fstype`, `mountpoint` | `/proc/1/mounts` and `statfs(2)` |
| `node_network` | the sixteen `node_network_{receive,transmit}_*_total{device}` counters, `node_network_up{device}` | `/proc/net/dev`, `/sys/class/net/<device>/operstate` |
| `node_disk` | `node_disk_{reads,writes}_completed_total`, `..._merged_total`, `node_disk_read_bytes_total`, `node_disk_written_bytes_total`, `node_disk_{read,write}_time_seconds_total`, `node_disk_io_now`, `node_disk_io_time_seconds_total`, `node_disk_io_time_weighted_seconds_total` | `/proc/diskstats` |
| `node_thermal` | `node_hwmon_temp_celsius{chip,sensor}`, `node_hwmon_chip_names{chip,chip_name}`, `node_thermal_zone_temp{type,zone}` | `/sys/class/hwmon`, `/sys/class/thermal` |

The filters are node_exporter's defaults, with three differences:

- The filesystem section also leaves out `tmpfs` and `ramfs`. They hold RAM, which
  `node_memory_Shmem_bytes` already reports, and the board has several of them.
- The network section leaves out `lo`.
- `node_uname_info` reads the `machine` label from `/proc/sys/kernel/arch`. A kernel before 6.1
  does not have that file, and then the label is the architecture the server was built for.

The disk section keeps node_exporter's default exclusion, `^(z?ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\d+n\d+p)\d+$`.
That expression does not match an `mmcblk0pN` partition, so each partition has its own series
beside `mmcblk0`, as on node_exporter.

`promtool check metrics` reports the `node_memory_*` names as camelCase. node_exporter's names
get the same warning, and a dashboard needs those names.

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
