# Web UI packages and the server contract

yuzi-co/ironkvm-dist#109. The web UI can be installed, and rolled back, without the server.
This page describes the contract between the two, the package format, where trust comes from,
and what the board does with a package. It is written so the format can move under RAUC or the
home-grown updater later (#51), and so the UI can be served from somewhere other than the board
(#110).

## The contract: `GET /api/version`

Answered without authentication, because the login page is a UI too.

```json
{"code": 0, "msg": "success",
 "data": {"version": "2.4.3", "api": 1, "features": ["direct-clock", "status-bar", "ui-update"]}}
```

| Field | Meaning |
| --- | --- |
| `version` | the installed application release, from `/kvmapp/version`, without the build stamp |
| `api` | `server/common/apicontract.APIVersion`, bumped on a breaking change to the UI's API |
| `features` | the registry in `server/common/apicontract/contract.go`, sorted |

What it exposes: the release number, an integer and a list of feature names. No build stamp
(which names a commit), no host name, no configuration, nothing about sessions. The login page
already loads `index.html` and its content-hashed assets, which identify the build exactly, so
the release number tells a visitor little that the static files did not. A test pins the three
fields, so adding one is a deliberate change.

**Adding a feature.** One line in `features` in `contract.go` on the server, one line in
`SERVER_FEATURES` in `web/src/lib/server-contract.ts`, and the UI code gates on
`useServerFeature('<name>')`. Both lists are sorted, so two branches that add features touch
different lines or adjacent ones, and the resolution is to keep both.

**Bumping the API.** Raise `APIVersion` on the server when an existing UI can no longer read
what the server answers. A UI states the versions it supports in `web/src/lib/ui-api-range.json`;
raise its `max` once the UI has been checked against the new server.

**Older servers.** A server without the endpoint answers 404. The UI reads that as API 1 with
unknown features. It keeps asking for the features that existed before the contract
(`status-bar`, `direct-clock`) and copes with a 404 as it did before; a feature added since
(`ui-update`) is off. A server that answers the contract is taken at its word.

**A UI outside the server's range** shows a notice and keeps running, so the update page that
fixes the mismatch still works.

## The package

```
ironkvm-ui_<version>.tar.gz             gzip tar
├── manifest.json
└── web/                                web/dist as the build leaves it, .gz siblings included
ironkvm-ui_<version>.tar.gz.minisig     minisign signature over the tarball's bytes
ui-latest.json                          the feed entry, published beside latest.json
```

`manifest.json`:

```json
{"format": 1, "name": "ironkvm-ui", "version": "2.4.4",
 "api": {"min": 1, "max": 1},
 "files": {"index.html": "<sha256 hex>", "assets/index-abc.js": "<sha256 hex>", "...": "..."}}
```

`files` lists every file under `web/` and nothing else. The archive holds only directories and
regular files, only those two top-level names, no duplicate paths and no `..`; the server refuses
anything else before it writes a byte.

`ui-latest.json`:

```json
{"format": 1, "version": "2.4.4", "name": "ironkvm-ui_2.4.4.tar.gz",
 "url": "https://...", "size_bytes": 8136728, "sha512": "<base64>", "api": {"min": 1, "max": 1}}
```

`url` is optional and must be absolute https, as in `latest.json`. The signature is fetched from
the package URL with `.minisig` appended. The feed's `api` lets the page say whether the package
fits before downloading it; the manifest inside the signed package is what the install trusts.

**Building one.** `make ui-package VERSION=x.y.z` builds `web/` with `UI_VERSION=x.y.z` (which
writes `dist/ui-version.json`) and runs `scripts/package-ui.sh`, which writes the three files to
`build/ui/`. The archive is reproducible: sorted names, no owners, a fixed time, `gzip -n`.

## Trust chain

1. **The signing key.** `SIGN_KEY` is the minisign secret key, the same kind
   `abslots/sign-image.sh` signs slot images with; a password is read from
   `SIGN_PASSWORD_FILE` on stdin, as there. The secret key never enters a repository.
   Releases are signed with the IronKVM release key, `ironkvm-release` (key ID
   `32F88FAC69726707`, public key `RWQHZ3JprI/4MjT286dEbA+7YYAFxH75jphjig4pEW1n4YKTxtH3z7OX`).
   Its public half is `keys/ironkvm-release.pub` here and `signing/ironkvm-release.pub` in
   ironkvm-dist, whose images install it as `/usr/share/ironkvm/keys/ironkvm-release.pub`.
   The secret half and its password live with the operator and in this repository's GitHub
   Actions secrets `MINISIGN_SECRET_KEY` and `MINISIGN_PASSWORD`.
2. **The build checks the chain.** With `VERIFY_KEYS` set to the public keys boards trust (the
   release key images ship), the build fails and removes the signature unless one of them
   verifies it.
3. **The board's keyring** is the one `slot install` uses: every `*.pub` in
   `/usr/share/ironkvm/keys` (the key the running image was built with) and `/data/ironkvm/keys`
   (the owner's). One key covers images and UI packages.
4. **The server verifies** the signature in Go (`server/service/webui/minisign.go`), both the
   prehashed (`ED`, minisign's default) and the legacy (`Ed`) form, and the trusted comment's
   signature. It does not need the minisign binary, which slot A lacks.
5. **The manifest's hashes** cover every unpacked file, so the unpacked tree is checked against
   the signed manifest before it is served, and the entry document and its assets again on every
   start.

**The owner decides**, as for images (`docs/signing.md` in ironkvm-dist): the file
`/data/ironkvm/no-signature-check` turns the check off, and an empty keyring checks nothing. Both
install with a warning that the API returns and the update page shows. A board whose image ships
no key (slot A, the vendor image) therefore checks nothing until the owner puts a key in
`/data/ironkvm/keys`.

For a test server, `IRONKVM_KEYS_DIRS` (space separated) and `IRONKVM_SIGNATURE_CHECK_OFF` in the
server's environment replace the two paths.

**Test keys.** The Go tests make Ed25519 keys at run time and write them in minisign's format.
`server/service/webui/testdata` holds one public key and two signatures made by the minisign tool,
whose secret half was discarded; nothing trusts that key. `tools/release/test-package-ui.sh` makes
its key with `minisign -G -W` in a temporary directory. No secret key, test or real, is
committed; the only real key in the repository is the release key's public half.

**Releases.** `.github/workflows/ui-release.yml` runs on every `ironkvm-X.Y.Z` tag, the form
`tools/release/release.sh` tags. One job builds `web/` with `UI_VERSION=X.Y.Z` and never sees
the secrets, because `pnpm install` runs third-party code. A second job writes the secrets to
files in `$RUNNER_TEMP`, runs `scripts/package-ui.sh` on that build with `VERIFY_KEYS` set to
`keys/ironkvm-release.pub`, removes the files, and checks the signature with `minisign -V` and
`ui-latest.json` against the package. A third attaches the package, its `.minisig` and
`ui-latest.json` to the tag's GitHub release, whose download URL `ui-latest.json` names. It
waits for the release `release.sh` creates and creates one only if none appears. It does not
write `ui-latest.json` to the `gh-pages` feed: offering the package to boards stays a manual
step, as for `latest.json`. Running the workflow by hand is a dry run that publishes nothing.

## On the board

```
<server dir>/web/                  the UI the application package carries, always there
<server dir>/ui/state.json         {"current": "<id>", "previous": "<id>"}, "" = bundled
<server dir>/ui/releases/<id>/     manifest.json + web/, id = <version>-<sha256 prefix>
```

**Install** (`POST /api/ui/upload` with `package` and `signature`, or `POST /api/ui/update` from
the feed; admin only, under the application update lock):

1. verify the signature against the keyring;
2. scan the archive and read the manifest; refuse a package whose `api` range does not include
   the server's API version;
3. check free space (unpacked size plus 16 MiB);
4. unpack into a work directory, hold every file to the manifest, run the basic check (below);
5. rename the work directory into `releases/`;
6. write the new state (`current` = new, `previous` = what was served) by a rename;
7. serve `GET /` from the new directory through the server's own static handler; if that fails,
   write the old state back and keep serving the old UI;
8. remove every release the state does not name.

**Rollback** (`POST /api/ui/rollback`) serves the previous release, or the bundled UI when there
is none or it no longer passes the check, and removes the release it rolled back from. Rolling
back from the bundled UI is refused.

**Start.** The server serves the current release if it passes the basic check, else the previous
one, else the bundled UI. It changes no file doing so. The basic check: the manifest parses and
states this server's API version, `index.html` matches its hash, and every `/assets/...` file
`index.html` loads is present with its hash.

**Application updates** replace the server directory, `ui/` with it, so the board returns to the
UI the new application carries. A separately installed UI was checked against the server it was
installed on, not the next one.

## Later

- **RAUC or the home-grown updater (#51).** A UI package is a single signed artefact with a
  manifest that names its compatibility, which maps onto a RAUC bundle with one image for a
  `web` slot class, or onto the home-grown updater as a second package type. The keyring is
  already shared with slot images.
- **UI off the board (#110).** `GET /api/version` is the first call such a UI makes to each
  board, and the feature list is what it adapts to per board. The endpoint takes no cookie and
  returns plain JSON, so a CORS allow-list is the only server change it needs for that call.
