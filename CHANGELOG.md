# Changelog

## 1.0.0 (unreleased)

The first IronKVM release. It is a fork of the official NanoKVM firmware, and
[docs/CHANGES-FROM-OFFICIAL.md](docs/CHANGES-FROM-OFFICIAL.md) lists what it
changes and why.

### Added

- A/B root slots with a recovery slot, and a boot watchdog that reverts a failed
  trial or falls back to recovery.
- Identity that survives a slot switch: the root password, the web password, the
  ssh host keys and `/root/.ssh`. A password change writes itself back.
- USB audio. The board carries the host's audio to the browser over a UAC1
  gadget, because HDMI on this hardware cannot.
- A serial console for the managed host, off by default.
- An update feed of its own, with the official Sipeed server kept as a one-click
  preset.
- A boot-script rollback: an update that stops the board booting is undone after
  three failed boots.
- zram, an HTTP proxy setting, API keys, and a build stamp that identifies the
  running binary.
- `tools/release/release.sh`, which builds, verifies and publishes a release, and
  `tools/abslots/build-card.sh`, which assembles a flashable card image.

### Changed

- The board no longer answers DHCP on the wired network.
- The HDMI receiver is no longer power-cycled at startup.
- Runtime state that used to be written to the boot card continuously now lives
  in memory.
- The About panel shows `IronKVM <version> (based on NanoKVM <base>)`, carries
  the disclaimer, and no longer links Sipeed's support channels. The hardware
  links remain.

### Fixed

- Cross-site request forgery, path traversal and shell injection in the API.
- A password change that did not require the current password.
- A JWT check that accepted the algorithm the token declared.
- A secret key that fell back to a guessable default.
- Updates that were installed without verifying what was downloaded.
- HID that stopped working until a reboot after a USB gadget rebuild.
- A stalled viewer that froze the stream for every other viewer.
- A capture pipeline that reported success after a failed initialisation, and
  that was dismantled while its own threads were still using it.
- A resolution probe that spun forever on an unreadable HDMI input.
- A server that was not restarted when it died.

### Known limits

- **The OLED screen and the Wi-Fi access point still say NanoKVM.** Both are C++
  strings inside `kvm_system`, and changing them would put the whole MaixCDK
  toolchain in the release path and break the documented Wi-Fi provisioning flow.
- **An update installs in place on the running slot.** It is protected by the
  boot-script rollback rather than by a slot trial. Slot-aware updates are the
  next piece of work.
- **The card image needs a card of at least 32 GB and ignores anything beyond
  that.** The partition table is fixed and the data partition is not grown on
  first boot yet.
- **The hardware acceptance test for the boot-script rollback has not been run.**
  Until it has, that mechanism is tested only against a scratch tree.
- **Only the English strings are rebranded.** The other 24 locale files keep
  their current text, because a part-retranslated file is worse than an
  untouched one.
