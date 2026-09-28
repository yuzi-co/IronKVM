# Network Boot for the Host

**Date:** 2026-09-28
**Issue:** yuzi-co/ironkvm-dist#14
**Builds on:** the USB network link, yuzi-co/ironkvm-dist#10 (`2026-09-28-usb-network-design.md`).
**Repos:** IronKVM only. Nothing here needs a kernel option or a capability word beyond
`usb-network`, so ironkvm-dist does not change.

## Purpose

The host boots an installer or a rescue system from the network, together with virtual media, and
nobody has to touch it. There are three ways, from the cheapest to the most invasive.

1. **Boot media.** netboot.xyz as an ISO in the image directory, inserted into the virtual CD like
   any other image. Works on any host and needs nothing on the board.
2. **PXE over the USB link.** The board answers the host's PXE request on the link that #10 built,
   hands it iPXE, and iPXE loads a menu from the board: every ISO in the image directory, and
   netboot.xyz.
3. **Proxy DHCP on the LAN.** Off by default. The board answers PXE requests on the LAN with a boot
   loader only and never hands out an address.

## What the board has

BusyBox (busybox-extras) gives `udhcpd`, an inetd-style `tftpd` and `httpd`. There is no dnsmasq.
apk-tools 3.0.8 is present, and the NetBird add-on already fetches, verifies and extracts an Alpine
package into `/data` with it. The CPU is riscv64. The host gets no internet over the USB link, which
never routes.

## Decisions

### Boot media

- The image downloader gets a **Boot menu** button. It downloads netboot.xyz 3.0.3's
  `netboot.xyz.iso` from the GitHub release into the image directory through the existing download
  service, with the release's SHA-256 pinned in the server. The web UI sends no URL; the server
  owns the pin, so a changed page cannot point it anywhere else.

### dnsmasq, an add-on

- dnsmasq comes from Alpine v3.24 main, like NetBird comes from community: `apk update`,
  `apk fetch dnsmasq`, `apk verify`, `apk extract --no-chown`, and `usr/sbin/dnsmasq` moves to
  `/data/ironkvm/addons/netboot/dnsmasq`. Only that package is fetched. It needs
  `so:libc.musl-riscv64.so.1`, which the image has; `dnsmasq-common` holds only the sample
  configuration and the package scripts, which nothing here uses. 2.92_p2 today.
- The add-on is called `netboot` and keeps its boot files beside the binary in
  `/data/ironkvm/addons/netboot/tftp/`. It is recorded with `S85netboot` as its boot script, so
  `S04addons` copies the script to `/etc/init.d` while LAN proxy DHCP is on. It needs no link: the
  script runs the binary from `/data`, and `/usr/sbin/dnsmasq` stays absent, so `S80dnsmasq` still
  exits at once.
- Install needs a distribution image with `/data` mounted. The boot files go to `/data`, not to
  the slot and not to `/tmp`.

### Boot files, pinned

- **iPXE v2.0.0**, the first iPXE release with published binaries: `ipxeboot.tar.gz`
  (SHA-256 `01a526d4…17eee1`). The server downloads it into the add-on's directory, checks the
  archive, takes out three files and checks each against its own pinned sum:
  `x86_64/undionly.kpxe` (BIOS), `x86_64/ipxe.efi` (UEFI x86-64) and `arm64/ipxe.efi` (UEFI arm64,
  saved as `ipxe-arm64.efi`). The archive is removed afterwards.
- **netboot.xyz 3.0.3:** `netboot.xyz.kpxe`, `netboot.xyz.efi` and `netboot.xyz-arm64.efi`, each
  checked against the release's `netboot.xyz-sha256-checksums.txt`, pinned in the server. These
  carry netboot.xyz's own embedded script, which loads the netboot.xyz menu from the internet.
- The server also writes `boot.ipxe`, three lines that chain iPXE to the board's HTTP menu through
  `${next-server}`, so nothing in the TFTP root depends on the subnet.

### Who runs dnsmasq

`S03usbdev` owns the link and already runs its `udhcpd`, so it stays the one place that decides
what serves DHCP on the link. The VPN add-ons are run by their own init scripts, so dnsmasq gets
one too: `kvmapp/system/init.d/S85netboot`, which never decides, only starts and stops.

- `S85netboot start-usb <if> <host> <mask>` starts the link instance, and fails when network boot
  on the link is off (no `/etc/kvm/netboot/usb.conf`) or dnsmasq is not installed.
  `stop-usb` stops it.
- `S85netboot start`, `stop` and `restart` handle the LAN instance, from
  `/etc/kvm/netboot/lan.conf`. `start` exits 0 when that file is absent.
- `S03usbdev`'s `usb_net_start` asks `S85netboot start-usb` first and runs `udhcpd` only when it
  fails. `usb_net_stop` stops both. A new action, `S03usbdev dhcp`, restarts only the link's DHCP
  server, so the server can switch between `udhcpd` and dnsmasq without re-enumerating the gadget
  and costing the host its keyboard.
- The script joins the `addons` cgroup before it starts dnsmasq, as `S98netbird` does, and starts
  and stops by pid file with the `/proc/<pid>/cmdline` check, as `S98netbird` does. dnsmasq runs
  in the foreground under the script's `&`, drops to `nobody` after it binds, writes its leases
  and its log to `/tmp/netboot`.
- Two instances, one per side, each with one `--interface` and `bind-interfaces`. With exactly one
  interface dnsmasq binds its DHCP socket to the device, which is also how the Wi-Fi access
  point's `udhcpd` and the link's `udhcpd` already share port 67.

### The configuration

The server generates both files in Go from `server.yaml`'s `netboot` block and writes them to
`/etc/kvm/netboot`, on `/data`, so the link comes up with network boot at the next boot before the
server runs.

The link file has no subnet in it. `S03usbdev` passes the interface and
`--dhcp-range=<host>,<host>,<mask>,10d` on the command line, from the same `usb_net_parse` that
feeds `udhcpd`, so a subnet change needs no new file. It keeps #10's rules:

- `port=0`: no DNS server at all.
- `dhcp-option=option:router` and `dhcp-option=option:dns-server` with no value. dnsmasq otherwise
  offers its own address as both.
- `dhcp-lease-max=1` and `dhcp-authoritative`: the host is the only client.
- No forwarding: that is `S03usbdev`'s `usb_net_no_forward`, which runs as before.
- The boot file by option 93: `client-arch` 0 gets `undionly.kpxe`, 7 and 9 get `ipxe.efi`, 11 gets
  `ipxe-arm64.efi`. A client that says user-class `iPXE` gets `boot.ipxe` instead, which chains to
  `http://${next-server}:8069/menu.ipxe`.
- `enable-tftp` with the add-on's `tftp` directory as its root.

The LAN file is proxy DHCP and nothing else:

- `dhcp-range=<network>,proxy,<mask>`, for the network of the interface that holds the default
  route. There is no address range, no `dhcp-authoritative`, no lease file and no DHCP option.
- `pxe-service` entries for x86PC, X86-64_EFI, BC_EFI and ARM64_EFI, tagged `!ipxe`, give each
  architecture its netboot.xyz binary by TFTP. With one entry for each architecture a PXE ROM
  boots it at once, with no menu.
- netboot.xyz's iPXE then loads its menu from the internet. The board's HTTP menu, and the
  images in `/data`, are never offered on the LAN. The page says so.
- The server checks the LAN interface and its network every 10 seconds and rewrites the file and
  restarts the instance when either changes.

### The menu and the images, on the link only

- A second HTTP listener in the server, plain HTTP on `<board>:8069`, the board's address on the
  link. It binds only while network boot on the link is on and the address exists, and it follows
  a change of subnet. On top of the bind, every request whose client is not inside the link's
  subnet is refused with 403, so a packet sent to that address from the LAN gets nothing even
  though Linux would accept it on any interface.
- `GET /menu.ipxe`, generated per request: one `sanboot --no-describe` entry for each ISO in the
  image directory, netboot.xyz, an iPXE shell, and exit to the next boot device. An image whose
  path holds anything but letters, digits, `.`, `_`, `-` and `/` is left out, because the path goes
  into an iPXE script.
- The netboot.xyz entry chains the netboot.xyz binary for the platform over HTTP from the board.
  netboot.xyz needs the internet, which the link does not give; its binary tries the host's other
  network interfaces. The menu entry says so.
- `GET /iso/<path>` serves an image with range requests (`http.ServeContent`), which `sanboot`
  needs. Only a path the image listing returns is served, after `EvalSymlinks` keeps it inside
  the image directory.
- `GET /boot/<file>` serves the three netboot.xyz binaries and nothing else.

### Settings and API

`server.yaml`:

```yaml
netboot:
  usb: false   # network boot on the USB link
  lan: false   # proxy DHCP on the LAN
```

Admin routes under `/api/netboot`:

| Route | |
| --- | --- |
| `GET status` | settings, dnsmasq installed and its version, boot files present, both instances running, the link's mode and addresses, the LAN interface and network, the menu URL, the host's lease, the last boots the HTTP server saw, and the last lines of each instance's log |
| `POST settings` | `{usb, lan}`. Refused while the add-on is not installed. Rewrites the files, restarts what changed, saves |
| `POST install` | dnsmasq and the boot files |
| `POST uninstall` | turns both off, gives the link back to `udhcpd`, removes the add-on |

`POST /api/download/image/netboot` starts the netboot.xyz ISO download.

### Web

A **Network boot** page in Settings: install and uninstall, a switch for the USB link with the
link's state (and a pointer to Device settings when the link is off), a switch for the LAN behind a
confirmation with a warning, and the status. The image downloader gets the Boot menu button. All 24
locales carry the new strings.

## Serving `/data` images on the LAN: a proposal, not built

Proxy DHCP answers any PXE client on the LAN, and an HTTP server for `/data` on the LAN would hand
every image to anybody who asks. A safe version would be armed per boot:

- The operator picks an image and a host MAC address in the page and arms it for, say, ten
  minutes.
- The proxy offers the menu only to that MAC (a `dhcp-host=<mac>,set:armed` tag on the
  `dhcp-boot`), and the menu URL carries a random one-time token.
- The LAN listener serves only that image, only with the token, only to the address that fetched
  the menu, and closes when the time runs out or the host has booted.

Spoofing a MAC on the LAN is easy, so this narrows the window rather than closing it. It is worth
building only if the owner needs it.

## Tests

- Go: the link file (no router, no DNS, no range, one lease, arch matching, the user-class chain),
  the LAN file (a proxy range only, never an address range, never a DHCP option), the menu (ISOs,
  unsafe names left out, netboot.xyz), the image server (range requests, a path outside the
  listing, a client outside the link), the pins, the lease and log readers, the settings flow with
  stubs, and the agreement between the Go paths and `S85netboot`.
- Shell: `tools/service/test-netboot.sh` drives `S85netboot` with a stub dnsmasq, in the style of
  `test-netbird.sh`. `tools/usbdev/test-usb-network.sh` gains the dnsmasq branch of
  `usb_net_start`, the fallback to `udhcpd`, and `S03usbdev dhcp`. Both run in `alpine:3.24`.

## What a board test must check

- Install from the page: `/data/ironkvm/addons/netboot/dnsmasq --version` runs, the six boot files
  and `boot.ipxe` are in `tftp/`.
- dnsmasq drops to `nobody` and still reads the TFTP root on the exFAT `/data`.
- With the link on (NCM) and network boot on: `S03usbdev dhcp` stops the link's `udhcpd` and
  starts dnsmasq in the `addons` cgroup; the Wi-Fi AP's `udhcpd` keeps running.
- A UEFI x86-64 host PXE boots from the USB NIC (the firmware must offer PXE on a USB NIC), gets
  `ipxe.efi`, then the menu, and sanboots an ISO from `/data`. The same with a BIOS host and
  `undionly.kpxe` if one is at hand.
- The host's lease has no router and no DNS server, and the lease shows on the page.
- From the LAN, `curl http://<board-link-address>:8069/menu.ipxe` gets no menu.
- LAN proxy on: a PXE client on the LAN keeps its address from the LAN's DHCP server, gets
  netboot.xyz by TFTP, and loads the netboot.xyz menu. Turning it off stops dnsmasq on the LAN.
- A reboot with network boot on the link comes up with dnsmasq, not `udhcpd`, on the link.
