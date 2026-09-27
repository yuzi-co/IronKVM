# Two virtual drives: a disk and a CD-ROM at the same time

Issue: yuzi-co/ironkvm-dist #8. Redfish (#2) builds its virtual media on this.

## Goal

The managed host sees two drives at once: a writable removable disk and a CD/DVD-ROM. An
install ISO and a driver or answer-file image can then be attached together. Inserting or
ejecting either drive never re-enumerates the USB gadget, so the keyboard and mouse are never
disturbed by a media change.

## Today

`mass_storage.disk0` has one LUN, `lun.0`, with `removable=1`. The server's `MountImage`
switches it between disk and CD-ROM by writing `cdrom`, `ro` and `inquiry_string`, and the
host reads those only at attach time. So every mount closes the HID handles and unbinds and
rebinds the UDC. Swapping media on a removable LUN by writing `file` alone needs no rebind.

## Decision: a fixed CD drive on `lun.1`

`lun.1` is created once, at gadget build, as a CD-ROM, and never changes type. `lun.0` stays a
disk. Neither LUN changes `cdrom` or `inquiry_string` after that, so a media change is only a
write to `file`.

The CD drive is always present, even when empty. The host shows an empty optical drive (on
Windows, a drive letter with no disc), and a BIOS skips it. This was chosen over creating
`lun.1` on demand, which would need a UDC rebind each time it appears or goes away.

Both LUNs share the function's one bulk pair, so the endpoint budget (6 IN, 7 OUT) does not
change.

## Gadget: `kvmapp/system/init.d/S03usbdev`

Inside the existing `usb_kept disk` block, after `lun.0` is set up:

```sh
mkdir functions/mass_storage.disk0/lun.1
echo 1 > functions/mass_storage.disk0/lun.1/removable
echo 1 > functions/mass_storage.disk0/lun.1/cdrom
echo 1 > functions/mass_storage.disk0/lun.1/ro
echo "NanoKVM USB CD/DVD-ROM  0520" > functions/mass_storage.disk0/lun.1/inquiry_string
```

`lun.1/file` is left unset, so the drive reports no medium. The inquiry string follows the
format the server writes today: vendor padded to 8, product padded to 16, then the version.

`lun.0` keeps its setup, including `/boot/usb.disk0.ro`. Its CD mode is no longer used.

The drives exist only while the disk function is on (`/boot/usb.disk0`), as `lun.0` does today.

## Server: `server/service/storage`

### Drive layer

A new file, `drives.go`, owns both LUNs:

- `disk`: `lun.0`, type `disk`.
- `cdrom`: `lun.1`, type `cdrom`.

The configfs function directory is a package variable, so tests point it at a temporary
directory. A drive is listed only if its LUN directory exists. A server running on a gadget
built before this change therefore lists only `disk`.

**Insert** (`drive`, `file`, `ro`):

1. `file` must pass `isMountableImage` (a `.iso` or `.img` under `/data`, no traversal).
2. `file` must not be loaded in the other drive. A writable disk and a CD on one backing file
   would corrupt it.
3. If the drive holds a medium, eject it first (below).
4. Disk only: write `ro`. The kernel accepts a change to `ro` only while no medium is present,
   which is why this comes after the eject.
5. Write `file` through `writeMountTarget`.

**Eject**: write an empty line to `file`. If the host has locked the medium (Linux does while a
CD is mounted), the kernel returns `EBUSY`. The 5.10.270 kernel has no `forced_eject`
attribute (the board's `lun.0` lists `cdrom file inquiry_string nofua removable ro`), so there
is no fallback. The call fails with a message saying the host holds the medium and must eject
it first. The server does not rebind the UDC to force it.

Neither operation closes the HID handles, rebinds the UDC or calls
`hid.NoteUSBGadgetMutated()`, because neither changes the gadget.

### Routes

| Route | Body | Result |
|-------|------|--------|
| `GET /api/storage/drives` | | `{drives: [{id, type, file, ro}]}` |
| `POST /api/storage/drives/:id/insert` | `{file, ro}` | ok, or an error |
| `POST /api/storage/drives/:id/eject` | | ok, or an error |

`file` is `""` when the drive is empty. `ro` is always true for `cdrom`. In HID-only mode, or
with the disk function off, `drives` is empty and insert and eject fail with "no such drive".

### The old endpoints

They stay, as thin wrappers, so an older UI build or a script keeps working:

- `POST /api/storage/image/mount {file, cdrom}`: a non-empty `file` inserts into `cdrom` if
  `cdrom` is true, otherwise into `disk`. An empty `file` ejects both drives.
- `GET /api/storage/image/mounted`: the CD's file if it has one, otherwise the disk's.
- `GET /api/storage/cdrom`: 1 if the CD drive holds a medium, else 0.

On a gadget without `lun.1`, `mount {cdrom: true}` fails with "no CD drive". It does not fall
back to switching `lun.0`, because that path is the one being retired.

### Delete

`POST /api/storage/image/delete` refuses an image that is loaded in either drive. Today it has
no such check.

## Web UI: `web/src/pages/desktop/menu/image`

- A drives section at the top of the image menu, one row per listed drive: its name (Disk,
  CD), the loaded file or "empty", and an eject button. The disk row has a read-only switch,
  which applies at the next insert.
- In the image list, a click inserts the image: `.iso` into CD, `.img` into Disk. A small drive
  selector on the row overrides the default. An image loaded in a drive is marked with that
  drive, and the other drive cannot take it.
- The global CD-ROM mode switch goes away.
- An insert or eject that fails shows the server's message. The locked-medium case says to
  eject on the host.
- `web/src/api/storage.ts` gains `getDrives`, `insertDrive` and `ejectDrive`.
- New strings go in `en.ts`. The other locales fall back to English, the same gap #22 tracks.

## Deploy

The server and the gadget script deploy separately, and either order works:

- New server, old gadget: only `disk` is listed. The UI shows one drive.
- Old server, new gadget: the old server still switches `lun.0` with a UDC rebind, and `lun.1`
  sits empty. Nothing breaks.

The gadget change takes effect at the next gadget rebuild (a reboot, or S03usbdev
`stop_start`), which re-enumerates the gadget once.

## Verification

Go tests in `server/service/storage`, against a fake configfs directory:

- Insert and eject on each drive write the expected `file` contents.
- The disk's `ro` is written after the eject and before the new `file`.
- The same file cannot go into both drives.
- An `EBUSY` from `file` returns the locked-medium error.
- A missing `lun.1` lists only `disk`.
- The old endpoints map onto the drives as described.
- Delete refuses a loaded image.

On the board, with unraid as the host:

- With both drives loaded, the host shows the disk as `sdX` and the CD as `srX` at once.
- Inserting and ejecting each drive leaves the gadget's USB device number unchanged, and the
  keyboard and both pointers keep their `eventN` numbers.
- A 1 MiB ISO in the CD drive reads back byte-identical.
- An eject while the host has the CD mounted returns the locked-medium error.

Not tested: a Windows 11 install from the CD drive, in line with #9.

## Out of scope

- Keeping mounts across a reboot.
- More than two drives.
- Redfish (#2).
- Backporting `forced_eject` to 5.10.270. A possible follow-up if locked media turn out to be
  common.
