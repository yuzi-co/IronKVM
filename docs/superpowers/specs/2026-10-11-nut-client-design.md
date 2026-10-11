# NUT client

Issue: yuzi-co/ironkvm-dist #130. Related: #13 host watchdog.

## Goal

Hosts that cannot run NUT themselves get a clean shutdown on low battery, and hosts without
"restore on AC power" come back when mains returns. The KVM reads a UPS from a NUT server on the
network and presses the host's ATX power button.

## Decisions

- **A NUT protocol client in Go, not `upsmon`.** The protocol is one line per command on TCP
  3493. The client needs five commands (`USERNAME`, `PASSWORD`, `LOGIN`, `LOGOUT`, `LIST VAR`)
  and is about 250 lines in `service/nut/client.go`. A Go library would add a dependency of which
  this uses a fraction, with no read deadlines of its own. `upsmon` would add an Alpine package, a
  config file and a second process on a board with 158 MB of RAM, and its `SHUTDOWNCMD` and
  `NOTIFYCMD` would still have to call back into the server to press a button. The UPS state also
  has to reach the UI and the status bar, which a separate process would do through files.
- **In the server, off by default.** A package `service/nut` with one goroutine. Off, it reads the
  setting every 5 seconds and opens no connection.
- **Settings in server.yaml, under `nut`.** The password is in `/etc/kvm/nut.passwd` (0600, as the
  VNC password is), never in server.yaml, never logged, never returned by the API (`passwordSet`
  says whether there is one). The NUT protocol is plain text here: the user and password cross
  the network in the clear, as they do for `upsmon` without a certificate.
- **A shared power helper.** `service/hostpower` holds the press lengths (800 ms short, 5 s long)
  and the power LED check. The watchdog and the NUT client press through it. A checked press
  (`Shutdown`, `ForceOff`, `PowerOn`) reads the LED first and does nothing if the host is already
  in the wanted state or if the LED cannot say. The watchdog keeps its own gate (the LED says the
  host should be up) and its behaviour; it uses the helper for the LED read, the reset press and
  the power cycle.
- **Needs the power LED.** A short press on a host that is off switches it on, so without the LED
  the client cannot press safely. The server refuses to enable it without "Power LED connected",
  and at run time it presses nothing if the LED cannot be read (phase `needsLed`).

| Setting | Default | Range |
| --- | --- | --- |
| `host`, `port`, `ups` | empty, 3493, empty | host name or IP, 1 to 65535, `[A-Za-z0-9._@-]{1,64}` |
| `username` | empty | same shape as the UPS name |
| `onBatteryMinutes` | 0 (off) | 0 to 1440 |
| `forceOffMinutes` | 0 (off) | 0 to 1440 |
| `powerOnRestore` | off | |
| `powerOnMinCharge` | 0 (any) | 0 to 100 |
| `powerOnDelaySeconds` | 60 | 10 to 3600 |

## What it does

The client polls with `LIST VAR` every 5 seconds on one connection, and takes `ups.status`,
`battery.charge` and `battery.runtime`.

- **Shutdown.** A short press when the host is on and the UPS reports `FSD` (forced shutdown) or
  `LB` (low battery), or has been `OB` (on battery) for `onBatteryMinutes`. One press per outage.
- **Long press.** If the host is still on `forceOffMinutes` after the shutdown press and the reason
  still holds, a 5 second press. Once. Mains returning cancels it.
- **Power on.** After an outage that shut the host down, or that began with it running, and
  with `powerOnRestore` on: when the UPS is `OL` with no `OB`, `LB` or `FSD`, the charge is at least
  `powerOnMinCharge`, the host is off, and that has held for `powerOnDelaySeconds`, a short press.
  The press is checked against the LED again at the moment of pressing. A host that is off by
  choice, one that never went down in the outage, and one that came back by itself are left alone.
- **Login.** With a user and a password the client registers as a client of the UPS (`LOGIN`). A
  primary `upsmon` waits for its secondaries to log out, up to `HOSTSYNC`, before it cuts the UPS
  output. The client logs out once the host's LED is off, and does not log in again until the
  shutdown's reason is gone. Without credentials it only reads, and the primary does not wait.

## Power races and unreachable servers

- **Mains returns before the UPS cuts its output.** The host is down, the UPS never lost power,
  and "restore on AC power" never fires. The client sees `OL` with the host off, and powers on
  after the delay. The delay also lets a UPS that is about to cut its output do so first.
- **The server is unreachable.** Three polls in a row fail: the status is `unreachable` and an
  event is logged. The client decides nothing new from a UPS it cannot see: no shutdown, no
  power-on. A shutdown that is under way keeps its timers, so the long press still happens.
- **The NUT server is the host.** It goes down with the host. The outage was seen (`OB` with the
  host on), so when the server answers again with `OL` and the host is off, the host is powered
  on. If the host does not start by itself, nothing powers it on; the board cannot know that mains
  is back. A host powered by the same UPS as the KVM needs the KVM on a separate supply.
- **Shutdown requested, host never goes off.** After 15 minutes the client stops expecting it to
  go off, so a host the owner switches off by hand later is not switched on.

## API

| Route | Role | Content |
| --- | --- | --- |
| `GET /api/nut/settings` | admin | The settings, and whether a password is stored |
| `POST /api/nut/settings` | admin | Save them. `password` sets one, `clearPassword` removes it, no field keeps it |
| `GET /api/nut/state` | any login | Status, UPS state, flags, charge, runtime, time on battery, login, host LED, phase, the last 20 events |

Registered in the #109 contract as `nut-client`.

## Web UI

A NUT tab under Settings > Integrations, for administrators and only on a server with
`nut-client`: the switch, the settings, the state and the events. The status bar gets a UPS item
(state, charge, runtime) while the client is on.

## Not done

Events go to the log and to the page. The audit log (#117) and notifications (#118) do not exist
yet; the monitor's events are the hook for them. STARTTLS to the NUT server is not implemented.
