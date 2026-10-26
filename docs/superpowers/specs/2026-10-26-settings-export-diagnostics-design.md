# Settings export and diagnostics bundle

Issue: yuzi-co/ironkvm-dist #122.

## Goal

Two support features for an administrator:

1. Export the settings of a board to a file, and import that file on the same board or on another
   one. The file holds no secret. The import shows what changes before it writes anything.
2. Download a diagnostics bundle: one archive that holds what a support request needs, with every
   secret removed.

## Decisions

- **An allowlist, not a blocklist.** The export is built from structs in `service/settingsio` that
  copy named fields out of `server.yaml` and out of named files in `/etc/kvm`. A field that a later
  change adds to `config.NUT`, for example, does not reach the file until somebody adds it to the
  export struct. A new secret is therefore excluded by default.
- **Network settings are not imported, and not exported.** A wrong address, port, protocol or
  allowlist cuts the administrator off from the board, and a dialog that asks "are you sure" does
  not help after the page stops answering. The import therefore changes nothing that decides how
  the board is reached. The design leaves a confirmed, self-reverting network import for later.
- **The import applies at the next reboot.** The import writes the files. It does not restart the
  services one by one, because each service applies its settings in its own way and several read
  them only at start. After a successful import the page offers a reboot. Until the reboot, the
  running services keep their old settings.
- **Diagnostics are redacted twice.** The configuration in the bundle comes from the same
  allowlist as the export. Then a scrubber passes over every file in the bundle. It replaces each
  secret value that it can read on the board, and each string that has the shape of a token or a
  key.
- **Admin only.** Every route is behind `middleware.CheckToken` and
  `middleware.RequireRole(authn.RoleAdmin)`, like the other settings pages.

## The settings file

The file is JSON. The browser downloads it as `ironkvm-settings-<host>-<date>.json`.

```json
{
  "format": "ironkvm-settings",
  "version": 1,
  "exportedAt": "2026-10-26T09:00:00Z",
  "source": { "application": "2.4.3", "hostname": "kvm" },
  "sections": {
    "watchdog": { "enabled": true, "timeoutMinutes": 5, "action": "reset", "...": "..." },
    "macros": { "macros": [] }
  }
}
```

- `format` must be `ironkvm-settings`.
- `version` is the version of the file format. It is 1. The server changes it when a section
  changes in a way that an older server cannot read. A new section does not change it.
- `source` is information only. The import does not read it.

### What is exported

| Section | Source | Fields |
| --- | --- | --- |
| `logging` | `server.yaml` `logger` | `level` |
| `loginSecurity` | `server.yaml` `security` | `loginLockoutDuration`, `loginMaxFailures` |
| `sessions` | `server.yaml` `jwt` | `refreshTokenDuration`, `revokeTokensOnLogout` |
| `powerLed` | `server.yaml` `hardware` | `powerLed` |
| `redfish` | `server.yaml` `redfish` | `enabled` |
| `ipmi` | `server.yaml` `ipmi` | `enabled` |
| `watchdog` | `server.yaml` `watchdog` | `enabled`, `timeoutMinutes`, `action`, `cooldownMinutes`, `maxPerHour`, `pingHost` |
| `nut` | `server.yaml` `nut` | every field; the password is in another file |
| `netboot` | `server.yaml` `netboot` | `usb`, `lan` |
| `vnc` | `server.yaml` `vnc` | `enabled`, `port`, `maxFps`, `vncAuth` |
| `webrtc` | `server.yaml` `webrtc`, `stun` | `udpPort`, `publicAddress`, `stun` |
| `picoclaw` | `server.yaml` `picoclaw` | `screenshotWidth`, `screenshotQuality` |
| `hostLog` | `server.yaml` `hostlog` | every field |
| `webTitle` | `/etc/kvm/web-title` | `title` |
| `leaderKey` | `/etc/kvm/leader-key` | `key` |
| `shortcuts` | `/etc/kvm/shortcuts.json` | `shortcuts` |
| `macros` | `/etc/kvm/macros.json` | `macros` |
| `mouseJiggler` | `/etc/kvm/mouse-jiggler` | `enabled`, `mode` |
| `keyJiggler` | `/etc/kvm/key-jiggler` | `enabled`, `key` |
| `oled` | `/etc/kvm/oled_sleep`, `/etc/kvm/oled_contrast` | `sleepSeconds`, `brightness` |
| `hdmi` | `/etc/kvm/hdmi_idle_timeout` | `idleTimeoutMinutes` |
| `roomMic` | `/etc/kvm/room-mic` | `allowed`, `gain` |
| `notifications` | `/etc/kvm/notify.json` | `enabled`, `snapshot`, `events`, `loginThreshold`, `loginWindowSeconds` |

A section whose file is absent exports the value that the owning service uses when the file is
absent.

### What is excluded, and why

| Excluded | Why |
| --- | --- |
| Accounts and password hashes (`/etc/kvm/pwd`), IPMI passwords sealed in it | Secret |
| `jwt.secretKey`, `/etc/kvm/.jwt_secret` | Secret: signs every session |
| TOTP key (`/etc/kvm/totp.key`) | Secret |
| API keys (`/etc/kvm/api_keys.json`), share links (`/etc/kvm/share_links.json`) | Secret |
| MCP key (`/etc/kvm/mcp.json`), PicoClaw internal token | Secret |
| First boot password | Secret |
| TLS certificate and private key (`server.*`, `uploaded.*`) | The key is secret, and a certificate without its key is of no use |
| Wi-Fi SSID and password | Secret, and network |
| VNC password, NUT password, IPMI key | Secret |
| TURN user and credential (`turn`) | Secret. The TURN address goes with them, so the whole block stays out |
| `proxy` | A proxy URL can hold a user and a password |
| Update server URL (`application-update.json`) | A URL can hold a user and a password |
| Notification channels: webhook URL, ntfy server, topic and token, Telegram bot and chat | Secret: a webhook URL or a public ntfy topic is the credential |
| Ports, `proto`, `http2`, `host`, `authentication`, certificate paths | Network: decides how the board is reached |
| Allowed networks (`security.allowedNetworks`) | Network: can lock the administrator out |
| Ethernet, DNS, IPv6, hostname, mDNS, USB network subnet, remote syslog | Network, or the name of this board |
| SSH: on or off, port, keys only, authorized keys, host keys | Access: can lock the administrator out |
| Redfish UUID, USB identity, hardware version | Identity of this board |
| CPU frequency, video clock, `GOMEMLIMIT`, swap, zram, ION reserve floor | Board tuning: a wrong value can stop the boot or the video |
| USB gadget functions (`/boot/usb.*`), `hid_only`, `disable_hid`, `hdmi_disable` | Board wiring: a wrong value can remove the keyboard or the video |
| Scripts and autostart (`/etc/kvm/scripts`, `/etc/kvm/autostart`) | Executable content. An import must never run code |
| Agent tiers, AI control mode, PicoClaw runtime | Security policy of the agent. An import must not widen it silently |
| Audit settings and audit log, Wake-on-LAN devices, video stream settings, input region | Not in version 1. They can be added as sections later |

### Import

1. The browser reads the file and sends it to `POST /api/settings/preview`. The body limit is
   8 MiB, which holds 50 macros at their largest.
2. The server checks `format` and `version`:
   - A different `format` is refused.
   - A `version` above 1 is refused: "this file comes from newer firmware (version N); update this
     board first".
   - A `version` below 1 is refused. Version 1 is the first.
3. For each section the server starts from the current value, decodes the section over it, and
   validates the result with the same limits as the settings page of that section.
   - A field that the file does not carry keeps its current value.
   - A field that the server does not know is ignored, and the preview lists it as a warning.
   - A section that the server does not know is skipped, and the preview lists it as unknown.
4. The server checks the sections together. The watchdog power cycle and the NUT client need the
   power LED connected. Plain VNC authentication needs a VNC password on this board. The VNC port
   and the WebRTC UDP port must not be a web server port.
5. The preview returns, for each section, its status (`change`, `same`, `invalid`, `unknown`), the
   changed fields with the old and the new value, and any error or warning.
6. The administrator chooses the sections and confirms. `POST /api/settings/import` repeats every
   check. If one selected section fails, nothing is written.
7. The server computes every new file in memory, writes each one to a temporary file in its own
   directory and syncs it, then renames them into place. If a rename fails, the server puts back
   the files it already renamed, from the copies it read before.
8. The page offers a reboot.

## Diagnostics bundle

`GET /api/diagnostics/bundle` returns `ironkvm-diagnostics-<host>-<date>.tar.gz`. Only one bundle
is built at a time. A second request while one is built gets an error.

| File | Content | Limit |
| --- | --- | --- |
| `README.txt` | What the bundle holds, when it was made, and the redaction rules | |
| `versions.txt` | Application version, server build, `/boot/ver`, `/boot/ironkvm.ver`, `/proc/version`, hardware version | |
| `slot.txt` | `/boot/slot`, `/boot/slot.try`, `/boot/slot.prev`, `/boot/slot.good`, `/proc/cmdline`, the root mount | |
| `dmesg.txt` | `dmesg` | last 256 KiB, 5 s |
| `server-log.txt` | `/tmp/nanokvm-server.log` | last 256 KiB |
| `config.json` | The export sections, the network fields of `server.yaml` that are not secret, and a list of the files in `/etc/kvm` with size and mode only | |
| `network.txt` | `ip addr`, `ip route`, `ip -6 route`, `/etc/resolv.conf`, `/proc/net/dev` | 5 s each |
| `usb-gadget.txt` | For each gadget in configfs: the UDC, the functions, the links in each configuration. For each UDC in `/sys/class/udc`: `state`, `current_speed` | |
| `video.json` | Failsafe state, ION carveout state, whether video and HDMI capture are off, the HDMI idle timeout | |
| `system.txt` | `/proc/meminfo`, `/proc/loadavg`, `/proc/uptime`, `/proc/mounts`, space on `/`, `/data`, `/tmp` and `/boot`, and each process with its name and resident memory | |
| `errors.txt` | Each source that failed or timed out | |

Each source has its own timeout and its own size limit, and the whole bundle has a limit of 30
seconds. A source that fails adds a line to `errors.txt`, and the bundle continues.

What the bundle does not touch:

- No `devmem`, nothing under `/sys/kernel/debug` except the ION heap dump that
  `service/ion` already reads, no thermal sensor, nothing that reads a hardware register.
  On this board a read of a clock-gated block can stop the bus.
- No call into `libkvm`. The first call to `common.GetKvmVision()` builds the capture pipeline, so
  the HDMI signal is not in the bundle.
- No process command lines. They can hold a key (`tailscale up --authkey=...`), so the process list
  is built from `/proc/<pid>/stat` and `/proc/<pid>/status`.
- No file content from `/etc/kvm`, only names, sizes and modes.

### Redaction

1. **Allowlist.** The configuration comes from the export structs and from a second struct that
   names the non-secret network fields of `server.yaml`.
2. **Known values.** The server reads every secret it can find on the board: the JWT secret, the
   TURN credential, the proxy URL, the update server URL, the Wi-Fi password, the VNC, NUT and
   IPMI secrets, the TOTP key, the MCP key, the PicoClaw internal token, the first boot password,
   the API key file, the share links file, and the notification URL, topic and tokens. Each value
   of 6 characters or more is replaced with `[REDACTED]` wherever it occurs in the bundle.
3. **Patterns.** PEM private key blocks, JSON web tokens, `Bearer` tokens, URL user information,
   `password=`, `token=`, `secret=`, `key=` and similar assignments, Telegram bot tokens, API keys
   with the `nkvm_` prefix, MCP keys with the `nag_mcp_` prefix, and Tailscale keys with the
   `tskey-` prefix.

The tests seed every secret on a fake board with a sentinel value, build the export and the
bundle, and search the output for each sentinel.

## Routes

| Route | What it does |
| --- | --- |
| `GET /api/settings/export` | The settings file, as an attachment |
| `POST /api/settings/preview` | The preview of a file |
| `POST /api/settings/import` | The import of the chosen sections |
| `GET /api/diagnostics/bundle` | The diagnostics bundle |

The server declares the features `settings-backup` and `diagnostics-bundle`. The UI shows
Settings > Backup only to an administrator, and each part only when the server declares it.

## Checks on the board

The tests run on a workstation against a fake root. These paths exist only on the board and need
a check there:

- `/tmp/nanokvm-server.log`, `/boot/slot*`, `/proc/cmdline`.
- `dmesg` and `ip` from BusyBox: the options that the bundle uses, and the time they take.
- `/sys/kernel/config/usb_gadget/*` and `/sys/class/udc/*`.
- `/sys/kernel/debug/ion/cvi_carveout_heap_dump` on slot A. Slot B has no ION, and the section
  must say so without an error.
- An import followed by a reboot: each section takes effect.
