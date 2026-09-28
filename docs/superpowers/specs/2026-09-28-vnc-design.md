# VNC server

Issue: yuzi-co/ironkvm-dist #4.

## Goal

A standard VNC client (TigerVNC, Remmina and others that speak Tight) views and controls
the host. The board does no pixel work: the frames are the JPEG frames that the hardware encoder
already makes for the MJPEG stream.

## Decisions

- **In the server, off by default.** A package `service/vnc` holds a pure-Go RFB 3.8 server. It
  listens on TCP 5900 only while it is on. The port can be changed.
- **One session at a time.** A second client that authenticates while a session is open gets a
  failed SecurityResult with the reason "another VNC session is active". The open session is
  never taken over, so a stranger with a password cannot kick the owner out. The settings page
  can end the open session. At most four handshakes run at once, because every TLS handshake
  costs the core a public-key operation.
- **Settings in server.yaml, under `vnc`.** The web UI saves them the way it saves
  `redfish.enabled`: to server.yaml and to the running configuration, with no restart. A change
  of port, or turning the server off, closes the listener and ends the session.

| Setting | Default | Range |
| --- | --- | --- |
| `enabled` | false | |
| `port` | 5900 | 1 to 65535, not the web ports |
| `maxFps` | 15 | 1 to 60 |
| `vncAuth` | false | plain VNC authentication, see below |

## Frames

- **Source.** The VNC session subscribes to the MJPEG streamer in `service/stream/mjpeg`. A
  subscription is a viewer of that capture loop the same as a browser, without an HTTP
  response: it has its own frame slot and no writer goroutine. So a browser and a VNC client
  that watch at the same time share one hardware read per frame, the duplicate suppression
  applies to both, and the VNC viewer counts toward the HDMI capture demand. The MCP
  screenshot tool reads the encoder directly, one frame at a time. VNC does not use that path:
  a second loop of reads beside the MJPEG loop would read the same encoder twice per frame.
- **Pacing.** The session takes a frame from its slot only while the client has an update
  request outstanding, and not more often than `maxFps`. A frame that is not taken stays in the
  slot, and the capture loop reads nothing while every viewer holds a frame. So the capture
  rate follows the slowest viewer, the same as for a browser on a slow link. An unchanged
  screen sends nothing: an incremental request waits until the picture changes, or until the
  five-second refresh of the streamer.
- **Encoding.** Each update is one Tight rectangle over the whole framebuffer, with the JPEG
  compression-control byte (`0x90`), a compact length and the JPEG bytes as the encoder made
  them. The server does not decode, scale or re-encode. A non-incremental request is answered
  with the last frame sent, so a client that asks for a full refresh gets it at once.
- **Size.** The server reads the frame size from the JPEG header. The ServerInit message uses
  the current capture size. If a frame has another size, the server sends a DesktopSize
  pseudo-rectangle as its own update, and sends the frame on the next request. A client that
  did not ask for DesktopSize is disconnected when the size changes, with the reason in the log
  and on the settings page.
- **Client support.** The client must accept Tight encoding and a true-colour pixel format of
  16 or 32 bits. The server cannot fall back to Raw without decoding JPEG on the CPU, so a
  client without Tight is disconnected with a reason. RealVNC Viewer and macOS Screen Sharing
  do not offer Tight and are not supported.

## Input

- **Keyboard.** KeyEvent carries an X11 keysym. Printable keysyms (Latin-1 and the Unicode
  range `0x01000000`) become a character, and the character goes through the US layout table
  in `service/hid` that paste uses. The function and modifier keysyms (arrows, F1 to F24,
  keypad, Shift, Control, Alt, Super, AltGr) use a table in `service/vnc`. The server keeps
  the set of keys held and sends an 8-byte boot keyboard report on each change, with up to six
  keys. A character that needs Shift on the US layout adds Shift while its key is held, so a
  client that sends `!` without Shift still types `!`. The release of a keysym releases the key
  that its press chose, even if the modifiers changed in between.
- **Pointer.** PointerEvent goes to the absolute pointer. The position is scaled from the
  framebuffer size to the 1 to 32768 range that the web UI uses. Buttons 1, 2 and 3 are left,
  middle and right. Buttons 4 and 5 are one wheel step up and down, sent on the press only.
- **Arbitration.** Keyboard and pointer reports go through the same HID writers and the same
  manual-input reservation as the web UI websocket (`inputcontrol.ManualSession`), with the
  same PicoClaw session lock check. The control mode (`service/controlmode`) is respected in
  the same way: an AI-held mode blocks or pre-empts VNC input exactly as it does for a
  browser. When the session ends, held keys and buttons are released.

## Authentication

- **VeNCrypt, X509Plain (default).** Security type 19, subtype 262. The server offers VeNCrypt
  0.2 with X509Plain only. The TLS handshake uses the server's certificate from
  `cert.crt` and `cert.key` in server.yaml, loaded again for each connection. If those files do
  not load, the server calls `utils.EnsureCert` and uses `/etc/kvm/server.crt`. After TLS the
  client sends a username and a password, checked against the KVM accounts in `authn`, under
  the web login's brute-force limit, with the same two-second delay after a failure. A
  disabled account, or an account that must change its password first, is refused.
- **Plain VNC authentication (off by default).** Security type 2, offered only when `vncAuth`
  is on and a VNC password is set. The VNC password is separate from the accounts, is 6 to 8
  characters (DES uses the first 8), and is kept in `/etc/kvm/vnc.passwd` with mode 0600,
  never in server.yaml. The protocol sends frames and keystrokes without encryption. The
  settings page shows a warning while it is on. macOS Screen Sharing needs this mode, but it
  does not support Tight, so it still does not work.
- `authentication: disable` in server.yaml does not turn off VNC authentication.

## Settings page

`Settings > VNC`, for admins, next to Redfish. The page shows the switch, the port, the frame
rate limit, the plain-authentication switch with the password field and its warning, and the
state: listening or not, the last error, and the open session with the client address, the
user and the start time. A button ends the session.

API, all behind `CheckToken` and the admin role:

- `GET /api/vnc/settings`, `POST /api/vnc/settings`
- `GET /api/vnc/state`
- `POST /api/vnc/disconnect`

## Memory and CPU

- **Memory.** Per session: three goroutines, the TLS state (about 40KB with its record
  buffers), one small read buffer, and up to three JPEG references (the frame in the slot, a
  frame held for a resize, and the last frame sent for a full refresh). The streamer shares the
  JPEG slice, so no frame is copied. At 1080p a frame is 100KB to 350KB, so the worst case is
  about 1.1MB. With no session the server holds only the listener. The binary grows by
  `crypto/des` and the package, well under 200KB.
- **CPU.** The server does no pixel work. Per frame it scans the JPEG header for its size and
  writes the bytes to the socket. With TLS the main cost is the record encryption: Go picks
  ChaCha20-Poly1305 on a core without AES instructions, which the C906 is. Expected cost is
  below the MJPEG-over-HTTPS path at the same frame rate and size, because there is no HTTP
  framing and the frame rate is capped by `maxFps` and by the client's requests. The MJPEG
  path costs about 98% of the core per session today.

## Board test

The board test must measure, for each case, the server process CPU from `top` over 60 seconds,
the RSS of the server, the frames per second the client receives, and the capture reads per
second from the metrics endpoint:

1. One browser on MJPEG at 1080p, no VNC client (the reference).
2. One TigerVNC client over VeNCrypt X509Plain at `maxFps` 15 and 30, with a moving picture
   (a video on the host) and with a still picture.
3. The same with plain VNC authentication, to separate the TLS cost from the rest.
4. A browser on MJPEG and a VNC client at once, to confirm that they share one read per frame.

It must also check: the JPEG rectangles decode in TigerVNC and Remmina, and in noVNC through
websockify with plain VNC authentication; the keyboard types the full US layout, the modifiers
and the function keys; the pointer lands where it is clicked at two resolutions; a resolution change on the host resizes the client; a second client
is refused; the disconnect button ends the session and releases held keys; the RSS goes back
after the session ends.

## Not in this change

- Dirty-rectangle updates. Every update is the full frame.
- The QEMU extended key event, which gives layout-independent scancodes.
- Clipboard, audio, cursor shape, and the LED state pseudo-encoding.
- More than one session, and a view-only session.
