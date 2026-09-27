# Redfish API

Issue: yuzi-co/ironkvm-dist #2. Virtual media uses the two drives from #8.

## Goal

Standard tools drive the managed host through the KVM: Ansible `community.general.redfish_*`,
`redfishtool`, Metal3/Ironic (sushy) and MAAS. They power the host on and off, reset it, read
its power state, and attach and eject install media. The DMTF Redfish Service Validator passes
against the implemented subset.

## Decisions

- **In the server, no daemon.** A package `service/redfish` with its own routes under
  `/redfish`. Memory is the board's constraint, and the handlers are small JSON documents.
- **A fixed subset.** The server exposes exactly one system, one manager and one chassis, with
  the IDs `1`. No schema can be discovered beyond what is implemented, and every unimplemented
  path answers 404 with a Redfish error body.
- **Local images only.** `InsertMedia` takes an `Image` that names a file under `/data`, as a
  path or a bare file name. A URL (`http://`, `https://`) is refused with
  `ActionParameterValueNotInList`. Downloading an ISO onto the board is its own feature: it
  needs space checks, progress and cancel, and the board's `/data` is the same exFAT store the
  UI uploads to.
- **HTTP and HTTPS both, as the server is configured.** The Redfish specification requires
  TLS, and most clients default to it. The routes go on the server's existing listeners, so a
  board set to `proto: https` serves Redfish over TLS. On plain HTTP it still works, for clients
  that allow it. Recorded, not changed.

## Resources

| Path | Content |
| --- | --- |
| `/redfish` | `{"v1": "/redfish/v1/"}` |
| `/redfish/v1/` | ServiceRoot: `Systems`, `Managers`, `Chassis`, `SessionService`, `Links.Sessions`, `UUID`, `RedfishVersion` |
| `/redfish/v1/odata`, `/redfish/v1/$metadata` | Service document and a minimal CSDL that references the DMTF schemas used |
| `/redfish/v1/Systems`, `/Systems/1` | ComputerSystem: `PowerState` from the power LED, `Actions.#ComputerSystem.Reset` with `ResetType@Redfish.AllowableValues` |
| `/redfish/v1/Chassis`, `/Chassis/1` | Chassis: `PowerState`, links to the system and manager. sushy and the validator expect one. |
| `/redfish/v1/Managers`, `/Managers/1` | Manager: `FirmwareVersion` (the app and image versions), `ManagerType: "BMC"`, `VirtualMedia`, `EthernetInterfaces` |
| `/redfish/v1/Managers/1/EthernetInterfaces`, `/…/eth0` | MAC and IPv4 addresses from `GetInterfaceInfos()` |
| `/redfish/v1/Managers/1/VirtualMedia`, `/…/Cd`, `/…/Disk` | `MediaTypes`, `Image`, `ImageName`, `Inserted`, `WriteProtected`, `ConnectedVia: "Applet"`, actions `InsertMedia` and `EjectMedia` |
| `/redfish/v1/SessionService`, `/…/Sessions`, `/…/Sessions/<id>` | Session login and logout |

The system UUID is generated once and kept in `/etc/kvm/redfish-uuid`, so it stays the same
across restarts.

## Power

The board sees the host only through the power LED and drives it only through two buttons.
`ResetType` maps as follows:

| ResetType | LED on | LED off |
| --- | --- | --- |
| `On` | nothing | power press, 800 ms |
| `ForceOff` | power press, 5 s | nothing |
| `GracefulShutdown` | power press, 800 ms | nothing |
| `ForceRestart` | reset press, 800 ms | reset press, 800 ms |
| `PushPowerButton` | power press, 800 ms | power press, 800 ms |

`PowerCycle`, `Nmi` and `GracefulRestart` are not offered, because the board cannot carry them
out. The action answers 204 once the press is done. It does not wait for the LED: an ATX host
can take seconds to change state, and clients poll `PowerState`. A board without a power LED
wired (`GetGpio` read fails) reports `PowerState: null` and still accepts the
state-independent types.

The `vm` package gets two exported functions, `PressButton(kind string, d time.Duration) error`
and `PowerLED() (bool, error)`, which `SetGpio` and `GetGpio` then use too. A button press holds
a mutex, so a Redfish reset and a UI press never overlap.

## Virtual media

`Cd` maps to the `cdrom` drive and `Disk` to the `disk` drive. `storage` gets exported wrappers
over the functions from #8: `ListDrives`, `InsertDrive(id, file, ro)` and `EjectDrive(id)`. The
drive lock, the check that one image is never in both drives, and the forced eject from #25 all
apply as they do from the UI.

- `InsertMedia`: `Image` is required. `WriteProtected` defaults to true and applies to `Disk`
  only; `Cd` is always write-protected. `Inserted: false` is refused. `TransferMethod` and
  `TransferProtocolType` are accepted only when absent, because a local file has neither.
- `EjectMedia`: empties the drive, with the forced eject where the kernel has it.
- Errors map to Redfish messages: an unknown image is `ResourceNotFound`, an image in the
  other drive is `ResourceInUse`, and a medium the host locked on a kernel without
  `forced_eject` is `ActionNotSupported` with the storage error text.
- With the virtual disk function off (no drives), the VirtualMedia collection is empty.

## Authentication

- **Basic auth**: every request may carry it. It checks the account through
  `authn.DefaultStore.Authenticate` and applies the login's brute-force limit.
- **Sessions**: `POST /redfish/v1/SessionService/Sessions` with `UserName` and `Password` returns
  201, the `X-Auth-Token` header and a `Location`. Tokens are random, 32 bytes, and held in
  memory, with a 30-minute idle timeout and at most 16 at a time. The oldest is dropped when a
  new one would pass the limit. `DELETE` on the session logs out. A server restart ends every
  session, and the clients log in again.
- **API keys**: `X-Auth-Token` also accepts an existing API key, so automation can use a key
  instead of a password.
- **Roles**: any account may `GET`. Actions and session management of other users need admin.
- `/redfish` and `/redfish/v1/` answer without auth, as the specification requires. Every other
  path answers 401 with `WWW-Authenticate: Basic realm="IronKVM"`.
- On a board with authentication disabled, every request acts as admin, as the rest of the API
  does.

## Protocol details

`OData-Version: 4.0` on every response, `@odata.id`, `@odata.type` and `@odata.context` on
every resource, ETags on the system and on each virtual media resource. `If-Match` is not
required. Errors use the standard `error` object with `@Message.ExtendedInfo` from the Base
message registry.

## Units

- `service/redfish/router.go`: routes, the auth middleware, and the error writer.
- `service/redfish/session.go`: the session store.
- `service/redfish/system.go`, `manager.go`, `chassis.go`, `media.go`: one file per resource.
- `vm` and `storage` get the exported functions named above, and nothing in them moves.

## Testing

- Handler tests with `httptest` and fakes for the button, the LED and the drives: each
  `ResetType` in each LED state presses what the table says, the auth paths (none, basic, bad
  basic, session, expired session, API key, user vs admin), and each media error.
- A JSON test that every resource carries `@odata.id` and `@odata.type`, and that every link
  resolves inside the server.
- On the board: DMTF `Redfish-Service-Validator` from the workstation, `redfishtool` power and
  media commands, and one Ansible `redfish_command` run for `PowerOn` and `VirtualMediaInsert`.
- RSS of the server before and after, recorded in the issue.

## Out of scope

URL images, event subscriptions (`EventService`), `AccountService` writes, `SerialConsole`
(no host serial console exists), `UpdateService`, `LogServices`, and more than one system.
