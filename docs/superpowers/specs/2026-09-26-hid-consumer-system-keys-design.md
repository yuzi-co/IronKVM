# HID Consumer and System Control keys

Issues: yuzi-co/ironkvm-dist #5 (System Control: power, sleep, wake) and #6 (Consumer
Control: volume, media keys).

## Goal

The web UI can send media keys and the power, sleep and wake keys to the managed host. The
change spends no USB endpoint and does not touch the keyboard or the relative mouse.

## Decision: report IDs on the absolute pointer

The keyboard (`hid.GS0`) and the relative mouse (`hid.GS1`) declare the boot protocol. A BIOS
or UEFI setup screen reads their reports without a report ID, so an ID byte would break the
keyboard in the one place where a KVM keyboard matters most. The absolute pointer (`hid.GS2`)
is never a boot device, so report IDs are safe there. A fourth HID function would spend one of
the six IN endpoints, which #10, #11 and #12 also want.

Rejected:

- Report IDs on the keyboard. Hosts accept it, but boot-protocol firmware does not, and
  `f_hid` on 5.10 does not tell userspace which protocol the host selected.
- A new `hid.GS3`. Clean descriptors, but it costs an IN endpoint.

## Gadget: `kvmapp/system/init.d/S03usbdev`

If `/boot/usb.extkeys` exists, `hid.GS2` gets one report descriptor with three
top-level collections, and `report_length` changes from 6 to 7. Without the
marker, `hid.GS2` keeps its plain 6-byte pointer, as before.

The marker is opt-in because a server that predates the IDs sends the pointer
without one, and the host then drops or misreads it. A rolled-back server
deploy, an upstream application update, or any build of a branch without this
feature leaves that state. The marker also lets the owner turn the IDs off if a
firmware setup screen handles them badly.

| ID | Collection                     | Report after the ID                     | Bytes with ID |
| -- | ------------------------------ | --------------------------------------- | ------------- |
| 1  | Generic Desktop, Mouse         | Unchanged: buttons, X16, Y16, wheel     | 7             |
| 2  | Consumer, Consumer Control     | One 16-bit usage, 0 to 0x3FF, array     | 3             |
| 3  | Generic Desktop, System Control| One 8-bit usage, 0 to 0xFF, array       | 2             |

The descriptor, in hex, with the inserted and appended items marked:

```
05 01 09 02 a1 01  85 01                       ; ID 1 inserted after the application collection
09 01 a1 00 05 09 19 01 29 05 15 00 25 01 95 05 75 01 81 02
95 01 75 03 81 01 05 01 09 30 09 31 15 00 26 ff 7f 35 00 46 ff 7f
75 10 95 02 81 02 05 01 09 38 15 81 25 7f 35 00 45 00 75 08 95 01
81 06 c0 c0
05 0c 09 01 a1 01 85 02                        ; Consumer Control, ID 2
15 00 26 ff 03 19 00 2a ff 03 75 10 95 01 81 00 c0
05 01 09 80 a1 01 85 03                        ; System Control, ID 3
15 00 26 ff 00 19 00 29 ff 75 08 95 01 81 00 c0
```

A press writes the usage; a release writes usage 0. Common usages: Consumer 0xE2 Mute, 0xE9
Volume Up, 0xEA Volume Down, 0xCD Play/Pause, 0xB5 Next, 0xB6 Previous, 0xB7 Stop. System
0x81 Power Down, 0x82 Sleep, 0x83 Wake Up.

`S03usbhid` (hid-only mode) stays as it is. Its descriptors differ from `S03usbdev` on
purpose, and `tools/usbdev/test-usb-descriptors.sh` asserts that. Hid-only mode exists for
hosts that are picky about the gadget, so it keeps the 6-byte pointer without IDs, and the
new keys are not offered in that mode.

## Server: `server/service/hid`

- The pointer's `hidDevice` asks `writeHID` to put the report ID read from configfs in
  front of the report when that ID is non-zero. Every current writer of the absolute pointer (the
  websocket queue, the jiggler, MCP, the release path) keeps sending 6 bytes.
- When the server opens the devices, it reads
  `/sys/kernel/config/usb_gadget/g0/functions/hid.GS2/report_length`. A value of 7 turns the
  report IDs on; any other value, or a missing file, turns them off. So a new server on a
  gadget without IDs, including hid-only mode, keeps a working pointer. Without this check
  that mismatch breaks the absolute mouse: `f_hid` truncates a 7-byte write to 6. The
  reverse case, an older server on a gadget with IDs, has no guard in the server, because
  the older server cannot be changed. The marker is the guard: keep it absent on a board
  that can run an older server.
- `AbsoluteMouseReportLen` stays 6. `usb_report_test.go` holds `S03usbdev` to the two
  lengths it can give `hid.GS2` (7 with the marker, 6 without), and `S03usbhid` to 6.
- New route `POST /api/hid/key`, under `CheckToken` like the other HID routes. Body
  `{"page": "consumer" | "system", "usage": <int>}`. Validation: consumer 1 to 0x3FF,
  system 0x81 to 0xB7. The handler takes the mouse lock, writes the press, waits 50 ms and
  writes the release. The release is written even if the wait is interrupted, and is tried
  once more if it fails. If report IDs are off, the route returns an error code and writes
  nothing; the write decides that under the mouse lock, so a mode switch cannot slip a key
  report onto a pointer without IDs.
- `GET /api/hid/mode` gains `extendedKeys: bool`, which is the state of the report IDs.

## Web UI

- Power menu: a "Host OS" group below the ATX buttons with Sleep, Wake and Power Down.
  These follow the existing "show confirm" switch. The group is hidden when
  `extendedKeys` is false.
- Keyboard menu: a "Media keys" row with Mute, Volume Down, Volume Up, Previous,
  Play/Pause, Next and Stop. Hidden when `extendedKeys` is false.
- Strings go into `en.ts`; other locales fall back to English.
- The browser's own media keys are not captured. Browsers mostly consume them.

## Deploy

1. Copy `S03usbdev` to `/etc/init.d` and `/kvmapp/system/init.d`.
2. Deploy the server with `tools/deploy/deploy-server`. It reads `report_length` 6 and sends
   6-byte reports, as before.
3. Create `/boot/usb.extkeys`.
4. Run `S03usbdev stop_start`. HID drops for a few seconds; the server reopens the devices,
   reads 7 and turns the report IDs on. To undo, delete the marker and run
   `stop_start` again.
5. The next ironkvm-dist image takes the script through the manifest. The image does not
   create the marker.

## Verification

- On the managed host (unraid), `/proc/bus/input/devices` lists "Consumer Control" and
  "System Control" inputs for the NanoKVM, with `KEY_VOLUMEUP` and `KEY_SLEEP` in their key
  bitmaps.
- The owner confirms the absolute mouse moves and clicks after step 4.
- Volume Up from the UI produces an event on the host's input device.
- Sleep and Power Down are tested on the host only when the owner asks.

## Limits

- Wake Up, and any key sent to a sleeping host, needs USB remote wakeup, which is
  `/boot/usb.wakeup`. Without it, the host ignores reports while it sleeps.
- Power Down does what the host's power-key handler does. On a server that may be a
  shutdown.
