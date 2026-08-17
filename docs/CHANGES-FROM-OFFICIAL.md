# What IronKVM changes

IronKVM is a fork of [sipeed/NanoKVM](https://github.com/sipeed/NanoKVM). This
page says what it changes and why. It is written by hand from 254 commits,
because a generated commit list is not a document that anybody reads.

Nothing here is a criticism of the vendor. Most of it comes from running the
board as the only way into a machine, which is a narrower job than the firmware
is built for, and which makes some faults matter more than they otherwise would.

## Security

- **The API refuses cross-site requests, path traversal and shell injection.**
  Several endpoints accepted a request that another site could forge, and a
  handful built shell commands from request parameters. A stolen session or a
  crafted page was enough to take the device.
- **A password change now requires the current password.** Without it, a forged
  request could take the device permanently.
- **The JWT check accepts only the algorithm the server issues.** Accepting
  whatever the token declares is the classic way to forge one.
- **A missing secret key fails the server rather than falling back to a
  guessable one.** A default that works is a default nobody replaces.
- **Downloads are verified before they are trusted or stored.** The updater
  checks a size and a checksum before writing to the SD card, and refuses a
  manifest it cannot make sense of.
- **API keys.** Scripts authenticate with a key instead of borrowing a browser
  session, and the routes that mint keys need a session, so a stolen key cannot
  issue successors.
- **The USB gadget has a stable identity and a safe default.** The device no
  longer presents a changing identity to the host it controls.
- **The login lockout survives a table full of other addresses.** An attacker
  could previously flush their own record out of it.
- **The board stops answering DHCP on the wired network.** A KVM that hands out
  addresses on a network it was plugged into is a fault that is hard to trace.

## Reliability

This is the largest group, and the reason the fork exists.

- **A/B root slots and a recovery slot.** The card carries two full systems and a
  small recovery one. A new build is installed to the slot that is not running
  and tried once. If it does not come up, the board returns to the slot that
  works, and if neither works it lands in recovery, which runs a network, ssh and
  a serial console and nothing else.
- **A boot watchdog.** If the board is not reachable within five minutes it acts
  rather than waiting for a person. It distinguishes a slot being tried, which
  reverts, from a board that has lost both doors while running, which goes to
  recovery. An unplugged network cable stands it down, because rebooting cannot
  fix a cable.
- **Identity survives a slot switch.** The root password, the web password, the
  ssh host keys and `/root/.ssh` used to be part of whichever root filesystem was
  running, so switching slots reverted the board to factory credentials and
  changed its ssh fingerprint. They now live on the data partition, and a
  password change writes itself back with no command to remember.
- **The server restarts when it dies.** Nothing supervised it before.
- **A boot-script update can undo itself.** A package update that installs boot
  scripts keeps what it replaced. If the board cannot be reached afterwards, the
  watchdog puts them back and reboots, before it considers recovery: a board
  that stops answering right after an update is very probably broken by that
  update, and undoing it keeps video and HID that recovery does not. A counter
  covers the case where the power is cut before the watchdog's deadline.
- **HID survives a USB gadget rebuild.** Keyboard and mouse used to stop working
  until a reboot.
- **The capture pipeline tears down in the right order.** The server waited on
  threads after dismantling what they used, and reported the pipeline as running
  when its initialisation had failed.
- **The HDMI resolution probe stops spinning on an unreadable input.**
- **A stalled viewer no longer freezes the stream for everybody else.**
- **The terminal socket closes when its writer gives up.**
- **The mouse jiggler runs one loop, with its state guarded.**
- **A deploy that does not serve restores the previous server automatically.**

## Performance and memory

The board has one core and about 158 MB of usable memory, so this is not
micro-optimisation.

- **Every MJPEG client gets its own queue and writer.** One slow client used to
  hold up all of them.
- **Frames are copied fewer times**, and each WebRTC frame is packetised once
  for all viewers rather than once per viewer.
- **The H.264 path no longer builds a second copy of every frame.**
- **zram**, so the board has somewhere to put cold pages. Measured cost is
  nothing on both H.264 paths, and about 14.5% of frame rate on MJPEG under the
  same burst load.
- **Cache headers on the web UI**, so a browser stops refetching what has not
  changed.
- **Runtime state is kept off the SD card.** The frame rate counter, the HDMI
  presence flag and the server log wrote to the boot card continuously.
- **The OLED image moves**, to spread the wear of a static picture.
- **The capture log stops repeating a failure once per frame.**

## Features

- **USB audio.** The board presents a UAC1 gadget and carries the host's audio to
  the browser. HDMI on this hardware cannot carry audio: the pins are unwired,
  per Sipeed. This is the whole path, from capture to the speaker button.
- **A serial console for the managed host**, off by default. The host sees a USB
  serial port, so a machine whose network is down can still be reached through
  the board.
- **An HTTP proxy setting**, for a board on a network with no direct route out.
- **A custom update server**, which is what IronKVM's own feed uses.
- **Reporting of whether the HDMI port actually carries a picture**, rather than
  whether a cable is present.
- **A build stamp**, so the running binary can be identified. The application
  version alone is written by the updater and says nothing about the binary.
- **A `novision` build tag**, which stubs the device-native bindings so the tree
  can be built and tested off the board.

## Changed behaviour

Things an existing user would notice.

- The board no longer answers DHCP on `eth0`.
- The HDMI receiver is no longer power-cycled at startup.
- The default update server is IronKVM's feed. The official one is one click
  away in `Settings > Update`.
- The favicon and the login logo are IronKVM's. `/boot/logo.ico` still replaces
  both, exactly as before.
- Sipeed's X and Discord links are gone from the About panel. They are where the
  hardware vendor supports its own firmware, and a report about this one spends
  somebody else's time. The hardware links remain.

## What still says NanoKVM

Deliberately, and worth knowing before you go looking for a bug.

- **The OLED screen and the Wi-Fi access point SSID.** Both live in C++ inside
  `kvm_system`. Changing them means rebuilding through MaixCDK and shipping that
  binary in every release, which drags the whole toolchain into the release path
  for two strings, and changing the SSID would break the Wi-Fi provisioning flow
  for anyone following Sipeed's documentation.
- **Every path on disk.** `/kvmapp`, `/etc/kvm`, `NanoKVM-Server`, the API
  routes and the encryption key are unchanged. That is what lets an official
  package install over IronKVM, and lets you go back.
- **The strings that name the hardware.** The board is a NanoKVM whatever
  firmware it runs, so "press the BOOT button on the NanoKVM" stays correct.
- **`Image Version` in the About panel.** That value is the version of Sipeed's
  system image, and renaming it would misreport what it shows.

## Not here yet

- **Slot-aware updates.** An update installs in place on the running slot and is
  protected by the boot-script rollback rather than by a slot trial. Moving the
  updater onto the slot machinery is the next piece of work.
- **Growing the data partition.** The card image carries a fixed partition
  table, so a card larger than 32 GB leaves the extra space unused.
- **A preview channel.** The code supports one. There is nothing to preview yet,
  and a channel with nothing in it is a promise that has to be kept.
- **An upgrade in place from the official firmware.** You install IronKVM by
  writing the card image. A board running Sipeed's firmware cannot reach IronKVM
  through `Settings > Update`.

  The official updater refuses a package that is not named `nanokvm_X.Y.Z.tar.gz`,
  and it does not run the hook that installs the boot scripts. A package renamed
  to satisfy it would replace the application and leave `/etc/init.d` alone, so
  the board would get the server and the web user interface and none of the A/B
  slots, the boot watchdog or the identity carry-over. Nothing on the board would
  say so, and no later update would repair it.

  A migration needs the server to bring the boot scripts up to date by itself.
  That is a piece of work, not a rename, so 1.0 does not claim it.
