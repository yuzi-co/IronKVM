# External component audit and upgrade plan, 2026-08-29

Every third-party component this fork ships, builds with, or runs on, measured against its current
upstream release. The audit found one reachable vulnerability in the server, two open advisories in
the web UI that no patch release fixes, and three build inputs that no version pins.

Measured on `fork/integration` at `48785d8a`, application version 1.0.2.

## The seven layers

| Layer | Where it is pinned | Count | Behind |
| --- | --- | --- | --- |
| Go modules | `server/go.mod` | 22 direct, 68 indirect | 14 direct |
| npm packages | `web/package.json`, `web/pnpm-lock.yaml` | 53 | 35, of which 30 by a major |
| Native libraries | `server/dl_lib/` (committed binaries) | 40 | see below |
| Build toolchain | `docker/Dockerfile`, `tools/*/Dockerfile` | 7 inputs | 3 unpinned |
| CI actions | `.github/workflows/` | 4 actions | all 3 majors behind |
| Device rootfs | Sipeed firmware image, not built here | 5 packages | not ours to bump |
| Runtime download | Tailscale, fetched by the device | 1 | self-updating |

## Findings, worst first

### 1. A reachable vulnerability in the server binary

`govulncheck` reports one advisory that the fork's own code calls:

```
GO-2026-5970  Infinite loop on invalid input in golang.org/x/text
  Found in: golang.org/x/text@v0.37.0     Fixed in: golang.org/x/text@v0.39.0
  utils/proxy.go:60:31: utils.proxyFor calls httpproxy.Config.ProxyFunc,
                        which eventually calls norm.Form.Bytes
```

The path is real. `server/utils/proxy.go` resolves a proxy for every outbound request, so the update
feed and the Tailscale download both cross it. The input that reaches `x/text` is the request URL
and the operator's configured proxy string, so the exposure is small on a device that only calls out
to two known hosts. The fix costs nothing, so the small exposure is not a reason to wait.

Three more advisories are present in required modules but are not called: one in `x/net`, two in
`x/crypto`. One of the `x/crypto` pair, `GO-2026-5932`, has no fixed version at all.

**This fix shipped on 2026-08-29**, in `def3b03e`, merged as `2681b116`. The whole novision gate
passed: `go vet`, the 28 test packages, and a riscv64 cross-build. `govulncheck` went from one
called vulnerability to none. The build is on the device, confirmed by comparing the sha256 of
`/proc/<pid>/exe` against the local binary.

```shell
go get golang.org/x/text@v0.41.0
go get golang.org/x/net@v0.58.0
go get golang.org/x/crypto@v0.55.0
go get golang.org/x/sys@v0.47.0
go mod tidy
```

### 2. Two open advisories in the web UI that need a major version

GitHub reports two moderate alerts against `react-router`, and both name the same fixed version:

| Advisory | Vulnerable range | First patched |
| --- | --- | --- |
| Open redirect via backslash in `<Link>` and `useNavigate` (CVE-2025-68470 bypass) | `>= 6.0.0, < 7.18.0` | 7.18.0 |
| Arbitrary constructor injection in `deserializeErrors()` on SSR hydration | `>= 6.4.0, < 7.18.0` | 7.18.0 |

The lockfile holds `react-router-dom@6.30.6`, which is the last 6.x release. **No 6.x patch exists**,
so closing these alerts is a version 6 to version 7 migration and not a lockfile bump.

The second advisory does not apply to this product. It needs server-side rendering, and this
frontend is a hash-routed single page application that gin serves as static files. There is no
hydration path. The first advisory is a real class of bug, though every route here is internal.

The migration is small. Ten files import the package, and they use eight APIs:

```
34 json          10 useNavigate     4 redirect      4 Outlet
 3 Navigate       2 useSearchParams 2 createHashRouter/RouterProvider   1 Link
```

Version 7 keeps all of them. The work is the package rename from `react-router-dom` to
`react-router`, and the future flags that version 6.30 already warns about.

### 3. Three build inputs that no version pins

The builder image is not reproducible, and two of its three downloads are not verified.

| Input | How it is fetched | Problem |
| --- | --- | --- |
| MaixCDK | `git clone https://github.com/Sipeed/MaixCDK` in `docker/Dockerfile:58` | No tag and no SHA. Every image rebuild takes whatever the default branch holds. Upstream tagged `v4.11.3` on 2025-07-29, and the branch has moved to `30f4b8b7e3` since. |
| host-tools (the riscv64 musl toolchain) | `wget` from `sophon-file.sophon.cn`, in two Dockerfiles | No checksum. A changed or replaced tarball would be built without notice. |
| Go tarball | `wget` from `go.dev`, version pinned to 1.25.0 | Version is pinned in two places that must stay in step. No checksum. |

Only `.devcontainer/Dockerfile` verifies anything, and only for Node. Outside Docker,
`tools/opusbench/build.sh` does verify the Opus tarball with `sha256sum -c`, which is the pattern the
other three should follow.

The MaixCDK clone is the one that matters most. `libkvm.so` is built against it, the AGENTS.md
warning about `./build update_lib` already shows how quietly a MaixCDK mismatch fails, and an
unpinned clone means two builds of the same commit of this repository can produce different
libraries.

### 4. The Go modules that are behind but hold no advisory

Fourteen direct dependencies:

```
gin                1.10.0    -> 1.12.0        pion/webrtc/v4   4.0.1   -> 4.2.19
validator/v10      10.20.0   -> 10.30.3       pion/rtp         1.8.18  -> 1.10.5
viper              1.19.0    -> 1.21.0        pion/dtls/v3     3.1.4   -> 3.1.8
logrus             1.9.3     -> 1.10.2        mcp go-sdk       1.6.1   -> 1.7.0
unrolled/secure    1.15.0    -> 1.17.0        gin-gonic/contrib, rs/cors/wrapper/gin (both pseudo-versions)
```

None is a major version step. `pion/webrtc` is the largest jump and the one with real behaviour
behind it, because it carries the whole WebRTC stack in its indirect set.

### 5. The frontend is behind almost everywhere, and none of it needs React 19

Thirty-five of fifty-three packages are behind, thirty of them by a major version. The important
finding is what the peer dependencies allow:

| Package | Now | Latest | Accepts React 18 |
| --- | --- | --- | --- |
| antd | 5.29.3 | 6.6.2 | yes, `react@>=18.0.0` |
| react-router | 6.30.6 | 7.18.3 | yes, `react@>=18` |
| @ant-design/icons | 5.6.1 | 6.3.2 | yes |
| react-error-boundary | 4.1.2 | 6.1.3 | yes |
| react-helmet-async | 2.0.5 | 3.0.0 | yes |
| lucide-react | 0.562.0 | 1.37.0 | yes |
| vaul | 0.9.9 | 1.1.2 | yes |
| @xterm/xterm | 5.5.0 | 6.0.0 | no peer constraint |
| i18next | 23.16.8 | 26.4.0 | not a React package |
| react-i18next | 14.1.3 | 17.0.12 | yes, `react@>= 16.8.0` |

**Nothing forces the React 18 to 19 jump.** The whole runtime stack can move to its current major
while React stays where it is. That separates the risk cleanly: the library upgrades can be judged
on their own, and React 19 becomes a decision to take later on its own merits.

The build-time packages are a separate group and change no shipped code: `tailwindcss` 3 to 4,
`typescript` 5 to 7, `eslint` 9 to 10 with its plugins.

### 6. The CI actions are three majors behind

`actions/checkout`, `actions/setup-node` and `actions/upload-artifact` are all at `v4`, against `v7`
upstream, and `actions/download-artifact` is at `v4` against `v8`. No advisory drives this. The cost
of leaving them is that GitHub eventually retires a runner feature the old action needs, and the
release workflow is the thing that breaks.

### 7. The device rootfs is old, and it is not this repository's to change

| Component | On the device | Note |
| --- | --- | --- |
| Linux | 5.10.4, built 2025-04-17 | Sipeed vendor kernel for the SG2002 |
| Buildroot | 2023.11.2 | |
| BusyBox | 1.36.1 | |
| OpenSSH | 9.6p1, released 2023-12-18 | in the CVE-2024-6387 range |
| OpenSSL | 3.1.4 | the 3.1 series is past its end of support |
| watchdog | 5.16 | current; this is the last release |

None of these is built here. The four deliverables in this repository are `server/`, `web/`,
`support/sg2002/` and `kvmapp/`. Everything in the table comes from the Sipeed firmware image, so
moving any of it means either a Sipeed release or a Buildroot rootfs the fork builds itself.

That is a much larger change than it looks, and two known facts price it. A rootfs rebuild silently
removes the zram modules, because they are hand-installed on the root slot. A rebuild also reverts
identity: `/etc/shadow`, `authorized_keys` and the ssh host key all go back to factory. Neither is a
reason never to do it, and both are reasons not to do it as part of a dependency sweep.

### 8. Two components that are already current

- **Opus is at 1.5.2**, which is the latest release, from 2024-09-11. `tools/opusbench/build.sh`
  pins the version and verifies the tarball. Nothing to do.
- **Tailscale needs no pin.** The device holds 1.80.2 against 1.102.3 upstream, but
  `S98tailscaled` and `install.go` both fetch `tailscale_latest_riscv64.tgz`. This is a stale
  install and not a stale pin: a reinstall from the UI takes the current release.

The forty committed libraries in `server/dl_lib/` come from the CVITEK/Sophon SDK and move only when
MaixCDK moves. Pinning MaixCDK, in the plan below, is what gives them a version at all.

## The plan

Six phases. Each is one branch off `main`, merged into `fork/integration` with `git merge --no-ff`,
in the order below. The order is by risk-adjusted value and not by size.

Phases 1, 3 and 4 change the server binary, so each one deploys through
`tools/deploy/deploy-server` and is verified from `/proc/PID/exe`. Phases 2 and 5 change only the web
UI, which cannot wedge the board. Phase 6 changes nothing that runs on the device.

### Phase 1: the Go security bumps (done, 2026-08-29)

`fix/go-security-bumps`. Four `go get` lines. Closes the one reachable advisory and two of the three
unreachable ones.

Gate: `go vet -tags novision`, `go test -tags novision ./...`, the riscv64 cross-build, and
`govulncheck` reporting no called vulnerabilities. Then deploy and confirm from `/proc/PID/exe`.

All of it passed. The deploy guard reported "OK, serving within 240s and running what was
installed", and the sha256 of the running image matches the local build. Cost: about an hour, most
of it the deploy.

### Phase 2: react-router 6 to 7 (done, 2026-08-29)

`fix/react-router-advisories`. Closes both open Dependabot alerts, which is the only reason this
outranks the rest of the frontend.

Rename the import in ten files, drop `react-router-dom` for `react-router`, and clear the version 6
future-flag warnings. `createHashRouter` and `RouterProvider` carry over unchanged.

Gate: `pnpm build`, then a real device with a deployed `web/`. Walk login, the stream page, settings
and the Wi-Fi provisioning page, because `ProtectedRoute` and the redirect behaviour are exactly what
this package owns.

Done. 7.18.2 rather than 7.18.3: both carry the fix, and taking the newer one
would have meant writing a release-age bypass into pnpm-workspace.yaml. The
device walk-through covered the redirect, the nested admin route, the query
string and the lazy chunks. The pages behind the login wall were not covered.
Both GitHub alerts closed once the fix reached `main`, which is the branch
Dependabot scans.

### Phase 3: pin the build inputs (done, 2026-08-29)

`fix/pin-build-inputs`. The supply-chain finding, and the phase with the longest tail if it is left.

- Pin MaixCDK to a tag or a SHA in `docker/Dockerfile`, and record which one in `AGENTS.md`
  beside the existing `update_lib` warning.
- Add a `sha256sum -c` check to the host-tools download in both `docker/Dockerfile` and
  `tools/build/Dockerfile`, in the shape `tools/opusbench/build.sh` already uses.
- Add the same check to the Go tarball, and reduce the version to one place.

Gate: rebuild the builder image from scratch, then `make app` and `make vision`, then compare
`patchelf --print-needed libkvm.so` against the committed library. AGENTS.md already states that the
two lists must agree, and this is the change most likely to move them.

Do this before Phase 4 and not after. A dependency bump that is investigated on top of an
unpinned toolchain cannot be told apart from a toolchain that moved underneath it.

Done, and the gate held. Both images were rebuilt from scratch, the two
checksums verified, and a libkvm.so rebuilt from the pinned MaixCDK commit has
a NEEDED list identical to the committed library, all twenty-three entries.
`tools/build/test-pinned-inputs.sh` keeps the pins in place and catches all
seven mutations tried against it.

Two extra things came out of it. `.dockerignore` did not exclude `.pnpm-store`
or the opus build output, and each of them stops `docker build` outright on
Windows. The go download also asked for `go1.25.0.linux-aarch64.tar.gz` on an
arm64 host, a file that has never existed.

### Phase 4: the rest of the Go direct dependencies (done, 2026-08-29)

`chore/go-dependency-refresh`. Fourteen modules, no major steps, no advisories.

Take `pion/webrtc` on its own commit inside the branch. It is the only one whose failure mode is a
stream that negotiates and then carries nothing, which no test here would catch.

Gate: the novision suite, then hardware acceptance on all three video paths. `capture-costs-80-percent-of-core`
gives the baselines to compare against: MJPEG about 90% of the core, H.264 direct about 24%, no
viewer about 7%.

Done, and two of the fourteen were held back rather than taken.

**gin stays at 1.10.0.** Every release after it links `quic-go/http3`, in 1.11 as well as 1.12,
and 1.12 also pulls mongo-driver in through `gin/binding`. This board will never speak QUIC or
read BSON. With gin 1.12 in the tree govulncheck reported GO-2026-5676 in quic-go as reachable,
against nothing reachable before it, and gin 1.10.0 carries no advisory of its own. The trade was
backwards.

**modelcontextprotocol/go-sdk stays at 1.6.1.** 1.7.0 stops returning an `Mcp-Session-Id` header
on initialize. That is a protocol change on the surface PicoClaw drives, so it needs its own
branch.

The other twelve moved. The three-path acceptance ran against the deployed build with a live HDMI
signal, in 20-second samples from `/proc/stat`:

| path | busy | delivered |
| --- | --- | --- |
| no viewer | 6.0% | n/a |
| MJPEG | 98.2% | 48.9MB in 20s |
| H.264 direct | 18.9% | 285 KB/s |

WebRTC negotiated fully: an 883-byte answer carrying H.264, `sendonly`, a DTLS fingerprint and the
playout-delay extension, with three ICE candidates trickled. SRTP media flow is the one thing a
scripted client cannot exercise, because it needs a real ICE and DTLS handshake.

The `service/ws` heartbeat test failed once in the full run and passes five times out of five on
its own. That is the known load flake, not this change.

### Phase 5: the frontend runtime majors, with React held at 18 (done, 2026-08-29)

`chore/frontend-majors`. Ten packages, listed in finding 5. React and `react-dom` stay at 18.3.1, and
`@types/react` stays at 18.

Take `antd` 5 to 6 as its own commit. It is the largest single change in the tree by rendered
surface, and it is the one that will need screenshots rather than a build gate.

Gate: `pnpm build`, `pnpm lint`, then a deployed walk-through of every page. There is no frontend
test runner, so the walk-through is the whole gate.

Done. Twelve packages moved and React stayed at 18.3.1 throughout, which the
peer dependencies allowed exactly as finding 5 predicted.

antd was the only one of the twelve that needed source changes, and tsc named
both. `duration: null` no longer disables auto-close on a notification: v6 uses
`false`, and five notifications here must stay up until the operator dismisses
them. Modal's semantic DOM dropped `styles.content` for `styles.container`.

Four packages are pinned below their newest release because pnpm's release-age
gate holds those back and taking them would have written a bypass into
pnpm-workspace.yaml. antd 6.6.2 was a day old, lucide-react 1.37.0 was eight
hours old.

antd 6 is smaller than antd 5 here: the desktop chunk falls from 1005kB to
991kB and LockOutlined from 191kB to 148kB.

**Build on the branch, then build again on the merge.** The branch was cut from
`main`, and `main` lacks 39 of the fork's frontend files. The branch compiled
clean; the merge into `fork/integration` then failed on
`hid-status/input-disconnected.tsx`, a fork-only file carrying the same
`duration: null`. A single build on the branch would have shipped that.

### Phase 6: the build-time packages and the CI actions (done, 2026-08-29)

`chore/build-tooling`. Nothing here reaches the device.

- `tailwindcss` 3 to 4, which is a configuration rewrite rather than a version bump.
- `typescript` 5 to 7, `eslint` 9 to 10, and the plugin set that follows eslint.
- `actions/checkout` and `actions/setup-node` to v7, `actions/upload-artifact` to v7,
  `actions/download-artifact` to v8.

Gate: `pnpm build` byte-compared against the previous `web/dist`, and a full dry run of the release
workflow. Tailwind 4 changes how classes are emitted, so a diff of the built CSS is the evidence
that matters.

Cost: two days, dominated by Tailwind.

Done, and one target moved. The phase says "nothing here reaches the device", and that was wrong
about one thing: Tailwind emits the stylesheet the device serves, so this phase ended with a deploy
like the others.

**TypeScript stops at 6, not 7.** typescript-eslint refuses version 7 outright, with
`Error: typescript-eslint does not support TS 7.0`, and `pnpm lint` then exits 2 having linted
nothing. Its peer range is `>=4.8.4 <6.1.0`. The compiler itself is ready: tsc 7.0.2, the native
build, type-checks this tree with no errors against the same tsconfig. Only the linter blocks it.
TypeScript 6 is worth taking on its own account, because it reports what 7 removes, and it named one
thing here: `baseUrl`, which is now gone from tsconfig.json.

**eslint 10 brings 47 findings that are not lint noise.** eslint-plugin-react-hooks 7 puts the React
Compiler rules in its recommended set: 20 `set-state-in-effect`, 16 `immutability`, 10 `refs` and
one `purity`. Every one is worth reading, and fixing them means changing how the desktop UI renders,
on a device where a broken UI is a KVM nobody can reach. They are warnings, with the reasoning in
`eslint.config.js` beside them. `rules-of-hooks` stays an error and reports nothing.

This is the phase's one deliberate debt, and it is real work rather than a formality.

**Tailwind 4 kept the page identical, and the evidence is a class-level diff.** Four utilities moved
one step down their scale and were renamed at nine call sites, so the emitted declarations match the
old build exactly: `shadow-sm` to `shadow-xs`, `rounded-sm` to `rounded-xs`, `backdrop-blur-sm` to
`backdrop-blur-xs`, and `outline-none` to `outline-hidden`. Preflight is left out by importing the
theme and utilities layers by name, because `tailwind.config.js` is gone and antd owns the reset.

Comparing the two builds found 27 utilities gone and 22 new, and all 27 are accounted for. Five of
them were never real: Tailwind 3's scanner read `if (!container)` out of JavaScript and emitted a
`.\!container` utility for it.

The palette is taken as it comes. Version 4 generates its colours in OKLCH, so sky-400 moves from
`#38bdf8` to `#00bcfe` and red-500 from `#ef4444` to `#fb2c36`. The greys do not move, and they are
most of this UI. Pinning the old table back would mean carrying it forever.

**The two-build rule found nothing this time, which is the answer it exists to give.** The branch was
cut from `main`, which lacks eleven of this tree's frontend files, so the class comparison was run
again on the merged tree against a Tailwind 3 build of `fork/integration`. Same 27 gone, same 22 new,
one for one.

**Verified on the device.** The desktop menu bar, the settings modal and the Appearance tab all
render correctly under antd 6 and Tailwind 4, which also closes the "pages behind the login wall"
gap that Phase 5 left open. The only console message is the react-router `HydrateFallback` warning
that Phase 2 introduced. `.rounded` computes to a 4px radius in the browser, the same as before.

The deploy taught one thing worth writing down. `S95nanokvm` copies `/kvmapp/server` to
`/tmp/server` at boot and runs the binary from there, and the web root follows the executable. A
frontend written only to `/kvmapp/server/web` changes nothing until a restart, while looking
entirely correct on disk. Both trees have to be updated, and swapping the `/tmp` one in place costs
no downtime.

Two things came out of it that were not planned. `eslint-plugin-react` was in devDependencies and
`eslint.config.js` never loaded it, so it linted nothing; it is removed. And `tsconfig.node.json`
sets `composite: true` with no `noEmit`, so every type-check writes two `.tsbuildinfo` caches and
emits `vite.config.js` into the working tree. Those are in `.gitignore` now.

### The follow-up, done the same day

Phase 6 left one thing open: the four GitHub actions were pinned to tags, not to commit shas,
and the account that owns an action can move a tag at any time. Phase 3 stopped the two
Dockerfiles trusting a movable name for exactly that reason, so the workflows were the last
build input a third party could still change under us. `chore/pin-workflow-actions` closes it.
All ten references now name a commit, each with the version beside it:

| action | commit | version | used |
| --- | --- | --- | --- |
| `actions/checkout` | `3d3c42e5` | v7.0.1 | 6 times |
| `actions/setup-node` | `82076278` | v7.0.0 | once |
| `actions/upload-artifact` | `043fb46d` | v7.0.1 | once |
| `actions/download-artifact` | `3e5f45b2` | v8.0.1 | once |

The shas are what the tags resolved to on 2026-08-29, so the commit changes which commit runs,
not which version runs.

`tools/build/test-pinned-actions.sh` holds them there. It reads the workflow files, runs no
build, and makes no network call. It fails if an external `uses:` names anything but a
40-character sha, if a pin loses its `# vX.Y.Z` comment, if one action is pinned to two
different commits, or if a local reusable workflow points at a file that is absent. The third
case is the one that motivated the suite: `actions/checkout` is used six times, so a bump that
reaches five of them leaves the sixth behind and both builds still pass. Each of the four cases
was shown to fail under a mutation, and each mutation was confirmed to have applied, before the
suite was committed.

The suite asks GitHub nothing. A test that compared a pin against the tag would fail on the day
an action published a release and would report a correct pin as a defect.

**The gap this opens, and it is real.** A sha never follows a release.
`dependabot_security_updates` is enabled on the repository, so an advisory against one of these
four actions still produces a pull request. But `.github/dependabot.yml` exists on neither
`main` nor `fork/integration`, so no ecosystem here is watched for routine releases: `npm`,
`gomod` and `github-actions` are all unwatched. The pins will now age in silence. Closing that
needs a `dependabot.yml` on `main`, because Dependabot reads its configuration from the default
branch, and it needs `target-branch: fork/integration` so the pull requests reach the branch
that carries the pins. That is standing configuration and it is not added here.

## What this plan deliberately leaves alone

**React 18 to 19.** Nothing needs it, as finding 5 shows. It is a decision about the frontend's
future and not a dependency chore, and folding it into a sweep would make every other change in
Phase 5 harder to judge.

**The device rootfs.** Kernel, BusyBox, OpenSSH and OpenSSL are all old and none is built here.
Moving them is a Buildroot project with two known traps, described in finding 7. If the OpenSSH
version is the concern, the cheaper answer is to close port 22 from outside the LAN, because the
supervisor's ssh door now keeps the daemon alive for local repair either way.

**`server/dl_lib/`.** These forty libraries follow MaixCDK. Phase 3 gives them a version; there is
nothing to bump directly.

**Opus and Tailscale.** Already current, for the reasons in finding 8.

## One correction to the record (made)

`AGENTS.md` says the backend is Go 1.24. `server/go.mod` says `go 1.25.0`, and both Dockerfiles
install Go 1.25.0. The document was the thing that was wrong, and it now says 1.25.
