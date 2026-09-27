# NetBird beside Tailscale, and one improved page for both

Issue: yuzi-co/ironkvm-dist #1. Decided with the owner on 2026-09-27: a NetBird page like the
Tailscale one, six improvements to both, and one service at a time.

## Goal

The KVM joins a NetBird network from the web UI, as it joins a tailnet today. Both services
get the same page, which also shows what the Tailscale page lacks: start at boot, peers,
version and updates, the daemon's memory against its group's limit, and the reason for a
failure.

## Measured, on image w (slot b, 2026-09-27)

Alpine's `netbird` 0.78.2 for riscv64, joined with a setup key, no peers, uses the kernel
WireGuard interface. After about 15 minutes idle, in the `addons` group beside tailscaled:

- netbird: 31.8 MB resident, 6.7 MB swapped. tailscaled: 23.9 MB resident, 2.1 MB swapped.
- `addons` at 61.6 MB against `memory.high` 64M: 354 `high` events, no `oom_kill`.
- MemAvailable 104 MB with both, 121 MB after netbird stopped.
- netbird's CPU while idle: about 0.3% of the core.

Not measured: memory and CPU with a peer streaming over NetBird. The owner has no peer yet.

## Decisions

- **One at a time.** Both together hold the `addons` group at its limit, idle. Starting,
  enabling at boot, or installing one while the other runs or is enabled at boot is refused
  with a message that names the other. The page shows that notice before the button is pressed.
- **NetBird comes from Alpine.** Upstream publishes no riscv64 build; Alpine v3.24 community
  does. The server runs `apk fetch` into `/data`, which checks Alpine's signature, and extracts
  `usr/bin/netbird`. The binary lives in `/data/ironkvm/addons/netbird/` and is linked into
  `/usr/bin` through the existing add-on records, so a slot install keeps it. No package is
  installed into the slot's root filesystem.
- **Identity on `/data`.** NetBird's config and state go to `/data/identity-system/netbird/`,
  beside Tailscale's, so a slot install keeps the peer.
- **The memory limit is derived, not typed.** Both init scripts set `GOMEMLIMIT` to 7/8 of the
  `addons` `memory.high`, as `S98tailscaled` already does. The free-text memory limit field goes;
  `/etc/kvm/GOMEMLIMIT` still caps it if an operator wrote one.
- **Logout for NetBird is `netbird deregister`,** which removes the peer from the account, as
  Tailscale's logout expires the node. The page says so before it runs.

## Server

A shared shape for both services, in `proto/vpn.go`:

```go
type VpnStatus struct {
    State       string    // notInstall, notRunning, notLogin, stopped, running
    Version     string
    IP          string
    Name        string    // host name or FQDN
    Account     string    // tailnet name; NetBird management URL host
    Control     bool      // connected to the coordination or management server
    Peers       []VpnPeer
    UptimeSec   int64     // of the daemon process, 0 when not running
    BootEnabled bool
    Memory      VpnMemory
    BlockedBy   string    // "tailscale" or "netbird" when the other one runs or starts at boot
}
type VpnPeer struct{ Name, IP string; Online bool }
type VpnMemory struct{ DaemonRSS, GroupCurrent, GroupHigh, GroupMax int64 }
```

`GetTailscaleStatusRsp` keeps its four fields for old clients and gains the rest.

Routes, under `/api/extensions/<name>/` with the admin group, `<name>` being `tailscale` or
`netbird`:

| Route | Change |
| --- | --- |
| `GET status` | returns `VpnStatus` |
| `POST install`, `uninstall`, `start`, `stop`, `restart`, `up`, `down`, `logout` | as Tailscale's today. `start` no longer changes start at boot. |
| `POST login` | Tailscale as today. NetBird: `{setupKey}` runs `netbird up --setup-key`; an empty body runs the SSO login and returns its URL, read from the CLI's output as Tailscale's is. |
| `POST boot` | new: `{enabled}` sets start at boot |
| `GET update` | new: `{current, latest}`. Tailscale: the latest version from `pkgs.tailscale.com`. NetBird: `apk update` then the candidate version. Cached for an hour. |
| `POST update` | new: installs the latest, keeps the state, restarts the daemon if it ran |

Start at boot is a separate switch now. Today `start` turns it on and `stop` turns it off; after
this change they only start and stop.

Errors: every failing handler returns the last lines of the CLI's output as the message, so the
page can show why. The existing error codes stay.

The exclusivity check is one function in `service/extensions/addon`, which both services call
in `install`, `start`, `up` and `boot`.

`service/extensions/netbird` follows `tailscale`'s files: `service.go`, `cli.go`, `install.go`.
Shared helpers move to `service/extensions/vpn`: the daemon's uptime and RSS from `/proc`, the
group's memory from `/sys/fs/cgroup/addons`, the login URL reader, and the output tail for
errors.

## Init script: `kvmapp/system/init.d/S98netbird`

Modeled on `S98tailscaled`: the pid file, the `addons` group, the derived `GOMEMLIMIT`, and
environment overrides for the tests. It runs `netbird service run --config
/data/identity-system/netbird/config.json --daemon-addr unix:///var/run/netbird.sock
--log-file /var/log/netbird.log`, and falls back to `/var/lib/netbird` when `/data` is absent.
Like `S98tailscaled`, it is copied to `/etc/init.d` when the add-on is enabled, and
`S04addons` handles it at boot with no change.

## Web

`web/src/pages/desktop/menu/settings/vpn/` holds the shared parts:

- the page frame, which switches on `state` as the Tailscale page does
- the header with restart, stop, uninstall and update
- start at boot, a switch
- status: control connection, IP, name, account, version, uptime
- peers, a list with online state
- memory: the daemon's RSS and the group's use against `memory.high` and `memory.max`, as bars
- the exclusivity notice
- the error detail under any failed action

`settings/tailscale/index.tsx` and `settings/netbird/index.tsx` each give the frame their API
module and their login form. NetBird's login form has the setup key field and an SSO button.
The swap control stays on both pages. A NetBird entry joins Settings below Tailscale, with its
own icon. New strings in English only; the other locales fall back to English.

## Testing

- Go: status parsing for both CLIs from recorded JSON (the NetBird one from the board), peers,
  the exclusivity check in each entry point, the boot switch, the update check with a fake
  HTTP server and a fake `apk`, the error tail, and the login URL reader for NetBird's output.
- Shell: `tools/service/test-netbird.sh` runs `S98netbird` against a scratch tree with a stub
  daemon, as `test-tailscaled.sh` does; the memlimit and cgroup tests cover the new script.
- On the board: install NetBird from the page, join with the setup key, check the IP and that
  the identity survives a server restart; the notice blocks Tailscale while NetBird runs; the
  boot switch survives a reboot; update reports current equal to latest.

## Out of scope

NetBird networks and routes, NetBird SSH, exit nodes, self-hosted management URLs (the field
can come later), and both services at once.
