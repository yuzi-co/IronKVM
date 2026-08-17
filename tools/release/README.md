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
| `sfdisk`, `mkfs.vfat`, `mtools`, `e2fsprogs`, `zstd`, `xz` | For the card image. |
| `u-boot-tools`, `cpio` | For the boot image, and the ability to `mknod`. |
| `gh`, authenticated | Creates the release. |
| A `base/` directory | See below. |

### Running it from a Windows workstation

A Windows workstation has almost none of that. `Dockerfile` beside this file is
the host as an image: build it once, then run the release inside it.

```shell
docker build -t ironkvm-release-host tools/release
```

From a WSL shell, in the checkout:

```shell
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$PWD:$PWD" -w "$PWD" \
  -e BUILD_UID=<the id the builder image was built with> \
  -e BUILD_GID=<the same for the group> \
  ironkvm-release-host tools/release/release.sh --dry-run 1.0.0
```

Three details make that work, and each one fails differently without it.

**The checkout is mounted at its own path.** `release.sh` starts a second
container and passes `$PWD` as the bind source. The daemon resolves that path
itself, so it has to mean the same thing inside the container and outside it.

**The Docker socket is the workstation's.** The container holds the client and
nothing else. The MaixCDK builder image never has to exist twice.

**`BUILD_UID` names the identity the builder image was built with**, which is
the Windows account's id, not the id a WSL shell or a container reports. The
builder bakes the ownership of `/home/build` in at build time and its entry
point drops to whatever id it is given, so a mismatch leaves `go` unable to
write its module cache. `docker images` shows the id in the image name.

The container runs as root on purpose: `repack-boot.sh` unpacks an initramfs
holding `dev/console`, and `mknod` is refused to anybody else.

A real release also tags, pushes and creates the GitHub release, so it needs the
credentials a dry run does not. Add `-e GH_TOKEN` and mount the key `git push`
uses:

```shell
  -e GH_TOKEN -v "$HOME/.ssh:/root/.ssh:ro"
```

The card is assembled inside the container and only the compressed image is
moved to `RELEASE_OUT`. `build-card.sh` creates the card at its full 28.85 GiB
and truncates it afterwards, and that hole is free on a filesystem with sparse
files. A Windows drive through a bind mount has none: an 8 GiB hole measured
8.0G allocated there against 0 on ext4. Assembling the card on the bind mount
therefore wrote about 25 GB for real and read it back twice to check the slots,
which is where the first dry run spent an hour. The build host's own filesystem
always has sparse files, so nothing about this needs configuring.

## The base

`base/` holds the pinned Sipeed inputs. It is gitignored: the files are Sipeed's
and they total about 280 MB.

```
base/rootfs.tar.zst        251 MB  the base root filesystem, from p2 of the official image
base/nanokvm_2.5.0.tar.gz   16 MB  the official application, layered under the fork's own
base/boot/                  12 MB  the official /boot, including boot.sd and fip.bin
base/version                        the official application version these came from
```

`base/version` becomes `/kvmapp/base-version` on the device, and the About panel
reads it to show `IronKVM 1.0.0 (based on NanoKVM 2.5.0)`.

The release script verifies `rootfs.tar.zst` and the application against
`tools/abslots/BASE.sha256` and refuses to build from anything unpinned. That
file is the only record of which bytes an image was built from: the official
system image ships no checksum of its own, so its hash is a pin rather than a
verification.

### Rebuilding base/

`tools/release/fetch-base.sh` does the whole thing: it downloads both artifacts
from Sipeed, checks each against the pin, reads the image's own partition table,
and extracts `/boot` and the root filesystem. It needs Linux and the ability to
loop-mount. It downloads about 1.6 GB and leaves about 280 MB.

```shell
tools/release/fetch-base.sh
```

This fork does not republish the base. The GPL parts could be redistributed, but
the image also carries vendor binaries for the SG2002 whose terms are not stated,
and fetching from the publisher gets the same reproducibility without answering
that question on somebody else's behalf.

By hand, the same steps are these. The rootfs comes from partition 2 of the
official card image and the boot directory from partition 1.

```shell
# p1, the boot partition: 16 MiB at sector 1
mount -o loop,offset=512,ro 20260610_NanoKVM_Rev1_4_3.img /mnt/p1
cp -a /mnt/p1/. base/boot/
umount /mnt/p1

# p2, the root filesystem. Its offset comes from the image's own table.
mount -o loop,offset=$((<p2_start> * 512)),ro 20260610_NanoKVM_Rev1_4_3.img /mnt/p2
( cd /mnt/p2 && tar --numeric-owner -cf - . ) | zstd -q -o base/rootfs.tar.zst
umount /mnt/p2

sha256sum base/rootfs.tar.zst base/nanokvm_*.tar.gz   # must appear in BASE.sha256
```

Do not build the base from a running board. `tools/abslots/BASE.sha256` records
why: the slot archived on 2026-08-15 is older than v1.4.3, and building from a
board is how a factory `/etc/kvm/ssh_stop` and a CRLF init script reached a slot.

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
| `SHA256SUMS` | Release assets | For a person checking a download is intact. Unsigned, see below. |

The feed and the packages live apart on purpose. GitHub Pages is right for a few
hundred bytes of JSON and GitHub Releases is right for a 26 MB tarball, and a
release asset lives under a per-tag path that no fixed base URL can reach. The
manifest therefore names the package with an absolute `url`.

### The checksums are not signed

`SHA256SUMS` proves a download is intact. It proves nothing about who produced
it: anybody who can replace the artifacts can replace this file beside them.

A signature was designed in and then dropped for 1.0. It is worth exactly what
the key's safekeeping is worth. An offline key would let somebody who pinned it
detect a later compromise of this repository, which is real value; a key sitting
on the build machine with no backup adds ceremony and no protection, and losing
it forces every user to re-trust from scratch. Saying the checksums are unsigned
is more honest than a signature nobody can rely on.

The device never reads this file. Its update path checks a sha512 from
`latest.json` over TLS, so signing here would not have protected it either way.
Signature verification on the device is a different and larger piece of work.

## The feed branch

Done on 2026-08-16. `gh-pages` exists, GitHub Pages serves it at
`https://yuzi-co.github.io/IronKVM/`, and HTTPS is enforced. The branch holds an
`index.html` that says what the URL is for, and a `.nojekyll` so Pages does not
run a Jekyll build over a directory containing one JSON file.

It was created as an orphan with plumbing rather than `git switch --orphan`,
which empties the working tree of every tracked file and puts them back on the
way out. On a checkout this size that is a lot of churn for a branch holding two
files.

`latest.json` returns 404 until the first release writes it. That is correct: a
device asked to check for updates before then reports that the update server is
inaccessible. Check the URL answers before announcing anything.

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

### 2026-08-17, installing 1.0.0 on a board: PASS, with one open question

`ironkvm_1.0.0.tar.gz` uploaded through `Settings > Update` to a board on slot A.
Every check passed.

| Check | Result |
| --- | --- |
| `/kvmapp/version` | `1.0.0` |
| `/kvmapp/base-version` | `2.5.0` |
| `server/dl_lib` | 38 libraries |
| `kvm_system`, `system/tool` | present, 2 files |
| `libkvm.so` mapped by the running server | yes, 0 relocation errors |
| Boot scripts changed | `S95nanokvm` replaced, `S99vidiag` added |
| avahi, ssdpd, tailscaled, picoclaw, wifi, usbhid | none installed |
| Backup file left in `/etc/init.d` | none |

The two fixes that mattered are both proved here. The 38 libraries and the clean
loader are the layering fix: a package built from `kvmapp/` alone would have
left 14 and a server that cannot start. The two-line backup manifest is the
install list: without it, ten more scripts would have been written and six
daemons would have started at the next boot.

**The open question is the reboot.** The board rebooted during the update, and
nothing in the logs says which component asked for it. The supervisor did not:
it reboots at five short runs and recorded one. The watchdog stood down. The
shutdown was orderly, so something called `reboot` deliberately.

What the supervisor log does show is a race:

```
18:51:09  NanoKVM-Server is gone after 5s (short_runs=1), restarting in 10s
18:51:19  started /tmp/server/NanoKVM-Server as pid 2322
18:51:32  NanoKVM-Server is gone after 3s (short_runs=1), restarting in 20s
18:51:43  supervisor stopped
```

The supervisor started the server while the updater was moving `/kvmapp`. The
existing sentinel file could not have prevented it: that one is a download lock,
removed when the upload handler returns, which is before `restartServices` runs,
and `S98supervise` never read it.

**Fixed on 2026-08-17.** The updater writes `/tmp/nanokvm-updating` before it
touches `/kvmapp`, and `S98supervise` returns a new `updating` verdict while
that marker is fresh, which suppresses every other decision. The updater does
not remove the marker, because the restart is inside the window it protects: the
process that would clean up is the one being replaced. The next server to start
clears it, and a failed install clears it too, since a caller that gets an error
never restarts anything.

Two bounds stop a marker nothing clears from suspending the supervisor for good.
It lives in tmpfs, so a reboot clears it. And it is ignored once it is older than
`SUPERVISE_UPDATE_STANDOFF`, 300 seconds by default, so an update that dies
halfway costs one window rather than the board. An unreadable timestamp resumes
supervision rather than extending the stand-off: a dead server nothing restarts
is only cleared by a reboot, and that is the failure that stays silent.

**The reboot itself is still unexplained.** Removing the race removes the most
likely trigger, since the server the supervisor started died with its files
being moved, but nothing in the logs named the component that called `reboot`,
and this is not evidence that it will not happen again. The next update on real
hardware is the test.

### 2026-08-17, first end-to-end dry run: PASS, after four faults

`tools/release/release.sh --dry-run 1.0.0`, in the image above, against
`fork/integration` at `30a674e9`. It produced a 12,484,962 byte package, a
372,265,848 byte compressed card image, `latest.json` and `SHA256SUMS`. Both
slots pass `e2fsck` when read back out of the compressed artifact, the table
carries all six partitions, and `p1` holds `fip.bin` and the repacked `boot.sd`.

It took four attempts. The faults are what the run was for.

**The web build stopped on a question.** pnpm asks before it purges a modules
directory another platform installed and refuses to purge without a TTY. It was
the first step, so the release stopped having built nothing.

**The builder identity was the host's.** `release.sh` passed `id -u` to an image
that bakes the ownership of `/home/build` in at build time. Any host but the one
that built the image ran `go` as a user that cannot write its module cache.

**The package left out the identity script.** The image manifest installs five
boot scripts from `tools/`, and the package carried three. `S02identity` is the
one that stops a slot switch reverting the root password to the factory one.
`rcS` is the other, and it is now held back on purpose: it is what runs the
watchdog, so a valid but wrong `rcS` would leave nothing to repair the board.

**The package was missing 27 files, and this is the one that mattered.** It was
assembled from `kvmapp/` alone, which holds only what the fork changes: 14 of
the 37 libraries in `server/dl_lib`, no `kvm_system`, no `system/tool`. The
updater replaces `/kvmapp` rather than merging into it, so installing 1.0.0
would have moved away `libopencv_core`, `libcvi_audio` and twenty others and
left a server that cannot load `libkvm.so`.

The board would have stayed reachable through it. ssh answers and the address is
up, so the watchdog calls it healthy and rolls nothing back. A KVM that is online
and cannot show a screen is the failure this project exists to prevent, and it
was one command away from shipping.

The package is now built from the same two layers as the image, and a check
after the copy refuses any package missing a file the official one carries.

### 2026-08-16, boot-script rollback: PASS

Run with `tools/release/acceptance-initd-rollback.sh root@10.0.0.222`, against a
board on slot A running `dev.20260816.1824.2c702a6f`.

`S30eth` was replaced with a valid script that configures nothing, so the board
lost its only way in. It came back on its own after **367 seconds**, with the
original `S30eth` restored, the manifest spent, and no recovery marker set.

```
302s  no way in after 300s: carrier=up address=down web=up ssh=up
302s  no way in after an update: restored the previous /etc/init.d, rebooting
 12s  reachable after 10s (web=down, ssh=up)
```

Three things worth keeping from it.

**`web=up ssh=up` on an unreachable board is not a contradiction.** Those probes
hit `127.0.0.1`, which answers with no address on `eth0`. `address_up` is the
signal that decides reachability, and `healthy` checks it first. A future change
that reorders those probes would make the watchdog blind.

**Breaking `S30eth` is the right break, and `S01fs` is not.** A board with a
broken `S01fs` usually still answers ssh, so the watchdog would call it healthy
and never roll anything back. The test has to cost the board both doors.

**The counter was not what fired.** The undo came from `escalate`, at the
deadline. The counter in `start()` covers a power cut before that deadline and a
board with no recovery slot; on a board with this watchdog it is the second
trigger, not the first.

The run also turned up an unrelated fault: `/etc/init.d/S02identity.rollback`
was present and executable, so `rcS` ran an older identity script after the
current one at every boot, binding `/etc/kvm` a second time. It was moved to
`/root/initd-backups/`, and `build-image.sh` now refuses an image that carries a
backup file in `/etc/init.d`.
