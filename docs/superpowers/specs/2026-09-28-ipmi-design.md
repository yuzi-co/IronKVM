# IPMI over LAN

Issue: yuzi-co/ironkvm-dist #3.

## Goal

`ipmitool -I lanplus -H <kvm> -U <user> -P <password> power on|off|cycle|reset|soft|status`
against the managed host, through the ATX lines, plus `chassis status`, `chassis identify` and
`mc info`. Nothing else: no SEL, SDR, FRU, LAN configuration or Serial over LAN.

## Decisions

- **In the server, off by default.** A package `service/ipmi`, pure Go, one UDP socket on port
  623 and one goroutine reading it. The socket is open only while the service is on. The setting
  is `ipmi.enabled` in server.yaml, saved the way `redfish.enabled` is: to the file and to the
  running configuration, with no restart. A file without the block loads with the service off.
- **IPMI 2.0 RMCP+ only.** Cipher suites 3 (RAKP-HMAC-SHA1, HMAC-SHA1-96, AES-CBC-128) and 17
  (RAKP-HMAC-SHA256, HMAC-SHA256-128, AES-CBC-128). Everything else, cipher suite 0 included,
  gets "no cipher suite match". Get Channel Authentication Capabilities answers in IPMI 1.5
  framing, because that is how every client asks, but it offers no IPMI 1.5 authentication
  type, and Get Session Challenge and Activate Session are refused. Every message in a session
  is authenticated and encrypted.
- **Users are the KVM accounts, with a separate IPMI password.** RAKP proves knowledge of the
  password by HMAC, so the board needs the password itself, and it only keeps a bcrypt hash of
  the web password. An admin sets an IPMI password per account on the IPMI settings page. It
  is 12 to 20 printable ASCII characters (20 is the IPMI 2.0 limit), and it must differ from
  the web password: RAKP message 2 hands anyone who knows a user name an HMAC keyed by that
  password, which can be cracked offline, and the web password must not be the thing exposed.
  The page offers a generated 20 character password. Only accounts with a name of at most 16
  bytes (the IPMI limit) can have one. An account without an IPMI password cannot log in over
  IPMI. A disabled account cannot either.
- **Stored encrypted, next to its key.** The IPMI password lives in the account record in
  `/etc/kvm/pwd`, sealed with AES-256-GCM under a random key in `/etc/kvm/ipmi.key` (mode
  0600), with the user name as additional data, so a sealed value copied to another account
  does not open. This is not protection against someone who can read `/etc/kvm`. It keeps the
  password out of the account file itself, which the legacy account mirror and any copy of
  that file would otherwise carry. Deleting the account deletes its IPMI password. Changing the
  web password does not change it.
- **Privilege from the role.** An admin account gets up to ADMINISTRATOR, a user account up to
  USER, as Redfish lets only admins press buttons. A session starts at USER and Set Session
  Privilege Level raises it to at most that limit. ipmitool asks for ADMINISTRATOR by default,
  so a user account needs `-L USER`, and can then read the power state but not change it.
- **Failed logins** count against the web login's brute-force limit, per address and account,
  the same record Redfish uses. A locked out pair is refused at RAKP message 1.
- **The UI says IPMI authentication is weak by design**, and why, on the page and next to the
  switch.

## Protocol

RMCP class IPMI over UDP. The pre-session commands are Get Channel Authentication Capabilities
(IPMI 1.5 framing, session 0), Get Channel Cipher Suites and Get System GUID (RMCP+ framing,
session 0, unauthenticated). Then Open Session, RAKP 1 to 4, and messages in the session.

- The managed system GUID is the Redfish service UUID, from `/etc/kvm/redfish-uuid`, so both
  services name the same machine.
- KG is not set, so the SIK is keyed by the user's password, as the specification requires
  when KG is all zeros. Get Channel Authentication Capabilities reports it that way.
- Session IDs and the managed system random number come from `crypto/rand`. At most 16
  sessions exist at once, counting ones still in the handshake; a session is dropped after 60
  seconds without a valid message. A handshake that fails drops its session.
- Inbound session sequence numbers are checked against a 32 message sliding window, and a
  number seen before is dropped.
- ipmitool resends a request with the same request sequence number when an answer is lost. The
  last answer of each session is kept, and a repeated request within 5 seconds gets it again
  without the command running twice. This matters for Chassis Control.

## Commands

| NetFn | Command | Privilege | Answer |
| --- | --- | --- | --- |
| App | Get Device ID | USER | Firmware version from the application version, IPMI 2.0, chassis device |
| App | Get System GUID | none | The GUID above |
| App | Get Channel Authentication Capabilities | none | RMCP+ only, non-null user names, KG all zeros |
| App | Get Channel Cipher Suites | none | Suites 3 and 17 |
| App | Get Session Challenge, Activate Session | none | Refused: IPMI 1.5 is off |
| App | Set Session Privilege Level | CALLBACK | Up to the session's limit |
| App | Close Session | CALLBACK | Its own session only |
| Chassis | Get Chassis Status | USER | Power on or off from the power LED |
| Chassis | Chassis Control | OPERATOR | See below |
| Chassis | Chassis Identify | OPERATOR | Acknowledged, does nothing: there is no light to turn on |

Anything else answers "invalid command".

## Power

As in Redfish (`service/redfish/system.go`), the power state comes from `vm.PowerLED` only
when `vm.PowerLEDConnected()` says the LED header is wired. Without it the state is unknown:
Get Chassis Status answers "not supported in present state", and every Chassis Control
action that depends on the state is refused the same way. Hard reset presses reset whatever
the state, so it is the only action a board without the LED offers.

| Chassis Control | LED on | LED off | LED unknown |
| --- | --- | --- | --- |
| power down (0) | power, 5 s | nothing | refused |
| power up (1) | nothing | power, 800 ms | refused |
| power cycle (2) | power 5 s, pause 5 s, power 800 ms | refused | refused |
| hard reset (3) | reset, 800 ms | reset, 800 ms | reset, 800 ms |
| diagnostic interrupt (4) | refused | refused | refused |
| soft off (5) | power, 800 ms | nothing | refused |

The command answers as soon as the press is decided. The press runs after, because a 5
second hold is longer than ipmitool waits. Presses are serialized, and after one the lock is
kept for 2 seconds so the next command reads an LED that has had time to follow. A command
that arrives while a press is running answers "node busy". Presses go through
`vm.PressButton`, which also serializes them with the UI, Redfish and the watchdog.

## API

All routes need a web UI login with the admin role.

- `GET /api/ipmi/settings`: the switch, the port, whether the power LED is wired, and the
  accounts with their role, whether they are enabled, whether they have an IPMI password and
  whether their name fits.
- `POST /api/ipmi/settings` `{enabled}`: turns the service on or off. Off ends every session.
- `POST /api/ipmi/users/:username/password` `{password}`: sets an account's IPMI password
  and ends its sessions.
- `DELETE /api/ipmi/users/:username/password`: removes it and ends its sessions.

## Web UI

`Settings > IPMI`, next to Redfish: the switch with the warning, the port, the example
command, whether power on and off are available (the power LED), and the accounts, each with
a set or generate and a remove action.

## Tests

- Unit tests of the RAKP math (the RAKP 2 and 3 HMACs, SIK, K1, K2, the RAKP 4 ICV) against
  vectors computed independently from the specification's formulas, and of AES-CBC-128 and
  the integrity trailer both ways.
- In-process tests with a Go client that runs the whole handshake, and that show cipher suite
  0, IPMI 1.5, a wrong password, a replayed sequence number and an unknown user fail.
- Chassis Control against a fake `PressButton` that records presses, with the LED on, off and
  unknown. No test presses a real button.
- An end-to-end test that runs the real `ipmitool -I lanplus` against the service started in
  the test process: `chassis status`, `mc info`, `power status` and `power on` for cipher
  suites 3 and 17. It skips when ipmitool is not installed; run it in a container that has
  it.
