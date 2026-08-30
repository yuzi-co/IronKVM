# OLED

Protect the status panel from burn-in, without changing `kvm_system`.

## Why this panel is exposed

The NanoKVM status screen is close to the worst case for an OLED. It shows the
same labels in the same pixels for the whole life of the device: `HDMI`, `USB`,
`ETH`, an IP address, a resolution. An OLED pixel dims as it is driven, so
pixels that are always lit lose brightness against pixels that never are, and
the layout stays visible as a ghost after the content changes.

Three levers exist, and they are independent:

| lever            | where it lives                     | effect                        |
| ---------------- | ---------------------------------- | ----------------------------- |
| screen off       | `/etc/kvm/oled_sleep`, Web UI      | the strongest, all or nothing |
| lower the drive  | `S97oled-nudge`, `OLED_CONTRAST`   | slows the wear everywhere     |
| move the image   | `S97oled-nudge`, `OLED_NUDGE_MAX`  | spreads the wear              |

**Set the sleep timer first, because a device that has never had one is the
case this whole directory exists for.** A missing `/etc/kvm/oled_sleep` does not
mean a default: `oled_auto_sleep` reads the absent file as 0, which is "never
sleep", and `GetOLED` reports 0 to the Web UI. So a board nobody has configured
runs a static screen for its whole life. The file is read within a second of
being written and no restart is needed:

```shell
echo 300 > /etc/kvm/oled_sleep         # or use Settings in the Web UI
```

The value is in seconds. Below 10, `kvm_system` disables sleeping rather than
shortening it, and the server clamps to that. A file that exists but is empty
means 30 seconds.

Any change of state restarts the countdown and wakes the panel, because
`kvm_main_ui_disp` calls `oled_auto_sleep_time_update` whenever
`kvm_state_is_changed()` is true. Plugging HDMI, an address change or a stream
starting all bring the screen back. **On an enclosure whose only buttons are
POWER and RESET there is nothing else that wakes it**, so the panel is dark
whenever the board is genuinely idle. That is the point, and it is worth knowing
before somebody wonders why the screen is off.

## What the wear looks like

Measured on a board that had run without a sleep timer since it was new: fill
the panel with every pixel lit, and the aged pixels are visibly darker, so the
old content reads as a photographic negative.

```shell
# every pixel on, eight pages of 128 columns, four i2cset blocks per page
for page in 0 1 2 3 4 5 6 7; do
    i2cset -y 5 0x3d 0x00 $((0xB0 + page)) 0x00 0x10 i
    for chunk in 1 2 3 4; do
        i2cset -y 5 0x3d 0x40 $(for i in $(seq 32); do printf '0xff '; done) i
    done
done
```

Pause `kvm_system` with `kill -STOP` first, or it redraws over the fill, and
`kill -CONT` afterwards. Do not kill it: `S98supervise` would start a second
copy, and two processes writing this panel is what corrupts it. Restarting
`kvm_system` afterwards is what puts the real screen back, because only a full
redraw rewrites every pixel.

The ghost on that board showed `TYPE: H264` and `QUALITY: Middle` while the live
screen read `MJPG` and `EXTRA`, and several IP addresses on top of each other.
The wear records what the panel showed for the longest, not what it shows now.

## How the nudge works

The SSD1306 display offset is a property of the controller, not of the frame
buffer. Command `0xD3` selects which COM line row zero drives, so the whole
image moves and the drawing code never knows. `kvm_system` writes that register
once in `OLED_Init` and never again, so a value set from outside survives every
redraw.

That is the reason this needs no rebuild of `kvm_system`. A content shift inside
the drawing code would work too, and it would not wrap - see the limit below -
but it needs the MaixCDK builder and a replacement binary.

`S97oled-nudge` walks the offset 0, 1, 2, 1, 0 and back again, one row at a
time, by default every 600 seconds. The walk is a triangle rather than a
saw-tooth: a saw-tooth snaps the whole image back in one step, which reads as a
glitch rather than as a screen that does not sit still.

## Install

```shell
scp tools/oled/S97oled-nudge root@<device>:/etc/init.d/S97oled-nudge
ssh root@<device> 'chmod 755 /etc/init.d/S97oled-nudge && /etc/init.d/S97oled-nudge start'
```

The number puts it after `S95nanokvm`, which starts `kvm_system` and therefore
initialises the display. The first move happens one period later in any case.

```shell
/etc/init.d/S97oled-nudge status        # panel, travel, period, running
/etc/init.d/S97oled-nudge demo 3        # walk 0..3..0 with a second between rows
/etc/init.d/S97oled-nudge stop          # stops, and puts the offset back to 0
```

Use `demo` to judge the movement on the panel. If the screen has gone to sleep,
wake it first, or the walk happens with the display off.

Environment: `OLED_NUDGE_MAX` rows of travel, 0 disables movement;
`OLED_NUDGE_PERIOD` seconds between moves; `OLED_CONTRAST` drive current, or
`keep` to leave it alone; `OLED_BUS` and `OLED_ADDR` to skip detection.

## Lowering the drive

`kvm_system` writes contrast `0xCF` in `OLED_Init`, which is about 81% of full
drive, and never writes it again. The script sets `0x60` by default.

A pixel ages with the light it has emitted and the relationship is worse than
linear, so a dimmer panel buys back more life than the brightness it costs. This
screen is read from arm's length in a rack rather than in sunlight.

The value is clamped into `0x10..0xFF`, because a drive of zero is a legal
command and an unreadable screen, and a script that runs at boot must not be
able to produce one. Anything that is not a number is treated as `keep`.

The loop re-applies it every period. That is not redundancy: a restart of
`kvm_system` runs `OLED_Init` again and puts the panel back to `0xCF`, so
re-applying is what makes the setting hold without anything having to watch for
the restart. `stop` puts `0xCF` back, so a stopped script never leaves a dim
panel behind with nothing to explain it.

## Why not periodic inversion

Command `0xA7` lights every pixel the page leaves dark, so running inverted part
of the time evens the wear out instead of letting it accumulate in one pattern.
It is the wrong trade here.

The status page lights roughly 15 to 20% of the pixels, so an inverted panel
drives 80 to 85%: four to five times the emitted light, and the wear follows the
light. It converts a legible ghost into uniform dimming, and it spends much more
of the panel to do it.

It also cannot repair what is already there. Equalising an existing ghost needs
the dark pixels to accumulate drive-hours comparable to what the lit ones banked
over the life of the board. At a duty cycle anybody would tolerate looking at,
that takes longer than the panel has left.

Inversion earns its place on a panel that must stay lit permanently, where
uniform dimming beats a readable ghost. A board with a sleep timer is not that
panel.

## Which bus and address

The script probes, and reports what it found. Do not assume:

| board            | `/etc/kvm/hw` | bus | address |
| ---------------- | ------------- | --- | ------- |
| Cube, Lite       | `beta`        | 5   | `0x3d`  |
| earlier revision | `alpha`       | 1   | `0x3d`  |
| PCIe             |               | 5   | `0x3c`  |

`kvm_system` builds both `oled_alpha(1)` and `oled_beta(5)` and picks by
hardware version. Probing is safer than reading `/etc/kvm/hw`, because a wrong
guess sends display commands to whatever else answers at that address. On this
board `i2cdetect` also finds `0x2b` and `0x44` on bus 4.

## The limit: the shift wraps

The offset moves the whole image and wraps. Rows pushed off the bottom reappear
at the top. With a travel of one or two rows on a 64 row panel that is invisible
while the bottom rows are blank, and obvious if they are not. Look at the screen
once with `demo` before you raise the travel.

If a larger travel is wanted without wrapping, the shift has to move the content
instead, which means `kvm_system`. `make support` builds it, and it can be
driven without a TTY the way `tools/build` drives the app build - the `-it` in
the Makefile is the only reason it appears to need one.

## The cost, and one race worth knowing

One I2C transaction per period. `kvm_system` writes the panel far more often
than that.

It sends each command byte as its own transaction, and the SSD1306 takes a
command's argument from the next byte it receives whatever transaction that byte
arrives in. So a write from here landing between `0x81` and its value is read as
the value, and the display is left mis-configured rather than merely
mis-drawn. No redraw corrects that, because a redraw writes pixels and not
configuration.

The exposure is narrower than it sounds. The only commands here that take an
argument are in `OLED_Init`, which runs when `kvm_system` starts, so the window
is a few milliseconds against a period of ten minutes. Cursor positioning is
three single-byte commands with no arguments, and a write landing between them
is harmless.

There is no way to hold an I2C bus against another process, so the race cannot
be closed from here, only made unlikely.

**A second `kvm_system` is the dangerous case, not this script.** It writes each
pixel byte as its own transaction after positioning the cursor, so two of them
interleave one another's bursts and the data lands at the wrong column. That
corruption does persist, because the UI redraws only the fields whose value
changed. It is reached by killing `kvm_system` and starting a copy by hand:
`S98supervise` fills the gap with one of its own. Replace the binary and then
kill the process, so exactly one is ever running.

## Tests

```shell
sh tools/oled/test-nudge.sh                        # the repository copy
sh /tmp/test-nudge.sh /etc/init.d/S97oled-nudge    # what is installed, on the board
```

`test-nudge.sh` extracts the offset walk and the probe order from the script
with `sed`, so the test cannot drift from what ships. Run it on the board as
well: busybox `ash` is not the shell it was written in.

## One trap

The nudge loop never returns, so it must not inherit the calling shell's stdio.
Started over ssh without `< /dev/null > /dev/null 2>&1` it holds the file
descriptor open and the ssh command never comes back. `tools/build/README.md`
records the same trap for `S95nanokvm`.
