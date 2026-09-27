# NetBird Beside Tailscale Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The KVM joins a NetBird network from the web UI as it joins a tailnet today, and both services share one improved settings page that also shows start at boot, peers, version and updates, the daemon's memory against its group, and the reason for a failure.

**Architecture:** A new `service/extensions/vpn` package holds what both add-ons share (error tails, the init-script runner, the login URL reader, `/proc` and cgroup readers, the update cache, the boot and refusal handlers). `service/extensions/addon` gains the one exclusivity check and the start-at-boot record. `service/extensions/netbird` mirrors `tailscale` (service.go, cli.go, install.go, status.go) and installs Alpine's package with `apk fetch` into `/data`. `S98netbird` is modeled on `S98tailscaled`. The web gets `settings/vpn/` for the shared frame, and `settings/tailscale/` and `settings/netbird/` pass it their API module and login form.

**Tech Stack:** Go 1.25 (gin, logrus), POSIX sh init scripts, React + TypeScript + antd 6 + i18next (Vite), Alpine apk-tools 3.0.8 on the board.

**Spec:** `docs/superpowers/specs/2026-09-27-vpn-extensions-design.md` (on `fork/integration`, commit 38f51392). Issue: yuzi-co/ironkvm-dist #1.

## Global Constraints

- Branch `feat/netbird` off `fork/integration` in `D:\projects\NanoKVM`; the dist change goes on `feat/netbird` off `main` in `D:\projects\ironkvm-dist`.
- Commits carry NO `Co-Authored-By`, `Claude-Session` or any other assistant attribution line.
- The owner's NetBird setup key is given at run time. Never write it into any file, test, fixture, shell history file or commit. This plan calls it "the setup key the owner gave".
- Tailscale is in live use on the board (root@10.0.0.222). Stop it only with the owner's explicit go-ahead, and restore it (running, and start at boot as it was) before the task ends.
- One at a time: install, start, up, login and boot-on of one service are refused while the other runs or is enabled at boot, with a message that names the other.
- NetBird comes from Alpine v3.24 community (`netbird` 0.78.2-r0, binary `/usr/bin/netbird`, 43 MB). The server runs `apk fetch` into `/data`, checks the signature, extracts `usr/bin/netbird` into `/data/ironkvm/addons/netbird/`, and links it into `/usr/bin` through the add-on records. No package is installed into the slot's root filesystem.
- NetBird's config and state live in `/data/identity-system/netbird/`, falling back to `/var/lib/netbird` when `/data` is absent.
- `GOMEMLIMIT` is derived in both init scripts: 7/8 of the `addons` group's `memory.high`; `/etc/kvm/GOMEMLIMIT` still caps it if an operator wrote one. The server no longer writes or deletes that file from the VPN handlers, and the page's memory switch goes.
- NetBird logout is `netbird deregister`; the page warns that it removes the peer from the account before it runs.
- Start at boot is its own `POST boot` route with `{enabled}`. `start` and `stop` no longer change it.
- `GET update` returns `{current, latest}`, cached for one hour; `POST update` installs the latest, keeps the state, and restarts the daemon only if it ran.
- Every failing handler returns the last lines of the CLI's output as its message. The existing error codes (-1, -2) stay.
- New UI strings in English only (`web/src/i18n/locales/en.ts`); other locales fall back to English. The swap control stays on both pages.
- Out of scope: NetBird networks, routes, SSH, exit nodes, self-hosted management URLs, both services at once.
- `gotest ARGS` in this plan means exactly:
  `cd /d/projects/NanoKVM/server && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM\\server:/src" -v nanokvm-gomod:/go/pkg/mod -v nanokvm-gocache:/root/.cache/go-build -w /src golang:1.25 go test ARGS`
- Before every Go commit, format the touched files: `cd /d/projects/NanoKVM/server && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM\\server:/src" -w /src golang:1.25 gofmt -w <files>`. The code blocks in this plan are written to gofmt's layout, but column alignment in composite literals is gofmt's call.
- Shell tests hang under Git Bash. Run them in a container:
  `cd /d/projects/NanoKVM && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM:/r" -w /r alpine:3.24 sh tools/service/<suite>.sh`
- Web: pnpm is not installed; use npm from `web/`: `npm run build`, `npm run lint`. Never run prettier over the whole tree (it rewrites about 60 unrelated files). Format only touched files: `npx prettier --write <file> ...`.
- Board build of the server:
  `cd /d/projects/NanoKVM && MSYS_NO_PATHCONV=1 docker run -e UID=1000 -e GID=1000 -v "$(pwd -W):/home/build/NanoKVM" --rm nanokvm-builder-local-1000-1000 /bin/bash -c 'cd /home/build/NanoKVM/server && CGO_ENABLED=1 GOOS=linux GOARCH=riscv64 CC=riscv64-unknown-linux-musl-gcc CGO_CFLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d" go build -buildvcs=false -ldflags "-s -w" && patchelf --add-rpath "\$ORIGIN/dl_lib" NanoKVM-Server'`
- Server deploy: `scp server/NanoKVM-Server root@10.0.0.222:/data/NanoKVM-Server.new && scp tools/deploy/deploy-server root@10.0.0.222:/data/ && ssh root@10.0.0.222 'DEPLOY_TIMEOUT=240 sh /data/deploy-server /data/NanoKVM-Server.new'`. Stage on `/data`, never `/tmp` (a 79 MB tmpfs).
- Web deploy (AGENTS.md: "rename web/dist to web and upload to /kvmapp/server/"): `scp -r web/dist root@10.0.0.222:/kvmapp/server/web.new && ssh root@10.0.0.222 'rm -rf /kvmapp/server/web.old && mv /kvmapp/server/web /kvmapp/server/web.old && mv /kvmapp/server/web.new /kvmapp/server/web'`, then the server deploy (which restarts the server) or `/etc/init.d/S95nanokvm restart`.

## Review Focus

1. A setup key pasted with spaces or a trailing newline, or a CLI that echoes the key in its error: the key is trimmed before use and never appears in a response message or the log (Task 11 `TestSetupKeyNeverReachesTheMessage`, Task 12 `TestLoginTrimsTheSetupKey`).
2. A board with no `apk` (vendor firmware) or no `/data`: NetBird install fails at once with a message that says it needs an IronKVM image, and nothing is left behind (Task 10 `TestInstallWithoutApkSaysWhy`).
3. A stale pid file after a power cut that names a pid another process now holds: it does not count as the other VPN running, so it neither blocks nor gets signalled (Task 5 `TestPIDNeedsTheDaemonsCommandLine`).
4. `netbird status` hanging while the management server is unreachable: the page's status request returns "notRunning" within seconds instead of hanging the handler (Task 12 `TestStatusDoesNotHangOnAStuckCLI`).
5. Update pressed while the daemon is stopped: it stays stopped afterwards; pressed while it runs, it runs again afterwards (Task 9 `TestUpdateRestartsOnlyADaemonThatRan`, Task 12 `TestUpdateKeepsAStoppedDaemonStopped`).

---

## File Structure

Server (`D:\projects\NanoKVM\server`):

| File | Change | Responsibility |
| --- | --- | --- |
| `proto/vpn.go` | create | `VpnState`, `VpnStatus`, `VpnPeer`, `VpnMemory`, `VpnBootReq`, `VpnUpdateRsp`, `VpnLoginRsp`, `NetbirdLoginReq` |
| `proto/vpn_test.go` | create | the JSON names old clients read |
| `proto/tailscale.go` | modify | Tailscale names become aliases of the VPN ones |
| `service/extensions/addon/exclusive.go` | create | `Daemon`, `Tailscale`, `NetBird`, `PID`, `Running`, `BootEnabled`, `SetBoot`, `BlockedBy`, `CheckExclusive`, `BlockedError` |
| `service/extensions/addon/exclusive_test.go` | create | the check and the boot record |
| `service/extensions/vpn/tail.go` | create | `Tail`, `CmdError`, `Run`, `Message`, `Script` |
| `service/extensions/vpn/login.go` | create | `LoginURL`, `ReadLoginURL` (moved from tailscale/cli.go) |
| `service/extensions/vpn/proc.go` | create | `UptimeSec`, `RSS`, `GroupMemory`, `Fill`, `CgroupDir` |
| `service/extensions/vpn/version.go` | create | `VersionCache`, `UpdateTTL` |
| `service/extensions/vpn/handlers.go` | create | `Refuse`, `Boot` |
| `service/extensions/vpn/*_test.go` | create | one per file above |
| `service/extensions/tailscale/status.go` | create | `TsStatus`, `TsPeer`, `StateMap`, `parseStatus`, `toVpnStatus` |
| `service/extensions/tailscale/cli.go` | modify | script-based start/stop, CLI error tails, `Version` |
| `service/extensions/tailscale/service.go` | modify | refusal, boot, update, no GOMEMLIMIT writes, `VpnStatus` |
| `service/extensions/tailscale/install.go` | modify | `resolveRedirect`, `versionFromPackageURL`, `latestVersion` |
| `service/extensions/tailscale/cli_test.go` | delete | moved to `vpn/login_test.go` |
| `service/extensions/tailscale/addon_test.go` | modify | drop `TestRecordEnabledOnlyOnData` |
| `service/extensions/tailscale/{status,service,update}_test.go`, `testdata/status-running.json` | create | |
| `service/extensions/netbird/{service,cli,status,install}.go` | create | the NetBird add-on |
| `service/extensions/netbird/{helpers,status,cli,install,service}_test.go`, `testdata/*.json` | create | |
| `router/extensions.go` | modify | boot, update and the NetBird routes |

Scripts and tools (`D:\projects\NanoKVM`):

| File | Change |
| --- | --- |
| `kvmapp/system/init.d/S98netbird` | create, mode 755 |
| `kvmapp/system/init.d/S98tailscaled` | modify: `tailscale_memlimit_mib` becomes `memlimit_mib` |
| `kvmapp/system/init.d.package-only` | modify: `S98netbird not-installed-by-policy` |
| `tools/service/test-netbird.sh` | create, mode 755 |
| `tools/service/test-tailscale-memlimit.sh` | rename to `test-addon-memlimit.sh`, covers both scripts |
| `tools/service/test-cgroup-join.sh` | modify: add `S98netbird:addons` |

Web (`D:\projects\NanoKVM\web\src`):

| File | Change |
| --- | --- |
| `api/extensions/tailscale.ts` | modify: `setBoot`, `getUpdate`, `update`, long timeouts |
| `api/extensions/netbird.ts` | create |
| `pages/desktop/menu/settings/vpn/{types.ts,format.ts,page.tsx,header.tsx,boot.tsx,details.tsx,peers.tsx,memory.tsx,notice.tsx,error-detail.tsx,install.tsx,run.tsx,device.tsx,uninstall.tsx,swap.tsx}` | create |
| `pages/desktop/menu/settings/tailscale/index.tsx`, `login.tsx` | modify |
| `pages/desktop/menu/settings/tailscale/install-help.tsx` | create |
| `pages/desktop/menu/settings/tailscale/{header,install,run,device,uninstall,memory,swap,types}.ts(x)` | delete |
| `pages/desktop/menu/settings/netbird/{index,login}.tsx` | create |
| `pages/desktop/menu/settings/index.tsx` | modify: NetBird tab below Tailscale |
| `components/icons/netbird.tsx`, `assets/images/netbird.svg` | create |
| `i18n/locales/en.ts` | modify: `settings.vpn`, `settings.netbird`, trimmed `settings.tailscale` |

Dist (`D:\projects\ironkvm-dist`):

| File | Change |
| --- | --- |
| `devices/kernel-requirements` | add `netbird software CONFIG_TUN CONFIG_WIREGUARD` |
| `devices/fake-board/kernel.config` | add `CONFIG_WIREGUARD=y`, comment says five options |
| `devices/memory-profiles` | comment: NetBird, like Tailscale, needs no entry |
| `generic/S01cgroups` | comment: the addons group holds NetBird too |

`generic/S04addons` needs no change: it reads any add-on's `initd` and `enabled` files.

---

### Task 1: Branch, and the facts the parsers need

No production code. This task fixes the fixtures and the apk method before any parser or installer is written.

**Files:**
- Create: `server/service/extensions/netbird/testdata/status-connected.json`
- Create: `server/service/extensions/netbird/testdata/status-peers.json`
- Create: `server/service/extensions/netbird/testdata/status-needslogin.json`
- Create: `server/service/extensions/tailscale/testdata/status-running.json`

**Interfaces:**
- Consumes: nothing.
- Produces: the four fixtures above (read by Tasks 7, 11, 12); the apk extraction decision (A, B or C below) recorded in the commit message and applied in Task 10; the confirmed `daemonStatus` values `Idle`, `Connecting`, `Connected`, `NeedsLogin`, `LoginFailed`, `SessionExpired` (used in Task 11).

- [ ] **Step 1: Create the branch and commit this plan**

```bash
cd /d/projects/NanoKVM
git status --short            # must be empty apart from this plan file
git switch -c feat/netbird fork/integration
git add docs/superpowers/plans/2026-09-27-vpn-extensions.md
git commit -m "plans: NetBird beside Tailscale, one improved page for both (#1)"
```

- [ ] **Step 2: Capture the status JSON shape from the NetBird source at v0.78.2**

```bash
cd "$TMPDIR" 2>/dev/null || cd /tmp
curl -fsSL https://raw.githubusercontent.com/netbirdio/netbird/v0.78.2/client/status/status.go -o nb-status.go
curl -fsSL https://raw.githubusercontent.com/netbirdio/netbird/v0.78.2/client/internal/peer/conn_status.go -o nb-conn.go
grep -n 'type PeerStateDetailOutput' -A 17 nb-status.go
grep -n 'DaemonStatus[A-Za-z]* *DaemonStatus = ' nb-status.go
grep -n 'return "' nb-conn.go
```

Expected (confirmed while writing this plan):
- `PeerStateDetailOutput` has `json:"fqdn"`, `json:"netbirdIp"`, `json:"status"` (plus publicKey, lastStatusUpdate, connectionType, iceCandidateType, iceCandidateEndpoint, relayAddress, lastWireguardHandshake, transferReceived, transferSent, latency, quantumResistance, networks).
- Daemon statuses: `Idle`, `Connecting`, `Connected`, `NeedsLogin`, `LoginFailed`, `SessionExpired`.
- Peer `status` values: `Idle`, `Connecting`, `Connected`.
- `client/cmd/status.go` prints JSON even in NeedsLogin when `--json` is given (the human-readable "Run UP command" text is only for the non-JSON output).

If any tag differs, use the source's tag in the fixtures below and in Task 11's `NbPeer` and `NbStatus` tags.

- [ ] **Step 3: Write the connected fixture (the board sample, verbatim)**

`server/service/extensions/netbird/testdata/status-connected.json`:

```json
{"peers":{"total":0,"connected":0,"details":null},"cliVersion":"0.78.2","daemonVersion":"0.78.2","daemonStatus":"Connected","management":{"url":"https://api.netbird.io:443","connected":true,"error":""},"signal":{"url":"https://signal.netbird.io:443","connected":true,"error":""},"netbirdIp":"100.73.212.105/16","fqdn":"ironkvm.netbird.cloud","usesKernelInterface":true}
```

- [ ] **Step 4: Write the peers fixture, built from the v0.78.2 struct**

`server/service/extensions/netbird/testdata/status-peers.json`:

```json
{"peers":{"total":2,"connected":1,"details":[{"fqdn":"phone.netbird.cloud","netbirdIp":"100.73.10.21","publicKey":"cGhvbmUtcHVibGljLWtleS0wMDAwMDAwMDAwMDAwMDA=","status":"Idle","lastStatusUpdate":"2026-09-27T09:00:00Z","connectionType":"-","iceCandidateType":{"local":"","remote":""},"iceCandidateEndpoint":{"local":"","remote":""},"relayAddress":"","lastWireguardHandshake":"0001-01-01T00:00:00Z","transferReceived":0,"transferSent":0,"latency":0,"quantumResistance":false,"networks":[]},{"fqdn":"laptop.netbird.cloud","netbirdIp":"100.73.10.20","publicKey":"bGFwdG9wLXB1YmxpYy1rZXktMDAwMDAwMDAwMDAwMDA=","status":"Connected","lastStatusUpdate":"2026-09-27T10:00:00Z","connectionType":"P2P","iceCandidateType":{"local":"host","remote":"srflx"},"iceCandidateEndpoint":{"local":"192.168.1.5:51820","remote":"203.0.113.7:51820"},"relayAddress":"","lastWireguardHandshake":"2026-09-27T10:00:05Z","transferReceived":1024,"transferSent":2048,"latency":12000000,"quantumResistance":false,"networks":[]}]},"cliVersion":"0.78.2","daemonVersion":"0.78.2","daemonStatus":"Connected","management":{"url":"https://api.netbird.io:443","connected":true,"error":""},"signal":{"url":"https://signal.netbird.io:443","connected":true,"error":""},"netbirdIp":"100.73.212.105/16","fqdn":"ironkvm.netbird.cloud","usesKernelInterface":true}
```

- [ ] **Step 5: Probe apk on the board (read-only for Tailscale)**

This does not touch Tailscale. Everything is staged on `/data` because `/tmp` is a 79 MB tmpfs.

```bash
ssh root@10.0.0.222 'set -x
apk --version
apk update >/dev/null && apk search -e netbird; echo search_e=$?
rm -rf /data/nbprobe && mkdir -p /data/nbprobe && cd /data/nbprobe
apk fetch -o /data/nbprobe netbird; echo fetch=$?; ls -l /data/nbprobe
PKG=$(ls /data/nbprobe/netbird-[0-9]*.apk)
apk verify "$PKG"; echo verify=$?
mkdir -p x && apk extract --no-chown --destination /data/nbprobe/x "$PKG"; echo extract=$?
ls -l x/usr/bin/netbird
tar -tzf "$PKG" 2>&1 | head -5; echo tar_list=$?
cp "$PKG" bad.apk && printf X | dd of=bad.apk bs=1 seek=4096 conv=notrunc 2>/dev/null
apk verify bad.apk; echo bad_verify=$?
mkdir -p y && apk extract --no-chown --destination /data/nbprobe/y bad.apk; echo bad_extract=$?'
```

Record the output. Expected: `search_e` prints `netbird-0.78.2-r0`; `fetch=0`; one `netbird-0.78.2-r0.apk`.

Pick the extraction method from the result, and write the letter in the Step 8 commit message:
- **A** `verify=0`, `extract=0`, `x/usr/bin/netbird` is about 43 MB, and `bad_verify` or `bad_extract` is non-zero: keep Task 10's `extractPackage` and `install` exactly as written.
- **B** `extract` is non-zero (unknown applet, or a chown or permission error on exFAT) but `tar_list` shows `usr/bin/netbird` and `bad_verify` is non-zero: in Task 10 use the tar variant of `extractPackage` given there.
- **C** `apk verify` does not exist (`verify` prints a usage or "unknown command") but `bad_extract` is non-zero: in Task 10 delete the `apk("verify", pkg)` call from `install`, and the fake apk still answers `verify`.
- If neither `bad_verify` nor `bad_extract` is non-zero, nothing checks the signature: stop and ask the owner before Task 10.

If `apk search -e` prints nothing but `apk search -x netbird` does, use `-x` in Task 10's `latestVersion` and fake apk.

- [ ] **Step 6: Capture a fresh daemon's status on the board, and where it keeps its files**

Still beside the running Tailscale; this daemon is not logged in, makes no interface, and runs for about ten seconds.

```bash
ssh root@10.0.0.222 'cd /data/nbprobe
NB=/data/nbprobe/x/usr/bin/netbird
[ -x "$NB" ] || { mkdir -p x && tar -xzf netbird-[0-9]*.apk -C x usr/bin/netbird; }
ls /var/lib/netbird 2>&1
NB_STATE_DIR=/data/nbprobe/state $NB service run --config /data/nbprobe/state/config.json --daemon-addr unix:///var/run/nbprobe.sock --log-file /data/nbprobe/probe.log </dev/null >/dev/null 2>&1 &
echo $! > /data/nbprobe/pid; sleep 10
$NB status --json --daemon-addr unix:///var/run/nbprobe.sock; echo; echo status_rc=$?
ls -la /data/nbprobe/state; ls -la /var/lib/netbird 2>&1
kill "$(cat /data/nbprobe/pid)"; sleep 3
$NB status --json --daemon-addr unix:///var/run/nbprobe.sock; echo down_rc=$?'
```

Expected: `daemonStatus` is `NeedsLogin`; `/data/nbprobe/state` holds `default.json` or `config.json`, `active_profile.json` and `state.json`; `/var/lib/netbird` gained nothing new; with the daemon gone, `status` exits non-zero (the page reads that as notRunning).

If files other than logs appeared under `/var/lib/netbird`, stop: `S98netbird` (Task 13) exports `NB_STATE_DIR` for exactly this, and a gap means the identity would stay on the slot. Ask the owner before going on.

Save the JSON line as `server/service/extensions/netbird/testdata/status-needslogin.json`. If the capture could not be made, use this, which is what `ConvertToStatusOutputOverview` produces for a fresh profile:

```json
{"peers":{"total":0,"connected":0,"details":null},"cliVersion":"0.78.2","daemonVersion":"0.78.2","daemonStatus":"NeedsLogin","management":{"url":"https://api.netbird.io:443","connected":false,"error":""},"signal":{"url":"","connected":false,"error":""},"netbirdIp":"N/A","fqdn":"","usesKernelInterface":false}
```

Clean up:

```bash
ssh root@10.0.0.222 'rm -rf /data/nbprobe /var/run/nbprobe.sock'
```

- [ ] **Step 7: Check the Tailscale key names without copying the board's data**

```bash
ssh root@10.0.0.222 'tailscale status --json' | grep -o '"\(Version\|BackendState\|Self\|HostName\|DNSName\|TailscaleIPs\|Online\|CurrentTailnet\|Name\|Peer\)"' | sort | uniq -c
```

Expected: every one of the ten keys appears at least once. The fixture below uses these names with made-up values; the board's real names and addresses are not written anywhere.

`server/service/extensions/tailscale/testdata/status-running.json`:

```json
{
  "Version": "1.90.1-t8b5c4a1e2-g3f6d7c8b9",
  "BackendState": "Running",
  "Self": {
    "HostName": "nanokvm",
    "DNSName": "ironkvm.tail1234.ts.net.",
    "TailscaleIPs": ["100.101.102.103", "fd7a:115c:a1e0::1"],
    "Online": true
  },
  "CurrentTailnet": { "Name": "owner@example.com", "MagicDNSSuffix": "tail1234.ts.net" },
  "Peer": {
    "nodekey:bbb": {
      "HostName": "Phone",
      "DNSName": "phone.tail1234.ts.net.",
      "TailscaleIPs": ["100.64.0.3"],
      "Online": false
    },
    "nodekey:aaa": {
      "HostName": "LAPTOP-1",
      "DNSName": "laptop.tail1234.ts.net.",
      "TailscaleIPs": ["100.64.0.2", "fd7a:115c:a1e0::2"],
      "Online": true
    }
  }
}
```

- [ ] **Step 8: Commit the fixtures**

```bash
cd /d/projects/NanoKVM
git add server/service/extensions/netbird/testdata server/service/extensions/tailscale/testdata
git commit -m "extensions: status fixtures for NetBird and Tailscale

The NetBird peer shape and daemon statuses are from client/status at
v0.78.2. apk on the board: method <A|B|C>, <one line of the probe result>."
```

---

### Task 2: The shared status shape in proto

**Files:**
- Create: `server/proto/vpn.go`
- Modify: `server/proto/tailscale.go` (whole file)
- Test: `server/proto/vpn_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `proto.VpnState` with `VpnNotInstall`, `VpnNotRunning`, `VpnNotLogin`, `VpnStopped`, `VpnRunning`; `proto.VpnStatus{State, Version, IP, Name, Account, Control, Peers []VpnPeer, UptimeSec int64, BootEnabled, Memory VpnMemory, BlockedBy}`; `proto.VpnPeer{Name, IP, Online}`; `proto.VpnMemory{DaemonRSS, GroupCurrent, GroupHigh, GroupMax int64}`; `proto.VpnBootReq{Enabled}`; `proto.VpnUpdateRsp{Current, Latest}`; `proto.VpnLoginRsp{Url}`; `proto.NetbirdLoginReq{SetupKey}`. The Tailscale names stay as aliases.

- [ ] **Step 1: Write the failing test**

`server/proto/vpn_test.go`:

```go
package proto

import (
	"encoding/json"
	"strings"
	"testing"
)

// Old clients read state, name, ip and account from the Tailscale status.
// Those four names must not move when the shape grows.
func TestTailscaleStatusKeepsItsFourFields(t *testing.T) {
	b, err := json.Marshal(GetTailscaleStatusRsp{
		State:   TailscaleRunning,
		Name:    "kvm",
		IP:      "100.1.2.3",
		Account: "owner@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"state":"running"`, `"name":"kvm"`, `"ip":"100.1.2.3"`, `"account":"owner@example.com"`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s is missing from %s", want, b)
		}
	}
}

func TestVpnStatusFieldNames(t *testing.T) {
	b, err := json.Marshal(VpnStatus{
		State:       VpnStopped,
		Version:     "0.78.2",
		Control:     true,
		Peers:       []VpnPeer{{Name: "laptop", IP: "100.73.10.20", Online: true}},
		UptimeSec:   61,
		BootEnabled: true,
		Memory:      VpnMemory{DaemonRSS: 1, GroupCurrent: 2, GroupHigh: 3, GroupMax: 4},
		BlockedBy:   "tailscale",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"state":"stopped"`, `"version":"0.78.2"`, `"control":true`,
		`"peers":[{"name":"laptop","ip":"100.73.10.20","online":true}]`,
		`"uptimeSec":61`, `"bootEnabled":true`,
		`"memory":{"daemonRss":1,"groupCurrent":2,"groupHigh":3,"groupMax":4}`,
		`"blockedBy":"tailscale"`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("%s is missing from %s", want, b)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./proto/ -run 'TestTailscaleStatusKeepsItsFourFields|TestVpnStatusFieldNames' -v`
Expected: FAIL to compile, `undefined: VpnStatus`.

- [ ] **Step 3: Write the implementation**

`server/proto/vpn.go`:

```go
package proto

// VpnState is where a VPN add-on stands, as the settings page shows it.
type VpnState string

const (
	VpnNotInstall VpnState = "notInstall"
	VpnNotRunning VpnState = "notRunning"
	VpnNotLogin   VpnState = "notLogin"
	VpnStopped    VpnState = "stopped"
	VpnRunning    VpnState = "running"
)

// VpnStatus is the status of Tailscale or NetBird, in one shape for both.
type VpnStatus struct {
	State       VpnState  `json:"state"`
	Version     string    `json:"version"`
	IP          string    `json:"ip"`
	Name        string    `json:"name"`    // host name or FQDN
	Account     string    `json:"account"` // tailnet name; NetBird management URL host
	Control     bool      `json:"control"` // connected to the coordination or management server
	Peers       []VpnPeer `json:"peers"`
	UptimeSec   int64     `json:"uptimeSec"` // of the daemon process, 0 when not running
	BootEnabled bool      `json:"bootEnabled"`
	Memory      VpnMemory `json:"memory"`
	BlockedBy   string    `json:"blockedBy"` // "tailscale" or "netbird" when the other one runs or starts at boot
}

type VpnPeer struct {
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Online bool   `json:"online"`
}

// VpnMemory is in bytes. GroupHigh and GroupMax are 0 when the group sets no
// limit or does not exist.
type VpnMemory struct {
	DaemonRSS    int64 `json:"daemonRss"`
	GroupCurrent int64 `json:"groupCurrent"`
	GroupHigh    int64 `json:"groupHigh"`
	GroupMax     int64 `json:"groupMax"`
}

type VpnBootReq struct {
	Enabled bool `json:"enabled"`
}

type VpnUpdateRsp struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
}

type VpnLoginRsp struct {
	Url string `json:"url"`
}

// NetbirdLoginReq joins with a setup key. An empty key asks for SSO instead.
type NetbirdLoginReq struct {
	SetupKey string `json:"setupKey"`
}
```

`server/proto/tailscale.go` (whole file):

```go
package proto

// The Tailscale names predate NetBird. They stay, as aliases, for the code
// and the clients that use them.
type TailscaleState = VpnState

const (
	TailscaleNotInstall = VpnNotInstall
	TailscaleNotRunning = VpnNotRunning
	TailscaleNotLogin   = VpnNotLogin
	TailscaleStopped    = VpnStopped
	TailscaleRunning    = VpnRunning
)

// GetTailscaleStatusRsp keeps state, name, ip and account where old clients
// read them, and carries the rest of VpnStatus beside them.
type GetTailscaleStatusRsp = VpnStatus

type LoginTailscaleRsp = VpnLoginRsp
```

- [ ] **Step 4: Run the tests and the build**

Run: `gotest ./proto/ ./service/extensions/... -v -run 'TestTailscaleStatusKeepsItsFourFields|TestVpnStatusFieldNames|TestPlaceBinaries'`
Expected: PASS (the tailscale package still compiles against the aliases).

- [ ] **Step 5: Commit**

```bash
git add server/proto/vpn.go server/proto/vpn_test.go server/proto/tailscale.go
git commit -m "proto: one status shape for Tailscale and NetBird"
```

---

### Task 3: Error tails and the init-script runner

**Files:**
- Create: `server/service/extensions/vpn/tail.go`
- Test: `server/service/extensions/vpn/tail_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `vpn.TailLines = 8`; `vpn.Tail(out []byte, n int) string`; `type vpn.CmdError struct{ Err error; Tail string }` with `Error()` and `Unwrap()`; `vpn.Run(cmd *exec.Cmd) ([]byte, error)`; `vpn.Message(what string, err error) string`; `vpn.Script(path, action, logFile string) error`.

- [ ] **Step 1: Write the failing test**

`server/service/extensions/vpn/tail_test.go`:

```go
//go:build linux

package vpn

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTailKeepsTheLastNonBlankLines(t *testing.T) {
	got := Tail([]byte("one\n\n two \nthree\r\nfour\n"), 2)
	if got != "three\nfour" {
		t.Fatalf("got %q", got)
	}
}

func TestTailCapsItsSize(t *testing.T) {
	got := Tail([]byte(strings.Repeat("x", 5000)), TailLines)
	if len(got) != maxTailBytes {
		t.Fatalf("got %d bytes, want %d", len(got), maxTailBytes)
	}
}

func TestRunCarriesTheTail(t *testing.T) {
	_, err := Run(exec.Command("sh", "-c", "echo one; echo two; echo why >&2; exit 3"))
	var ce *CmdError
	if !errors.As(err, &ce) {
		t.Fatalf("want a *CmdError, got %v", err)
	}
	if ce.Tail != "one\ntwo\nwhy" {
		t.Fatalf("tail is %q", ce.Tail)
	}
	if got := Message("start failed", err); got != "start failed:\none\ntwo\nwhy" {
		t.Fatalf("message is %q", got)
	}
}

func TestMessageWithoutOutputUsesTheError(t *testing.T) {
	if got := Message("stop failed", errors.New("boom")); got != "stop failed: boom" {
		t.Fatalf("message is %q", got)
	}
}

func script(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "S98test")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// The boot scripts exit 0 when the daemon did not come up and print FAIL.
func TestScriptTakesAFailLineAsAFailedStart(t *testing.T) {
	s := script(t, `echo "GOMEMLIMIT set to 56MiB"; echo "Starting netbird: FAIL"`)
	err := Script(s, "start", "")
	if err == nil {
		t.Fatal("a start that printed FAIL must fail")
	}
	if msg := Message("start failed", err); !strings.Contains(msg, "Starting netbird: FAIL") {
		t.Fatalf("message is %q", msg)
	}
}

func TestScriptAddsTheDaemonsLog(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "netbird.log")
	if err := os.WriteFile(logFile, []byte(
		"2026-09-27T10:00:00Z INFO starting\n"+
			"2026-09-27T10:00:01Z FATA failed to create interface wt0: operation not permitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := script(t, `echo "Starting netbird: FAIL"`)
	msg := Message("start failed", Script(s, "start", logFile))
	if !strings.Contains(msg, "failed to create interface wt0") {
		t.Fatalf("the daemon's own reason is missing: %q", msg)
	}
}

func TestScriptStartThatSucceeds(t *testing.T) {
	if err := Script(script(t, `echo "Starting netbird: OK"`), "start", ""); err != nil {
		t.Fatal(err)
	}
}

// Stop with nothing running prints FAIL and is still a successful stop.
func TestScriptStopWithNothingRunningIsNotAnError(t *testing.T) {
	if err := Script(script(t, `echo "Stopping netbird: FAIL"`), "stop", ""); err != nil {
		t.Fatal(err)
	}
}

func TestScriptExitStatusIsAFailure(t *testing.T) {
	err := Script(script(t, `echo "/usr/bin/netbird not found"; exit 1`), "start", "")
	if msg := Message("start failed", err); !strings.Contains(msg, "/usr/bin/netbird not found") {
		t.Fatalf("message is %q", msg)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/vpn/ -v`
Expected: FAIL to compile, `undefined: Tail`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/vpn/tail.go`:

```go
// Package vpn holds what the Tailscale and NetBird add-ons share: the tail of
// a failed command's output, which the page shows as the reason, the runner
// for their boot scripts, the login URL a CLI prints, the daemon's uptime and
// memory, the addons group's memory, the update check's cache, and the
// handlers for start at boot and for refusing while the other VPN runs.
package vpn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TailLines is how many lines of a failed command's output reach the page.
const TailLines = 8

// maxTailBytes caps one tail, so a CLI that prints a wall of text cannot fill
// the page.
const maxTailBytes = 2048

// logTailBytes is how much of a daemon's log is read for its last lines.
const logTailBytes = 16 * 1024

// Tail returns the last n non-blank lines of out, trimmed, one per line.
func Tail(out []byte, n int) string {
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	s := strings.Join(lines, "\n")
	if len(s) > maxTailBytes {
		s = s[len(s)-maxTailBytes:]
	}
	return s
}

// CmdError is a command that failed, with the last lines it printed.
type CmdError struct {
	Err  error
	Tail string
}

func (e *CmdError) Error() string {
	if e.Tail == "" {
		return e.Err.Error()
	}
	return e.Err.Error() + ": " + e.Tail
}

func (e *CmdError) Unwrap() error { return e.Err }

// Run runs cmd and returns its combined output. On failure the error is a
// *CmdError that carries the output's last lines.
func Run(cmd *exec.Cmd) ([]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, &CmdError{Err: err, Tail: Tail(out, TailLines)}
	}
	return out, nil
}

// Message is what a failing handler returns: what failed, then why. The why
// is the command's last lines when there are any, and the error otherwise.
func Message(what string, err error) string {
	if err == nil {
		return what
	}
	var ce *CmdError
	if errors.As(err, &ce) && ce.Tail != "" {
		return what + ":\n" + ce.Tail
	}
	return what + ": " + err.Error()
}

// Script runs an add-on's boot script with one action.
//
// The scripts exit 0 even when the daemon did not come up; they print
// "Starting <daemon>: FAIL" instead. So a start or restart that prints such a
// line fails here too. logFile, when set, is the daemon's own log: a daemon
// that dies at once says why there and not in the script's output.
func Script(path, action, logFile string) error {
	out, err := Run(exec.Command("sh", path, action))
	if err == nil && (action == "start" || action == "restart") && startFailed(out) {
		err = &CmdError{
			Err:  fmt.Errorf("%s %s failed", filepath.Base(path), action),
			Tail: Tail(out, TailLines),
		}
	}
	if err != nil && logFile != "" {
		var ce *CmdError
		if errors.As(err, &ce) {
			if lines := Tail(readEnd(logFile, logTailBytes), TailLines); lines != "" {
				ce.Tail = strings.TrimSpace(ce.Tail + "\n" + lines)
			}
		}
	}
	return err
}

func startFailed(out []byte) bool {
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Starting ") && strings.HasSuffix(line, "FAIL") {
			return true
		}
	}
	return false
}

// readEnd returns up to n bytes from the end of the file, or nothing.
func readEnd(path string, n int64) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	if fi, err := f.Stat(); err == nil && fi.Size() > n {
		if _, err := f.Seek(-n, io.SeekEnd); err != nil {
			return nil
		}
	}
	b, _ := io.ReadAll(io.LimitReader(f, n))
	return b
}
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/vpn/ -v`
Expected: PASS, 9 tests.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/vpn/tail.go server/service/extensions/vpn/tail_test.go
git commit -m "extensions/vpn: the reason a command failed, and the boot script runner"
```

---

### Task 4: The login URL reader moves to vpn

**Files:**
- Create: `server/service/extensions/vpn/login.go`
- Create: `server/service/extensions/vpn/login_test.go`
- Delete: `server/service/extensions/tailscale/cli_test.go`
- Modify: `server/service/extensions/tailscale/cli.go` (imports, `Login`, and the removal of `whitespace`, `loginResult`, `loginURL`, `readLoginURL`, `waitDelay`)

**Interfaces:**
- Consumes: `vpn.CmdError`, `vpn.Tail`, `vpn.TailLines` (Task 3).
- Produces: `vpn.LoginURL(cmd *exec.Cmd, stdout bool, timeout, life time.Duration) (string, error)`; `vpn.ReadLoginURL(r io.Reader) (string, error)`.

- [ ] **Step 1: Write the failing test**

`server/service/extensions/vpn/login_test.go` (the four Tailscale tests move here unchanged in substance, and three are new):

```go
//go:build linux

package vpn

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const tailscaleLine = "To authenticate, visit:\n\n\thttps://login.tailscale.com/a/abcdef\n\n"

// NetBird's openURL with --no-browser: the URL, a space, and an empty code
// message when the code is already in the URL.
const netbirdLine = "Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n"

func waitForFile(t *testing.T, path string, within time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func waitForExit(t *testing.T, pid int, within time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		// A killed but unreaped child is still a zombie and signal 0 finds it.
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestLoginReturnsTheURL(t *testing.T) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("printf %q >&2", tailscaleLine))
	url, err := LoginURL(cmd, false, 5*time.Second, 0)
	if err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if url != "https://login.tailscale.com/a/abcdef" {
		t.Fatalf("unexpected url %q", url)
	}
}

func TestLoginReadsNetBirdFromStdout(t *testing.T) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("printf %q", netbirdLine))
	url, err := LoginURL(cmd, true, 5*time.Second, 0)
	if err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if url != "https://login.netbird.io/activate?user_code=ABCD-EFGH" {
		t.Fatalf("unexpected url %q", url)
	}
}

func TestLoginLeavesTheCommandRunning(t *testing.T) {
	// The login keeps running until the user finishes in the browser. Closing
	// the pipe as soon as the URL is read hands the child a SIGPIPE on its
	// next line and kills the login it was told to complete.
	marker := filepath.Join(t.TempDir(), "finished")
	script := fmt.Sprintf("printf %q >&2; sleep 0.4; echo still-here >&2; touch %q", tailscaleLine, marker)
	if _, err := LoginURL(exec.Command("sh", "-c", script), false, 5*time.Second, 0); err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if !waitForFile(t, marker, 3*time.Second) {
		t.Fatal("expected the login command to run to completion")
	}
}

func TestLoginGivesUpWhenNoURLAppears(t *testing.T) {
	// Production runs the binary directly, so the test does too: killing a
	// shell wrapper would leave the real command holding the pipe.
	start := time.Now()
	if _, err := LoginURL(exec.Command("sleep", "30"), false, 300*time.Millisecond, 0); err == nil {
		t.Fatal("expected an error when no url is printed")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("expected LoginURL to give up quickly, took %s", elapsed)
	}
}

func TestLoginReapsACommandThatGaveUp(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if _, err := LoginURL(cmd, false, 200*time.Millisecond, 0); err == nil {
		t.Fatal("expected an error when no url is printed")
	}
	if !waitForExit(t, cmd.Process.Pid, 5*time.Second) {
		t.Fatal("expected the abandoned command to be reaped")
	}
}

// netbird up has no timeout of its own, so a login nobody finishes is ended
// after its life.
func TestLoginEndsTheCommandAfterItsLife(t *testing.T) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("printf %q; exec sleep 30", netbirdLine))
	if _, err := LoginURL(cmd, true, 5*time.Second, 300*time.Millisecond); err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if !waitForExit(t, cmd.Process.Pid, 5*time.Second) {
		t.Fatal("expected the login to be ended after its life")
	}
}

func TestLoginWithoutAURLCarriesTheOutput(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'daemon is not running' >&2")
	_, err := LoginURL(cmd, false, 5*time.Second, 0)
	if err == nil || !strings.Contains(Message("login failed", err), "daemon is not running") {
		t.Fatalf("the CLI's reason is missing: %v", err)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/vpn/ -run TestLogin -v`
Expected: FAIL to compile, `undefined: LoginURL`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/vpn/login.go`:

```go
package vpn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// waitDelay bounds how long Wait tolerates a still-open pipe after the process
// itself has gone.
const waitDelay = 5 * time.Second

type loginResult struct {
	url string
	err error
}

// LoginURL starts a login command and returns the URL the user has to visit.
// Tailscale prints it on stderr, NetBird on stdout.
//
// The command keeps running afterwards, until the login is completed in the
// browser, so its output has to keep draining: closing the pipe early hands it
// a SIGPIPE on its next line and kills the login. It also has to be reaped
// rather than left as an orphan, once, on every path out of here.
//
// timeout bounds the wait for the URL. life, when positive, bounds how long
// the command may run after it: netbird up has no timeout of its own and would
// otherwise wait for the browser forever.
func LoginURL(cmd *exec.Cmd, stdout bool, timeout, life time.Duration) (string, error) {
	var pipe io.ReadCloser
	var err error
	if stdout {
		pipe, err = cmd.StdoutPipe()
	} else {
		pipe, err = cmd.StderrPipe()
	}
	if err != nil {
		return "", err
	}

	// Safety net if the command ever leaves a child holding the pipe open.
	cmd.WaitDelay = waitDelay

	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Buffered, so this goroutine still finishes if nobody is listening.
	results := make(chan loginResult, 1)

	go func() {
		url, err := ReadLoginURL(pipe)
		results <- loginResult{url: url, err: err}

		// Wait closes the pipe itself, so it must not be closed here.
		_, _ = io.Copy(io.Discard, pipe)
		_ = cmd.Wait()
	}()

	select {
	case result := <-results:
		if result.err == nil && life > 0 {
			// Kill after Wait is harmless: it reports the process done.
			time.AfterFunc(life, func() { _ = cmd.Process.Kill() })
		}
		return result.url, result.err

	case <-time.After(timeout):
		// Otherwise the handler blocks for the command's whole life.
		_ = cmd.Process.Kill()
		return "", fmt.Errorf("timed out waiting for the login url")
	}
}

// ReadLoginURL reads lines until one holds an https URL and returns that URL.
// Without one, the error carries the last lines it read, which say why.
func ReadLoginURL(reader io.Reader) (string, error) {
	buffered := bufio.NewReader(reader)
	var seen bytes.Buffer
	for {
		line, err := buffered.ReadString('\n')
		for _, field := range strings.Fields(line) {
			if strings.HasPrefix(field, "https://") {
				return field, nil
			}
		}
		seen.WriteString(line)
		if err != nil {
			return "", &CmdError{Err: errors.New("no login url in the output"), Tail: Tail(seen.Bytes(), TailLines)}
		}
	}
}
```

Now switch Tailscale to it. Delete `server/service/extensions/tailscale/cli_test.go`:

```bash
git rm server/service/extensions/tailscale/cli_test.go
```

In `server/service/extensions/tailscale/cli.go`, replace the import block with:

```go
import (
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)
```

Replace the `const (...)` block with:

```go
const (
	ScriptPath       = "/etc/init.d/S98tailscaled"
	ScriptBackupPath = "/kvmapp/system/init.d/S98tailscaled"

	// loginURLTimeout bounds how long the handler waits for the login URL.
	// The command itself runs for ten minutes waiting on the browser.
	loginURLTimeout = 60 * time.Second
)
```

Replace the `Login` method and delete everything from `var whitespace = regexp.MustCompile` through the end of `readLoginURL` (the `whitespace` var, the `loginResult` type, `loginURL` and `readLoginURL`). The new `Login`:

```go
func (c *Cli) Login() (string, error) {
	// No shell: killing "sh -c tailscale ..." leaves tailscale holding the
	// stderr pipe, so the timeout could never take effect.
	cmd := exec.Command("tailscale", "login", "--accept-dns=false", "--timeout=10m")

	return vpn.LoginURL(cmd, false, loginURLTimeout, 0)
}
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/... -v -run 'TestLogin|TestPlaceBinaries'`
Expected: PASS, 7 `TestLogin*` in vpn and the tailscale placement tests.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/vpn/login.go server/service/extensions/vpn/login_test.go server/service/extensions/tailscale/cli.go
git commit -m "extensions/vpn: the login URL reader, shared, for stdout and stderr"
```

---

### Task 5: One exclusivity check, and the start-at-boot record

**Files:**
- Create: `server/service/extensions/addon/exclusive.go`
- Test: `server/service/extensions/addon/exclusive_test.go`

**Interfaces:**
- Consumes: `addon.OnData()`, `addon.Dir(name)`, `addon.SetEnabled(name, on)` (existing).
- Produces:
  - vars `addon.InitdDir = "/etc/init.d"`, `addon.PkgInitdDir = "/kvmapp/system/init.d"`, `addon.ProcDir = "/proc"`
  - `type addon.Daemon struct{ Name, Title, Initd, PidFile, Process string }`, method `(d Daemon) Script() string`
  - vars `addon.Tailscale`, `addon.NetBird` (Daemon values; tests change their `PidFile`)
  - `addon.PID(d Daemon) (int, bool)`, `addon.Running(d Daemon) bool`
  - `addon.BootEnabled(d Daemon) bool`, `addon.SetBoot(d Daemon, on bool) error`
  - `addon.BlockedBy(self string) string`, `addon.CheckExclusive(self string) error`, `type addon.BlockedError struct{ Self, By Daemon }`

- [ ] **Step 1: Write the failing test**

`server/service/extensions/addon/exclusive_test.go`:

```go
//go:build linux

package addon

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// scratchDaemons points the pid files, /proc and both init.d directories at a
// temporary tree, and puts a boot script for each daemon in the package copy.
// onData decides whether it also looks like a distribution image with /data.
func scratchDaemons(t *testing.T, onData bool) {
	t.Helper()
	if onData {
		scratch(t)
	} else {
		saved := DistroMarker
		t.Cleanup(func() { DistroMarker = saved })
		DistroMarker = filepath.Join(t.TempDir(), "absent")
	}
	base := t.TempDir()
	savedTs, savedNb := Tailscale, NetBird
	savedInitd, savedPkg, savedProc := InitdDir, PkgInitdDir, ProcDir
	t.Cleanup(func() {
		Tailscale, NetBird = savedTs, savedNb
		InitdDir, PkgInitdDir, ProcDir = savedInitd, savedPkg, savedProc
	})
	InitdDir = filepath.Join(base, "etc-init.d")
	PkgInitdDir = filepath.Join(base, "kvmapp-init.d")
	ProcDir = filepath.Join(base, "proc")
	for _, d := range []string{InitdDir, PkgInitdDir, ProcDir, filepath.Join(base, "run")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	Tailscale.PidFile = filepath.Join(base, "run", "tailscaled.pid")
	NetBird.PidFile = filepath.Join(base, "run", "netbird.pid")
	for _, s := range []string{Tailscale.Initd, NetBird.Initd} {
		if err := os.WriteFile(filepath.Join(PkgInitdDir, s), []byte("#!/bin/sh\n# "+s+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// fakeRunning writes d's pid file and a /proc entry whose command line is argv.
func fakeRunning(t *testing.T, d Daemon, pid int, argv ...string) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(strings.Join(argv, "\x00")+"\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPIDNeedsTheDaemonsCommandLine(t *testing.T) {
	scratchDaemons(t, false)
	if _, ok := PID(NetBird); ok {
		t.Fatal("no pid file is not running")
	}

	// After a power cut the pid file can name a pid something else now holds.
	fakeRunning(t, NetBird, 4242, "/bin/sleep", "60")
	if Running(NetBird) {
		t.Fatal("a pid that is not netbird's is not netbird running")
	}

	fakeRunning(t, NetBird, 4243, "/usr/bin/netbird", "service", "run")
	if pid, ok := PID(NetBird); !ok || pid != 4243 {
		t.Fatalf("got %d %v, want 4243 true", pid, ok)
	}

	if err := os.WriteFile(NetBird.PidFile, []byte("junk\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Running(NetBird) {
		t.Fatal("a pid file that is not a number is not running")
	}
}

func TestBootOffADistributionImageIsTheScriptInInitd(t *testing.T) {
	scratchDaemons(t, false)
	if BootEnabled(NetBird) {
		t.Fatal("nothing enabled yet")
	}
	if err := SetBoot(NetBird, true); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(InitdDir, "S98netbird"))
	if err != nil || fi.Mode().Perm()&0o100 == 0 {
		t.Fatalf("the boot script must be in init.d and executable: %v", err)
	}
	if !BootEnabled(NetBird) {
		t.Fatal("start at boot must read back on")
	}
	if err := SetBoot(NetBird, false); err != nil {
		t.Fatal(err)
	}
	if BootEnabled(NetBird) {
		t.Fatal("start at boot must read back off")
	}
	if err := SetBoot(NetBird, false); err != nil {
		t.Fatalf("turning it off twice is fine: %v", err)
	}
}

func TestBootOnADistributionImageIsTheEnabledFile(t *testing.T) {
	scratchDaemons(t, true)
	if err := SetBoot(Tailscale, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(Dir("tailscale"), "enabled")); err != nil {
		t.Fatal("the enabled file must be on /data")
	}
	// A new image has no copy in /etc/init.d until S04addons puts it back.
	_ = os.Remove(filepath.Join(InitdDir, "S98tailscaled"))
	if !BootEnabled(Tailscale) {
		t.Fatal("on /data the enabled file is the record")
	}
	if err := SetBoot(Tailscale, false); err != nil {
		t.Fatal(err)
	}
	if BootEnabled(Tailscale) {
		t.Fatal("start at boot must read back off")
	}
}

func TestBlockedByNamesTheOtherWhenItRuns(t *testing.T) {
	scratchDaemons(t, false)
	fakeRunning(t, Tailscale, 100, "/usr/sbin/tailscaled", "--state=x")
	if got := BlockedBy("netbird"); got != "tailscale" {
		t.Fatalf("got %q", got)
	}
	if got := BlockedBy("tailscale"); got != "" {
		t.Fatalf("a daemon does not block itself: %q", got)
	}
}

func TestBlockedByNamesTheOtherWhenItStartsAtBoot(t *testing.T) {
	scratchDaemons(t, false)
	if err := SetBoot(NetBird, true); err != nil {
		t.Fatal(err)
	}
	if got := BlockedBy("tailscale"); got != "netbird" {
		t.Fatalf("got %q", got)
	}
}

func TestNothingBlocksWhenBothAreIdle(t *testing.T) {
	scratchDaemons(t, false)
	if err := CheckExclusive("netbird"); err != nil {
		t.Fatal(err)
	}
	if err := CheckExclusive("tailscale"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckExclusiveNamesTheOther(t *testing.T) {
	scratchDaemons(t, false)
	fakeRunning(t, Tailscale, 100, "/usr/sbin/tailscaled")
	err := CheckExclusive("netbird")
	var be *BlockedError
	if !errors.As(err, &be) || be.By.Name != "tailscale" {
		t.Fatalf("want a BlockedError by tailscale, got %v", err)
	}
	for _, want := range []string{"Tailscale is running or starts at boot", "before you use NetBird"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("%q is missing from %q", want, err.Error())
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/addon/ -v`
Expected: FAIL to compile, `undefined: Tailscale`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/addon/exclusive.go`:

```go
package addon

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Paths the boot and exclusivity checks read. Variables so the tests can point
// them at a scratch tree.
var (
	InitdDir    = "/etc/init.d"
	PkgInitdDir = "/kvmapp/system/init.d"
	ProcDir     = "/proc"
)

// Daemon is an add-on that runs as a daemon in the addons memory group.
type Daemon struct {
	Name    string // the add-on's name, as in Spec
	Title   string // how the page names it
	Initd   string // its boot script
	PidFile string // written by the boot script
	Process string // an argument its /proc/<pid>/cmdline holds, by base name
}

// Tailscale and NetBird each hold the addons group near its memory.high on
// their own, idle; together they hold it at the limit (measured on image w,
// 2026-09-27). So only one of them may run or start at boot.
var (
	Tailscale = Daemon{
		Name: "tailscale", Title: "Tailscale", Initd: "S98tailscaled",
		PidFile: "/var/run/tailscaled.pid", Process: "tailscaled",
	}
	NetBird = Daemon{
		Name: "netbird", Title: "NetBird", Initd: "S98netbird",
		PidFile: "/var/run/netbird.pid", Process: "netbird",
	}
)

// exclusive is read at every call, so a test that moves a pid file is seen.
func exclusive() []Daemon { return []Daemon{Tailscale, NetBird} }

// Script is the boot script in the package copy, which start, stop and restart
// run. /etc/init.d holds a copy only while the daemon starts at boot.
func (d Daemon) Script() string { return filepath.Join(PkgInitdDir, d.Initd) }

// PID is the daemon's process id, when its pid file names a live process
// whose command line holds d.Process. A stale pid file after a power cut can
// name a pid that something else now holds, and that is not the daemon.
func PID(d Daemon) (int, bool) {
	b, err := os.ReadFile(d.PidFile)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	cmdline, err := os.ReadFile(filepath.Join(ProcDir, strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return 0, false
	}
	for _, arg := range bytes.Split(cmdline, []byte{0}) {
		if len(arg) > 0 && filepath.Base(string(arg)) == d.Process {
			return pid, true
		}
	}
	return 0, false
}

// Running reports whether the daemon runs now.
func Running(d Daemon) bool {
	_, ok := PID(d)
	return ok
}

// BootEnabled reports whether the daemon starts at boot. On a distribution
// image that is the add-on's enabled file on /data, which S04addons reads;
// anywhere else it is the script in /etc/init.d.
func BootEnabled(d Daemon) bool {
	if OnData() {
		return exists(filepath.Join(Dir(d.Name), "enabled"))
	}
	return exists(filepath.Join(InitdDir, d.Initd))
}

// SetBoot turns start at boot on or off. It puts the boot script into
// /etc/init.d or takes it out, which is the whole record off a distribution
// image; on one it also writes the enabled file, so the next image keeps it.
func SetBoot(d Daemon, on bool) error {
	dst := filepath.Join(InitdDir, d.Initd)
	if on {
		if err := copyFile(d.Script(), dst, 0o755); err != nil {
			return err
		}
	} else if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if OnData() {
		return SetEnabled(d.Name, on)
	}
	return nil
}

// BlockedBy names the other daemon when it runs or starts at boot, and is
// empty when self may go ahead.
func BlockedBy(self string) string {
	for _, d := range exclusive() {
		if d.Name != self && (Running(d) || BootEnabled(d)) {
			return d.Name
		}
	}
	return ""
}

// BlockedError refuses an action because the other VPN runs or starts at
// boot.
type BlockedError struct {
	Self, By Daemon
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("%s is running or starts at boot. Only one VPN runs at a time: "+
		"stop %s and turn off its start at boot before you use %s.",
		e.By.Title, e.By.Title, e.Self.Title)
}

// CheckExclusive is the one check both add-ons call before install, start,
// up, login and turning start at boot on.
func CheckExclusive(self string) error {
	by := BlockedBy(self)
	if by == "" {
		return nil
	}
	return &BlockedError{Self: byName(self), By: byName(by)}
}

func byName(name string) Daemon {
	for _, d := range exclusive() {
		if d.Name == name {
			return d
		}
	}
	return Daemon{Name: name, Title: name}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// copyFile writes a copy through a temporary file and a rename, so a boot
// that reads /etc/init.d never sees half a script.
func copyFile(src, dst string, mode os.FileMode) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, b, mode); err != nil {
		return err
	}
	// WriteFile honours the umask.
	if err := os.Chmod(tmp, mode); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/addon/ -v`
Expected: PASS, the existing addon tests and the 7 new ones.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/addon/exclusive.go server/service/extensions/addon/exclusive_test.go
git commit -m "extensions/addon: one VPN at a time, and start at boot as its own record"
```

---

### Task 6: Uptime, memory, and the update cache

**Files:**
- Create: `server/service/extensions/vpn/proc.go`
- Create: `server/service/extensions/vpn/version.go`
- Test: `server/service/extensions/vpn/proc_test.go`, `server/service/extensions/vpn/version_test.go`

**Interfaces:**
- Consumes: `addon.ProcDir`, `addon.PID`, `addon.BootEnabled`, `addon.BlockedBy`, `addon.Daemon` (Task 5); `proto.VpnStatus`, `proto.VpnPeer` (Task 2).
- Produces: var `vpn.CgroupDir = "/sys/fs/cgroup/addons"`; `vpn.UptimeSec(pid int) int64`; `vpn.RSS(pid int) int64`; `vpn.GroupMemory() (current, high, max int64)`; `vpn.Fill(st *proto.VpnStatus, d addon.Daemon)`; `const vpn.UpdateTTL = time.Hour`; `type vpn.VersionCache struct{ TTL time.Duration; Now func() time.Time; ... }` with `Latest(fetch func() (string, error)) (string, error)` and `Reset()`.

- [ ] **Step 1: Write the failing tests**

`server/service/extensions/vpn/proc_test.go`:

```go
package vpn

import (
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
)

// scratchProc is a /proc with one netbird at pid 4242 that started 500 s after
// boot on a system up 1700 s, and an addons group at 61.6 MB of 64M.
func scratchProc(t *testing.T) {
	t.Helper()
	base := t.TempDir()
	saved := struct {
		proc, cg, marker, initd string
		ts, nb                  addon.Daemon
	}{addon.ProcDir, CgroupDir, addon.DistroMarker, addon.InitdDir, addon.Tailscale, addon.NetBird}
	t.Cleanup(func() {
		addon.ProcDir, CgroupDir, addon.DistroMarker, addon.InitdDir = saved.proc, saved.cg, saved.marker, saved.initd
		addon.Tailscale, addon.NetBird = saved.ts, saved.nb
	})
	addon.ProcDir = filepath.Join(base, "proc")
	CgroupDir = filepath.Join(base, "addons")
	addon.DistroMarker = filepath.Join(base, "absent")
	addon.InitdDir = filepath.Join(base, "init.d")
	addon.Tailscale.PidFile = filepath.Join(base, "tailscaled.pid")
	addon.NetBird.PidFile = filepath.Join(base, "netbird.pid")

	files := map[string]string{
		"proc/4242/stat":        "4242 (net bird) S 1 4242 4242 0 -1 4194560 100 0 0 0 5 3 0 0 20 0 9 0 50000 44000000 7950\n",
		"proc/4242/status":      "Name:\tnetbird\nVmPeak:\t   50000 kB\nVmRSS:\t   31800 kB\n",
		"proc/4242/cmdline":     "/usr/bin/netbird\x00service\x00run\x00",
		"proc/uptime":           "1700.25 3000.00\n",
		"netbird.pid":           "4242\n",
		"addons/memory.current": "64592691\n",
		"addons/memory.high":    "67108864\n",
		"addons/memory.max":     "max\n",
	}
	for name, body := range files {
		path := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUptimeSec(t *testing.T) {
	scratchProc(t)
	if got := UptimeSec(4242); got != 1200 {
		t.Fatalf("got %d, want 1200", got)
	}
	if got := UptimeSec(9999); got != 0 {
		t.Fatalf("a missing process has no uptime, got %d", got)
	}
}

func TestRSS(t *testing.T) {
	scratchProc(t)
	if got := RSS(4242); got != 31800*1024 {
		t.Fatalf("got %d", got)
	}
}

func TestGroupMemoryReadsMaxAsNoLimit(t *testing.T) {
	scratchProc(t)
	current, high, max := GroupMemory()
	if current != 64592691 || high != 67108864 || max != 0 {
		t.Fatalf("got %d %d %d", current, high, max)
	}
	CgroupDir = filepath.Join(t.TempDir(), "absent")
	if c, h, m := GroupMemory(); c != 0 || h != 0 || m != 0 {
		t.Fatalf("no group reads as zeros, got %d %d %d", c, h, m)
	}
}

func TestFill(t *testing.T) {
	scratchProc(t)
	st := proto.VpnStatus{State: proto.VpnRunning}
	Fill(&st, addon.NetBird)
	if st.UptimeSec != 1200 || st.Memory.DaemonRSS != 31800*1024 {
		t.Fatalf("daemon figures: %+v", st)
	}
	if st.Memory.GroupCurrent != 64592691 || st.Memory.GroupHigh != 67108864 || st.Memory.GroupMax != 0 {
		t.Fatalf("group figures: %+v", st.Memory)
	}
	if st.BootEnabled || st.BlockedBy != "" {
		t.Fatalf("nothing enabled and nothing blocking: %+v", st)
	}
	if st.Peers == nil {
		t.Fatal("peers must be an empty list, not null, for the page")
	}
}
```

`server/service/extensions/vpn/version_test.go`:

```go
package vpn

import (
	"errors"
	"testing"
	"time"
)

func TestVersionCacheKeepsAnAnswerForItsTTL(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	c := &VersionCache{TTL: UpdateTTL, Now: func() time.Time { return now }}
	calls := 0
	fetch := func() (string, error) { calls++; return "0.79.0", nil }

	for i := 0; i < 3; i++ {
		if v, err := c.Latest(fetch); err != nil || v != "0.79.0" {
			t.Fatalf("got %q %v", v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("fetched %d times within the hour", calls)
	}

	now = now.Add(UpdateTTL + time.Second)
	if _, err := c.Latest(fetch); err != nil || calls != 2 {
		t.Fatalf("an hour later it must ask again: calls=%d err=%v", calls, err)
	}

	c.Reset()
	if _, err := c.Latest(fetch); err != nil || calls != 3 {
		t.Fatalf("after Reset it must ask again: calls=%d err=%v", calls, err)
	}
}

func TestVersionCacheDoesNotKeepAFailure(t *testing.T) {
	c := &VersionCache{TTL: UpdateTTL}
	if _, err := c.Latest(func() (string, error) { return "", errors.New("offline") }); err == nil {
		t.Fatal("the failure must reach the caller")
	}
	if v, err := c.Latest(func() (string, error) { return "1.90.1", nil }); err != nil || v != "1.90.1" {
		t.Fatalf("a failure must not be cached: %q %v", v, err)
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `gotest ./service/extensions/vpn/ -run 'TestUptimeSec|TestRSS|TestGroupMemory|TestFill|TestVersionCache' -v`
Expected: FAIL to compile, `undefined: UptimeSec`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/vpn/proc.go`:

```go
package vpn

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
)

// CgroupDir is the addons memory group that ironkvm-dist's S01cgroups makes.
// A variable for the tests.
var CgroupDir = "/sys/fs/cgroup/addons"

// clockTicks is USER_HZ, the unit of the start time in /proc/<pid>/stat. It is
// 100 on every Linux this server runs on.
const clockTicks = 100

// UptimeSec is how long the process has run, in whole seconds, or 0 when it
// cannot be read.
func UptimeSec(pid int) int64 {
	stat, err := os.ReadFile(filepath.Join(addon.ProcDir, strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0
	}
	// The command name is in parentheses and may hold spaces, so the fields
	// are counted from the last ')'. After it, field 3 (state) is index 0, and
	// field 22 (starttime) is index 19.
	s := string(stat)
	i := strings.LastIndexByte(s, ')')
	if i < 0 {
		return 0
	}
	fields := strings.Fields(s[i+1:])
	if len(fields) < 20 {
		return 0
	}
	start, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil {
		return 0
	}
	up, err := os.ReadFile(filepath.Join(addon.ProcDir, "uptime"))
	if err != nil {
		return 0
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(up)), " ")
	system, err := strconv.ParseFloat(first, 64)
	if err != nil {
		return 0
	}
	if sec := int64(system) - start/clockTicks; sec > 0 {
		return sec
	}
	return 0
}

// RSS is the process's resident memory in bytes, from VmRSS, or 0.
func RSS(pid int) int64 {
	b, err := os.ReadFile(filepath.Join(addon.ProcDir, strconv.Itoa(pid), "status"))
	if err != nil {
		return 0
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		rest, ok := strings.CutPrefix(sc.Text(), "VmRSS:")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			return 0
		}
		kb, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0
		}
		return kb * 1024
	}
	return 0
}

// GroupMemory reads the addons group's use and its two limits, in bytes. A
// limit of "max", or a group that is not there, reads as 0.
func GroupMemory() (current, high, max int64) {
	return readBytes("memory.current"), readBytes("memory.high"), readBytes("memory.max")
}

func readBytes(name string) int64 {
	b, err := os.ReadFile(filepath.Join(CgroupDir, name))
	if err != nil {
		return 0
	}
	v, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// Fill adds what both add-ons report the same way: start at boot, the other
// VPN when it blocks this one, and the daemon's uptime and memory beside its
// group's.
func Fill(st *proto.VpnStatus, d addon.Daemon) {
	st.BootEnabled = addon.BootEnabled(d)
	st.BlockedBy = addon.BlockedBy(d.Name)
	if pid, ok := addon.PID(d); ok {
		st.UptimeSec = UptimeSec(pid)
		st.Memory.DaemonRSS = RSS(pid)
	}
	st.Memory.GroupCurrent, st.Memory.GroupHigh, st.Memory.GroupMax = GroupMemory()
	if st.Peers == nil {
		st.Peers = []proto.VpnPeer{}
	}
}
```

`server/service/extensions/vpn/version.go`:

```go
package vpn

import (
	"sync"
	"time"
)

// UpdateTTL is how long a looked-up latest version is kept. The page asks on
// every visit; the package servers need not hear about each one.
const UpdateTTL = time.Hour

// VersionCache keeps the latest version an update check found.
type VersionCache struct {
	TTL time.Duration
	// Now is the clock, for the tests. Nil means time.Now.
	Now func() time.Time

	mu     sync.Mutex
	at     time.Time
	latest string
}

// Latest returns the cached version while it is younger than TTL, and asks
// fetch otherwise. A failed fetch is not cached. The lock is held through the
// fetch, so two pages open at once make one request.
func (c *VersionCache) Latest(fetch func() (string, error)) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	if c.latest != "" && now().Sub(c.at) < c.TTL {
		return c.latest, nil
	}
	v, err := fetch()
	if err != nil {
		return "", err
	}
	c.latest, c.at = v, now()
	return v, nil
}

// Reset forgets the cached version, after an update.
func (c *VersionCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latest = ""
	c.at = time.Time{}
}
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/vpn/ -v`
Expected: PASS, every vpn test.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/vpn/proc.go server/service/extensions/vpn/proc_test.go server/service/extensions/vpn/version.go server/service/extensions/vpn/version_test.go
git commit -m "extensions/vpn: uptime and memory from /proc and the addons group, and the update cache"
```

---

### Task 7: Tailscale status in the shared shape, with peers

**Files:**
- Create: `server/service/extensions/tailscale/status.go`
- Modify: `server/service/extensions/tailscale/cli.go` (imports, `TsStatus` removed, `Status`)
- Modify: `server/service/extensions/tailscale/service.go` (`StateMap` removed, `GetStatus`, imports)
- Test: `server/service/extensions/tailscale/status_test.go`

**Interfaces:**
- Consumes: `proto.VpnStatus`, `proto.VpnPeer`, `proto.VpnState` constants (Task 2); `vpn.Fill` (Task 6); `addon.Tailscale` (Task 5); fixture `testdata/status-running.json` (Task 1).
- Produces: `type TsPeer struct{ HostName, DNSName string; TailscaleIPs []string; Online bool }`; `type TsStatus struct{ Version, BackendState string; Self TsPeer; CurrentTailnet struct{ Name string }; Peer map[string]TsPeer }`; `StateMap map[string]proto.VpnState`; `parseStatus(out []byte) (*TsStatus, error)`; `toVpnStatus(ts *TsStatus) (proto.VpnStatus, error)`. `GET /api/extensions/tailscale/status` returns a full `VpnStatus`.

- [ ] **Step 1: Write the failing test**

`server/service/extensions/tailscale/status_test.go`:

```go
package tailscale

import (
	"os"
	"reflect"
	"testing"

	"NanoKVM-Server/proto"
)

func TestParseStatusRunning(t *testing.T) {
	b, err := os.ReadFile("testdata/status-running.json")
	if err != nil {
		t.Fatal(err)
	}
	ts, err := parseStatus(b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := toVpnStatus(ts)
	if err != nil {
		t.Fatal(err)
	}
	want := proto.VpnStatus{
		State:   proto.VpnRunning,
		Version: "1.90.1",
		IP:      "100.101.102.103",
		Name:    "nanokvm",
		Account: "owner@example.com",
		Control: true,
		Peers: []proto.VpnPeer{
			{Name: "laptop", IP: "100.64.0.2", Online: true},
			{Name: "phone", IP: "100.64.0.3", Online: false},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestParseStatusNeedsLogin(t *testing.T) {
	ts, err := parseStatus([]byte(`{"BackendState":"NeedsLogin","Self":{"HostName":"nanokvm","TailscaleIPs":null,"Online":false},"CurrentTailnet":null,"Peer":null}`))
	if err != nil {
		t.Fatal(err)
	}
	got, err := toVpnStatus(ts)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != proto.VpnNotLogin || got.Name != "nanokvm" || got.Peers == nil || len(got.Peers) != 0 {
		t.Fatalf("got %+v", got)
	}
}

// The CLI prints a version warning before the JSON when it and the daemon
// differ.
func TestParseStatusSkipsALeadingWarning(t *testing.T) {
	ts, err := parseStatus([]byte("Warning: client version \"1.88.3\" != tailscaled server version \"1.90.1\"\n" +
		`{"BackendState":"Stopped","Self":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := toVpnStatus(ts); err != nil || got.State != proto.VpnStopped {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestParseStatusRejectsAnUnknownState(t *testing.T) {
	ts, err := parseStatus([]byte(`{"BackendState":"Exploded"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := toVpnStatus(ts); err == nil {
		t.Fatal("an unknown state must be an error")
	}
}

func TestParseStatusRejectsGarbage(t *testing.T) {
	if _, err := parseStatus([]byte("failed to connect to local tailscaled")); err == nil {
		t.Fatal("output with no JSON must be an error")
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/tailscale/ -run TestParseStatus -v`
Expected: FAIL to compile, `undefined: parseStatus`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/tailscale/status.go`:

```go
package tailscale

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"

	"NanoKVM-Server/proto"
)

// TsPeer is one node in `tailscale status --json`: Self, or an entry of Peer.
type TsPeer struct {
	HostName     string   `json:"HostName"`
	DNSName      string   `json:"DNSName"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Online       bool     `json:"Online"`
}

// TsStatus is the part of `tailscale status --json` the page shows.
type TsStatus struct {
	Version        string `json:"Version"`
	BackendState   string `json:"BackendState"`
	Self           TsPeer `json:"Self"`
	CurrentTailnet struct {
		Name string `json:"Name"`
	} `json:"CurrentTailnet"`
	Peer map[string]TsPeer `json:"Peer"`
}

var StateMap = map[string]proto.VpnState{
	"NoState":          proto.VpnNotRunning,
	"Starting":         proto.VpnNotRunning,
	"NeedsLogin":       proto.VpnNotLogin,
	"NeedsMachineAuth": proto.VpnNotLogin,
	"InUseOtherUser":   proto.VpnNotLogin,
	"Running":          proto.VpnRunning,
	"Stopped":          proto.VpnStopped,
}

// parseStatus reads the CLI's JSON. The CLI can print a warning line before
// it, so everything before the first brace is skipped.
func parseStatus(out []byte) (*TsStatus, error) {
	i := bytes.IndexByte(out, '{')
	if i < 0 {
		return nil, errors.New("unknown output")
	}
	var st TsStatus
	if err := json.Unmarshal(out[i:], &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// toVpnStatus is the page's view of the CLI's status. Uptime, memory, start at
// boot and the blocker are vpn.Fill's.
func toVpnStatus(ts *TsStatus) (proto.VpnStatus, error) {
	state, ok := StateMap[ts.BackendState]
	if !ok {
		return proto.VpnStatus{}, fmt.Errorf("unknown tailscale state: %s", ts.BackendState)
	}
	st := proto.VpnStatus{
		State:   state,
		Version: shortVersion(ts.Version),
		IP:      ipv4(ts.Self.TailscaleIPs),
		Name:    ts.Self.HostName,
		Account: ts.CurrentTailnet.Name,
		Control: ts.Self.Online,
		Peers:   []proto.VpnPeer{},
	}
	for _, p := range ts.Peer {
		st.Peers = append(st.Peers, proto.VpnPeer{Name: peerName(p), IP: ipv4(p.TailscaleIPs), Online: p.Online})
	}
	sort.Slice(st.Peers, func(i, j int) bool { return st.Peers[i].Name < st.Peers[j].Name })
	return st, nil
}

// shortVersion drops the build suffix: 1.90.1-t8b5c4a1e2-g3f6d7c8b9 is 1.90.1.
func shortVersion(v string) string {
	s, _, _ := strings.Cut(v, "-")
	return s
}

// ipv4 is the node's IPv4 address, the one the page has always shown.
func ipv4(ips []string) string {
	for _, s := range ips {
		if ip := net.ParseIP(s); ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}

// peerName is the first label of the MagicDNS name, which is how the tailnet
// names the machine, or the OS host name when there is none.
func peerName(p TsPeer) string {
	if label, _, _ := strings.Cut(p.DNSName, "."); label != "" {
		return label
	}
	return p.HostName
}
```

In `server/service/extensions/tailscale/cli.go`: delete the `TsStatus` type, replace the import block, and replace `Status`:

```go
import (
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)
```

```go
func (c *Cli) Status() (*TsStatus, error) {
	output, err := exec.Command("sh", "-c", "tailscale status --json").CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseStatus(output)
}
```

In `server/service/extensions/tailscale/service.go`: delete `StateMap` (it moved), remove `"net"` from the imports and add `"NanoKVM-Server/service/extensions/vpn"`, and replace `GetStatus`:

```go
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	st := proto.VpnStatus{State: proto.VpnNotInstall}
	if isInstalled() {
		st = proto.VpnStatus{State: proto.VpnNotRunning}
		if ts, err := NewCli().Status(); err != nil {
			log.Debugf("failed to get tailscale status: %s", err)
		} else if st, err = toVpnStatus(ts); err != nil {
			log.Errorf("%s", err)
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}

	vpn.Fill(&st, addon.Tailscale)
	rsp.OkRspWithData(c, &st)
}
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/tailscale/ -v`
Expected: PASS, the 5 new tests and the existing placement tests.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/tailscale/status.go server/service/extensions/tailscale/status_test.go server/service/extensions/tailscale/cli.go server/service/extensions/tailscale/service.go
git commit -m "tailscale: status in the shared shape, with version, control and peers"
```

---

### Task 8: Tailscale start, stop and boot apart; refusal; error tails

**Files:**
- Create: `server/service/extensions/vpn/handlers.go`
- Modify: `server/service/extensions/tailscale/cli.go` (whole file)
- Modify: `server/service/extensions/tailscale/service.go` (whole file)
- Modify: `server/service/extensions/tailscale/addon_test.go` (delete `TestRecordEnabledOnlyOnData`)
- Modify: `server/router/extensions.go` (add the boot route)
- Test: `server/service/extensions/tailscale/service_test.go`

**Interfaces:**
- Consumes: `addon.Tailscale`, `addon.NetBird`, `addon.CheckExclusive`, `addon.SetBoot`, `addon.BootEnabled`, `addon.InitdDir`, `addon.PkgInitdDir`, `addon.ProcDir` (Task 5); `vpn.Script`, `vpn.Run`, `vpn.Message` (Task 3); `vpn.LoginURL` (Task 4); `vpn.Fill` (Task 6); `proto.VpnBootReq`, `proto.VpnLoginRsp` (Task 2).
- Produces: `vpn.Refuse(c *gin.Context, d addon.Daemon) bool`; `vpn.Boot(c *gin.Context, d addon.Daemon, installed bool)`; `(*tailscale.Service).Boot`; route `POST /api/extensions/tailscale/boot`. Test helpers in package tailscale: `stub`, `lifecycle`, `fakeRunning`, `callsOf`, `call` (used by Task 9).

- [ ] **Step 1: Write the failing test**

`server/service/extensions/tailscale/service_test.go`:

```go
//go:build linux

package tailscale

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
)

func stub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// lifecycle is scratchImage off a distribution image, with stub binaries, a
// stub boot script in the package copy, and scratch pid files, /proc and
// /etc/init.d. Every stub appends to the file it returns. STUB_RESULT=FAIL
// makes the script's start print FAIL; STUB_FAIL makes `tailscale up` fail.
func lifecycle(t *testing.T) (calls string) {
	t.Helper()
	fsroot, _ := scratchImage(t, false)
	base := filepath.Dir(fsroot)
	calls = filepath.Join(base, "calls")

	savedTs, savedNb := addon.Tailscale, addon.NetBird
	savedInitd, savedPkg, savedProc := addon.InitdDir, addon.PkgInitdDir, addon.ProcDir
	savedLimit := utils.GoMemLimitFile
	t.Cleanup(func() {
		addon.Tailscale, addon.NetBird = savedTs, savedNb
		addon.InitdDir, addon.PkgInitdDir, addon.ProcDir = savedInitd, savedPkg, savedProc
		utils.GoMemLimitFile = savedLimit
	})
	addon.InitdDir = filepath.Join(base, "etc-init.d")
	addon.PkgInitdDir = filepath.Join(base, "kvmapp-init.d")
	addon.ProcDir = filepath.Join(base, "proc")
	for _, d := range []string{addon.InitdDir, addon.PkgInitdDir, addon.ProcDir, filepath.Join(base, "run")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	addon.Tailscale.PidFile = filepath.Join(base, "run", "tailscaled.pid")
	addon.NetBird.PidFile = filepath.Join(base, "run", "netbird.pid")
	utils.GoMemLimitFile = filepath.Join(base, "GOMEMLIMIT")

	stub(t, filepath.Join(addon.PkgInitdDir, "S98tailscaled"), `echo "script $1" >> "`+calls+`"
case "$1" in
start) echo "GOMEMLIMIT set to 56MiB"; echo "Starting tailscaled[1.2.3]: ${STUB_RESULT:-OK}" ;;
stop) echo "Stopping tailscaled: OK" ;;
esac`)
	stub(t, TailscalePath, `echo "tailscale $*" >> "`+calls+`"
case "$1" in
version) echo "1.88.3"; echo "  tailscale commit: abc" ;;
up) if [ -n "$STUB_FAIL" ]; then echo "backend error: tailscaled is not running" >&2; exit 1; fi ;;
esac`)
	stub(t, TailscaledPath, `exit 0`)
	return calls
}

// fakeRunning writes d's pid file and a /proc entry whose command line starts
// with argv0.
func fakeRunning(t *testing.T, d addon.Daemon, pid int, argv0 string) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(addon.ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(argv0+"\x00service\x00run\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func callsOf(t *testing.T, calls string) string {
	t.Helper()
	b, err := os.ReadFile(calls)
	if err != nil {
		return ""
	}
	return string(b)
}

func call(t *testing.T, h gin.HandlerFunc, body string) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h(c)
	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func TestEntryPointsRefuseWhileNetBirdRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	s := NewService()
	for name, h := range map[string]gin.HandlerFunc{
		"install": s.Install, "start": s.Start, "up": s.Up, "login": s.Login,
	} {
		rsp := call(t, h, "")
		if rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird is running or starts at boot") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	if rsp := call(t, s.Boot, `{"enabled":true}`); rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird") {
		t.Fatalf("boot: got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
	if addon.BootEnabled(addon.Tailscale) {
		t.Fatal("start at boot must stay off")
	}
}

func TestEntryPointsRefuseWhileNetBirdStartsAtBoot(t *testing.T) {
	calls := lifecycle(t)
	if err := os.WriteFile(filepath.Join(addon.InitdDir, "S98netbird"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if rsp := call(t, NewService().Start, ""); rsp.Code == 0 || !strings.Contains(rsp.Msg, "NetBird") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestBootOffIsAllowedWhileBlocked(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Boot, `{"enabled":false}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}

func TestStartAndStopLeaveStartAtBootAlone(t *testing.T) {
	calls := lifecycle(t)
	s := NewService()
	if rsp := call(t, s.Start, ""); rsp.Code != 0 {
		t.Fatalf("start: %q", rsp.Msg)
	}
	if addon.BootEnabled(addon.Tailscale) {
		t.Fatal("start must not turn on start at boot")
	}
	if rsp := call(t, s.Boot, `{"enabled":true}`); rsp.Code != 0 {
		t.Fatalf("boot: %q", rsp.Msg)
	}
	if rsp := call(t, s.Stop, ""); rsp.Code != 0 {
		t.Fatalf("stop: %q", rsp.Msg)
	}
	if !addon.BootEnabled(addon.Tailscale) {
		t.Fatal("stop must not turn off start at boot")
	}
	if got := callsOf(t, calls); got != "script start\nscript stop\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestStartFailureCarriesTheScriptOutput(t *testing.T) {
	lifecycle(t)
	t.Setenv("STUB_RESULT", "FAIL")
	rsp := call(t, NewService().Start, "")
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "Starting tailscaled[1.2.3]: FAIL") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}

func TestUpFailureCarriesTheCLIOutput(t *testing.T) {
	lifecycle(t)
	t.Setenv("STUB_FAIL", "1")
	rsp := call(t, NewService().Up, "")
	if rsp.Code != -1 || !strings.Contains(rsp.Msg, "backend error: tailscaled is not running") {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
}

func TestStartNoLongerWritesGoMemLimit(t *testing.T) {
	lifecycle(t)
	if rsp := call(t, NewService().Start, ""); rsp.Code != 0 {
		t.Fatalf("start: %q", rsp.Msg)
	}
	if utils.IsGoMemLimitExist() {
		t.Fatal("S98tailscaled derives the limit; the server must not write it")
	}
}
```

Delete `TestRecordEnabledOnlyOnData` (the whole function, lines 95 to 117) from `server/service/extensions/tailscale/addon_test.go`. `recordEnabled` is gone; `addon.SetBoot` has its own tests.

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/tailscale/ -run 'TestEntryPoints|TestBootOff|TestStartAndStop|TestStartFailure|TestUpFailure|TestStartNoLonger' -v`
Expected: FAIL to compile, `s.Boot undefined`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/vpn/handlers.go`:

```go
package vpn

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"

	"github.com/gin-gonic/gin"
)

// Refuse answers the request with the reason when the other VPN runs or
// starts at boot, and reports whether it did. Install, start, up, login and
// boot on both add-ons call it before they do anything.
func Refuse(c *gin.Context, d addon.Daemon) bool {
	err := addon.CheckExclusive(d.Name)
	if err == nil {
		return false
	}
	var rsp proto.Response
	rsp.ErrRsp(c, -1, err.Error())
	return true
}

// Boot handles POST boot for either add-on: {enabled} turns start at boot on
// or off. Turning it on is refused while the other VPN runs or starts at boot;
// turning it off never is.
func Boot(c *gin.Context, d addon.Daemon, installed bool) {
	var req proto.VpnBootReq
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if req.Enabled {
		if !installed {
			rsp.ErrRsp(c, -1, d.Title+" is not installed")
			return
		}
		if Refuse(c, d) {
			return
		}
	}
	if err := addon.SetBoot(d, req.Enabled); err != nil {
		rsp.ErrRsp(c, -1, Message("start at boot failed", err))
		return
	}
	rsp.OkRsp(c)
}
```

`server/service/extensions/tailscale/cli.go` (whole file):

```go
package tailscale

import (
	"os/exec"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
)

// loginURLTimeout bounds how long the handler waits for the login URL. The
// command itself runs for ten minutes waiting on the browser.
const loginURLTimeout = 60 * time.Second

type Cli struct{}

func NewCli() *Cli {
	return &Cli{}
}

// Start runs the boot script from the package copy. It no longer touches
// start at boot, which is the boot switch's, through addon.SetBoot.
func (c *Cli) Start() error {
	for _, filePath := range []string{TailscalePath, TailscaledPath} {
		if err := utils.EnsurePermission(filePath, 0o100); err != nil {
			return err
		}
	}
	return vpn.Script(addon.Tailscale.Script(), "start", "")
}

func (c *Cli) Restart() error {
	return vpn.Script(addon.Tailscale.Script(), "restart", "")
}

func (c *Cli) Stop() error {
	return vpn.Script(addon.Tailscale.Script(), "stop", "")
}

func (c *Cli) Up() error {
	_, err := vpn.Run(exec.Command(TailscalePath, "up", "--accept-dns=false"))
	return err
}

func (c *Cli) Down() error {
	_, err := vpn.Run(exec.Command(TailscalePath, "down"))
	return err
}

func (c *Cli) Status() (*TsStatus, error) {
	output, err := exec.Command(TailscalePath, "status", "--json").CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseStatus(output)
}

func (c *Cli) Login() (string, error) {
	// No shell: killing "sh -c tailscale ..." leaves tailscale holding the
	// stderr pipe, so the timeout could never take effect.
	cmd := exec.Command(TailscalePath, "login", "--accept-dns=false", "--timeout=10m")

	return vpn.LoginURL(cmd, false, loginURLTimeout, 0)
}

func (c *Cli) Logout() error {
	_, err := vpn.Run(exec.Command(TailscalePath, "logout"))
	return err
}
```

`server/service/extensions/tailscale/service.go` (whole file):

```go
package tailscale

import (
	"os"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Service struct{}

// Variables rather than constants so the tests can point them at a scratch
// root. On a distribution image both are links into the add-on on /data.
var (
	TailscalePath  = "/usr/bin/tailscale"
	TailscaledPath = "/usr/sbin/tailscaled"
)

// addonSpec is what Tailscale needs back from the root filesystem after a new
// image: its two binaries in their usual places and its boot script while it is
// enabled. Its login is already on /data, where S98tailscaled keeps it.
func addonSpec() addon.Spec {
	return addon.Spec{
		Name: addon.Tailscale.Name,
		Links: []addon.Link{
			{Path: "/usr/bin/tailscale", File: "tailscale"},
			{Path: "/usr/sbin/tailscaled", File: "tailscaled"},
		},
		Initd: addon.Tailscale.Initd,
	}
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	if !isInstalled() {
		if err := install(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("install failed", err))
			return
		}

		if err := NewCli().Start(); err != nil {
			log.Errorf("failed to start tailscale after install: %s", err)
		}
	}

	rsp.OkRsp(c)
	log.Debugf("install tailscale successfully")
}

func (s *Service) Uninstall(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Debugf("failed to stop tailscale before uninstall: %s", err)
	}
	if err := addon.SetBoot(addon.Tailscale, false); err != nil {
		log.Errorf("failed to turn off tailscale at boot: %s", err)
	}

	if addon.OnData() {
		if err := addon.Remove(addonSpec()); err != nil {
			log.Errorf("failed to remove the tailscale add-on: %s", err)
		}
	}
	_ = os.Remove(TailscalePath)
	_ = os.Remove(TailscaledPath)

	rsp.OkRsp(c)
	log.Debugf("uninstall tailscale successfully")
}

// Start starts the daemon. Start at boot is the boot route's alone now, and
// the Go memory limit is S98tailscaled's, derived from the addons group.
func (s *Service) Start(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	if err := NewCli().Start(); err != nil {
		log.Errorf("failed to run tailscale start: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale start successfully")
}

func (s *Service) Restart(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Restart(); err != nil {
		log.Errorf("failed to run tailscale restart: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("restart failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale restart successfully")
}

func (s *Service) Stop(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Errorf("failed to run tailscale stop: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale stop successfully")
}

func (s *Service) Up(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	if err := NewCli().Up(); err != nil {
		log.Errorf("failed to run tailscale up: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("tailscale up failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("run tailscale up successfully")
}

func (s *Service) Down(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Down(); err != nil {
		log.Errorf("failed to run tailscale down: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("tailscale down failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("run tailscale down successfully")
}

func (s *Service) Login(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.Tailscale) {
		return
	}

	cli := NewCli()
	status, err := cli.Status()
	if err != nil {
		if err := cli.Start(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
			return
		}
		status, err = cli.Status()
	}

	if err != nil {
		log.Errorf("failed to get tailscale status: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("unknown status", err))
		return
	}

	if status.BackendState == "Running" {
		rsp.OkRspWithData(c, &proto.VpnLoginRsp{})
		return
	}

	url, err := cli.Login()
	if err != nil {
		log.Errorf("failed to run tailscale login: %s", err)
		rsp.ErrRsp(c, -2, vpn.Message("login failed", err))
		return
	}

	rsp.OkRspWithData(c, &proto.VpnLoginRsp{Url: url})
	log.Debugf("tailscale login url: %s", url)
}

func (s *Service) Logout(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Logout(); err != nil {
		log.Errorf("failed to run tailscale logout: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("logout failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("tailscale logout successfully")
}

func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	st := proto.VpnStatus{State: proto.VpnNotInstall}
	if isInstalled() {
		st = proto.VpnStatus{State: proto.VpnNotRunning}
		if ts, err := NewCli().Status(); err != nil {
			log.Debugf("failed to get tailscale status: %s", err)
		} else if st, err = toVpnStatus(ts); err != nil {
			log.Errorf("%s", err)
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
	}

	vpn.Fill(&st, addon.Tailscale)
	rsp.OkRspWithData(c, &st)
}

// Boot turns start at boot on or off.
func (s *Service) Boot(c *gin.Context) {
	vpn.Boot(c, addon.Tailscale, isInstalled())
}
```

In `server/router/extensions.go`, add after the restart route:

```go
	api.POST("/tailscale/boot", ts.Boot)           // tailscale start at boot on or off
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/... ./router/ -v`
Expected: PASS. `go vet` is part of `go test`; no "declared and not used" or unused imports.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/vpn/handlers.go server/service/extensions/tailscale server/router/extensions.go
git commit -m "tailscale: start at boot is its own switch, one VPN at a time, and the reason for a failure

start and stop no longer turn start at boot on and off. The server no
longer writes /etc/kvm/GOMEMLIMIT: S98tailscaled derives the limit from
the addons group."
```

---

### Task 9: Tailscale update check and update

**Files:**
- Modify: `server/service/extensions/tailscale/install.go` (`getDownloadURL` split, new `versionFromPackageURL`, `latestVersion`, `regexp` import)
- Modify: `server/service/extensions/tailscale/cli.go` (`Version`, `strings` import)
- Modify: `server/service/extensions/tailscale/service.go` (vars, `Install` uses `installPackage`, `GetUpdate`, `Update`)
- Modify: `server/router/extensions.go` (two update routes)
- Test: `server/service/extensions/tailscale/update_test.go`

**Interfaces:**
- Consumes: `vpn.VersionCache`, `vpn.UpdateTTL` (Task 6); `addon.Running` (Task 5); `lifecycle`, `call`, `callsOf`, `fakeRunning` test helpers (Task 8); `proto.VpnUpdateRsp` (Task 2).
- Produces: `resolveRedirect(client *http.Client, rawURL string) (string, error)`; `versionFromPackageURL(u string) (string, error)`; `latestVersion() (string, error)`; `(*Cli).Version() (string, error)`; package vars `updates *vpn.VersionCache`, `fetchLatest func() (string, error)`, `installPackage func() error`; `(*Service).GetUpdate`, `(*Service).Update`; routes `GET` and `POST /api/extensions/tailscale/update`.

- [ ] **Step 1: Write the failing test**

`server/service/extensions/tailscale/update_test.go`:

```go
//go:build linux

package tailscale

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
)

func TestVersionFromPackageURL(t *testing.T) {
	v, err := versionFromPackageURL("https://pkgs.tailscale.com/stable/tailscale_1.90.1_riscv64.tgz")
	if err != nil || v != "1.90.1" {
		t.Fatalf("got %q %v", v, err)
	}
	if _, err := versionFromPackageURL("https://pkgs.tailscale.com/stable/tailscale_latest_riscv64.tgz"); err == nil {
		t.Fatal("the unresolved alias has no version")
	}
}

// A fake release server: the latest alias redirects to a versioned package.
func TestResolveRedirectFollowsToThePackage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stable/tailscale_latest_riscv64.tgz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/stable/tailscale_1.90.1_riscv64.tgz", http.StatusFound)
	})
	mux.HandleFunc("/stable/tailscale_1.90.1_riscv64.tgz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("the check must not download the package, got %s", r.Method)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resolved, err := resolveRedirect(srv.Client(), srv.URL+"/stable/tailscale_latest_riscv64.tgz")
	if err != nil {
		t.Fatal(err)
	}
	if v, err := versionFromPackageURL(resolved); err != nil || v != "1.90.1" {
		t.Fatalf("got %q %v from %s", v, err, resolved)
	}
}

func TestGetUpdateReportsCurrentAndLatestAndCaches(t *testing.T) {
	lifecycle(t)
	savedCache, savedFetch := updates, fetchLatest
	t.Cleanup(func() { updates, fetchLatest = savedCache, savedFetch })
	updates = &vpn.VersionCache{TTL: vpn.UpdateTTL}
	fetches := 0
	fetchLatest = func() (string, error) { fetches++; return "1.90.1", nil }

	for i := 0; i < 2; i++ {
		rsp := call(t, NewService().GetUpdate, "")
		data, _ := rsp.Data.(map[string]any)
		if rsp.Code != 0 || data["current"] != "1.88.3" || data["latest"] != "1.90.1" {
			t.Fatalf("got %d %q %v", rsp.Code, rsp.Msg, rsp.Data)
		}
	}
	if fetches != 1 {
		t.Fatalf("the release server was asked %d times within the hour", fetches)
	}
}

func appendLine(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(line + "\n")
	return err
}

func TestUpdateRestartsOnlyADaemonThatRan(t *testing.T) {
	calls := lifecycle(t)
	savedInstall := installPackage
	t.Cleanup(func() { installPackage = savedInstall })
	installPackage = func() error { return appendLine(calls, "install") }
	s := NewService()

	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "install\n" {
		t.Fatalf("a stopped daemon must stay stopped:\n%s", got)
	}

	_ = os.Remove(calls)
	fakeRunning(t, addon.Tailscale, 4343, "/usr/sbin/tailscaled")
	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script stop\ninstall\nscript start\n" {
		t.Fatalf("a running daemon must run again:\n%s", got)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/tailscale/ -run 'TestVersionFromPackageURL|TestResolveRedirect|TestGetUpdate|TestUpdateRestarts' -v`
Expected: FAIL to compile, `undefined: versionFromPackageURL`.

- [ ] **Step 3: Write the implementation**

In `server/service/extensions/tailscale/install.go`, add `"regexp"` to the imports and replace `getDownloadURL` (the last function in the file) with:

```go
// packageVersion is the version in a resolved package name,
// tailscale_1.90.1_riscv64.tgz. The unresolved alias, latest, does not match.
var packageVersion = regexp.MustCompile(`/tailscale_([0-9][0-9A-Za-z.]*)_riscv64\.tgz$`)

func versionFromPackageURL(u string) (string, error) {
	m := packageVersion.FindStringSubmatch(u)
	if m == nil {
		return "", fmt.Errorf("no version in %q", u)
	}
	return m[1], nil
}

// resolveRedirect asks for rawURL with a HEAD and returns where the redirects
// ended. A HEAD is enough: a GET would pull the whole archive only to discard
// it and fetch it again.
func resolveRedirect(client *http.Client, rawURL string) (string, error) {
	resp, err := client.Head(rawURL)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Request.URL.String(), nil
}

// getDownloadURL resolves the "latest" alias to the versioned package, and
// keeps it on the release host.
func getDownloadURL() (string, error) {
	resolved, err := resolveRedirect(utils.OutboundClient(checksumTimeout), OriginalURL)
	if err != nil {
		return "", err
	}

	if err := checkDownloadHost(resolved); err != nil {
		return "", err
	}

	return resolved, nil
}

// latestVersion is the version the release server's "latest" alias points at.
func latestVersion() (string, error) {
	resolved, err := getDownloadURL()
	if err != nil {
		return "", err
	}
	return versionFromPackageURL(resolved)
}
```

In `server/service/extensions/tailscale/cli.go`, add `"strings"` to the imports and add:

```go
// Version is the installed CLI's version: the first line of `tailscale version`.
func (c *Cli) Version() (string, error) {
	out, err := vpn.Run(exec.Command(TailscalePath, "version"))
	if err != nil {
		return "", err
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(first), nil
}
```

In `server/service/extensions/tailscale/service.go`, add below the path vars:

```go
// The update check's cache, the lookup and the installer. Variables so the
// tests need no network.
var (
	updates        = &vpn.VersionCache{TTL: vpn.UpdateTTL}
	fetchLatest    = latestVersion
	installPackage = install
)
```

In `Install`, replace `if err := install(); err != nil {` with `if err := installPackage(); err != nil {`. Then add at the end of the file:

```go
// GetUpdate reports the installed version and the latest one, which is
// looked up at most once an hour.
func (s *Service) GetUpdate(c *gin.Context) {
	var rsp proto.Response

	if !isInstalled() {
		rsp.ErrRsp(c, -1, "tailscale is not installed")
		return
	}

	current, err := NewCli().Version()
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("version failed", err))
		return
	}

	latest, err := updates.Latest(fetchLatest)
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("update check failed", err))
		return
	}

	rsp.OkRspWithData(c, &proto.VpnUpdateRsp{Current: current, Latest: latest})
}

// Update installs the latest release over the installed one. The login stays
// on /data, and the daemon runs again afterwards only if it ran before.
func (s *Service) Update(c *gin.Context) {
	var rsp proto.Response

	if !isInstalled() {
		rsp.ErrRsp(c, -1, "tailscale is not installed")
		return
	}

	cli := NewCli()
	wasRunning := addon.Running(addon.Tailscale)
	if wasRunning {
		if err := cli.Stop(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
			return
		}
	}

	err := installPackage()
	updates.Reset()

	if wasRunning {
		if startErr := cli.Start(); startErr != nil && err == nil {
			err = startErr
		}
	}
	if err != nil {
		log.Errorf("failed to update tailscale: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("update failed", err))
		return
	}

	rsp.OkRsp(c)
	log.Debugf("update tailscale successfully")
}
```

In `server/router/extensions.go`, add after the boot route:

```go
	api.GET("/tailscale/update", ts.GetUpdate)     // tailscale current and latest version
	api.POST("/tailscale/update", ts.Update)       // install the latest tailscale
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/... ./router/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/tailscale server/router/extensions.go
git commit -m "tailscale: check for and install the latest release"
```

---

### Task 10: NetBird install from Alpine's package

**Files:**
- Create: `server/service/extensions/netbird/install.go`
- Test: `server/service/extensions/netbird/helpers_test.go`, `server/service/extensions/netbird/install_test.go`

**Interfaces:**
- Consumes: `addon.OnData`, `addon.Dir`, `addon.Record`, `addon.Spec`, `addon.Link`, `addon.NetBird` (Task 5); `vpn.Run`, `vpn.Message` (Task 3); `utils.MoveFile`.
- Produces: vars `NetbirdPath = "/usr/bin/netbird"`, `ApkPath = "apk"`, `FallbackWorkspace = "/root/.netbird-fetch"`; `addonSpec() addon.Spec`; `isInstalled() bool`; `install() error`; `latestVersion() (string, error)`; `versionFromSearch(out []byte) (string, error)`. Test helpers in package netbird: `scratchImage(t, onDistro) (fsroot string)`, `stub`, `fakeApk(t, version) (calls string)`, `callsOf`.

- [ ] **Step 1: Write the failing test**

`server/service/extensions/netbird/helpers_test.go`:

```go
//go:build linux

package netbird

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
)

// scratchImage points the add-on package and NetbirdPath at a temporary root.
// onDistro decides whether it looks like a distribution image with /data
// mounted. Off one, the fetch goes to a scratch FallbackWorkspace.
func scratchImage(t *testing.T, onDistro bool) (fsroot string) {
	t.Helper()
	base := t.TempDir()
	fsroot = filepath.Join(base, "root")
	data := filepath.Join(base, "data")
	for _, d := range []string{fsroot + "/usr/bin", data} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(base, "deviceinfo")
	if onDistro {
		if err := os.WriteFile(marker, []byte("DEVICE=one-board\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mounts := filepath.Join(base, "mounts")
	if err := os.WriteFile(mounts, []byte("/dev/x "+data+" exfat rw 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	saved := struct{ root, dataDir, marker, mounts, fsroot, nb, ws string }{
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot, NetbirdPath, FallbackWorkspace}
	t.Cleanup(func() {
		addon.Root, addon.DataDir, addon.DistroMarker, addon.Mounts, addon.FsRoot = saved.root, saved.dataDir, saved.marker, saved.mounts, saved.fsroot
		NetbirdPath, FallbackWorkspace = saved.nb, saved.ws
	})
	addon.Root = filepath.Join(data, "ironkvm", "addons")
	addon.DataDir = data
	addon.DistroMarker = marker
	addon.Mounts = mounts
	addon.FsRoot = fsroot
	NetbirdPath = fsroot + "/usr/bin/netbird"
	FallbackWorkspace = filepath.Join(base, "fetch")
	return fsroot
}

func stub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func callsOf(t *testing.T, calls string) string {
	t.Helper()
	b, err := os.ReadFile(calls)
	if err != nil {
		return ""
	}
	return string(b)
}

// fakeApk puts a stub in place of apk-tools 3. It answers update, search,
// fetch, verify and extract, builds the package it fetches with tar, and
// appends every call to the file it returns. BAD_SIGNATURE in the environment
// makes verify and extract refuse the package, as apk does a tampered one.
func fakeApk(t *testing.T, version string) (calls string) {
	t.Helper()
	dir := t.TempDir()
	calls = filepath.Join(dir, "calls")
	src := filepath.Join(dir, "src")
	path := filepath.Join(dir, "apk")
	stub(t, path, fmt.Sprintf(`echo "$*" >> %q
case "$1" in
update) echo "OK: 25418 distinct packages available" ;;
search) echo "netbird-%[2]s-r0" ;;
fetch)
	mkdir -p %[3]q/usr/bin
	printf 'netbird %[2]s\n' > %[3]q/usr/bin/netbird
	tar -czf "$3/netbird-%[2]s-r0.apk" -C %[3]q usr
	echo "Downloading netbird-%[2]s-r0" ;;
verify) [ -z "$BAD_SIGNATURE" ] || { echo "$2: UNTRUSTED signature" >&2; exit 1; } ;;
extract)
	[ -z "$BAD_SIGNATURE" ] || { echo "ERROR: $5: UNTRUSTED signature" >&2; exit 1; }
	tar -xzf "$5" -C "$4" ;;
*) echo "unexpected: $*" >&2; exit 2 ;;
esac`, calls, version, src))

	saved := ApkPath
	t.Cleanup(func() { ApkPath = saved })
	ApkPath = path
	return calls
}
```

`server/service/extensions/netbird/install_test.go`:

```go
//go:build linux

package netbird

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
)

func TestInstallOnADistributionImagePutsTheBinaryOnData(t *testing.T) {
	fsroot := scratchImage(t, true)
	calls := fakeApk(t, "0.79.0")
	if err := install(); err != nil {
		t.Fatal(err)
	}

	dir := addon.Dir("netbird")
	bin := filepath.Join(dir, "netbird")
	if b, err := os.ReadFile(bin); err != nil || string(b) != "netbird 0.79.0\n" {
		t.Fatalf("the binary is not on /data: %q %v", b, err)
	}
	if fi, err := os.Stat(bin); err != nil || fi.Mode().Perm()&0o111 == 0 {
		t.Fatal("the binary must be executable")
	}
	if target, err := os.Readlink(fsroot + "/usr/bin/netbird"); err != nil || target != bin {
		t.Fatalf("/usr/bin/netbird is not a link into the add-on: %q %v", target, err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "initd")); err != nil || string(b) != "S98netbird\n" {
		t.Fatalf("initd is %q (%v)", b, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".fetch")); err == nil {
		t.Fatal("the fetched package must not be left on /data")
	}
	if !isInstalled() {
		t.Fatal("isInstalled must follow the link and answer true")
	}

	got := callsOf(t, calls)
	last := -1
	for _, want := range []string{"update\n", "fetch -o ", "verify "} {
		i := strings.Index(got, want)
		if i <= last {
			t.Fatalf("want %q after the previous step in:\n%s", want, got)
		}
		last = i
	}
}

func TestInstallElsewhereUsesUsrBin(t *testing.T) {
	fsroot := scratchImage(t, false)
	fakeApk(t, "0.79.0")
	if err := install(); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(fsroot + "/usr/bin/netbird")
	if err != nil || !fi.Mode().IsRegular() {
		t.Fatalf("/usr/bin/netbird must be the binary itself off a distribution image: %v", err)
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("nothing may be written to /data off a distribution image")
	}
	if _, err := os.Stat(FallbackWorkspace); err == nil {
		t.Fatal("the workspace must be removed")
	}
}

func TestInstallRefusesAnUnsignedPackage(t *testing.T) {
	fsroot := scratchImage(t, true)
	fakeApk(t, "0.79.0")
	t.Setenv("BAD_SIGNATURE", "1")

	err := install()
	if err == nil || !strings.Contains(vpn.Message("install failed", err), "UNTRUSTED signature") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Lstat(fsroot + "/usr/bin/netbird"); err == nil {
		t.Fatal("an unsigned package must not be installed")
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("a failed first install must leave no add-on directory behind")
	}
}

// The vendor firmware has no apk. NetBird needs an IronKVM image.
func TestInstallWithoutApkSaysWhy(t *testing.T) {
	scratchImage(t, true)
	saved := ApkPath
	t.Cleanup(func() { ApkPath = saved })
	ApkPath = filepath.Join(t.TempDir(), "no-apk")

	err := install()
	if err == nil || !strings.Contains(err.Error(), "needs an IronKVM image") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(addon.Dir("netbird")); err == nil {
		t.Fatal("nothing may be left behind")
	}
}

func TestVersionFromSearch(t *testing.T) {
	for out, want := range map[string]string{
		"netbird-0.78.2-r0\n":                           "0.78.2",
		"netbird-openrc-0.78.2-r0\nnetbird-0.78.2-r1\n": "0.78.2",
		"netbird-0.79.0_rc1-r0":                         "0.79.0_rc1",
	} {
		if got, err := versionFromSearch([]byte(out)); err != nil || got != want {
			t.Fatalf("%q: got %q %v, want %q", out, got, err, want)
		}
	}
	if _, err := versionFromSearch([]byte("")); err == nil {
		t.Fatal("no package must be an error")
	}
}

func TestLatestVersionAsksApk(t *testing.T) {
	scratchImage(t, true)
	calls := fakeApk(t, "0.79.0")
	v, err := latestVersion()
	if err != nil || v != "0.79.0" {
		t.Fatalf("got %q %v", v, err)
	}
	if got := callsOf(t, calls); got != "update\nsearch -e netbird\n" {
		t.Fatalf("calls:\n%s", got)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/netbird/ -v`
Expected: FAIL to compile, `undefined: install`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/netbird/install.go` (method A from Task 1; the B and C changes follow the code):

```go
package netbird

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"

	log "github.com/sirupsen/logrus"
)

// Variables rather than constants so the tests can point them at a scratch
// root and a stub package manager. On a distribution image NetbirdPath is a
// link into the add-on on /data.
var (
	NetbirdPath = "/usr/bin/netbird"
	ApkPath     = "apk"
	// FallbackWorkspace is where the package is fetched off a distribution
	// image. On one it is the add-on's directory on /data: /tmp is a 79 MB
	// tmpfs and the binary alone is 43 MB.
	FallbackWorkspace = "/root/.netbird-fetch"
)

const (
	packageName = "netbird"
	apkTimeout  = 10 * time.Minute
)

// addonSpec is what NetBird needs back from the root filesystem after a new
// image: its binary in /usr/bin and its boot script while it is enabled. Its
// identity is already on /data, where S98netbird keeps it.
func addonSpec() addon.Spec {
	return addon.Spec{
		Name:  addon.NetBird.Name,
		Links: []addon.Link{{Path: "/usr/bin/netbird", File: "netbird"}},
		Initd: addon.NetBird.Initd,
	}
}

func isInstalled() bool {
	_, err := os.Stat(NetbirdPath)
	return err == nil
}

func workspace() string {
	if addon.OnData() {
		return filepath.Join(addon.Dir(addon.NetBird.Name), ".fetch")
	}
	return FallbackWorkspace
}

// apk runs the package manager with a deadline. A failure carries its output.
func apk(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apkTimeout)
	defer cancel()
	return vpn.Run(exec.CommandContext(ctx, ApkPath, args...))
}

// install fetches Alpine's netbird package, checks its signature, and takes
// usr/bin/netbird out of it. Upstream publishes no riscv64 build; Alpine v3.24
// community does. Nothing is installed into the root filesystem: the binary
// goes to the add-on's directory on /data, linked from /usr/bin, or to
// /usr/bin itself off a distribution image.
func install() error {
	if _, err := exec.LookPath(ApkPath); err != nil {
		return errors.New("apk not found: NetBird installs from Alpine's packages and needs an IronKVM image")
	}

	ws := workspace()
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(ws)
		if addon.OnData() {
			// Only removes the add-on's directory when a failed first
			// install left it empty.
			_ = os.Remove(addon.Dir(addon.NetBird.Name))
		}
	}()

	if _, err := apk("update"); err != nil {
		return err
	}
	if _, err := apk("fetch", "-o", ws, packageName); err != nil {
		return err
	}
	pkg, err := fetchedPackage(ws)
	if err != nil {
		return err
	}
	if _, err := apk("verify", pkg); err != nil {
		return err
	}
	root := filepath.Join(ws, "root")
	if err := extractPackage(pkg, root); err != nil {
		return err
	}
	if err := placeBinary(filepath.Join(root, "usr", "bin", "netbird")); err != nil {
		return err
	}
	log.Debugf("install netbird from %s", filepath.Base(pkg))
	return nil
}

// fetchedPackage is the one netbird-<version>.apk that apk fetch wrote.
func fetchedPackage(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, packageName+"-[0-9]*.apk"))
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("apk fetch left %d netbird packages in %s", len(matches), dir)
	}
	return matches[0], nil
}

// extractPackage unpacks the package into dir. apk extract checks the
// signature again and reads both package formats. --no-chown because /data is
// exFAT, which has no owners.
func extractPackage(pkg, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	_, err := apk("extract", "--no-chown", "--destination", dir, pkg)
	return err
}

// placeBinary moves the extracted binary into place and, on a distribution
// image, records the add-on so S04addons puts the link back after a new image.
func placeBinary(src string) error {
	fi, err := os.Stat(src)
	if err != nil || !fi.Mode().IsRegular() {
		return errors.New("the netbird package has no usr/bin/netbird")
	}
	if err := os.Chmod(src, 0o755); err != nil {
		return err
	}

	dst := NetbirdPath
	onData := addon.OnData()
	if onData {
		dst = filepath.Join(addon.Dir(addon.NetBird.Name), "netbird")
	}
	if err := utils.MoveFile(src, dst); err != nil {
		return err
	}
	if onData {
		return addon.Record(addonSpec())
	}
	return nil
}

// latestVersion refreshes the index and reads the version of the netbird
// package it offers.
func latestVersion() (string, error) {
	if _, err := apk("update"); err != nil {
		return "", err
	}
	out, err := apk("search", "-e", packageName)
	if err != nil {
		return "", err
	}
	return versionFromSearch(out)
}

// versionFromSearch reads "netbird-0.78.2-r0" and returns "0.78.2". Names
// such as netbird-openrc-... are other packages and are skipped.
func versionFromSearch(out []byte) (string, error) {
	for _, word := range strings.Fields(string(out)) {
		rest, ok := strings.CutPrefix(word, packageName+"-")
		if !ok || rest == "" || rest[0] < '0' || rest[0] > '9' {
			continue
		}
		if i := strings.LastIndex(rest, "-r"); i > 0 {
			rest = rest[:i]
		}
		return rest, nil
	}
	return "", errors.New("apk offers no netbird package")
}
```

If Task 1 chose **B**, replace `extractPackage` with the tar variant (the `apk verify` call stays and is the signature check):

```go
// extractPackage unpacks the binary with tar: Alpine v3.24 still ships the v2
// package format, which is a gzip stream of tar segments. apk verify has
// already checked the signature.
func extractPackage(pkg, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	_, err := vpn.Run(exec.Command("tar", "-xzf", pkg, "-C", dir, "usr/bin/netbird"))
	return err
}
```

If Task 1 chose **C**, delete these three lines from `install` (apk extract is then the signature check):

```go
	if _, err := apk("verify", pkg); err != nil {
		return err
	}
```

and in `TestInstallOnADistributionImagePutsTheBinaryOnData` replace `"verify "` in the order list with `"extract "`.

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/netbird/ -v`
Expected: PASS, 6 tests.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/netbird/install.go server/service/extensions/netbird/helpers_test.go server/service/extensions/netbird/install_test.go
git commit -m "netbird: install the binary from Alpine's signed package onto /data"
```

---

### Task 11: NetBird CLI and status

**Files:**
- Create: `server/service/extensions/netbird/cli.go`
- Create: `server/service/extensions/netbird/status.go`
- Test: `server/service/extensions/netbird/status_test.go`, `server/service/extensions/netbird/cli_test.go`

**Interfaces:**
- Consumes: `NetbirdPath`, `scratchImage`, `stub` (Task 10); `vpn.Run`, `vpn.Script`, `vpn.Message`, `vpn.CmdError` (Task 3); `vpn.LoginURL` (Task 4); `addon.NetBird` (Task 5); fixtures (Task 1).
- Produces: `const LogFile = "/var/log/netbird.log"`; var `statusTimeout = 15 * time.Second`; `type Cli` with `NewCli()`, `Start()`, `Restart()`, `Stop()`, `Up()`, `Down()`, `Deregister()`, `Status() (*NbStatus, error)`, `Version() (string, error)`, `JoinWithSetupKey(key string) error`, `LoginSSO() (string, error)`; `type NbPeer`, `type NbStatus`, `StateMap map[string]proto.VpnState`, `parseStatus(out []byte) (*NbStatus, error)`, `toVpnStatus(nb *NbStatus) (proto.VpnStatus, error)`.

- [ ] **Step 1: Write the failing tests**

`server/service/extensions/netbird/status_test.go`:

```go
package netbird

import (
	"os"
	"reflect"
	"testing"

	"NanoKVM-Server/proto"
)

func load(t *testing.T, name string) *NbStatus {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	nb, err := parseStatus(b)
	if err != nil {
		t.Fatal(err)
	}
	return nb
}

// The JSON recorded on the board on 2026-09-27.
func TestParseConnectedStatusFromTheBoard(t *testing.T) {
	got, err := toVpnStatus(load(t, "status-connected.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := proto.VpnStatus{
		State:   proto.VpnRunning,
		Version: "0.78.2",
		IP:      "100.73.212.105",
		Name:    "ironkvm.netbird.cloud",
		Account: "api.netbird.io",
		Control: true,
		Peers:   []proto.VpnPeer{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestParsePeers(t *testing.T) {
	got, err := toVpnStatus(load(t, "status-peers.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := []proto.VpnPeer{
		{Name: "laptop", IP: "100.73.10.20", Online: true},
		{Name: "phone", IP: "100.73.10.21", Online: false},
	}
	if !reflect.DeepEqual(got.Peers, want) {
		t.Fatalf("got %+v", got.Peers)
	}
}

func TestParseNeedsLogin(t *testing.T) {
	got, err := toVpnStatus(load(t, "status-needslogin.json"))
	if err != nil || got.State != proto.VpnNotLogin {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestDaemonStatuses(t *testing.T) {
	for status, want := range map[string]proto.VpnState{
		"Idle":           proto.VpnStopped,
		"Connecting":     proto.VpnRunning,
		"Connected":      proto.VpnRunning,
		"NeedsLogin":     proto.VpnNotLogin,
		"LoginFailed":    proto.VpnNotLogin,
		"SessionExpired": proto.VpnNotLogin,
	} {
		got, err := toVpnStatus(&NbStatus{DaemonStatus: status})
		if err != nil || got.State != want {
			t.Fatalf("%s: got %q %v, want %q", status, got.State, err, want)
		}
	}
	if _, err := toVpnStatus(&NbStatus{DaemonStatus: "Exploded"}); err == nil {
		t.Fatal("an unknown daemon status must be an error")
	}
}

func TestParseStatusRejectsGarbage(t *testing.T) {
	if _, err := parseStatus([]byte("Error: failed to connect to daemon")); err == nil {
		t.Fatal("output with no JSON must be an error")
	}
}
```

`server/service/extensions/netbird/cli_test.go`:

```go
//go:build linux

package netbird

import (
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/vpn"
)

// A CLI that echoes the key in its error must not put it on the page.
func TestSetupKeyNeverReachesTheMessage(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `echo "login failed: invalid setup key $3" >&2; exit 1`)

	err := NewCli().JoinWithSetupKey("SECRET-KEY-123")
	msg := vpn.Message("join failed", err)
	if err == nil || strings.Contains(msg, "SECRET-KEY-123") || !strings.Contains(msg, "invalid setup key ***") {
		t.Fatalf("message is %q", msg)
	}
	if strings.Contains(err.Error(), "SECRET-KEY-123") {
		t.Fatalf("the error must not carry the key either: %q", err.Error())
	}
}

func TestLoginSSOReadsTheURLFromStdout(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `[ "$1 $2" = "up --no-browser" ] || exit 9
printf 'Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n'`)

	url, err := NewCli().LoginSSO()
	if err != nil || url != "https://login.netbird.io/activate?user_code=ABCD-EFGH" {
		t.Fatalf("got %q %v", url, err)
	}
}

func TestVersion(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `[ "$1" = version ] && echo 0.78.2`)
	if v, err := NewCli().Version(); err != nil || v != "0.78.2" {
		t.Fatalf("got %q %v", v, err)
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `gotest ./service/extensions/netbird/ -v`
Expected: FAIL to compile, `undefined: parseStatus`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/netbird/status.go`:

```go
package netbird

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"NanoKVM-Server/proto"
)

// NbPeer is one entry of peers.details in `netbird status --json`. The names
// are PeerStateDetailOutput's in client/status/status.go at v0.78.2.
type NbPeer struct {
	FQDN   string `json:"fqdn"`
	IP     string `json:"netbirdIp"`
	Status string `json:"status"` // Idle, Connecting, Connected
}

// NbStatus is the part of `netbird status --json` the page shows.
type NbStatus struct {
	Peers struct {
		Details []NbPeer `json:"details"`
	} `json:"peers"`
	DaemonVersion string `json:"daemonVersion"`
	DaemonStatus  string `json:"daemonStatus"`
	Management    struct {
		URL       string `json:"url"`
		Connected bool   `json:"connected"`
	} `json:"management"`
	IP   string `json:"netbirdIp"`
	FQDN string `json:"fqdn"`
}

// StateMap maps the daemon statuses of client/status at v0.78.2. A fresh
// daemon with no login reports NeedsLogin; `netbird down` leaves it Idle.
var StateMap = map[string]proto.VpnState{
	"Idle":           proto.VpnStopped,
	"Connecting":     proto.VpnRunning,
	"Connected":      proto.VpnRunning,
	"NeedsLogin":     proto.VpnNotLogin,
	"LoginFailed":    proto.VpnNotLogin,
	"SessionExpired": proto.VpnNotLogin,
}

// parseStatus reads the CLI's JSON, skipping anything printed before it.
func parseStatus(out []byte) (*NbStatus, error) {
	i := bytes.IndexByte(out, '{')
	if i < 0 {
		return nil, errors.New("unknown output")
	}
	var st NbStatus
	if err := json.Unmarshal(out[i:], &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// toVpnStatus is the page's view of the CLI's status. Uptime, memory, start at
// boot and the blocker are vpn.Fill's.
func toVpnStatus(nb *NbStatus) (proto.VpnStatus, error) {
	state, ok := StateMap[nb.DaemonStatus]
	if !ok {
		return proto.VpnStatus{}, fmt.Errorf("unknown netbird status: %s", nb.DaemonStatus)
	}
	st := proto.VpnStatus{
		State:   state,
		Version: nb.DaemonVersion,
		IP:      stripPrefix(nb.IP),
		Name:    nb.FQDN,
		Account: managementHost(nb.Management.URL),
		Control: nb.Management.Connected,
		Peers:   []proto.VpnPeer{},
	}
	for _, p := range nb.Peers.Details {
		st.Peers = append(st.Peers, proto.VpnPeer{
			Name:   peerName(p.FQDN),
			IP:     stripPrefix(p.IP),
			Online: p.Status == "Connected",
		})
	}
	sort.Slice(st.Peers, func(i, j int) bool { return st.Peers[i].Name < st.Peers[j].Name })
	return st, nil
}

// stripPrefix drops the network length: 100.73.212.105/16 is 100.73.212.105.
func stripPrefix(ip string) string {
	addr, _, _ := strings.Cut(ip, "/")
	return addr
}

// managementHost is the account the page shows: the management server's host.
func managementHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return raw
	}
	return u.Hostname()
}

// peerName is the first label of the peer's FQDN.
func peerName(fqdn string) string {
	if label, _, _ := strings.Cut(fqdn, "."); label != "" {
		return label
	}
	return fqdn
}
```

`server/service/extensions/netbird/cli.go`:

```go
package netbird

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/utils"
)

// LogFile is where S98netbird points the daemon's log. A start that fails
// reports its last lines.
const LogFile = "/var/log/netbird.log"

const (
	// loginURLTimeout bounds how long the handler waits for the SSO URL.
	loginURLTimeout = 60 * time.Second
	// ssoLife bounds an SSO login nobody finishes: netbird up has no timeout
	// of its own and would wait for the browser forever.
	ssoLife = 10 * time.Minute
	// upTimeout bounds up and a join with a setup key, which talk to the
	// management server.
	upTimeout = 2 * time.Minute
	// cliTimeout bounds down, deregister and version.
	cliTimeout = 30 * time.Second
	// waitDelay bounds how long a command's pipes may outlive it.
	waitDelay = 2 * time.Second
)

// statusTimeout bounds `netbird status`, which the page runs on every visit.
// A variable for the tests.
var statusTimeout = 15 * time.Second

type Cli struct{}

func NewCli() *Cli {
	return &Cli{}
}

// run runs the CLI with a deadline. WaitDelay closes the pipes if the CLI
// leaves a child holding them, so the deadline holds.
func run(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, NetbirdPath, args...)
	cmd.WaitDelay = waitDelay
	return vpn.Run(cmd)
}

// Start runs the boot script from the package copy. Start at boot is the boot
// route's alone.
func (c *Cli) Start() error {
	if err := utils.EnsurePermission(NetbirdPath, 0o100); err != nil {
		return err
	}
	return vpn.Script(addon.NetBird.Script(), "start", LogFile)
}

func (c *Cli) Restart() error {
	return vpn.Script(addon.NetBird.Script(), "restart", LogFile)
}

func (c *Cli) Stop() error {
	return vpn.Script(addon.NetBird.Script(), "stop", "")
}

func (c *Cli) Up() error {
	_, err := run(upTimeout, "up")
	return err
}

func (c *Cli) Down() error {
	_, err := run(cliTimeout, "down")
	return err
}

// Deregister removes this peer from the NetBird account and deletes its
// configuration. It is NetBird's logout; joining again needs a new key or SSO.
func (c *Cli) Deregister() error {
	_, err := run(cliTimeout, "deregister")
	return err
}

func (c *Cli) Status() (*NbStatus, error) {
	out, err := run(statusTimeout, "status", "--json")
	if err != nil {
		return nil, err
	}
	return parseStatus(out)
}

// Version is the installed CLI's version, the first line of `netbird version`.
func (c *Cli) Version() (string, error) {
	out, err := run(cliTimeout, "version")
	if err != nil {
		return "", err
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(first), nil
}

// JoinWithSetupKey runs `netbird up --setup-key`. The key is never logged,
// and a CLI that echoes it has it replaced in the error, which reaches the
// page and the server log.
func (c *Cli) JoinWithSetupKey(key string) error {
	_, err := run(upTimeout, "up", "--setup-key", key)
	return redact(err, key)
}

// LoginSSO starts an SSO login and returns the URL. NetBird prints it on
// stdout with --no-browser.
func (c *Cli) LoginSSO() (string, error) {
	cmd := exec.Command(NetbirdPath, "up", "--no-browser")
	return vpn.LoginURL(cmd, true, loginURLTimeout, ssoLife)
}

func redact(err error, secret string) error {
	if err == nil || secret == "" {
		return err
	}
	var ce *vpn.CmdError
	if errors.As(err, &ce) {
		ce.Tail = strings.ReplaceAll(ce.Tail, secret, "***")
		ce.Err = errors.New(strings.ReplaceAll(ce.Err.Error(), secret, "***"))
		return ce
	}
	return errors.New(strings.ReplaceAll(err.Error(), secret, "***"))
}
```

- [ ] **Step 4: Run the tests**

Run: `gotest ./service/extensions/netbird/ -v`
Expected: PASS, the Task 10 tests plus 8 new ones.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/netbird/cli.go server/service/extensions/netbird/status.go server/service/extensions/netbird/status_test.go server/service/extensions/netbird/cli_test.go
git commit -m "netbird: the CLI and its status in the shared shape"
```

---

### Task 12: NetBird handlers and routes

**Files:**
- Create: `server/service/extensions/netbird/service.go`
- Modify: `server/router/extensions.go` (whole file)
- Test: `server/service/extensions/netbird/service_test.go`

**Interfaces:**
- Consumes: everything from Tasks 10 and 11; `vpn.Refuse`, `vpn.Boot` (Task 8); `vpn.Fill`, `vpn.VersionCache`, `vpn.UpdateTTL` (Task 6); `addon.Running`, `addon.SetBoot`, `addon.Remove`, `addon.OnData` (Task 5 and existing); `proto.NetbirdLoginReq`, `proto.VpnLoginRsp`, `proto.VpnUpdateRsp` (Task 2).
- Produces: `netbird.NewService()` with handlers `Install`, `Uninstall`, `GetStatus`, `Start`, `Stop`, `Restart`, `Up`, `Down`, `Login`, `Logout`, `Boot`, `GetUpdate`, `Update`; package vars `updates`, `fetchLatest`, `installPackage`; routes under `/api/extensions/netbird/` with the same names as Tailscale's.

- [ ] **Step 1: Write the failing test**

`server/service/extensions/netbird/service_test.go`:

```go
//go:build linux

package netbird

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
)

// lifecycle is scratchImage off a distribution image, with a stub netbird
// that answers from the board's status fixture, a stub boot script in the
// package copy, and scratch pid files, /proc and /etc/init.d. Every stub
// appends to the file it returns. STUB_HANG makes `netbird status` hang.
func lifecycle(t *testing.T) (calls string) {
	t.Helper()
	fsroot := scratchImage(t, false)
	base := filepath.Dir(fsroot)
	calls = filepath.Join(base, "calls")
	fixture, err := filepath.Abs("testdata/status-connected.json")
	if err != nil {
		t.Fatal(err)
	}

	savedTs, savedNb := addon.Tailscale, addon.NetBird
	savedInitd, savedPkg, savedProc := addon.InitdDir, addon.PkgInitdDir, addon.ProcDir
	savedStatus, savedCache, savedFetch, savedInstall := statusTimeout, updates, fetchLatest, installPackage
	t.Cleanup(func() {
		addon.Tailscale, addon.NetBird = savedTs, savedNb
		addon.InitdDir, addon.PkgInitdDir, addon.ProcDir = savedInitd, savedPkg, savedProc
		statusTimeout, updates, fetchLatest, installPackage = savedStatus, savedCache, savedFetch, savedInstall
	})
	addon.InitdDir = filepath.Join(base, "etc-init.d")
	addon.PkgInitdDir = filepath.Join(base, "kvmapp-init.d")
	addon.ProcDir = filepath.Join(base, "proc")
	for _, d := range []string{addon.InitdDir, addon.PkgInitdDir, addon.ProcDir, filepath.Join(base, "run")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	addon.Tailscale.PidFile = filepath.Join(base, "run", "tailscaled.pid")
	addon.NetBird.PidFile = filepath.Join(base, "run", "netbird.pid")
	updates = &vpn.VersionCache{TTL: vpn.UpdateTTL}

	r := strings.NewReplacer("CALLS", calls, "FIXTURE", fixture)
	stub(t, filepath.Join(addon.PkgInitdDir, "S98netbird"), r.Replace(`echo "script $1" >> "CALLS"
case "$1" in
start) echo "Starting netbird: ${STUB_RESULT:-OK}" ;;
stop) echo "Stopping netbird: OK" ;;
esac`))
	stub(t, NetbirdPath, r.Replace(`echo "netbird $*" >> "CALLS"
case "$1" in
version) echo "0.78.2" ;;
status) [ -z "$STUB_HANG" ] || sleep 30; cat "FIXTURE" ;;
up)
	if [ "$2" = "--no-browser" ]; then
		printf 'Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n'
		exit 0
	fi
	echo "Connected" ;;
deregister) echo "Deregistered successfully" ;;
esac`))
	return calls
}

func fakeRunning(t *testing.T, d addon.Daemon, pid int, argv0 string) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(addon.ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(argv0+"\x00service\x00run\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func call(t *testing.T, h gin.HandlerFunc, body string) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h(c)
	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func data(t *testing.T, rsp proto.Response) map[string]any {
	t.Helper()
	m, ok := rsp.Data.(map[string]any)
	if rsp.Code != 0 || !ok {
		t.Fatalf("got %d %q %v", rsp.Code, rsp.Msg, rsp.Data)
	}
	return m
}

func TestEntryPointsRefuseWhileTailscaleRuns(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
	s := NewService()
	for name, c := range map[string]struct {
		h    gin.HandlerFunc
		body string
	}{
		"install": {s.Install, ""},
		"start":   {s.Start, ""},
		"up":      {s.Up, ""},
		"login":   {s.Login, `{"setupKey":"KEY-123"}`},
		"boot":    {s.Boot, `{"enabled":true}`},
	} {
		rsp := call(t, c.h, c.body)
		if rsp.Code == 0 || !strings.Contains(rsp.Msg, "Tailscale is running or starts at boot") {
			t.Fatalf("%s: got %d %q", name, rsp.Code, rsp.Msg)
		}
	}
	if got := callsOf(t, calls); got != "" {
		t.Fatalf("nothing may run while refused, ran:\n%s", got)
	}
}

func TestLoginTrimsTheSetupKey(t *testing.T) {
	calls := lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, NewService().Login, `{"setupKey":"  KEY-123 \n"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "netbird up --setup-key KEY-123\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestLoginStartsAStoppedDaemonFirst(t *testing.T) {
	calls := lifecycle(t)
	if rsp := call(t, NewService().Login, `{"setupKey":"KEY-123"}`); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script start\nnetbird up --setup-key KEY-123\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestLoginWithoutAKeyReturnsTheSSOURL(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	d := data(t, call(t, NewService().Login, ""))
	if d["url"] != "https://login.netbird.io/activate?user_code=ABCD-EFGH" {
		t.Fatalf("got %v", d)
	}
}

func TestGetStatusRunning(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	d := data(t, call(t, NewService().GetStatus, ""))
	if d["state"] != "running" || d["ip"] != "100.73.212.105" || d["account"] != "api.netbird.io" || d["blockedBy"] != "" {
		t.Fatalf("got %v", d)
	}
}

func TestStatusNotInstalledStillNamesTheBlocker(t *testing.T) {
	lifecycle(t)
	if err := os.Remove(NetbirdPath); err != nil {
		t.Fatal(err)
	}
	fakeRunning(t, addon.Tailscale, 100, "/usr/sbin/tailscaled")
	d := data(t, call(t, NewService().GetStatus, ""))
	if d["state"] != "notInstall" || d["blockedBy"] != "tailscale" {
		t.Fatalf("got %v", d)
	}
}

// With the management server unreachable the CLI can hang. The page must not.
func TestStatusDoesNotHangOnAStuckCLI(t *testing.T) {
	lifecycle(t)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	t.Setenv("STUB_HANG", "1")
	statusTimeout = 300 * time.Millisecond

	start := time.Now()
	d := data(t, call(t, NewService().GetStatus, ""))
	if d["state"] != "notRunning" {
		t.Fatalf("got %v", d)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("status took %s", elapsed)
	}
}

func TestLogoutDeregisters(t *testing.T) {
	calls := lifecycle(t)
	if rsp := call(t, NewService().Logout, ""); rsp.Code != 0 {
		t.Fatalf("got %d %q", rsp.Code, rsp.Msg)
	}
	if got := callsOf(t, calls); got != "netbird deregister\n" {
		t.Fatalf("calls:\n%s", got)
	}
}

func TestGetUpdateReportsCurrentAndLatest(t *testing.T) {
	lifecycle(t)
	fetchLatest = func() (string, error) { return "0.78.2", nil }
	d := data(t, call(t, NewService().GetUpdate, ""))
	if d["current"] != "0.78.2" || d["latest"] != "0.78.2" {
		t.Fatalf("got %v", d)
	}
}

func TestUpdateKeepsAStoppedDaemonStopped(t *testing.T) {
	calls := lifecycle(t)
	installPackage = func() error {
		f, err := os.OpenFile(calls, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString("install\n")
		return err
	}
	s := NewService()

	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "install\n" {
		t.Fatalf("a stopped daemon must stay stopped:\n%s", got)
	}

	_ = os.Remove(calls)
	fakeRunning(t, addon.NetBird, 4242, "/usr/bin/netbird")
	if rsp := call(t, s.Update, ""); rsp.Code != 0 {
		t.Fatalf("update: %q", rsp.Msg)
	}
	if got := callsOf(t, calls); got != "script stop\ninstall\nscript start\n" {
		t.Fatalf("a running daemon must run again:\n%s", got)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `gotest ./service/extensions/netbird/ -v`
Expected: FAIL to compile, `undefined: NewService`.

- [ ] **Step 3: Write the implementation**

`server/service/extensions/netbird/service.go`:

```go
package netbird

import (
	"errors"
	"io"
	"os"
	"strings"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Service struct{}

// The update check's cache, the lookup and the installer. Variables so the
// tests need neither the network nor apk.
var (
	updates        = &vpn.VersionCache{TTL: vpn.UpdateTTL}
	fetchLatest    = latestVersion
	installPackage = install
)

func NewService() *Service {
	return &Service{}
}

func (s *Service) Install(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.NetBird) {
		return
	}

	if !isInstalled() {
		if err := installPackage(); err != nil {
			log.Errorf("failed to install netbird: %s", err)
			rsp.ErrRsp(c, -1, vpn.Message("install failed", err))
			return
		}

		if err := NewCli().Start(); err != nil {
			log.Errorf("failed to start netbird after install: %s", err)
		}
	}

	rsp.OkRsp(c)
	log.Debugf("install netbird successfully")
}

// Uninstall removes the binary and the add-on. The identity on /data stays, as
// Tailscale's does, so a reinstall is the same peer.
func (s *Service) Uninstall(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Debugf("failed to stop netbird before uninstall: %s", err)
	}
	if err := addon.SetBoot(addon.NetBird, false); err != nil {
		log.Errorf("failed to turn off netbird at boot: %s", err)
	}
	if addon.OnData() {
		if err := addon.Remove(addonSpec()); err != nil {
			log.Errorf("failed to remove the netbird add-on: %s", err)
		}
	}
	_ = os.Remove(NetbirdPath)

	rsp.OkRsp(c)
	log.Debugf("uninstall netbird successfully")
}

func (s *Service) Start(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.NetBird) {
		return
	}
	if err := NewCli().Start(); err != nil {
		log.Errorf("failed to start netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Restart(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Restart(); err != nil {
		log.Errorf("failed to restart netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("restart failed", err))
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Stop(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Stop(); err != nil {
		log.Errorf("failed to stop netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Up(c *gin.Context) {
	var rsp proto.Response

	if vpn.Refuse(c, addon.NetBird) {
		return
	}
	if err := NewCli().Up(); err != nil {
		log.Errorf("failed to run netbird up: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("netbird up failed", err))
		return
	}
	rsp.OkRsp(c)
}

func (s *Service) Down(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Down(); err != nil {
		log.Errorf("failed to run netbird down: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("netbird down failed", err))
		return
	}
	rsp.OkRsp(c)
}

// Login joins with {setupKey}, or with an empty body starts an SSO login and
// returns its URL. The key is never logged.
func (s *Service) Login(c *gin.Context) {
	var rsp proto.Response
	var req proto.NetbirdLoginReq

	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if vpn.Refuse(c, addon.NetBird) {
		return
	}

	cli := NewCli()
	if !addon.Running(addon.NetBird) {
		if err := cli.Start(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("start failed", err))
			return
		}
	}

	if key := strings.TrimSpace(req.SetupKey); key != "" {
		if err := cli.JoinWithSetupKey(key); err != nil {
			log.Errorf("failed to join netbird with a setup key: %s", err)
			rsp.ErrRsp(c, -2, vpn.Message("join failed", err))
			return
		}
		rsp.OkRspWithData(c, &proto.VpnLoginRsp{})
		return
	}

	url, err := cli.LoginSSO()
	if err != nil {
		log.Errorf("failed to start the netbird sso login: %s", err)
		rsp.ErrRsp(c, -2, vpn.Message("login failed", err))
		return
	}
	rsp.OkRspWithData(c, &proto.VpnLoginRsp{Url: url})
}

// Logout is netbird deregister: it removes the peer from the account. The page
// warns before it calls this.
func (s *Service) Logout(c *gin.Context) {
	var rsp proto.Response

	if err := NewCli().Deregister(); err != nil {
		log.Errorf("failed to deregister netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("deregister failed", err))
		return
	}
	rsp.OkRsp(c)
}

// GetStatus asks the CLI only while the daemon runs; without it the CLI would
// only fail to connect.
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response

	st := proto.VpnStatus{State: proto.VpnNotInstall}
	if isInstalled() {
		st = proto.VpnStatus{State: proto.VpnNotRunning}
		if addon.Running(addon.NetBird) {
			if nb, err := NewCli().Status(); err != nil {
				log.Debugf("failed to get netbird status: %s", err)
			} else if st, err = toVpnStatus(nb); err != nil {
				log.Errorf("%s", err)
				rsp.ErrRsp(c, -1, err.Error())
				return
			}
		}
	}

	vpn.Fill(&st, addon.NetBird)
	rsp.OkRspWithData(c, &st)
}

// Boot turns start at boot on or off.
func (s *Service) Boot(c *gin.Context) {
	vpn.Boot(c, addon.NetBird, isInstalled())
}

// GetUpdate reports the installed version and the one Alpine offers, which is
// looked up at most once an hour.
func (s *Service) GetUpdate(c *gin.Context) {
	var rsp proto.Response

	if !isInstalled() {
		rsp.ErrRsp(c, -1, "netbird is not installed")
		return
	}
	current, err := NewCli().Version()
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("version failed", err))
		return
	}
	latest, err := updates.Latest(fetchLatest)
	if err != nil {
		rsp.ErrRsp(c, -1, vpn.Message("update check failed", err))
		return
	}
	rsp.OkRspWithData(c, &proto.VpnUpdateRsp{Current: current, Latest: latest})
}

// Update installs Alpine's current package over the installed binary. The
// identity stays on /data, and the daemon runs again only if it ran before.
func (s *Service) Update(c *gin.Context) {
	var rsp proto.Response

	if !isInstalled() {
		rsp.ErrRsp(c, -1, "netbird is not installed")
		return
	}

	cli := NewCli()
	wasRunning := addon.Running(addon.NetBird)
	if wasRunning {
		if err := cli.Stop(); err != nil {
			rsp.ErrRsp(c, -1, vpn.Message("stop failed", err))
			return
		}
	}

	err := installPackage()
	updates.Reset()

	if wasRunning {
		if startErr := cli.Start(); startErr != nil && err == nil {
			err = startErr
		}
	}
	if err != nil {
		log.Errorf("failed to update netbird: %s", err)
		rsp.ErrRsp(c, -1, vpn.Message("update failed", err))
		return
	}
	rsp.OkRsp(c)
}
```

`server/router/extensions.go` (whole file):

```go
package router

import (
	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/extensions/netbird"
	"NanoKVM-Server/service/extensions/tailscale"

	"github.com/gin-gonic/gin"
)

func extensionsRouter(r *gin.Engine) {
	api := r.Group("/api/extensions").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)

	ts := tailscale.NewService()

	api.POST("/tailscale/install", ts.Install)     // install tailscale
	api.POST("/tailscale/uninstall", ts.Uninstall) // uninstall tailscale
	api.GET("/tailscale/status", ts.GetStatus)     // get tailscale status
	api.POST("/tailscale/up", ts.Up)               // run tailscale up
	api.POST("/tailscale/down", ts.Down)           // run tailscale down
	api.POST("/tailscale/login", ts.Login)         // tailscale login
	api.POST("/tailscale/logout", ts.Logout)       // tailscale logout
	api.POST("/tailscale/start", ts.Start)         // tailscale start
	api.POST("/tailscale/stop", ts.Stop)           // tailscale stop
	api.POST("/tailscale/restart", ts.Restart)     // tailscale restart
	api.POST("/tailscale/boot", ts.Boot)           // tailscale start at boot on or off
	api.GET("/tailscale/update", ts.GetUpdate)     // tailscale current and latest version
	api.POST("/tailscale/update", ts.Update)       // install the latest tailscale

	nb := netbird.NewService()

	api.POST("/netbird/install", nb.Install)     // install netbird from Alpine's package
	api.POST("/netbird/uninstall", nb.Uninstall) // uninstall netbird
	api.GET("/netbird/status", nb.GetStatus)     // get netbird status
	api.POST("/netbird/up", nb.Up)               // run netbird up
	api.POST("/netbird/down", nb.Down)           // run netbird down
	api.POST("/netbird/login", nb.Login)         // join with a setup key, or start an SSO login
	api.POST("/netbird/logout", nb.Logout)       // netbird deregister
	api.POST("/netbird/start", nb.Start)         // netbird start
	api.POST("/netbird/stop", nb.Stop)           // netbird stop
	api.POST("/netbird/restart", nb.Restart)     // netbird restart
	api.POST("/netbird/boot", nb.Boot)           // netbird start at boot on or off
	api.GET("/netbird/update", nb.GetUpdate)     // netbird current and latest version
	api.POST("/netbird/update", nb.Update)       // install the latest netbird
}
```

- [ ] **Step 4: Run all server tests touched so far**

Run: `gotest ./proto/ ./router/ ./service/extensions/... -v`
Expected: PASS.

Then check formatting:

```bash
cd /d/projects/NanoKVM/server && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM\\server:/src" -w /src golang:1.25 gofmt -l proto router service/extensions
```

Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add server/service/extensions/netbird/service.go server/service/extensions/netbird/service_test.go server/router/extensions.go
git commit -m "netbird: the API, beside Tailscale's"
```

---

### Task 13: S98netbird and the shell tests

**Files:**
- Create: `kvmapp/system/init.d/S98netbird` (mode 755)
- Create: `tools/service/test-netbird.sh` (mode 755)
- Modify: `kvmapp/system/init.d/S98tailscaled:129` and `:154` (function rename)
- Rename: `tools/service/test-tailscale-memlimit.sh` to `tools/service/test-addon-memlimit.sh` (and rewrite)
- Modify: `tools/service/test-cgroup-join.sh:5-9` and `:21`
- Modify: `kvmapp/system/init.d.package-only` (after the `S98tailscaled` line)

**Interfaces:**
- Consumes: nothing from Go. The server runs `sh /kvmapp/system/init.d/S98netbird {start|stop|restart}` (Task 11) and reads `/var/run/netbird.pid` (Task 5) and `/var/log/netbird.log` (Task 11).
- Produces: `S98netbird` honouring `NETBIRD`, `PIDFILE`, `SOCKET`, `LOGFILE`, `STATEDIR`, `KVMDIR`, `MOUNTS`, `DATA`, `IDENTITY_DIR`, `START_WAIT`, `STOP_WAIT`, `CGROUP_PROCS`; a `memlimit_mib` function inside `# --- memlimit ---` markers in both scripts.

- [ ] **Step 1: Write the failing shell test**

`tools/service/test-netbird.sh`:

```sh
#!/bin/sh
# Drive kvmapp/system/init.d/S98netbird against a scratch tree and a stub
# daemon.
#
#   test-netbird.sh [path-to-S98netbird]
#
# The script starts and stops the daemon by pid file, without
# start-stop-daemon, as S98tailscaled does. And the peer's identity lives on
# /data, so both slots are the same peer on the NetBird network.
S98=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S98netbird}
[ -f "$S98" ] || { echo "usage: test-netbird.sh <S98netbird>"; exit 1; }

for tool in pgrep mktemp; do
    command -v "$tool" >/dev/null 2>&1 || { echo "needs $tool"; exit 2; }
done
[ -d /proc/self ] || { echo "needs /proc"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
cleanup() {
    for p in $(pgrep -f "$work/bin/netbird"); do kill -9 "$p" 2>/dev/null; done
    rm -rf "$work"
}
trap cleanup EXIT

# restart runs "$0" itself, which needs an executable file. A checkout on a
# Windows host mounted into a container does not promise the bit, so the test
# runs a copy.
cp "$S98" "$work/S98netbird" && chmod 755 "$work/S98netbird"
S98=$work/S98netbird

mkdir -p "$work/bin"
# The stub records its environment and arguments and then stays up the way
# `netbird service run` does. Its command line names netbird, which is what the
# script checks a pid against. STUB_EXIT makes it die at once.
cat > "$work/bin/netbird" <<'STUB'
#!/bin/sh
echo "NB_STATE_DIR=$NB_STATE_DIR GOMEMLIMIT=$GOMEMLIMIT $*" >> "$STUB_LOG"
[ -n "$STUB_EXIT" ] && exit 1
trap 'exit 0' TERM
while :; do sleep 1; done
STUB
chmod 755 "$work/bin/netbird"

reset_tree() {
    rm -rf "$work/root" "$work/stub.log"
    mkdir -p "$work/root/data" "$work/root/etc/kvm"
    : > "$work/stub.log"
    if [ "$1" = mounted ]; then
        printf '/dev/mmcblk0p6 %s exfat rw 0 0\n' "$work/root/data" > "$work/root/mounts"
    else
        : > "$work/root/mounts"
    fi
}

run() {
    NETBIRD="$work/bin/netbird" \
    PIDFILE="$work/root/var/run/netbird.pid" \
    SOCKET="$work/root/var/run/netbird.sock" \
    LOGFILE="$work/root/var/log/netbird.log" \
    STATEDIR="$work/root/var/lib/netbird" \
    KVMDIR="$work/root/etc/kvm" MOUNTS="$work/root/mounts" DATA="$work/root/data" \
    CGROUP_PROCS="$work/root/no-cgroup/cgroup.procs" \
    START_WAIT=1 STOP_WAIT=5 STUB_LOG="$work/stub.log" \
        sh "$S98" "$@" > "$work/out" 2>&1
}

daemons() { pgrep -f "$work/bin/netbird service run" | wc -l | tr -d ' '; }
started_with() { grep -- "service run" "$work/stub.log" | tail -1; }

shared="$work/root/data/identity-system/netbird"
slot="$work/root/var/lib/netbird"

echo "===== start and stop ====="
reset_tree mounted
run start
grep -q '^Starting netbird: OK$' "$work/out" \
    && note "start reports OK" OK \
    || { note "start reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(daemons)" = 1 ] \
    && note "one daemon is running" OK \
    || note "one daemon is running (got $(daemons))" FAIL
pid=$(cat "$work/root/var/run/netbird.pid" 2>/dev/null)
[ -n "$pid" ] && grep -q netbird "/proc/$pid/cmdline" 2>/dev/null \
    && note "the pid file names the daemon" OK \
    || note "the pid file names the daemon" FAIL
case "$(started_with)" in
    *"--daemon-addr unix://$work/root/var/run/netbird.sock "*) note "it serves the CLI on the socket" OK ;;
    *) note "it serves the CLI on the socket" FAIL; echo "    $(started_with)" ;;
esac
case "$(started_with)" in
    *"--log-file $work/root/var/log/netbird.log"*) note "it logs to the log file" OK ;;
    *) note "it logs to the log file" FAIL ;;
esac
case "$(started_with)" in
    *"GOMEMLIMIT=512MiB "*) note "with no group and no setting the limit is 512MiB" OK ;;
    *) note "with no group and no setting the limit is 512MiB" FAIL; echo "    $(started_with)" ;;
esac

run start
[ "$(daemons)" = 1 ] \
    && note "a second start does not start a second daemon" OK \
    || note "a second start does not start a second daemon (got $(daemons))" FAIL

old=$(cat "$work/root/var/run/netbird.pid" 2>/dev/null)
run restart
new=$(cat "$work/root/var/run/netbird.pid" 2>/dev/null)
[ "$(daemons)" = 1 ] && [ -n "$new" ] && [ "$new" != "$old" ] \
    && note "restart leaves one new daemon" OK \
    || note "restart leaves one new daemon (got $(daemons), $old -> $new)" FAIL

run stop
grep -q '^Stopping netbird: OK$' "$work/out" \
    && note "stop reports OK" OK \
    || { note "stop reports OK" FAIL; sed 's/^/    /' "$work/out"; }
[ "$(daemons)" = 0 ] \
    && note "no daemon is left running" OK \
    || note "no daemon is left running (got $(daemons))" FAIL
[ ! -e "$work/root/var/run/netbird.pid" ] \
    && note "the pid file is removed" OK \
    || note "the pid file is removed" FAIL

echo
echo "===== the identity lives on /data ====="
reset_tree mounted
run start; run stop
case "$(started_with)" in
    *"NB_STATE_DIR=$shared "*) note "the profile directory is on /data" OK ;;
    *) note "the profile directory is on /data" FAIL; echo "    $(started_with)" ;;
esac
case "$(started_with)" in
    *"--config $shared/config.json "*) note "and so is the config" OK ;;
    *) note "and so is the config" FAIL ;;
esac

reset_tree unmounted
run start; run stop
case "$(started_with)" in
    *"NB_STATE_DIR=$slot "*"--config $slot/config.json "*) note "without /data it stays on the slot" OK ;;
    *) note "without /data it stays on the slot" FAIL; echo "    $(started_with)" ;;
esac
[ ! -d "$work/root/data/identity-system" ] \
    && note "and nothing is written under an unmounted /data" OK \
    || note "and nothing is written under an unmounted /data" FAIL

echo
echo "===== a pid file it cannot trust ====="
reset_tree mounted
sleep 60 &
bystander=$!
mkdir -p "$work/root/var/run"
echo "$bystander" > "$work/root/var/run/netbird.pid"
run stop
kill -0 "$bystander" 2>/dev/null \
    && note "stop does not signal a process that is not netbird" OK \
    || note "stop does not signal a process that is not netbird" FAIL
echo "$bystander" > "$work/root/var/run/netbird.pid"
run start
[ "$(daemons)" = 1 ] \
    && note "start is not fooled into thinking it is running" OK \
    || note "start is not fooled into thinking it is running" FAIL
run stop
kill "$bystander" 2>/dev/null

echo
echo "===== a daemon that dies at once ====="
reset_tree mounted
STUB_EXIT=1 run start
grep -q '^Starting netbird: FAIL$' "$work/out" \
    && note "start reports FAIL" OK \
    || { note "start reports FAIL" FAIL; sed 's/^/    /' "$work/out"; }
[ ! -e "$work/root/var/run/netbird.pid" ] \
    && note "and leaves no pid file behind" OK \
    || note "and leaves no pid file behind" FAIL

echo
echo "===== no binary ====="
reset_tree mounted
NETBIRD="$work/bin/absent" sh "$S98" start > "$work/out" 2>&1
st=$?
[ "$st" -ne 0 ] && grep -q "not found" "$work/out" \
    && note "a missing binary is reported and fails" OK \
    || note "a missing binary is reported and fails (status $st)" FAIL

echo
echo "===== the script still parses ====="
sh -n "$S98" 2>/dev/null \
    && note "sh -n accepts the script" OK \
    || note "sh -n rejects the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
```

`tools/service/test-addon-memlimit.sh` (replaces `test-tailscale-memlimit.sh`):

```bash
cd /d/projects/NanoKVM && git mv tools/service/test-tailscale-memlimit.sh tools/service/test-addon-memlimit.sh
```

Then its whole content:

```sh
#!/bin/sh
# The Go memory limit S98tailscaled and S98netbird give their daemons.
#
#   test-addon-memlimit.sh
#
# The limit is the owner's /etc/kvm/GOMEMLIMIT, or 512 MiB without it, capped
# at seven eighths of the addons group's memory.high. Only the block between
# "# --- memlimit ---" and "# --- end memlimit ---" runs here, never a script.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if [ "$2" = "$3" ]; then note "$1" OK; else note "$1" FAIL; echo "    got '$2' want '$3'"; fi; }

# limit <owner setting or -> <memory.high or ->
limit() {
    rm -rf "$WORK/kvm" "$WORK/cg"; mkdir -p "$WORK/kvm" "$WORK/cg"
    [ "$1" = - ] || printf '%s\n' "$1" > "$WORK/kvm/GOMEMLIMIT"
    [ "$2" = - ] || printf '%s\n' "$2" > "$WORK/cg/memory.high"
    KVMDIR=$WORK/kvm CGROUP_PROCS=$WORK/cg/cgroup.procs \
        sh -c '. "$1"; memlimit_mib' _ "$WORK/block.sh"
}

for s in S98tailscaled S98netbird; do
    S=$HERE/../../kvmapp/system/init.d/$s
    echo "$s"
    sed -n '/^# --- memlimit ---$/,/^# --- end memlimit ---$/p' "$S" > "$WORK/block.sh"
    check "the block can be extracted" "$([ -s "$WORK/block.sh" ] && echo yes)" "yes"
    check "start uses it" "$(grep -c 'GOMEMLIMIT="$(memlimit_mib)MiB"' "$S")" "1"

    check "no setting, no group: 512" "$(limit - -)" "512"
    check "an owner's 75, no group: 75" "$(limit 75 -)" "75"
    check "no setting, memory.high 64M: 56" "$(limit - 67108864)" "56"
    check "an owner's 75, memory.high 64M: 56" "$(limit 75 67108864)" "56"
    check "an owner's 40, memory.high 64M: 40" "$(limit 40 67108864)" "40"
    check "memory.high max (no limit): the setting stands" "$(limit 75 max)" "75"
    check "a setting that is not a number: 512" "$(limit 75MiB -)" "512"
    check "a setting of 0: 512" "$(limit 0 -)" "512"
    check "a memory.high under 2 MiB caps nothing" "$(limit 75 1048576)" "75"
done

[ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
[ "$fails" -eq 0 ]
```

In `tools/service/test-cgroup-join.sh`, change the header comment's second paragraph to:

```sh
# ironkvm-dist's S01cgroups makes two groups, kvm and addons. S95nanokvm joins
# kvm; S98tailscaled, S98netbird and S96picoclaw join addons. The join is the
# block between "# --- cgroup ---" and "# --- end cgroup ---" in each script.
# Only that block is run here, never the script, because running an init
# script starts or stops real services.
```

and the loop line to:

```sh
for pair in S95nanokvm:kvm S98tailscaled:addons S98netbird:addons S96picoclaw:addons; do
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /d/projects/NanoKVM
for s in test-netbird test-addon-memlimit test-cgroup-join; do
  MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM:/r" -w /r alpine:3.24 sh tools/service/$s.sh; echo "$s exit=$?"
done
```

Expected: `test-netbird` exits 1 with "usage" (no script yet); `test-addon-memlimit` FAILs "start uses it" for S98tailscaled and every S98netbird case; `test-cgroup-join` FAILs the S98netbird block.

- [ ] **Step 3: Write the script and rename the function**

In `kvmapp/system/init.d/S98tailscaled`, line 129 `tailscale_memlimit_mib() {` becomes `memlimit_mib() {`, and line 154 becomes:

```sh
                export GOMEMLIMIT="$(memlimit_mib)MiB"
```

`kvmapp/system/init.d/S98netbird`:

```sh
#!/bin/sh
#
# NetBird's daemon. The web UI starts, stops and restarts it through this
# script in /kvmapp/system/init.d. While start at boot is on, the script is
# also in /etc/init.d, where S04addons puts it back after a new image.
# Modeled on S98tailscaled.

DAEMON="netbird"

# Every path is a variable so tools/service/test-netbird.sh can drive this
# script against a scratch tree and a stub daemon. On the device nothing sets
# them.
NETBIRD=${NETBIRD:-/usr/bin/$DAEMON}
PIDFILE=${PIDFILE:-/var/run/$DAEMON.pid}
SOCKET=${SOCKET:-/var/run/$DAEMON.sock}
LOGFILE=${LOGFILE:-/var/log/$DAEMON.log}
STATEDIR=${STATEDIR:-/var/lib/netbird}
KVMDIR=${KVMDIR:-/etc/kvm}
MOUNTS=${MOUNTS:-/proc/mounts}
DATA=${DATA:-/data}
START_WAIT=${START_WAIT:-5}
STOP_WAIT=${STOP_WAIT:-15}

# The peer's identity is its WireGuard key and its login, in NetBird's profile
# files. It belongs to the board, not to a slot, as Tailscale's state does, so
# a board that boots its other slot is still the same peer.
#
# NetBird 0.78 keeps its profiles in one directory: default.json,
# active_profile.json and state.json. NB_STATE_DIR moves that directory, and
# --config names the default profile inside it. With /data mounted both point
# there; without it everything stays under /var/lib/netbird.
IDENTITY_DIR=${IDENTITY_DIR:-$DATA/identity-system/netbird}

if [ ! -x "$NETBIRD" ]; then
    echo "$NETBIRD not found, install NetBird from the web UI"
    exit 1
fi

state_dir() {
    if grep -q " $DATA " "$MOUNTS" 2>/dev/null && mkdir -p "$IDENTITY_DIR" 2>/dev/null; then
        echo "$IDENTITY_DIR"
        return
    fi
    mkdir -p "$STATEDIR"
    echo "$STATEDIR"
}

# Started and stopped by pid file, in the shell, as S98tailscaled explains:
# Alpine's busybox has no start-stop-daemon. A pid file is trusted only while
# its process is still netbird, so a stale file after a power cut never gets
# another process signalled.
running_pid() {
    [ -f "$PIDFILE" ] || return 1
    pid=$(cat "$PIDFILE" 2>/dev/null)
    case "$pid" in
        "" | *[!0-9]*) return 1 ;;
    esac
    kill -0 "$pid" 2>/dev/null || return 1
    grep -q "$DAEMON" "/proc/$pid/cmdline" 2>/dev/null || return 1
    echo "$pid"
}

# --- cgroup ---
# ironkvm-dist's S01cgroups makes an addons memory group with a soft and a hard
# limit. This script joins it before it starts the daemon, so the daemon
# inherits it however the script is run: at boot, or from the web UI through
# the server, whose own group it would otherwise inherit. Without the group
# nothing changes.
CGROUP_PROCS=${CGROUP_PROCS:-/sys/fs/cgroup/addons/cgroup.procs}
if [ -w "$CGROUP_PROCS" ]; then
    { echo $$ > "$CGROUP_PROCS"; } 2>/dev/null || true
fi
# --- end cgroup ---

# --- memlimit ---
# The Go memory limit, in MiB. /etc/kvm/GOMEMLIMIT is the owner's setting,
# and without it the limit is 512. A value that is not a whole number is
# ignored: Go refuses to start under a GOMEMLIMIT it cannot parse.
#
# In the addons group the kernel throttles the daemon at memory.high and kills
# it at memory.max, and Go frees memory only as its heap nears GOMEMLIMIT. So
# the limit is capped at seven eighths of memory.high, 56 for 64M, which leaves
# room for what Go does not count. An owner's setting below that stays.
memlimit_mib() {
    mib=512
    if [ -f "$KVMDIR/GOMEMLIMIT" ]; then
        v=$(cat "$KVMDIR/GOMEMLIMIT" 2>/dev/null)
        case "$v" in
            ''|*[!0-9]*|0) ;;
            *) mib=$v ;;
        esac
    fi
    high=$(cat "${CGROUP_PROCS%/*}/memory.high" 2>/dev/null)
    case "$high" in
        ''|*[!0-9]*) ;;
        *)
            cap=$((high / 1048576 * 7 / 8))
            if [ "$cap" -gt 0 ] && [ "$cap" -lt "$mib" ]; then
                mib=$cap
            fi
            ;;
    esac
    echo "$mib"
}
# --- end memlimit ---

case "$1" in
        start)
                export GOMEMLIMIT="$(memlimit_mib)MiB"
                echo "GOMEMLIMIT set to ${GOMEMLIMIT}"

                if pid=$(running_pid); then
                    echo "$DAEMON is already running as $pid"
                    exit 0
                fi

                mkdir -p "$(dirname "$PIDFILE")" "$(dirname "$SOCKET")" "$(dirname "$LOGFILE")"
                dir=$(state_dir)
                export NB_STATE_DIR="$dir"

                printf "Starting $DAEMON: "
                "$NETBIRD" service run \
                    --config "$dir/config.json" \
                    --daemon-addr "unix://$SOCKET" \
                    --log-file "$LOGFILE" </dev/null >/dev/null 2>&1 &
                echo $! > "$PIDFILE"
                sleep "$START_WAIT"
                if running_pid >/dev/null; then
                    echo "OK"
                else
                    rm -f "$PIDFILE"
                    echo "FAIL"
                fi
                ;;
        stop)
                printf "Stopping $DAEMON: "
                if pid=$(running_pid); then
                    kill "$pid" 2>/dev/null
                    waited=0
                    while kill -0 "$pid" 2>/dev/null && [ "$waited" -lt "$STOP_WAIT" ]; do
                        sleep 1
                        waited=$((waited + 1))
                    done
                    if kill -0 "$pid" 2>/dev/null; then
                        echo "FAIL"
                    else
                        rm -f "$PIDFILE"
                        echo "OK"
                    fi
                else
                    rm -f "$PIDFILE"
                    echo "FAIL"
                fi
                ;;
        restart|reload)
                "$0" stop
                "$0" start
                ;;
        *)
                echo "Usage: $0 {start|stop|restart}"
                exit 1
esac

exit 0
```

`restart` runs `"$0"` itself, so in the test harness the environment carries through to both halves.

In `kvmapp/system/init.d.package-only`, add after the `S98tailscaled` line:

```
S98netbird       not-installed-by-policy  started on demand, not at boot
```

Set the modes (core.filemode is false on this checkout):

```bash
cd /d/projects/NanoKVM
git add kvmapp/system/init.d/S98netbird tools/service/test-netbird.sh
git update-index --chmod=+x kvmapp/system/init.d/S98netbird tools/service/test-netbird.sh
```

- [ ] **Step 4: Run the shell tests**

```bash
cd /d/projects/NanoKVM
for s in service/test-netbird service/test-addon-memlimit service/test-cgroup-join service/test-tailscaled release/test-init-install-list; do
  MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM:/r" -w /r alpine:3.24 sh tools/$s.sh; echo "$s exit=$?"
done
```

Expected: every suite prints "all cases passed" (or only `ok` lines for the list test) and exits 0. If `test-netbird` exits 2, the container lacks `pgrep`: use `alpine:3.24` with `apk add procps` first, and never read an exit 2 as a pass.

- [ ] **Step 5: Commit**

```bash
git add kvmapp/system/init.d/S98netbird kvmapp/system/init.d/S98tailscaled kvmapp/system/init.d.package-only tools/service/test-netbird.sh tools/service/test-addon-memlimit.sh tools/service/test-cgroup-join.sh
git commit -m "S98netbird: NetBird's daemon in the addons group, its identity on /data

GOMEMLIMIT is derived as in S98tailscaled, whose memlimit function is now
memlimit_mib so one test covers both scripts."
```

---

### Task 14: The shared VPN page parts, and the Tailscale API additions

The web has no unit test runner. The test cycle for Tasks 14 to 16 is the type check and build (`npm run build` runs `tsc`) and the linter; the behaviour is checked on the board in Task 18.

**Files:**
- Modify: `web/src/api/extensions/tailscale.ts` (whole file)
- Create: `web/src/pages/desktop/menu/settings/vpn/types.ts`, `format.ts`, `error-detail.tsx`, `notice.tsx`, `boot.tsx`, `details.tsx`, `peers.tsx`, `memory.tsx`, `install.tsx`, `run.tsx`, `device.tsx`, `uninstall.tsx`, `swap.tsx`, `header.tsx`, `page.tsx`
- Modify: `web/src/i18n/locales/en.ts` (a `vpn` block before `tailscale: {`, around line 634)

**Interfaces:**
- Consumes: the routes of Tasks 8, 9 and 12; the `VpnStatus` JSON names of Task 2 (`state`, `version`, `ip`, `name`, `account`, `control`, `peers[{name, ip, online}]`, `uptimeSec`, `bootEnabled`, `memory{daemonRss, groupCurrent, groupHigh, groupMax}`, `blockedBy`); `{current, latest}` from `GET update`.
- Produces: `VpnPage({ vpn: VpnInfo, setIsLocked })` in `vpn/page.tsx`; `ErrorDetail({ message })` in `vpn/error-detail.tsx`; types `State`, `Peer`, `Memory`, `Status`, `UpdateInfo`, `Rsp`, `VpnApi`, `VpnInfo` in `vpn/types.ts`; `formatBytes`, `formatUptime`, `vpnTitles` in `vpn/format.ts`; i18n keys `settings.vpn.*`; Tailscale API functions `setBoot(enabled)`, `getUpdate()`, `update()`.

- [ ] **Step 1: Record the baseline**

```bash
cd /d/projects/NanoKVM/web && npm run build && npm run lint > /tmp/lint-before.txt 2>&1; echo "lint exit=$?"; tail -3 /tmp/lint-before.txt
```

Expected: the build passes. Note the lint error and warning counts; the tasks below must not add errors.

- [ ] **Step 2: Write the API additions**

`web/src/api/extensions/tailscale.ts`:

```ts
import { http } from '@/lib/http.ts';

// Install and update download a release, which can take minutes on the
// board's link. The default request timeout is one minute.
const LONG = { timeout: 10 * 60 * 1000 };

// install tailscale
export function install() {
  return http.post('/api/extensions/tailscale/install', undefined, LONG);
}

// uninstall tailscale
export function uninstall() {
  return http.post('/api/extensions/tailscale/uninstall');
}

// get tailscale status
export function getStatus() {
  return http.get('/api/extensions/tailscale/status');
}

// start tailscale
export function start() {
  return http.post('/api/extensions/tailscale/start');
}

// restart tailscale
export function restart() {
  return http.post('/api/extensions/tailscale/restart');
}

// stop tailscale
export function stop() {
  return http.post('/api/extensions/tailscale/stop');
}

// run tailscale up
export function up() {
  return http.post('/api/extensions/tailscale/up');
}

// run tailscale down
export function down() {
  return http.post('/api/extensions/tailscale/down');
}

// login tailscale
export function login() {
  return http.post('/api/extensions/tailscale/login');
}

// logout tailscale
export function logout() {
  return http.post('/api/extensions/tailscale/logout');
}

// turn start at boot on or off
export function setBoot(enabled: boolean) {
  return http.post('/api/extensions/tailscale/boot', { enabled });
}

// get the installed and the latest version
export function getUpdate() {
  return http.get('/api/extensions/tailscale/update');
}

// install the latest version
export function update() {
  return http.post('/api/extensions/tailscale/update', undefined, LONG);
}
```

- [ ] **Step 3: Write the shared types and helpers**

`web/src/pages/desktop/menu/settings/vpn/types.ts`:

```ts
import type { ReactNode } from 'react';

export type State = 'notInstall' | 'notRunning' | 'notLogin' | 'stopped' | 'running';

export type Peer = { name: string; ip: string; online: boolean };

// Bytes. groupHigh and groupMax are 0 when the group sets no limit.
export type Memory = {
  daemonRss: number;
  groupCurrent: number;
  groupHigh: number;
  groupMax: number;
};

export type Status = {
  state: State;
  version: string;
  ip: string;
  name: string;
  account: string;
  control: boolean;
  peers: Peer[] | null;
  uptimeSec: number;
  bootEnabled: boolean;
  memory: Memory;
  blockedBy: string;
};

export type UpdateInfo = { current: string; latest: string };

// The server's reply, as lib/http.ts returns it.
export type Rsp = { code: number; msg: string; data: any };

// What the shared page needs from an add-on's API module. Login differs
// between the two and is each add-on's own form.
export type VpnApi = {
  install: () => Promise<Rsp>;
  uninstall: () => Promise<Rsp>;
  getStatus: () => Promise<Rsp>;
  start: () => Promise<Rsp>;
  stop: () => Promise<Rsp>;
  restart: () => Promise<Rsp>;
  up: () => Promise<Rsp>;
  down: () => Promise<Rsp>;
  logout: () => Promise<Rsp>;
  setBoot: (enabled: boolean) => Promise<Rsp>;
  getUpdate: () => Promise<Rsp>;
  update: () => Promise<Rsp>;
};

export type VpnInfo = {
  id: 'tailscale' | 'netbird';
  title: string;
  api: VpnApi;
  logoutLabel: string;
  logoutWarning: string;
  renderLogin: (onSuccess: () => void) => ReactNode;
  installHelp?: ReactNode;
};
```

`web/src/pages/desktop/menu/settings/vpn/format.ts`:

```ts
export const vpnTitles: Record<string, string> = {
  tailscale: 'Tailscale',
  netbird: 'NetBird'
};

export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '-';
  const mib = bytes / 1024 / 1024;
  return mib >= 1024 ? `${(mib / 1024).toFixed(1)} GiB` : `${mib.toFixed(1)} MiB`;
}

export function formatUptime(sec: number): string {
  if (!sec || sec <= 0) return '-';
  const days = Math.floor(sec / 86400);
  const hours = Math.floor((sec % 86400) / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}
```

- [ ] **Step 4: Write the small parts**

`web/src/pages/desktop/menu/settings/vpn/error-detail.tsx`:

```tsx
type ErrorDetailProps = {
  message: string;
};

// ErrorDetail shows why an action failed. The server sends the last lines of
// the command's output, so the line breaks are kept.
export const ErrorDetail = ({ message }: ErrorDetailProps) => {
  if (!message) return null;

  return (
    <pre className="mt-5 max-h-[200px] w-full overflow-auto whitespace-pre-wrap break-words rounded bg-neutral-800/60 p-3 font-mono text-xs text-red-400">
      {message}
    </pre>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/notice.tsx`:

```tsx
import { Alert } from 'antd';
import { useTranslation } from 'react-i18next';

import { vpnTitles } from './format.ts';

type NoticeProps = {
  blockedBy: string;
};

// Notice says, before any button is pressed, that the other VPN runs or
// starts at boot, and that only one runs at a time.
export const Notice = ({ blockedBy }: NoticeProps) => {
  const { t } = useTranslation();
  if (!blockedBy) return null;

  return (
    <Alert
      className="mt-5"
      type="warning"
      showIcon
      message={t('settings.vpn.blocked', { other: vpnTitles[blockedBy] ?? blockedBy })}
    />
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/boot.tsx`:

```tsx
import { useState } from 'react';
import { Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import type { VpnInfo } from './types.ts';

type BootProps = {
  vpn: VpnInfo;
  enabled: boolean;
  blocked: boolean;
  onChange: () => void;
  onError: (msg: string) => void;
};

// Boot is start at boot. It is its own switch: starting and stopping the
// daemon no longer change it.
export const Boot = ({ vpn, enabled, blocked, onChange, onError }: BootProps) => {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  function toggle(next: boolean) {
    if (isLoading) return;
    setIsLoading(true);
    onError('');

    vpn.api
      .setBoot(next)
      .then((rsp) => {
        if (rsp.code !== 0) {
          onError(rsp.msg);
          return;
        }
        onChange();
      })
      .catch((err) => onError(err?.message || 'Failed to set start at boot'))
      .finally(() => setIsLoading(false));
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col">
        <span>{t('settings.vpn.boot')}</span>
        <span className="text-xs text-neutral-500">
          {t('settings.vpn.bootDesc', { name: vpn.title })}
        </span>
      </div>
      <Switch
        checked={enabled}
        loading={isLoading}
        disabled={blocked && !enabled}
        onChange={toggle}
      />
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/details.tsx`:

```tsx
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

import { formatUptime } from './format.ts';
import type { Status } from './types.ts';

type DetailsProps = {
  status: Status;
};

export const Details = ({ status }: DetailsProps) => {
  const { t } = useTranslation();

  const rows: [string, ReactNode][] = [
    [
      t('settings.vpn.control'),
      status.control ? (
        <span className="text-green-500">{t('settings.vpn.connected')}</span>
      ) : (
        <span className="text-red-400">{t('settings.vpn.disconnected')}</span>
      )
    ],
    [t('settings.vpn.deviceName'), status.name || '-'],
    [t('settings.vpn.deviceIP'), status.ip || '-'],
    [t('settings.vpn.account'), status.account || '-'],
    [t('settings.vpn.version'), status.version || '-'],
    [t('settings.vpn.uptime'), formatUptime(status.uptimeSec)]
  ];

  return (
    <div className="flex flex-col space-y-3">
      {rows.map(([label, value]) => (
        <div key={label} className="flex justify-between">
          <span>{label}</span>
          <span className="text-neutral-300">{value}</span>
        </div>
      ))}
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/peers.tsx`:

```tsx
import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

import type { Peer } from './types.ts';

type PeersProps = {
  peers: Peer[];
};

export const Peers = ({ peers }: PeersProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col space-y-2">
      <span>{t('settings.vpn.peers')}</span>

      {peers.length === 0 ? (
        <span className="text-sm text-neutral-500">{t('settings.vpn.noPeers')}</span>
      ) : (
        <ul className="flex flex-col space-y-1">
          {peers.map((peer) => (
            <li key={`${peer.name}-${peer.ip}`} className="flex items-center justify-between text-sm">
              <span className="flex items-center space-x-2">
                <span
                  role="img"
                  aria-label={peer.online ? t('settings.vpn.online') : t('settings.vpn.offline')}
                  title={peer.online ? t('settings.vpn.online') : t('settings.vpn.offline')}
                  className={clsx(
                    'inline-block h-2 w-2 rounded-full',
                    peer.online ? 'bg-green-500' : 'bg-neutral-600'
                  )}
                />
                <span>{peer.name}</span>
              </span>
              <span className="font-mono text-neutral-400">{peer.ip}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/memory.tsx`:

```tsx
import { Progress } from 'antd';
import { useTranslation } from 'react-i18next';

import { formatBytes } from './format.ts';
import type { Memory } from './types.ts';

type MemoryBarsProps = {
  memory: Memory;
};

// MemoryBars puts the daemon and the whole addons group against the group's
// memory.high, where the kernel starts to throttle, and names memory.max,
// where it kills.
export const MemoryBars = ({ memory }: MemoryBarsProps) => {
  const { t } = useTranslation();

  const limit = memory.groupHigh || memory.groupMax;
  const percent = (value: number) => (limit > 0 ? Math.min(100, (value / limit) * 100) : 0);
  const pressed = memory.groupHigh > 0 && memory.groupCurrent >= memory.groupHigh * 0.9;
  const limits = [
    memory.groupHigh > 0 ? t('settings.vpn.high', { size: formatBytes(memory.groupHigh) }) : '',
    memory.groupMax > 0 ? t('settings.vpn.max', { size: formatBytes(memory.groupMax) }) : ''
  ].filter(Boolean);

  return (
    <div className="flex flex-col space-y-2">
      <span>{t('settings.vpn.memory')}</span>

      <Bar
        label={t('settings.vpn.daemonRss')}
        value={memory.daemonRss}
        percent={percent(memory.daemonRss)}
        hasLimit={limit > 0}
      />

      {limit > 0 ? (
        <>
          <Bar
            label={t('settings.vpn.group')}
            value={memory.groupCurrent}
            percent={percent(memory.groupCurrent)}
            warn={pressed}
            hasLimit
          />
          <span className="text-xs text-neutral-500">{limits.join(', ')}</span>
        </>
      ) : (
        <span className="text-xs text-neutral-500">{t('settings.vpn.noGroup')}</span>
      )}
    </div>
  );
};

type BarProps = {
  label: string;
  value: number;
  percent: number;
  warn?: boolean;
  hasLimit: boolean;
};

const Bar = ({ label, value, percent, warn, hasLimit }: BarProps) => (
  <div className="flex flex-col">
    <div className="flex justify-between text-sm">
      <span className="text-neutral-400">{label}</span>
      <span className="text-neutral-300">{formatBytes(value)}</span>
    </div>
    {hasLimit && (
      <Progress
        percent={percent}
        showInfo={false}
        size="small"
        strokeColor={warn ? '#f59e0b' : undefined}
      />
    )}
  </div>
);
```

`web/src/pages/desktop/menu/settings/vpn/swap.tsx` (the Tailscale page's swap control, moved, with the shared strings):

```tsx
import { useEffect, useState } from 'react';
import { Switch, Tooltip } from 'antd';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';

export const Swap = () => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(true);
  const [isEnabled, setIsEnabled] = useState(false);

  useEffect(() => {
    api
      .getSwap()
      .then((rsp) => {
        if (rsp.data?.size > 0) {
          setIsEnabled(true);
        }
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  function update(enable: boolean) {
    if (isLoading) return;
    setIsLoading(true);

    const size = enable ? 256 : 0;

    api
      .setSwap(size)
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }

        setIsEnabled(enable);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="flex h-[40px] cursor-pointer items-center justify-between space-x-6 rounded px-2 text-neutral-300 hover:bg-neutral-700/70">
      <div className="flex items-center space-x-1">
        <span>{t('settings.vpn.swap.title')}</span>
        <Tooltip
          title={t('settings.vpn.swap.tip')}
          className="cursor-pointer text-neutral-500"
          placement="top"
          styles={{ root: { maxWidth: '400px' } }}
        >
          <CircleHelpIcon size={15} />
        </Tooltip>
      </div>

      <Switch value={isEnabled} loading={isLoading} size="small" onChange={update} />
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/uninstall.tsx`:

```tsx
import { useState } from 'react';
import { Modal } from 'antd';
import { Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { VpnInfo } from './types.ts';

type UninstallProps = {
  vpn: VpnInfo;
  onSuccess: () => void;
};

export const Uninstall = ({ vpn, onSuccess }: UninstallProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);

  function uninstall() {
    if (isLoading) return;
    setIsLoading(true);

    vpn.api.uninstall().finally(() => {
      setIsModalOpen(false);
      setIsLoading(false);
      onSuccess();
    });
  }

  const title = (
    <div className="flex items-center space-x-1 text-red-500">
      <Trash2Icon size={18} />
      <span>{t('settings.vpn.uninstall', { name: vpn.title })}</span>
    </div>
  );

  return (
    <>
      <div
        className="flex h-[30px] cursor-pointer items-center space-x-1 rounded px-2 py-1 text-neutral-300 hover:bg-neutral-700/70"
        onClick={() => setIsModalOpen(true)}
      >
        <span>{t('settings.vpn.uninstall', { name: vpn.title })}</span>
      </div>

      <Modal
        title={title}
        open={isModalOpen}
        centered={true}
        okType="danger"
        okText={t('settings.vpn.okBtn')}
        cancelText={t('settings.vpn.cancelBtn')}
        onOk={uninstall}
        onCancel={() => setIsModalOpen(false)}
        confirmLoading={isLoading}
      >
        <div className="py-5">
          <p className="text-base">{t('settings.vpn.uninstallDesc', { name: vpn.title })}</p>
        </div>
      </Modal>
    </>
  );
};
```

- [ ] **Step 5: Write the state views**

`web/src/pages/desktop/menu/settings/vpn/install.tsx`:

```tsx
import { useState } from 'react';
import { DownloadOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { Button, Card, Result } from 'antd';
import { useTranslation } from 'react-i18next';

import type { VpnInfo } from './types.ts';

type InstallProps = {
  vpn: VpnInfo;
  blocked: boolean;
  setIsLocked: (isLocked: boolean) => void;
  onSuccess: () => void;
  onError: (msg: string) => void;
};

export const Install = ({ vpn, blocked, setIsLocked, onSuccess, onError }: InstallProps) => {
  const { t } = useTranslation();

  const [state, setState] = useState<'' | 'installing' | 'failed'>('');

  function install() {
    if (state === 'installing' || blocked) return;
    setState('installing');
    setIsLocked(true);
    onError('');

    vpn.api
      .install()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setState('failed');
          onError(rsp.msg);
          return;
        }

        setState('');
        onSuccess();
      })
      .catch((err) => {
        setState('failed');
        onError(err?.message || 'Install failed');
      })
      .finally(() => {
        setIsLocked(false);
      });
  }

  if (state === 'failed') {
    return (
      <Result
        status="warning"
        title={t('settings.vpn.installFailed')}
        icon={<InfoCircleOutlined />}
        extra={
          <div className="flex flex-col items-center space-y-4">
            <Button onClick={() => setState('')}>{t('settings.vpn.retry')}</Button>
            {vpn.installHelp}
          </div>
        }
      />
    );
  }

  return (
    <Card>
      <Result
        icon={<DownloadOutlined />}
        subTitle={t('settings.vpn.notInstall', { name: vpn.title })}
        extra={
          <Button
            key="install"
            type="primary"
            loading={state === 'installing'}
            disabled={blocked}
            onClick={install}
          >
            {state === 'installing' ? t('settings.vpn.installing') : t('settings.vpn.install')}
          </Button>
        }
      />
    </Card>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/run.tsx`:

```tsx
import { useState } from 'react';
import { PauseCircleOutlined } from '@ant-design/icons';
import { Button, Card, Result } from 'antd';
import { useTranslation } from 'react-i18next';

import type { VpnInfo } from './types.ts';

type RunProps = {
  vpn: VpnInfo;
  blocked: boolean;
  onSuccess: () => void;
  onError: (msg: string) => void;
};

export const Run = ({ vpn, blocked, onSuccess, onError }: RunProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);

  function run() {
    if (isLoading || blocked) return;
    setIsLoading(true);
    onError('');

    vpn.api
      .start()
      .then((rsp) => {
        if (rsp.code !== 0) {
          onError(rsp.msg);
          return;
        }
        onSuccess();
      })
      .catch((err) => onError(err?.message || 'Failed to start'))
      .finally(() => setIsLoading(false));
  }

  return (
    <Card>
      <Result
        icon={<PauseCircleOutlined />}
        subTitle={t('settings.vpn.notRunning', { name: vpn.title })}
        extra={
          <Button key="run" type="primary" loading={isLoading} disabled={blocked} onClick={run}>
            {t('settings.vpn.run')}
          </Button>
        }
      />
    </Card>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/device.tsx`:

```tsx
import { useState } from 'react';
import { LogoutOutlined } from '@ant-design/icons';
import { Button, Divider, Popconfirm, Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import { Details } from './details.tsx';
import { MemoryBars } from './memory.tsx';
import { Peers } from './peers.tsx';
import type { Status, VpnInfo } from './types.ts';

type DeviceProps = {
  vpn: VpnInfo;
  status: Status;
  onChange: () => void;
  onError: (msg: string) => void;
};

export const Device = ({ vpn, status, onChange, onError }: DeviceProps) => {
  const { t } = useTranslation();

  const [isRunning, setIsRunning] = useState(status.state === 'running');
  const [isUpdating, setIsUpdating] = useState(false);
  const [isLogging, setIsLogging] = useState(false);

  // A new status from the parent replaces whatever the switch last set. This
  // is done during render rather than in an effect, so the switch never paints
  // the old value against the new status.
  const [prevStatus, setPrevStatus] = useState(status);
  if (status !== prevStatus) {
    setPrevStatus(status);
    setIsRunning(status.state === 'running');
  }

  async function toggle() {
    if (isUpdating) return;
    setIsUpdating(true);
    onError('');

    try {
      const rsp = isRunning ? await vpn.api.down() : await vpn.api.up();
      if (rsp.code !== 0) {
        onError(rsp.msg);
        return;
      }
      setIsRunning(!isRunning);
    } catch (err: any) {
      onError(err?.message || 'Request failed');
    } finally {
      setIsUpdating(false);
    }
  }

  function logout() {
    if (isLogging) return;
    setIsLogging(true);
    onError('');

    vpn.api
      .logout()
      .then((rsp) => {
        if (rsp.code !== 0) {
          onError(rsp.msg);
          return;
        }
        onChange();
      })
      .catch((err) => onError(err?.message || 'Failed to logout'))
      .finally(() => setIsLogging(false));
  }

  return (
    <div className="flex flex-col space-y-6 pt-5">
      <div className="flex justify-between">
        <span>{t('settings.vpn.enable', { name: vpn.title })}</span>
        <Switch checked={isRunning} loading={isUpdating} onClick={toggle} />
      </div>

      <Details status={status} />
      <Divider className="my-0" />
      <Peers peers={status.peers ?? []} />
      <Divider className="my-0" />
      <MemoryBars memory={status.memory} />
      <Divider className="my-0" />

      <div className="flex justify-center pt-3">
        <Popconfirm
          placement="bottom"
          title={<div className="max-w-[320px]">{vpn.logoutWarning}</div>}
          okText={t('settings.vpn.okBtn')}
          cancelText={t('settings.vpn.cancelBtn')}
          onConfirm={logout}
        >
          <Button
            danger
            type="primary"
            size="large"
            shape="round"
            icon={<LogoutOutlined />}
            loading={isLogging}
          >
            {vpn.logoutLabel}
          </Button>
        </Popconfirm>
      </div>
    </div>
  );
};
```

- [ ] **Step 6: Write the header and the page frame**

`web/src/pages/desktop/menu/settings/vpn/header.tsx`:

```tsx
import { useEffect, useState } from 'react';
import { Popconfirm, Popover } from 'antd';
import {
  CircleArrowUpIcon,
  CircleStopIcon,
  EllipsisIcon,
  LoaderIcon,
  RotateCwIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import semver from 'semver';

import { Swap } from './swap.tsx';
import type { Rsp, State, UpdateInfo, VpnInfo } from './types.ts';
import { Uninstall } from './uninstall.tsx';

type HeaderProps = {
  vpn: VpnInfo;
  state: State | undefined;
  setIsLocked: (isLocked: boolean) => void;
  onChange: () => void;
  onError: (msg: string) => void;
};

type Loading = '' | 'restarting' | 'stopping' | 'updating';

function isNewer(latest: string, current: string) {
  if (!latest || !current) return false;
  if (semver.valid(latest) && semver.valid(current)) return semver.gt(latest, current);
  return latest !== current;
}

export const Header = ({ vpn, state, setIsLocked, onChange, onError }: HeaderProps) => {
  const { t } = useTranslation();

  const [loading, setLoading] = useState<Loading>('');
  const [update, setUpdate] = useState<UpdateInfo>();
  const installed = !!state && state !== 'notInstall';

  // The server caches the answer for an hour, so asking on every visit is cheap.
  useEffect(() => {
    if (!installed) return;
    vpn.api
      .getUpdate()
      .then((rsp) => {
        if (rsp.code === 0) setUpdate(rsp.data);
      })
      .catch(() => {});
  }, [installed, vpn.api]);

  const hasUpdate = installed && !!update && isNewer(update.latest, update.current);

  function act(kind: Loading, request: () => Promise<Rsp>, lock = false) {
    if (loading !== '') return;
    setLoading(kind);
    if (lock) setIsLocked(true);
    onError('');

    request()
      .then((rsp) => {
        if (rsp.code !== 0) onError(rsp.msg);
      })
      .catch((err) => onError(err?.message || 'Request failed'))
      .finally(() => {
        setLoading('');
        if (lock) setIsLocked(false);
        onChange();
      });
  }

  function runUpdate() {
    act(
      'updating',
      () =>
        vpn.api.update().then((rsp) => {
          if (rsp.code === 0) setUpdate(undefined);
          return rsp;
        }),
      true
    );
  }

  return (
    <div className="flex items-center justify-between">
      <span className="text-base">{vpn.title}</span>

      <div className="flex items-center space-x-2">
        {hasUpdate && update && (
          <Popconfirm
            title={t('settings.vpn.update', { name: vpn.title, version: update.latest })}
            description={t('settings.vpn.updateDesc')}
            onConfirm={runUpdate}
            okText={t('settings.vpn.okBtn')}
            cancelText={t('settings.vpn.cancelBtn')}
            placement="bottom"
            disabled={loading !== ''}
          >
            <div className="flex cursor-pointer rounded p-1 text-blue-500 hover:bg-neutral-600 hover:text-blue-500/80">
              {loading === 'updating' ? (
                <LoaderIcon className="animate-spin" size={18} />
              ) : (
                <CircleArrowUpIcon size={18} />
              )}
            </div>
          </Popconfirm>
        )}

        {state && ['notLogin', 'stopped', 'running'].includes(state) && (
          <>
            <Popconfirm
              title={t('settings.vpn.restart', { name: vpn.title })}
              onConfirm={() => act('restarting', vpn.api.restart)}
              okText={t('settings.vpn.okBtn')}
              cancelText={t('settings.vpn.cancelBtn')}
              placement="bottom"
              disabled={loading !== ''}
            >
              <div className="flex cursor-pointer rounded p-1 text-green-500 hover:bg-neutral-600 hover:text-green-500/80">
                {loading === 'restarting' ? (
                  <LoaderIcon className="animate-spin" size={18} />
                ) : (
                  <RotateCwIcon size={18} />
                )}
              </div>
            </Popconfirm>

            <Popconfirm
              title={t('settings.vpn.stop', { name: vpn.title })}
              description={t('settings.vpn.stopDesc')}
              onConfirm={() => act('stopping', vpn.api.stop)}
              okText={t('settings.vpn.okBtn')}
              cancelText={t('settings.vpn.cancelBtn')}
              placement="bottom"
              disabled={loading !== ''}
            >
              <div className="flex cursor-pointer rounded p-1 text-red-500 hover:bg-neutral-600 hover:text-red-500/80">
                {loading === 'stopping' ? (
                  <LoaderIcon className="animate-spin" size={18} />
                ) : (
                  <CircleStopIcon size={18} />
                )}
              </div>
            </Popconfirm>
          </>
        )}

        {installed && (
          <Popover
            content={
              <div className="flex min-w-[250px] flex-col">
                <Swap />
                <Uninstall vpn={vpn} onSuccess={onChange} />
              </div>
            }
            placement="bottom"
            trigger="click"
          >
            <div className="flex cursor-pointer rounded p-1 text-white hover:bg-neutral-700/50">
              <EllipsisIcon size={18} />
            </div>
          </Popover>
        )}
      </div>
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/vpn/page.tsx`:

```tsx
import { useEffect, useState } from 'react';
import { Divider } from 'antd';
import { LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { Boot } from './boot.tsx';
import { Device } from './device.tsx';
import { ErrorDetail } from './error-detail.tsx';
import { Header } from './header.tsx';
import { Install } from './install.tsx';
import { Notice } from './notice.tsx';
import { Run } from './run.tsx';
import type { Status, VpnInfo } from './types.ts';

type VpnPageProps = {
  vpn: VpnInfo;
  setIsLocked: (isLocked: boolean) => void;
};

// VpnPage is the settings page of Tailscale and of NetBird. Each gives it its
// API module and its login form; the rest is the same for both.
export const VpnPage = ({ vpn, setIsLocked }: VpnPageProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<Status>();
  const [errMsg, setErrMsg] = useState('');

  const getStatus = useStableCallback(() => {
    if (isLoading) return;
    setIsLoading(true);

    vpn.api
      .getStatus()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }
        setStatus(rsp.data);
      })
      .catch((err) => {
        setErrMsg(err?.message || 'Failed to get status');
      })
      .finally(() => {
        setIsLoading(false);
      });
  });

  useEffect(() => {
    getStatus();
  }, [getStatus]);

  const blocked = !!status?.blockedBy;
  const installed = !!status && status.state !== 'notInstall';

  return (
    <>
      <Header
        vpn={vpn}
        state={status?.state}
        setIsLocked={setIsLocked}
        onChange={getStatus}
        onError={setErrMsg}
      />
      <Divider className="opacity-50" />

      {isLoading ? (
        <div className="flex w-full items-center justify-center space-x-2 pt-5 text-neutral-500">
          <LoaderCircleIcon className="animate-spin" size={18} />
          <span>{t('settings.vpn.loading')}</span>
        </div>
      ) : (
        <>
          <Notice blockedBy={status?.blockedBy ?? ''} />

          {status?.state === 'notInstall' && (
            <Install
              vpn={vpn}
              blocked={blocked}
              setIsLocked={setIsLocked}
              onSuccess={getStatus}
              onError={setErrMsg}
            />
          )}

          {status?.state === 'notRunning' && (
            <Run vpn={vpn} blocked={blocked} onSuccess={getStatus} onError={setErrMsg} />
          )}

          {status?.state === 'notLogin' && vpn.renderLogin(getStatus)}

          {(status?.state === 'stopped' || status?.state === 'running') && (
            <Device vpn={vpn} status={status} onChange={getStatus} onError={setErrMsg} />
          )}

          {installed && status && (
            <div className="pt-6">
              <Boot
                vpn={vpn}
                enabled={status.bootEnabled}
                blocked={blocked}
                onChange={getStatus}
                onError={setErrMsg}
              />
            </div>
          )}

          <ErrorDetail message={errMsg} />
        </>
      )}
    </>
  );
};
```

- [ ] **Step 7: Add the shared English strings**

In `web/src/i18n/locales/en.ts`, insert this block directly before the line `      tailscale: {` (around line 634):

```ts
      vpn: {
        loading: 'Loading...',
        okBtn: 'Yes',
        cancelBtn: 'No',
        restart: 'Restart {{name}}?',
        stop: 'Stop {{name}}?',
        stopDesc: 'The daemon stops now. Start at boot is a separate switch and stays as it is.',
        update: 'Update {{name}} to {{version}}?',
        updateDesc: 'The daemon restarts if it is running. The login is kept.',
        notInstall: '{{name}} is not installed.',
        install: 'Install',
        installing: 'Installing',
        installFailed: 'Install failed',
        retry: 'Try again',
        notRunning: '{{name}} is not running. Start it to continue.',
        run: 'Start',
        boot: 'Start at boot',
        bootDesc: 'Start {{name}} when the KVM boots.',
        enable: 'Enable {{name}}',
        control: 'Control server',
        connected: 'Connected',
        disconnected: 'Not connected',
        deviceName: 'Device name',
        deviceIP: 'Device IP',
        account: 'Account',
        version: 'Version',
        uptime: 'Uptime',
        peers: 'Peers',
        noPeers: 'No peers yet.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Memory',
        daemonRss: 'Daemon',
        group: 'Add-ons group',
        high: 'throttled above {{size}}',
        max: 'stopped by the kernel above {{size}}',
        noGroup: 'No add-ons memory group on this board.',
        uninstall: 'Uninstall {{name}}',
        uninstallDesc: 'Are you sure you want to uninstall {{name}}? The login stays on the board.',
        blocked:
          '{{other}} is running or starts at boot. Only one VPN runs at a time: stop {{other}} and turn off its start at boot first.',
        swap: {
          title: 'Swap memory',
          tip: 'If the daemon runs short of memory, try enabling swap memory. This sets the swap file size to 256MB by default, which can be adjusted in "Settings > Device".'
        }
      },
```

- [ ] **Step 8: Format, build and lint**

```bash
cd /d/projects/NanoKVM/web
npx prettier --write src/api/extensions/tailscale.ts src/pages/desktop/menu/settings/vpn/*.ts src/pages/desktop/menu/settings/vpn/*.tsx src/i18n/locales/en.ts
npm run build
npm run lint > /tmp/lint-after.txt 2>&1; echo "lint exit=$?"; grep -n "settings/vpn\|extensions/tailscale" /tmp/lint-after.txt
```

Expected: the build passes; no lint error in the new files (a `react-hooks` warning is allowed only if the same pattern already warns in the old Tailscale files); `git status` lists only the files of this task.

- [ ] **Step 9: Commit**

```bash
cd /d/projects/NanoKVM
git add web/src/api/extensions/tailscale.ts web/src/pages/desktop/menu/settings/vpn web/src/i18n/locales/en.ts
git commit -m "web: the shared VPN page, with start at boot, peers, memory, updates and the reason for a failure"
```

---

### Task 15: The Tailscale page on the shared frame

**Files:**
- Modify: `web/src/pages/desktop/menu/settings/tailscale/index.tsx` (whole file)
- Modify: `web/src/pages/desktop/menu/settings/tailscale/login.tsx` (whole file)
- Create: `web/src/pages/desktop/menu/settings/tailscale/install-help.tsx`
- Delete: `web/src/pages/desktop/menu/settings/tailscale/{header,install,run,device,uninstall,memory,swap}.tsx`, `types.ts`
- Modify: `web/src/i18n/locales/en.ts` (the `tailscale:` block)

**Interfaces:**
- Consumes: `VpnPage`, `VpnInfo`, `ErrorDetail` (Task 14); `@/api/extensions/tailscale.ts` (Task 14).
- Produces: `Tailscale({ setIsLocked })`, unchanged for `settings/index.tsx`.

- [ ] **Step 1: Remove the old parts**

```bash
cd /d/projects/NanoKVM/web/src/pages/desktop/menu/settings/tailscale
git rm header.tsx install.tsx run.tsx device.tsx uninstall.tsx memory.tsx swap.tsx types.ts
```

The memory switch goes with `memory.tsx`: it wrote `/etc/kvm/GOMEMLIMIT` through `/api/vm/memory/limit`, and the init scripts derive the limit now.

- [ ] **Step 2: Run the build to see it fail**

Run: `cd /d/projects/NanoKVM/web && npm run build`
Expected: FAIL, `Cannot find module './header.tsx'` from `tailscale/index.tsx`.

- [ ] **Step 3: Write the page**

`web/src/pages/desktop/menu/settings/tailscale/install-help.tsx`:

```tsx
import { Card } from 'antd';
import { useTranslation } from 'react-i18next';

// InstallHelp is the manual install, shown when the install from the page
// fails.
export const InstallHelp = () => {
  const { t } = useTranslation();

  return (
    <Card styles={{ body: { padding: 0 } }}>
      <p className="px-4 pt-3 text-left text-sm text-neutral-400">{t('settings.tailscale.retry')}</p>
      <ul className="list-decimal text-left font-mono text-sm text-neutral-300">
        <li>
          {t('settings.tailscale.download')}
          <a
            className="px-1"
            href="https://pkgs.tailscale.com/stable/tailscale_latest_riscv64.tgz"
            target="_blank"
          >
            {t('settings.tailscale.package')}
          </a>
          {t('settings.tailscale.unzip')}
        </li>
        <li>{t('settings.tailscale.upTailscale')}</li>
        <li>{t('settings.tailscale.upTailscaled')}</li>
        <li>{t('settings.tailscale.refresh')}</li>
      </ul>
    </Card>
  );
};
```

`web/src/pages/desktop/menu/settings/tailscale/login.tsx`:

```tsx
import { useState } from 'react';
import { UserSwitchOutlined } from '@ant-design/icons';
import { Button, Card } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/tailscale.ts';

import { ErrorDetail } from '../vpn/error-detail.tsx';

type LoginProps = {
  onSuccess: () => void;
};

export const Login = ({ onSuccess }: LoginProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [loginUrl, setLoginUrl] = useState('');
  const [errMsg, setErrMsg] = useState('');

  function login() {
    if (isLoading) return;
    setIsLoading(true);
    setErrMsg('');

    api
      .login()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }

        const url = rsp.data.url;
        if (!url) {
          onSuccess();
          return;
        }

        setLoginUrl(url);
        window.open(url, '_blank');
        setTimeout(() => setLoginUrl(''), 10 * 60 * 1000);
      })
      .catch((err) => {
        setErrMsg(err?.message || 'Failed to login');
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="flex flex-col items-center justify-center space-y-10">
      <Card>{t('settings.tailscale.notLogin')}</Card>

      {loginUrl === '' ? (
        <Button
          type="primary"
          size="large"
          shape="round"
          icon={<UserSwitchOutlined />}
          loading={isLoading}
          onClick={login}
        >
          {t('settings.tailscale.login')}
        </Button>
      ) : (
        <div className="flex w-full flex-col items-center justify-center space-y-5">
          <Button type="link" href={loginUrl} target="_blank">
            {loginUrl}
          </Button>

          <span className="text-xs text-neutral-600">{t('settings.tailscale.urlPeriod')}</span>

          <Button type="primary" size="large" shape="round" onClick={onSuccess}>
            {t('settings.tailscale.loginSuccess')}
          </Button>
        </div>
      )}

      <ErrorDetail message={errMsg} />
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/tailscale/index.tsx`:

```tsx
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/tailscale.ts';

import { VpnPage } from '../vpn/page.tsx';
import type { VpnInfo } from '../vpn/types.ts';
import { InstallHelp } from './install-help.tsx';
import { Login } from './login.tsx';

type TailscaleProps = {
  setIsLocked: (isLocked: boolean) => void;
};

export const Tailscale = ({ setIsLocked }: TailscaleProps) => {
  const { t } = useTranslation();

  const vpn: VpnInfo = {
    id: 'tailscale',
    title: 'Tailscale',
    api,
    logoutLabel: t('settings.tailscale.logout'),
    logoutWarning: t('settings.tailscale.logoutDesc'),
    renderLogin: (onSuccess) => <Login onSuccess={onSuccess} />,
    installHelp: <InstallHelp />
  };

  return <VpnPage vpn={vpn} setIsLocked={setIsLocked} />;
};
```

In `web/src/i18n/locales/en.ts`, replace the whole `tailscale: { ... },` block (from `      tailscale: {` through the `      },` before `      update: {`) with the keys the Tailscale files still use:

```ts
      tailscale: {
        title: 'Tailscale',
        retry: 'Please refresh and try again. Or try to install manually',
        download: 'Download the',
        package: 'installation package',
        unzip: 'and unzip it',
        upTailscale: 'Upload tailscale to NanoKVM directory /usr/bin/',
        upTailscaled: 'Upload tailscaled to NanoKVM directory /usr/sbin/',
        refresh: 'Refresh current page',
        notLogin:
          'The device has not been bound yet. Please login and bind this device to your account.',
        urlPeriod: 'This url is valid for 10 minutes',
        login: 'Login',
        loginSuccess: 'Login Success',
        logout: 'Logout',
        logoutDesc: 'Are you sure you want to logout?'
      },
```

The other locales keep their old `tailscale` keys; the ones nothing reads any more are harmless, and every `settings.vpn` key falls back to English.

- [ ] **Step 4: Format, build and lint**

```bash
cd /d/projects/NanoKVM/web
npx prettier --write src/pages/desktop/menu/settings/tailscale/*.tsx src/i18n/locales/en.ts
npm run build
npm run lint > /tmp/lint-after.txt 2>&1; grep -n "settings/tailscale\|settings/vpn" /tmp/lint-after.txt
grep -rn "settings.tailscale\.\(memory\|swap\|restart\|stop\|enable\|deviceName\|uninstall\|okBtn\)" src/pages
```

Expected: the build passes; no lint errors in these files; the last grep prints nothing.

- [ ] **Step 5: Commit**

```bash
cd /d/projects/NanoKVM
git add -A web/src/pages/desktop/menu/settings/tailscale web/src/i18n/locales/en.ts
git commit -m "web: the Tailscale page on the shared frame; the memory switch goes"
```

---

### Task 16: The NetBird page, its icon, and its menu entry

**Files:**
- Create: `web/src/api/extensions/netbird.ts`
- Create: `web/src/pages/desktop/menu/settings/netbird/index.tsx`, `login.tsx`
- Create: `web/src/components/icons/netbird.tsx`, `web/src/assets/images/netbird.svg`
- Modify: `web/src/pages/desktop/menu/settings/index.tsx:23`, `:32`, `:57-61`
- Modify: `web/src/i18n/locales/en.ts` (a `netbird` block after the `tailscale` block)

**Interfaces:**
- Consumes: `VpnPage`, `VpnInfo`, `ErrorDetail` (Task 14); the NetBird routes (Task 12); `POST login` with `{setupKey}` or no body, answering `{url}`.
- Produces: `Netbird({ setIsLocked })`; the `netbird` settings tab below `tailscale`.

- [ ] **Step 1: Add the tab first, to see the build fail**

In `web/src/pages/desktop/menu/settings/index.tsx`, after line 23 add:

```tsx
import { Netbird as NetbirdIcon } from '@/components/icons/netbird';
```

after `import { Network } from './network';` add:

```tsx
import { Netbird } from './netbird';
```

and after the `tailscale` tab object (lines 57 to 61) add:

```tsx
          {
            id: 'netbird',
            icon: <NetbirdIcon />,
            component: <Netbird setIsLocked={setIsLocked} />
          },
```

Run: `cd /d/projects/NanoKVM/web && npm run build`
Expected: FAIL, `Cannot find module '@/components/icons/netbird'`.

- [ ] **Step 2: Write the API module**

`web/src/api/extensions/netbird.ts`:

```ts
import { http } from '@/lib/http.ts';

// Install and update fetch Alpine's package, about 15 MB, and unpack a 43 MB
// binary onto the SD card. The default request timeout is one minute.
const LONG = { timeout: 10 * 60 * 1000 };

// install netbird
export function install() {
  return http.post('/api/extensions/netbird/install', undefined, LONG);
}

// uninstall netbird
export function uninstall() {
  return http.post('/api/extensions/netbird/uninstall');
}

// get netbird status
export function getStatus() {
  return http.get('/api/extensions/netbird/status');
}

// start netbird
export function start() {
  return http.post('/api/extensions/netbird/start');
}

// restart netbird
export function restart() {
  return http.post('/api/extensions/netbird/restart');
}

// stop netbird
export function stop() {
  return http.post('/api/extensions/netbird/stop');
}

// run netbird up
export function up() {
  return http.post('/api/extensions/netbird/up');
}

// run netbird down
export function down() {
  return http.post('/api/extensions/netbird/down');
}

// join with a setup key, or, without one, start an SSO login and get its URL
export function login(setupKey?: string) {
  return http.post('/api/extensions/netbird/login', setupKey ? { setupKey } : undefined, {
    timeout: 3 * 60 * 1000
  });
}

// netbird deregister: removes this peer from the account
export function logout() {
  return http.post('/api/extensions/netbird/logout');
}

// turn start at boot on or off
export function setBoot(enabled: boolean) {
  return http.post('/api/extensions/netbird/boot', { enabled });
}

// get the installed and the latest version
export function getUpdate() {
  return http.get('/api/extensions/netbird/update');
}

// install the latest version
export function update() {
  return http.post('/api/extensions/netbird/update', undefined, LONG);
}
```

- [ ] **Step 3: Write the icon**

`web/src/assets/images/netbird.svg` is NetBird's own mark, from `client/ui/frontend/src/assets/logos/netbird.svg` at tag v0.78.2 (BSD-3-Clause, as the client):

```svg
<svg width="31" height="23" viewBox="0 0 31 23" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M21.4631 0.523438C17.8173 0.857913 16.0028 2.95675 15.3171 4.01871L4.66406 22.4734H17.5163L30.1929 0.523438H21.4631Z" fill="#F68330"/>
<path d="M17.5265 22.4737L0 3.88525C0 3.88525 19.8177 -1.44128 21.7493 15.1738L17.5265 22.4737Z" fill="#F68330"/>
<path d="M14.9236 4.70563L9.54688 14.0208L17.5158 22.4747L21.7385 15.158C21.0696 9.44682 18.2851 6.32784 14.9236 4.69727" fill="#F05252"/>
</svg>
```

Check it against the source before committing:

```bash
curl -fsSL https://raw.githubusercontent.com/netbirdio/netbird/v0.78.2/client/ui/frontend/src/assets/logos/netbird.svg | diff - /d/projects/NanoKVM/web/src/assets/images/netbird.svg && echo same
```

`web/src/components/icons/netbird.tsx`:

```tsx
import icon from '@/assets/images/netbird.svg';

export const Netbird = () => {
  return <img src={icon} className="h-[18px] w-[18px] object-contain" alt="netbird" />;
};
```

- [ ] **Step 4: Write the login form and the page**

`web/src/pages/desktop/menu/settings/netbird/login.tsx`:

```tsx
import { useState } from 'react';
import { KeyOutlined, UserSwitchOutlined } from '@ant-design/icons';
import { Button, Card, Divider, Input } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/netbird.ts';

import { ErrorDetail } from '../vpn/error-detail.tsx';

type LoginProps = {
  onSuccess: () => void;
};

// Login joins with a setup key, or logs in with SSO. The key stays in this
// form's state and goes to the server once; nothing stores it.
export const Login = ({ onSuccess }: LoginProps) => {
  const { t } = useTranslation();

  const [setupKey, setSetupKey] = useState('');
  const [loading, setLoading] = useState<'' | 'key' | 'sso'>('');
  const [loginUrl, setLoginUrl] = useState('');
  const [errMsg, setErrMsg] = useState('');

  function join() {
    const key = setupKey.trim();
    if (loading !== '' || !key) return;
    setLoading('key');
    setErrMsg('');

    api
      .login(key)
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }
        setSetupKey('');
        onSuccess();
      })
      .catch((err) => setErrMsg(err?.message || 'Failed to join'))
      .finally(() => setLoading(''));
  }

  function sso() {
    if (loading !== '') return;
    setLoading('sso');
    setErrMsg('');

    api
      .login()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }

        const url = rsp.data?.url;
        if (!url) {
          onSuccess();
          return;
        }

        setLoginUrl(url);
        window.open(url, '_blank');
        setTimeout(() => setLoginUrl(''), 10 * 60 * 1000);
      })
      .catch((err) => setErrMsg(err?.message || 'Failed to login'))
      .finally(() => setLoading(''));
  }

  return (
    <div className="flex flex-col items-center justify-center space-y-8 pt-5">
      <Card>{t('settings.netbird.notLogin')}</Card>

      {loginUrl === '' ? (
        <div className="flex w-full max-w-[420px] flex-col space-y-4">
          <span>{t('settings.netbird.setupKey')}</span>
          <div className="flex space-x-2">
            <Input.Password
              value={setupKey}
              placeholder={t('settings.netbird.setupKeyPlaceholder')}
              autoComplete="off"
              onChange={(e) => setSetupKey(e.target.value)}
              onPressEnter={join}
            />
            <Button
              type="primary"
              icon={<KeyOutlined />}
              loading={loading === 'key'}
              disabled={!setupKey.trim() || loading === 'sso'}
              onClick={join}
            >
              {t('settings.netbird.join')}
            </Button>
          </div>

          <Divider plain>{t('settings.netbird.or')}</Divider>

          <Button
            size="large"
            shape="round"
            icon={<UserSwitchOutlined />}
            loading={loading === 'sso'}
            disabled={loading === 'key'}
            onClick={sso}
          >
            {t('settings.netbird.sso')}
          </Button>
        </div>
      ) : (
        <div className="flex w-full flex-col items-center justify-center space-y-5">
          <Button type="link" href={loginUrl} target="_blank">
            {loginUrl}
          </Button>

          <span className="text-xs text-neutral-600">{t('settings.netbird.urlPeriod')}</span>

          <Button type="primary" size="large" shape="round" onClick={onSuccess}>
            {t('settings.netbird.loginSuccess')}
          </Button>
        </div>
      )}

      <ErrorDetail message={errMsg} />
    </div>
  );
};
```

`web/src/pages/desktop/menu/settings/netbird/index.tsx`:

```tsx
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/netbird.ts';

import { VpnPage } from '../vpn/page.tsx';
import type { VpnInfo } from '../vpn/types.ts';
import { Login } from './login.tsx';

type NetbirdProps = {
  setIsLocked: (isLocked: boolean) => void;
};

export const Netbird = ({ setIsLocked }: NetbirdProps) => {
  const { t } = useTranslation();

  const vpn: VpnInfo = {
    id: 'netbird',
    title: 'NetBird',
    api,
    logoutLabel: t('settings.netbird.logout'),
    logoutWarning: t('settings.netbird.logoutDesc'),
    renderLogin: (onSuccess) => <Login onSuccess={onSuccess} />
  };

  return <VpnPage vpn={vpn} setIsLocked={setIsLocked} />;
};
```

In `web/src/i18n/locales/en.ts`, insert directly after the `tailscale` block's closing `      },`:

```ts
      netbird: {
        title: 'NetBird',
        notLogin:
          'This device has not joined a NetBird network yet. Join with a setup key, or log in with SSO.',
        setupKey: 'Setup key',
        setupKeyPlaceholder: 'Paste a setup key from the NetBird dashboard',
        join: 'Join',
        or: 'or',
        sso: 'Log in with SSO',
        urlPeriod: 'This url is valid for 10 minutes',
        loginSuccess: 'Login Success',
        logout: 'Deregister',
        logoutDesc:
          'Deregister removes this peer from your NetBird account and deletes its configuration here. Joining again needs a setup key or an SSO login, and the peer may get a new IP. Continue?'
      },
```

- [ ] **Step 5: Format, build and lint**

```bash
cd /d/projects/NanoKVM/web
npx prettier --write src/api/extensions/netbird.ts src/pages/desktop/menu/settings/netbird/*.tsx src/components/icons/netbird.tsx src/pages/desktop/menu/settings/index.tsx src/i18n/locales/en.ts
npm run build
npm run lint > /tmp/lint-after.txt 2>&1; grep -n "settings/netbird\|icons/netbird\|extensions/netbird\|settings/index" /tmp/lint-after.txt
```

Expected: the build passes; no lint errors in these files.

- [ ] **Step 6: Commit**

```bash
cd /d/projects/NanoKVM
git add web/src/api/extensions/netbird.ts web/src/pages/desktop/menu/settings/netbird web/src/components/icons/netbird.tsx web/src/assets/images/netbird.svg web/src/pages/desktop/menu/settings/index.tsx web/src/i18n/locales/en.ts
git commit -m "web: a NetBird page below Tailscale, with a setup key and SSO"
```

---

### Task 17: The dist side: NetBird's kernel requirement

`generic/S04addons` needs no change: it reads any add-on's `initd` and `enabled`. The kernel already carries WireGuard (`devices/sipeed-nanokvm/kernel.config:1145`, `CONFIG_WIREGUARD=y`, kept by the A5 audit for NetBird), and nothing holds it there. This task makes the gate hold it.

**Files:**
- Modify: `D:\projects\ironkvm-dist\devices\kernel-requirements` (after the `tailscale` line)
- Modify: `D:\projects\ironkvm-dist\devices\fake-board\kernel.config` (comment and one option)
- Modify: `D:\projects\ironkvm-dist\devices\memory-profiles` ("Tailscale needs no entry" section)
- Modify: `D:\projects\ironkvm-dist\generic\S01cgroups:14` (comment)

**Interfaces:**
- Consumes: nothing from the NanoKVM tasks.
- Produces: a `netbird` software feature that `gates/check-kernel-config.sh` holds every device to.

- [ ] **Step 1: Branch, and write the requirement first**

```bash
cd /d/projects/ironkvm-dist
git status --short   # must be empty
git switch -c feat/netbird main
```

In `devices/kernel-requirements`, after the line `tailscale             software    CONFIG_TUN`, add:

```
# netbird is off by default for the same reason as tailscale. It needs the tun
# device for its user-space fallback and the kernel's WireGuard for the
# interface it uses on this board (usesKernelInterface in `netbird status`,
# measured 2026-09-27). Without WIREGUARD it falls back to user space, which
# costs memory the addons group does not have.
netbird               software    CONFIG_TUN CONFIG_WIREGUARD
```

- [ ] **Step 2: Run the gate to see the fixture fail**

```bash
cd /d/projects && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects:/p" -w /p/ironkvm-dist alpine:3.24 sh gates/check-kernel-config.sh; echo exit=$?
```

Expected: exit 1, `fake-board` misses `CONFIG_WIREGUARD` (sipeed-nanokvm passes: it has `=y`). If it exits 2, read its last line: it names what is missing (for example the NanoKVM checkout beside this one), and a 2 is not a pass.

- [ ] **Step 3: Update the fixture and the comments**

In `devices/fake-board/kernel.config`, change the sentence `# The four options that follow are the software features, which are part of the` to `# The five options that follow are the software features, which are part of the`, and add after `CONFIG_TUN=y`:

```
CONFIG_WIREGUARD=y
```

In `devices/memory-profiles`, replace the section heading and first sentence:

```
# == Tailscale and NetBird need no entry ==
#
# Both are off by default on every device, decided on 2026-09-17 for Tailscale
# and on 2026-09-27 for NetBird, and the reason is that a board joins a network
# when somebody joins it rather than when it boots. That is not a memory
# question, so writing it here would say that a board with more memory would
# have them on, which is false. That only one of them runs at a time is a
# memory question, and the web UI enforces it, because the addons group holds
# one of them near its limit.
```

In `generic/S01cgroups`, line 14, change `#   addons  Tailscale and PicoClaw.` to `#   addons  Tailscale, NetBird and PicoClaw.` (keep the rest of the comment).

- [ ] **Step 4: Run the gates**

```bash
cd /d/projects && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects:/p" -w /p/ironkvm-dist alpine:3.24 sh -c 'sh gates/check-kernel-config.sh; echo check=$?; sh run-tests.sh gates'
```

Expected: `check=0`; every `gates` suite PASS (a SKIP line must name why, and is not a pass).

- [ ] **Step 5: Commit**

```bash
cd /d/projects/ironkvm-dist
git add devices/kernel-requirements devices/fake-board/kernel.config devices/memory-profiles generic/S01cgroups
git commit -m "kernel-requirements: netbird needs TUN and WIREGUARD (#1)"
```

---

### Task 18: On the board

Every step here is on root@10.0.0.222. Stopping Tailscale needs the owner's go-ahead (Step 4). The setup key is the one the owner gave; type it into the page only, never into a shell, a file or a commit.

**Files:** none changed. This task deploys and checks.

**Interfaces:**
- Consumes: the server build, the web build, `S98netbird`, `S98tailscaled` from Tasks 1 to 16.
- Produces: the evidence the spec's board tests ask for, reported to the owner.

- [ ] **Step 1: Run every automated suite once more**

```bash
cd /d/projects/NanoKVM
# gotest from Global Constraints:
cd server && MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM\\server:/src" -v nanokvm-gomod:/go/pkg/mod -v nanokvm-gocache:/root/.cache/go-build -w /src golang:1.25 go test ./proto/ ./router/ ./service/extensions/... ./utils/ && cd ..
for s in service/test-netbird service/test-addon-memlimit service/test-cgroup-join service/test-tailscaled release/test-init-install-list; do
  MSYS_NO_PATHCONV=1 docker run --rm -v "D:\\projects\\NanoKVM:/r" -w /r alpine:3.24 sh tools/$s.sh >/dev/null; echo "$s exit=$?"
done
cd web && npm run build && cd ..
```

Expected: Go PASS; every shell suite exit 0; the web build passes.

- [ ] **Step 2: Build and deploy**

Record Tailscale's state first, so it can be put back exactly:

```bash
ssh root@10.0.0.222 'tailscale status --json | grep -m1 BackendState; tailscale ip -4; ls /data/ironkvm/addons/tailscale/enabled 2>&1; cat /etc/kvm/GOMEMLIMIT 2>&1'
```

Write the four answers down (running state, IPv4, whether `enabled` exists, the GOMEMLIMIT file).

Build the server with the board build command from Global Constraints. Then deploy the init scripts (package copy only; `S98netbird` is not-installed-by-policy, so never into `/etc/init.d` by hand), the web, and the server:

```bash
cd /d/projects/NanoKVM
scp kvmapp/system/init.d/S98netbird root@10.0.0.222:/kvmapp/system/init.d/S98netbird
scp kvmapp/system/init.d/S98tailscaled root@10.0.0.222:/kvmapp/system/init.d/S98tailscaled
ssh root@10.0.0.222 'chmod 755 /kvmapp/system/init.d/S98netbird /kvmapp/system/init.d/S98tailscaled; [ -f /etc/init.d/S98tailscaled ] && cp /kvmapp/system/init.d/S98tailscaled /etc/init.d/S98tailscaled; echo ok'
scp -r web/dist root@10.0.0.222:/kvmapp/server/web.new
ssh root@10.0.0.222 'rm -rf /kvmapp/server/web.old && mv /kvmapp/server/web /kvmapp/server/web.old && mv /kvmapp/server/web.new /kvmapp/server/web'
scp server/NanoKVM-Server root@10.0.0.222:/data/NanoKVM-Server.new
scp tools/deploy/deploy-server root@10.0.0.222:/data/
ssh root@10.0.0.222 'DEPLOY_TIMEOUT=240 sh /data/deploy-server /data/NanoKVM-Server.new'
```

Expected: deploy-server reports the new binary serving. `tailscale status` is unchanged (the server restart does not touch tailscaled).

- [ ] **Step 3: With Tailscale running, the NetBird page refuses**

In the web UI, open Settings. Check:
- A NetBird entry with its icon sits below Tailscale.
- The Tailscale page shows Running, the IP recorded in Step 2, peers, version, uptime, the memory bars against 64 MiB, and the start-at-boot switch matching the recorded `enabled` answer. The more menu has the swap switch and no memory-optimization switch.
- The NetBird page shows "Tailscale is running or starts at boot..." and the Install button is disabled.

- [ ] **Step 4: Ask the owner before stopping Tailscale**

Ask: "The next steps stop Tailscale and turn off its start at boot for the length of the test, about 20 minutes and one reboot, and put both back afterwards. Go ahead?" Wait for a yes. Without it, stop here and report Steps 1 to 3.

With the go-ahead: on the Tailscale page, turn start at boot off, then stop Tailscale. Check the NetBird page's notice is gone.

- [ ] **Step 5: Install and join**

On the NetBird page, press Install. Expected within a few minutes: the page moves to the login form. On the board:

```bash
ssh root@10.0.0.222 'ls -l /usr/bin/netbird; ls -la /data/ironkvm/addons/netbird; apk info -e netbird; echo apk_installed=$?; cat /var/run/netbird.pid; tr "\0" "\n" < /proc/$(cat /var/run/netbird.pid)/environ | grep -E "GOMEMLIMIT|NB_STATE_DIR"; cat /proc/$(cat /var/run/netbird.pid)/cgroup'
```

Expected: `/usr/bin/netbird` is a link into `/data/ironkvm/addons/netbird/netbird`; the add-on directory has `links`, `initd`, the binary, and no `.fetch`; `apk info -e netbird` prints nothing and `apk_installed=1` (nothing installed into the slot); `GOMEMLIMIT=56MiB` (or the operator's lower value) and `NB_STATE_DIR=/data/identity-system/netbird`; the cgroup line ends in `/addons`.

Paste the setup key the owner gave into the Setup key field and press Join. Expected: the page shows Running, control Connected, an IP in 100.64.0.0/10, the name `*.netbird.cloud`, account `api.netbird.io`, version 0.78.2.

```bash
ssh root@10.0.0.222 'ls -la /data/identity-system/netbird; ls -la /var/lib/netbird 2>&1; netbird status | head -12; ip -br link show wt0'
```

Expected: the profile files are under `/data/identity-system/netbird`; `/var/lib/netbird` holds nothing of NetBird's identity; `wt0` exists.

- [ ] **Step 6: The identity survives a server restart and a daemon restart**

```bash
ssh root@10.0.0.222 'netbird status --json | grep -o "\"netbirdIp\":\"[^\"]*\""; /etc/init.d/S95nanokvm restart; sleep 20; netbird status --json | grep -o "\"netbirdIp\":\"[^\"]*\""'
```

Then press Restart on the NetBird page, wait for Running, and compare the IP. Expected: the same IP all three times, and the owner's NetBird dashboard lists one peer for this board, not two.

- [ ] **Step 7: The notice blocks Tailscale while NetBird runs**

Open the Tailscale page. Expected: "NetBird is running or starts at boot..." and the Start button disabled. Force the server-side check too (the admin session cookie is in the browser; use the browser's devtools console on the KVM page):

```js
fetch('/api/extensions/tailscale/start', { method: 'POST' }).then((r) => r.json()).then(console.log)
```

Expected: `code: -1` and a message that names NetBird.

- [ ] **Step 8: Start at boot survives a reboot**

On the NetBird page turn start at boot on. Then:

```bash
ssh root@10.0.0.222 'ls -l /data/ironkvm/addons/netbird/enabled /etc/init.d/S98netbird; reboot'
```

After the board is back (about two minutes): the NetBird page shows Running with the same IP; `ssh root@10.0.0.222 'cat /proc/$(cat /var/run/netbird.pid)/cgroup; pgrep tailscaled || echo no-tailscaled'` shows `/addons` and `no-tailscaled`.

- [ ] **Step 9: Update reports current equal to latest, and the page reads well**

The NetBird header shows no update icon. In the devtools console:

```js
fetch('/api/extensions/netbird/update').then((r) => r.json()).then(console.log)
```

Expected: `current: "0.78.2", latest: "0.78.2"`. The memory bars show the daemon at about 30 MB and the group against 64 MiB; press Stop, then Start, and check the start-at-boot switch did not move.

- [ ] **Step 10: Put Tailscale back**

With the owner's go-ahead from Step 4 still standing: on the NetBird page turn start at boot off and press Stop (leave NetBird installed and joined unless the owner says to uninstall; do not press Deregister, which deletes the peer from the owner's account). Then on the Tailscale page press Start, and turn start at boot back on if the recorded `enabled` said so.

```bash
ssh root@10.0.0.222 'tailscale status --json | grep -m1 BackendState; tailscale ip -4; ls /data/ironkvm/addons/tailscale/enabled 2>&1; ls /data/ironkvm/addons/netbird/enabled 2>&1; pgrep -f "netbird service run" || echo netbird-stopped'
```

Expected: the four answers from Step 2 again (Running, the same IPv4, `enabled` as it was), NetBird not enabled and stopped. Report to the owner: each step's result, and anything that differed from "Expected".

---

## Self-review

Spec coverage, section by section:

| Spec requirement | Task |
| --- | --- |
| One at a time; one check in `addon`, called from install, start, up and boot | 5 (check), 8 and 12 (called from install, start, up, login, boot on) |
| The page shows the notice before the button is pressed | 6 (`BlockedBy` in `Fill`), 14 (`Notice`, disabled buttons) |
| NetBird from Alpine: `apk fetch` into `/data`, signature, extract `usr/bin/netbird`, linked via add-on records, no package in the slot | 1 (method), 10, 18 Step 5 |
| Identity at `/data/identity-system/netbird/` | 13 (`S98netbird`), 18 Steps 5 and 6 |
| GOMEMLIMIT derived at 7/8 of `memory.high`; free-text field removed | 13 (both scripts, one test), 8 (server stops writing it), 15 (switch removed) |
| Logout is `netbird deregister`, with a warning | 11, 12, 16 (`logoutDesc`) |
| `VpnStatus` shape; `GetTailscaleStatusRsp` keeps four fields | 2 |
| Routes incl. `POST boot`, `GET`/`POST update`; start and stop no longer change boot | 8, 9, 12 |
| NetBird login: `{setupKey}` or SSO URL from the CLI output | 4 (reader), 11, 12, 16 |
| Update: Tailscale from pkgs.tailscale.com, NetBird via `apk update`; one-hour cache; restart only if it ran | 6, 9, 10, 12 |
| Error tail in every failing handler | 3, 8, 12 |
| `vpn` package: uptime and RSS from `/proc`, group memory, login URL reader, output tail | 3, 4, 6 |
| `S98netbird` modeled on `S98tailscaled`, falls back to `/var/lib/netbird`, handled by `S04addons` unchanged | 13; 17 notes S04addons |
| Web: shared frame, header, boot switch, status, peers, memory bars, notice, error detail; swap on both; NetBird below Tailscale with its own icon; English only | 14, 15, 16 |
| Testing: Go, shell, board | every task; 18 |

Decisions this plan adds where the spec is silent, each for a reason found in the code:
- `S98netbird` exports `NB_STATE_DIR` beside `--config`: NetBird 0.78 keeps `default.json`, `active_profile.json` and `state.json` in the profile directory, which `--config` alone does not move (`client/internal/profilemanager/service.go` at v0.78.2). Task 1 Step 6 checks it on the board.
- `login` is refused too, since it starts the daemon.
- Start, stop and restart run the script in `/kvmapp/system/init.d`, and `/etc/init.d` holds a copy only while start at boot is on, so off a distribution image the copy is the boot record.
- The "free-text memory limit field" in the code is a switch (`tailscale/memory.tsx`) that wrote 75 to `/etc/kvm/GOMEMLIMIT` through `/api/vm/memory/limit`. The switch goes; the route stays, because `server/main.go` also reads that file for the server's own limit, and that is outside this spec.
- `GET update` has the default one-minute request timeout: `http.get` in `web/src/lib/http.ts` takes no config.
