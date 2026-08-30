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

---

## Status, 2026-08-25, upstream `71ab9127`

`sipeed/NanoKVM` landed one commit since the 2026-08-19 rebase: `71ab9127`,
"feat: add secure multi-user support (#876)", 59 files, +2702/-510. It rewrites
the authentication subsystem this fork had five branches sitting on: a new
`server/authn` package with users, roles and token versions, a rewritten
`middleware/jwt.go`, a new `middleware/session.go`, `service/auth/account.go`
deleted, and every router re-plumbed for role checks.

`main` and `fork/integration` are rebased onto it. The nine extraction branches
are rebased onto the same commit.

### What upstream took, and what it left

Three of this fork's open pull requests describe faults `#876` also fixed. It
did not merge them; it wrote its own. The three are not equal, so the answers
differ:

| Fault | Upstream now | This fork |
| --- | --- | --- |
| Current password required to change the password | Required every time, checked against the account store | Dropped. `#848` closed. |
| JWT algorithm confusion | Method check plus an explicit method list in `ParseJWT` | Implementation dropped. `alg=none`, which upstream does not test, kept as a test. |
| Cross-site requests | `CheckWebSocketOrigin` on the five upgraders | Kept. Upstream guards the upgrades; the API around them still takes any origin. |

The origin rule is the one worth stating plainly, because "upstream has an
origin check now" reads like the work is done. Upstream's check covers websocket
upgrades only, compares the port as well as the host, and does not know about
`authentication: disable`. This fork applies the rule to every authenticated
request, compares the hostname alone because the device serves the same UI over
http and https, and stands down in the mode that exists for frontend
development. The two implementations disagree, so the fork keeps one of them and
removes the other rather than running both.

### What had to be re-expressed rather than replayed

- **API keys.** A key now carries the account it was issued to and that
  account's role. Before `#876` the device had one account and a key could carry
  every authority it had. With roles, a key that spoke for nobody in particular
  would let any operator mint an administrator, so `apikey.Verify` returns the
  key, the middleware resolves the account, and a disabled or deleted account
  takes its keys with it. Keys issued before this are adopted by the system
  account.
- **The default account hash.** The fork derived it once instead of per login
  attempt. Upstream stores the hash in the account file, which removes most of
  the cost, but `defaultDatabase()` still runs bcrypt on every read while the
  file is absent, and `Get` reads on every request that carries a session. The
  fix is now `sync.OnceValues` in `authn/store.go`.
- **The secret key.** Upstream deleted the rotation on logout, because a session
  is revoked through the account's token version now. It kept the clock-derived
  fallback in `generateRandomSecretKey`, which is the half that mattered: that
  fallback signs every session and the PicoClaw internal token. The fork's
  generator replaces it and the rotation is gone.
- **The web session cookie.** `web/src/lib/cookie.ts` is deleted with upstream.
  The server sets the cookie itself, with `SameSite=Strict`, `Secure` on https
  and `HttpOnly`, which is stronger than anything the page could do. The two
  places that used the helper changed shape: the login page confirms the session
  by asking for the account rather than reading the cookie back, and the TLS
  switch calls logout instead of deleting the cookie from JavaScript.

### Pull requests

| PR | State |
| --- | --- |
| #846 novision build tag | Rebased, mergeable. |
| #847 origin check | Rebased and re-scoped to the API half. Retitled. |
| #848 current password | Closed, superseded by `#876`. |
| #849 request paths | Rebased, mergeable. |
| #850 JWT | Rebased and re-scoped to the signing key. Retitled. |
| #851 mass storage | Rebased, mergeable. |
| #852 frame rate counter | Rebased, mergeable. |
| #871 MCP protocol | Rebased, mergeable. |
| #873 zram | Rebased. Routes moved to the admin group. |

No maintainer has commented on any of them. `#873` is the only one with a
conversation, and it is with another user rather than a maintainer.

### Item 8, PR #751, runtime state on tmpfs

Taken, and it found something. `now_fps` and `state` were already redirected
here by symlink. The `state` link never held: libkvm publishes that file by
renaming a temporary one over the path, and a rename replaces a symlink instead
of following it, so the first publish after every start put a regular file back
on the boot medium. `prepare_runtime_state` re-made the link at the next
restart, which hid it.

Both publishers now write to `/tmp/kvm/state` directly. The link stays for the
readers that know the old path, and the writes move as soon as a library built
from these sources ships: `server/dl_lib/libkvm.so` still holds the old path.

`wifi_state` is redirected, which is all it needed: kvm_system writes it with a
shell redirect and a redirect follows a symlink.

`width` and `height` stay on the card. libkvm reads them back at capture init to
restore the resolution the board last saw, so a tmpfs copy seeded with 0 changes
what the sensor is configured with on the first frame after a boot. That is a
device behaviour change and it needs hardware to answer.

### What is not done

- `libkvm.so` and `kvm_system` are not rebuilt, so the `state` move is in the
  sources and not yet in the binaries this repository ships.
- The extraction branches that have no pull request open, `security/api-key-auth`
  above all, are not rebased onto `71ab9127`. They carry the pre-`#876` shape of
  the auth code and have to be re-expressed the way `main` was.

---

## Status, 2026-08-25, upstream `2ba45a21`

Upstream moved again the day after `#876`. Three commits, all from the same
author, all pushed to `main` with no pull request and no review:

| Commit      | Subject                                                |
| ----------- | ------------------------------------------------------ |
| `527caeff`  | `fix(web): correct mouse coordinates for direct h264 frames` |
| `9e6d2808`  | `feat(web): add configurable mouse input regions`      |
| `2ba45a21`  | `feat(web): enhance configurable input regions`        |

Together they add one feature: the operator marks which part of the captured
frame is the host's real picture, and the absolute pointer maps into that
rectangle instead of the whole frame. The board scales every source into a
fixed 1920x1080 output, so a 4:3 or 5:4 host arrives with black bars down both
sides, and the pointer drifts further from the cursor the closer it gets to an
edge.

The feature has three modes. `off` uses the whole frame. `auto` samples the
rendered pixels and looks for black borders. `manual` stores a rectangle the
operator drew. `off` is the default on both sides of the wire, so nothing
changes for a board that never opens the menu.

Three new routes, all on the token group rather than the admin group, which
matches `/vm/gpio` and `/vm/screen` next to them. The setting is stored as
`/etc/kvm/input-region.json`.

### What it cost to take

`main` rebased onto `2ba45a21` with no conflict at all.

`fork/integration` had four, and three of them were import lists. The fourth is
the only one worth recording. `9e6d2808` introduces a `ScreenViewport`
component and moves every stream into it, which is also where the video scale
now lives, so upstream deleted the per-screen `getVideoScale` effect from
`mjpeg.tsx` and `h264-webrtc.tsx`. The fork's audio commit had added an
`<audio>` element and a second copy of that same effect to the WebRTC screen.
The effect is gone with upstream's, and the `<audio>` element now sits beside
`<ScreenViewport>` rather than inside it: the component clones a single child
and rewrites its style, so a second child there would be dropped.

### One fork commit is dropped

`89ee7021`, "Tell the mouse how large a direct H.264 frame really is", took the
first half of upstream PR `#877`: the direct H.264 worker owns the canvas
through `transferControlToOffscreen`, so the element this thread can see keeps
its default 300x150 and the pointer mapped against a 2:1 picture. `527caeff` is
that pull request, landed, with the same mechanism down to the `frame-size`
message and the `dataset` attributes. Keeping the fork's copy would have been a
second implementation of one fix, so it is skipped.

That commit's message argued at length against the second half of `#877`, the
black-border sampling, on the grounds that a console or a BIOS screen has exact
black at both edges and would be read as padding. Upstream's answer is the mode
switch: the sampling runs only when the operator selects `auto`, and the
default is `off`. The objection stands as a description of what `auto` will do
on a dark host, and it is now the operator's choice to make rather than a
behaviour every board gets.

### What it confirms

`width` and `height` stay on the SD card, and there is now a second reason.
`GetInputResolution` reads `/kvmapp/kvm/width` and `/kvmapp/kvm/height` to offer
the original-resolution presets. Had item 8 moved those two files to tmpfs, the
new endpoint would have failed on every boot until the first resolution change,
in a feature that has no other way to learn the source size.

`/etc/kvm/input-region.json` needs no work to survive a slot switch. `S02identity`
binds `/data/identity` over the whole of `/etc/kvm`, so anything the server
writes there is already on the data partition.

### Two things in the new code, not acted on

- `removeInputRegion` in `server/service/vm/input_region.go` has no caller.
- Neither `Regions` nor `Resolutions` has a length limit, and the server has no
  request body cap outside the offline updater. An authenticated user can post
  an arbitrarily large array, which is decoded into memory on a board that
  wedges rather than OOM-kills below about 30 MB free. This is not specific to
  the new routes: every JSON endpoint here has the same exposure, so the fix is
  one body cap in the middleware rather than a length check per handler. It is
  a fork item, not a defect to report upstream on its own.

### Pull requests

All eight open pull requests still report `MERGEABLE` against `2ba45a21`, and a
local `git merge-tree` against the new upstream agrees for every one of them.
None needed a rebase. Only `#873` touches the same files at all: it appends its
routes and its proto types below upstream's, and its web changes are in the
settings panel rather than in the screen.

### Still not done

Both items from the `71ab9127` section stand, against the newer base:

- `libkvm.so` and `kvm_system` are not rebuilt, so the `state` move is in the
  sources and not in the binaries this repository ships.
- The extraction branches with no pull request open, `security/api-key-auth`
  above all, are not rebased. They still carry the pre-`#876` shape of the auth
  code.

---

## Status, 2026-08-28

A second survey of the same two sources, plus the fork pool. `sipeed/NanoKVM`
has not moved: `upstream/main` is still `2ba45a21` and `fork/integration` is 366
ahead, 0 behind. Everything new is in the pull request pool and in third-party
forks.

### Pull requests this document had not triaged

Thirty-three are open. Five were not covered above, four of them opened after
the 2026-08-25 pass.

| PR | Author | What | Verdict |
| --- | --- | --- | --- |
| #888 | `@watermeko` | Share H.264 capture between direct and WebRTC | Idea taken, code not |
| #880 | `@SiYue-ZO` | Harden multi-user sessions and proxy authentication | Two parts taken |
| #881 | `@SiYue-ZO` | `make server` staging target | Overlaps `tools/deploy`, and has no rollback |
| #879 | `@rajr0` | Reboot button in the toolbar | 27 files, 24 of them locale stubs |
| #795 | `@bilibilifmk` | Login persistence and desktop interaction | Stores the password in `localStorage` |

**#888** is right about the defect and wrong about the fix. Direct mode and
WebRTC mode each ran their own ticker and each called `ReadH264`. One encoder
sits behind that call and hands each caller whatever frame is ready, so two
viewers on different modes divided the stream instead of each receiving it.

Its own hand-off blocks the capture loop on a four-deep channel, with a WebRTC
track write behind that, so one slow peer stalls the direct viewers as well.
That is what `fix/stream-stalled-viewer` and `FrameSlot` exist to prevent. It
also commits `libopencv_video.so.409` to `dl_lib` to satisfy the cross-linker.
Nothing in that library is used: measured against the binary the pull request
adds, `libkvm.so` has 320 undefined symbols, the library exports 694 functions,
and the two sets do not meet. `patchelf --remove-needed` is what this fork
already does, and upstream's own committed `libkvm.so` carries the entry while
this fork's does not. Its frame rate figure also rests on an empty interceptor
registry, which removes the NACK responder and TWCC.

Taken as `perf/shared-h264-source`, on `FrameSlot`.

**#880** carries two things this fork wanted. Its `Decrypt` reimplementation
fixes a fault that is in this tree: `utils.Decrypt` panics on a malformed cipher
text, `POST /api/auth/login` runs it on the request body before any credential is
checked, and almost any garbage of the right length reaches it, because the last
byte of a wrong decryption is above the block size fifteen times in sixteen.
`gin.Recovery()` turns that into a 500 rather than a stopped server, and there
are six call sites. Its cookie change is the other: the `Secure` attribute has to
come from the request rather than from `conf.Proto`.

Taken as `fix/decrypt-panic` and `fix/secure-cookie-scheme`.

### The endpoint budget was wrong, and the controller says so

The flat budget of nine in `S03usbdev` and `endpoints.go` was fitted to
configurations tried by hand. It is the two directions added together.

```
# cat /sys/kernel/debug/usb/4340000.usb/hw_params
num_dev_ep       : 7
total_fifo_size  : 3072
# cat /sys/kernel/debug/usb/4340000.usb/fifo
Periodic TXFIFOs:
    DPTXFIFO 1: Size 768   DPTXFIFO 2: Size 512   DPTXFIFO 3: Size 512
    DPTXFIFO 4: Size 384   DPTXFIFO 5: Size 128   DPTXFIFO 6: Size 128
```

Seven endpoint pairs beside ep0 and six dedicated transmit FIFOs. dwc2 runs in
dedicated FIFO mode and refuses an IN endpoint it cannot give a FIFO of its own
to, so the ceiling is six inbound and seven outbound. A seventh FIFO does not
fit either: those depths plus the RX and non-periodic FIFOs are 3000 of the 3072
words the part has. The shipped gadget sits exactly on the inbound ceiling, with
all six FIFOs seated and `ep7in` idle.

One combination changes: the console and the network together, which is seven
inbound of six. It was built on the board to see what that does.

```
dwc2 4340000.usb: new device is high-speed
dwc2 4340000.usb: new address 6
dwc2 4340000.usb: dwc2_hsotg_ep_enable: No suitable fifo found
dwc2 4340000.usb: dwc2_hsotg_ep_enable: No suitable fifo found
```

The kernel ring is the only place it appears. At the same moment the UDC
reported `configured`, `/dev/hidg0` through `/dev/hidg2` and `/dev/ttyGS0` were
all present, and `usb0` was up, so every check anyone would think to make says
the gadget is fine. Which endpoint loses is whichever is enabled seventh rather
than a fixed one: in that run the HID endpoints kept their places and a bulk
pair did not.

**Do not read a successful bind, a `configured` UDC, or the presence of
`/dev/hidg*` as proof that a gadget configuration works.** The 2026-08-05 round
recorded console with network as binding, and it does. Only the bind was
checked.

`RobbyV2/NanoKVM` models the same constraint, and its inbound six and its FIFO
depths match this hardware exactly. Its outbound five does not: the controller
reports seven.

Taken as `fix/usb-endpoint-direction`.

### The audio gadget was writing attributes nothing accepted

`req_number` is how many isochronous requests `u_audio` keeps in flight. A
service interval the host does not fill still completes its request, and
`u_audio` copies that request's buffer into the ALSA ring again, so the capture
repeats the audio from `req_number` milliseconds earlier. `RobbyV2/NanoKVM`
measured one millisecond in 5.6 as a byte-identical repeat at 4 requests against
one in 46 at 8. Not re-measured here. The cost is 1536 bytes of buffer rather
than 576.

Setting it exposed an older fault in the same block. configfs refuses these
attributes while a config holds the function:

```
# echo 8 > .../functions/uac1.usb0/req_number
sh: write error: Resource busy
```

Neither the function directory nor its config entry goes away on a
`stop_start`, so every rebuild after the first writes to a function the config
is still holding. Both writes are redirected to `/dev/null`, so they fail
silently. Deployed and measured: a `stop_start` with the block writing 8 left
`req_number` at 3 and said nothing. `p_chmask`, which carries the whole "the
host gets a speaker and not a microphone" decision, had the same problem.

The config entry now comes out first, then both attributes, then the link.
Re-measured after that change: 3 to 8, with the gadget and all three HID
functions intact.

Taken as `fix/uac-request-number`.

### What was taken

| Branch | What |
| --- | --- |
| `fix/decrypt-panic` | Read the salted AES format here, padding checked, instead of panicking on it |
| `fix/secure-cookie-scheme` | `Secure` from the request; delete the cookie while HTTPS is still on |
| `perf/shared-h264-source` | One capture loop for both H.264 paths, handed off through `FrameSlot` |
| `fix/usb-endpoint-direction` | Six inbound and seven outbound, per function and per direction |
| `fix/uac-request-number` | `req_number` 8, and both audio attributes actually applied |
| `fix/usb-test-case-timeout` | `CASE_TIMEOUT` 240, so the suite stops failing on a slow host |

The last one is not an adoption. Two cases in `test-usb-descriptors.sh` failed on
a tree with nothing wrong with it, because each case runs the whole of
`start_usb_dev` and `usb_bind` spends about nine seconds retrying when no
controller is present. On Git Bash on Windows the case does not finish inside 60
seconds. Confirmed by running the same unmodified tree at 240, where both pass.

### Deployed

`fork/integration` at `7f3e95b2`, cross-compiled and installed through
`tools/deploy/deploy-server` with `DEPLOY_TIMEOUT=240`. The guard reported
`OK, serving within 240s and running what was installed`, and the running
process, the on-disk copy and the staged candidate all hash the same.
`S03usbdev` went to `/etc/init.d` and `/kvmapp/system/init.d` both, and was
applied with a `stop_start` behind a check that restores the previous script if
the gadget that comes back is not the one that was there.

The web assets were not rebuilt. The only frontend change is one English string,
and the numbers the panel shows come from the API.

### Still not done

The two items from the `71ab9127` section stand. Beyond them:

- The other twenty-eight open pull requests are as this document already
  described them.
- `#888` has not been commented on upstream.
- The comment beside `patchelf` in `tools/build/build-app.sh` says `libkvm.so`
  links against five OpenCV libraries, and that
  `libopencv_video.so.409 ... not found` is expected during linking. That is
  true of upstream's committed library and no longer true of this fork's, which
  records four.

---

## Extraction branch rebase, 2026-08-28

The item that has stood since the `71ab9127` section, "the extraction branches with no pull request
open are not rebased", is done. Forty branches were in scope at first: every branch cut from an
upstream base that had no pull request open. The eight branches that did have one were then closed
upstream, which removed the reason to leave them alone, so they were rebased in the same pass. All
forty-eight extraction branches are now accounted for and forty-six sit on current `upstream/main`.
The two that do not are dead, and the reason is below.

`pr675`, `pr814` and `pr877` are copies of other contributors' pull requests rather than
extractions, so they stay as they are.

### A rebase produces the wrong branch

`git rebase --onto upstream/main` replays the branch's own old commits. Twenty-four branches rebased
with no conflict, and thirteen of those twenty-four came out carrying a version of the change that
`main` has since moved past. A clean rebase is therefore not evidence that the branch is right.

`main` sits on current `upstream/main`, so `main` already holds the reconciled version of every
change these branches carry. The operation that works is to cut the branch again from
`upstream/main` and cherry-pick `main`'s commits. That is how `security/api-key-auth` got the
post-`#876` shape of the auth code, which is what this item was about.

Cherry-picking brings `main`'s commit message with it, and that message is the wrong one. `main`'s
messages carry fork context: a reference to `hdmi_idle.go` that the extracted diff does not contain,
a pointer to `tools/README.md`, and in one case the line "NOT COMPILED AND NOT RUN". The extraction
branch's own message is the upstream-facing one. In the probe case it is also the more current one,
because it records a hardware check that came later. Every re-cut branch had its message put back,
and the tree was compared against the tree before that rewrite to prove that only the message moved.

Three further traps:

- `cherry-pick -x` writes `(cherry picked from commit ...)` with a hash that resolves nowhere
  upstream. Twelve branches already carried such a line from an earlier pass. They are stripped.
- `main` squashed several single-purpose fork commits into one. `c7a7943e` alone carries the logger,
  the router, `service/hid`, the frame rate counter, `service/ws` and `utils/memory`. The extraction
  branches are the unsquashed versions, which is their purpose, so "patch-equal to `main`" can no
  longer be required of them. Those changes were taken out by path instead.
- A worktree needs `core.autocrlf=false` set before the checkout, not after. Setting it after leaves
  every file reading as modified, and `git rebase` then refuses to start.

### Three branches were dead or duplicated

- `security/password-change-requires-current` asks for what upstream already does.
  `ChangePasswordReq` in `upstream/main` carries `CurrentPassword`, which `#876` brought in. Pull
  request `#848` is closed for the same reason.
- `fix/hdmi-signal-reported` asks for what upstream already does. `GetGetHdmiStateRsp` carries
  `Signal bool` from `#859`, and `main` dropped the fork's own reader in `a189c906`.
- `perf/webrtc-per-viewer-writer` has never compiled on its own. It uses `stream.FrameSlot`, which
  `perf/mjpeg-per-client` defines, and it did not carry that commit. Its final patch is identical to
  the last commit of `perf/webrtc-shared-packetizer`. Adding the missing prerequisite gives the two
  branches the same tree, so one of the two should go. This fault predates the rebase.

Two more branches lost part of what they proposed, because upstream landed that part:

- `fix/etc-kvm-file-modes` had three hunks and now has one. After `#876`, `authn/store.go` writes
  the account file 0o600 through a temporary file, and it calls `MkdirAll` with 0o755. Only the mode
  that `/etc/kvm` itself is created with still needs correcting.
- `security/usb-gadget-identity` dropped its MAC derivation. Upstream merged `#828` with the same
  fix, which is what `586ee30e` records.

### What was verified

Forty-two branches touch Go. Each one was stacked on `build/novision-tag` and put through
`go vet -tags novision ./...` and `go test -tags novision ./...` in `golang:1.25`. All forty-two
pass. Five failed on the first run. Four of the five were the known `picoclaw` and `ws` timing
flakes and passed on a second run. The fifth was `perf/webrtc-per-viewer-writer` above.

Stack the tag from the `build/novision-tag` branch, not from `main`'s copy of that commit. `main`'s
first version of the stub predates `HasHDMISignal`, which upstream added in `#859`. It fails
`go vet` on every branch, and the failure says nothing about the branch under test.

Three branches touch C++, and all three build in the MaixCDK builder. The build log confirms that
`kvm_vision.cpp` compiled for `fix/hdmi-probe-spin` and `fix/kvmv-thread-join`, and that
`kvm_mmf.cpp` compiled for `fix/vi-init-race`. Read the log rather than the exit status. This is the
first time the probe change has been compiled at all: `26781dc4` says "NOT COMPILED AND NOT RUN".

`fix/p2-resize-guard` touches shell only. Both of its files pass `sh -n`.

### The eight pull requests are closed

`#846`, `#847`, `#849`, `#850`, `#851`, `#852`, `#871` and `#873` are closed on `sipeed/NanoKVM`,
with no comment left on any of them. No branch was deleted, so any of the eight can be reopened
while its branch exists.

Those eight were the only branches still on an old base, and closing the pull requests removed the
reason to leave them there. They were cut from `71ab9127`, which is `#876` itself, so they were
already past the change that made the other batch stale and they were only three web-only commits
behind. All eight rebased with no conflict and none lost a commit. All eight pass the same novision
gate.

This also settles the `security/contain-request-paths` question. It is rebased now, so
`security/download-verify` and `feat/device-http-proxy` no longer stack on a copy of it that a pull
request contradicts.

### Still not done

- `feat/zram-swap` adds six files under `tools/`, which belongs to the fork. Whether a zram pull
  request should carry its module build tooling is a decision rather than a defect, so the branch
  keeps them.
- `fix/vi-init-race` and `main` no longer agree. The branch serialises with an RAII guard class.
  `main` folded the same lock into `160d9fc8` with an explicit unlock at each exit, beside a
  different change. The branch is the better single-purpose pull request, so the divergence stays.
- Nothing is pushed. Every rewritten branch keeps its previous head under
  `refs/rebase-backup/20260828/`, which is how the one accident in this pass was found and undone.

## Gadget supervision, adopted 2026-08-28

`mrjeeves/NanoKVM` `50b9e3c8`, "fix(usb): supervise the gadget instead of only repairing it at
startup". The first adoption made after the fork stopped contributing upstream, so it goes straight
onto a feature branch and into `fork/integration`. It is `fix/usb-gadget-supervision`.

Its diagnosis is correct here as well. Recovery ran at two moments and only two: `S03usbdev`'s boot
watch, which spends a fixed budget in the first seconds of uptime and exits for good, and an
operator pressing "reset USB". A link that failed at any other moment stayed failed. The board keeps
answering on `eth0` throughout, so every remote signal reads healthy and only the keyboard is gone.

Half of the source commit is already here, and half of it is wrong for this board.

### What was taken

The shape: a ticker that reads the controller, a debounce that is longer when the link has never
been seen working, a backoff, a settle window around deliberate gadget operations, and failed HID
writes as the evidence that separates a wedged gadget from a healthy one in a computer that is
switched off. Those two read identically at the controller, and that is the real difficulty in
acting on the reading at all.

The fault classifier was taken as an idea rather than as code. `mrjeeves` adds a package-level
counter of consecutive failed writes. This fork already records per-endpoint health with timestamps,
in `service/hid/health.go`, for `GET /api/hid/status`. So `ESHUTDOWN` and `ENODEV` became a fourth
endpoint state, `detached`, beside `stalled`. `Hid.linkFaultSince` folds the three endpoints into
the one answer the supervisor wants. That is a smaller change than the source commit makes, it
improves a response the web UI already renders, and it removes the import cycle that put the source
commit's supervisor in `service/storage`: this one lives in `service/hid` with everything it calls.

### What was rejected, and why

**The health test.** The source commit treats `configured` as health. On this board a gadget that
enumerated at full speed reports `configured` and does not work: the periodic bandwidth in a
full-speed frame does not hold three HID interrupt endpoints beside the console, the disk and the
speaker, so the host schedules what fits and silently stops polling the rest. On 2026-08-19 that
left this board with a working absolute mouse and a dead keyboard. `usb_link_ok` in `S03usbdev`
already makes the two-part test for the boot window; `usb_link.go` now makes it for the rest of
uptime. A full-speed link is a fault in its own right here, and it is one the source commit's
supervisor cannot see at all.

**The escalation to `restart_phy`.** The source commit rebinds, then resets the PHY. `restart_phy`
unbinds the dwc2 controller from its driver, and a bind that then fails leaves the gadget wedged
with no way back: rewriting `UDC` does not recover it and neither does another `restart_phy`. Only a
full configfs teardown does, and that needs a shell on the device. This board has no remote power
cycle. An unattended escalation that can strand it is not worth having however rare the failure is,
so the ladder stops at `stop_start`, which flips the controller role and composes the gadget again,
and which was observed returning a full-speed link to high-speed. `restart_phy` stays where it is,
behind an operator who is watching.

**The removal of `stop_start`.** The source commit removes it, on the grounds that `stop` leaves the
configfs tree standing so the following `start` mutates a live composite under an attached host.
That is true of their `stop`, which only unbinds the UDC. This fork's `stop` is `start_usb_host`,
which also writes `host` to `/proc/cviusb/otg_role`. That flips the hardware ID pin and takes the
controller through a full `dwc2_core_init()`, so the composite is not live when `start_usb_dev`
runs again. `stop_start` is the proven recovery here and it stays.

**The `S03usbdev` bind hardening.** Already present. `usb_bind` has taken its retry count from the
environment and read the `UDC` back since the boot watch was added.

### What is deliberately not covered

A gadget that is `configured` at high-speed and still has dead endpoints, because the composition
asked for more IN endpoints than the controller has transmit FIFOs. That reads healthy to this
supervisor by design. It is a composition fault, no amount of rebinding fixes it, and the endpoint
budget in `S03usbdev` and `service/vm/endpoints.go` refuses it before the bind.

### Verification

`go vet`, `go test` and a `GOARCH=riscv64` build, all under `-tags novision`, all clean. The
decision logic is a pure function and is covered case by case, including the full-speed grading, the
suspended-host case that must not trigger anything, and the ladder ending in a give-up rather than
rebinding forever.

**Not validated on hardware.** The source commit says the same about itself. What is untested here
is not the decision logic but the sysfs semantics it rests on, and the fork has hardware evidence
for those from 2026-08-19 that the source commit did not have.

### Still to adopt

- `RobbyV2/NanoKVM` `server/service/media`. **Read on 2026-08-30, and nothing is adopted.** The
  last section of this document says what the two files hold and why neither transfers.
- `eringiriri/ERINGI_JPN_NanoKVM`: horizontal scroll in relative mouse mode, and a composition guard
  that leaves the JIS Zenkaku key stranded. Both small and additive. **Both are adopted in the next
  section, and both are on the device since 2026-08-30.**
- `pi-bmc/nanokvm-app`, twelve `cvi` commits. "Stop the drivers' error reporting from killing the
  board" and "Drain the encoder even when it has just refused a frame" are the interesting two. It
  is a Go rewrite of the capture path, so it stays read-and-reimplement.

## Input fixes from eringiriri, adopted 2026-08-28

`eringiriri/ERINGI_JPN_NanoKVM` carries two changes on top of `71ab9127` that upstream does not
have. Both are input correctness, both are small, and neither needed reworking for this fork beyond
the descriptor.

### The composition guard stranded a key pressed

`ee4c48bf`, taken as `fix/keyboard-composition-release`.

The keyup handler returned early for every event while a composition was active. That is right for
a key the IME is consuming and wrong for a key the page has already sent a keydown for. The JIS
Zenkaku/Hankaku key is the second case: it toggles the local IME, so `compositionstart` can arrive
between its own keydown and its keyup. The keyup was dropped, the host kept the key held, and
because `pressedKeys` still held it, the next press was suppressed as a repeat. One press and the
key stopped working for the rest of the session, with a key held down on the managed host.

Release now runs for any key already tracked as pressed. A key the IME really is consuming was
never in `pressedKeys`, so it is still skipped. `releaseKeys` clears the composing flag too, which
the source commit also does: leaving it set after releasing everything recreates the fault.

The second half sends `Lang5` from the on-screen Japanese keyboard's Backquote position. `Lang5` is
already in `web/src/lib/keymap.ts`, so this is one line plus the key's label.

### Horizontal scroll had nowhere to go

`68bfdd74`, taken as `feat/mouse-horizontal-scroll`.

The relative mouse report was buttons, X, Y and a vertical wheel. A tilt wheel or a thumb wheel
reached the browser and stopped there. The fix adds a fifth byte, AC Pan from the Consumer page,
appended so that bytes 0 to 2 keep the USB boot mouse layout.

Three things were done differently from the source commit.

**Named constants rather than a literal swap.** The source replaces every `4` with `5`. In this
tree the literal appeared in seven files and doubled as a type discriminator: the mouse queue, the
websocket and the jiggler cooldown all told a relative report from an absolute one by its length.
`KeyboardReportLen`, `RelativeMouseReportLen` and `AbsoluteMouseReportLen` now carry that meaning.

**A test that holds the shell and the Go together.** This is the part worth keeping.
`server/service/hid/usb_report_test.go` reads both init scripts, decodes the `hid.GS1` descriptor,
totals the bits its Input items ask for, and holds `report_length` and the descriptor and the Go
constants to one number. It also checks that the horizontal wheel really is AC Pan and that it comes
after X and Y, because a descriptor of the right length with the wrong usage or the wrong order
would otherwise pass. Same pattern as `endpoints_shell_test.go` and the endpoint budget. Verified by
mutation: setting `report_length` back to 4 and stripping the AC Pan item each fail the right test
with the right sentence.

**`S03usbhid` as well as `S03usbdev`.** The HID-only mode carries its own copy of the descriptors.
The source commit changes both; it is worth saying out loud, because changing one leaves the other
mode broken and nothing on a running device would say so.

The cooldown test needed more than a new number, which the source commit also spotted. It asked
whether the last byte was non-zero to detect wheel movement. That was right while the wheel was
last. Left alone it would have stopped recognising an ordinary scroll as input, and the jiggler
would have nudged the pointer under someone reading a document. Here the two layouts are handled
apart as well, because byte 3 of an absolute report is half of the Y coordinate and counting it
would have made pointer movement suspend the jiggler.

### Deployed, 2026-08-30

Both halves are on the device, and the descriptor and the server agree about the report length.

The two init scripts in `/kvmapp/system/init.d` have the same md5 as the blobs on
`fork/integration`: `a6d582041d3057992e7845b23827681e` for `S03usbdev` and
`f8069690f8799ccd4d6baf8a41569e90` for `S03usbhid`. `/etc/init.d/S03usbdev`, which is the copy that
boot reads, matches as well. There is no `/etc/init.d/S03usbhid`, and there does not have to be:
`server/service/hid/status.go` runs the HID-only script from `/kvmapp` by its absolute path. The
gadget agrees with the scripts, because
`/sys/kernel/config/usb_gadget/g0/functions/hid.GS1/report_length` reads 5.

The running server is `dev.20260829.1819.ce311bb5`, and `feat/mouse-horizontal-scroll` is an
ancestor of that commit. The binary therefore sends five-byte reports to a gadget that asks for
five. The composition fix is frontend only, and the served tree is current: all 42 files under
`/tmp/server/web` are byte-identical to a fresh build of the branch head, and `/kvmapp/server/web`
matches the same build.

Before the deploy, both changes were verified off-device: `go vet`, the full `go test` twice, a
`GOARCH=riscv64` build, `tsc --noEmit`, eslint and prettier. The scroll change alters the USB
descriptor, so the two halves had to arrive together. A server that sends five bytes to a gadget
composed from the old script has its reports truncated, and the horizontal wheel is dropped with no
error anywhere.

## `feat/usb-audio` is superseded, 2026-08-30

The branch holds 28 commits cut from an old `main`. Every one of them reached `fork/integration` in
a later form, and the Opus migration then rewrote that form. `git merge-base --is-ancestor` still
reports the branch as unmerged, which is true of the SHAs and false of the content.

A re-cut onto today's `main` was tried and thrown away. It applied with four small conflicts, passed
`go vet`, all 29 novision test packages, a `GOARCH=riscv64` build, `tsc --noEmit`, eslint and the
frontend build, and it added nothing: `g711.go`, `resample.go` and one plan document are the only
files on it that `fork/integration` does not have, and the Opus work deleted the first two on
purpose.

Compare the files before acting on the merge-base. `origin` still holds the branch at `06871b81`.

## RobbyV2 `service/media`, read 2026-08-30

`hold.go` and `pcmloop.go` are read. Neither is adopted, and the bullet that asked for them was
wrong about one of the two.

### `pcmloop.go` describes a loop this fork does not have

The file is 41 lines. It sorts one non-blocking PCM transfer into three outcomes: a whole period
moved, the endpoint had nothing this tick (`EAGAIN`), or the stream is broken. The third case is the
one worth the reading. ALSA parks a substream after an under- or overrun, and it moves no more
samples until something prepares it again, so a loop that treats a broken transfer as an idle one
goes quiet for the rest of the binding.

That distinction belongs to a capture loop that calls ALSA directly through cgo. This fork captures
through an `arecord` child, so no return code reaches Go. A quiet host is a child that blocks and
delivers nothing. A parked stream is a child that exits. The equivalent of the ALSA prepare is
starting a new child, and `Source.Run` already does that.

The policy the file argues for is also already in `source.go`, arrived at from the device rather
than from this fork. Capture is never retired, because a host that plays nothing is this board's
ordinary idle state. The backoff climbs to 5 seconds, the log falls quiet after 5 consecutive
failures, and the child's own stderr line rides on the same lines so it falls quiet with them.

### `hold.go` is about cameras, not audio

The file keeps one `open(2)` per linked UVC video node. `f_uvc` sets `bind_deactivated`, so the
controller holds its soft-disconnect for the whole gadget until something opens the node, and the
kernel's deactivation counter is asymmetric: the release always increments, the activate refuses to
go below zero. A second overlapping open therefore leaks a deactivation that no later open pays
back, and the gadget stays off the bus with HID, the network and mass storage on it.

None of that reaches a fork with no UVC function. This one has none.

### The one idea worth carrying, and this tree already has it

Both halves of `hold.go` run their syscall on its own goroutine against a deadline, and retire the
node rather than reopen it, because a kernel that never returns must not take the manager with it.
That is a good rule on this board: a read of `/proc/cvitek/vb` blocks forever in D state and
survives every signal.

This tree applies the rule where it matters already. `Stream.Stop` bounds its wait at 2 seconds and
says why in a comment that names the same `/proc/cvitek/vb` case, and `service/hid/hid.go` sets a
write deadline before every write to `/dev/hidg*`.

### What is still worth reading there

`SeatFIFOs`, before any function that wants a wide isochronous IN endpoint is added, because only
FIFO 1 on this controller holds more than 512 words. `manager.go` and `output_linux.go` are 78KB
between them and stay a reimplementation rather than an adoption.
