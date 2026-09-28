# HID touch screen digitizer

Issue: yuzi-co/ironkvm-dist #7.

## Goal

iPadOS and Android handle an absolute mouse poorly. A touch screen digitizer (usage page
0x0D, usage 0x04) gives them direct touch from the web UI, and gives Windows and Linux a
second absolute mode for hosts where the absolute mouse misbehaves. No USB endpoint is added
and the absolute mouse reports do not change.

## Decision: a report ID next to the extended keys

The absolute pointer (`hid.GS2`) has no report ID by default. It has ID 1 only with
`/boot/usb.extkeys` (see `2026-09-26-hid-consumer-system-keys-design.md`), and then its
report is 7 bytes: the ID and the unchanged 6-byte pointer report.

Adding an ID to the default descriptor would change every pointer report, so the default
descriptor is left alone. The digitizer is added to the descriptor that already carries
IDs, behind its own marker, `/boot/usb.touch`. With the marker the pointer is ID 1 and
its reports are byte for byte what `/boot/usb.extkeys` gives today. The marker implies the
IDs, so the Consumer and System Control keys come with it.

Rejected:

- A mouse mode "touch" that swaps the descriptor. It re-enumerates the gadget on every mode
  switch, and the absolute mouse is gone while touch is selected. It remains the fallback if
  a host cannot take a mouse and a digitizer on one interface.
- IDs on the default descriptor. Every pointer report changes, and an older server sends the
  pointer without an ID.
- A new HID function. It costs one of the six IN endpoints.

## Report layout

`report_length` stays 7, the size of the pointer report with its ID. A digitizer report with
two contacts does not fit in 7 bytes, so the touch report carries one contact and uses the
hybrid mode that Windows and Linux both support: a frame with two contacts is two reports,
the first with Contact Count 2 and the second with Contact Count 0. Keeping 7 means the
server's existing check (report_length 7 turns the IDs on) still holds, and a server that
knows the extended keys but not touch keeps a working pointer on a touch gadget.

| ID | Type    | Content                                                             | Bytes with ID |
| -- | ------- | ------------------------------------------------------------------- | ------------- |
| 1  | Input   | Pointer, unchanged                                                  | 7             |
| 2  | Input   | Consumer Control, unchanged                                         | 3             |
| 3  | Input   | System Control, unchanged                                           | 2             |
| 4  | Input   | Tip Switch (1 bit), Contact ID (7 bits), X16, Y16, Contact Count 8  | 7             |
| 5  | Feature | Contact Count Maximum, logical maximum 2                            | 2             |
| 6  | Feature | Vendor 0xFF00 usage 0xC5, 256 bytes (the Windows certification blob)| 257           |

X and Y are 0 to 32767 with the physical range the pointer uses.

The descriptor appended to the extended-keys descriptor:

```
05 0d 09 04 a1 01 85 04          ; Digitizer, Touch Screen, application, ID 4
09 22 a1 02                      ; Finger, logical
  09 42 15 00 25 01 75 01 95 01 81 02      ; Tip Switch
  09 51 25 7f 75 07 95 01 81 02            ; Contact Identifier
  05 01 09 30 09 31 26 ff 7f 35 00 46 ff 7f 75 10 95 02 81 02   ; X, Y
  35 00 45 00                              ; physical range back to the logical one
c0
05 0d 09 54 25 02 75 08 95 01 81 02        ; Contact Count
85 05 09 55 25 02 75 08 95 01 b1 02        ; Contact Count Maximum, feature, ID 5
06 00 ff 85 06 09 c5 15 00 26 ff 00 75 08 96 00 01 b1 02   ; certification blob, ID 6
c0
```

The blob is needed on Linux, not only on Windows. `hid-multitouch` binds any device with a
Contact Identifier, and in its default class it ignores every collection that is not a touch
collection or a key collection, so the absolute mouse would stop working on the same
interface. The blob puts the device in the Windows 8 class, which exports all collections.

### Known limit: GET_REPORT

`f_hid` on this kernel answers every GET_REPORT with zeros. The host therefore reads Contact
Count Maximum as 0 and the blob as zeros. Linux then takes the logical maximum, 2. Windows
behaviour with a zero maximum and an invalid blob is not known and must be tested on a host.

### Held contacts

The Windows 8 class in `hid-multitouch` releases contacts that see no report for about
100 ms. The web UI repeats the current frame every 50 ms while a contact is down, so a
finger held still stays down.

## Gadget: `kvmapp/system/init.d/S03usbdev`

`/boot/usb.touch` selects the descriptor above, report_length 7. Otherwise
`/boot/usb.extkeys` and the default behave as before. `S03usbhid` does not change.

## Server: `server/service/hid`

- `TouchReportID` 4. When the server opens `/dev/hidg2` it reads `report_desc` from configfs
  next to `report_length` and walks the items. Touch is available when the descriptor has a
  Touch Screen application collection with a 48-bit input report under ID 4. The watchdog's
  refresh reopens the handle when either the ID or the touch state changes.
- A touch frame is one or two contacts: Tip Switch, contact ID 0 to 127, X and Y 0 to
  32767. It is written as one report per contact under the mouse lock, so a pointer report
  cannot fall between the two. A write without touch declared is refused and writes
  nothing.
- The mouse queue carries touch frames. Contacts still down when the queue closes or a write
  fails are lifted, the same way held buttons are released.
- `GET /api/hid/mode` gains `touch: bool`.

## Websocket

Message type 3, touch:

```
03 count { flags id xLo xHi yLo yHi } x count
```

`count` is 1 or 2, `flags` bit 0 is Tip Switch, ids are distinct and below 128, X and Y are
at most 32767. Anything else is dropped. The frame lists every contact that is down, plus
once more with Tip Switch clear each contact that has just lifted.

## Web UI

- The mouse mode menu offers "Touch" when `touch` is true. The mode is stored like the
  others. If it is stored but the gadget no longer declares touch, the absolute mouse is
  used.
- Touch mode uses Pointer Events on the screen with `touch-action: none`. Up to two pointers
  become contacts 0 and 1. A mouse also works: the left button is the finger. Taps,
  two-finger scroll and other gestures are interpreted by the host, not the browser.
- Moves are sent once per animation frame; down and up are sent at once.
- The stalled-endpoint warning covers touch mode too, since it uses the same endpoint.
- Strings in every locale.

## Deploy

1. Copy `S03usbdev` to `/etc/init.d` and `/kvmapp/system/init.d`, deploy the server.
2. Create `/boot/usb.touch` and run `S03usbdev stop_start`. The gadget re-enumerates.
3. To undo, delete the marker and run `stop_start` again.

## Verification on a board and a host

- Linux host: `/proc/bus/input/devices` lists a touchscreen input for the NanoKVM, bound by
  `hid-multitouch`, and the absolute mouse, Consumer Control and System Control inputs still
  exist and work. `evtest` shows `ABS_MT_POSITION_X` and two slots.
- The absolute mouse moves and clicks with the marker set.
- Windows: Device Manager shows a HID-compliant touch screen without an error, and Pen and
  Touch in Settings reports touch input with 2 touch points. Tap, drag, two-finger scroll
  and press-and-hold work.
- Android: taps and scrolling.
- iPadOS: whether it takes a USB touch screen at all is not known. Test it before the
  issue's premise is taken as met.
