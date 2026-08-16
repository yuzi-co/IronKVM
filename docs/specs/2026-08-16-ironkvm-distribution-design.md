# IronKVM Distribution Design

**Date:** 2026-08-16
**Status:** approved design, ready for an implementation plan
**Scope:** release 1.0 of IronKVM. Slot-aware updates are deferred to a separate design.

## Goal

Turn this fork from a private branch stack into a named, versioned, installable
firmware distribution for the Sipeed NanoKVM, with published releases, an update
feed, and a written account of how it differs from the official firmware.

## What the fork is today

`fork/integration` carries 84 commits on `main` above `upstream/main`, and 159
more above `main`. Measured against `upstream/main` that is 296 files changed
and about 37,000 inserted lines. The fork adds an A/B/recovery slot layout, an
identity carry-over mechanism, a boot watchdog, zram, USB audio, HTTP proxy
support, API key authentication, and a large set of reliability and performance
fixes.

The fork already contains half of an update system. `service/application/`
fetches `latest.json` from a configurable base URL, checks a sha512, applies
size limits, and honours a proxy. `Settings > Update` exposes the URL. What is
missing is a server to point it at, artifacts to serve, and a name.

## Position and naming

The distribution is called **IronKVM**.

The name follows the convention of the product category (PiKVM, JetKVM,
RustKVM). "Iron" is a claim about the firmware, not a claim about who made the
board. Searches found no GitHub user, organisation or repository on `ironkvm`,
`iron-kvm` or `IronKVM`, and no registered trademark surfaced for `IRON KVM` or
`IRONKVM`. The USPTO database needs a session that a plain fetch cannot open, so
that check is at search-engine level only. Run the official search before the
name goes on physical goods.

The domains `ironkvm.com`, `.org`, `.dev`, `.io` and `.net` had no NS records
when checked on 2026-08-16.

Every user-facing surface carries this line:

> IronKVM: hardened community firmware for the Sipeed NanoKVM. Not affiliated
> with Sipeed.

The upstream project stays credited, and the licence stays GPL-3.0.

The repository is renamed to `IronKVM` under the `yuzi-co` account. GitHub keeps
a redirect from the old name, so existing clones continue to work. The remote
named `upstream` continues to point at `sipeed/NanoKVM`.

### The rename is deliberately shallow

On-disk layout does not change. `/kvmapp`, `/etc/kvm`, `NanoKVM-Server`, the
API paths and the compiled-in AES key all stay as they are.

The reason is the escape hatch. A user who does not want IronKVM must be able to
install the official firmware over it, and an official package must remain
installable through the fork's own updater. A deep rename removes that, breaks
every device-side tool in `tools/`, and invalidates the image manifests. It buys
a tidier directory listing and costs the exit.

## Scope

### In scope for 1.0

1. The branding surface listed below.
2. A version line of its own, semver from 1.0.0.
3. Release artifacts and the script that builds and publishes them.
4. The update feed, and the two small client changes it needs.
5. The package install hook, and the boot-success rollback that makes it safe.
6. The documentation set, including the written diff from the official firmware.

### Out of scope, deferred to its own design

**Slot-aware updates.** The correct end state is that an update writes the new
root filesystem to the inactive slot, marks a trial, reboots, and lets the
watchdog confirm it or roll it back in one boot. All of that machinery exists.
Wiring the updater into it is a project of its own, and release 1.0 must not
wait for it.

Nothing built for 1.0 is wasted by that deferral. The boot-success counter
specified below is the signal a slot trial needs, and boards on stock
partitioning keep the protection permanently.

**A preview channel.** `PreviewURL` exists in the code. One channel with nothing
in it is a promise that must then be kept. Add it when there is something to
preview.

**OLED and Wi-Fi access point strings.** See "Deferred branding" below.

## Versioning

IronKVM uses semver starting at 1.0.0. The version lives in `/kvmapp/version`,
where the updater already writes it, and `common/version.Decorate` continues to
attach the build stamp as build metadata.

The fork cannot encode the upstream number in its own version. `2.4.3+iron.5`
compares equal to `2.4.3` because semver ignores build metadata, and
`2.4.3-iron.5` sorts below it. Both break update detection, because
`Settings > Update` orders versions with `semver.gte`.

The upstream base is therefore recorded beside the version, not inside it. The
About panel reads:

> IronKVM 1.0.0 (based on NanoKVM <base>)

The base value comes from the version file of the Sipeed image the release was
built from. It is read at build time and written into the release, not
hardcoded in source.

Starting at 1.0.0 keeps the escape hatch working through the user interface. A
device on IronKVM 1.x that points its feed back at Sipeed sees the official
2.4.x release as an upgrade and can install it with one click.

## Release artifacts

A tag `v1.0.0` produces four things.

| Artifact | Published to | Purpose |
| --- | --- | --- |
| `ironkvm-1.0.0-sdcard.img.xz` | Release assets | First install. Full A/B/recovery card image. |
| `ironkvm_1.0.0.tar.gz` | Release assets | In-user-interface update, and offline upload. |
| `latest.json` | GitHub Pages | The feed the device polls. |
| `SHA256SUMS` | Release assets | For a human checking a download is intact. Unsigned; see the amendment below. |

The tarball's top-level directory is `ironkvm_1.0.0/`, because `update.go`
derives the expected root from the package name.

**Amended 2026-08-16, during implementation: `SHA256SUMS` is not signed.**

The design called for `minisign`, on the reasoning that it is one binary and a
one-line public key. Implementation made the trade-off concrete and it did not
hold up for 1.0. A signature is worth what the key's safekeeping is worth: an
offline key lets somebody who pinned it detect a later compromise of the
repository, but a key on the build machine with no backup adds ceremony and no
protection, and losing it forces every user to re-trust from scratch.

The device was never going to read it either way. Its update path checks a
sha512 from the manifest, fetched over TLS from the same origin as the package,
so signing the release artifacts protects a human at a terminal and nothing
else. The README states that the checksums are unsigned rather than implying
otherwise.

Signature verification on the device stays out of scope, and is the piece that
would actually change the threat model.

## Build and publish

Release 1.0 is built by `tools/release/release.sh`, run on a Linux host, and
published with the `gh` command line tool.

The reason is a constraint, not a preference. The SD image build needs two
inputs that a hosted runner does not have: the Sipeed base image, and the
MaixCDK builder image that exists only as a locally built
`nanokvm-builder-local-<uid>-<gid>`. A workflow file that cannot run is worse
than no workflow file.

A GitHub Actions workflow can be added later. The tarball half of the release
needs only Go and pnpm, so it can move to CI first.

The release script does these steps:

1. Refuse to run on a dirty tree, or on a tag that already exists.
2. Build the web user interface with `pnpm build`.
3. Cross-compile the server, then patch its RPATH to `$ORIGIN/dl_lib`.
4. Assemble the payload, then produce `ironkvm_<version>.tar.gz`.
5. Build the two slot filesystems with `tools/abslots/build-image.sh`, which runs
   its own gates, then assemble them into a card with
   `tools/abslots/build-card.sh`. The second script did not exist when this
   design was written: `build-image.sh` produces one root filesystem, and the
   slots in use had been written to a card partitioned by hand.
6. Compute checksums, then write `latest.json`.
7. Verify its own output: the tarball top directory matches the package name,
   and the sha512 in `latest.json` matches the artifact it names.
8. Create the GitHub release, upload the assets, then push `latest.json` to the
   Pages branch.

Step 7 exists because this repository has already lost three guards to rot. A
release script that cannot check its own output will produce a feed that points
at a file that does not match it.

The script has a dry-run mode that performs every step except the two that
publish.

## The update feed

`latest.json` gains one optional field, `url`:

```json
{
  "manifest_version": 2,
  "version": "1.0.0",
  "name": "ironkvm_1.0.0.tar.gz",
  "url": "https://github.com/yuzi-co/IronKVM/releases/download/v1.0.0/ironkvm_1.0.0.tar.gz",
  "sha512": "...",
  "size_bytes": 26214400,
  "unpacked_size_bytes": 41943040
}
```

The client rule: if `url` is present, it must be absolute and it must use
`https`, or the update is refused. If `url` is absent, the client falls back to
today's behaviour and joins the base URL with the package name. The proxy rules
and the redirect rules do not change.

This grants no new trust. The sha512 already comes from the same manifest, so a
hostile feed can already serve any bytes it likes. The field only removes the
requirement that the manifest and the package sit in one directory.

That requirement is what makes the field necessary. GitHub Pages is free, has no
uptime obligation, and is the right place for a few hundred bytes of JSON.
GitHub Releases is the right place for a 26 MB tarball and a several hundred MB
image. Release assets live under a per-tag path, so a fixed base URL cannot
reach them.

The feed is served from `https://yuzi-co.github.io/IronKVM/latest.json`, and can
later move to `feed.ironkvm.dev` without a client change.

Images ship with `/etc/kvm/application-update.json` already set to that feed and
enabled. The user interface keeps the URL field, and gains a preset that
restores the official Sipeed CDN in one click.

### Package name pattern

`packageNamePattern` currently accepts only `nanokvm_X.Y.Z.tar.gz`. It is
widened to accept `(nanokvm|ironkvm)_X.Y.Z.tar.gz`.

IronKVM then ships packages under its own name, and an official Sipeed package
stays installable through the fork's updater. This is an addition, not a rename,
so it is consistent with the shallow-rename decision.

## Branding surface

| Location | Now | Becomes |
| --- | --- | --- |
| `web/index.html` | `<title>NanoKVM</title>`, `/sipeed.ico` | `IronKVM`, `/ironkvm.svg` |
| `web/src/components/head.tsx` | document title | IronKVM |
| `web/src/pages/auth/login/index.tsx` | `/sipeed.ico` | `/ironkvm.svg` |
| `settings/about/community.tsx` | Sipeed wiki, repository, FAQ | IronKVM documentation and repository, and a "Hardware (Sipeed)" link that stays on purpose |
| About panel | version only | `IronKVM 1.0.0 (based on NanoKVM <base>)` and the disclaimer |
| `server/service/vm/web_title.go:26` | `"NanoKVM"` is the default sentinel | `"IronKVM"` |
| `server/service/application/service.go` | `StableURL` is the Sipeed CDN | the IronKVM feed |
| `web/.../update/custom-server.tsx` | `OFFICIAL_UPDATE_SERVER` | the IronKVM feed, with Sipeed kept as a preset |
| `README.md` and variants | NanoKVM fork | IronKVM, with the disclaimer and upstream credit |

### The rule for translated strings

`en.ts` alone contains 27 occurrences of "NanoKVM", and there are 25 locale
files.

Do not replace them mechanically. A string that names the **product** changes. A
string that names the **hardware** does not, because it is still correct: the
board is a NanoKVM whatever firmware it runs.

Each of the 27 English strings is classified by hand. Only `en.ts` receives new
English text. The other 24 locales keep their existing strings until a
translator revises them. A partly retranslated file is worse than an untouched
one, because the reader cannot tell which half is current.

### Deferred branding

Two strings stay as they are in 1.0:

- The OLED shows `NanoKVM` (`support/sg2002/kvm_system/main/lib/oled_ui/oled_ui.cpp:413`).
- The Wi-Fi access point SSID is `NanoKVM`
  (`support/sg2002/kvm_system/main/lib/system_ctrl/system_ctrl.cpp:13` and `:119`).

Both live in C++ inside `kvm_system`. Changing them means rebuilding through
MaixCDK and shipping a new `kvm_system` binary in every release, which drags the
whole toolchain into the release path for two strings. Changing the SSID also
breaks the Wi-Fi provisioning flow for anyone following Sipeed's documentation.

The documentation states that these two strings still read NanoKVM, and why.

## The install hook

A package update must be able to install boot scripts. Nothing in `kvmapp`
copies `kvmapp/system/init.d/*` into `/etc/init.d` today. Only the image
manifest does. Without a hook, a release would ship the fork's boot behaviour in
the image and silently omit it from the package, and the two would diverge with
nothing to show it.

### `kvmapp/system/install.sh`

The package carries it. The updater runs it after the package is unpacked and
before the service restarts.

- Syntax-check every script with `sh -n` first. If any script fails, install
  nothing and exit non-zero.
- Copy in only the scripts that differ from what is installed.
- Save each replaced original to a backup directory, with a manifest that
  records whether the original existed at all. A script that is new has no
  original, and restoring must delete it rather than restore nothing.
- Never delete a script that this mechanism did not install.
- Be idempotent. A second run changes nothing.
- Install `S00awatchdog` last, because it is the script that performs the
  repair.

### The server hook

`installPreparedPackage` gains a step that runs `install.sh` if it is present
and executable, with a timeout. The call goes through a package-level variable
so tests can stub it, in the same pattern as `saveIdentity` in
`service/auth/password.go`.

A failure is logged and surfaced in the update result. It does not undo the
application install: the new application is already in place, and the previous
one is in `BackupDir`.

## Boot-success rollback

A package update lands in place on the running slot. There is no trial boot and
no slot switch, so a bad boot script would not be caught by the slot machinery.
This is the one genuinely dangerous part of the design, and it needs its own
guard.

- A counter is incremented early in boot, and cleared when the server answers.
- The success signal is the one the watchdog already has. `S00awatchdog`
  probes for reachability and already reports "confirmed" and "standing down".
  That same event clears the counter.
- On the third consecutive boot with no success, `S00awatchdog` restores the
  backed-up init scripts, clears the counter, logs the restore loudly, and
  continues booting.
- Because `S00awatchdog` runs first, it repairs `S01` through `S95` before they
  run. Recovery costs one boot, not two.

### Two limits, stated rather than hidden

1. If the broken script is `S00awatchdog` itself, it cannot repair itself. The
   mitigation is partial: `install.sh` writes that file last and syntax-checks
   it. Beyond that the recovery slot is the backstop, which is what the recovery
   slot is for.

2. Where the counter lives must be verified on hardware, not assumed. `/boot` is
   attractive because `/boot/slot.try` sets the precedent and `/boot` mounts
   before `/data`. Whether it is mounted when `S00awatchdog` runs is a question
   for the board, not for this document. The implementation checks it and picks
   the location that is actually available at that point in boot. Runtime state
   does not go under `/kvmapp`.

## Documentation

- `README.md`, rewritten. What IronKVM is, the disclaimer, install, update, and
  how to return to the official firmware. The last of these is a feature and is
  documented as one.
- `docs/CHANGES-FROM-OFFICIAL.md`. The real difference from the official
  firmware, grouped by security, reliability, performance and features, written
  by hand from the 254 commits. A generated commit list is not a document that
  anybody reads.
- `CHANGELOG.md`, one section per release, generated from tags and then edited.
- `README_ZH.md` and `README_JA.md` receive the identity header and the
  disclaimer only. The rest stays as it is, for the same reason the locale files
  do.

English source text follows ASD-STE100 where it fits, as the repository already
requires.

## Testing

- `install.sh` gets a shell test under `tools/`, in the pattern of
  `tools/abslots/device/test-identity.sh`, and every guard in it is
  mutation-tested.
- The boot-success rollback gets a hardware acceptance test: install a package
  that contains a deliberately broken `S01fs`, then prove the board restores
  itself and comes back. It shares the harness with the slot acceptance test.
- Go tests cover the `url` field validation, including rejection of a non-https
  and of a relative URL; the widened package name pattern, including that an
  official `nanokvm_` package still passes and that a name with a path separator
  still fails; and the install hook call, with a stubbed runner, including that a
  hook failure does not undo the application install.
- The release script is tested by its dry-run mode, and by the self-verification
  in step 7.

Tests are written before the code they cover, and each is watched to fail first.

## Risks

| Risk | Response |
| --- | --- |
| A package update installs a bad boot script and the board does not return. | The boot-success rollback, plus the recovery slot for the case it cannot cover. |
| The feed and the release assets fall out of step. | The release script verifies its own output before it publishes, and refuses to publish a manifest that does not match the artifact. |
| A user believes IronKVM is an official Sipeed build. | The disclaimer on the README, the login page and the About panel. The name does not contain "Nano" or "Sipeed". |
| The name turns out to be claimed. | Checked at search-engine level for GitHub, the web and domains. The registered-trademark check is still open and is stated as open. |
| The fork drifts from upstream and rebases become expensive. | Unchanged by this work. The existing branch discipline in `CLAUDE.md` continues to apply. |
| Releasing publicly attracts support requests. | The README states that no support is promised. |
