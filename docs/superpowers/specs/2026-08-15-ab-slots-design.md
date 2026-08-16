# A/B root filesystems, second design

Status: design approved 2026-08-15. Not implemented.

## Why this exists

On 2026-08-15 the board stopped answering. It replied to ping and to ARP for
more than 25 minutes, over a cold power cycle, and no TCP port accepted a
connection. The USB gadget did not enumerate on the managed host, so the ACM
console was gone as well. The card had to come out.

Three recovery mechanisms were in place and none of them acted:

- The initramfs fallback catches a slot that will not mount. A root did mount,
  so the fallback never ran.
- `S00awatchdog` reverts the marker and reboots. It was not installed in the
  running slot, and a boot that stops before `S00` cannot reach it anyway.
- `deploy-server` restores the previous binary. The server binary was not the
  fault.

The symptom matches a failure this repository has already recorded in
`tools/slots/README.md`: `S03usbdev` and `S95nanokvm` exit 127, and `S50sshd`
exits 0 having started nothing. A CRLF in a device script gives the first. A
factory `/etc/kvm/ssh_stop` gives the second.

The lesson is one sentence. **A guard that the suspect runs is not a guard.**

## Decisions

| decision | choice |
| --- | --- |
| Layout | Two real root partitions, plus a small recovery root. |
| Image source | A pinned base plus a manifest. The cloner discovers the first manifest and is then a capture tool only. |
| Who disarms a bad slot | The initramfs disarms. Userspace confirms. The watchdog stays as a second layer. |
| What confirms | Reachable: `eth0` has an address, and the web server or `sshd` answers. |
| Identity | Persistent on `/data`, bind-mounted over `/etc/kvm`. Images hold no secrets. |
| Base release | The Sipeed release closest to what the board runs, read from `/kvmapp/version`. |

## Partition layout

Measured on the card, not assumed:

```
Disk /dev/sdg: 60506112 sectors of 512 bytes, 28.85 GiB, MBR
sdg1  *      1 ..    32768     32768 sectors   16M  c  W95 FAT32 (LBA)
sdg2      32769 .. 16000000  15967232 sectors  7.6G  83 Linux
sdg3   16001024 .. 60506111  44505088 sectors 21.2G  83 Linux
```

**p1 starts at sector 1.** There is no gap before it, and `fip.bin` is a 440832
byte *file* inside the FAT partition beside `boot.sd`. An earlier draft of this
document said `fip.bin` lives in raw sectors before p1. That is wrong for this
board, and it mattered, because it made the migration guard the wrong thing.

The real constraint is therefore: **p1 keeps its type, its boot flag, its start
sector and its contents.** The boot ROM finds `fip.bin` by reading the first
FAT partition, so the partition may not move, may not change type, and may not
be reformatted. Everything from sector 32769 upward is ours.

```
p1  FAT16     /boot       boot.sd and the markers
p2  ext4      root A
p3  ext4      root B
p4  extended
 p5 ext4      recovery
 p6 exfat     /data       images, identity, logs, backups
```

MBR gives four primary partitions and this design needs five filesystems, so
one extended partition holds the last two. GPT is tidier. It is not used here,
because the boot ROM and this u-boot have not been tested with it.

### Sizes

The rule is: p1 unchanged, recovery 256MB, root A and root B equal at 1.5 times
the present root image, `/data` takes the remainder. A card that cannot satisfy
it is too small for A/B, and the operator must learn that before the partition
table is written.

Applied to this card. The present root is a 3 GiB image that holds 254109 of
786432 blocks, so about 993MB is in use and 4.5GB per slot leaves room to grow
without being wasteful:

```
p1  FAT16      1 ..    32768        16M   /boot     unchanged, boot flag kept
p2  ext4   32769 ..  9469952       4.5G   root A
p3  ext4 9469953 .. 18907136       4.5G   root B
p4  extended
 p5 ext4                            256M  recovery
 p6 exfat                          ~19.5G /data
```

`/data` loses about 1.7GB against the 21.2GB it has today. Its present contents
are copied off before the table is written, so the loss is planned, not
discovered.

### The partition number trap

`S01fs` runs `mkfs.exfat` on `/dev/mmcblk0p3` when `/boot/usb.disk0` exists and
`/etc/kvm.disk0` does not. In this layout p3 is a root filesystem. The rule is
therefore absolute: **no script computes a device from a partition number.**
`S01fs` reads its device from a manifest value, and the initramfs takes the
three device names from variables that the build writes.

## Marker protocol

Four files on the FAT p1. FAT because the initramfs can write it with no
journal and no `fsck`.

| file | meaning |
| --- | --- |
| `/boot/slot` | The trusted slot. If the file is absent, the trusted slot is A. |
| `/boot/slot.try` | A slot on trial. The operator writes it. The initramfs deletes it. |
| `/boot/slot.prev` | What `slot` held before the last confirm. |
| `/boot/recovery` | If the file is present, boot recovery and do nothing else. |

### Initramfs sequence

1. If `/boot/recovery` is present, mount recovery and hand over.
2. If `/boot/slot.try` is present, read it, delete it, and sync.
3. If the delete did not persist, discard the trial and boot the trusted slot.
4. Otherwise mount the trusted slot, or A if the marker is absent.
5. If a mount fails, try the trusted slot, then recovery, then `msc`.

Step 2 is the whole design. The initramfs disarms the trial before it hands
over, so a hang, a panic, an oops, a power cut and a slot that starts no
listener all reach the same next boot, which is the trusted slot. The suspect
gets one attempt and never votes on its own trial.

Step 3 exists because a delete on FAT is a directory-entry write. If the write
does not land and the trial slot hangs on every boot, the board loops for ever,
which is worse than the dead board this design prevents. Proof of disarm before
the trial is honoured turns that loop into one wasted boot.

### Confirm

The trial slot writes its own marker when it is reachable. Reachable has one
definition, and both users of it read the same function: `healthy()` in
`S00awatchdog`, graded by the table in the watchdog section below. The confirm
step calls that function rather than repeating its rule, because two copies of a
health test drift, and the copy that drifts is the one nobody runs.

```sh
cp /boot/slot /boot/slot.prev
echo <slot> > /boot/slot
sync
```

Reachable is the bar because reachable is the property that decides whether the
card comes out. Broken HID, dead capture or a wrong binary are repairable over
`ssh`, so they must not hold a trial open.

### Failure table

| failure | before | after |
| --- | --- | --- |
| Slot will not mount | Falls back to A | Falls back to recovery |
| Slot mounts, boot stops early | Card comes out | One reboot, trusted slot |
| Slot boots, no listener | Card comes out | One reboot, trusted slot |
| Kernel panic in the trial | Card comes out | One reboot, trusted slot |
| The trusted slot rots | Card comes out | The watchdog escalates to recovery |

## Initramfs changes

The patch extends `tools/slots/init-slot-selection.inc` and
`tools/slots/init-mount-dispatch.inc`. The try marker is handled in the
selection block, because the dispatch block runs after `/boot` is unmounted.

Device names are written at the top of `/init` by the build:

```sh
SLOT_A=/dev/mmcblk0p2
SLOT_B=/dev/mmcblk0p3
RECOVERY=/dev/mmcblk0p5
```

`mount_fs` keeps its single `e2fsck -fvp` retry. Every step logs to `/dev/kmsg`
with the `nanokvm-slot:` prefix, so a board that fell back explains itself over
the network.

`loop:` support is removed. Real partitions make it unnecessary, and each branch
is a branch that must be tested. The block-device form stays, because it costs
one `case` arm and it is how an operator boots something ad hoc.

The initramfs uses `/busybox <applet>` for any applet that is not one of the
twelve symlinks. `test-commands.sh` already fails a build that breaks this rule.

`repack-boot.sh` is reused without change. It refuses to emit an image unless
the kernel and the device tree come back byte-identical, only `/init` differs,
`dev/console` survives as char 5,1, all twelve symlinks survive, and every
command resolves under `PATH=/`.

## Recovery root

Recovery is the base image with a manifest of four entries:

| entry | why |
| --- | --- |
| `S30eth` | The network door. |
| `sshd`, started unconditionally | The way in. |
| `S03usbdev`, reduced to the `acm` function only | A door that needs no network. |
| `slot`, `e2fsck`, `mount`, `tar` | So an operator can repair from here. |

Recovery carries no `kvmapp`, no server, no HID, no capture, no zram and no
picoclaw. It cannot fail the way the board failed, because it does not run the
things that failed.

The ACM door matters because the two doors fail for unrelated reasons. On
2026-08-15 the network was healthy and the board was still lost. The opposite
case, a healthy board with a broken network configuration, is equally real.

Precondition: the root password must not be the factory password, because
whoever controls the managed host gets a login prompt.

Recovery is sticky. It does not delete its own marker. An operator leaves with
`rm /boot/recovery && reboot`, which is possible because recovery answers
`ssh`. The one-shot disarm belongs to trials, not to the safe place.

## Watchdog, second layer

`S00awatchdog` goes in the manifest of every full root, so "it was not
installed" cannot happen again.

The health rule changes. The present rule is `net_up AND (web_up OR ssh_up)`,
and `net_up` needs carrier. An unplugged cable therefore reboots the board into
recovery, where the cable is still unplugged. That is a reboot loop dressed as a
safety feature, and it punishes an external fault.

| carrier | address | listener | verdict |
| --- | --- | --- | --- |
| No | any | any | Stand down. Recovery cannot repair a cable. |
| Yes | No | any | Escalate. |
| Yes | Yes | None | Escalate. |
| Yes | Yes | Web or ssh | Healthy. |

To escalate is to log to `/watchdog.log` in the image, `touch /boot/recovery`,
`sync`, and `reboot -f`.

Escalation has two destinations, and the watchdog chooses between them:

| This boot | Escalation goes to | Why |
| --- | --- | --- |
| The trusted slot | Recovery | There is no better slot to fall back to. |
| A slot on trial | The trusted slot | It was working minutes ago. |

A trial that fails must not reach recovery. Recovery serves no video and no
HID, so it costs the board its whole function, and its marker is sticky, so it
costs the operator a second reboot to leave. The trusted slot is a better
fallback and it is already armed: the initramfs deleted `slot.try` before the
handover, so a plain `reboot -f` lands on it.

No marker records that a boot is a trial, and none is necessary. The running
slot is what is mounted on `/`, the trusted slot is `/boot/slot`, and a boot
where those two differ is a trial. `slot status` reads both. If either reads
blank, or the running slot is `unknown` or `recovery`, the watchdog takes the
recovery branch: a watchdog that cannot tell where it is must not assume it has
a good slot to fall back to.

The watchdog probes a listener, not a process. `pidof sshd` passes on a board
whose server died.

## Image build

### Base

One pinned Sipeed release rootfs. Its sha256 is recorded in the repository. A
slot's provenance is then two facts: which base, and which manifest.

### Manifest

The manifest needs three lists, not one:

```
add:     kvmapp/ , the fork init scripts, NanoKVM-Server, dl_lib/, web/
remove:  /etc/kvm/ssh_stop and every other file the base has and we do not want
touch:   /etc/kvm.disk0
set:     DATA_DEV=/dev/mmcblk0p6
```

The `remove` list is not an afterthought. `cp -a` adds files and overwrites
files. It never deletes. A file that the factory has and the running board does
not will survive into the image, which is how `/etc/kvm/ssh_stop` reached a
slot and stopped `sshd` while reporting success.

### Identity

`/etc/kvm/` holds the bcrypt password, the TLS key and certificate, the picoclaw
internal token and the API keys. A slot built from a factory base loses all of
it and the board returns as `admin`/`admin`. An image that carries it holds a
device private key.

Identity therefore lives in `/data/identity/` and is bind-mounted over
`/etc/kvm` early in boot. A slot switch keeps the password, the certificates and
the tokens. Images hold no secrets. `ssh_stop` cannot return from a base,
because `/etc/kvm` is never the base copy.

Recovery must handle a missing or unmountable `/data`. It falls back to the
image copy of `/etc/kvm` and says so on the console.

### Mechanics

The image is built in a container on a workstation, not on the board:

```sh
mke2fs -d <staged tree> -t ext4 -L slotb -U <fixed uuid> slotb.img
```

`mke2fs -d` populates from a directory with no loop mount and no privileged
mount, which is what makes the build possible on Windows through Docker.

Byte-for-byte reproducibility is not promised, because ext4 records timestamps.
Content reproducibility is: build twice and compare file inventories with modes,
links and device nodes, the way `repack-boot.sh` compares initramfs inventories.

### Gates

An image is not written unless every gate passes. Each gate names a fault this
board has had:

| gate | fault |
| --- | --- |
| Every `/etc/init.d/S*` and `*.sh` is LF only | `rc=127`, silent and total |
| Every `/etc/init.d/S*` is mode 755 | The script never runs |
| `/etc/kvm.disk0` is present | The first boot reformats `/data` |
| `/etc/kvm/ssh_stop` is absent | `S50sshd` exits 0 and starts nothing |
| `S00awatchdog` is present and 755 | The 2026-08-15 outage |
| Every file in the `remove` list is gone | The same trap, generally |
| `sh -n` accepts every shell script | A syntax error at boot |
| No `/swapfile` | Swap inside a loop image deadlocks reclaim |

## Operator interface

One tool, `slot`, installed in every full root and in recovery.

```
slot status
slot install b /data/slot-20260816.img.gz
slot try b
slot confirm
slot revert
slot recovery
```

Each image carries `/etc/slot-manifest`, written at build time, so `slot status`
reports what every partition holds:

```
running   B   base 1.4.3  manifest a91c3f4  built 2026-08-16
trusted   B
trial     none
A         base 1.4.3  manifest 77b0e21  built 2026-08-09
recovery  base 1.4.3  manifest 0c4d19a  built 2026-08-16
prev      A
```

`install` follows the guards that `install-boot.sh` already applies to
`boot.sd`: refuse the running root, refuse an image that does not fit, sha256
the source, `dd`, read back, and compare. It also refuses an image with no
`/etc/slot-manifest`, so an unidentified blob cannot become a slot.

Recovery adds the repair verbs, because recovery is where an operator lands
when the board is bad:

```
slot mount a
slot fsck a
slot restore a /data/slot-known-good.img.gz
slot boot a
```

### A/B is not the deploy path

A root partition is about 3GB. Compressed it is 300MB to 600MB over the network,
and then `dd` writes it to a slow SD card. One slot install costs 10 to 20
minutes and consumes write endurance.

`deploy-server` stays the common case. It replaced the server binary on
2026-08-15 in about 90 seconds, with automatic restore. Slot cycling is for a
new base, a new kernel, a new init layout, or a rescue. The README must say so,
or operators reach for the heavy path from habit.

## Testing

| test | subject |
| --- | --- |
| `test-dispatch.sh` | The initramfs scenarios below, in a sandbox |
| `test-mutation.sh` | Breaks dispatch on purpose and fails if the checks pass |
| `test-watchdog.sh` | `healthy()`, including the carrier rule, extracted from the script |
| `test-manifest.sh` | The eight build gates |
| `test-slot-tool.sh` | `install` guards against a loop file |
| `test-image-build.sh` | Two builds, compared by inventory |

Dispatch scenarios:

| scenario | expected |
| --- | --- |
| No markers | Slot A |
| `slot=b`, no trial | Slot B |
| `slot=b`, `try=a` | Slot A, and `slot.try` is gone |
| The trial will not mount | Trusted, then recovery |
| The trial marker cannot be deleted | The trial is discarded, trusted boots |
| The trusted slot will not mount | Recovery |
| Everything fails | `msc` |
| The marker is unreadable | Trusted, and the reason is logged |

### Acceptance

Two images are built to be broken on purpose:

1. `S95nanokvm` saved with CRLF, so it exits 127.
2. A script that never returns, placed before `S50sshd`.

`slot try` each one and reboot. The board must return to the trusted slot within
one boot, and `dmesg | grep nanokvm-slot` must explain why. If this test does
not pass on hardware, the design is not delivered.

## Migration

The order matters, because step 2 is the only copy of a board that worked.

1. Read the card before writing to it: `fdisk -l`, `/watchdog.log`,
   `ls -la /etc/init.d/`, `/kvmapp/version`, and the logs under `/data/deploy/`.
2. Take a full compressed card image to the workstation.
3. Copy `/data` off.
4. Extract `/etc/kvm/` from the good slot. It becomes `/data/identity/`.
5. Clone the good slot, compare it against the pinned base, and keep the
   difference. That difference is the first manifest.
6. Repartition, and leave the p1 entry byte-identical: same start sector, same
   length, same type `c`, same boot flag, and no reformat. The boot ROM finds
   `fip.bin` by reading that partition, so losing it is the one mistake in this
   plan that no recovery in it can undo. Rewrite the entries for p2 upward only,
   and read the table back before writing any filesystem.
7. Write recovery first and boot it on purpose. Prove `ssh` answers before root
   A exists. If recovery does not work, nothing below it works either.
8. Boot root A, verify it, then build root B from the same manifest.

## Out of scope

- A board that is reachable but useless. Carrier, address and a listener are all
  present, and the KVM does nothing. That board is repairable over `ssh`, and
  reachable is the agreed bar.
- Reproducible image bytes. Content reproducibility is specified instead.
- GPT partitioning.
- Any change to `boot.sd` other than `/init`. The kernel and the device tree
  come back byte-identical, and `repack-boot.sh` enforces it.
