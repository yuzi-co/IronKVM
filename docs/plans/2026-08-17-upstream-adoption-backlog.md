# Upstream adoption backlog

Survey date: 2026-08-17. Two upstream repositories were reviewed for work this fork can adopt:

- `sipeed/NanoKVM`, the application source this fork tracks as `upstream`.
- `sipeed/LicheeRV-Nano-Build`, the Buildroot and kernel tree for the SG2002 board.

Third-party forks and the OneKVM organization are covered in the companion document,
`2026-08-17-fork-ecosystem-survey.md`.

Every claim below was checked against this checkout. Where the fork already carries the change, the
file and line are given so the finding can be re-verified.

## Fork position

`git fetch upstream` on 2026-08-17 gave:

```
ahead:  84
behind: 1
merge-base: 9415afb3591ee7a83abc947601ba1f2aeb6c8f10
```

The single missing commit is `d382f062`, "fix: report and retry WebRTC connection failures (#866)".
`sipeed/NanoKVM` has one branch, `main`, so no branch work is hidden there.

Nine of the open pull requests on `sipeed/NanoKVM` are this fork's own extractions (`#846`, `#847`,
`#848`, `#849`, `#850`, `#851`, `#852`, `#871`, `#873`). The third-party pool is 20 pull requests.

## Status, 2026-08-19

The survey above is a record of 2026-08-17 and is left as written. This section
records what changed after it.

The fork is no longer behind. `fork/integration` and `main` were rebased onto
`d382f062` on 2026-08-19, which was step 1 of the suggested order. All nine of
this fork's own extraction branches were rebased onto the same commit and
force-pushed, and each still carries the identical patch it carried before.

Three items are done:

| Item | State |
| --- | --- |
| PR #813, gate the data disk until the filesystem is ready | Done. `S01fs` now provisions through `provision_disk0`, covered by `tools/abslots/device/test-s01fs-disk0.sh`. |
| PR #749, request DHCP option 121 | Done, for `S30eth` and `S30wifi` both, covered by `tools/network/test-dhcp-options.sh`. |
| PR #764, set the `/data` partition type and label | Done. The A/B path already did it; the stock path now does it too. |
| PR #759, `GOMEMLIMIT` for the server | The init-script half is done, covered by `tools/service/test-server-memlimit.sh`. The NetBird VPN is not taken. |

Three corrections to the text above, from checking it against the tree rather
than re-reading it:

- Item 1 says the #813 defect is "in this tree today". True as code, but it
  cannot run on an IronKVM card: `may_autopartition` refuses it whenever
  `/etc/nanokvm-slots.conf` is readable, and the image manifest creates
  `/etc/kvm.disk0` as well. The population it reached is a stock-layout board
  that installs this firmware over the air, because the update package carries
  `S01fs` and carries neither of those files.
- Item 8 understates what is already done. `S95nanokvm` links `state` to tmpfs
  as well as `now_fps`. What remains of #751 is `wifi_state`, `width` and
  `height`.
- Item 7 describes a different failure here. The server generates and owns
  `/etc/kvm/.picoclaw_internal_token` and the bridge script reads it, so the
  cache cannot be overtaken by PicoClaw. It can go stale only if `/etc/kvm` is
  rebuilt underneath the running server.

## Status, 2026-08-19, second round

Four more items are taken. None of them was taken as written, and the reasons
are worth keeping.

| Item | State |
| --- | --- |
| PR #675, HID mode without a reboot | Done, by a different route. `applyHidMode` runs `stop_start` on the installed script and reads the mode back, the way the disk, network, console and speaker toggles already work. The pull request rebuilds the gadget in Go and calls `log.Fatalf` on the way, which would stop the server. |
| PR #877, mouse coordinates on direct H.264 | Half taken. The canvas size fix is correct and is in. The letterbox inference is not: it treats black columns at both edges as padding, which a console or a dark desktop has, and it needs the bars to decode to exactly zero, which H.264 limited-range black need not. |
| PR #814, `tests/usb-init-scripts-test.sh` | Not portable. It drives an `S03usb-common` this fork does not have through nine environment variables these scripts do not read. Its four report descriptors were taken, as a second transcription to check ours against, and the harness is the one `test-acm-console.sh` already used. |
| `a3ccb527`, UDC bind retry | Taken with a bound. That loop is `while true` in a script `rcS` waits on, so a controller that never appears would stop the boot before the network and the server start. |

The HID mode switch needed two changes in the gadget scripts before the server
change could work at all, and neither is obvious from the pull request:

- `f_hid` copies `subclass`, `protocol`, `report_length` and the report
  descriptor into the instance when the function is linked into a
  configuration, and returns `EBUSY` for every write to them while the link
  exists. A rebuild that does not unlink first rebinds carrying the descriptors
  of the mode it left, while `bcdDevice` says otherwise.
- `stop` writes an empty UDC and leaves `configs/c.1` alone, so hid-only mode
  arrives with whatever normal mode linked. Console, disk and network together
  are six endpoints of nine before HID asks for three, so left in place the
  gadget refuses to bind and every `/dev/hidg*` disappears.

`S03usbdev` also now writes `bcdUSB` and `bcdDevice` rather than leaving them at
the kernel's default. The default was the wanted value by coincidence: the
kernel derives `bcdDevice` from its own version, and 5.10 gives `0x0510`.

---

Item 3 was also confirmed on hardware rather than assumed: the board's
`/usr/share/udhcpc/default.script:73` does test `$staticroutes`, so requesting
the option is enough and no handler was needed.

---

## Status, 2026-08-19, PR #800

Closed with no production change. The mechanism the pull request fixes does not
exist here, and the board says so rather than the reading of it.

Upstream caches a token that PicoClaw regenerates. Here the direction is
reversed: the server generates `/etc/kvm/.picoclaw_internal_token`, owns it, and
the bridge script only reads it. Nothing can overtake the cache. `/etc/kvm` is
mounted at `S02identity`, long before `S95nanokvm` starts the server, and no code
in this tree rebuilds it while the server runs.

Measured on the device against `/api/picoclaw/screenshot`, which is one of the
five loopback paths:

| Request | Status |
| --- | --- |
| The header holds the token from the file | 200 |
| The header holds a different token | 401 |
| No header | 401 |

The cached token and the file agree on a running board. The fork also answers
the pull request's own mechanism separately: `ensurePicoclawPicoToken` reconciles
PicoClaw's `.security.yml` channel token at every runtime start.

What the item did find is a second copy of the token, and no test on it.
`defaultPicoclawMCPServer` writes the token into PicoClaw's `config.json` as a
request header, and that copy is what PicoClaw sends. `setMCPServer` overwrites
it when it differs, which is correct, but nothing held that behaviour in place.
The failure it prevents is silent: a stale header returns 401 to every MCP call,
the agent stops being able to touch the machine, and no page reports an error.
`runtime_defaults_test.go` now pins it. Two mutations confirm the tests bite:
making the field loop create-only, and dropping `e.changed`.

---

## Priority 0: a live bug here, or a small clean change

### 1. PR #813, gate the default data disk until the filesystem is ready

Author `@mjc`. `+1253/-61`, 9 files, `mergeable=clean`. Includes tests.

`kvmapp/system/init.d/S01fs:134` creates `/etc/kvm.disk0`. Line 139 then runs
`(mkfs.exfat /dev/mmcblk0p3) &` in the background. The marker therefore exists before the format
completes. If the board loses power during the format, the next boot reads the marker, treats a
half-formatted partition as usable, and exports it over USB mass storage.

The pull request tracks an interrupted format with `/etc/kvm.disk0.formatting` and retries it on the
next boot.

This is the highest-value item on the list, because the defect is in this tree today.

### 2. PR #675, change HID mode without a reboot

Author `@scpcom`. `+1019/-46`, 11 files, `mergeable=dirty`, last updated 2026-02-16.

`server/service/hid/status.go:120` still logs "reboot system..." and calls `exec.Command("reboot")`.
A restart on this board costs 135 seconds under the graceful-stop init script.

The branch conflicts and is old. Read it for the approach, then implement it here. Do not
cherry-pick.

### 3. PR #749, request DHCP option 121

Author `@phaidros7`. `+3/-3`, 1 file, `mergeable=clean`.

`kvmapp/system/init.d/S30eth:60` and `:68` call `udhcpc` without `-O 121`, so the server never
offers classless static routes. The existing `default.script` already handles them per RFC 3442.

### 4. PR #764, set the `/data` partition type and label

Author `@jjmaestro`. `+4/-3`, 1 file, `mergeable=clean`.

Creates `mmcblk0p3` with the parted NTFS type, so the MBR id is `0x07` instead of Linux `0x83`, and
gives the exfat filesystem a label. `S01fs` sets neither today.

Check this against the A/B slot caveat recorded at `kvmapp/system/init.d/S01fs:46-50` before you
apply it.

---

## Priority 1: worth the work

### 5. PR #858, safer OLED sleep and monotonic timers

Author `@SiYue-ZO`. `+759/-577`, 12 files, draft, `mergeable=clean`.

No monotonic clock exists anywhere in `support/sg2002/kvm_system/main/lib/oled_ctrl/` or
`.../oled_ui/`. An NTP step at boot can therefore blank or wake the panel for no reason. The pull
request moves inactivity, carousel, event-wake, Wi-Fi and button timers onto monotonic time, and
separates the OLED power state from the UI subpage.

This sits next to the existing `fix/oled-sleep-never` branch.

### 6. PR #759, take the init-script half only

Author `@AndrewMoryakov`. `+1651/-16`, 23 files, `mergeable=dirty`.

Do not take the NetBird VPN. Take the `S95nanokvm` changes: `GOMEMLIMIT` and OOM protection for
`NanoKVM-Server` itself.

This fork sets `GOMEMLIMIT` only for tailscaled, at `kvmapp/system/init.d/S98tailscaled:46-53`.
Memory pressure is the worst failure mode on this board, so the server deserves the same treatment.

### 7. PR #800, read the PicoClaw token dynamically

Author `@BeaconCat`. `+52/-0`, 1 file, `mergeable=clean`.

The pull request reads the token from `~/.picoclaw/.security.yml` on every call, because PicoClaw
regenerates it on restart.

This fork uses a different file and caches the value: `server/config/picoclaw_internal.go:24` returns
`picoclawInternalToken.value` when it is not empty, and the file is
`/etc/kvm/.picoclaw_internal_token`. The mechanism differs, but the failure shape is the same.
Confirm whether the cache goes stale here before you write any code.

### 8. PR #751, move more runtime state to tmpfs

Author `@winstar0070`. `+33/-32`, 8 files, `mergeable=dirty`.

The `perf/now-fps-off-sd-card` branch covers `now_fps`. This pull request also covers `state`,
`wifi_state`, `width` and `height`, and it changes the C++ side in `kvm_vision.cpp` and
`system_state.cpp`, which this fork has not touched.

Diff his file list against the fork branch and take the remainder.

---

## Priority 2: features, larger

| PR   | Author          | What                                                        | Size                |
| ---- | --------------- | ----------------------------------------------------------- | ------------------- |
| #797 | `@Schattenwelt` | Multi-user authentication with admin, operator, viewer roles | +1807/-277, 41 files |
| #809 | `@YipKo`        | Wi-Fi 802.1X, static IP, DNS, network detail panel            | +3706/-220, 40 files |
| #867 | `@SiYue-ZO`     | Toolbar for mobile and touch devices                          | +1832/-558, 73 files |
| #825 | `@ethanperrine` | PicoClaw provider and auth UI, model testing                  | +2824/-206, 23 files |
| #864 | `@jordimra`     | Spanish (Spain ISO) paste layout                              | +148/-7, additive    |
| #682 | `@imguoguo`     | Screen recorder                                               | +147/-2, 5 files     |

`server/service/hid/paste.go` carries only `de` (line 40) and `fr` (line 94), so #864 is purely
additive.

---

## Priority 3: already carried here, or skip

- **#746**, USB serial number from the SoC UID. Already here.
  `kvmapp/system/init.d/S03usbdev:271-280` reads `/sys/class/cvi-base/base_uid`. The
  `security/usb-gadget-identity` branch goes further.
- **#814**, host USB HID boot, mode and recovery. Mostly here.
  `kvmapp/system/init.d/S03usbhid:41-70` already sets the boot subclass and protocol, and records why
  the absolute pointer is never a boot device. `fix/hid-gadget-rebuild` and
  `fix/hid-endpoint-reporting` cover the server side. **Take his `tests/usb-init-scripts-test.sh`.**
- **#741**, do not expose the eMMC as mass storage. Covered by `security/usb-mass-storage-default`.
- **#819** and **#775**, USB CDC ACM console. Covered by `fix/usb-acm-console` and `/boot/usb.acm`.
  #819 is a draft proof of concept.
- **#758**, USB descriptor customization. Partly here through `/boot/usb.vid` and `/boot/usb.pid`.
  The pull request adds a settings UI and eight vendor presets. It is `dirty`.
- **#700**, Tailscale auto-update and Ethernet settings. `dirty` since January, and it overlaps #809.
- **#868**, TypeScript 6. Draft. Wait for it.

---

## Open upstream issues this fork already answers

These are not code to adopt. They show which upstream problems the fork branches already solve, and
which users feel them.

| Issue | Fork work that addresses it                          |
| ----- | ---------------------------------------------------- |
| #806  | Safari caches the Web UI: `perf/web-cache-headers`   |
| #804  | WebRTC to MJPEG wedges the server: ION exhaustion work |
| #634  | Stuck `cvitask_isp_*` threads: `fix/vi-init-race`, `fix/capture-gate` |
| #149  | Tailscale auth lost on app update (26 comments): the identity problem |
| #219  | Remote firmware update (26 comments): the slot tooling |
| #540  | PCIe cannot install Windows from ISO: see the mass-storage patch below |

**#875**, filed 2026-08-15: application 2.4.3 and 2.5.0 capture zero frames on NanoKVM Cube with
LT6911UXC, VIFPS reads 0, and 2.3.6 works. This is current and severe. Check that this device is not
affected.

Large unbuilt requests, for direction only: #79 Zerotier (48 comments, 31 reactions), #295 MFA (13
comments, 8 reactions), #71 OpenVPN, #565 Clear CMOS, #575 lock the keyboard in fullscreen.

---

## From `sipeed/LicheeRV-Nano-Build`

That repository holds a branch named `NanoKVM`, 26 ahead and 18 behind its `main`, pushed as recently
as 2026-07-07. It is where Sipeed keeps the KVM image recipe. Watch that branch, not only `main`.

### Virtual DVD in the mass-storage gadget

Commit `1f3fa10d` on the `NanoKVM` branch, 2025-09-13, "support windows iso on usb mass storage".
It patches `linux_5.10/drivers/usb/gadget/function/f_mass_storage.c` and adds the MMC command set:
`GET_CONFIGURATION`, `READ_HEADER`, `READ_TRACK_INFORMATION`, `READ_DISC_STRUCTURE`, the Mechanical
Status mode page, and a `cd_as_dvd` LUN flag that reports `MMC_PROFILE_DVD_ROM` for large images.

`kvmapp/system/init.d/S03usbdev:404-408` sets `lun.0/cdrom` to `0`, and only touches it on the
read-only path. Stock `f_mass_storage` advertises CD-ROM poorly, which is why a Windows installer ISO
does not always boot.

This is kernel source, not a script. This fork does not build a kernel, so record it and know that a
Sipeed image built after September 2025 already carries it. It is the answer to upstream issue #540.

### PR #836, "More RAM"

Sets `BOOTLOGO_SIZE = 0` in `memmap.py`, guards the device tree with
`#if (CVIMMAP_FRAMEBUFFER_SIZE > 0)`, and forces `ENABLE_BOOTLOGO=0` in `envsetup_soc.sh`. The author
reports `free -m` total of 166 MB.

`tools/fbmem/` already does this by repacking the FDT in `boot.sd`, and `tools/fbmem/README.md:58`
already states that `memmap.py` is the better place. The device tree guard and the `ENABLE_BOOTLOGO`
hook are the two hunks this fork does not have. Take them if this fork ever builds an image.

### ION size

Commit `71316159` on `main` raises ION from 75 MB to 105 MB. Commit `bf7570c8` on the `NanoKVM`
branch puts it back to 75 MB for the KVM board. The 75 MB carveout is deliberate.

### PR #835, minimal NanoKVM image target

Adds a board `sg2002_licheervnano_sd_minimal` with a Buildroot allowlist. It removes Python 3 and its
module set, FFmpeg, OpenCV, ALSA utilities, mpg123, Tcl, Expect, tmux, htop, nano, vim, GDB, strace,
the benchmark and memory-test tools, the IPMI, serial, LCD and TPU demos, and the extra archivers.

This fork ships no image, so the pull request is not adoptable. Keep the allowlist as a reference for
what a smaller rootfs would drop.

### USB gadget details worth copying

- **UDC bind retry.** Commit `a3ccb527` replaces the single write with a loop that retries every 0.4
  seconds until `cat UDC` returns a non-empty value. `kvmapp/system/init.d/S03usbdev:476` still does
  the single write. A bounded retry is cheap.
- **Empty-UDC detection.** Commit `5602fd9d` adds `adbd_monitor.sh`, which polls
  `/sys/kernel/config/usb_gadget/g0/UDC` and restarts the gadget when it reads empty. Their recovery
  will not work here, because a rebind fails and `rmdir acm.GS0` blocks on the ttyGS0 getty. The
  detection half is still a free health signal for `S99vidiag` or `/api/hid/status`.

### Ideas, not code

- Commit `e5a5c4df` adds `BR2_PACKAGE_OPENIPMI=y` and bumps OpenIPMI to 2.0.36 for an IPMI server
  simulator. Sipeed is moving the KVM toward answering as a BMC. That would let normal datacenter
  tooling drive ATX control.
- Issue #833, "VB pool not released after MMF program exits", reports the ION exhaustion failure
  against stock Sipeed firmware, with no answer from Sipeed. It confirms the defect is in the vendor
  middleware. It names `cvidaemon` as the process that claims the pool at boot, and asks whether
  reloading `cvi_vb.ko` frees it. `tools/README.md:784` already records that `rmmod soph_vpss` answers
  "Resource temporarily unavailable", so that door looks shut, but `cvi_vb` was not tried.

### Small items

- Commit `198086d1` fixes an `int` to `uint` overflow in `u-boot-2021.10/cmd/i2c.c`. It matters only
  if this fork builds U-Boot.
- Commit `017c2724`, "Add USB audio support", enables `CONFIG_SND_USB_AUDIO` so the board can drive a
  USB sound card as a host. It is not the UAC1 gadget path. Do not confuse the two.
- Issue #116 reports that `ipset` is not functional. Check that before promising anything advanced in
  the Tailscale extension.
- The `kvm/` directory on the `NanoKVM` branch holds prebuilt `frp 0.59.0` and `tailscale 1.80.2`
  riscv64 binaries, and `sensor_cfg.ini` variants for LT, OA, SC035, alpha and beta. PR #827 adds a
  LONTIUM LT6911 configuration.

---

## Suggested order

1. Rebase onto `d382f062`.
2. PR #813, the data-disk gating. The defect is in this tree.
3. PR #675, HID mode without a reboot. It returns 135 seconds per mode change.
4. PR #749 and PR #764. Seven lines in total, and both merge clean.
5. PR #759, `GOMEMLIMIT` and OOM protection for the server.
6. PR #858, monotonic OLED timers.
