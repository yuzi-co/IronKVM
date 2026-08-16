# Releasing IronKVM

`release.sh` builds, verifies and publishes one release. It runs on a Linux host,
not in CI.

## Why it is not a workflow

The card image needs two inputs a hosted runner does not have: Sipeed's base
image, and the MaixCDK builder image, which exists only as a locally built
`nanokvm-builder-local-<uid>-<gid>`. A workflow file that cannot run is worse
than no workflow file. The package half needs only Go and pnpm, so that part can
move to CI later.

## What the host needs

| Requirement | Note |
| --- | --- |
| Docker | The server cross-compile and the image build both use it. |
| The MaixCDK builder image | Build it once with `make shell`. |
| `pnpm` | For the web user interface. |
| `sfdisk`, `mkfs.vfat`, `mtools`, `e2fsprogs`, `xz` | For the card image. |
| `minisign` with the release key | Signs `SHA256SUMS`. |
| `gh`, authenticated | Creates the release. |
| A `base/` directory | See below. |

The `base/` directory holds the pinned Sipeed inputs. It is not in the
repository, because it is Sipeed's build and it is large.

```
base/rootfs.tar.zst   the base root filesystem
base/boot/            the contents of the base /boot partition
base/boot.sd          the stock boot.sd, repacked by the release
base/version          the official version these came from
```

`base/version` becomes `/kvmapp/base-version` on the device, and the About panel
reads it to show `IronKVM 1.0.0 (based on NanoKVM <base>)`.

## Running it

```shell
tools/release/release.sh --dry-run 1.0.0    # build and verify, publish nothing
tools/release/release.sh 1.0.0              # build, verify and publish
tools/release/release.sh --verify-only 1.0.0
```

The dry run performs every step except tagging, the GitHub release and the feed
push. Use it first.

## What it produces

| Artifact | Goes to | For |
| --- | --- | --- |
| `ironkvm-<v>-sdcard.img.xz` | Release assets | First install. Flash the card. |
| `ironkvm_<v>.tar.gz` | Release assets | In-user-interface update, and offline upload. |
| `latest.json` | The `gh-pages` branch | The feed a device polls. |
| `SHA256SUMS`, `SHA256SUMS.minisig` | Release assets | For a person checking a download. |

The feed and the packages live apart on purpose. GitHub Pages is right for a few
hundred bytes of JSON and GitHub Releases is right for a 26 MB tarball, and a
release asset lives under a per-tag path that no fixed base URL can reach. The
manifest therefore names the package with an absolute `url`.

## The first release only

Create the feed branch once:

```shell
git switch --orphan gh-pages
git commit --allow-empty -m "Start the IronKVM feed"
git push -u origin gh-pages
git switch fork/integration
```

Then turn on GitHub Pages for that branch, and check that
`https://yuzi-co.github.io/IronKVM/latest.json` answers before announcing
anything. A device whose feed 404s reports that the update server is
inaccessible, which is correct but unhelpful.

## What the script refuses

- A dirty tree, and a tag that already exists.
- A version that is not `X.Y.Z`. A prerelease sorts below the release it came
  from, and build metadata compares equal to it, so either would break update
  detection on every device.
- Its own output, when the manifest and the package disagree. It compares the
  name, the sha512, the size and the tarball's top directory before it publishes
  anything.

That last check is not ceremony. A feed pointing at a package it does not
describe fails nowhere until a device tries to install it, and three guards in
this repository have already rotted into passing while testing nothing.

## Acceptance record

Nothing recorded yet. The hardware acceptance test for the boot-script rollback
has not been run. Add its outcome, its timings, and anything surprising here when
it is. A test that ran once and was never written down is a test nobody can trust
later.
