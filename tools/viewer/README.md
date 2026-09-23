# viewer

A headless client for the board's three video paths: WebRTC, H.264 direct and
MJPEG. It connects the way the web UI does, counts what arrives, and discards
it. Use it to measure what each path costs the board without a browser, and
without a person to switch modes.

## Build

The tool uses pion, the same WebRTC library as the server, so a container with Go
is enough:

```shell
MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/v" -v nanokvm-gomod:/go/pkg/mod \
  -w /v -e CGO_ENABLED=0 -e GOOS=windows -e GOARCH=amd64 golang:1.25 \
  go build -o viewer.exe .
```

Run it on the workstation, not in a container. A WebRTC session needs a UDP
path to the board, and container networking gets in the way.

## A session token

Every stream route needs a session. `viewer mint` signs one from the board's own
secret. Do this only on a board whose owner agreed to it, because it goes around
the login.

```shell
ssh root@<device> 'printf "%s\n" "$(cat /etc/kvm/.jwt_secret)"; cat /etc/kvm/pwd' \
  | ./viewer.exe mint -out kvm-token -ttl 3h
```

The token names the first enabled administrator and carries that account's token
version. A password change or a logout with `revokeTokensOnLogout` changes that
version, and the token stops working at once. The file is mode 0600. Do not
print it.

## Measure

```shell
./viewer.exe -mode webrtc -secs 40 -token kvm-token
./viewer.exe -mode direct -secs 40 -token kvm-token
./viewer.exe -mode mjpeg  -secs 40 -token kvm-token
```

It prints the receive rate every 5 seconds and a total at the end. Sample the
board at the same time, from `/tmp` and never from `/kvmapp`. Every 5 seconds,
take the deltas of the `cpu` line in `/proc/stat` for the board, of fields 14
and 15 of `/proc/<server pid>/stat` for the server, and of `tx_bytes` for eth0
in `/proc/net/dev`. There is one core, so the ticks in a row are the whole
board. Take the phase times from the board's clock.

Nothing measures anything without an HDMI signal. Read `/api/vm/hdmi` first.

## Measured on 2026-09-23

Image j, 1000 MHz, 1080p, 40 seconds for each path:

| Path         | Board busy | Server | Sent      |
| ------------ | ---------- | ------ | --------- |
| No viewer    | 5%         | 0%     | 0         |
| WebRTC       | 28.5%      | 21%    | 176 KB/s  |
| H.264 direct | 17%        | 10%    | 325 KB/s  |
| MJPEG        | 98%        | 90%    | 5.6 MB/s  |

WebRTC costs about 3.5 times as much as H.264 direct for each byte it delivers.
It sends about 143 encrypted RTP packets a second through Go. Direct sends one
WebSocket message for each frame.
