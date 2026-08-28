# Fork ecosystem survey

Survey date: 2026-08-17. Companion to `2026-08-17-upstream-adoption-backlog.md`, which covers
`sipeed/NanoKVM` and `sipeed/LicheeRV-Nano-Build`. This document covers the third-party forks and the
OneKVM organization.

## Method

`sipeed/NanoKVM` has 316 forks. 86 of them were pushed after they were created. Each of those 86 was
compared against `sipeed/NanoKVM:main`, and 20 are ahead on their default branch. Named forks were
also checked branch by branch, because several keep their work off the default branch.

```
ahead  behind  fork                            branch
  235      39  mrjeeves/NanoKVM                main
  232      41  woffko/Hardened_NanoKVM         main
  201      79  pi-bmc/nanokvm-app              main
   32      11  Schattenwelt/NanoKVM            fork-live
   16     222  Ihab-Zhaika/NanoKVM             main
   16      39  mjc/NanoKVM                     mjc/rust-webrtc-sidecar
   10      41  JuergenLeber/NanoKVM            main
    5      64  BeaconCat/NanoKVM               main
    4     333  C2022H/NanoKVM                  iso-de-keyboard
    3     110  ZagidullinRuslanG/NanoKVM       main
    3     128  lazerusrm/NanoKVM               main
    3     436  PuwenHsiao/NanoKVM              main
    1       3  S33G/NanoKVM                    main
    1     319  twormtwo/NanoKVM                main
    1     332  aliaksei-by/NanoKVM             main
    1      39  deasmi/NanoKVM                  main
    1      39  DrEVILish/NanoKVM               main
    1      39  ethanperrine/NanoKVM            main
    1     542  maggo1404/NanoKVM               main
    1      83  Voranto/NanoKVM                 main
```

`mjc/NanoKVM` keeps `main` at zero ahead and puts everything on 20 topic branches, so the table
understates it.

---

## Tier 1: forks with real engineering

### pi-bmc/nanokvm-app, 201 ahead, active 2026-08-16

The application is restructured into Go packages. The changed paths concentrate in `api/redfish` (31
files), `pkg/video` (28), `pkg/firmware` (23), `kvmapp/system` (16), `api/vm` (12), `server/dl_lib`
(11), `pkg/usbgadget` (11), `pkg/network` (11), `pkg/efivars` (10), `pkg/ipmi` (9).

The important part for this fork is the capture series dated 2026-08-16. They are reimplementing the
SG2002 HDMI capture path in Go, without `libkvm`:

```
cvi: Load the media drivers rather than assume someone else did
cvi: Give VI the DMA working memory it never allocates for itself
cvi: Mux the MIPI RX pads before configuring the receiver
cvi: Correct the LT6911 receiver attributes
fix: Dispatch VI SDK ioctls via VI_IOCTL_SDK_CTRL
lt6911: Follow Sipeed's read sequence
lt6911: Add WriteBank for bring-up
lt6911: Start the CSI transmitter
lt6911: Drive the transmitter from MIPI_TX_CTRL after all
lt6911: Give the bridge an EDID when its storage is blank
fix: Stop LT6911 watchdog when opening registers
feat: Add the SG2002 HDMI capture pipeline
```

This is the only place where somebody has written down, in readable source, how the VI, MIPI and
LT6911 bring-up actually works on this board. The capture stack in this fork is a prebuilt
`libkvm.so` black box, so that value is high, whether or not any code is copied.

The rest is a BMC direction: Redfish, IPMI and EFI variables in-tree. Earlier work includes an EEPROM
handling rework with a signature file, generated API docs, and a UI redesigned after JetKVM and
PiKVM.

### mrjeeves/NanoKVM, 235 ahead, active 2026-08-15, releases 0.1.x

Two separate themes.

The first is a fleet product: a MyOwnMesh daemon, device claim and unclaim, and a CEC feature that
is a remote-support handoff. A user raises a hand, a technician is granted access, the grant is
time-boxed to three hours, and eviction frees the video lane. Not useful here.

The second is a USB gadget reliability campaign through August 2026, which is very useful here:

```
fix(usb): stop swapping the drive image under a live gadget
fix(usb): guard the drive swap inside installUsbDisk, not at one call site
fix(usb): stop rebuilding the gadget at startup
fix: repair USB gadget at app startup
fix: recover USB after media failures
fix: recover NanoKVM USB device role safely
fix NanoKVM USB health detection
fix unsafe USB gadget lifecycle
document USB role failure precisely
```

Note the conflict: this fork's own measurement says a media swap needs no UDC reset, because the LUN
is removable. His first two commits move the other way. Read why before you trust either.

The HDMI commits matter too:

```
fix(hdmi): stop cycling the receiver on every server restart
revert(hdmi): never power the receiver down
fix(hdmi): clear a stale disable flag once, so dark devices come back
```

And the USB network work names problems this fork will meet: NCM by default with an RNDIS to NCM
migration, a separate `S32usbdhcp`, "don't leave two DHCP servers on one link", "stop hiding a udhcpd
that refused to start", and "stop the uplink-sharing NAT eating the device's own DNS".

One more: "feat(update): deliver init scripts over the air", and "fix(update): make an OTA deliver
boot scripts on the release that adds them". That is the two-places init script problem, solved.

### woffko/Hardened_NanoKVM, 232 ahead, last push 2026-07-11, at RC9

A Rust rewrite of the backend. `kvm_system` responsibilities move into Rust services, including the
watchdog state and the Wi-Fi reconnect. Take the ideas, not the code:

- "Deinitialize libkvm on backend shutdown", which is the same ground as `fix/kvmv-thread-join`.
- "Add USB wakeup control and WebRTC keyframe recovery".
- "Sync NTP time during service startup".
- "Allow validated mass storage image uploads".
- Mobile work: TouchSync screen geometry scaling, pinch and pan, and absolute mouse X calibration in
  480 mode.

### mjc/NanoKVM, 20 topic branches, June 2026, every branch about 40 behind

The best organised fork of the set. Each branch is single-purpose and carries tests. Every branch is
about 40 commits behind upstream, so cherry-picks will conflict. Read and reimplement.

| Branch | Ahead | What it holds |
| --- | --- | --- |
| `mjc/fix-usb-virtual-media-host` | 44 | A real `server/service/usb` package: `gadget.go`, `mass_storage.go`, `rndis.go`, `fs.go`, `transaction.go`, all with tests. Transactional gadget state: preserve state on disable failure, skip mutations after a detach failure, do not expose unpersisted functions. |
| `mjc/cmd-argv-cleanups` | 16 | Centralize safe command runners in `utils`, remove the remaining Go shell-outs for file operations, reject unsafe mdns pid values. |
| `mjc/cleanup-usb-gadget-media` | 15 | "reopen stale HID gadget handles", plus `tests/usb-init-scripts-test.sh`, a fake configfs harness that tests the USB init scripts off-device. |
| `mjc/updater-test-coverage` | 14 | Updater hardening, including "assert updater restores backup on apply failure". |
| `mjc/auth-password-hardening-followups` | 9 | Password storage, account file hardening, Wi-Fi password as a private file, shared security helpers. |
| `mjc/fix-script-runner` | 7 | Command injection in `server/service/vm/script.go`, with regression coverage. |
| `mjc/revoke-password-change-tokens` | 5 | Revoke tokens after a password change, restore the account on sync failure. |
| `mjc/auth-device-control-security` | 5 | Validation contracts for autostart, hostname, jiggler, web title. |
| `mjc/auth-*` (7 more) | 1 to 2 | Security contracts and behaviour tests for core auth, client requests, network, storage, stream, PicoClaw, Tailscale CLI, and the updater and untar paths. |
| `mjc/fix-kvm-vision-teardown` | 1 | "stop kvm vision threads before teardown". |
| `mjc/investigate-purple-hdr` | 1 | "Force RGB-only EDID for LT6911 capture", plus `tools/nanokvm_update_edid/test_rgb_only_edid.py`. |
| `mjc/replace-ntpd-with-busybox-continuous` | 1 | Replaces the bundled ntpd with busybox ntpd in a new `S31ntpd`. |
| `mjc/rust-webrtc-sidecar` | 16 | A Rust WebRTC sidecar behind a Go backend switch, with a nix flake and a Beads issue tracker. |
| `mjc:nanokvm` | 11 | A nix dev shell, an image builder script, and a deploy script. |

---

## Tier 2: single-idea forks

- **Ihab-Zhaika/NanoKVM**, 16 ahead, last push 2025-12-13, stale. Virtual audio input and output
  called NanoInput and NanoOutput, with sound meters and mute controls, and **UAC2 kernel module
  support**. The `feat/usb-audio` branch here uses UAC1, so compare the two. The fork also carries CI
  workflows that produce flashable OS images and an `upgrade-from-pr.sh` test script.
- **Schattenwelt/NanoKVM**, branch `fork-live`, 32 ahead. The RBAC work of upstream PR #797, plus
  "support lt6911d" and changes under `tools/nanokvm_update_edid`.
- **JuergenLeber/NanoKVM**, 10 ahead. A polished version of the USB descriptor customization in
  upstream PR #758, plus "fix(server): embed RPATH so libkvm.so is found without LD_LIBRARY_PATH".
  This fork already patches the RPATH, so only the descriptor work is new.
- **lazerusrm/NanoKVM**, 3 ahead plus a `pr-700` branch. `now_fps` moved to `/tmp`, an absolute mouse
  coordinate scaling fix, and Alt-Shift layout switching in the virtual keyboard.
- **ZagidullinRuslanG/NanoKVM**, 3 ahead. USB device identity configuration and an
  `assemble_build.sh`.
- **BeaconCat** (5), **C2022H** `iso-de-keyboard` (4), **deasmi**, **DrEVILish**, **ethanperrine**,
  **Voranto** (1 each). These are the branches behind their own upstream pull requests, already
  covered in the other document.

---

## OneKVM organization

24 repositories, created 2026-07-31, still being pushed daily. It is a separate distribution for the
same hardware, built on Yocto and OpenEmbedded. The core, `onekvm-server`, is closed. Everything else
is being opened progressively, under MIT.

### onekvm-nanokvm-mmf

The video and crypto backend, deployed as one shared library at
`/usr/lib/onekvm/video-backends/nanokvm-mmf.so`. It captures from the LT6911, encodes H.264, H.265
and MJPEG in hardware, and reports the HDMI signal state with a built-in no-signal frame. Standalone
CMake, and it does not depend on MaixCDK.

Release builds compile the whole driver stack from official Sophgo sources rather than the Sipeed
prebuilt MMF, pinned to a single matched set dated 2026-06-30:

| Part | Repository | Revision |
| --- | --- | --- |
| Linux kernel | `sophgo/linux_5.10` | `sg200x-dev`, `767d3c5ab10b066d2d5c7c0bd1eab8a5340e923d` |
| Video kernel drivers | `sophgo/osdrv` | `sg200x-dev`, `aa542c41df94f7bc656cb740f6622a5dca7dc403` |
| Video userspace libraries | `sophgo/cvi_mpi` | `sg200x-dev` weekly, `75c181ee6e25baca9729a4a9b415f36180b54f93` |
| LT6911 support | `sophgo/SensorSupportList` | `sg200x-dev`, `f064b02ba8a82746f3e87a2c5bb3bd683ff95db0` |
| Video codec firmware | `sophgo/ramdisk` | `1ec8fcb63a358c17c369bac38eb42dc16f30a3bb` |

Their note says these revisions form a matched set, and that mixing revisions can stop the video
stack from starting.

This is the documented escape route from the `libkvm.so` black box. Keep the table even if nothing is
built from it.

### linux-onekvm-nbd

An out-of-tree NBD client for Linux 5.10, for browser-backed virtual media. It keeps the standard NBD
ioctl and netlink protocol but registers as `onekvm_nbd` and `/dev/onekvm-nbd0`, so it can coexist
with the upstream driver. Changes from upstream:

- An adaptive read-cache ceiling, `max_cache_kb`, 128 to 4096 KiB, exposed through sysfs and applied
  while the queue is frozen.
- `io_depth`, 1 to 1024, default 32.
- Adaptive read-ahead that starts conservatively and turns on after 16384 KiB of contiguous reads.
- **Receive workqueues are unbound and set to the lowest normal priority, so NBD network processing
  does not compete with latency-sensitive video and input work.**

That last point is the lesson for a single-core board, and it applies to more than NBD. The feature
itself, mounting an ISO straight from the browser with no local storage, is worth considering.

### linux-sg2002-cryptodma

Two kernel modules, `onekvm_sg2002_ghash_opt.ko` and `onekvm_crypto_offload.ko`, giving optimized
GHASH and AES-GCM offload through the CVITEK SPACC engine. On a board where HTTPS costs real CPU
against the capture path, this is worth measuring.

### The rest

- **onekvm-nanokvm-recovery**: an initramfs recovery hook, runtime marker commands, a recovery
  console, an input helper, and systemd integration, for USB recovery and factory reset. It is the
  same problem the slot tooling here solves, approached differently.
- **onekvm-system-lifecycle**: persistent configuration, extension state, package state,
  kernel-module updates, and storage preparation.
- **onekvm-plugin-manager** (Go): installs and validates extension IPKs, manages extension processes
  through an extension host, and enforces resource, device, port, credential and filesystem policy.
- **onekvm-extension-host** (Rust), and one repository per extension: tailscale, netbird, zerotier,
  rustdesk, vnc, rtsp, mcp, wake-on-lan, mouse-jiggler, gpio-atx, the OLED designer, and api-docs.
- **onekvm-nanokvm-machine**: the machine support layer, including a native I2C probe.

---

## Priority

### Take first, small and unblocking

1. **`tests/usb-init-scripts-test.sh` from `mjc/cleanup-usb-gadget-media`.** A fake configfs harness
   that lets `S03usbdev` and `S03usbhid` be tested without a device. Everything else in the USB area
   gets cheaper once this exists.
2. **`mjc/replace-ntpd-with-busybox-continuous`.** One commit. It drops the bundled ntpd for busybox
   ntpd, which saves memory and addresses upstream issue #793, where a failed time sync at boot stops
   Tailscale from connecting.
3. **`mjc/fix-script-runner`.** A command injection fix in `server/service/vm/script.go`, with
   regression coverage.

### Read and reimplement

4. **`mjc/fix-usb-virtual-media-host`.** The transactional `server/service/usb` package. It is the
   right shape for the gadget failures recorded here, where a failed bind wedges `g0`.
5. **The mrjeeves USB reliability series**, all nine commits. Resolve the media-swap disagreement
   against our own measurement before acting.
6. **EDID work.** `mjc/investigate-purple-hdr` forces an RGB-only EDID for LT6911 capture, and
   pi-bmc gives the bridge an EDID when its storage is blank. Both touch upstream issue #875, where
   application 2.4.3 and 2.5.0 capture zero frames on a Cube with LT6911UXC.
7. **`mjc/updater-test-coverage`**, specifically "restore backup on apply failure".

### Research, not code

8. **pi-bmc's `cvi` and `lt6911` commits** as documentation of the capture bring-up this fork
   currently treats as opaque.
9. **The OneKVM pinned driver set** as the route to building the video stack from Sophgo sources
   instead of the Sipeed prebuilt `libkvm.so`.
10. **The OneKVM NBD workqueue priority choice.** Moving non-critical kernel work to unbound,
    lowest-priority workqueues is a general lesson for this board, given that softirq alone already
    costs a measured 28 percent.
11. **`linux-sg2002-cryptodma`** for HTTPS cost.
12. **Ihab-Zhaika's UAC2 audio** against the UAC1 approach in `feat/usb-audio`.

### Skip

The mesh and CEC product in `mrjeeves`, the Rust rewrite in `woffko`, and the Redfish and IPMI
surface in `pi-bmc`, unless this fork decides to become a BMC.

---

## Should this fork rebase onto one of them?

No. Keep `upstream/main` as the base.

### The measurements

`fork/integration` is 267 commits and 323 changed files away from `upstream/main`. `main` alone is 84
commits. A rebase replays all of that onto a foreign base.

Every candidate is behind upstream, and this fork is one commit behind:

| Fork | Ahead of upstream | Behind upstream | Files changed | Files also changed here |
| --- | --- | --- | --- | --- |
| `woffko/Hardened_NanoKVM` | 232 | 41 | 300 (API cap) | 75 or more |
| `mrjeeves/NanoKVM` | 235 | 39 | 193 | 32 |
| `pi-bmc/nanokvm-app` | 201 | 79 | 300 (API cap) | 18, and see below |
| `mjc/fix-usb-virtual-media-host` | 44 | 39 | 22 | 8 |

The GitHub compare API returns at most 300 files, so the `woffko` and `pi-bmc` counts are floors.

### Why the overlap is the problem

The shared files are not scattered. They are the hot files, and they are the same ones in every case:

```
kvmapp/system/init.d/S01fs
kvmapp/system/init.d/S03usbdev
kvmapp/system/init.d/S03usbhid
kvmapp/system/init.d/S95nanokvm
server/common/kvm_vision.go
server/service/hid/hid.go
server/service/hid/status.go
server/service/storage/image.go
server/service/vm/virtual-device.go
server/config/*
```

Both sides already fixed the same defects in these files, independently and differently. That is
semantic conflict, not textual conflict. Git merges text. Git cannot say which of two USB gadget
rewrites is correct, and neither fork ships tests this one has reason to trust.

### Reasons per candidate

- **`pi-bmc/nanokvm-app`.** The overlap of 18 files looks small only because the tree was
  restructured into `api/` and `pkg/`. Most of the 323 paths here no longer exist there. Moving onto
  that base is not a rebase. It abandons this fork and restarts on theirs, discarding 267 commits.
- **`woffko/Hardened_NanoKVM`.** 75 or more shared files, and the backend is being rewritten in Rust.
  A rebase adopts Rust.
- **`mrjeeves/NanoKVM`.** About half of the 235 commits are a fleet mesh and CEC remote-support
  product this fork does not want. A rebase brings all of it.
- **`mjc/NanoKVM`.** The only fork with a compatible shape. Its branches hold 22 files or fewer each,
  so a cherry-pick is cheaper than a rebase in every case. The fork has had no push since June.

### Two things a rebase breaks

1. **The upstream pull request pipeline.** Nine pull requests are open at `sipeed/NanoKVM`. Every
   extraction branch is cut from `upstream/main`, so the pull request shows only its own change. If
   the base becomes a third-party fork, extraction stops working.
2. **The divergence measure.** `git rev-list --count upstream/main..main` only means something while
   `upstream/main` is the base.

Each of these forks is also one person's work, with no release promise. A rebase inherits their bus
factor.

### Do this instead

Pull their work toward this fork, rather than moving this fork onto theirs. Rebase their branch onto
`main` in a scratch worktree, read the conflicts, then reimplement in this fork's idiom:

```shell
git remote add mjc https://github.com/mjc/NanoKVM.git
git fetch mjc
git worktree add ../nkvm-mjc -b scratch/mjc-usb main
cd ../nkvm-mjc
git rebase --onto main upstream/main mjc/mjc/cleanup-usb-gadget-media
```

This is the flow that `CLAUDE.md` describes for extraction branches, run in the opposite direction.
It is cheap because each `mjc` branch is single-purpose.

The capture knowledge in `pi-bmc` and the video stack in `onekvm-nanokvm-mmf` are reading material,
and possibly a separately built artifact. Neither is a reason to move git history.

### The one exception

If this fork stops being a NanoKVM fork and becomes a BMC, exposing Redfish and IPMI, then
`pi-bmc/nanokvm-app` has already done a year of that work and is the right base. That is a restart on
their tree rather than a rebase of this one, and it costs all 267 commits. It is a product decision,
not a git decision.

---

## Second pass, 2026-08-28

Forks pushed since the survey above: `pi-bmc/nanokvm-app` (2026-08-25),
`mrjeeves/NanoKVM` (2026-08-24), `woffko/Hardened_NanoKVM` (2026-08-21, no
commits), and four that the survey did not have. One of the four matters.

### RobbyV2/NanoKVM, 234 ahead, 0 behind, active 2026-08-26

Not in the survey above, and the most useful fork found so far for the work this
one is doing. It rebases onto upstream rather than drifting, so it is 0 behind.

It carries a `server/service/presentation` package that treats the USB gadget as
a compiled profile: a capability table for the controller, an endpoint
accounting pass, a plan, and a reconcile against what is already linked. Beside
that there is USB passthrough over WebUSB, functionfs profiles, an EDID parser
with a recovery UI, an ethernet bridge, a wstunnel tunnel, media sources, and a
tier of kernel integration tests behind a build tag.

Two things were taken from it on 2026-08-28, and both are recorded in
`2026-08-17-upstream-adoption-backlog.md`:

- **The endpoint model.** Its `staticV1` carries `MaxInEndpoints: 6` and
  `InFIFOWords: {768, 512, 512, 384, 128, 128}`. Those are this board's dwc2
  parameters exactly, read back from `/sys/kernel/debug/usb/4340000.usb/fifo`.
  Its `MaxOutEndpoints: 5` is not: `hw_params` reports `num_dev_ep: 7` and the
  outbound direction needs no dedicated FIFO.
- **`req_number` for the audio gadget.** Their measurement of `u_audio`
  repeating a stale request buffer is the reason the value moved from the kernel
  default of 3 to 8.

Worth reading before `feat/usb-audio` is finished: their `service/media`
separates a quiet host from a parked stream, which an `arecord`-based capture
path has to answer differently but still has to answer. `SeatFIFOs` is worth
reading before any function that wants a wide isochronous IN endpoint is added,
because only FIFO 1 holds more than 512 words.

Their commit style and the shape of their comments are close enough to this
fork's that the two trees read as siblings. That is a reason to compare findings
carefully rather than to adopt them, as the outbound five shows.

### The rest

- **`mrjeeves/NanoKVM`**: one commit worth the name,
  "supervise the gadget instead of only repairing it at startup". Relevant to
  the recovery path, not yet read.
- **`eringiriri/ERINGI_JPN_NanoKVM`**: horizontal scroll on relative mode, and a
  fix for a composition guard that strands the JIS Zenkaku key pressed. Both
  small and additive.
- **`pi-bmc/nanokvm-app`**: twelve more `cvi` commits the survey missed by one
  day, including "Stop the drivers' error reporting from killing the board" and
  "Drain the encoder even when it has just refused a frame". Still a Go rewrite
  of the capture path, so still read and reimplement rather than adopt.
- `woffko/Hardened_NanoKVM` and `Schattenwelt/NanoKVM` have not moved.
