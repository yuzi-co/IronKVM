# Host watchdog

Issue: yuzi-co/ironkvm-dist #13.

## Goal

If the host should be up, and the picture has not changed or there is no HDMI signal for N
minutes, the board presses the ATX reset button, or power-cycles the host, as the owner chose.
Every action goes to a log that the web UI shows, with a screenshot taken before the press.

## Decisions

- **In the server, off by default.** A package `service/watchdog` with one goroutine. When the
  watchdog is off, the goroutine reads the setting every 10 seconds and does nothing else: no
  capture, no ping, no GPIO read.
- **Settings in server.yaml, under `watchdog`.** `enabled`, `timeoutMinutes`, `action`
  (`reset` or `power`), `cooldownMinutes`, `maxPerHour` and `pingHost`. A file without the
  block loads with the watchdog off, and a zero value means the default. The web UI saves them
  the way it saves `hardware.powerLed` and `redfish.enabled`: to server.yaml and to the running
  configuration, with no restart.

| Setting | Default | Range |
| --- | --- | --- |
| `timeoutMinutes` | 5 | 1 to 1440 |
| `action` | `reset` | `reset`, `power` |
| `cooldownMinutes` | 15 | 1 to 1440 |
| `maxPerHour` | 3 | 1 to 20 |
| `pingHost` | empty (off) | an IPv4 or IPv6 address |

## Inputs

The watchdog samples every 10 seconds.

- **Frame change.** One JPEG through the same capture path as the MCP screenshot tool, at the
  stream's current size and a low quality. It never changes the capture resolution, because
  that would resize every viewer's stream. The watchdog keeps a SHA-256 of the JPEG bytes. The
  HDMI input is digital and the encoder is deterministic, so a frozen picture gives the same
  bytes. Any change, even a blinking cursor, counts as a change. That errs on the side of no
  reset. When MJPEG frame detection is on for a viewer, the library answers "image not
  changed" for a still picture, and the watchdog counts that as unchanged.
- **HDMI signal.** Read after the sample, because the sample resumes capture if the idle timer
  stopped it. The return of the signal counts as a change.
- **Power LED.** Used only when "Power LED connected" is on. If the LED is off, or cannot be
  read, the host is not "should be up": the watchdog does nothing and restarts its timer.
  Without the LED the watchdog assumes the host should be up.
- **Ping.** Optional. One ICMP echo each sample, with a 2 second timeout. A reply counts as a
  sign of life, the same as a picture change. This keeps a headless server with a still
  console, or a host whose display went to sleep, from being reset.

A sample that fails for any other reason counts as a sign of life. The watchdog cannot prove a
hang from a failed capture, and a false reset costs more than a missed one.

If HDMI capture is turned off in the UI, there is no picture to judge, and the watchdog waits.

## Trigger

The host should be up, and the last sign of life is at least `timeoutMinutes` old. The
watchdog then acts, unless the last action is less than `cooldownMinutes` old or
`maxPerHour` actions already happened in the last hour. The start of the process, turning the
watchdog on, a change of settings, the LED going off and every action restart the timer.

- `reset`: reset press, 800 ms.
- `power`: power press, 5 s (force off), a 5 second pause, then power press, 800 ms (on). A
  short press alone does nothing to a hung host.

Both go through `vm.PressButton`, which serializes them with the UI and Redfish.

## Log

Before each press the watchdog takes a screenshot at the stream's size and quality. The log
keeps the last 50 actions in `/data/ironkvm/watchdog/log.json`, and each screenshot as
`/data/ironkvm/watchdog/<id>.jpg`. An entry that falls off the end takes its screenshot with
it. The log is written only when the watchdog acts, at most `maxPerHour` times an hour, so the
card sees no steady writes. An entry records the time, the action, the reason (`frozen` or
`noSignal`), how long the host showed no sign of life, the screenshot, and the press error if
there is one.

## API

All routes need a web UI login with the admin role, as the Redfish page's routes do.

| Route | Content |
| --- | --- |
| `GET /api/watchdog/settings` | The settings |
| `POST /api/watchdog/settings` | Save the settings |
| `GET /api/watchdog/state` | The detector state: last change, signal, LED, ping, the time until the watchdog acts, and why it waits |
| `GET /api/watchdog/log` | The log, newest first |
| `GET /api/watchdog/log/:id/screenshot` | The screenshot of one entry, `image/jpeg` |

`/api/metrics` carries `ironkvm_watchdog_actions_total{action}`.

## Cost

With the watchdog on, the board encodes one JPEG every 10 seconds and hashes it. The sample
also keeps HDMI capture from stopping for idle, so the capture core keeps running. That is the
price of watching the picture, and the reason the watchdog is off by default.

## Web UI

A Watchdog page under Settings, for administrators: the switch, the settings, the detector
state, and the log with a screenshot for each entry.
