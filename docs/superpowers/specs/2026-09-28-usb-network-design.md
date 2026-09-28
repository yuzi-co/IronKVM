# USB Network Link to the Host

**Date:** 2026-09-28
**Issue:** yuzi-co/ironkvm-dist#10
**Repos:** IronKVM (`S03usbdev`, the server, the web UI) and ironkvm-dist (the capability word
and its kernel requirement).

## Purpose

The gadget can carry a network function. It gives the KVM and the managed host a private link
with no LAN behind it. The link serves file transfer, a host agent, and network boot of the host.

## What exists today

`S03usbdev` already builds `ncm.usb0` when `/boot/usb.ncm` exists, or `rndis.usb0` when
`/boot/usb.rndis0` exists. The web UI has one switch for it, and that switch writes the RNDIS
marker. Nothing gives usb0 an address and nothing answers DHCP on it, so the host sees an adapter
with no network.

The kernel builds NCM and ECM in (`CONFIG_USB_F_NCM=y`, `CONFIG_USB_F_ECM=y`,
`CONFIG_USB_CONFIGFS_NCM=y`, `CONFIG_USB_CONFIGFS_ECM=y` in
`socs/sophgo-sg2002/kernel/evidence/5.10.270-ironkvm0/kernel.config`). No kernel change is needed.

## Decisions

- **Mode.** One setting: `off`, `ncm` or `ecm`. It is off by default. NCM is the default when the
  operator turns the link on: Linux, macOS and Windows 11 drive it with an in-box driver. ECM is
  the fallback for a host that has no NCM driver.
- **Markers.** `/boot/usb.ncm` and `/boot/usb.ecm`. The server writes exactly one and removes the
  others. `/boot/usb.rndis0` stays readable, so a board that an older server set up keeps its
  link, but the UI no longer offers RNDIS. If more than one marker exists, `S03usbdev` builds NCM
  first, then ECM, then RNDIS.
- **Subnet.** `172.31.255.0/30` by default. The operator can set another private IPv4 network with
  a prefix from /24 to /30. The board takes the first address and the host takes the second. The
  setting lives in `/etc/kvm/usb-network.subnet`, on `/data`, as one line in CIDR form.
  `S03usbdev` falls back to the default if the file is absent or not valid.
- **DHCP.** The image has no dnsmasq binary. `S80dnsmasq` exits when `/etc/dnsmasq.conf` is
  absent, and the dist image does not carry dnsmasq. BusyBox `udhcpd` is in the image, from
  busybox-extras, for the Wi-Fi access point. `S03usbdev` runs a separate `udhcpd` instance for the
  USB link, with its own configuration and pid file under `/tmp/usb-network`. It serves the one
  interface, leases one address (the host address), and offers the subnet mask only: no router
  and no DNS server.
- **No routing.** The board sets `forwarding` to 0 for the interface, and inserts `FORWARD` drop
  rules for `usb+` in both directions when `iptables` exists. The per-interface value alone is not
  enough, because a later write to `net.ipv4.ip_forward` (NetBird does that on a routing peer)
  resets every interface. There is no NAT.
- **Interface name.** u_ether registers the net device at the first bind and keeps it for the life
  of the function directory, and `S03usbdev` never removes a function directory. A switch from NCM
  to ECM without a reboot therefore gives ECM `usb1`. The script reads the name from the
  function's `ifname` attribute, and it clears the address from every other network function, so
  two interfaces never hold the same subnet.
- **Endpoint budget.** NCM and ECM each take a bulk pair and an interrupt IN endpoint for
  notifications, so two IN and one OUT, the same as the RNDIS entry. The budget stays six IN and
  seven OUT. With HID (3/3) the combinations that fit are console + disk + audio, and disk +
  network + audio. The console and the network do not fit together.
- **Refusal.** The server refuses to turn the link on when the network does not fit beside the
  functions that are on, with the existing refusal sentence that names what to turn off. A switch
  between NCM and ECM costs nothing more, so it always fits. Turning the link off always fits.
- **Fit list.** `GET /api/vm/device/virtual` also returns `fits`: every largest set of optional
  functions that fits beside HID. The UI shows it, so the operator sees which functions fit
  together before a refusal.
- **Re-enumeration.** Applying a change runs `S03usbdev stop` and `start`, so the host sees the
  whole gadget disconnect and return. The UI says so and asks for confirmation before it applies.

## API

`GET /api/vm/device/usb-network`

```json
{ "mode": "ncm", "subnet": "172.31.255.0/30", "board": "172.31.255.1",
  "host": "172.31.255.2", "active": true, "fits": true, "refusal": "" }
```

`mode` can also be `rndis` on a board that an older server set up. `fits` answers whether the
network can be turned on now, and `refusal` is the reason when it cannot.

`POST /api/vm/device/usb-network` with `{ "mode": "ecm", "subnet": "172.31.255.0/30" }`.
The server validates the subnet, checks that it does not overlap an address on another interface
of the board, checks the budget, writes the subnet file and the marker, and rebuilds the gadget.

The old switch, `POST /api/vm/device/virtual` with `network`, stays. It now writes the NCM marker
and removes all three.

## Web

The Device settings page already holds the USB functions and the endpoint budget. The network
switch there becomes a USB network section: a mode selector, the subnet, the two addresses, and an
Apply button behind a confirmation that says the gadget re-enumerates. The budget panel shows the
fit list. Every locale file carries the new strings.

## ironkvm-dist

`usb-network` joins the `CAPABILITIES` vocabulary, with `CONFIG_USB_CONFIGFS_NCM` and
`CONFIG_USB_CONFIGFS_ECM` as its kernel requirement. The NanoKVM declares it, in both copies of its
description.

## Tests

- Shell: `tools/usbdev/test-usb-network.sh` covers the marker choice, the subnet parse, the
  `udhcpd` configuration, and the wiring in `start_usb_dev`. `tools/service/test-usb-endpoints.sh`
  follows the new network directories.
- Go: the mode and subnet validation, the budget check for each transition, the fit list, and the
  agreement between the Go table and the shell script.

## What a board test must check

- NCM on a Linux host: the host gets 172.31.255.2/30 and no default route and no DNS, and reaches
  the web UI at 172.31.255.1.
- ECM on a Linux or macOS host, the same checks.
- A switch from NCM to ECM without a reboot: the ECM interface (probably `usb1`) holds the address,
  `usb0` holds none, and `udhcpd` serves the new interface.
- From the host, with a route added by hand through 172.31.255.1: nothing on the LAN answers.
- The console and the network together are refused, and the keyboard keeps working after each
  change.
- Windows 11 with NCM: the adapter appears with no added driver.
