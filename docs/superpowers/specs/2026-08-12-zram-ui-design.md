# zram toggle and status in the web UI

Date: 2026-08-12
Status: approved, ready for an implementation plan

## Summary

zram already works on this board, but only as a manual installation. `tools/zram/build-modules.sh`
builds two loadable modules against the stock kernel, and `tools/zram/S01zram` enables compressed
swap at boot. Both steps are done by hand over SSH, and the web UI shows nothing about them.

This design adds one row to `Settings > Device > Advanced`. The row turns zram on or off, it makes
the choice survive a reboot, and it reports what zram currently holds. It also repairs a defect in
the existing file-swap control, which turns zram off as a side effect.

The kernel modules stay a manual step. The server does not build them and does not ship them.

## Scope

In scope:

- A `GET`/`POST /api/vm/zram` pair in the Go backend.
- One React row in `Settings > Device > Advanced`, above the existing Swap row.
- Boot persistence, through an init script that the server installs and removes.
- A narrower `swapoff` in the file-swap service, so the two controls stop interfering.
- A fixed swap priority for zram, so it is preferred over the SD-card swap file.

Out of scope:

- Building or shipping `zsmalloc.ko` and `zram.ko`. The device must already have them.
- Changing the size, the memory limit, or the compression algorithm from the UI. These stay the
  constants in the init script: 96 MB of disksize, a 40 MB memory limit, and `lzo-rle`.
- Translations beyond English. The other locale files fall back to English.
- Any change to `kvm_system` (C++). The design avoids a MaixCDK rebuild on purpose.

## Why zram is not simply always on

The board reserves 75 MB for the video pipeline and leaves about 158 MB for Linux. Swap therefore
helps. But there is no disk swap behind zram, so exceeding zram calls the OOM killer instead of
paging slowly, and the largest process is `NanoKVM-Server`. An OOM there takes video down with it.
`tools/zram/README.md` records the measurements and the trade. An operator must be able to turn the
feature off, and must be able to see how close it is to its limit. That is what this row provides.

## Three states, not one

The server reports three independent booleans. A single "on" flag hides the failure that matters.

| Field | Source | Meaning |
| --- | --- | --- |
| `available` | `/mnt/system/ko/zsmalloc.ko` and `/mnt/system/ko/zram.ko` both exist | The modules are installed. |
| `enabled` | `/etc/init.d/S01zram` exists | The setting survives a reboot. |
| `active` | `/dev/zram0` appears in `/proc/swaps` | Compressed swap runs now. |

If `available` is false, the UI disables the toggle and points the operator at
`tools/zram/README.md`. If `enabled` is true and `active` is false, zram is configured but did not
start. The UI reports that as a warning, not as "off", because reporting it as "off" invites the
operator to toggle it on and see nothing change.

## API

`GET /api/vm/zram` returns:

```go
type GetZramRsp struct {
    Available  bool   `json:"available"`
    Enabled    bool   `json:"enabled"`
    Active     bool   `json:"active"`
    Algorithm  string `json:"algorithm"`
    DiskSize   int64  `json:"diskSize"`
    Original   int64  `json:"original"`
    Compressed int64  `json:"compressed"`
    MemUsed    int64  `json:"memUsed"`
    MemLimit   int64  `json:"memLimit"`
    SwapIn     int64  `json:"swapIn"`
    SwapOut    int64  `json:"swapOut"`
}
```

All sizes are bytes. `POST /api/vm/zram` takes `{"enabled": true|false}`.

Both routes go under `middleware.CheckToken`, beside the existing swap routes in
`server/router/vm.go`. Responses use the `proto.Response` helpers.

## Reading the numbers

`Algorithm` comes from `/sys/block/zram0/comp_algorithm`, which marks the active entry with square
brackets. The server extracts the bracketed name.

`DiskSize` comes from `/sys/block/zram0/disksize`.

`Original`, `Compressed`, `MemUsed` and `MemLimit` come from `/sys/block/zram0/mm_stat`. That file
holds one line of whitespace-separated integers. The first four are `orig_data_size`,
`compr_data_size`, `mem_used_total` and `mem_limit`. The parser must tolerate a missing file, a
short line, and a non-numeric field. In each of those cases it returns zeros. It must not return an
error, because a device without zram is a normal state, not a fault.

`mem_limit` is write-only in sysfs. Field 4 of `mm_stat` is the only way to read it back.

`SwapIn` and `SwapOut` come from the `pswpin` and `pswpout` lines of `/proc/vmstat`. These counters
are system-wide: they include the file swap, and they do not reset when zram restarts. The UI text
must say so.

The compression ratio is not sent. The frontend computes it from `Original` and `Compressed`, and
shows a dash when `Compressed` is zero.

## Turning zram on and off

The init script list that `kvm_system` copies into `/etc/init.d` is hard-coded in
`support/sg2002/kvm_system/main/lib/system_init/system_init.cpp`. Adding a name to that list needs a
MaixCDK rebuild and a `kvm_system` redeploy on every device. The server installs the script instead.
`S98tailscaled` already follows this pattern: the presence of the file in `/etc/init.d` is what
marks the feature as installed.

`git mv tools/zram/S01zram kvmapp/system/init.d/S01zram` gives the server a source to copy from.
`.gitattributes` already forces LF on `kvmapp/system/init.d/*`, so the line endings are correct
without a new rule.

To enable:

1. Confirm both modules exist. If either is missing, return an error and change nothing.
2. Copy `/kvmapp/system/init.d/S01zram` to `/etc/init.d/S01zram`, mode 755.
3. Run `/etc/init.d/S01zram start`.
4. Read `/proc/swaps` again. If `/dev/zram0` is absent, remove `/etc/init.d/S01zram` and return an
   error.

Step 4 keeps `enabled` honest. A script that stays installed after a failed start would report
"survives reboot" for a feature that does not run.

To disable:

1. Run `/etc/init.d/S01zram stop`, if the script is present.
2. Remove `/etc/init.d/S01zram`.

Both directions are idempotent. Enabling an enabled device, or disabling a disabled one, succeeds
and changes nothing.

## The file-swap collision

`disableSwap()` in `server/service/vm/swap.go` runs `swapoff -a`, which stops every swap device.
`SetSwap` calls it when the operator selects "Disable", and `enableSwap()` calls it before it
resizes an existing swap file. So any change to the file-swap control turns zram off, and the zram
row would report "enabled but not active" for a reason the operator cannot see.

The fix is to name the file: `swapoff /swapfile`. A missing file must not make the call fail, so the
service checks `os.Stat` first and skips the command when the file is absent.

This changes nothing on a device without zram. It is a defect in the existing code either way,
because `swapoff -a` was never the stated intent of a control labelled "swap file size".

## Swap priority

Both swap devices take a kernel-assigned priority in `swapon` order. `S01zram` runs from `rcS`, and
the file swap runs from the `si11::sysinit` line that `enableInittab()` appends to `/etc/inittab`.
The order between them is not defined. If the file swap wins, the kernel writes to the SD card
before it writes to compressed RAM, which is the opposite of the intent.

`S01zram` therefore uses `swapon -p 100 /dev/zram0`.

BusyBox `swapon` is documented to accept `-p PRI`, but the applet on this image may be built without
it. Verify on hardware before this lands. If the flag is rejected, keep the plain `swapon` and record
the limitation in `tools/zram/README.md` instead.

## Frontend

`web/src/pages/desktop/menu/settings/device/advanced/zram.tsx` renders one row, above `<Swap />` in
`index.tsx`. It follows the layout of `swap.tsx`: a title with an information tooltip on the left, a
description below it, and the control on the right. The control is an antd `Switch` rather than a
`Select`, because the setting is binary.

Under the description, a compact status line, for example:

```
Active - 3.2 MB of 96 MB, 2.8x
```

The tooltip on the status line adds the algorithm, `mem_used` against `mem_limit`, and the
system-wide swap-in and swap-out counters.

The row fetches status when the Advanced section opens and again after each toggle. It does not
poll. A board with one core should not answer a repeating request to show a number that changes
slowly.

`web/src/api/vm.ts` gets `getZram()` and `setZram(enabled: boolean)`, beside the existing swap
functions. Strings go under `settings.device.zram` in `web/src/i18n/locales/en.ts`.

## Testing

Go tests run under `-tags novision`:

- The `mm_stat` parser, table-driven: a normal line, a short line, a non-numeric field, an empty
  file, and a missing file.
- The `comp_algorithm` parser, with and without a bracketed entry.
- Status assembly across the `available`, `enabled` and `active` combinations, including
  `enabled && !active`.
- `disableSwap()` with the swap file absent, which must succeed.

The path constants in the zram service become package variables so a test can point them at a
temporary directory.

Hardware verification on the device:

1. Toggle on. Confirm `/proc/swaps` lists `/dev/zram0`, then reboot and confirm it is still listed.
2. Toggle off. Confirm the device is gone, then reboot and confirm it stays gone.
3. With zram on, set the file swap to 64 MB and then to Disable. Confirm zram stays active through
   both.
4. Confirm `swapon -p 100` is accepted, and that `/proc/swaps` gives zram the higher priority.

There is no frontend test runner, so the React row is checked by hand.

## Documentation

`tools/zram/README.md` stays where it is. It gains a pointer to the new script location and a note
that the UI toggle installs and removes the script. `tools/README.md` gains the same pointer.

## Branch

This is fork work, cut from `main`. It does not go upstream: the upstream image carries neither the
zram modules nor `tools/zram`, so the toggle would report `available: false` on every upstream
device.
