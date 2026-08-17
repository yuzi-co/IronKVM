# IronKVM

**Hardened community firmware for the Sipeed NanoKVM. Not affiliated with Sipeed.**

IronKVM is a fork of [sipeed/NanoKVM](https://github.com/sipeed/NanoKVM) for
people who use the board as the only way into a machine. It adds A/B root slots
with a recovery slot, a boot watchdog, identity that survives a slot switch, USB
audio, and a large set of security and reliability fixes.

**No support is promised.** This is published because it may be useful, not
because anybody is on call for it. Issues are welcome and may go unanswered.

Read [what it changes](docs/CHANGES-FROM-OFFICIAL.md) before installing.

## Hardware

NanoKVM Cube (Lite and Full) and NanoKVM PCIe. Anything built on the SG2002.

NanoKVM-Pro is a different chip and a different repository. IronKVM does not run
on it.

The card image needs a microSD card of at least 32 GB. Space beyond that is
unused for now.

## Install

Installing replaces everything on the card, including the data partition.

1. Download `ironkvm-<version>-sdcard.img.xz` from
   [Releases](https://github.com/yuzi-co/IronKVM/releases).
2. Check it against `SHA256SUMS`. That proves the download is intact. It is not
   signed, so it does not prove who built it: anybody who could replace the
   image could replace the checksums beside it.
3. Write it to the card with your usual imaging tool.
4. Put the card in the board and power it on. The first boot makes the data
   partition and takes longer than later ones.
5. Open the board in a browser. The default login is `admin` / `admin`.

**Change the password before you do anything else.** Doing it in the web UI sets
the web password and the root password together, and writes both somewhere a
slot switch cannot lose them.

## Update

`Settings > Update` checks IronKVM's feed and installs from it.

An update replaces the application and, if the package carries them, the boot
scripts. If the board then fails to come up three times, it puts the previous
boot scripts back by itself. It is not yet a slot trial: see
[Not here yet](docs/CHANGES-FROM-OFFICIAL.md#not-here-yet).

You can also download the `.tar.gz` from a release and upload it in the same
page, which needs no network access from the board.

Both paths need a board that already runs IronKVM. You cannot upgrade to IronKVM
from the official firmware: write the card image instead. See
[Not here yet](docs/CHANGES-FROM-OFFICIAL.md#not-here-yet) for why.

## Going back to the official firmware

This is a supported path, not an apology.

1. Open `Settings > Update`.
2. Turn on the custom update server and press **Use the official Sipeed server**.
3. Save, then update.

Every path on disk is unchanged from upstream, so an official package installs
over IronKVM the same way it installs over itself. If you prefer to start clean,
flash Sipeed's card image instead.

The boot-level changes live in the card image rather than in the package, so an
application-level rollback leaves the slots and the watchdog in place. Flashing
the official image removes them.

## Build from source

The device build is a RISC-V cross-compile against libraries that only exist on
the board, so it does not build natively on a workstation.

```shell
make app        # cross-compile the server in the Docker builder
make support    # build kvm_system
make vision     # build the capture libraries
```

For anything that does not need the device libraries, the `novision` tag stubs
them out:

```shell
cd server && go test -tags novision ./...
```

See [CLAUDE.md](CLAUDE.md) for the rest, including the traps that cost real time.

## Release

`tools/release/release.sh` builds and publishes a release. See
[tools/release/README.md](tools/release/README.md) for what the host needs.

## Licence and credit

GPL-3.0, the same as upstream.

The hardware, the base system image and the great majority of this code are
Sipeed's work. IronKVM is a fork, and it is not endorsed by or affiliated with
Sipeed. Hardware questions belong to
[their wiki](https://wiki.sipeed.com/nanokvm), not here.
