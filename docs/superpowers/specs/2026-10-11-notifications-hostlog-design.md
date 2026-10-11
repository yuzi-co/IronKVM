# Notifications and the host log receiver

ironkvm-dist#118 and #116. Two services that share one event bus.

## Event bus (`server/service/events`)

A leaf package, so any service can publish without importing the notifier. An
event has a type, a time, a title, a message, string fields and, optionally, a
JPEG the publisher already holds. Each subscriber has its own goroutine and a
64-event queue; a full queue drops its oldest event, so `Publish` never blocks
the login handler or the ATX press.

| Type | Published by |
| --- | --- |
| `signal.lost`, `signal.restored` | the notifier's own monitor: it reads the HDMI signal every 5 s and reports a change that held for two reads |
| `watchdog.action` | `service/watchdog`, after a press, with the screenshot taken before it |
| `login.failure` (raw) | `auth.RecordLoginFailure`, which every login path counts through, the lockout being off or on |
| `login.failures` | the notifier, when `loginThreshold` raw failures fall inside `loginWindowSeconds` |
| `atx.press` | `vm.PressButton`, the one choke point of the UI, Redfish, IPMI and sequences. The watchdog presses through `PressButtonQuiet` and reports itself |
| `failsafe.entered` | the router, once, when the server starts in failsafe mode (`failsafe.Init` runs before anything listens) |
| `hostlog.match` | `service/hostlog`, for a line that matches an alert phrase |

## Notifications (`server/service/notify`)

- Settings live in `/etc/kvm/notify.json`, mode 0600, written by rename. They
  hold the tokens, so they are not in `server.yaml`, which is 0644. Unknown keys are ignored, so both slots can share it.
- Off by default. Every event type has a rule (on, cooldown); a missing rule is
  on with a default cooldown.
- Rate limit: a cooldown per event type. Events inside it are counted and the
  next notification carries `suppressed: N`.
- Channels: a webhook (JSON; the snapshot is added as base64 only when asked),
  ntfy (the JPEG is the attachment, the text is in headers, non-ASCII as RFC
  2047), Telegram (`sendPhoto` with a caption, else `sendMessage`).
- The snapshot goes through the same path as the watchdog's screenshot
  (`watchdogCapture` over the MCP snapshotter). It is taken once per event, only
  if a channel is on, and never for `signal.lost`.
- Secrets: the API says only whether a token is set. A save without a secret key
  keeps it; an empty string clears it. Transport errors are scrubbed
  (`url.Error` loses its URL, and every secret is masked), and a test holds the
  log free of all three.
- `POST /api/notify/test` sends on one channel or on every channel that is set
  up, ignoring the switches and cooldowns.
- Outbound calls use `utils.OutboundClient`, so the configured proxy applies.

## Host log (`server/service/hostlog`)

- Settings in `server.yaml` under `hostlog` (no secrets). Syslog (default 514)
  and netconsole (default 6666) are separate listeners, both off by default.
- `bind` is empty (every interface) or `usb` (the board's address on the USB
  link, bound once the link is up; the service checks every 10 s).
- Source filter: the owner's list if set; else the host on the USB link while it
  is up; else the watchdog's ping host; else any sender. The filter in force is
  shown on the page.
- Parsers: RFC 5424 (structured data skipped), RFC 3164, plain and extended
  netconsole. Control characters are dropped and a line is cut at 2 KB.
- Bounds: the in-memory ring is capped in bytes (64 KB to 4 MB, default 1 MB).
  The optional disk log (off by default, `/data/ironkvm/hostlog`) is two files
  that rotate at half the cap (256 KB to 64 MB, default 4 MB) and is read back
  on start. A token bucket (500 lines/s, burst 1000) protects the board from a
  panic loop; dropped lines are counted.
- The page opens `/api/hostlog/stream`, a websocket guarded by the session
  cookie like the terminal: the newest lines for the search, then batches of new
  ones. Search is whole words in any case over host, tag and message, with a
  severity cut. `/api/hostlog/download` returns the filtered lines as text.
- Alert phrases are Go regular expressions (RE2, so no catastrophic
  backtracking), case-insensitive. At most one `hostlog.match` event per
  second leaves the receiver.

## Contract (#109)

`notifications` (`/api/notify`) and `host-log` (`/api/hostlog`), in
`server/common/apicontract/contract.go` and `web/src/lib/server-contract.ts`.
The settings tabs appear only when the server advertises them.

## Not done

Screen-blank, on-screen phrase (OCR) and power-LED-change triggers from #118,
and forwarding the host log to the remote syslog from #116.
