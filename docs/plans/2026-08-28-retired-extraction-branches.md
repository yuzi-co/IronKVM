# Retired extraction branches, 2026-08-28

The fork stopped contributing to `sipeed/NanoKVM` on 2026-08-28. The eight open pull requests were
closed, and the forty-eight extraction branches were retired from the local repository on the same
day. This file is the record of what they were and how to get any of them back.

An extraction branch was a single-purpose copy of work that already lives in `main`, shaped so a
pull request to `sipeed/NanoKVM` showed only that change. `AGENTS.md` explains the two branch kinds.
Nothing here is fork work that `fork/integration` lacks, with one exception named below.

## Nothing was destroyed

Each branch is preserved twice, and the two copies are not the same:

- **`origin` still has every one of the forty-eight**, at the commit it had before 2026-08-28. The
  fork was never pushed after the rebase of that day, so `origin` holds the **pre-rebase** version.
  These are the commits the closed pull requests point at, which is why they were left alone: a
  closed pull request can be reopened while its head branch exists, and deleting the branch would
  end that.
- **`refs/archive/extractions-20260828/<branch>` holds the rebased version**, local to this clone.
  That is the 2026-08-28 work: forty-six of the forty-eight moved onto `upstream/main` at
  `2ba45a21`, verified with the novision gate and, for the three C++ branches, with a MaixCDK build.
  This copy exists nowhere else. If this clone is lost, the rebase is lost and `origin` is what
  remains.

`refs/rebase-backup/20260828/<branch>` also survives and holds the pre-rebase heads, which is a
third copy of what `origin` has.

Restore a branch from either copy:

```shell
git checkout -b <branch> refs/archive/extractions-20260828/<branch>   # rebased, 2026-08-28
git checkout -b <branch> origin/<branch>                              # pre-rebase, what the PR shows
```

## The one branch that holds unique work

`fix/hdmi-signal-reported` carries the fork's own HDMI signal reader, `server/utils/hdmi.go` and its
48-line `hdmi_test.go`. `main` dropped both in `a189c906` because upstream replaced them in `#859`,
so that branch is the only copy of the fork's version. It is superseded rather than lost, and it is
recorded here because deleting it on the "already in main" signal would have been wrong.

`security/password-change-requires-current` is the other branch that was never rebased. Upstream
implemented the same requirement in `#876`, so it asks for what upstream already does. The files it
appears to hold uniquely, `server/service/auth/account.go` and `web/src/lib/cookie.ts`, are
upstream's own files from before `#876` deleted them, not fork work.

## The forty-eight

`local` is the rebased tip in `refs/archive/extractions-20260828/`. `origin` is the pre-rebase tip
still on the fork. They differ for forty-six of the forty-eight; they match for the two branches
that were found dead and deliberately not rebased.

| Branch | local (rebased) | origin (pre-rebase) | Tip subject |
| --- | --- | --- | --- |
| `build/git-build-stamp` | `5b959b11` | `8ec5547c` | Stop the container build asking git for a VCS stamp |
| `build/novision-tag` | `d8082d55` | `8887fbd6` | Add a build tag that stubs the capture bindings |
| `feat/device-http-proxy` | `150d6959` | `1681ad13` | Let the device reach the internet through a proxy |
| `feat/mcp-protocol-2026-07-28` | `7e6b20ac` | `015cdb93` | Pin the newest MCP protocol revision the server offers |
| `feat/zram-swap` | `43f91fca` | `849253b9` | Look for the zram modules where an update will not remove them |
| `fix/capture-gate` | `9e3897c7` | `93fe12b5` | Guard the cgo capture boundary against teardown |
| `fix/etc-kvm-file-modes` | `81dcbe02` | `f96c7bd0` | Correct the mode /etc/kvm is created with |
| `fix/formatting-faults` | `f0a4bc1d` | `b81f04c8` | Fix the three formatting faults the line-ending noise was hiding |
| `fix/gomemlimit-not-restored` | `ad82144a` | `afebe8c0` | Apply the configured memory limit at startup |
| `fix/hdmi-probe-spin` | `09d75ae4` | `16996322` | Stop the resolution probe walking the list against a dead input |
| `fix/hdmi-signal-reported` | `a7274b7f` | `a7274b7f` | Report whether the HDMI port is actually carrying a picture |
| `fix/hdmi-switch-refusal` | `769e34b7` | `d423c330` | Notice when the library refuses to switch HDMI |
| `fix/hid-endpoint-reporting` | `b935a56e` | `6838a915` | Stop allocating on every HID report to check for a deleted node |
| `fix/hid-gadget-rebuild` | `319d63e3` | `0e52464b` | Keep HID working across a gadget rebuild, and describe it honestly |
| `fix/hid-queue-never-blocks` | `55d46bd1` | `8a4e4f18` | Never block the websocket read loop on a full HID queue |
| `fix/jiggler-race` | `c9d89dc1` | `d24dcb68` | Guard the mouse jiggler's state, and start only one loop |
| `fix/kvmv-thread-join` | `93c66bc0` | `13fa421c` | Wait for libkvm's threads before dismantling what they use |
| `fix/oled-sleep-never` | `5211cc71` | `d3005191` | Stop the OLED sleep API from silently meaning "never" |
| `fix/p2-resize-guard` | `b6fb526d` | `9d202fd9` | Resize p2 only when it can still grow |
| `fix/power-button-hold-bound` | `3b9b3e73` | `482a95c6` | Bound how long a power button can be held, and stop presses overlapping |
| `fix/stream-stalled-viewer` | `14152cc2` | `5b869cc2` | Stop one stalled viewer from freezing the stream, and fix the screen race |
| `fix/terminal-socket-leak` | `7ee28045` | `4f06b40e` | Close the terminal socket when its writer gives up |
| `fix/vi-init-race` | `a5ac79bb` | `380e65d3` | Serialise mmf_vi_init and mmf_vi_deinit |
| `hardening/http-server-timeouts` | `1bc00f71` | `bed913f7` | Bound how long a client may take to send its request headers |
| `hardening/websocket-read-limits` | `a0afd883` | `5afb1587` | Bound the size of one frame on the HID event socket too |
| `perf/capture-log-flood` | `8f718af7` | `2e903d01` | Say a repeated capture read failure once, not once per frame |
| `perf/frame-copy-reduction` | `30e0e406` | `2ed9159f` | Stop copying every video frame more times than needed |
| `perf/frame-detect-async` | `3c358e1e` | `12ae0bd2` | Stop holding a request open to pause frame detection |
| `perf/frame-rate-write-on-change` | `1915214c` | `3b22c65e` | Write the frame rate counter only when it changes |
| `perf/h264-no-per-frame-copy` | `498d7a44` | `4c25cc54` | Stop building a second copy of every H.264 frame |
| `perf/hid-report-no-alloc` | `d7f8da6d` | `0a469d64` | Stop allocating per HID report to reach the device handle |
| `perf/logger-caller-off` | `eff9d74a` | `bb221892` | Report the caller only at the levels that ask for it |
| `perf/login-hash-once` | `472daa8e` | `8fda4fa0` | Derive the default account hash once instead of per login attempt |
| `perf/mjpeg-per-client` | `2c0c206d` | `4c98a805` | Give every MJPEG client its own queue and writer goroutine |
| `perf/now-fps-off-sd-card` | `888cf1cb` | `2ed5ef4c` | Keep the frame rate counter off the SD card |
| `perf/static-skips-api` | `7e292b48` | `195d9219` | Keep the static middleware off the API paths |
| `perf/web-cache-headers` | `75e3e913` | `1e58e035` | Tell browsers which parts of the web UI may be cached |
| `perf/webrtc-per-viewer-writer` | `86e476a5` | `04c9431d` | Give every WebRTC viewer its own writer, and packetize each frame once |
| `perf/webrtc-shared-packetizer` | `86e476a5` | `aea57d94` | Give every WebRTC viewer its own writer, and packetize each frame once |
| `security/api-key-auth` | `91889fd5` | `57f77ed7` | Let scripts authenticate with a key instead of a session |
| `security/contain-request-paths` | `a060d734` | `f749d018` | Contain the file paths that arrive in a request |
| `security/download-verify` | `eac51e23` | `bbf56726` | Verify what gets downloaded before trusting or storing it |
| `security/jwt-hardening` | `bafd275f` | `3972c6ad` | Refuse a guessable JWT signing key, and test the algorithm pin |
| `security/login-lockout-eviction` | `a8796749` | `d525056b` | Keep a login lockout when other addresses fill the record table |
| `security/password-change-requires-current` | `e09f00ec` | `e09f00ec` | Require the current password to change the password |
| `security/usb-gadget-identity` | `c80aef29` | `2815ead9` | Give the USB gadget a safe default and a stable identity |
| `security/usb-mass-storage-default` | `c02e5643` | `2a64275c` | Stop exposing the raw eMMC partition as USB mass storage |
| `security/websocket-origin-check` | `125baded` | `2b98aa85` | Hold the whole API to the same origin, not only the websocket upgrades |

