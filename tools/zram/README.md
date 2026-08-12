# zram

Compressed swap in RAM, as two loadable modules against the stock kernel. No
new image, no kernel replacement, no boot-path change; `rmmod` undoes it.

## Why it is possible

```
# CONFIG_MODVERSIONS is not set     only vermagic must match, no symbol CRCs
# CONFIG_MODULE_SIG is not set      no signature enforcement
CONFIG_MODULES=y                    ~30 .ko already load from /mnt/system/ko
CONFIG_SWAP=y  CONFIG_CRYPTO_LZO=y  CONFIG_CRYPTO_ZSTD=y
# CONFIG_ZSMALLOC is not set        so zram needs zsmalloc built too
```

The board's own toolchain is the one in `tools/build` — Xuantie V2.6.1, gcc
10.2.0 — which is what `CONFIG_CC_VERSION_TEXT` on the device names.

## Why lzo-rle

`lz4` is not an option: `CONFIG_CRYPTO_LZ4` is unset, so `/proc/crypto` offers
only `deflate`, `lzo`, `lzo-rle` and `zstd`. The real choice is lzo-rle against
zstd.

The scarce resource on this board is CPU, not compression ratio: one in-order
C906 at 1GHz with no acceleration, also running H.264 encode. Compression is
paid on swap-*out*, which happens under memory pressure — exactly when that core
is already saturated. zstd costs several times more there for a better ratio.

## Reconsidered, with numbers

The reason above is a *kernel config* constraint, and this repository can now
build modules against the stock kernel — that is what `build-modules.sh` does.
So `CONFIG_CRYPTO_LZ4` could be built and lz4 offered to zram. Measured on a
live board, it would buy nothing:

```
zram offers   : lzo [lzo-rle] zstd
ratio         : 2.76x   (orig 3.21 MB -> compressed 1.17 MB)
mem_used      : 1.63 MB against the 40 MB mem_limit -> 4% of the cap
swap in use   : 3.2 MB of 96 MB
lifetime      : pswpout 40449 pages, pswpin 6503
```

Ratio is not the binding constraint: at 4% of the cap, a much worse ratio would
still fit. lz4 over lzo-rle is a few percent of compression time on a board that
idles at 5% of one core. And lzo-rle is the kernel's default for zram because of
its run-length handling of zero pages, which is most of what swap holds — lz4
has no equivalent. Against that, an extra module pinned to the kernel's vermagic
has to be rebuilt whenever the kernel moves.

If this is revisited, measure CPU during a sustained swap-out, not the ratio.
The ratio is already comfortable and is not what costs anything here.

An earlier version of this file claimed lzo-rle "reached about 5x". Today it
measures 2.76x on live pages. Both are real; they are different workloads, and
a single figure should not have been stated as the property of the algorithm.

Switching is one write, before `disksize` is set:

```shell
echo zstd > /sys/block/zram0/comp_algorithm
```

## Build and install

The modules are the only manual step. The server installs the init script.

```shell
ssh root@<device> 'zcat /proc/config.gz' > kernel.config
tools/zram/build-modules.sh /path/to/linux_5.10 kernel.config ./ko
scp ko/*.ko root@<device>:/mnt/system/ko/
```

Then open `Settings > Device > Advanced` in the web UI and turn on **Compressed
swap (zram)**. The toggle copies `/kvmapp/system/init.d/S01zram` to
`/etc/init.d/`, makes it executable, and starts it. Turning the toggle off
stops the device and removes the script again.

The row reports "The kernel modules are not installed on this device" and the
toggle stays disabled until both `.ko` files are in place.

`build-modules.sh` refuses to emit modules whose vermagic does not match
`5.10.4-tag- preempt mod_unload riscv` exactly, so a module that cannot load
never reaches a device.

`loadsystemko.sh` lists modules explicitly and does not glob, so dropping files
into `/mnt/system/ko` does not make them load. `S01zram` is what enables them.

## Where the script lives

`S01zram` moved to `kvmapp/system/init.d/` so that it ships in the install
package and the server has a source to copy from. It is not in the list of
scripts that `kvm_system` copies at boot: that list is hard-coded C++, and
adding a name to it needs a MaixCDK rebuild and a redeploy on every device. The
presence of `/etc/init.d/S01zram` is what marks the feature as enabled, the
same way `S98tailscaled` already works.

`tools/zram/test-zram-swapon.sh` checks the swap priority and its fallback:

```shell
sh tools/zram/test-zram-swapon.sh
```

## Swap priority

The script asks for `swapon -p 100`. zram and the optional swap file are both
swapped on during boot, and the order between them is not defined: `S01zram`
runs from `rcS`, and the swap file runs from the `si11::sysinit` line that the
server appends to `/etc/inittab`. Without a priority, the swap file can win, and
the kernel then writes to the SD card before it writes to compressed RAM.

BusyBox documents `swapon -p PRI`, but an applet built without the option
rejects it. The script falls back to a plain `swapon`, so a rejected option
costs the priority and not the swap.

## The trade

With no disk swap behind it, exceeding zram means the OOM killer rather than
slow paging, and the largest process is `NanoKVM-Server` — so an OOM takes video
with it. `mem_limit` caps the RAM zram may consume so it cannot spiral, but the
tail behaviour is a cliff rather than a slope. On a board that measured 143
pages ever swapped, and zero under a live streaming session, that cliff is a
long way off.
